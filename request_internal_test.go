// SPDX-FileCopyrightText: 2023 Iván Szkiba
//
// SPDX-License-Identifier: MIT

package muxpress

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"strings"
	"testing"

	"github.com/dop251/goja"
	"github.com/julienschmidt/httprouter"
	"github.com/stretchr/testify/assert"
)

func Test_request(t *testing.T) {
	t.Parallel()

	runtime := goja.New()

	from := httptest.NewRequest(http.MethodGet, "/", nil)

	req := newRequest(runtime, from)

	assert.NotNil(t, req)
	assert.Nil(t, req.cookiesObj)
	assert.Nil(t, req.queryObj)
	assert.Nil(t, req.paramsObj)
	assert.Nil(t, req.bodyValue)

	assert.NotNil(t, req.cookies())
	assert.NotNil(t, req.cookiesObj)

	assert.NotNil(t, req.query())
	assert.NotNil(t, req.queryObj)

	assert.NotNil(t, req.params())
	assert.NotNil(t, req.paramsObj)

	assert.NotNil(t, req.body())
	assert.NotNil(t, req.bodyValue)
}

func Test_request_body(t *testing.T) {
	t.Parallel()

	runtime := goja.New()

	str := `{"prop-name":"prop-value"}`

	body := strings.NewReader(str)
	from := httptest.NewRequest(http.MethodGet, "/", body)
	from.Header.Add("content-type", "application/json")
	from.Header.Add("content-length", strconv.Itoa(len(str)))

	req := newRequest(runtime, from)

	obj, ok := req.body().(*goja.Object)

	assert.True(t, ok, "body must be object")
	assert.Equal(t, "prop-value", obj.Get("prop-name").String())

	from = httptest.NewRequest(http.MethodGet, "/", nil)

	req = newRequest(runtime, from)

	assert.Equal(t, goja.Undefined(), req.body())
}

func Test_request_cookies(t *testing.T) {
	t.Parallel()

	runtime := goja.New()

	from := httptest.NewRequest(http.MethodGet, "/", nil)
	from.AddCookie(&http.Cookie{Name: "cookie-name", Value: "cookie-value"}) //nolint:exhaustruct

	req := newRequest(runtime, from)

	assert.NotNil(t, req.cookies())
	assert.Equal(t, "cookie-value", req.cookies().Get("cookie-name").String())

	from = httptest.NewRequest(http.MethodGet, "/", nil)

	req = newRequest(runtime, from)
	assert.NotNil(t, req.cookies())
}

func Test_request_query(t *testing.T) {
	t.Parallel()

	runtime := goja.New()

	from := httptest.NewRequest(http.MethodGet, "/?param-name=param-value", nil)
	req := newRequest(runtime, from)

	assert.NotNil(t, req.query())
	assert.Equal(t, "param-value", req.query().Get("param-name").String())

	from = httptest.NewRequest(http.MethodGet, "/", nil)
	req = newRequest(runtime, from)
	assert.NotNil(t, req.query())
}

func Test_request_params(t *testing.T) {
	t.Parallel()

	runtime := goja.New()

	from := httptest.NewRequest(http.MethodGet, "/", nil)
	params := httprouter.Params{httprouter.Param{Key: "param-name", Value: "param-value"}}
	ctx := context.WithValue(context.TODO(), httprouter.ParamsKey, params)

	from = from.WithContext(ctx)
	from.AddCookie(&http.Cookie{Name: "cookie-name", Value: "cookie-value"}) //nolint:exhaustruct

	req := newRequest(runtime, from)

	assert.NotNil(t, req.params())
	assert.Equal(t, "param-value", req.params().Get("param-name").String())

	from = httptest.NewRequest(http.MethodGet, "/", nil)
	req = newRequest(runtime, from)
	assert.NotNil(t, req.params())
}

func Test_wrap_request(t *testing.T) {
	t.Parallel()

	runtime := goja.New()

	str := `{"prop-name":"prop-value"}`

	from := httptest.NewRequest(http.MethodGet, "http://localhost/path/dir?query-name=query-value", strings.NewReader(str))
	from.Header.Add("content-type", "application/json")
	from.Header.Add("content-length", strconv.Itoa(len(str)))
	from.AddCookie(&http.Cookie{Name: "cookie-name", Value: "cookie-value"}) //nolint:exhaustruct

	params := httprouter.Params{httprouter.Param{Key: "param-name", Value: "param-value"}}
	ctx := context.WithValue(context.TODO(), httprouter.ParamsKey, params)

	from = from.WithContext(ctx)

	req := wrapRequest(runtime, from)

	obj, isObject := req.Get("body").(*goja.Object)

	assert.True(t, isObject, "body must be object")
	assert.Equal(t, "prop-value", obj.Get("prop-name").String())

	obj, isObject = req.Get("params").(*goja.Object)

	assert.True(t, isObject, "params must be object")
	assert.Equal(t, "param-value", obj.Get("param-name").String())

	obj, isObject = req.Get("query").(*goja.Object)

	assert.True(t, isObject, "query must be object")
	assert.Equal(t, "query-value", obj.Get("query-name").String())

	obj, isObject = req.Get("cookies").(*goja.Object)

	assert.True(t, isObject, "cookies must be object")
	assert.Equal(t, "cookie-value", obj.Get("cookie-name").String())

	assert.Equal(t, "/path/dir", req.Get("path").String())
	assert.Equal(t, "http", req.Get("protocol").String())
	assert.Equal(t, "GET", req.Get("method").String())
	assert.Equal(t, "localhost", req.Get("host").String())

	var get goja.Callable

	assert.NoError(t, runtime.ExportTo(req.Get("get"), &get))

	value, err := get(req, runtime.ToValue("content-type"))

	assert.NoError(t, err)
	assert.Equal(t, "application/json", value.String())
}

