package grpcadapter

import "log/slog"

type Server struct {
	addr string
	log  *slog.Logger
}

func NewServer(addr string, log *slog.Logger) *Server {
	return &Server{addr: addr, log: log}
}

func (s *Server) Start() {
	if s.log != nil {
		s.log.Info("grpc.listen.stub", "addr", s.addr)
	}
}
