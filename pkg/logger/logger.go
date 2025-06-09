package logger

import (
	"context"
	"log/slog"
	"os"
)

func New(level slog.Level, humanReadable bool) *slog.Logger {
	opts := &slog.HandlerOptions{
		Level: level,
	}

	var handler slog.Handler
	if humanReadable {
		handler = slog.NewTextHandler(os.Stdout, opts)
	} else {
		handler = slog.NewJSONHandler(os.Stdout, opts)
	}
	ctxHandler := &ctxHandler{Handler: handler}

	return slog.New(ctxHandler)
}

type ctxHandler struct {
	slog.Handler
}

func (h *ctxHandler) Handle(ctx context.Context, r slog.Record) error {
	r.AddAttrs(Attrs(ctx)...)
	return h.Handler.Handle(ctx, r)
}

type fieldsCtxKey struct{}

func WithAttrs(ctx context.Context, fields ...slog.Attr) context.Context {
	return context.WithValue(ctx, fieldsCtxKey{}, fields)
}

func Attrs(ctx context.Context) []slog.Attr {
	attrs, ok := ctx.Value(fieldsCtxKey{}).([]slog.Attr)
	if !ok {
		return nil
	}
	return attrs
}
