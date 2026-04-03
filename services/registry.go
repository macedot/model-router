package services

import "model-router/models"

// ModelRegistry provides read-only access to registered models.
type ModelRegistry struct {
	models []models.InternalModel
}

// RegistryReader defines the read-only interface for model registry access.
// Handlers should accept this interface to avoid modifying internal state.
type RegistryReader interface {
	Get(name string) *models.InternalModel
	List() []models.InternalModel
}

func NewRegistry(modelsList []models.InternalModel) *ModelRegistry {
	return &ModelRegistry{models: modelsList}
}

func (r *ModelRegistry) Get(name string) *models.InternalModel {
	for i := range r.models {
		if r.models[i].Name == name {
			// Return a copy to prevent callers from modifying internal state.
			model := r.models[i]
			return &model
		}
	}
	return nil
}

func (r *ModelRegistry) List() []models.InternalModel {
	// Return a copy to prevent callers from modifying internal state.
	out := make([]models.InternalModel, len(r.models))
	copy(out, r.models)
	return out
}

// Ensure ModelRegistry implements RegistryReader.
var _ RegistryReader = (*ModelRegistry)(nil)
