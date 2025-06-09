package middleware

import (
	"log/slog"
	"net/http"

	"github.com/google/uuid"
	"github.com/pirellik/sequence-api/pkg/logger"
)

func Apply(handler http.Handler, middlewares ...func(http.Handler) http.Handler) http.Handler {
	for _, middleware := range middlewares {
		handler = middleware(handler)
	}
	return handler
}

func RequestID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestID := r.Header.Get("X-Request-ID")
		if requestID == "" {
			requestID = uuid.New().String()
			r.Header.Set("X-Request-ID", requestID)
		}
		attrs := logger.Attrs(r.Context())
		attrs = append(attrs, slog.String("request_id", requestID))
		ctx := logger.WithAttrs(r.Context(), attrs...)
		r = r.WithContext(ctx)
		next.ServeHTTP(w, r)
	})
}

func Logging(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		writer := &writerWithStatus{ResponseWriter: w}
		next.ServeHTTP(writer, r)
		slog.InfoContext(r.Context(), "request", "method", r.Method, "path", r.URL.Path, "status", writer.status)
	})
}

type writerWithStatus struct {
	http.ResponseWriter
	status int
}

func (w *writerWithStatus) WriteHeader(status int) {
	w.status = status
	w.ResponseWriter.WriteHeader(status)
}
