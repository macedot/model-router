package handlers

import (
	"github.com/gofiber/fiber/v2"

	"model-router/services"
)

type ModelsHandler struct {
	registry *services.ModelRegistry
}

func NewModelsHandler(registry *services.ModelRegistry) *ModelsHandler {
	return &ModelsHandler{registry: registry}
}

func (h *ModelsHandler) List(c *fiber.Ctx) error {
	modelsList := h.registry.List()

	names := make([]string, len(modelsList))
	for i, m := range modelsList {
		names[i] = m.Name
	}

	return c.JSON(fiber.Map{"models": names})
}
