// SPDX-FileCopyrightText: 2023 Iván Szkiba
//
// SPDX-License-Identifier: MIT

package scripts_test

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/dop251/goja"
	"github.com/imroc/req/v3"
	"github.com/stretchr/testify/assert"
	"github.com/szkiba/muxpress"
)

func testFunction(t *testing.T, runtime *goja.Runtime) func(string, goja.Callable) {
	t.Helper()

	return func(name string, code goja.Callable) {
		t.Run(name, func(t *testing.T) {
			savedAssert := runtime.Get("assert")
			savedConsole := runtime.Get("console")

			assert.NoError(t, runtime.Set("assert", assert.New(t)))
			assert.NoError(t, runtime.Set("console", newConsole(t, runtime)))

			_, err := code(runtime.GlobalObject())

			assert.NoError(t, err)
			assert.NoError(t, runtime.Set("console", savedConsole))
			assert.NoError(t, runtime.Set("assert", savedAssert))
		})
	}
}

func js(t *testing.T, script string) {
	t.Helper()

	runtime := goja.New()
	ctor, err := muxpress.NewApplicationConstructor(runtime)

	assert.NoError(t, err)
	assert.NoError(t, runtime.Set("console", newConsole(t, runtime)))
	assert.NoError(t, runtime.Set("Application", ctor))
	assert.NoError(t, runtime.Set("assert", assert.New(t)))
	assert.NoError(t, runtime.Set("client", req.NewClient()))
	assert.NoError(t, runtime.Set("test", testFunction(t, runtime)))

	prog, err := goja.Compile(t.Name(), script, true)

	assert.NoError(t, err, "JavaScript syntax error")

	_, err = runtime.RunProgram(prog)

	assert.NoError(t, err)
}

type console struct {
	t *testing.T
}

func newConsole(t *testing.T, runtime *goja.Runtime) *goja.Object {
	t.Helper()

	con := &console{t}

	this := runtime.NewObject()

	assert.NoError(t, this.Set("log", con.log))
	assert.NoError(t, this.Set("debug", con.log))
	assert.NoError(t, this.Set("info", con.log))
	assert.NoError(t, this.Set("warn", con.log))
	assert.NoError(t, this.Set("error", con.logerror))

	return this
}

func (c console) log(args ...goja.Value) {
	c.t.Log(c.format(args...))
}

func (c console) logerror(args ...goja.Value) {
	c.t.Error(c.format(args...))
}

func (c console) format(args ...goja.Value) string {
	var strs strings.Builder

	for i := 0; i < len(args); i++ {
		if i > 0 {
			strs.WriteString(" ")
		}

		strs.WriteString(c.formatValue(args[i]))
	}

	return strs.String()
}

func (c console) formatValue(value goja.Value) string {
	marshaler, ok := value.(json.Marshaler)
	if !ok {
		return value.String()
	}

	bin, err := json.Marshal(marshaler)
	if err != nil {
		return value.String()
	}

	return string(bin)
}
