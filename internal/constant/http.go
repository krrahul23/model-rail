package constant

// HTTP methods, routes and header names used by the public API.
const (
	HTTPMethodGet  = "GET"
	HTTPMethodPost = "POST"

	PathHealth          = "/healthz"
	PathChatCompletions = "/v1/chat/completions"

	RouteHealth          = HTTPMethodGet + " " + PathHealth
	RouteChatCompletions = HTTPMethodPost + " " + PathChatCompletions

	HeaderRequestID   = "X-Request-Id"
	HeaderTaskClass   = "modelrail-task"
	HeaderContentType = "Content-Type"
	HeaderRetryAfter  = "Retry-After"

	ContentTypeJSON = "application/json"
	ContentTypeText = "text/plain; charset=utf-8"

	// HealthOK is the body returned by the liveness endpoint.
	HealthOK = "ok"
)

// OpenAI-compatible response fields and error types.
const (
	ObjectChatCompletion = "chat.completion"

	ErrorTypeInvalidRequest = "invalid_request_error"
	ErrorTypeAuthentication = "authentication_error"
	ErrorTypeRateLimit      = "rate_limit_error"
	ErrorTypeAPI            = "api_error"
)

// Request ID generation settings.
const (
	// RequestIDBytes is the number of random bytes rendered as the request ID.
	RequestIDBytes = 16
	// RequestIDFallbackBase is the numeral base used when random generation fails.
	RequestIDFallbackBase = 36
	// MinRetryAfterSeconds is the smallest Retry-After value emitted.
	MinRetryAfterSeconds = 1
)
