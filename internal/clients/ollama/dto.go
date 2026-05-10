package ollama

import "time"

type GenerateRequest struct {
	Model   string                  `json:"model"`
	Prompt  string                  `json:"prompt"`
	Stream  bool                    `json:"stream"`
	Options *map[string]interface{} `json:"options,omitempty"`
}

type GenerateResponse struct {
	Model     string    `json:"model"`
	CreatedAt time.Time `json:"created_at"`
	Response  string    `json:"response"`
	Done      bool      `json:"done"`
}

type PostGenerateOut struct {
	Response string
}
