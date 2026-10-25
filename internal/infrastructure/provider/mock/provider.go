package mock

import (
	"context"
	"errors"
	"time"

	"modelrail/internal/constant"
	"modelrail/internal/dto/request"
	"modelrail/internal/dto/response"
	"modelrail/internal/ports"
)

// Compile-time assertion that the mock satisfies the provider port.
var _ ports.Provider = (*Provider)(nil)

// Provider is a deterministic provider adapter implementing ports.Provider.
// Every failure it returns is a response.ProviderError, never an SDK error.
type Provider struct {
	cfg Config
}

// New builds a mock provider from cfg and applies defaults for zero values.
func New(cfg Config) (provider *Provider) {
	provider = &Provider{cfg: cfg.withDefaults()}
	return
}

// Name returns the configured provider identifier.
func (p *Provider) Name() (name string) {
	name = p.cfg.ProviderName
	return
}

// Capabilities returns the configured capability set for every model; the mock
// does not model per-model capability differences.
func (p *Provider) Capabilities(model constant.ModelRef) (caps response.Capabilities) {
	caps = p.cfg.Capabilities
	return
}

// Invoke performs a simulated unary completion. It honours context cancellation
// and deadlines before returning content or a configured error.
func (p *Provider) Invoke(ctx context.Context, model constant.ModelRef, req request.CanonicalRequest) (result response.Result, err error) {
	start := time.Now()

	// Simulate provider latency, returning immediately if the caller cancels or
	// the request deadline expires.
	if waitErr := wait(ctx, p.cfg.Latency); waitErr != nil {
		err = p.contextError(waitErr)
		return
	}

	// Reproduce a configured provider failure, if any.
	if p.cfg.InvokeError != nil {
		err = *p.cfg.InvokeError
		return
	}

	result = response.Result{
		Content: []request.ContentPart{{
			Type: constant.ContentText,
			Text: p.cfg.ResponseText,
		}},
		Usage:        p.cfg.effectiveUsage(),
		ModelServed:  model,
		ProviderName: p.cfg.ProviderName,
		FinishReason: constant.FinishStop,
		Latency: response.LatencyBreakdown{
			TTFT:  p.cfg.Latency,
			Total: time.Since(start),
		},
	}
	return
}

// OpenStream opens a deterministic stream. A configured OpenError fails before
// any event is produced; StreamError fails after the configured chunk count.
func (p *Provider) OpenStream(ctx context.Context, model constant.ModelRef, req request.CanonicalRequest) (providerStream ports.ProviderStream, err error) {
	if p.cfg.OpenError != nil {
		err = *p.cfg.OpenError
		return
	}
	providerStream = newStream(ctx, p.cfg)
	return
}

// Probe performs the minimal health check used for half-open recovery. The mock
// probe always succeeds unless a ProbeError is configured.
func (p *Provider) Probe(ctx context.Context, model constant.ModelRef) (result response.ProbeResult, err error) {
	start := time.Now()

	// Observe a pre-cancelled or expired context without adding artificial delay.
	if waitErr := wait(ctx, 0); waitErr != nil {
		err = p.contextError(waitErr)
		return
	}
	if p.cfg.ProbeError != nil {
		err = *p.cfg.ProbeError
		return
	}

	result = response.ProbeResult{
		Healthy: true,
		Latency: time.Since(start),
	}
	return
}

// contextError converts a context failure into the provider taxonomy. A
// cancellation is unattributed (ClassUnknown) and must not be retried; a
// deadline is classified as transient so the reliability layer may retry while
// budget remains.
func (p *Provider) contextError(err error) (providerErr response.ProviderError) {
	class := constant.ClassUnknown
	code := constant.ProviderCodeContextCanceled
	if errors.Is(err, context.DeadlineExceeded) {
		class = constant.ClassTransient
		code = constant.ProviderCodeContextDeadlineExceeded
	}
	providerErr = response.ProviderError{
		Class:        class,
		Provider:     p.cfg.ProviderName,
		ProviderCode: code,
		Cause:        err,
	}
	return
}

// wait sleeps for delay while remaining responsive to context cancellation. A
// non-positive delay only checks whether the context is already done.
func wait(ctx context.Context, delay time.Duration) (err error) {
	if delay <= 0 {
		select {
		case <-ctx.Done():
			err = ctx.Err()
		default:
		}
		return
	}

	timer := time.NewTimer(delay)
	defer timer.Stop()

	select {
	case <-ctx.Done():
		err = ctx.Err()
	case <-timer.C:
	}
	return
}
