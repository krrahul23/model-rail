package chat

import (
	"strings"
	"time"

	"modelrail/internal/constant"
	dtoresponse "modelrail/internal/dto/response"
)

// Response is the supported subset of the OpenAI chat completion response.
type Response struct {
	ID      string   `json:"id"`
	Object  string   `json:"object"`
	Created int64    `json:"created"`
	Model   string   `json:"model"`
	Choices []choice `json:"choices"`
	Usage   usage    `json:"usage"`
}

// choice is one completion choice.
type choice struct {
	Index        int              `json:"index"`
	Message      assistantMessage `json:"message"`
	FinishReason string           `json:"finish_reason"`
}

// assistantMessage is the assistant message returned to the client.
type assistantMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// usage mirrors the OpenAI usage object.
type usage struct {
	PromptTokens     int64 `json:"prompt_tokens"`
	CompletionTokens int64 `json:"completion_tokens"`
	TotalTokens      int64 `json:"total_tokens"`
}

// FromResult maps the canonical result onto the v1 response object.
func FromResult(requestID string, result dtoresponse.Result) (payload Response) {
	var text strings.Builder
	for _, part := range result.Content {
		if part.Type == constant.ContentText {
			text.WriteString(part.Text)
		}
	}

	payload = Response{
		ID:      requestID,
		Object:  constant.ObjectChatCompletion,
		Created: time.Now().Unix(),
		Model:   string(result.ModelServed),
		Choices: []choice{{
			Index: 0,
			Message: assistantMessage{
				Role:    string(constant.RoleAssistant),
				Content: text.String(),
			},
			FinishReason: string(result.FinishReason),
		}},
		Usage: usage{
			PromptTokens:     result.Usage.InputTokens,
			CompletionTokens: result.Usage.OutputTokens,
			TotalTokens:      result.Usage.InputTokens + result.Usage.OutputTokens,
		},
	}
	return
}
