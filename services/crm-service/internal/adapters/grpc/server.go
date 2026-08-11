package grpcadapter

import "log/slog"

// Server is a stub gRPC server (codegen later).
type Server struct {
	Addr string
	Log  *slog.Logger
}

// NewServer returns a stub gRPC server.
func NewServer(addr string, log *slog.Logger) *Server {
	return &Server{Addr: addr, Log: log}
}

// Start logs that gRPC is stubbed until proto codegen.
func (s *Server) Start() {
	if s.Log != nil {
		s.Log.Info("grpc.stub", "addr", s.Addr, "note", "proto codegen not wired; REST is primary")
	}
}
