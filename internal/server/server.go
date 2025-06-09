package server

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/pirellik/sequence-api/internal/openapi"
	"github.com/pirellik/sequence-api/pkg/middleware"
	"github.com/pkg/errors"
)

func New(svc openapi.StrictServerInterface, port int) *http.Server {
	strictHandler := openapi.NewStrictHandlerWithOptions(svc, nil, openapi.StrictHTTPServerOptions{
		RequestErrorHandlerFunc:  errorHandler,
		ResponseErrorHandlerFunc: errorHandler,
	})

	r := http.NewServeMux()
	handler := openapi.HandlerFromMux(strictHandler, r)
	handler = middleware.Apply(handler,
		middleware.Logging,
		middleware.RequestID,
	)
	return &http.Server{
		Addr:    fmt.Sprintf(":%d", port),
		Handler: handler,
	}
}

func errorHandler(w http.ResponseWriter, r *http.Request, err error) {
	slog.ErrorContext(r.Context(), "error", "err", err)
	if err == nil {
		writeErrorResponse(w, r, openapi.Error{
			Message: "internal server error",
			Status:  http.StatusInternalServerError,
		})
		return
	}

	err = errors.Cause(err)
	apiErr, ok := err.(*APIError)
	if !ok {
		writeErrorResponse(w, r, openapi.Error{
			Message: "internal server error",
			Status:  http.StatusInternalServerError,
		})
		return
	}

	writeErrorResponse(w, r, openapi.Error{
		Message: apiErr.Error(),
		Status:  int64(apiErr.StatusCode()),
	})
}

func writeErrorResponse(w http.ResponseWriter, r *http.Request, e openapi.Error) {
	err := json.NewEncoder(w).Encode(e)
	if err != nil {
		slog.ErrorContext(r.Context(), "failed to encode error response", "error", err)
	}
}
