// SPDX-FileCopyrightText: 2023 Iván Szkiba
//
// SPDX-License-Identifier: MIT

package muxpress

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/dop251/goja"
)

func wrapResponseWriter(runtime *goja.Runtime, from http.ResponseWriter) *goja.Object {
	return wrapResponse(runtime, newResponse(runtime, from))
}

func wrapResponse(runtime *goja.Runtime, resp *response) *goja.Object {
	this := runtime.NewObject()

	mustSet(runtime, this, "json", resp.json)
	mustSet(runtime, this, "text", resp.textf)
	mustSet(runtime, this, "html", resp.html)
	mustSet(runtime, this, "binary", resp.binary)
	mustSet(runtime, this, "send", resp.send)
	mustSet(runtime, this, "status", resp.status)
	mustSet(runtime, this, "type", resp.contentType)
	mustSet(runtime, this, "vary", resp.vary)
	mustSet(runtime, this, "set", resp.set)
	mustSet(runtime, this, "append", resp.append)
	mustSet(runtime, this, "redirect", resp.redirect)

	return this
}

type response struct {
	http.ResponseWriter
	runtime *goja.Runtime
}

func newResponse(runtime *goja.Runtime, writer http.ResponseWriter) *response {
	return &response{ResponseWriter: writer, runtime: runtime}
}

func (resp *response) json(v interface{}) {
	resp.Header().Set("Content-Type", "application/json; charset=utf-8")

	b, err := json.Marshal(v)

	must(resp.runtime, err)

	_, err = resp.Write(b)

	must(resp.runtime, err)
}

func (resp *response) textf(format string, v ...interface{}) {
	resp.Header().Set("Content-Type", "text/plain; charset=utf-8")

	_, err := resp.Write([]byte(fmt.Sprintf(format, v...)))

	must(resp.runtime, err)
}

func (resp *response) html(b []byte) {
	resp.Header().Set("Content-Type", "text/html; charset=utf-8")

	_, err := resp.Write(b)

	must(resp.runtime, err)
}

func (resp *response) binary(b []byte) {
	resp.Header().Set("Content-Type", "application/octet-stream")

	_, err := resp.Write(b)

	must(resp.runtime, err)
}

func (resp *response) send(data interface{}) {
	switch val := data.(type) {
	case string:
		resp.html([]byte(val))
	case []byte:
		resp.binary(val)
	default:
		resp.json(data)
	}
}

func (resp *response) status(code int) {
	resp.WriteHeader(code)
}

func (resp *response) contentType(mime string) {
	resp.Header().Set("Content-Type", mime)
}

func (resp *response) vary(header string) {
	resp.Header().Set("Vary", header)
}

func (resp *response) set(field, value string) {
	resp.Header().Set(field, value)
}

func (resp *response) append(field string, value string) {
	resp.Header().Add(field, value)
}

func (resp *response) redirect(code int, loc string) {
	resp.WriteHeader(code)
	resp.Header().Set("Location", loc)
}
