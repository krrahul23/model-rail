// Package v1 registers routes for API version 1.
package v1

import (
	"net/http"

	applicationinference "modelrail/internal/application/inference"
	"modelrail/internal/constant"
	"modelrail/internal/controller/http/v1/chat"
)

// Deps groups the use cases required by API version 1.
type Deps struct {
	Inference applicationinference.InferenceService
}

// Register mounts the v1 endpoints on the router.
func Register(mux *http.ServeMux, deps Deps) {
	mux.Handle(constant.RouteChatCompletions, chat.NewHandler(deps.Inference))
}
