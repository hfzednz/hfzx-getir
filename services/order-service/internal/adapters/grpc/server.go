// Package grpcadapter provides a stub gRPC server for order-service.
package grpcadapter

import "log/slog"

// Server is a placeholder gRPC server (proto defined; listener not started in memory mode).
type Server struct {
	Addr string
	Log  *slog.Logger
}

// NewServer returns a stub gRPC server descriptor.
func NewServer(addr string, log *slog.Logger) *Server {
	if log == nil {
		log = slog.Default()
	}
	return &Server{Addr: addr, Log: log}
}

// Start logs that gRPC is stubbed.
func (s *Server) Start() {
	s.Log.Info("grpc.stub", "addr", s.Addr, "note", "proto defined; wire generated stubs later")
}
