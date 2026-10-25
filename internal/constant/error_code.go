package constant

// ErrorCode is the stable, client-visible identifier of a failure. Codes are
// part of the API contract: once published they are not renamed.
type ErrorCode string

// ErrorCode constants
const (
	// Admission, authentication and limit failures.
	CodeSourceRateLimited      ErrorCode = "source_rate_limited"
	CodeUnauthorized           ErrorCode = "unauthorized"
	CodeKeyRevoked             ErrorCode = "key_revoked"
	CodeRateLimited            ErrorCode = "rate_limited"
	CodeTPMLimit               ErrorCode = "tpm_limit"
	CodeTenantConcurrencyLimit ErrorCode = "tenant_concurrency_limit"
	CodePlatformOverloaded     ErrorCode = "platform_overloaded"
	CodeRequestInvalid         ErrorCode = "request_invalid"
	CodeModelNotAllowed        ErrorCode = "model_not_allowed"
	CodeNoCandidate            ErrorCode = "no_candidate"
	CodeBudgetExceeded         ErrorCode = "budget_exceeded"
	CodeTokenQuotaExceeded     ErrorCode = "token_quota_exceeded"

	// Provider-origin failures, classified by the adapter.
	CodeProviderThrottled   ErrorCode = "provider_throttled"
	CodeProviderTimeout     ErrorCode = "provider_timeout"
	CodeProviderUnavailable ErrorCode = "provider_unavailable"
	CodeProviderOverloaded  ErrorCode = "provider_overloaded"
	CodePlatformConfigError ErrorCode = "platform_config_error"

	// Response validation and stream degradation.
	CodeOutputValidationFailed ErrorCode = "output_validation_failed"
	CodeDegradedStream         ErrorCode = "degraded_stream"

	// Money-path failures.
	CodeSettlementFailed ErrorCode = "settlement_failed"

	// Internal failures that have no more specific classification.
	CodeInternalError ErrorCode = "internal_error"
)
