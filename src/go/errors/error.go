package errors

import (
	"errors"
	"fmt"
)

var Is = errors.Is
var As = errors.As

type E struct {
	Msg     string
	Err     error
	Context map[string]any
}

func (e E) Error() string {
	if e.Err == nil {
		return e.Msg
	}
	return fmt.Sprintf(`%s: %s`, e.Msg, e.Err.Error())
}

func (e E) Unwrap() error {
	return e.Err
}

func New(msg string, keyvals ...any) error {
	return E{
		Msg:     msg,
		Err:     nil,
		Context: fromKeyvals(keyvals...),
	}
}

func Newf(msg string, args ...any) error {
	return New(fmt.Sprintf(msg, args...))
}

func Wrap(err error, msg string, keyvals ...any) error {
	if err == nil {
		return nil
	}

	return E{
		Msg:     msg,
		Err:     err,
		Context: fromKeyvals(keyvals...),
	}
}

func Wrapf(err error, msg string, args ...any) error {
	return Wrap(err, fmt.Sprintf(msg, args...))
}

func With(err error, keyvals ...any) error {
	var e E
	switch err.(type) {
	case E:
		e = err.(E)
	default:
		e = E{Err: err}
	}

	if e.Context == nil {
		e.Context = fromKeyvals(keyvals...)
		return e
	}

	for k, v := range fromKeyvals(keyvals...) {
		e.Context[k] = v
	}

	return err
}

func fromKeyvals(keyvals ...any) map[string]any {
	if len(keyvals) == 0 {
		return nil
	}

	if len(keyvals)%2 != 0 {
		keyvals = append(keyvals, "")
	}

	res := make(map[string]any, len(keyvals)/2)
	for i := 0; i < len(keyvals); i += 2 {
		var key string
		switch k := keyvals[i].(type) {
		case string:
			key = k
		case interface{ String() string }:
			key = k.String()
		default:
			key = fmt.Sprintf("%v", k)
		}

		res[key] = keyvals[i+1]
	}

	return res
}
