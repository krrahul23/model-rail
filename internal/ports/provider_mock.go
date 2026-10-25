package ports

import (
	"context"
	"errors"
	"time"

	"modelrail/internal/constant"
	"modelrail/internal/dto/request"
	"modelrail/internal/dto/response"
)

// Compile-time assertions that the test doubles satisfy the port contracts.
var (
	_ Provider       = (*MockProvider)(nil)
	_ ProviderStream = (*MockProviderStream)(nil)
)

// MockProvider is a configurable test double for Provider. It records calls and
// returns canned results or errors, so domain tests never need a real adapter.
type MockProvider struct {
	ProviderName     string
	CapabilitySet    response.Capabilities
	InvokeResult     response.Result
	InvokeErr        error
	InvokeDelay      time.Duration
	Stream           ProviderStream
	OpenStreamErr    error
	ProbeResultValue response.ProbeResult
	ProbeErr         error

	InvokeCalls     int
	OpenStreamCalls int
	ProbeCalls      int
	LastModel       constant.ModelRef
	LastRequest     request.CanonicalRequest
}

// NewMockProvider builds a test double with a canned unary result and error.
func NewMockProvider(result response.Result, err error) (provider *MockProvider) {
	provider = &MockProvider{InvokeResult: result, InvokeErr: err}
	return
}

// Name returns the configured provider name, defaulting to the mock provider.
func (m *MockProvider) Name() (name string) {
	name = m.ProviderName
	if name == constant.EmptyString {
		name = constant.ProviderMock
	}
	return
}

// Capabilities returns the configured capability set.
func (m *MockProvider) Capabilities(model constant.ModelRef) (caps response.Capabilities) {
	caps = m.CapabilitySet
	return
}

// Invoke records the call and returns the configured result or error. When a
// delay is configured it waits for the timer or the request context, mapping a
// context failure into the provider taxonomy.
func (m *MockProvider) Invoke(ctx context.Context, model constant.ModelRef, req request.CanonicalRequest) (result response.Result, err error) {
	m.recordCall(model, req)

	if ctxErr := ctx.Err(); ctxErr != nil {
		err = contextProviderError(m.Name(), ctxErr)
		return
	}

	if m.InvokeDelay > 0 {
		timer := time.NewTimer(m.InvokeDelay)
		defer timer.Stop()

		select {
		case <-ctx.Done():
			err = contextProviderError(m.Name(), ctx.Err())
			return
		case <-timer.C:
		}
	}

	result = m.InvokeResult
	err = m.InvokeErr
	return
}

// OpenStream records the call and returns the configured stream or error. A
// default closed stream is returned when none was supplied.
func (m *MockProvider) OpenStream(ctx context.Context, model constant.ModelRef, req request.CanonicalRequest) (stream ProviderStream, err error) {
	m.OpenStreamCalls++
	m.LastModel = model
	m.LastRequest = req

	if ctxErr := ctx.Err(); ctxErr != nil {
		err = contextProviderError(m.Name(), ctxErr)
		return
	}
	if m.OpenStreamErr != nil {
		err = m.OpenStreamErr
		return
	}

	stream = m.Stream
	if stream == nil {
		stream = NewMockProviderStream()
	}
	return
}

// Probe records the call and returns the configured probe result or error.
func (m *MockProvider) Probe(ctx context.Context, model constant.ModelRef) (result response.ProbeResult, err error) {
	m.ProbeCalls++
	m.LastModel = model

	if ctxErr := ctx.Err(); ctxErr != nil {
		err = contextProviderError(m.Name(), ctxErr)
		return
	}

	result = m.ProbeResultValue
	err = m.ProbeErr
	return
}

// recordCall stores the last invocation for assertions.
func (m *MockProvider) recordCall(model constant.ModelRef, req request.CanonicalRequest) {
	m.InvokeCalls++
	m.LastModel = model
	m.LastRequest = req
}

// MockProviderStream is a test double for ProviderStream. The constructor
// returns a stream whose event channel is closed after the supplied events, so
// consumers never block.
type MockProviderStream struct {
	EventChannel chan response.StreamEvent
	CloseErr     error
	CloseCalls   int
}

// NewMockProviderStream builds a stream that emits events and then closes.
func NewMockProviderStream(events ...response.StreamEvent) (stream *MockProviderStream) {
	channel := make(chan response.StreamEvent, len(events))
	for _, event := range events {
		channel <- event
	}
	close(channel)
	stream = &MockProviderStream{EventChannel: channel}
	return
}

// Events returns the event channel.
func (s *MockProviderStream) Events() (events <-chan response.StreamEvent) {
	events = s.EventChannel
	return
}

// Close records the call and returns the configured error.
func (s *MockProviderStream) Close() (err error) {
	s.CloseCalls++
	err = s.CloseErr
	return
}

// contextProviderError maps a context failure onto the provider taxonomy.
func contextProviderError(provider string, err error) (providerErr response.ProviderError) {
	class := constant.ClassUnknown
	code := constant.ProviderCodeContextCanceled
	if errors.Is(err, context.DeadlineExceeded) {
		class = constant.ClassTransient
		code = constant.ProviderCodeContextDeadlineExceeded
	}
	providerErr = response.ProviderError{
		Class:        class,
		Provider:     provider,
		ProviderCode: code,
		Cause:        err,
	}
	return
}
