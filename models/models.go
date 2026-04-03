package models

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
	Name   string          `json:"name"`
	URL    string          `json:"url"`
	APIKey string          `json:"api_key"`
	Format RequestFormat   `json:"format"`
}

type InternalModel struct {
	Name            string          `json:"name"`
	RequestFormat   RequestFormat   `json:"request_format"`
	Strategy        Strategy        `json:"strategy"`
	RetryDelaySecs  uint32          `json:"retry_delay_secs"`
	Externals       []ExternalModel `json:"externals"`
}

type Config struct {
	Port   uint16           `json:"port"`
	Models []InternalModel  `json:"models"`
}

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type OpenAIRequest struct {
	Model       string    `json:"model"`
	Messages    []Message `json:"messages"`
	Temperature *float64  `json:"temperature,omitempty"`
	MaxTokens   *int      `json:"max_tokens,omitempty"`
	Stream      *bool     `json:"stream,omitempty"`
}

type AnthropicRequest struct {
	Model       string    `json:"model"`
	Messages    []Message `json:"messages"`
	MaxTokens   int       `json:"max_tokens"`
	Temperature *float64  `json:"temperature,omitempty"`
}
