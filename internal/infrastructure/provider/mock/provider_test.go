package mock_test

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"modelrail/internal/constant"
	"modelrail/internal/dto/request"
	"modelrail/internal/dto/response"
	"modelrail/internal/infrastructure/provider/mock"
	"modelrail/internal/ports"
)

const testModel = constant.ModelRef("mock-model")

// Compile-time assertion that the exported type satisfies the provider port.
var _ ports.Provider = (*mock.Provider)(nil)

// durationPtr returns a pointer to d for optional duration fields.
func durationPtr(d time.Duration) (ptr *time.Duration) {
	ptr = &d
	return
}

// collect drains a stream until it closes or the timeout expires.
func collect(t *testing.T, events <-chan response.StreamEvent, timeout time.Duration) (collected []response.StreamEvent) {
	t.Helper()
	timer := time.NewTimer(timeout)
	defer timer.Stop()

	for {
		select {
		case ev, ok := <-events:
			if !ok {
				return
			}
			collected = append(collected, ev)
		case <-timer.C:
			t.Fatalf("stream did not close within %s", timeout)
			return
		}
	}
}

func TestProviderMetadata(t *testing.T) {
	t.Parallel()

	provider := mock.New(mock.Config{ProviderName: "mock-test"})

	if name := provider.Name(); name != "mock-test" {
		t.Fatalf("Name() = %q, want %q", name, "mock-test")
	}

	caps := provider.Capabilities(testModel)
	if !caps.SupportsStreaming {
		t.Fatal("Capabilities() must advertise streaming support")
	}
	if caps.ContextWindow == 0 {
		t.Fatal("Capabilities() must advertise a non-zero context window")
	}
}

func TestInvokeReturnsNormalizedResult(t *testing.T) {
	t.Parallel()

	cfg := mock.Config{
		ProviderName: "mock-test",
		ResponseText: "hello from the mock",
		Latency:      5 * time.Millisecond,
		Usage: response.Usage{
			InputTokens:  7,
			OutputTokens: 4,
			Source:       constant.UsageReported,
		},
	}
	provider := mock.New(cfg)

	result, err := provider.Invoke(t.Context(), testModel, request.CanonicalRequest{Model: testModel})
	if err != nil {
		t.Fatalf("Invoke() error = %v", err)
	}

	if len(result.Content) != 1 {
		t.Fatalf("len(Content) = %d, want 1", len(result.Content))
	}
	if got := result.Content[0].Text; got != cfg.ResponseText {
		t.Fatalf("Content[0].Text = %q, want %q", got, cfg.ResponseText)
	}
	if result.Usage != cfg.Usage {
		t.Fatalf("Usage = %+v, want %+v", result.Usage, cfg.Usage)
	}
	if result.ModelServed != testModel {
		t.Fatalf("ModelServed = %q, want %q", result.ModelServed, testModel)
	}
	if result.ProviderName != cfg.ProviderName {
		t.Fatalf("ProviderName = %q, want %q", result.ProviderName, cfg.ProviderName)
	}
	if result.FinishReason != constant.FinishStop {
		t.Fatalf("FinishReason = %q, want %q", result.FinishReason, constant.FinishStop)
	}
	if result.Latency.Total < cfg.Latency {
		t.Fatalf("Latency.Total = %s, want at least %s", result.Latency.Total, cfg.Latency)
	}
}

func TestInvokeReturnsConfiguredError(t *testing.T) {
	t.Parallel()

	retryAfter := 2 * time.Second
	providerErr := response.ProviderError{
		Class:        constant.ClassThrottled,
		Provider:     "mock",
		ProviderCode: "429",
		RetryAfter:   durationPtr(retryAfter),
	}
	provider := mock.New(mock.Config{InvokeError: &providerErr})

	_, err := provider.Invoke(t.Context(), testModel, request.CanonicalRequest{})
	var got response.ProviderError
	if !errors.As(err, &got) {
		t.Fatalf("Invoke() error = %v, want response.ProviderError", err)
	}
	if got.Class != constant.ClassThrottled {
		t.Fatalf("Class = %v, want %v", got.Class, constant.ClassThrottled)
	}
	if got.ProviderCode != "429" {
		t.Fatalf("ProviderCode = %q, want %q", got.ProviderCode, "429")
	}
	if got.RetryAfter == nil || *got.RetryAfter != retryAfter {
		t.Fatalf("RetryAfter = %v, want %s", got.RetryAfter, retryAfter)
	}
}

