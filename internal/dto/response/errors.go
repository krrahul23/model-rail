package response

import (
	"fmt"
	"time"

	"modelrail/internal/constant"
)

// ProviderError is returned by adapters. It carries no transport knowledge:
// the controller maps taxonomy codes to HTTP statuses. Class drives retry and
// fallback decisions, ProviderCode is diagnostics only, and RetryAfter is set
// only when the provider supplied one (typically on 429 or 503).
type ProviderError struct {
	Class        constant.ErrorClass
	Provider     string
	ProviderCode string
	RetryAfter   *time.Duration
	Cause        error
}

// Error renders a diagnostic string. It is intended for logs; clients receive
// the mapped ErrorCode and envelope instead.
func (e ProviderError) Error() string {
	// Provider codes are optional, so only include them when the adapter
	// captured one from the upstream response.
	if e.ProviderCode != constant.EmptyString {
		return fmt.Sprintf("provider %s: %s (%s)", e.Provider, e.Class, e.ProviderCode)
	}
	return fmt.Sprintf("provider %s: %s", e.Provider, e.Class)
}

// Unwrap exposes the underlying cause to errors.Is and errors.As.
func (e ProviderError) Unwrap() error {
	return e.Cause
}

// GatewayError is produced by the gateway pipeline itself. Code is the stable,
// client-visible identifier; the controller maps it to an HTTP status and the
// OpenAI-compatible error envelope.
type GatewayError struct {
	Class      constant.ErrorClass
	Code       constant.ErrorCode
	RetryAfter *time.Duration
	Cause      error
}

// Error renders the code plus the wrapped cause, when one exists.
func (e GatewayError) Error() string {
	// A gateway error may wrap an infrastructure failure; include it for
	// diagnostics without changing the stable code.
	if e.Cause != nil {
		return fmt.Sprintf("%s: %v", e.Code, e.Cause)
	}
	return string(e.Code)
}

// Unwrap exposes the underlying cause to errors.Is and errors.As.
func (e GatewayError) Unwrap() error {
	return e.Cause
}
