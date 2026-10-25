package constant

// Structured-log field keys shared by every component.
const (
	// FieldRequestID correlates log lines for one gateway request.
	FieldRequestID = "request_id"
	// FieldTraceID carries a distributed-trace identifier when one exists.
	FieldTraceID = "trace_id"
	// FieldMethod records the HTTP method.
	FieldMethod = "method"
	// FieldPath records the HTTP path.
	FieldPath = "path"
	// FieldStatus records the HTTP status code.
	FieldStatus = "status"
	// FieldDuration records an elapsed duration.
	FieldDuration = "duration"
	// FieldAddress records a listen address.
	FieldAddress = "addr"
	// FieldEnvironment records the deployment environment.
	FieldEnvironment = "environment"
)
