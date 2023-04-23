// SPDX-FileCopyrightText: 2023 Iván Szkiba
//
// SPDX-License-Identifier: MIT

package muxpress

import (
	"context"
	"os"
	"sync"

	"github.com/dop251/goja"
	"github.com/sirupsen/logrus"
	"github.com/spf13/afero"
)

// RunnerFunc is used to execute middlewares on incoming requests.
type RunnerFunc func(func() error)

type options struct {
	runner     RunnerFunc
	logger     logrus.FieldLogger
	filesystem afero.Fs
	context    func() context.Context
}

func getopts(with ...Option) (*options, error) {
	opts := new(options)

	for _, o := range with {
		o(opts)
	}

	if opts.logger == nil {
		opts.logger = logrus.StandardLogger()
	}

	if opts.filesystem == nil {
		cwd, err := os.Getwd()
		if err != nil {
			return nil, err
		}

		opts.filesystem = afero.NewBasePathFs(afero.NewOsFs(), cwd)
	}

	if opts.runner == nil {
		opts.runner = syncRunner()
	}

	if opts.context == nil {
		opts.context = context.TODO
	}

	return opts, nil
}

// Option is an option for the [NewApplicationConstructor] factory function.
type Option = func(*options)

// WithLogger returns an Option that specifies a [logrus.FieldLogger] logger to be used for logging.
func WithLogger(logger logrus.FieldLogger) Option {
	return func(o *options) {
		o.logger = logger
	}
}

// WithFS returns an Option that specifies a [afero.Fs] filesystem to be used for accessing static files.
func WithFS(filesystem afero.Fs) Option {
	return func(o *options) {
		o.filesystem = filesystem
	}
}

// WithContext returns an Option that specifies a [context.Context] getter function to be used for stopping application when context is canceled or done.
// Default is to use [context.TODO].
func WithContext(context func() context.Context) Option {
	return func(o *options) {
		o.context = context
	}
}

// WithRunner returns an Option that specifies a runner function to be used for execute middlewares for incoming requests.
// This option allows you to schedule middleware calls in the event loop.
//
// Since [goja.Runtime] is not goroutine-safe, the default is to execute middlewares in synchronous way.
//
//   func syncRunner() RunnerFunc {
//     var mu sync.Mutex
//
//     return func(fn func() error) {
//       mu.Lock()
//       defer mu.Unlock()
//
//       if err := fn(); err != nil {
//         panic(err)
//       }
//     }
//   }
//nolint:gci,gofmt,gofumpt,goimports
func WithRunner(runner RunnerFunc) Option {
	return func(o *options) {
		o.runner = runner
	}
}

// WithRunOnLoop returns an Option that specifies [RunOnLoop] function from [goja_nodejs] package to be used for execute middlewares for incoming requests.
//
// [RunOnLoop]: https://pkg.go.dev/github.com/dop251/goja_nodejs/eventloop#EventLoop.RunOnLoop
// [goja_nodejs]: https://github.com/dop251/goja_nodejs
func WithRunOnLoop(runOnLoop func(func(*goja.Runtime))) Option {
	return WithRunner(runOnLoopRunner(runOnLoop))
}

func runOnLoopRunner(runOnLoop func(func(*goja.Runtime))) RunnerFunc {
	return func(fn func() error) {
		runOnLoop(func(runtime *goja.Runtime) {
			must(runtime, fn())
		})
	}
}

func syncRunner() RunnerFunc {
	var mu sync.Mutex

	return func(fn func() error) {
		mu.Lock()
		defer mu.Unlock()

		if err := fn(); err != nil {
			panic(err)
		}
	}
}
