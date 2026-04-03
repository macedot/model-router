package handlers

import (
	"io"
	"time"

	"github.com/gofiber/fiber/v2"

	"model-router/models"
	"model-router/services"
)

type OpenAIHandler struct {
	registry  services.RegistryReader
	forwarder *services.Forwarder
}

func NewOpenAIHandler(registry services.RegistryReader, forwarder *services.Forwarder) *OpenAIHandler {
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

	if req.Model == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": fiber.Map{
				"message": "model is required",
				"type":    "invalid_request_error",
			},
		})
	}

	if len(req.Messages) == 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": fiber.Map{
				"message": "messages is required and cannot be empty",
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

	// For streaming, use Fiber's SendStream with an io.Pipe so the forwarder
	// writes chunks directly to the client as they arrive from the upstream provider.
	// Fallback retry is not supported for streaming since the response is already
	// in-flight once writing begins.
	if isStream {
		c.Set("Content-Type", "text/event-stream")
		c.Set("Transfer-Encoding", "chunked")

		pipeReader, pipeWriter := io.Pipe()
		go func() {
			err := h.forwarder.ForwardOpenAIStream(ctx, &req, internalModel.Externals[0], pipeWriter)
			if err != nil {
				// Close the pipe with error so the reader sees it
				pipeWriter.CloseWithError(err)
			} else {
				pipeWriter.Close()
			}
		}()

		return c.SendStream(pipeReader)
	}

	// Non-streaming path supports fallback retry
	var lastErr error
	for i, external := range internalModel.Externals {
		var respBody []byte
		var err error

		respBody, err = h.forwarder.ForwardOpenAI(ctx, &req, external)
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
			"message": "All providers failed: " + lastErr.Error(),
			"type":    "api_error",
		},
	})
}
