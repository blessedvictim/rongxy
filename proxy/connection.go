package proxy

import (
	"net"

	"go.uber.org/zap"
)

type Server struct {
	serveAddr net.Addr
	logger    *zap.SugaredLogger
}

func New(addr net.Addr) Server {
	logger, _ := zap.NewDevelopment()
	sugared := logger.Sugar()

	return Server{
		serveAddr: addr,
		logger:    sugared,
	}
}

func (s Server) ServeConnections() error {
	l, err := net.Listen("tcp", s.serveAddr.String())
	if err != nil {
		s.logger.Debugf("unable to listen on SOCKS port [%s]: %v", s.serveAddr.String(), err)
		s.logger.Debugf("shutting down")
		return nil
	}
	defer func() {
		_ = l.Close()
	}()
	s.logger.Debugf("listening for incoming SOCKS connections on [%s]\n", s.serveAddr.String())

	for {
		c, err := l.Accept()
		if err != nil {
			s.logger.Debugf("failed to accept incoming SOCKS connection: %v", err)
		}
		go s.HandleConn(c.(*net.TCPConn))
	}
}
