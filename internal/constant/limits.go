package constant

import "time"

// Application limits and defaults applied before a request reaches a provider.
const (
	// DefaultMaxOutputTokens is used when the client does not request a limit.
	DefaultMaxOutputTokens = 512
	// MaxRequestBodyBytes caps the accepted inference request body.
	MaxRequestBodyBytes = 1 << 20
	// DefaultRequestTimeout bounds a whole inference request.
	DefaultRequestTimeout = 60 * time.Second
)

// Accepted temperature range.
const (
	MinTemperature = 0.0
	MaxTemperature = 2.0
)
