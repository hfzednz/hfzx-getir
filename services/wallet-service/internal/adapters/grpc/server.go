package grpcadapter

import "log/slog"

// Server is a placeholder for future gRPC wiring from proto/wallet/v1.
type Server struct {
	Log *slog.Logger
}

// ListenAndServe is a no-op stub until protobuf codegen is wired.
func (s *Server) ListenAndServe(addr string) error {
	if s.Log != nil {
		s.Log.Info("grpc.stub", "addr", addr, "note", "proto defined; listener not started")
	}
	return nil
}
