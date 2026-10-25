// Package request defines the canonical, provider-neutral request DTOs shared
// by the controller, application and domain layers.
package request

import (
	"time"

	"modelrail/internal/constant"
)

// ContentPart is one piece of a message payload. Today only text is supported;
// additional content types are added when the corresponding capability ships.
type ContentPart struct {
	Type constant.ContentType
	Text string
}

// Message is one turn in a conversation, authored by a role and made of one or
// more content parts.
type Message struct {
	Role  constant.Role
	Parts []ContentPart
}

// ToolSpec describes a tool the model may call. Parameters is the JSON schema
// of the tool's arguments.
type ToolSpec struct {
	Name        string
	Description string
	Parameters  []byte
}

// FormatSpec requests structured output: Type names the format (for example a
// JSON schema mode) and Schema carries the expected document shape.
type FormatSpec struct {
	Type   string
	Schema []byte
}

// CanonicalRequest is the normalized inference request. The controller maps the
// OpenAI wire format into this type, and provider adapters map it into their
// own dialect.
type CanonicalRequest struct {
	Model           constant.ModelRef  // requested alias, or "auto" for routing
	TaskClass       constant.TaskClass // routing class; defaults to TaskGeneral
	Messages        []Message          // conversation, oldest first
	Stream          bool               // selects SSE streaming instead of unary
	MaxOutputTokens int                // clamped by policy before reservation
	Temperature     *float64           // nil keeps the provider default
	Tools           []ToolSpec         // empty means tools are disabled
	ResponseFormat  *FormatSpec        // non-nil requests structured output
	TotalDeadline   time.Duration      // parent deadline for every phase
}
