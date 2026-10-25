// Package httpcontroller builds the public HTTP router and owns unversioned
// routes such as health checks.
package httpcontroller

import (
	"net/http"

	applicationinference "modelrail/internal/application/inference"
	"modelrail/internal/constant"
	"modelrail/internal/controller/http/v1"
	httpmiddleware "modelrail/internal/middleware/http"
)

// Deps groups the dependencies required by the HTTP router.
type Deps struct {
	Inference applicationinference.InferenceService
}

// NewRouter builds the public HTTP router and applies transport middleware.
func NewRouter(deps Deps) (handler http.Handler) {
	mux := http.NewServeMux()
	mux.HandleFunc(constant.RouteHealth, health)
	v1.Register(mux, v1.Deps{Inference: deps.Inference})
	handler = httpmiddleware.WithRequestContext(mux)
	return
}

// health reports process liveness.
func health(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set(constant.HeaderContentType, constant.ContentTypeText)
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(constant.HealthOK))
}
