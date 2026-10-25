package v1shared

import (
	"errors"
	"net/http"
	"strconv"
	"time"

	"modelrail/internal/constant"
	"modelrail/internal/dto/response"
)

// errorEnvelope is the OpenAI-compatible error body for API version 1.
type errorEnvelope struct {
	Error errorBody `json:"error"`
}

// errorBody carries the human message, the error category and the stable code.
type errorBody struct {
	Message string `json:"message"`
	Type    string `json:"type"`
	Code    string `json:"code"`
}

// RequestError wraps a request-mapping failure in the gateway taxonomy.
func RequestError(err error) (gatewayErr response.GatewayError) {
	gatewayErr = response.GatewayError{
		Class: constant.ClassTerminalRequest,
		Code:  constant.CodeRequestInvalid,
		Cause: err,
	}
	return
}

// WriteError renders a gateway or internal failure as the v1 error envelope.
func WriteError(w http.ResponseWriter, err error) {
	status, errorType, code, message, retryAfter := classifyError(err)
	if retryAfter != nil {
		w.Header().Set(constant.HeaderRetryAfter, strconv.Itoa(retryAfterSeconds(*retryAfter)))
	}
	_ = WriteJSON(w, status, errorEnvelope{Error: errorBody{
		Message: message,
		Type:    errorType,
		Code:    code,
	}})
}

// classifyError maps an error onto the HTTP status, envelope type and message.
func classifyError(err error) (status int, errorType, code, message string, retryAfter *time.Duration) {
	var gatewayErr response.GatewayError
	if !errors.As(err, &gatewayErr) {
		status = http.StatusInternalServerError
		errorType = constant.ErrorTypeAPI
		code = string(constant.CodeInternalError)
		message = "internal error"
		return
	}

	status, errorType = statusForCode(gatewayErr.Code)
	code = string(gatewayErr.Code)
	retryAfter = gatewayErr.RetryAfter
	message = messageFor(gatewayErr)
	return
}

// statusForCode returns the HTTP status and envelope type for a stable code.
func statusForCode(code constant.ErrorCode) (status int, errorType string) {
	switch code {
	case constant.CodeRequestInvalid:
		status, errorType = http.StatusBadRequest, constant.ErrorTypeInvalidRequest
	case constant.CodeOutputValidationFailed:
		status, errorType = http.StatusUnprocessableEntity, constant.ErrorTypeInvalidRequest
	case constant.CodeUnauthorized, constant.CodeKeyRevoked:
		status, errorType = http.StatusUnauthorized, constant.ErrorTypeAuthentication
	case constant.CodeModelNotAllowed:
		status, errorType = http.StatusForbidden, constant.ErrorTypeInvalidRequest
	case constant.CodeBudgetExceeded, constant.CodeTokenQuotaExceeded:
		status, errorType = http.StatusPaymentRequired, constant.ErrorTypeInvalidRequest
	case constant.CodeSourceRateLimited, constant.CodeRateLimited, constant.CodeTPMLimit,
		constant.CodeTenantConcurrencyLimit, constant.CodeProviderThrottled:
		status, errorType = http.StatusTooManyRequests, constant.ErrorTypeRateLimit
	case constant.CodeProviderTimeout:
		status, errorType = http.StatusGatewayTimeout, constant.ErrorTypeAPI
	case constant.CodeProviderUnavailable, constant.CodePlatformConfigError:
		status, errorType = http.StatusBadGateway, constant.ErrorTypeAPI
	case constant.CodeProviderOverloaded, constant.CodePlatformOverloaded, constant.CodeNoCandidate:
		status, errorType = http.StatusServiceUnavailable, constant.ErrorTypeAPI
	default:
		status, errorType = http.StatusInternalServerError, constant.ErrorTypeAPI
	}
	return
}

// messageFor selects a client-safe message. Only gateway-originated causes are
// echoed; provider diagnostics stay in logs.
func messageFor(err response.GatewayError) (message string) {
	switch err.Class {
	case constant.ClassTerminalRequest, constant.ClassTerminalPolicy:
		if err.Cause != nil {
			message = err.Cause.Error()
			return
		}
		message = "request rejected"
		return
	}
	message = "upstream provider request failed"
	return
}

// retryAfterSeconds converts a duration to the whole seconds clients expect.
func retryAfterSeconds(delay time.Duration) (seconds int) {
	seconds = int(delay.Seconds())
	if seconds < constant.MinRetryAfterSeconds {
		seconds = constant.MinRetryAfterSeconds
	}
	return
}
