//nolint
package proxy

import "errors"

type VersionMessage struct {
	Version  byte
	NMethods byte
	Methods  []byte
}

func makeVersionMessage(buf []byte) (*VersionMessage, error) {
	if len(buf) < 3 {
		return nil, errors.New("hello message corrupted")
	}
	return &VersionMessage{
		Version:  buf[0],
		NMethods: buf[1],
		Methods:  buf[2:],
	}, nil
}

//The SOCKS request is formed as follows:
//
//+----+-----+-------+------+----------+----------+
//|VER | CMD |  RSV  | ATYP | DST.ADDR | DST.PORT |
//+----+-----+-------+------+----------+----------+
//| 1  |  1  | X'00' |  1   | Variable |    2     |
//+----+-----+-------+------+----------+----------+
//
//Where:
//
//o  VER    protocol version: X'05'
//o  CMD
//o  CONNECT X'01'
//o  BIND X'02'
//o  UDP ASSOCIATE X'03'
//o  RSV    RESERVED
//o  ATYP   address type of following address
//o  IP V4 address: X'01'
//o  DOMAINNAME: X'03'
//o  IP V6 address: X'04'
//o  DST.ADDR       desired destination address
//o  DST.PORT desired destination port in network octet
//order

type SocksRequest struct {
	Version byte
	Cmd     byte
	RSV     byte
	Atyp    byte
	DstAddr []byte
	DstPort []byte
}

func makeSocksRequestMessage(buf []byte) (*SocksRequest, error) {
	if len(buf) < 10 {
		return nil, errors.New("request message corrupted")
	}

	n := len(buf)

	version := buf[0]
	cmd := buf[1]
	rsv := buf[2]
	atyp := buf[3]
	dstAddr := buf[4 : n-2]
	dstPort := buf[n-2 : n]

	return &SocksRequest{
		Version: version,
		Cmd:     cmd,
		RSV:     rsv,
		Atyp:    atyp,
		DstAddr: dstAddr,
		DstPort: dstPort,
	}, nil
}

//CMD
////o  CONNECT X'01'
////o  BIND X'02'
////o  UDP ASSOCIATE X'03'

const (
	CmdConnect      byte = 0x01
	CmdBind         byte = 0x02
	CmdUdpAssociate byte = 0x03
)

//ATYP   address type of following address
//o  IP V4 address: X'01'
//o  DOMAINNAME: X'03'
//o  IP V6 address: X'04'

const (
	AddrTypeIPv4       byte = 0x01
	AddrTypeDomainName byte = 0x03
	AddrTypeIPv6       byte = 0x04
)
