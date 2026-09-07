package http

import (
	"context"
	"data-service/config"
	"errors"
	"net"
	"net/http"
)

type Server struct {
	httpServer *http.Server
}

func NewServer(cfg config.HttpServer, handler http.Handler) *Server {
	return &Server{
		httpServer: &http.Server{
			Addr:         net.JoinHostPort(cfg.Address, cfg.Port),
			Handler:      handler,
			ReadTimeout:  cfg.Timeout,
			WriteTimeout: cfg.Timeout,
			IdleTimeout:  cfg.IdleTimeout,
		},
	}
}

func (s *Server) Run() error {
	if err := s.httpServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		return err
	}
	return nil
}

func (s *Server) Shutdown(ctx context.Context) error {
	return s.httpServer.Shutdown(ctx)
}
