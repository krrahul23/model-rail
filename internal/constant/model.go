package constant

// ModelRef identifies a model as exposed to clients, for example
// "deepseek-chat" or "auto" before routing resolves it to a concrete
// provider and model pair.
type ModelRef string

// TaskClass labels the kind of work a request performs. It is supplied by the
// client through the modelrail-task hint and defaults to TaskGeneral; routing
// uses it to look up evaluation evidence for that class.
type TaskClass string

// TaskClass constants
const (
	// TaskGeneral is the default class when no hint is supplied.
	TaskGeneral TaskClass = "general"
	// TaskCoding marks code generation and code reasoning workloads.
	TaskCoding TaskClass = "coding"
	// TaskReasoning marks multi-step reasoning workloads.
	TaskReasoning TaskClass = "reasoning"
	// TaskStructuredExtraction marks schema-constrained extraction workloads.
	TaskStructuredExtraction TaskClass = "structured_extraction"
)
