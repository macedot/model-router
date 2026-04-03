package handlers

import (
	"time"

	"github.com/gofiber/fiber/v2"

	"model-router/models"
	"model-router/services"
)

type AnthropicHandler struct {
	registry  services.RegistryReader
	forwarder *services.Forwarder
}

func NewAnthropicHandler(registry services.RegistryReader, forwarder *services.Forwarder) *AnthropicHandler {
	return &AnthropicHandler{
		registry:  registry,
		forwarder: forwarder,
	}
}

func (h *AnthropicHandler) Handle(c *fiber.Ctx) error {
	var req models.AnthropicRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": fiber.Map{
				"type":    "invalid_request_error",
				"message": "Invalid request body: " + err.Error(),
			},
		})
	}

	if req.Model == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": fiber.Map{
				"type":    "invalid_request_error",
				"message": "model is required",
			},
		})
	}

	if len(req.Messages) == 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": fiber.Map{
				"type":    "invalid_request_error",
				"message": "messages is required and cannot be empty",
			},
		})
	}

	modelName := req.Model
	internalModel := h.registry.Get(modelName)
	if internalModel == nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": fiber.Map{
				"type":    "invalid_request_error",
				"message": "Model not found: " + modelName,
			},
		})
	}

	ctx := c.Context()

	var lastErr error
	for i, external := range internalModel.Externals {
		respBody, err := h.forwarder.ForwardAnthropic(ctx, &req, external)
		if err == nil {
			return c.Type("json").Send(respBody)
		}
		lastErr = err

		// Retry delay between externals (skip delay after last one)
		if internalModel.Strategy == models.StrategyFallback &&
			i < len(internalModel.Externals)-1 &&
			internalModel.RetryDelaySecs > 0 {
			time.Sleep(time.Duration(internalModel.RetryDelaySecs) * time.Second)
		}
	}

	return c.Status(fiber.StatusBadGateway).JSON(fiber.Map{
		"error": fiber.Map{
			"type":    "api_error",
			"message": "All providers failed: " + lastErr.Error(),
		},
	})
}
