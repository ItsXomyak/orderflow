package httpserver

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"go.temporal.io/sdk/client"

	"orderflow/internal/handlers"
	"orderflow/pkg/logger"
)

type Server struct {
	server          *http.Server
	temporalClient  client.Client
	orderHandler    *handlers.OrderHandler
}

func NewServer(port int, temporalClient client.Client) *Server {
	orderHandler := handlers.NewOrderHandler(temporalClient)
	
	mux := http.NewServeMux()
	
	mux.HandleFunc("/api/orders", orderHandler.CreateOrder)
	mux.HandleFunc("/api/orders/status", orderHandler.GetOrderStatus)
	mux.HandleFunc("/api/orders/cancel", orderHandler.CancelOrder)
	mux.HandleFunc("/api/orders/state", orderHandler.GetWorkflowState)
	
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"ok","timestamp":"` + time.Now().Format(time.RFC3339) + `"}`))
	})

	server := &http.Server{
		Addr:         fmt.Sprintf(":%d", port),
		Handler:      mux,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	return &Server{
		server:         server,
		temporalClient: temporalClient,
		orderHandler:   orderHandler,
	}
}

func (s *Server) Start() error {
	logger.Info("Starting HTTP server", "port", s.server.Addr)
	return s.server.ListenAndServe()
}

func (s *Server) Shutdown(ctx context.Context) error {
	logger.Info("Shutting down HTTP server...")
	return s.server.Shutdown(ctx)
}
