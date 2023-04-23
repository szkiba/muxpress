// SPDX-FileCopyrightText: 2023 Iván Szkiba
//
// SPDX-License-Identifier: MIT

package muxpress

import (
	"fmt"

	"github.com/dop251/goja"
)

func throw(runtime *goja.Runtime, err error) {
	if e, ok := err.(*goja.Exception); ok { //nolint:errorlint
		panic(e)
	}

	panic(runtime.NewGoError(err))
}

func throwf(runtime *goja.Runtime, format string, args ...any) {
	throw(runtime, fmt.Errorf(format, args...)) //nolint:goerr113
}

func must(runtime *goja.Runtime, err error) {
	if err != nil {
		throw(runtime, err)
	}
}

func mustSet(runtime *goja.Runtime, obj *goja.Object, name string, value interface{}) {
	must(runtime, obj.Set(name, value))
}

func mustSetGetter(runtime *goja.Runtime, obj *goja.Object, name string, getter interface{}) {
	must(runtime, obj.DefineAccessorProperty(name, runtime.ToValue(getter), goja.Undefined(), goja.FLAG_FALSE, goja.FLAG_TRUE))
}
