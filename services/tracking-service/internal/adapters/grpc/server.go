package grpcadapter

import (
	"log/slog"
	"net"
)

// Server is a gRPC stub listener (HTTP is primary).
type Server struct {
	addr string
	log  *slog.Logger
}

// NewServer returns a stub gRPC server.
func NewServer(addr string, log *slog.Logger) *Server {
	if log == nil {
		log = slog.Default()
	}
	return &Server{addr: addr, log: log}
}

// Start logs that gRPC is stubbed (no real listener in memory mode).
func (s *Server) Start() {
	if s.addr == "" {
		return
	}
	s.log.Info("grpc.stub", "addr", s.addr, "note", "proto defined; HTTP is primary")
	_ = net.JoinHostPort
}
