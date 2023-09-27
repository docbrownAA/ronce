package log

import (
	"context"
	"log/slog"
	"ronce/src/go/errors"
)

type ErrorsHandler struct {
	next slog.Handler
}

func (h ErrorsHandler) Enabled(ctx context.Context, level slog.Level) bool {
	return h.next.Enabled(ctx, level)
}

func (h ErrorsHandler) Handle(ctx context.Context, r slog.Record) error {
	r.Attrs(func(a slog.Attr) bool {
		err, ok := a.Value.Any().(error)
		if !ok {
			return true
		}

		var e errors.E
		if errors.As(err, &e) {
			for k, v := range e.Context {
				r.AddAttrs(slog.Any(k, v))
			}
		}

		return true
	})

	return h.next.Handle(ctx, r)
}

func (h ErrorsHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return ErrorsHandler{h.next.WithAttrs(attrs)}
}

func (h ErrorsHandler) WithGroup(name string) slog.Handler {
	return ErrorsHandler{h.next.WithGroup(name)}
}
