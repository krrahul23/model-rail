package httpcontroller_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	applicationmock "modelrail/internal/application/inference"
	"modelrail/internal/constant"
	httpcontroller "modelrail/internal/controller/http"
)

func TestHealth(t *testing.T) {
	t.Parallel()

	router := httpcontroller.NewRouter(httpcontroller.Deps{
		Inference: &applicationmock.MockInferenceService{},
	})
	req := httptest.NewRequest(http.MethodGet, constant.PathHealth, nil)
	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusOK)
	}
	if recorder.Body.String() != constant.HealthOK {
		t.Fatalf("body = %q, want %q", recorder.Body.String(), constant.HealthOK)
	}
	if recorder.Header().Get(constant.HeaderRequestID) == constant.EmptyString {
		t.Fatal("response is missing the request ID header")
	}
}
