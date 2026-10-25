// Package mock provides a deterministic, configurable implementation of
// ports.Provider for tests and local development. It never performs network
// calls; latency, stream pacing and failures are controlled by Config.
package mock

import (
	"time"

	"modelrail/internal/constant"
	"modelrail/internal/dto/response"
)

// Config controls the behaviour of the mock provider. Zero values select
// sensible defaults; error fields are optional and let tests reproduce provider
// faults without network simulation.
type Config struct {
	ProviderName     string                  // provider identifier reported by Name
	Capabilities     response.Capabilities   // advertised model capabilities
	ResponseText     string                  // text returned for unary and streaming calls
	ChunkSize        int                     // characters per stream chunk
	Latency          time.Duration           // unary latency; also the default stream TTFT
	TTFT             time.Duration           // delay before the first stream event
	ChunkDelay       time.Duration           // delay between stream chunks
	Usage            response.Usage          // reported token usage
	InvokeError      *response.ProviderError // returned by Invoke when set
	OpenError        *response.ProviderError // returned by OpenStream when set
	StreamError      *response.ProviderError // emitted mid-stream when set
	ErrorAfterChunks int                     // chunks emitted before StreamError
	ProbeError       *response.ProviderError // returned by Probe when set
}

// withDefaults returns a copy of the configuration with defaults applied. It
// never mutates the caller's value.
func (c Config) withDefaults() (cfg Config) {
	cfg = c
	if cfg.ProviderName == constant.EmptyString {
		cfg.ProviderName = constant.DefaultProviderName
	}
	if cfg.ChunkSize <= 0 {
		cfg.ChunkSize = constant.DefaultChunkSize
	}
	if cfg.Capabilities == (response.Capabilities{}) {
		cfg.Capabilities = defaultCapabilities()
	}
	if cfg.Usage.Source == constant.EmptyString {
		cfg.Usage.Source = constant.UsageReported
	}
	if cfg.TTFT == 0 && cfg.Latency > 0 {
		cfg.TTFT = cfg.Latency
	}
	return
}

// defaultCapabilities returns the capability set advertised by the mock when
// the caller does not override it.
func defaultCapabilities() (caps response.Capabilities) {
	caps = response.Capabilities{
		ContextWindow:           constant.DefaultContextWindow,
		SupportsStreaming:       true,
		SupportsTools:           true,
		SupportsJSONMode:        true,
		SupportsCachedTokens:    true,
		SupportsReasoningTokens: true,
	}
	return
}

// effectiveUsage returns the configured usage, deriving an output-token count
// from the response text when the caller supplied no token counts at all.
func (c Config) effectiveUsage() (usage response.Usage) {
	usage = c.Usage
	if usage.InputTokens == 0 && usage.CachedInputTokens == 0 &&
		usage.OutputTokens == 0 && usage.ReasoningTokens == 0 {
		usage.OutputTokens = int64(len([]rune(c.ResponseText)))
	}
	return
}
