package proxy

import (
	"bytes"
	"encoding/binary"
	"io"
	"net"
	"sync"

	"github.com/blessedvictim/rongxy/proxy/auth"
)

var implementations = []func(*Server, *net.TCPConn, []byte){nil, nil, nil, nil, ImplSocks4, ImplSocks5}

//nolint
func ImplSocks4(s *Server, conn *net.TCPConn, buf []byte) {
	switch command := buf[1]; command {
	case 1:
		port := binary.BigEndian.Uint16(buf[2:4])
		ip := net.IP(buf[4:8])
		addr := &net.TCPAddr{IP: ip, Port: int(port)}
		buf := buf[8:]
		i := bytes.Index(buf, []byte{0})
		if i < 0 {
			s.logger.Debugf("[%s] unable to locate SOCKS4 user", conn.RemoteAddr())
			return
		}
		user := buf[:i]
		s.logger.Debugf("[%s] incoming SOCKS4 TCP/IP stream connection, user=%q, raddr=%s", conn.RemoteAddr(), user, addr)
		remote, err := net.DialTCP("tcp4", conn.RemoteAddr().(*net.TCPAddr), addr)
		if err != nil {
			s.logger.Debugf("[%s] unable to connect to remote host: %v", conn.RemoteAddr(), err)
			conn.Write([]byte{0, 0x5b, 0, 0, 0, 0, 0, 0})
			return
		}
		conn.Write([]byte{0, 0x5a, 0, 0, 0, 0, 0, 0})
		transfer(conn, remote)
	default:
		s.logger.Debugf("[%s] unsupported command, closing connection", conn.RemoteAddr())
	}
}

func ImplSocks5(s *Server, conn *net.TCPConn, buf []byte) {
	messVersion, err := makeVersionMessage(buf)
	if err != nil {
		s.logger.Debugf("[%s] corrupted hello SOCKS5 message from ", conn.RemoteAddr())
		return
	}

	if !bytes.Contains(messVersion.Methods, auth.SupportedAuth) {
		s.logger.Debugf("[%s] send unsupported SOCKS5 authentication methods", conn.RemoteAddr())
		_, _ = conn.Write([]byte{SOCKS5Version, auth.NoAcceptable})
		return
	}
	_, _ = conn.Write([]byte{SOCKS5Version, auth.None})
	buf = make([]byte, 256)
	n, err := conn.Read(buf)
	if err != nil {
		s.logger.Debugf("[%s] unable to read SOCKS header: %v", conn.RemoteAddr(), err)
		return
	}
	buf = buf[:n]
	requestMessage, err := makeSocksRequestMessage(buf)
	if err != nil {
		s.logger.Debugf("[%s] corrupted SOCKS5 TCP/IP connection request %v : %v", conn.RemoteAddr(), err, buf)
		_, _ = conn.Write([]byte{SOCKS5Version, ReplyGeneralFailure, 0x00, AddrTypeIPv4, 0, 0, 0, 0, 0, 0})
		return
	}

	switch version := requestMessage.Version; version {
	case SOCKS5Version:
		switch command := requestMessage.Cmd; command {
		case CmdConnect:
			switch addrtype := requestMessage.Atyp; addrtype {
			case AddrTypeIPv4:
				ip := net.IP(requestMessage.DstAddr)
				port := binary.BigEndian.Uint16(requestMessage.DstPort)
				addr := &net.TCPAddr{IP: ip, Port: int(port)}
				s.logger.Debugf("[%s] incoming SOCKS5 TCP/IP stream connection, raddr=%s", conn.RemoteAddr(), addr)
				remote, err := net.DialTCP("tcp", nil, addr)
				if err != nil {
					s.logger.Debugf("[%s] unable to connect to remote host: %v", conn.RemoteAddr(), err)
					_, _ = conn.Write([]byte{SOCKS5Version, ReplyHostUnreachable, 0x00, 0x01, 0, 0, 0, 0, 0, 0})
					return
				}
				_, _ = conn.Write([]byte{SOCKS5Version, ReplySucceeded, 0x00, 0x01, ip[0], ip[1], ip[2], ip[3], byte(port >> 8), byte(port)})
				transfer(conn, remote)
			case AddrTypeDomainName:
				ip, err := net.ResolveIPAddr("ip", string(requestMessage.DstAddr))
				if err != nil {
					s.logger.Debugf("[%s] unable to resolve IP address: %q, %v", conn.RemoteAddr(), requestMessage.DstAddr, err)
					_, _ = conn.Write([]byte{SOCKS5Version, ReplyHostUnreachable, 0x00, AddrTypeIPv4, 0, 0, 0, 0, 0, 0})
					return
				}
				port := binary.BigEndian.Uint16(requestMessage.DstPort)
				addr := &net.TCPAddr{IP: ip.IP, Port: int(port)}
				remote, err := net.DialTCP("tcp", conn.RemoteAddr().(*net.TCPAddr), addr)
				if err != nil {
					s.logger.Debugf("[%s] unable to connect to remote host: %v", conn.RemoteAddr(), err)
					_, _ = conn.Write([]byte{SOCKS5Version, ReplyHostUnreachable, 0x00, AddrTypeIPv4, 0, 0, 0, 0, 0, 0})
					return
				}
				_, _ = conn.Write([]byte{
					SOCKS5Version, ReplySucceeded, 0x00, AddrTypeIPv4,
					addr.IP[0], addr.IP[1], addr.IP[2], addr.IP[3],
					byte(port >> 8), byte(port),
				})
				transfer(conn, remote)

			default:
				s.logger.Debugf("[%s] unsupported SOCKS5 address type: %d", conn.RemoteAddr(), addrtype)
				_, _ = conn.Write([]byte{SOCKS5Version, ReplyAddrTypeNotSupported, 0x00, AddrTypeIPv4, 0, 0, 0, 0, 0, 0})
			}
		default:
			s.logger.Debugf("[%s] unknown SOCKS5 command: %d", conn.RemoteAddr(), command)
			_, _ = conn.Write([]byte{SOCKS5Version, ReplyCommandNotSupported, 0x00, AddrTypeIPv4, 0, 0, 0, 0, 0, 0})
		}
	default:
		s.logger.Debugf("[%s] unnknown version after SOCKS5 handshake: %d", conn.RemoteAddr(), version)
		_, _ = conn.Write([]byte{0x05, 0x07, 0x00, 0x01, 0, 0, 0, 0, 0, 0})
	}
}

func transfer(in, out net.Conn) {
	wg := new(sync.WaitGroup)
	wg.Add(2)
	f := func(in, out net.Conn, wg *sync.WaitGroup) {
		_, _ = io.Copy(out, in)
		//nolint
		// s.logger.Debugf("xfer done: in=%v\tout=%v\ttransfered=%d\terr=%v", in.RemoteAddr(), out.RemoteAddr(), n, err)
		if conn, ok := in.(*net.TCPConn); ok {
			_ = conn.CloseWrite()
		}
		if conn, ok := out.(*net.TCPConn); ok {
			_ = conn.CloseRead()
		}
		wg.Done()
	}
	go f(in, out, wg)
	go f(out, in, wg)
	wg.Wait()
	_ = out.Close()
}
