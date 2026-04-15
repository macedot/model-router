package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"model-router/config"
	"model-router/handlers"
	"model-router/services"
)

// loggingMiddleware wraps an http.Handler to log requests.
type loggingMiddleware struct {
	handler http.Handler
}

func (m *loggingMiddleware) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	log.Printf("%s %s", r.Method, r.URL.Path)
	m.handler.ServeHTTP(w, r)
}

// recoverMiddleware catches panics and returns 500.
type recoverMiddleware struct {
	handler http.Handler
}

func (m *recoverMiddleware) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	defer func() {
		if err := recover(); err != nil {
			log.Printf("panic: %v", err)
			http.Error(w, `{"error":{"message":"Internal server error"}}`, http.StatusInternalServerError)
		}
	}()
	m.handler.ServeHTTP(w, r)
}

func main() {
	log.Printf("model-router v%s starting...", GetVersion())
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Check port availability
	addr := fmt.Sprintf(":%d", cfg.Port)
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		log.Fatalf("Port %d is already in use", cfg.Port)
	}
	ln.Close()

	// Initialize services
	registry := services.NewRegistry(cfg.Models)
	forwarder := services.NewForwarder()

	// Initialize handlers
	modelsHandler := handlers.NewModelsHandler(registry)
	openaiHandler := handlers.NewOpenAIHandler(registry, forwarder)
	anthropicHandler := handlers.NewAnthropicHandler(registry, forwarder)

	// Create server with timeouts from config
	server := &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.Port),
		Handler:      nil, // use default mux
		ReadTimeout:  config.Defaults.ReadTimeout,
		WriteTimeout: config.Defaults.WriteTimeout,
	}

	// Register routes on default mux with middleware
	mux := http.NewServeMux()
	mux.HandleFunc("/models", modelsHandler)
	mux.HandleFunc("/v1/chat/completions", openaiHandler)
	mux.HandleFunc("/v1/messages", anthropicHandler)

	// Wrap with middleware (recover first, then logging)
	server.Handler = &recoverMiddleware{handler: &loggingMiddleware{handler: mux}}

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	errCh := make(chan error, 1)

	go func() {
		log.Printf("Listening on %s", server.Addr)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			errCh <- err
		}
	}()

	select {
	case <-quit:
		log.Println("Shutting down server...")
	case err := <-errCh:
		log.Printf("Server error: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), config.Defaults.ShutdownTimeout)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Server exited")
}
