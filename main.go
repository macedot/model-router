package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"

	"model-router/config"
	"model-router/handlers"
	"model-router/services"
)

func main() {
	log.Printf("model-router v%s starting...", FullVersion)
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Initialize services
	registry := services.NewRegistry(cfg.Models)
	forwarder := services.NewForwarder()

	// Initialize handlers
	modelsHandler := handlers.NewModelsHandler(registry)
	openaiHandler := handlers.NewOpenAIHandler(registry, forwarder)
	anthropicHandler := handlers.NewAnthropicHandler(registry, forwarder)

	// Create Fiber app
	app := fiber.New(fiber.Config{
		ReadTimeout:  config.Defaults.ReadTimeout,
		WriteTimeout: config.Defaults.WriteTimeout,
		BodyLimit:    config.Defaults.BodyLimit,
	})

	// Middleware
	app.Use(recover.New())
	app.Use(logger.New())

	// Routes
	app.Get("/models", modelsHandler.List)
	app.Post("/v1/chat/completions", openaiHandler.Handle)
	app.Post("/v1/messages", anthropicHandler.Handle)

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		addr := fmt.Sprintf(":%d", cfg.Port)
		if err := app.Listen(addr); err != nil {
			log.Fatalf("Server error: %v", err)
		}
	}()

	<-quit
	log.Println("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), config.Defaults.ShutdownTimeout)
	defer cancel()

	if err := app.ShutdownWithContext(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Server exited")
}
