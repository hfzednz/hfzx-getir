package grpcadapter

import "log/slog"

// Server is a gRPC stub placeholder (proto defined; HTTP is primary).
type Server struct {
	log *slog.Logger
}

// NewServer returns a gRPC server stub.
func NewServer(log *slog.Logger) *Server {
	if log == nil {
		log = slog.Default()
	}
	return &Server{log: log}
}
