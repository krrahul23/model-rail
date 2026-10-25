package chat_test

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	applicationinference "modelrail/internal/application/inference"
	applicationmock "modelrail/internal/application/inference"
	"modelrail/internal/constant"
	"modelrail/internal/controller/http/v1/chat"
	dtorequest "modelrail/internal/dto/request"
	dtoresponse "modelrail/internal/dto/response"
	httpmiddleware "modelrail/internal/middleware/http"
)

// newHandler wraps the chat handler with the request middleware so request IDs
// and logging behave as they do in production.
func newHandler(useCase applicationinference.InferenceService) (handler http.Handler) {
	handler = httpmiddleware.WithRequestContext(chat.NewHandler(useCase))
	return
}

// successResult returns a canonical result for the wire tests.
func successResult() (result dtoresponse.Result) {
	result = dtoresponse.Result{
		Content: []dtorequest.ContentPart{{Type: constant.ContentText, Text: "hello back"}},
		Usage: dtoresponse.Usage{
			InputTokens:  3,
			OutputTokens: 4,
			Source:       constant.UsageReported,
		},
		ModelServed:  "mock-model",
		ProviderName: constant.ProviderMock,
		FinishReason: constant.FinishStop,
	}
	return
}

func TestChatCompletionsSuccess(t *testing.T) {
	t.Parallel()

	useCase := &applicationmock.MockInferenceService{Result: successResult()}
	handler := newHandler(useCase)
	body := `{"model":"mock-model","messages":[{"role":"user","content":"hello"}]}`
	req := httptest.NewRequest(http.MethodPost, constant.PathChatCompletions, strings.NewReader(body))
	recorder := httptest.NewRecorder()

	handler.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusOK)
	}
	requestID := recorder.Header().Get(constant.HeaderRequestID)
	if requestID == constant.EmptyString {
		t.Fatal("response is missing the request ID header")
	}

	var payload struct {
		ID      string `json:"id"`
		Object  string `json:"object"`
		Model   string `json:"model"`
		Choices []struct {
			Message struct {
				Role    string `json:"role"`
				Content string `json:"content"`
			} `json:"message"`
			FinishReason string `json:"finish_reason"`
		} `json:"choices"`
		Usage struct {
			PromptTokens     int64 `json:"prompt_tokens"`
			CompletionTokens int64 `json:"completion_tokens"`
			TotalTokens      int64 `json:"total_tokens"`
		} `json:"usage"`
	}
	if err := json.Unmarshal(recorder.Body.Bytes(), &payload); err != nil {
		t.Fatalf("decode response: %v", err)
	}

	if payload.ID != requestID {
		t.Fatalf("response id = %q, want request ID %q", payload.ID, requestID)
	}
	if payload.Object != constant.ObjectChatCompletion {
		t.Fatalf("object = %q, want %q", payload.Object, constant.ObjectChatCompletion)
	}
	if payload.Model != "mock-model" {
		t.Fatalf("model = %q, want %q", payload.Model, "mock-model")
	}
	if len(payload.Choices) != 1 {
		t.Fatalf("len(choices) = %d, want 1", len(payload.Choices))
	}
	if payload.Choices[0].Message.Role != string(constant.RoleAssistant) {
		t.Fatalf("role = %q, want %q", payload.Choices[0].Message.Role, constant.RoleAssistant)
	}
	if payload.Choices[0].Message.Content != "hello back" {
		t.Fatalf("content = %q, want %q", payload.Choices[0].Message.Content, "hello back")
	}
	if payload.Choices[0].FinishReason != string(constant.FinishStop) {
		t.Fatalf("finish_reason = %q, want %q", payload.Choices[0].FinishReason, constant.FinishStop)
	}
	if payload.Usage.PromptTokens != 3 || payload.Usage.CompletionTokens != 4 || payload.Usage.TotalTokens != 7 {
		t.Fatalf("usage = %+v, want prompt 3, completion 4, total 7", payload.Usage)
	}
	if useCase.Calls != 1 {
		t.Fatalf("application calls = %d, want 1", useCase.Calls)
	}
	if len(useCase.LastRequest.Messages) != 1 || useCase.LastRequest.Messages[0].Parts[0].Text != "hello" {
		t.Fatalf("canonical request = %+v, want the decoded message", useCase.LastRequest)
	}
}

func TestChatCompletionsPassesTaskClassHeader(t *testing.T) {
	t.Parallel()

	useCase := &applicationmock.MockInferenceService{Result: successResult()}
	handler := newHandler(useCase)
	body := `{"model":"mock-model","messages":[{"role":"user","content":"hello"}]}`
	req := httptest.NewRequest(http.MethodPost, constant.PathChatCompletions, strings.NewReader(body))
	req.Header.Set(constant.HeaderTaskClass, string(constant.TaskCoding))
	recorder := httptest.NewRecorder()

	handler.ServeHTTP(recorder, req)

	if useCase.LastRequest.TaskClass != constant.TaskCoding {
		t.Fatalf("TaskClass = %q, want %q", useCase.LastRequest.TaskClass, constant.TaskCoding)
	}
}

func TestChatCompletionsMapsGatewayError(t *testing.T) {
	t.Parallel()

	retryAfter := 2 * time.Second
	useCase := &applicationmock.MockInferenceService{Err: dtoresponse.GatewayError{
		Class:      constant.ClassThrottled,
		Code:       constant.CodeProviderThrottled,
		RetryAfter: &retryAfter,
		Cause:      errors.New("provider said 429"),
	}}
	handler := newHandler(useCase)
	body := `{"model":"mock-model","messages":[{"role":"user","content":"hello"}]}`
	req := httptest.NewRequest(http.MethodPost, constant.PathChatCompletions, strings.NewReader(body))
	recorder := httptest.NewRecorder()

	handler.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusTooManyRequests {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusTooManyRequests)
	}
	if got := recorder.Header().Get(constant.HeaderRetryAfter); got != "2" {
		t.Fatalf("Retry-After = %q, want %q", got, "2")
	}

	var payload struct {
		Error struct {
			Type    string `json:"type"`
			Code    string `json:"code"`
			Message string `json:"message"`
		} `json:"error"`
	}
	if err := json.Unmarshal(recorder.Body.Bytes(), &payload); err != nil {
		t.Fatalf("decode error response: %v", err)
	}
	if payload.Error.Type != constant.ErrorTypeRateLimit {
		t.Fatalf("error type = %q, want %q", payload.Error.Type, constant.ErrorTypeRateLimit)
	}
	if payload.Error.Code != string(constant.CodeProviderThrottled) {
		t.Fatalf("error code = %q, want %q", payload.Error.Code, constant.CodeProviderThrottled)
	}
	if payload.Error.Message == "provider said 429" {
		t.Fatal("provider diagnostics must not be exposed to clients")
	}
}

func TestChatCompletionsRejectsInvalidJSON(t *testing.T) {
	t.Parallel()

	handler := newHandler(&applicationmock.MockInferenceService{})
	req := httptest.NewRequest(http.MethodPost, constant.PathChatCompletions, strings.NewReader(`{"model":`))
	recorder := httptest.NewRecorder()

	handler.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusBadRequest)
	}
}

func TestChatCompletionsRejectsUnsupportedContent(t *testing.T) {
	t.Parallel()

	handler := newHandler(&applicationmock.MockInferenceService{})
	body := `{"model":"mock-model","messages":[{"role":"user","content":[{"type":"image","text":"x"}]}]}`
	req := httptest.NewRequest(http.MethodPost, constant.PathChatCompletions, strings.NewReader(body))
	recorder := httptest.NewRecorder()

	handler.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusBadRequest)
	}
}
