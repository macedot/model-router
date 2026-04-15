package models

import "encoding/json"

type RequestFormat string

const (
	FormatOpenAI    RequestFormat = "openai"
	FormatAnthropic RequestFormat = "anthropic"
)

type Strategy string

const (
	StrategyFallback Strategy = "fallback"
)

type ExternalModel struct {
	ID     string        `json:"id,omitempty"`
	Name   string        `json:"name"`
	URL    string        `json:"url"`
	APIKey string        `json:"api_key"`
	Format RequestFormat `json:"format"`
}

type Provider struct {
	ID     string        `json:"id"`
	Name   string        `json:"name"`
	URL    string        `json:"url"`
	APIKey string        `json:"api_key"`
	Format RequestFormat `json:"format"`
}

func (p Provider) ToExternal() ExternalModel {
	return ExternalModel{
		ID:     p.ID,
		Name:   p.Name,
		URL:    p.URL,
		APIKey: p.APIKey,
		Format: p.Format,
	}
}

type InternalModel struct {
	Name           string          `json:"name"`
	RequestFormat  RequestFormat   `json:"request_format"`
	Strategy       Strategy        `json:"strategy"`
	RetryDelaySecs uint32          `json:"retry_delay_secs"`
	Externals      []ExternalModel `json:"externals"`
}

type Config struct {
	Port      uint16          `json:"port"`
	Providers []Provider      `json:"providers"`
	Models    []InternalModel `json:"models"`
}

// RequestEnvelope extracts only the fields needed for validation/routing.
// The full request body is handled as map[string]interface{} for generic passthrough.
type RequestEnvelope struct {
	Model    string          `json:"model"`
	Messages json.RawMessage `json:"messages"`
	Stream   *bool           `json:"stream,omitempty"`
}

// HasMessages returns true if the messages field is present and non-empty.
func (e *RequestEnvelope) HasMessages() bool {
	if len(e.Messages) == 0 {
		return false
	}
	var arr []interface{}
	if err := json.Unmarshal(e.Messages, &arr); err != nil {
		return false
	}
	return len(arr) > 0
}
