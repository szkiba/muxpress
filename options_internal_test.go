// SPDX-FileCopyrightText: 2023 Iván Szkiba
//
// SPDX-License-Identifier: MIT

package muxpress

import (
	"context"
	"reflect"
	"runtime"
	"testing"

	"github.com/dop251/goja"
	"github.com/sirupsen/logrus"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
)

func assertRunnerFuncEqual(t *testing.T, expected, actual RunnerFunc) {
	t.Helper()

	expectedName := runtime.FuncForPC(reflect.ValueOf(expected).Pointer()).Name()
	actualName := runtime.FuncForPC(reflect.ValueOf(actual).Pointer()).Name()

	assert.Equal(t, expectedName, actualName)
}

func Test_With(t *testing.T) {
	t.Parallel()

	opts := new(options)

	WithContext(context.TODO)(opts)
	assert.Equal(t, context.TODO(), opts.context())

	fs := afero.NewOsFs()

	WithFS(fs)(opts)
	assert.Equal(t, fs, opts.filesystem)

	logger := logrus.StandardLogger().WithField("foo", "bar")

	WithLogger(logger)(opts)
	assert.Equal(t, logger, opts.logger)

	runner := RunnerFunc(func(func() error) {})

	WithRunner(runner)(opts)
	assertRunnerFuncEqual(t, runner, opts.runner)

	opts.runner = nil
	WithRunOnLoop(func(f func(*goja.Runtime)) {})(opts)
	assert.NotNil(t, opts.runner)
}

func Test_getopts(t *testing.T) {
	t.Parallel()

	opts, err := getopts()

	assert.NoError(t, err)
	assert.NotNil(t, opts)
	assert.NotNil(t, opts.context)
	assert.NotNil(t, opts.filesystem)
	assert.NotNil(t, opts.logger)
	assert.NotNil(t, opts.runner)

	filesystem := afero.NewOsFs()
	logger := logrus.StandardLogger().WithField("foo", "bar")
	runner := RunnerFunc(func(func() error) {})

	opts, err = getopts(WithContext(context.TODO), WithFS(filesystem), WithLogger(logger), WithRunner(runner))

	assert.NoError(t, err)
	assert.NotNil(t, opts)

	assert.Equal(t, context.TODO(), opts.context())
	assert.Equal(t, filesystem, opts.filesystem)
	assert.Equal(t, logger, opts.logger)
	assertRunnerFuncEqual(t, runner, opts.runner)
}
