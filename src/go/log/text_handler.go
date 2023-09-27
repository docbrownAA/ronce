package log

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log/slog"
	"sync"
	"time"
)

const Reset = 0

const (
	Black int = iota + 30
	Red
	Green
	Yellow
	Blue
	Magenta
	Cyan
	White
)

const JustifyThres = 25

type TextHandler struct {
	writer io.Writer
	level  slog.Level
	lock   *sync.Mutex
	groups []byte
	attrs  []slog.Attr
}

func NewTextHandler(w io.Writer, l slog.Level) *TextHandler {
	return &TextHandler{
		writer: w,
		level:  l,
		lock:   &sync.Mutex{},
	}
}

func (h TextHandler) Enabled(ctx context.Context, level slog.Level) bool {
	return h.level <= level
}

func (h TextHandler) Handle(ctx context.Context, r slog.Record) error {
	// Initialize the buffer with enough space for most log lines. This
	// avoids the need for a buffer pool for reasonable amoounts of logs
	// when performance isn't an issue.
	var buf = make([]byte, 0, 1024)

	// Short formatting of the timestamp. With essentially remove the date
	// part because it's not pertinent for development.
	buf = append(buf, []byte(r.Time.Format("15:04:05.000"))...)

	// Colorize the log level for readability.
	var color int
	switch r.Level {
	case slog.LevelError:
		color = Red
	case slog.LevelWarn:
		color = Yellow
	case slog.LevelInfo:
		color = Green
	case slog.LevelDebug:
		color = Blue
	}
	buf = append(buf, ' ')
	buf = append(buf, []byte(fmt.Sprintf("\x1b[%dm", color))...)
	buf = append(buf, []byte(r.Level.String())...)
	buf = append(buf, []byte(fmt.Sprintf("\x1b[%dm", Reset))...)
	buf = append(buf, bytes.Repeat([]byte{' '}, 5-len(r.Level.String()))...)

	// Add the message, with eventual justification if below the threshold so
	// messages and attributes are mostly aligned.
	buf = append(buf, ' ')
	buf = append(buf, []byte(r.Message)...)
	if len(r.Message) < JustifyThres {
		buf = append(buf, bytes.Repeat([]byte{' '}, JustifyThres-len(r.Message))...)
	}

	// Append any preformatted attributes. The key already has the group of the
	// attr prefixed.
	for _, attr := range h.attrs {
		// Resolve the value first, which ahdnles the LogValuer interface.
		attr.Value = attr.Value.Resolve()
		// If the attr is empty, ignore it.
		if attr.Equal(slog.Attr{}) {
			continue
		}
		buf = h.appendAttr(buf, attr, color)
	}

	r.Attrs(func(attr slog.Attr) bool {
		// Resolve the value first, which ahdnles the LogValuer interface.
		attr.Value = attr.Value.Resolve()
		// If the attr is empty, ignore it.
		if attr.Equal(slog.Attr{}) {
			return true
		}
		buf = h.appendAttr(buf, slog.Attr{
			Key:   fmt.Sprintf("%s%s", h.groups, attr.Key),
			Value: attr.Value,
		}, color)
		return true
	})
	buf = append(buf, '\n')

	h.lock.Lock()
	defer h.lock.Unlock()
	_, err := h.writer.Write(buf)
	return err
}

func (h *TextHandler) appendAttr(buf []byte, attr slog.Attr, color int) []byte {
	// Ignore the app and version keys, as we don't need them in debug mode.
	switch attr.Key {
	case "app", "version":
		return buf
	}

	// Color the key in blue to make them easy to spot and visualy parse.
	buf = append(buf, ' ')
	buf = append(buf, []byte(fmt.Sprintf("\x1b[%dm", color))...)
	buf = append(buf, []byte(attr.Key)...)
	buf = append(buf, []byte(fmt.Sprintf("\x1b[%dm", Reset))...)
	buf = append(buf, '=')

	// Format the value depending on its kind. Most types want a simple string
	// conversion, but some may be more practical with a special handling.
	switch attr.Value.Kind() {
	case slog.KindTime:
		buf = append(buf, []byte(attr.Value.Time().Format(time.RFC3339Nano))...)
	default:
		buf = append(buf, []byte(attr.Value.String())...)
	}

	return buf
}

func (h *TextHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	h2 := *h

	// Create a new copy of the preformatted slice, so we don't have to rely on
	// implicit copy-on-append behavior.
	var current = make([]slog.Attr, len(h.attrs))
	copy(current, h.attrs)
	for _, attr := range attrs {
		current = append(current, slog.Attr{
			Key:   fmt.Sprintf("%s%s", h.groups, attr.Key),
			Value: attr.Value,
		})
	}
	h2.attrs = current

	return &h2
}

func (h *TextHandler) WithGroup(name string) slog.Handler {
	h2 := *h

	// Create a new copy of the group slice, to avoid any shared-slice issue.
	var groups = make([]byte, len(h.groups))
	copy(groups, h.groups)
	groups = append(groups, []byte(name)...)
	groups = append(groups, '.')
	h2.groups = groups

	return &h2
}
