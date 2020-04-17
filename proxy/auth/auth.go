package auth

const None byte = 0x0

//nolint
// TODO implementations
const GSSAPI byte = 0x01
const UsernamePassword byte = 0x02
const NoAcceptable byte = 0xFF

var SupportedAuth = []byte{None}
