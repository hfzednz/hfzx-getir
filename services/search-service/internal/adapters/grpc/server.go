package grpcadapter

import "log/slog"

type Server struct {
	addr string
	log  *slog.Logger
}

func NewServer(addr string, log *slog.Logger) *Server {
	if log == nil {
		log = slog.Default()
	}
	return &Server{addr: addr, log: log}
}

func (s *Server) Start() {
	s.log.Info("grpc.stub", "addr", s.addr)
}
