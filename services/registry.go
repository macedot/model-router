package services

import "model-router/models"

// ModelRegistry provides read-only access to registered models and providers.
type ModelRegistry struct {
	models    []models.InternalModel
	providers []models.Provider
}

// RegistryReader defines the read-only interface for model registry access.
// Handlers should accept this interface to avoid modifying internal state.
type RegistryReader interface {
	Get(name string) *models.InternalModel
	List() []models.InternalModel
	GetProvider(id string) *models.Provider
	ListProviders() []models.Provider
}

func NewRegistry(modelsList []models.InternalModel, providers []models.Provider) *ModelRegistry {
	providersCopy := make([]models.Provider, len(providers))
	copy(providersCopy, providers)

	result := make([]models.InternalModel, len(modelsList))
	for i := range modelsList {
		externalsCopy := make([]models.ExternalModel, len(modelsList[i].Externals))
		copy(externalsCopy, modelsList[i].Externals)
		result[i] = modelsList[i]
		result[i].Externals = externalsCopy
	}
	return &ModelRegistry{models: result, providers: providersCopy}
}

func (r *ModelRegistry) Get(name string) *models.InternalModel {
	for i := range r.models {
		if r.models[i].Name == name {
			m := r.models[i]
			externalsCopy := make([]models.ExternalModel, len(m.Externals))
			copy(externalsCopy, m.Externals)
			m.Externals = externalsCopy
			return &m
		}
	}
	return nil
}

func (r *ModelRegistry) List() []models.InternalModel {
	out := make([]models.InternalModel, len(r.models))
	for i := range r.models {
		externalsCopy := make([]models.ExternalModel, len(r.models[i].Externals))
		copy(externalsCopy, r.models[i].Externals)
		out[i] = models.InternalModel{
			Name:           r.models[i].Name,
			RequestFormat: r.models[i].RequestFormat,
			Strategy:      r.models[i].Strategy,
			RetryDelaySecs: r.models[i].RetryDelaySecs,
			Externals:     externalsCopy,
		}
	}
	return out
}

func (r *ModelRegistry) GetProvider(id string) *models.Provider {
	for i := range r.providers {
		if r.providers[i].ID == id {
			p := r.providers[i]
			return &p
		}
	}
	return nil
}

func (r *ModelRegistry) ListProviders() []models.Provider {
	out := make([]models.Provider, len(r.providers))
	copy(out, r.providers)
	return out
}

// Ensure ModelRegistry implements RegistryReader.
var _ RegistryReader = (*ModelRegistry)(nil)
