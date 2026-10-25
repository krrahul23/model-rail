package chat

import (
	"encoding/json"
	"net/http"

	applicationinference "modelrail/internal/application/inference"
	"modelrail/internal/constant"
	v1shared "modelrail/internal/controller/http/v1/shared"
	"modelrail/internal/logger"
	httpmiddleware "modelrail/internal/middleware/http"
)

// Handler serves the unary chat completion endpoint.
type Handler struct {
	useCase applicationinference.InferenceService
}

// NewHandler builds a chat handler around the application use case.
func NewHandler(useCase applicationinference.InferenceService) (handler *Handler) {
	handler = &Handler{useCase: useCase}
	return
}

// ServeHTTP decodes the v1 request, delegates to the application layer and
// writes either the completion or a typed error.
func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	log := logger.FromContext(r.Context())

	var payload Request
	decoder := json.NewDecoder(http.MaxBytesReader(w, r.Body, constant.MaxRequestBodyBytes))
	if err := decoder.Decode(&payload); err != nil {
		log.Warn("invalid request body", logger.Err(err))
		v1shared.WriteError(w, v1shared.RequestError(err))
		return
	}

	req, err := payload.ToDTO()
	if err != nil {
		log.Warn("invalid request mapping", logger.Err(err))
		v1shared.WriteError(w, v1shared.RequestError(err))
		return
	}
	if hint := r.Header.Get(constant.HeaderTaskClass); hint != constant.EmptyString {
		req.TaskClass = constant.TaskClass(hint)
	}

	result, err := h.useCase.Execute(r.Context(), req)
	if err != nil {
		log.Warn("inference failed", logger.Err(err))
		v1shared.WriteError(w, err)
		return
	}

	if err := v1shared.WriteJSON(w, http.StatusOK, FromResult(httpmiddleware.RequestIDFromContext(r.Context()), result)); err != nil {
		log.Error("write response failed", logger.Err(err))
	}
}
