// Package grpc provides a stub gRPC listener placeholder.
package grpc

import "log/slog"

// Server is a noop gRPC stub (proto defined; listener not started in HTTP-first mode).
type Server struct {
	addr string
	log  *slog.Logger
}

// NewServer returns a gRPC stub.
func NewServer(addr string, log *slog.Logger) *Server {
	if log == nil {
		log = slog.Default()
	}
	return &Server{addr: addr, log: log}
}

// Start logs stub readiness (no real listener).
func (s *Server) Start() {
	s.log.Info("grpc.stub", "addr", s.addr, "note", "proto defined; listener not started")
}
