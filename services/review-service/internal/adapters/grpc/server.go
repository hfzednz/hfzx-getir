package grpcadapter

import "log/slog"

// Server is a gRPC stub placeholder for review RPCs.
type Server struct {
	addr string
	log  *slog.Logger
}

// NewServer creates a gRPC server stub.
func NewServer(addr string, log *slog.Logger) *Server {
	if log == nil {
		log = slog.Default()
	}
	return &Server{addr: addr, log: log}
}

// Start logs that gRPC is reserved (wire generated stubs in CI).
func (s *Server) Start() {
	s.log.Info("grpc.stub", "addr", s.addr, "note", "proto generated server not linked in memory mode")
}
