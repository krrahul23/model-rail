package mock

import (
	"context"
	"sync"

	"modelrail/internal/constant"
	"modelrail/internal/dto/request"
	"modelrail/internal/dto/response"
	"modelrail/internal/ports"
)

// Compile-time assertion that the stream satisfies the provider stream port.
var _ ports.ProviderStream = (*stream)(nil)

// stream is a single-consumer event stream. The producer goroutine owns and
// closes the events channel exactly once; Close cancels the producer and is
// safe to call repeatedly or concurrently.
type stream struct {
	events    chan response.StreamEvent
	usage     response.Usage
	cfg       Config
	cancel    context.CancelFunc
	closeOnce sync.Once
	closed    chan struct{}
}

// newStream derives a cancellable context from ctx and starts the producer.
func newStream(ctx context.Context, cfg Config) (s *stream) {
	streamCtx, cancel := context.WithCancel(ctx)
	s = &stream{
		events: make(chan response.StreamEvent),
		usage:  cfg.effectiveUsage(),
		cfg:    cfg,
		cancel: cancel,
		closed: make(chan struct{}),
	}
	go s.run(streamCtx)
	return
}

// Events returns the single-consumer event channel. The adapter closes it after
// the terminal event or cancellation; the consumer must never close it.
func (s *stream) Events() (events <-chan response.StreamEvent) {
	events = s.events
	return
}

// Close stops the producer and releases its resources. It is idempotent and
// non-blocking, and safe to call while another goroutine reads the channel.
func (s *stream) Close() (err error) {
	s.closeOnce.Do(func() {
		close(s.closed)
		s.cancel()
	})
	return
}

// run produces the configured events in order: chunks, optional error, usage,
// then the terminal done event. Every send is cancellable, so a consumer that
// stops reading or calls Close cannot block the producer indefinitely.
func (s *stream) run(ctx context.Context) {
	defer close(s.events)
	defer s.cancel()

	// Simulate time to first token and stop early on cancellation.
	if waitErr := wait(ctx, s.cfg.TTFT); waitErr != nil {
		return
	}

	chunks := splitRunes(s.cfg.ResponseText, s.cfg.ChunkSize)

	// Emit chunks, inserting the configured error after the requested count.
	for i := 0; i <= len(chunks); i++ {
		if s.cfg.StreamError != nil && s.cfg.ErrorAfterChunks == i {
			s.emit(ctx, response.ErrorEvent{Err: *s.cfg.StreamError})
			return
		}
		if i == len(chunks) {
			break
		}
		if !s.emit(ctx, response.Chunk{Part: request.ContentPart{
			Type: constant.ContentText,
			Text: chunks[i],
		}}) {
			return
		}
		if waitErr := wait(ctx, s.cfg.ChunkDelay); waitErr != nil {
			return
		}
	}

	// Usage is emitted before the terminal event so settlement can use it.
	if !s.emit(ctx, response.UsageEvent{Usage: s.usage}) {
		return
	}
	s.emit(ctx, response.DoneEvent{FinishReason: constant.FinishStop})
}

// emit attempts to deliver one event, returning false when the stream is closed
// or the context has ended. It never blocks indefinitely.
func (s *stream) emit(ctx context.Context, ev response.StreamEvent) (ok bool) {
	select {
	case <-s.closed:
		return
	case <-ctx.Done():
		return
	default:
	}

	select {
	case s.events <- ev:
		ok = true
	case <-s.closed:
	case <-ctx.Done():
	}
	return
}

// splitRunes splits text into size-rune chunks. Working in runes avoids
// breaking multi-byte characters across chunks.
func splitRunes(text string, size int) (parts []string) {
	if size <= 0 {
		size = constant.DefaultChunkSize
	}
	runes := []rune(text)
	for len(runes) > 0 {
		n := size
		if n > len(runes) {
			n = len(runes)
		}
		parts = append(parts, string(runes[:n]))
		runes = runes[n:]
	}
	return
}
