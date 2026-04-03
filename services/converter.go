package services

import (
	"model-router/models"
)

const defaultMaxTokens = 4096

func ToAnthropic(req *models.OpenAIRequest) *models.AnthropicRequest {
	anthropicReq := &models.AnthropicRequest{
		Model:       req.Model,
		Messages:    req.Messages,
		MaxTokens:   defaultMaxTokens,
		Temperature: req.Temperature,
	}
	if req.MaxTokens != nil {
		anthropicReq.MaxTokens = *req.MaxTokens
	}
	return anthropicReq
}

