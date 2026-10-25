package response

import (
	"modelrail/internal/constant"
	"modelrail/internal/dto/request"
)

// StreamEvent is a sealed event emitted by a provider stream. A stream always
// terminates with DoneEvent or ErrorEvent, after which the adapter closes the
// event channel; no events follow the terminal one.
type StreamEvent interface {
	isStreamEvent()
}

// Chunk carries an incremental piece of content. Chunks are the only events
// that advance the client-visible response, and the first one defines the
// streaming commit point.
type Chunk struct {
	Part request.ContentPart
}

// isStreamEvent is a marker method to indicate that Chunk implements the StreamEvent interface.
func (Chunk) isStreamEvent() {}

// UsageEvent carries the final provider-reported usage. It is emitted once,
// immediately before the terminal event, and is the input to settlement.
type UsageEvent struct {
	Usage Usage
}

// isStreamEvent is a marker method to indicate that UsageEvent implements the StreamEvent interface.
func (UsageEvent) isStreamEvent() {}

// DoneEvent is the successful terminal event and carries the finish reason.
type DoneEvent struct {
	FinishReason constant.FinishReason
}

// isStreamEvent is a marker method to indicate that DoneEvent implements the StreamEvent interface.
func (DoneEvent) isStreamEvent() {}

// ErrorEvent is the failure terminal event. After a commit it is the only
// honest way to end the stream; the adapter never switches providers.
type ErrorEvent struct {
	Err ProviderError
}

// isStreamEvent is a marker method to indicate that ErrorEvent implements the StreamEvent interface.
func (ErrorEvent) isStreamEvent() {}

// The following lines ensure that the types Chunk, UsageEvent, DoneEvent, and ErrorEvent implement the StreamEvent interface.
var (
	_ StreamEvent = Chunk{}
	_ StreamEvent = UsageEvent{}
	_ StreamEvent = DoneEvent{}
	_ StreamEvent = ErrorEvent{}
)