func TestInvokeHonoursContextCancellation(t *testing.T) {
	t.Parallel()

	provider := mock.New(mock.Config{Latency: time.Second})
	ctx, cancel := context.WithCancel(t.Context())
	cancel()

	_, err := provider.Invoke(ctx, testModel, request.CanonicalRequest{})

	var got response.ProviderError
	if !errors.As(err, &got) {
		t.Fatalf("Invoke() error = %v, want response.ProviderError", err)
	}
	if got.Class != constant.ClassUnknown {
		t.Fatalf("Class = %v, want %v", got.Class, constant.ClassUnknown)
	}
	if got.ProviderCode != "context_canceled" {
		t.Fatalf("ProviderCode = %q, want %q", got.ProviderCode, "context_canceled")
	}
}

func TestInvokeClassifiesDeadlineAsTransient(t *testing.T) {
	t.Parallel()

	provider := mock.New(mock.Config{Latency: time.Second})
	ctx, cancel := context.WithTimeout(t.Context(), 5*time.Millisecond)
	defer cancel()

	_, err := provider.Invoke(ctx, testModel, request.CanonicalRequest{})

	var got response.ProviderError
	if !errors.As(err, &got) {
		t.Fatalf("Invoke() error = %v, want response.ProviderError", err)
	}
	if got.Class != constant.ClassTransient {
		t.Fatalf("Class = %v, want %v", got.Class, constant.ClassTransient)
	}
	if got.ProviderCode != "context_deadline_exceeded" {
		t.Fatalf("ProviderCode = %q, want %q", got.ProviderCode, "context_deadline_exceeded")
	}
}

func TestOpenStreamEmitsChunksUsageThenDone(t *testing.T) {
	t.Parallel()

	usage := response.Usage{
		InputTokens:  2,
		OutputTokens: 5,
		Source:       constant.UsageReported,
	}
	provider := mock.New(mock.Config{
		ResponseText: "abcdef",
		ChunkSize:    2,
		Usage:        usage,
	})

	providerStream, err := provider.OpenStream(t.Context(), testModel, request.CanonicalRequest{Model: testModel})
	if err != nil {
		t.Fatalf("OpenStream() error = %v", err)
	}
	defer providerStream.Close()

	events := collect(t, providerStream.Events(), 2*time.Second)

	var text string
	seenUsage := false
	seenDone := false

	for _, ev := range events {
		switch e := ev.(type) {
		case response.Chunk:
			if seenUsage || seenDone {
				t.Fatalf("content chunk emitted after usage or done event")
			}
			text += e.Part.Text
		case response.UsageEvent:
			if seenUsage || seenDone {
				t.Fatalf("usage event emitted more than once or after done")
			}
			seenUsage = true
			if e.Usage != usage {
				t.Fatalf("usage = %+v, want %+v", e.Usage, usage)
			}
		case response.DoneEvent:
			if seenDone {
				t.Fatalf("done event emitted more than once")
			}
			seenDone = true
			if e.FinishReason != constant.FinishStop {
				t.Fatalf("FinishReason = %q, want %q", e.FinishReason, constant.FinishStop)
			}
		case response.ErrorEvent:
			t.Fatalf("unexpected error event: %v", e.Err)
		default:
			t.Fatalf("unexpected event type %T", ev)
		}
	}

	if text != "abcdef" {
		t.Fatalf("streamed text = %q, want %q", text, "abcdef")
	}
	if !seenUsage {
		t.Fatal("stream did not emit a usage event")
	}
	if !seenDone {
		t.Fatal("stream did not emit a done event")
	}
}

func TestCloseIsIdempotentAndStopsProduction(t *testing.T) {
	t.Parallel()

	provider := mock.New(mock.Config{
		ResponseText: "abcdef",
		ChunkSize:    1,
		ChunkDelay:   20 * time.Millisecond,
	})

	providerStream, err := provider.OpenStream(t.Context(), testModel, request.CanonicalRequest{})
	if err != nil {
		t.Fatalf("OpenStream() error = %v", err)
	}

	// Wait for the first chunk so the producer is definitely running.
	select {
	case _, ok := <-providerStream.Events():
		if !ok {
			t.Fatal("stream closed before the first chunk")
		}
	case <-time.After(time.Second):
		t.Fatal("timed out waiting for the first chunk")
	}

	// Concurrent Close calls must be safe and must not panic.
	var wg sync.WaitGroup
	for i := 0; i < 4; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_ = providerStream.Close()
		}()
	}
	wg.Wait()

	if err := providerStream.Close(); err != nil {
		t.Fatalf("Close() after repeated calls returned %v", err)
	}

	// The producer must close the channel promptly after Close.
	timeout := time.After(time.Second)
	for {
		select {
		case _, ok := <-providerStream.Events():
			if !ok {
				return
			}
		case <-timeout:
			t.Fatal("events channel did not close after Close")
		}
	}
}

