package services

import "model-router/models"

type ModelRegistry struct {
	models []models.InternalModel
}

func NewRegistry(modelsList []models.InternalModel) *ModelRegistry {
	return &ModelRegistry{models: modelsList}
}

func (r *ModelRegistry) Get(name string) *models.InternalModel {
	for i := range r.models {
		if r.models[i].Name == name {
			return &r.models[i]
		}
	}
	return nil
}

func (r *ModelRegistry) List() []models.InternalModel {
	return r.models
}
