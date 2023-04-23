// SPDX-FileCopyrightText: 2023 Iván Szkiba
//
// SPDX-License-Identifier: MIT

package muxpress

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
	"sync"

	"github.com/dop251/goja"
	"github.com/julienschmidt/httprouter"
)

func wrapRequest(runtime *goja.Runtime, from *http.Request) *goja.Object {
	req := newRequest(runtime, from)
	this := runtime.NewObject()

	mustSetGetter(runtime, this, "host", req.host)
	mustSetGetter(runtime, this, "method", req.method)
	mustSetGetter(runtime, this, "path", req.path)
	mustSetGetter(runtime, this, "protocol", req.protocol)
	mustSetGetter(runtime, this, "params", req.params)
	mustSetGetter(runtime, this, "query", req.query)
	mustSetGetter(runtime, this, "cookies", req.cookies)
	mustSetGetter(runtime, this, "body", req.body)

	mustSet(runtime, this, "get", req.get)

	return this
}

func (req *request) get(field string) string {
	return req.Header.Get(field)
}

func (req *request) host() string {
	return req.Host
}

func (req *request) method() string {
	return req.Method
}

func (req *request) path() string {
	return req.URL.Path
}

func (req *request) protocol() string {
	return req.URL.Scheme
}

type request struct {
	*http.Request
	runtime *goja.Runtime

	paramsOnce sync.Once
	paramsObj  *goja.Object

	queryOnce sync.Once
	queryObj  *goja.Object

	cookiesOnce sync.Once
	cookiesObj  *goja.Object

	bodyOnce  sync.Once
	bodyValue goja.Value
}

func newRequest(runtime *goja.Runtime, req *http.Request) *request {
	return &request{Request: req, runtime: runtime} //nolint:exhaustruct
}

func (req *request) params() *goja.Object {
	req.paramsOnce.Do(func() {
		req.paramsObj = wrapParams(req.runtime, httprouter.ParamsFromContext(req.Context()))
	})

	return req.paramsObj
}

func (req *request) query() *goja.Object {
	req.queryOnce.Do(func() {
		req.queryObj = wrapValues(req.runtime, req.URL.Query())
	})

	return req.queryObj
}

func (req *request) cookies() *goja.Object {
	req.cookiesOnce.Do(func() {
		req.cookiesObj = wrapCookies(req.runtime, req.Cookies())
	})

	return req.cookiesObj
}

func (req *request) body() goja.Value {
	req.bodyOnce.Do(func() {
		req.bodyValue = wrapBody(req.runtime, req.Request)
	})

	return req.bodyValue
}

func wrapValues(runtime *goja.Runtime, values url.Values) *goja.Object {
	out := runtime.NewObject()

	if len(values) == 0 {
		return out
	}

	for key, value := range values {
		if len(value) == 1 {
			mustSet(runtime, out, key, value[0])
		} else {
			all := []interface{}{}

			for _, str := range value {
				all = append(all, str)
			}

			mustSet(runtime, out, key, runtime.NewArray(all...))
		}
	}

	return out
}

func wrapParams(runtime *goja.Runtime, params httprouter.Params) *goja.Object {
	out := runtime.NewObject()

	for _, param := range params {
		mustSet(runtime, out, param.Key, param.Value)
	}

	return out
}

func wrapCookies(runtime *goja.Runtime, cookies []*http.Cookie) *goja.Object {
	out := runtime.NewObject()

	for _, c := range cookies {
		mustSet(runtime, out, c.Name, c.Value)
	}

	return out
}

func wrapBody(runtime *goja.Runtime, req *http.Request) goja.Value {
	if req.ContentLength == 0 || !strings.HasPrefix(req.Header.Get("Content-Type"), "application/json") {
		return goja.Undefined()
	}

	defer req.Body.Close()

	bin, err := ioutil.ReadAll(req.Body)
	if err != nil {
		throw(runtime, err)
	}

	out := map[string]interface{}{}

	must(runtime, json.Unmarshal(bin, &out))

	return runtime.ToValue(out)
}
