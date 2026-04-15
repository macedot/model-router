package services

import (
	"encoding/json"
	"log"

	"model-router/models"
)

// Field renames between formats.
var openAIToAnthropicRenames = map[string]string{
	"stop":                 "stop_sequences",
	"max_completion_tokens": "max_tokens",
}

var anthropicToOpenAIRenames = map[string]string{
	"stop_sequences": "stop",
}

// Fields that exist in one format but have no equivalent in the other.
var openAIOnlyFields = map[string]bool{
	"frequency_penalty": true,
	"presence_penalty":  true,
	"n":                 true,
	"logprobs":          true,
	"top_logprobs":      true,
	"response_format":   true,
	"seed":              true,
	"service_tier":      true,
	"reasoning_effort":  true,
}

var anthropicOnlyFields = map[string]bool{
	"top_k":    true,
	"thinking": true,
}

// PrepareRequest prepares a request body for forwarding to a target provider.
// It sets the model name and applies format conversion if needed.
// Returns marshaled JSON ready to send.
func PrepareRequest(body map[string]interface{}, modelName string, source, target models.RequestFormat) ([]byte, error) {
	body["model"] = modelName

	if source == target {
		return json.Marshal(body)
	}

	return convertFormat(body, source, target)
}

func convertFormat(body map[string]interface{}, source, target models.RequestFormat) ([]byte, error) {
	var renames map[string]string
	var sourceOnly map[string]bool

	if source == models.FormatOpenAI && target == models.FormatAnthropic {
		renames = openAIToAnthropicRenames
		sourceOnly = openAIOnlyFields
	} else {
		renames = anthropicToOpenAIRenames
		sourceOnly = anthropicOnlyFields
	}

	result := make(map[string]interface{}, len(body))

	for key, value := range body {
		if sourceOnly[key] {
			log.Printf("warning: field %q has no equivalent in %s format, dropping", key, target)
			continue
		}
		if newKey, ok := renames[key]; ok {
			result[newKey] = value
			continue
		}
		result[key] = value
	}

	return json.Marshal(result)
}
