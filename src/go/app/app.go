package app

import (
	"context"
	"io"
	"os"
	"os/signal"
	"reflect"
	"syscall"
)

// Those variables are injected at compilation to mark the software version.
var (
	Name    string
	Version string
)

var ctx context.Context

// Context returns the application root context. This context will be
// terminated on receiving an appropriate signal like SIGINT or SIGTERM.
func Context() context.Context {
	return ctx
}

func init() {
	var cancel func()
	ctx, cancel = context.WithCancel(context.Background())

	go func() {
		done := make(chan os.Signal, 1)
		signal.Notify(done, syscall.SIGINT, syscall.SIGTERM)
		<-done
		cancel()
	}()
}

// Cleanup the given struct by closing all fields implementing io.Closer. This
// function isn't recursive and doesn't return any error, as it is mainly used
// for cleaning up when shutting down the services and we can't do anything
// with it anyway. If you need robust and provably correct shutting down
// behavior, do it yourself.
func Cleanup(s interface{}) {
	v := reflect.ValueOf(s)
	for v.Kind() == reflect.Ptr {
		v = v.Elem()
	}

	var typeCloser = reflect.TypeOf((*io.Closer)(nil)).Elem()
	for i := 0; i < v.NumField(); i++ {
		f := v.Field(i)
		if !f.Type().Implements(typeCloser) {
			continue
		}

		_ = f.Interface().(io.Closer).Close()
	}
}
