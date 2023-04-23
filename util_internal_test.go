// SPDX-FileCopyrightText: 2023 Iván Szkiba
//
// SPDX-License-Identifier: MIT

package muxpress

import (
	"errors"
	"testing"

	"github.com/dop251/goja"
	"github.com/stretchr/testify/assert"
)

var errFoo = errors.New("foo")

func Test_throw(t *testing.T) {
	t.Parallel()

	runtime := goja.New()

	assert.Panics(t, func() {
		throw(runtime, errFoo)
	})

	ex := new(goja.Exception)

	assert.PanicsWithValue(t, ex, func() {
		throw(runtime, ex)
	})

	assert.Panics(t, func() {
		throwf(runtime, "foo")
	})
}

func Test_must(t *testing.T) {
	t.Parallel()

	runtime := goja.New()

	assert.Panics(t, func() { must(runtime, errFoo) })
	assert.NotPanics(t, func() { must(runtime, nil) })

	obj := runtime.NewObject()

	assert.NotPanics(t, func() { mustSet(runtime, obj, "foo", "bar") })
	assert.Equal(t, "bar", obj.Get("foo").String())

	assert.NotPanics(t, func() { mustSetGetter(runtime, obj, "dynamic", func() string { return "value" }) })
	assert.Equal(t, "value", obj.Get("dynamic").String())
}
