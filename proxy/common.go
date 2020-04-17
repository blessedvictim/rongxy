package proxy

const (
	SOCKS4Version byte = 0x05
	SOCKS5Version byte = 0x05
)

//nolint
//Replies
//o  REP    Reply field:
//o  X'00' succeeded
//o  X'01' general SOCKS server failure
//o  X'02' connection not allowed by ruleset
//o  X'03' Network unreachable
//o  X'04' Host unreachable
//o  X'05' Connection refused
//o  X'06' TTL expired
//o  X'07' Command not supported
//o  X'08' Address type not supported
//o  X'09' to X'FF' unassigned

const (
	ReplySucceeded            byte = 0x00
	ReplyGeneralFailure       byte = 0x01
	ReplyNotAllowed           byte = 0x02
	ReplyNetUnreachable       byte = 0x03
	ReplyHostUnreachable      byte = 0x04
	ReplyConnRefused          byte = 0x05
	ReplyTTLExpired           byte = 0x06
	ReplyCommandNotSupported  byte = 0x07
	ReplyAddrTypeNotSupported byte = 0x08
)
