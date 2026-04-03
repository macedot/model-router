package services

import (
	"testing"

	"model-router/models"
)

func TestModelRegistry_Get(t *testing.T) {
	externals := []models.ExternalModel{
		{Name: "provider-1", URL: "https://api.example.com", Format: models.FormatOpenAI},
	}
	registry := NewRegistry([]models.InternalModel{
		{Name: "test-model", RequestFormat: models.FormatOpenAI, Strategy: models.StrategyFallback, Externals: externals},
	})

	tests := []struct {
		name       string
		modelName  string
		wantFound  bool
		wantModel  string
	}{
		{"existing model", "test-model", true, "test-model"},
		{"non-existing model", "unknown", false, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := registry.Get(tt.modelName)
			if tt.wantFound {
				if result == nil {
					t.Fatal("expected model, got nil")
				}
				if result.Name != tt.wantModel {
					t.Errorf("Model = %q, want %q", result.Name, tt.wantModel)
				}
			} else {
				if result != nil {
					t.Errorf("expected nil, got %+v", result)
				}
			}
		})
	}
}

func TestModelRegistry_List(t *testing.T) {
	externals := []models.ExternalModel{
		{Name: "provider-1", URL: "https://api.example.com", Format: models.FormatOpenAI},
	}
	modelsList := []models.InternalModel{
		{Name: "model-1", Externals: externals},
		{Name: "model-2", Externals: externals},
	}
	registry := NewRegistry(modelsList)

	list := registry.List()
	if len(list) != 2 {
		t.Errorf("List() len = %d, want 2", len(list))
	}
	if list[0].Name != "model-1" {
		t.Errorf("List()[0].Name = %q, want %q", list[0].Name, "model-1")
	}
	if list[1].Name != "model-2" {
		t.Errorf("List()[1].Name = %q, want %q", list[1].Name, "model-2")
	}
}

func TestModelRegistry_Get_EmptyRegistry(t *testing.T) {
	registry := NewRegistry([]models.InternalModel{})

	result := registry.Get("any-model")
	if result != nil {
		t.Errorf("expected nil for empty registry, got %+v", result)
	}
}