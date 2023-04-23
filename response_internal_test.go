// SPDX-FileCopyrightText: 2023 Iván Szkiba
//
// SPDX-License-Identifier: MIT

package muxpress

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/dop251/goja"
	"github.com/stretchr/testify/assert"
)

func Test_response_json(t *testing.T) {
	t.Parallel()

	runtime := goja.New()
	rec := httptest.NewRecorder()
	res := newResponse(runtime, rec)
	obj := wrapResponse(runtime, res)
	value := runtime.ToValue

	data := map[string]interface{}{
		"foo":    "bar",
		"answer": 42.0,
	}

	callMethod(t, obj, "json", value(data))
	res.json(data)

	assert.Equal(t, "application/json; charset=utf-8", rec.Header().Get("content-type"))

	got := map[string]interface{}{}

	assert.NoError(t, json.NewDecoder(rec.Body).Decode(&got))
	assert.Equal(t, data, got)
}

func Test_response_text(t *testing.T) {
	t.Parallel()

	runtime := goja.New()
	rec := httptest.NewRecorder()
	res := newResponse(runtime, rec)
	obj := wrapResponse(runtime, res)
	value := runtime.ToValue

	callMethod(t, obj, "text", value("Hello, %s!"), value("World"))

	assert.Equal(t, "text/plain; charset=utf-8", rec.Header().Get("content-type"))

	got, err := io.ReadAll(rec.Body)

	assert.NoError(t, err)
	assert.Equal(t, "Hello, World!", string(got))
}

func Test_response_html(t *testing.T) {
	t.Parallel()

	runtime := goja.New()
	rec := httptest.NewRecorder()
	res := newResponse(runtime, rec)
	obj := wrapResponse(runtime, res)
	value := runtime.ToValue

	callMethod(t, obj, "html", value([]byte("<html></html>")))

	assert.Equal(t, "text/html; charset=utf-8", rec.Header().Get("content-type"))

	got, err := io.ReadAll(rec.Body)

	assert.NoError(t, err)
	assert.Equal(t, "<html></html>", string(got))
}

func Test_response_binary(t *testing.T) {
	t.Parallel()

	runtime := goja.New()
	rec := httptest.NewRecorder()
	res := newResponse(runtime, rec)
	obj := wrapResponse(runtime, res)
	value := runtime.ToValue

	data := []byte{1, 2, 3, 4, 5}

	callMethod(t, obj, "binary", value(data))

	assert.Equal(t, "application/octet-stream", rec.Header().Get("content-type"))

	got, err := io.ReadAll(rec.Body)

	assert.NoError(t, err)
	assert.Equal(t, data, got)
}

func Test_response_send(t *testing.T) {
	t.Parallel()

	runtime := goja.New()
	rec := httptest.NewRecorder()
	res := newResponse(runtime, rec)
	obj := wrapResponse(runtime, res)
	value := runtime.ToValue

	callMethod(t, obj, "send", value([]byte{1, 2, 3, 4, 5}))

	assert.Equal(t, "application/octet-stream", rec.Header().Get("content-type"))

	rec = httptest.NewRecorder()
	res = newResponse(runtime, rec)
	obj = wrapResponse(runtime, res)

	callMethod(t, obj, "send", value("<html></html>"))

	assert.Equal(t, "text/html; charset=utf-8", rec.Header().Get("content-type"))

	rec = httptest.NewRecorder()
	res = newResponse(runtime, rec)
	obj = wrapResponse(runtime, res)

	callMethod(t, obj, "send", value(map[string]string{"foo": "bar"}))

	assert.Equal(t, "application/json; charset=utf-8", rec.Header().Get("content-type"))
}

func Test_response_contentType(t *testing.T) {
	t.Parallel()

	runtime := goja.New()
	rec := httptest.NewRecorder()
	res := newResponse(runtime, rec)
	obj := wrapResponse(runtime, res)
	value := runtime.ToValue

	assert.Empty(t, rec.Header().Get("content-type"))
	callMethod(t, obj, "type", value("text/plain"))
	assert.Equal(t, "text/plain", rec.Header().Get("content-type"))
}

func Test_response_vary(t *testing.T) {
	t.Parallel()

	runtime := goja.New()
	rec := httptest.NewRecorder()
	res := newResponse(runtime, rec)
	obj := wrapResponse(runtime, res)
	value := runtime.ToValue

	assert.Empty(t, rec.Header().Get("vary"))
	callMethod(t, obj, "vary", value("user-agent"))
	assert.Equal(t, "user-agent", rec.Header().Get("vary"))
}

func Test_response_redirect(t *testing.T) {
	t.Parallel()

	runtime := goja.New()
	rec := httptest.NewRecorder()
	res := newResponse(runtime, rec)
	obj := wrapResponse(runtime, res)
	value := runtime.ToValue

	assert.Empty(t, rec.Header().Get("location"))
	callMethod(t, obj, "redirect", value(http.StatusPermanentRedirect), value("http://example.com"))
	assert.Equal(t, http.StatusPermanentRedirect, rec.Code)
	assert.Equal(t, "http://example.com", rec.Header().Get("location"))
}

func Test_response_set(t *testing.T) {
	t.Parallel()

	runtime := goja.New()
	rec := httptest.NewRecorder()
	res := newResponse(runtime, rec)
	obj := wrapResponse(runtime, res)
	value := runtime.ToValue

	assert.Empty(t, rec.Header().Get("foo"))
	callMethod(t, obj, "set", value("foo"), value("bar"))
	assert.Equal(t, "bar", rec.Header().Get("foo"))
	assert.Equal(t, []string{"bar"}, rec.Header().Values("foo"))
}

func Test_response_append(t *testing.T) {
	t.Parallel()

	runtime := goja.New()
	rec := httptest.NewRecorder()
	res := newResponse(runtime, rec)
	obj := wrapResponse(runtime, res)
	value := runtime.ToValue

	assert.Empty(t, rec.Header().Get("foo"))
	callMethod(t, obj, "append", value("foo"), value("bar"))
	assert.Equal(t, "bar", rec.Header().Get("foo"))
	assert.Equal(t, []string{"bar"}, rec.Header().Values("foo"))
	callMethod(t, obj, "append", value("foo"), value("dummy"))
	assert.Equal(t, []string{"bar", "dummy"}, rec.Header().Values("foo"))
}

func Test_response_status(t *testing.T) {
	t.Parallel()

	runtime := goja.New()
	rec := httptest.NewRecorder()
	res := newResponse(runtime, rec)
	obj := wrapResponse(runtime, res)
	value := runtime.ToValue

	assert.Equal(t, http.StatusOK, rec.Code)
	callMethod(t, obj, "status", value(http.StatusBadRequest))
	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func callMethod(t *testing.T, this *goja.Object, name string, args ...goja.Value) goja.Value {
	t.Helper()

	val := this.Get(name)

	assert.False(t, goja.IsNull(val))
	assert.False(t, goja.IsUndefined(val))

	call, ok := goja.AssertFunction(val)

	assert.Truef(t, ok, "property %s should be a method", name)

	ret, err := call(this, args...)

	assert.NoError(t, err)

	return ret
}

func Test_wrap_responseWriter(t *testing.T) {
	t.Parallel()

	runtime := goja.New()
	rec := httptest.NewRecorder()
	obj := wrapResponseWriter(runtime, rec)
	value := runtime.ToValue

	callMethod(t, obj, "status", value(http.StatusMovedPermanently))
	assert.Equal(t, http.StatusMovedPermanently, rec.Code)
}