func TestStreamStopsWhenContextIsCancelled(t *testing.T) {
	t.Parallel()

	provider := mock.New(mock.Config{
		ResponseText: "abcdef",
		ChunkSize:    1,
		ChunkDelay:   10 * time.Millisecond,
	})

	ctx, cancel := context.WithCancel(t.Context())
	providerStream, err := provider.OpenStream(ctx, testModel, request.CanonicalRequest{})
	if err != nil {
		t.Fatalf("OpenStream() error = %v", err)
	}
	defer providerStream.Close()

	// Consume one chunk, then cancel the caller's context.
	select {
	case ev, ok := <-providerStream.Events():
		if !ok {
			t.Fatal("stream closed before the first chunk")
		}
		if _, isChunk := ev.(response.Chunk); !isChunk {
			t.Fatalf("first event = %T, want response.Chunk", ev)
		}
	case <-time.After(time.Second):
		t.Fatal("timed out waiting for the first chunk")
	}
	cancel()

	for _, ev := range collect(t, providerStream.Events(), time.Second) {
		switch ev.(type) {
		case response.DoneEvent, response.ErrorEvent:
			t.Fatalf("unexpected terminal event after cancellation: %T", ev)
		}
	}
}

func TestStreamEmitsConfiguredErrorAfterChunks(t *testing.T) {
	t.Parallel()

	streamErr := response.ProviderError{
		Class:        constant.ClassTransient,
		Provider:     "mock",
		ProviderCode: "reset",
	}
	provider := mock.New(mock.Config{
		ResponseText:     "abcdef",
		ChunkSize:        2,
		StreamError:      &streamErr,
		ErrorAfterChunks: 1,
	})

	providerStream, err := provider.OpenStream(t.Context(), testModel, request.CanonicalRequest{})
	if err != nil {
		t.Fatalf("OpenStream() error = %v", err)
	}
	defer providerStream.Close()

	events := collect(t, providerStream.Events(), 2*time.Second)
	if len(events) != 2 {
		t.Fatalf("got %d events, want 2", len(events))
	}

	chunk, ok := events[0].(response.Chunk)
	if !ok {
		t.Fatalf("first event = %T, want response.Chunk", events[0])
	}
	if chunk.Part.Text != "ab" {
		t.Fatalf("first chunk = %q, want %q", chunk.Part.Text, "ab")
	}

	failure, ok := events[1].(response.ErrorEvent)
	if !ok {
		t.Fatalf("second event = %T, want response.ErrorEvent", events[1])
	}
	if failure.Err.ProviderCode != "reset" {
		t.Fatalf("ProviderCode = %q, want %q", failure.Err.ProviderCode, "reset")
	}
}

func TestOpenStreamReturnsConfiguredError(t *testing.T) {
	t.Parallel()

	openErr := response.ProviderError{
		Class:        constant.ClassPlatformConfig,
		Provider:     "mock",
		ProviderCode: "invalid_key",
	}
	provider := mock.New(mock.Config{OpenError: &openErr})

	providerStream, err := provider.OpenStream(t.Context(), testModel, request.CanonicalRequest{})
	if providerStream != nil {
		t.Fatalf("OpenStream() returned a stream together with error %v", err)
	}

	var got response.ProviderError
	if !errors.As(err, &got) {
		t.Fatalf("OpenStream() error = %v, want response.ProviderError", err)
	}
	if got.Class != constant.ClassPlatformConfig {
		t.Fatalf("Class = %v, want %v", got.Class, constant.ClassPlatformConfig)
	}
}

func TestProbe(t *testing.T) {
	t.Parallel()

	provider := mock.New(mock.Config{})
	result, err := provider.Probe(t.Context(), testModel)
	if err != nil {
		t.Fatalf("Probe() error = %v", err)
	}
	if !result.Healthy {
		t.Fatal("Probe() reported an unhealthy mock")
	}
}

func TestProbeHonoursContextCancellation(t *testing.T) {
	t.Parallel()

	provider := mock.New(mock.Config{})
	ctx, cancel := context.WithCancel(t.Context())
	cancel()

	_, err := provider.Probe(ctx, testModel)

	var got response.ProviderError
	if !errors.As(err, &got) {
		t.Fatalf("Probe() error = %v, want response.ProviderError", err)
	}
	if got.ProviderCode != "context_canceled" {
		t.Fatalf("ProviderCode = %q, want %q", got.ProviderCode, "context_canceled")
	}
}
