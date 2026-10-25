package constant

// ErrorClass is the coarse failure category that drives retry and fallback
// decisions. It deliberately says nothing about HTTP status: the transport
// layer maps stable ErrorCodes to statuses and the OpenAI error envelope.
type ErrorClass int

// ErrorClass constants, ordered from least to most severe handling.
const (
	// ClassUnknown is the zero value. Consumers must treat it conservatively:
	// no retry and no fallback, so an unclassified error cannot amplify cost.
	ClassUnknown ErrorClass = iota
	// ClassTransient covers provider 5xx responses, connection resets and
	// timeouts that occur before the first content chunk.
	ClassTransient
	// ClassThrottled covers provider 429 responses; RetryAfter may be honoured
	// within the remaining deadline.
	ClassThrottled
	// ClassTerminalRequest marks a request the provider rejected as invalid;
	// resending it unchanged cannot succeed.
	ClassTerminalRequest
	// ClassTerminalPolicy marks a gateway policy rejection such as an
	// unauthorized key, an exhausted budget or a rate limit.
	ClassTerminalPolicy
	// ClassPlatformConfig marks a platform-side misconfiguration, such as a
	// provider credential being rejected. It is an incident, not a client error.
	ClassPlatformConfig
	// ClassDegradedStream marks a failure after the response was committed to
	// the client; fallback is never allowed at that point.
	ClassDegradedStream
)

// Stable string names of error classes, used in logs and diagnostics.
const (
	ClassNameUnknown         = "unknown"
	ClassNameTransient       = "transient"
	ClassNameThrottled       = "throttled"
	ClassNameTerminalRequest = "terminal_request"
	ClassNameTerminalPolicy  = "terminal_policy"
	ClassNamePlatformConfig  = "platform_config"
	ClassNameDegradedStream  = "degraded_stream"
)

// String returns the stable, lower-case name used in logs and diagnostics.
func (c ErrorClass) String() string {
	switch c {
	case ClassTransient:
		return ClassNameTransient
	case ClassThrottled:
		return ClassNameThrottled
	case ClassTerminalRequest:
		return ClassNameTerminalRequest
	case ClassTerminalPolicy:
		return ClassNameTerminalPolicy
	case ClassPlatformConfig:
		return ClassNamePlatformConfig
	case ClassDegradedStream:
		return ClassNameDegradedStream
	default:
		// Unknown or zero values must not be mistaken for a retryable class.
		return ClassNameUnknown
	}
}
