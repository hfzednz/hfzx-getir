// Package grpcadapter provides a gRPC server stub for catalog v1.
package grpcadapter

import (
	"context"
	"fmt"
	"log/slog"
	"net"

	"github.com/nexora/catalog-service/internal/app"
)

// ServerConfig configures the gRPC stub listener.
type ServerConfig struct {
	Addr string
	Deps *app.Deps
	Log  *slog.Logger
}

// Server is a placeholder until protobuf codegen is wired.
type Server struct {
	cfg ServerConfig
	ln  net.Listener
}

// NewServer creates a gRPC stub server.
func NewServer(cfg ServerConfig) *Server {
	if cfg.Log == nil {
		cfg.Log = slog.Default()
	}
	return &Server{cfg: cfg}
}

// ListenAndServe accepts connections and logs stub status.
func (s *Server) ListenAndServe() error {
	ln, err := net.Listen("tcp", s.cfg.Addr)
	if err != nil {
		return err
	}
	s.ln = ln
	s.cfg.Log.Info("grpc.stub.listen", "addr", s.cfg.Addr, "note", "wire proto/catalog/v1 after protoc")
	for {
		conn, err := ln.Accept()
		if err != nil {
			return err
		}
		_ = conn.Close()
	}
}

// Shutdown closes the listener.
func (s *Server) Shutdown(_ context.Context) error {
	if s.ln != nil {
		return s.ln.Close()
	}
	return nil
}

// Health returns stub health for readiness probes.
func (s *Server) Health() error {
	if s.cfg.Addr == "" {
		return fmt.Errorf("grpc addr not configured")
	}
	return nil
}
