package grpcadapter

import "log/slog"

// Server is a gRPC stub (proto ready; listener not bound in memory mode).
type Server struct {
	addr string
	log  *slog.Logger
}

// NewServer creates a stub gRPC server.
func NewServer(addr string, log *slog.Logger) *Server {
	if log == nil {
		log = slog.Default()
	}
	return &Server{addr: addr, log: log}
}

// Start logs intent without binding a listener.
func (s *Server) Start() {
	s.log.Info("grpc.stub", "addr", s.addr, "note", "proto defined; implement generated server later")
}
