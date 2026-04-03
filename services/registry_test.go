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

func TestModelRegistry_List_ReturnsCopy(t *testing.T) {
	originals := []models.InternalModel{
		{Name: "model-1", Externals: []models.ExternalModel{{Name: "ext-1"}}},
		{Name: "model-2", Externals: []models.ExternalModel{{Name: "ext-2"}}},
	}
	registry := NewRegistry(originals)

	list := registry.List()
	list[0].Name = "modified"
	list[1].Externals[0].Name = "modified-ext"

	// Verify original is unchanged via fresh List call
	fresh := registry.List()
	if fresh[0].Name != "model-1" {
		t.Errorf("List() returned copy but original was modified")
	}
	if fresh[1].Externals[0].Name != "ext-2" {
		t.Errorf("List() nested externals were modified")
	}
}

func TestModelRegistry_Get_ReturnsCopy(t *testing.T) {
	originals := []models.InternalModel{
		{Name: "model-1", Externals: []models.ExternalModel{{Name: "ext-1"}}},
	}
	registry := NewRegistry(originals)

	model := registry.Get("model-1")
	model.Name = "modified"
	model.Externals[0].Name = "modified-ext"

	// Verify original is unchanged via another Get call
	fresh := registry.Get("model-1")
	if fresh.Name != "model-1" {
		t.Errorf("Get() returned copy but original Name was modified")
	}
	if fresh.Externals[0].Name != "ext-1" {
		t.Errorf("Get() returned copy but original Externals were modified")
	}
}