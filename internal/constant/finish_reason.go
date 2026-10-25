package constant

// FinishReason describes why a completion ended. It is part of the
// OpenAI-compatible response surface.
type FinishReason string

// FinishReason constants
const (
	// FinishStop means the model reached a natural stopping point.
	FinishStop FinishReason = "stop"
	// FinishLength means the output token limit was reached.
	FinishLength FinishReason = "length"
	// FinishToolCalls means the model requested one or more tool calls.
	FinishToolCalls FinishReason = "tool_calls"
	// FinishContentFilter means provider-side safety filtering ended the output.
	FinishContentFilter FinishReason = "content_filter"
	// FinishError means the completion ended because of an error.
	FinishError FinishReason = "error"
)