func Test_wrapCookies(t *testing.T) {
	t.Parallel()

	runtime := goja.New()

	cookies := []*http.Cookie{
		{Name: "cookie-name", Value: "cookie-value"},
		{Name: "other-cookie", Value: "other-value"},
	}

	obj := wrapCookies(runtime, cookies)

	assert.NotNil(t, obj)
	assert.Equal(t, "cookie-value", obj.Get("cookie-name").String())
	assert.Equal(t, "other-value", obj.Get("other-cookie").String())

	assert.NotNil(t, wrapCookies(runtime, nil))
	assert.NotNil(t, wrapCookies(runtime, []*http.Cookie{}))
}

func Test_wrapParams(t *testing.T) {
	t.Parallel()

	runtime := goja.New()

	params := httprouter.Params{
		{Key: "param-name", Value: "param-value"},
		{Key: "other-param", Value: "other-value"},
	}

	obj := wrapParams(runtime, params)

	assert.NotNil(t, obj)
	assert.Equal(t, "param-value", obj.Get("param-name").String())
	assert.Equal(t, "other-value", obj.Get("other-param").String())

	assert.NotNil(t, wrapParams(runtime, nil))
	assert.NotNil(t, wrapParams(runtime, httprouter.Params{}))
}

func Test_wrapValues(t *testing.T) {
	t.Parallel()

	runtime := goja.New()

	values := url.Values{
		"query-name":  []string{"query-value"},
		"other-query": []string{"other-value", "another-value"},
	}

	obj := wrapValues(runtime, values)

	assert.NotNil(t, obj)
	assert.Equal(t, "query-value", obj.Get("query-name").String())

	arr, isObject := obj.Get("other-query").(*goja.Object)

	assert.True(t, isObject)
	assert.Equal(t, "other-value", arr.Get("0").String())
	assert.Equal(t, "another-value", arr.Get("1").String())

	assert.NotNil(t, wrapValues(runtime, nil))
	assert.NotNil(t, wrapValues(runtime, url.Values{}))
}

func Test_wrapBody(t *testing.T) {
	t.Parallel()

	runtime := goja.New()

	body := map[string]interface{}{
		"prop-name": "prop-value",
		"nested": map[string]interface{}{
			"nested-name": "nested-value",
		},
	}

	bin, err := json.Marshal(body)

	assert.NoError(t, err)

	from := httptest.NewRequest(http.MethodGet, "/", bytes.NewReader(bin))

	from.Header.Add("content-type", "application/json")
	from.Header.Add("content-length", strconv.Itoa(len(bin)))

	obj, isObject := wrapBody(runtime, from).(*goja.Object)

	assert.True(t, isObject)
	assert.NotNil(t, obj)
	assert.Equal(t, "prop-value", obj.Get("prop-name").String())

	val := obj.Get("nested")

	assert.NotNil(t, val)
	assert.False(t, goja.IsNull(val))
	assert.False(t, goja.IsUndefined(val))

	obj, isObject = val.(*goja.Object)

	assert.True(t, isObject)
	assert.Equal(t, "nested-value", obj.Get("nested-name").String())

	from = httptest.NewRequest(http.MethodGet, "/", nil)

	assert.NotNil(t, wrapParams(runtime, nil))

	val = wrapBody(runtime, from)

	assert.NotNil(t, val)
	assert.True(t, goja.IsUndefined(val))
}

type errReader struct{}

func (errReader) Read(p []byte) (int, error) {
	return 0, errFoo
}

func Test_wrapBody_panic(t *testing.T) {
	t.Parallel()

	runtime := goja.New()

	from := httptest.NewRequest(http.MethodGet, "/", errReader{})

	from.Header.Add("content-type", "application/json")
	from.Header.Add("content-length", "1")

	assert.Panics(t, func() { wrapBody(runtime, from) })
}
