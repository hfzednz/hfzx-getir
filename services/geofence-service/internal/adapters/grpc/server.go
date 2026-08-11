package grpcadapter

import "log/slog"

// Server is a gRPC stub placeholder for geofence RPCs.
type Server struct {
	Addr string
	Log  *slog.Logger
}

// NewServer returns a gRPC stub (not listening until wired).
func NewServer(addr string, log *slog.Logger) *Server {
	if log == nil {
		log = slog.Default()
	}
	return &Server{Addr: addr, Log: log}
}

// Start logs that the gRPC stub is ready (no real listener in memory mode).
func (s *Server) Start() {
	s.Log.Info("grpc.stub", "addr", s.Addr, "note", "proto defined; listener not started in stub mode")
}
