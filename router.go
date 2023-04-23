// SPDX-FileCopyrightText: 2023 Iván Szkiba
//
// SPDX-License-Identifier: MIT

package muxpress

import (
	"net/http"
	"strings"

	"github.com/dop251/goja"
	"github.com/julienschmidt/httprouter"
	"github.com/spf13/afero"
)

type middleware func(req *goja.Object, res *goja.Object, next goja.Callable)

type middlewareChain []middleware

func (chain middlewareChain) callOne(req *goja.Object, res *goja.Object, mware middleware) bool {
	nextCalled := false

	mware(req, res, func(this goja.Value, args ...goja.Value) (goja.Value, error) {
		nextCalled = true

		return goja.Undefined(), nil
	})

	return nextCalled
}

func (chain middlewareChain) call(req *goja.Object, res *goja.Object, cascade ...middleware) {
	all := make([]middleware, len(chain)+len(cascade))

	copy(all, chain)
	copy(all[len(chain):], cascade)

	for _, mware := range all {
		if !chain.callOne(req, res, mware) {
			break
		}
	}
}

type router struct {
	*httprouter.Router
	runner RunnerFunc

	middlewares middlewareChain
	filesystem  afero.Fs
}

func newRouter(runner RunnerFunc, filesystem afero.Fs) *router {
	return &router{
		Router:      httprouter.New(),
		runner:      runner,
		filesystem:  filesystem,
		middlewares: make(middlewareChain, 0),
	}
}

func (r *router) use(middlewares ...middleware) {
	r.middlewares = append(r.middlewares, middlewares...)
}

func (r *router) runSync(fn func() error) {
	done := make(chan struct{}, 1)

	r.runner(func() error {
		err := fn()

		done <- struct{}{}

		return err
	})

	<-done
}

func (r *router) handle(runtime *goja.Runtime, response http.ResponseWriter, request *http.Request, middlewares ...middleware) {
	r.runSync(func() error {
		req := wrapRequest(runtime, request)
		res := wrapResponseWriter(runtime, response)
		r.middlewares.call(req, res, middlewares...)

		return nil
	})
}

func (r *router) handleMethod(runtime *goja.Runtime, method string, path string, middlewares ...middleware) {
	r.Router.HandlerFunc(method, path, func(response http.ResponseWriter, request *http.Request) {
		r.handle(runtime, response, request, middlewares...)
	})
}

func (r *router) fixpath(path string) string {
	if strings.HasSuffix(path, "/*filepath") {
		return path
	}

	if !strings.HasSuffix(path, "/") {
		path += "/"
	}

	return path + "*filepath"
}

func (r *router) static(path string, docroot string) {
	fs := afero.NewHttpFs(afero.NewBasePathFs(r.filesystem, docroot))
	fileserver := http.FileServer(fs)

	r.Router.HandlerFunc(http.MethodGet, r.fixpath(path), func(response http.ResponseWriter, request *http.Request) {
		params := httprouter.ParamsFromContext(request.Context())

		request.URL.Path = params.ByName("filepath")

		r.runSync(func() error {
			fileserver.ServeHTTP(response, request)

			return nil
		})
	})
}
