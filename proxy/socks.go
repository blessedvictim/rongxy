package proxy

import (
	"net"
	"sync"
)

var connections = new(sync.WaitGroup)

func (s Server) HandleConn(conn *net.TCPConn) {
	connections.Add(1)
	defer func() {
		_ = conn.Close()
	}()
	defer connections.Done()

	// SOCKS does not include a length in the header, so take
	// a punt that each request will be readable in one go.
	buf := make([]byte, 256)
	n, err := conn.Read(buf)
	if err != nil || n < 2 {
		s.logger.Debugf("[%s] unable to read SOCKS header: %v", conn.RemoteAddr(), err)
		return
	}
	buf = buf[:n]

	socksVersion := buf[0]

	impl := implementations[socksVersion]
	if impl != nil {
		impl(&s, conn, buf)
	} else {
		s.logger.Debugf("[%s] unknown SOCKS version: %d", conn.RemoteAddr(), socksVersion)
	}
}
