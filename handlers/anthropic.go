package handlers

import (
	"github.com/gofiber/fiber/v2"

	"model-router/models"
	"model-router/services"
)

type AnthropicHandler struct {
	registry  *services.ModelRegistry
	forwarder *services.Forwarder
}

func NewAnthropicHandler(registry *services.ModelRegistry, forwarder *services.Forwarder) *AnthropicHandler {
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

	respBody, err := h.forwarder.ForwardAnthropic(ctx, &req, internalModel.Externals[0])
	if err != nil {
		return c.Status(fiber.StatusBadGateway).JSON(fiber.Map{
			"error": fiber.Map{
				"type":    "api_error",
				"message": "Forward failed: " + err.Error(),
			},
		})
	}

	return c.Type("json").Send(respBody)
}
