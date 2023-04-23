// SPDX-FileCopyrightText: 2023 Iván Szkiba
//
// SPDX-License-Identifier: MIT

package muxpress

import (
	_ "embed"
	"net"
	"net/http"
	"reflect"
	"strconv"
	"strings"

	"github.com/dop251/goja"
)

var httpMethods = []string{http.MethodGet, http.MethodHead, http.MethodOptions, http.MethodPost, http.MethodPut, http.MethodPatch, http.MethodDelete}

// NewApplicationConstructor creates an application constructor function. The returned constructor function is ready for use and assignable to any name in a given [goja.Runtime].
// This allow to decide the name of the JavaScript constructor.
// You can pass [Option] parameters to customize the muxpress runtime behavior.
func NewApplicationConstructor(runtime *goja.Runtime, option ...Option) (func(call goja.ConstructorCall) *goja.Object, error) {
	opts, err := getopts(option...)
	if err != nil {
		return nil, err
	}

	return func(call goja.ConstructorCall) *goja.Object {
		this := call.This
		app := newApplication(opts)

		for _, method := range httpMethods {
			mustSet(runtime, this, strings.ToLower(method), app.handlerFor(runtime, strings.ToUpper(method)))
		}

		mustSet(runtime, this, "static", app.static)

		mustSet(runtime, this, "use", app.use)
		mustSet(runtime, this, "listen", app.listen)
		mustSet(runtime, this, "shutdown", app.shutdown)

		mustSetGetter(runtime, this, "host", app.host)
		mustSetGetter(runtime, this, "hostname", app.hostname)
		mustSetGetter(runtime, this, "port", app.port)

		return this
	}, nil
}

type address struct {
	host     string
	hostname string
	port     int
}

type application struct {
	*router
	server  *server
	address *address
}

func newApplication(opts *options) *application {
	app := new(application)

	app.router = newRouter(opts.runner, opts.filesystem)
	app.server = newServer(opts.context, opts.logger)

	return app
}

func (app *application) listen(call goja.FunctionCall, runtime *goja.Runtime) goja.Value { // nolint:ireturn
	args := call.Arguments
	idx := 0
	addr := new(address)

	if len(args) > idx && args[idx].ExportType().Kind() == reflect.Int64 {
		addr.port = int(args[idx].ToInteger())
		idx++
	}

	if len(args) > idx && args[idx].ExportType().Kind() == reflect.String {
		addr.hostname = args[idx].String()
		idx++
	}

	addr.host = net.JoinHostPort(addr.hostname, strconv.Itoa(addr.port))

	tcp, err := app.server.listenAndServe(addr.host, app.router)

	must(runtime, err)

	if addr.port == 0 {
		addr.port = tcp.Port
	}

	if len(addr.hostname) == 0 {
		addr.hostname = defaultHost
	}

	addr.host = net.JoinHostPort(addr.hostname, strconv.Itoa(addr.port))

	app.address = addr

	if len(args) > idx {
		if callback, ok := goja.AssertFunction(args[idx]); ok {
			app.runner(func() error {
				_, err := callback(runtime.GlobalObject())

				return err
			})
		}
	}

	return nil
}

func (app *application) host(_ goja.FunctionCall, runtime *goja.Runtime) goja.Value {
	if app.address == nil {
		return goja.Null()
	}

	return runtime.ToValue(app.address.host)
}

func (app *application) hostname(_ goja.FunctionCall, runtime *goja.Runtime) goja.Value {
	if app.address == nil {
		return goja.Null()
	}

	return runtime.ToValue(app.address.hostname)
}

func (app *application) port(_ goja.FunctionCall, runtime *goja.Runtime) goja.Value {
	if app.address == nil {
		return goja.Null()
	}

	return runtime.ToValue(app.address.port)
}

func (app *application) shutdown(_ goja.FunctionCall, runtime *goja.Runtime) goja.Value { // nolint:ireturn
	app.server.shutdown()

	return goja.Undefined()
}

func (app *application) handlerFor(runtime *goja.Runtime, method string) func(goja.FunctionCall) goja.Value {
	return func(call goja.FunctionCall) goja.Value {
		args := call.Arguments
		idx := 0

		var path string

		if len(args) > idx {
			path = call.Argument(idx).String()

			idx++
		}

		middlewares := []middleware{}

		for _, arg := range args[idx:] {
			var m middleware

			must(runtime, runtime.ExportTo(arg, &m))

			middlewares = append(middlewares, m)
		}

		app.handleMethod(runtime, method, path, middlewares...)

		return goja.Undefined()
	}
}

func (app *application) static(call goja.FunctionCall, runtime *goja.Runtime) goja.Value {
	args := call.Arguments
	idx := 0

	if len(args) <= idx {
		throwf(runtime, "missing path parameter")
	}

	path := call.Argument(idx).String()

	idx++

	if len(args) <= idx {
		throwf(runtime, "missing docroot parameter")
	}

	docroot := call.Argument(idx).String()

	app.router.static(path, docroot)

	return goja.Undefined()
}

const defaultHost = "localhost"

//go:embed api/index.d.ts

// Declarations holds TypeScript declaration file contents for muxpress types.
var Declarations []byte
