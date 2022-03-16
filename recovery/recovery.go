package recovery

import (
	"fmt"
	"os"
	"runtime/debug"
	"strings"

	"github.com/socialpoint-labs/bsk/logx"
)

type options struct {
	exitFunc func()
}

type Options func(*options)

func Handler(l logx.Logger, opts ...Options) func() {
	o := &options{
		exitFunc: func() { os.Exit(2) },
	}

	for _, opt := range opts {
		opt(o)
	}

	return func() {
		if r := recover(); r != nil {
			l.Error(fmt.Sprintf("%v", r), logx.F("stack_trace", strings.Split(string(debug.Stack()), "\n")))
			o.exitFunc()
		}
	}
}

func WithExitFunction(exitFunc func()) Options {
	return func(o *options) {
		o.exitFunc = exitFunc
	}
}
