package log

import (
	"context"
	"log/slog"
	"os"
	"ronce/src/go/app"
	"ronce/src/go/errors"
)

func New() *Logger {
	return &Logger{
		Logger: slog.With("app", app.Name, "version", app.Version),
	}
}

type Logger struct {
	*slog.Logger
	Format string `key:"format" description:"log format [json, text]"`
	Level  string `key:"level"  description:"minimum log level [debug, info, warn, error]"`
}

func (l *Logger) Init() error {
	var level slog.Level
	switch l.Level {
	case "debug", "dbug":
		level = slog.LevelDebug
	case "info":
		level = slog.LevelInfo
	case "warn", "warning":
		level = slog.LevelWarn
	case "error", "eror":
		level = slog.LevelError
	default:
		return errors.Newf("unknown error level %q", l.Level)
	}

	var handler slog.Handler
	switch l.Format {
	case "text":
		handler = NewTextHandler(os.Stderr, level)
	case "json":
		handler = slog.NewJSONHandler(os.Stderr, &slog.HandlerOptions{
			Level: level,
		})
	default:
		return errors.Newf("invalid format %s: logger format can be either text or json", l.Format)
	}

	l.Logger = slog.New(ErrorsHandler{handler}).With("app", app.Name, "version", app.Version)

	return nil
}

func (l *Logger) With(args ...any) *Logger {
	return &Logger{
		Logger: l.Logger.With(args...),
	}
}

// AddContextFields adds every fields of m into the context for logging purposes.
// The fields are exposed by the WithContext function.
func AddContextFields(ctx context.Context, keyvals []any) context.Context {
	if len(keyvals)%2 != 0 {
		keyvals = append(keyvals, "N/C")
	}
	return context.WithValue(ctx, "log.fields", keyvals)
}

// WithContext injects context loggable data into the logger.
// It is similar to the package's function.
func WithContext(ctx context.Context, logger *Logger) *Logger {
	attrs, _ := ctx.Value("log.fields").([]any)
	return logger.With(attrs...)
}
