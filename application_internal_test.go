// SPDX-FileCopyrightText: 2023 Iván Szkiba
//
// SPDX-License-Identifier: MIT

package muxpress

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"github.com/dop251/goja"
	"github.com/stretchr/testify/assert"
)

func Test_application_properties(t *testing.T) {
	t.Parallel()

	runtime := goja.New()

	opts, err := getopts()

	assert.NoError(t, err)

	app := newApplication(opts)

	call := goja.FunctionCall{} // nolint:exhaustruct

	assert.Equal(t, goja.Null(), app.host(call, runtime))
	assert.Equal(t, goja.Null(), app.hostname(call, runtime))
	assert.Equal(t, goja.Null(), app.port(call, runtime))

	call.This = runtime.GlobalObject()
	call.Arguments = []goja.Value{}

	app.listen(call, runtime)

	port := app.port(call, runtime).ToInteger()

	assert.NotEmpty(t, port)

	assert.Equal(t, "localhost:"+strconv.Itoa(int(port)), app.host(call, runtime).String())
	assert.Equal(t, "localhost", app.hostname(call, runtime).String())

	app.shutdown(call, runtime)
}

func Test_application_listen(t *testing.T) {
	t.Parallel()

	runtime := goja.New()

	opts, err := getopts()

	assert.NoError(t, err)

	app := newApplication(opts)

	call := goja.FunctionCall{} // nolint:exhaustruct

	call.This = runtime.GlobalObject()

	assert.NotPanics(t, func() { app.listen(call, runtime) })
	app.shutdown(call, runtime)

	callbackCalled := false

	call.Arguments = append(call.Arguments, runtime.ToValue(func() {
		callbackCalled = true
	}))

	assert.NotPanics(t, func() { app.listen(call, runtime) })

	assert.True(t, callbackCalled)

	call.Arguments = nil

	app.shutdown(call, runtime)
}

func Test_application_listen_host(t *testing.T) {
	t.Parallel()

	runtime := goja.New()
	value := runtime.ToValue

	opts, err := getopts()

	assert.NoError(t, err)

	app := newApplication(opts)

	call := goja.FunctionCall{
		This: runtime.GlobalObject(),
		Arguments: []goja.Value{
			value("127.0.0.1"),
		},
	}

	assert.NotPanics(t, func() { app.listen(call, runtime) })

	port := app.port(call, runtime).ToInteger()

	assert.NotEmpty(t, port)

	assert.Equal(t, "127.0.0.1:"+strconv.Itoa(int(port)), app.host(call, runtime).String())
	assert.Equal(t, "127.0.0.1", app.hostname(call, runtime).String())

	app.shutdown(call, runtime)
}

func Test_application_listen_port(t *testing.T) {
	t.Parallel()

	runtime := goja.New()
	value := runtime.ToValue

	opts, err := getopts()

	assert.NoError(t, err)

	app := newApplication(opts)

	call := goja.FunctionCall{
		This: runtime.GlobalObject(),
		Arguments: []goja.Value{
			value(0),
		},
	}

	assert.NotPanics(t, func() { app.listen(call, runtime) })

	// new app on same port should panic
	call.Arguments[0] = value(app.address.port)
	app = newApplication(opts)

	assert.Panics(t, func() { app.listen(call, runtime) })
}

func Test_application_handlerFor(t *testing.T) {
	t.Parallel()

	runtime := goja.New()
	value := runtime.ToValue

	opts, err := getopts()

	assert.NoError(t, err)

	app := newApplication(opts)

	handler := app.handlerFor(runtime, http.MethodGet)

	call := goja.FunctionCall{
		This: runtime.GlobalObject(),
		Arguments: []goja.Value{
			value("/echo"),
			value(newEcho(t, runtime)),
		},
	}

	handler(call)

	call.Arguments = []goja.Value{}

	app.listen(call, runtime)

	url := "http://" + app.address.host + "/echo?message=dummy"
	req, err := http.NewRequestWithContext(context.TODO(), http.MethodGet, url, nil)

	assert.NoError(t, err)

	resp, err := http.DefaultClient.Do(req)
	defer func() {
		if err == nil {
			resp.Body.Close()
		}
	}()

	assert.NoError(t, err)

	got, err := io.ReadAll(resp.Body)

	assert.NoError(t, err)
	assert.Equal(t, "dummy", string(got))

	app.shutdown(call, runtime)
}

func Test_application_handlerFor_panic(t *testing.T) {
	t.Parallel()

	runtime := goja.New()
	value := runtime.ToValue

	opts, err := getopts()

	assert.NoError(t, err)

	app := newApplication(opts)

	handler := app.handlerFor(runtime, http.MethodGet)

	call := goja.FunctionCall{
		This: runtime.GlobalObject(),
		Arguments: []goja.Value{
			value(newEcho(t, runtime)),
		},
	}

	assert.Panics(t, func() { handler(call) })
}

func Test_application_static_panic(t *testing.T) {
	t.Parallel()

	runtime := goja.New()
	value := runtime.ToValue

	opts, err := getopts()

	assert.NoError(t, err)

	app := newApplication(opts)

	call := goja.FunctionCall{
		This:      runtime.GlobalObject(),
		Arguments: []goja.Value{},
	}

	assert.Panics(t, func() { app.static(call, runtime) })

	call.Arguments = []goja.Value{
		value("/foo"),
	}

	assert.Panics(t, func() { app.static(call, runtime) })

	call.Arguments = []goja.Value{
		value("/foo"),
		value("/bar"),
	}

	assert.NotPanics(t, func() { app.static(call, runtime) })
}

func Test_application_static(t *testing.T) {
	t.Parallel()

	runtime := goja.New()
	value := runtime.ToValue

	fs, cleanup := newStaticFs(t)
	defer cleanup()

	opts, err := getopts(WithFS(fs))

	assert.NoError(t, err)

	app := newApplication(opts)

	call := goja.FunctionCall{
		This: runtime.GlobalObject(),
		Arguments: []goja.Value{
			value("/dummy"),
			value("/"),
		},
	}

	assert.NotPanics(t, func() { app.static(call, runtime) })

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/dummy/foo/foo.txt", nil)

	app.router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)

	res := rec.Result()
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)

	assert.NoError(t, err)
	assert.Equal(t, "Hello, World!", string(body))
}

func Test_NewApplicationConstructor(t *testing.T) {
	t.Parallel()

	runtime := goja.New()
	value := runtime.ToValue

	fn, err := NewApplicationConstructor(runtime)

	assert.NoError(t, err)

	assert.NoError(t, runtime.Set("App", fn))

	ctor, ok := goja.AssertConstructor(runtime.Get("App"))

	assert.True(t, ok)

	app, err := ctor(runtime.NewObject())

	assert.NoError(t, err)

	for _, m := range methods {
		callMethod(t, app, m, value("/dummy"), value(newEcho(t, runtime)))
	}

	callMethod(t, app, "listen")

	for _, p := range properties {
		val := app.Get(p)

		assert.False(t, goja.IsNull(val))
		assert.False(t, goja.IsUndefined(val))
	}

	for _, f := range functions {
		_, isFunction := goja.AssertFunction(app.Get(f))

		assert.True(t, isFunction)
	}
}

var (
	methods    = []string{"get", "head", "post", "put", "patch", "delete", "options"}
	properties = []string{"host", "hostname", "port"}
	functions  = []string{"listen", "shutdown", "static", "use"}
)
