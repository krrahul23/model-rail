package inference

import (
	"context"
	"fmt"

	"modelrail/internal/constant"
	domaininference "modelrail/internal/domain/inference"
	"modelrail/internal/dto/request"
	"modelrail/internal/dto/response"
)

// inferenceService is the concrete application implementation.
type inferenceService struct {
	executor domaininference.InferenceService
}

// NewInferenceService builds the application service around the domain workflow.
func NewInferenceService(executor domaininference.InferenceService) (service InferenceService) {
	service = &inferenceService{executor: executor}
	return
}

// Execute normalizes and validates the request, then delegates to the domain.
func (s *inferenceService) Execute(ctx context.Context, req request.CanonicalRequest) (result response.Result, err error) {
	req = normalize(req)
	if err = validate(req); err != nil {
		return
	}
	result, err = s.executor.Complete(ctx, req)
	return
}

// normalize applies defaults for omitted values without changing explicit ones.
func normalize(req request.CanonicalRequest) (normalized request.CanonicalRequest) {
	normalized = req
	if normalized.TaskClass == constant.EmptyString {
		normalized.TaskClass = constant.TaskGeneral
	}
	if normalized.MaxOutputTokens == 0 {
		normalized.MaxOutputTokens = constant.DefaultMaxOutputTokens
	}
	if normalized.TotalDeadline <= 0 {
		normalized.TotalDeadline = constant.DefaultRequestTimeout
	}
	return
}

// validate enforces the business rules for an inference request.
func validate(req request.CanonicalRequest) (err error) {
	if req.Model == constant.EmptyString {
		err = invalidRequest("model is required")
		return
	}
	// Streaming is implemented in a later phase; rejecting it keeps the API
	// honest instead of returning a unary body for a streaming request.
	if req.Stream {
		err = invalidRequest("streaming is not supported yet")
		return
	}
	if len(req.Messages) == 0 {
		err = invalidRequest("messages must not be empty")
		return
	}
	if req.MaxOutputTokens < 0 {
		err = invalidRequest("max_tokens must not be negative")
		return
	}
	if req.Temperature != nil && (*req.Temperature < constant.MinTemperature || *req.Temperature > constant.MaxTemperature) {
		err = invalidRequest("temperature must be between %.0f and %.0f", constant.MinTemperature, constant.MaxTemperature)
		return
	}
	if !validTaskClass(req.TaskClass) {
		err = invalidRequest("unknown task class %q", req.TaskClass)
		return
	}
	err = validateMessages(req.Messages)
	return
}

// validateMessages checks every message role and content part.
func validateMessages(messages []request.Message) (err error) {
	for index, message := range messages {
		if !validRole(message.Role) {
			err = invalidRequest("messages[%d] has unknown role %q", index, message.Role)
			return
		}
		if len(message.Parts) == 0 {
			err = invalidRequest("messages[%d] must contain at least one content part", index)
			return
		}
		for _, part := range message.Parts {
			if part.Type != constant.ContentText {
				err = invalidRequest("messages[%d] has unsupported content type %q", index, part.Type)
				return
			}
			if part.Text == constant.EmptyString {
				err = invalidRequest("messages[%d] contains empty text", index)
				return
			}
		}
	}
	return
}

// validRole reports whether the role is part of the canonical protocol.
func validRole(role constant.Role) (valid bool) {
	switch role {
	case constant.RoleSystem, constant.RoleUser, constant.RoleAssistant, constant.RoleTool:
		valid = true
	}
	return
}

// validTaskClass reports whether the task class is one routing understands.
func validTaskClass(class constant.TaskClass) (valid bool) {
	switch class {
	case constant.TaskGeneral, constant.TaskCoding, constant.TaskReasoning, constant.TaskStructuredExtraction:
		valid = true
	}
	return
}

// invalidRequest builds the typed error returned for a rejected request.
func invalidRequest(format string, args ...any) (err error) {
	err = response.GatewayError{
		Class: constant.ClassTerminalRequest,
		Code:  constant.CodeRequestInvalid,
		Cause: fmt.Errorf(format, args...),
	}
	return
}
