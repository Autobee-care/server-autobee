// Package server provides the HTTP server lifecycle management.
package server

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"go.uber.org/zap"
)

const (
	readTimeout     = 15 * time.Second
	writeTimeout    = 15 * time.Second
	idleTimeout     = 60 * time.Second
	shutdownTimeout = 30 * time.Second
)

// Server wraps an http.Server with graceful shutdown support.
type Server struct {
	httpServer *http.Server
	log        *zap.Logger
}

// New creates a configured HTTP server.
func New(port string, handler http.Handler, log *zap.Logger) *Server {
	return &Server{
		httpServer: &http.Server{
			Addr:         fmt.Sprintf(":%s", port),
			Handler:      handler,
			ReadTimeout:  readTimeout,
			WriteTimeout: writeTimeout,
			IdleTimeout:  idleTimeout,
		},
		log: log,
	}
}

// Start begins listening for connections.
func (s *Server) Start() error {
	s.log.Info("starting HTTP server", zap.String("addr", s.httpServer.Addr))
	if err := s.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		return fmt.Errorf("http server: %w", err)
	}
	return nil
}

// Shutdown gracefully stops the HTTP server, waiting for active requests to complete.
func (s *Server) Shutdown() error {
	s.log.Info("shutting down HTTP server")
	ctx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()
	return s.httpServer.Shutdown(ctx)
}
