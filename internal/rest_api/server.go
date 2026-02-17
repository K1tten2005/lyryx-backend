package rest_api

import (
	"fmt"
	"net/http"
	"time"
)

type Server struct {
	httpServer *http.Server
}

func NewServer(handler http.Handler) *Server {
	return &Server{
		httpServer: &http.Server{
			ReadHeaderTimeout: time.Second,
			Addr:              ":8080",
			Handler:           handler,
		},
	}
}

func (s *Server) Start() error {
	if err := s.httpServer.ListenAndServe(); err != nil {
		return fmt.Errorf("server start failed: %v", err)
	}

	return nil
}
