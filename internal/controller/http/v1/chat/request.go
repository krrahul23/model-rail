// Package chat owns API version 1 chat completion requests and responses.
package chat

import (
	"encoding/json"
	"fmt"

	"modelrail/internal/constant"
	dtorequest "modelrail/internal/dto/request"
)

// Request is the supported subset of the OpenAI chat completion request.
type Request struct {
	Model       string    `json:"model"`
	Messages    []message `json:"messages"`
	MaxTokens   *int      `json:"max_tokens,omitempty"`
	Temperature *float64  `json:"temperature,omitempty"`
	Stream      bool      `json:"stream,omitempty"`
}

// message is one message in the wire format.
type message struct {
	Role    string  `json:"role"`
	Content content `json:"content"`
}

// content accepts either a plain string or an array of text parts.
type content struct {
	parts []dtorequest.ContentPart
}

// UnmarshalJSON decodes the two content shapes supported by the API.
func (c *content) UnmarshalJSON(data []byte) (err error) {
	var text string
	if err = json.Unmarshal(data, &text); err == nil {
		c.parts = []dtorequest.ContentPart{{Type: constant.ContentText, Text: text}}
		return
	}

	var parts []wireContentPart
	if err = json.Unmarshal(data, &parts); err != nil {
		err = fmt.Errorf("content must be a string or an array of text parts")
		return
	}
	for _, part := range parts {
		if part.Type != string(constant.ContentText) {
			err = fmt.Errorf("unsupported content part type %q", part.Type)
			return
		}
		c.parts = append(c.parts, dtorequest.ContentPart{Type: constant.ContentText, Text: part.Text})
	}
	return
}

// wireContentPart is one entry of an array-shaped content field.
type wireContentPart struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

// ToDTO converts the wire request into the canonical request DTO. It performs
// structural mapping only; business validation belongs to the application layer.
func (r Request) ToDTO() (req dtorequest.CanonicalRequest, err error) {
	req = dtorequest.CanonicalRequest{
		Model:     constant.ModelRef(r.Model),
		TaskClass: constant.TaskGeneral,
		Stream:    r.Stream,
	}
	if r.MaxTokens != nil {
		req.MaxOutputTokens = *r.MaxTokens
	}
	req.Temperature = r.Temperature

	for index, entry := range r.Messages {
		if len(entry.Content.parts) == 0 {
			err = fmt.Errorf("messages[%d] has no content", index)
			return
		}
		req.Messages = append(req.Messages, dtorequest.Message{
			Role:  constant.Role(entry.Role),
			Parts: entry.Content.parts,
		})
	}
	return
}
