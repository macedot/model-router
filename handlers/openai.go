package handlers

import (
	"github.com/gofiber/fiber/v2"

	"model-router/models"
	"model-router/services"
)

type OpenAIHandler struct {
	registry  *services.ModelRegistry
	forwarder *services.Forwarder
}

func NewOpenAIHandler(registry *services.ModelRegistry, forwarder *services.Forwarder) *OpenAIHandler {
	return &OpenAIHandler{
		registry:  registry,
		forwarder: forwarder,
	}
}

func (h *OpenAIHandler) Handle(c *fiber.Ctx) error {
	var req models.OpenAIRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": fiber.Map{
				"message": "Invalid request body: " + err.Error(),
				"type":    "invalid_request_error",
			},
		})
	}

	modelName := req.Model
	internalModel := h.registry.Get(modelName)
	if internalModel == nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": fiber.Map{
				"message": "Model not found: " + modelName,
				"type":    "invalid_request_error",
			},
		})
	}

	ctx := c.Context()

	// Check if streaming
	isStream := req.Stream != nil && *req.Stream

	if isStream {
		respBody, err := h.forwarder.ForwardOpenAIStream(ctx, &req, internalModel.Externals[0])
		if err != nil {
			return c.Status(fiber.StatusBadGateway).JSON(fiber.Map{
				"error": fiber.Map{
					"message": "Forward failed: " + err.Error(),
					"type":    "api_error",
				},
			})
		}
		return c.Type("json").Send(respBody)
	}

	respBody, err := h.forwarder.ForwardOpenAI(ctx, &req, internalModel.Externals[0])
	if err != nil {
		return c.Status(fiber.StatusBadGateway).JSON(fiber.Map{
			"error": fiber.Map{
				"message": "Forward failed: " + err.Error(),
				"type":    "api_error",
			},
		})
	}

	return c.Type("json").Send(respBody)
}
