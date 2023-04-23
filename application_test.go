// SPDX-FileCopyrightText: 2023 Iván Szkiba
//
// SPDX-License-Identifier: MIT

package muxpress_test

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/dop251/goja"
	req "github.com/imroc/req/v3"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/szkiba/muxpress"
)

// This example starts a server and listens for connections on a randomly assigned free port. The application responds with "Hello World!" for requests to the root URL.
// In this example, the name of the constructor is `WebApp`, but the name you use is up to you.
func ExampleNewApplicationConstructor() {
	const SCRIPT = `
	const app = new WebApp()

	app.get("/", (req, res) => {
			res.text("Hello World!")
	})

	app.listen()

	app.port // goja runtime returns the last evaluated expression
`

	runtime := goja.New()

	ctor, err := muxpress.NewApplicationConstructor(runtime)
	if err != nil {
		panic(err)
	}

	err = runtime.Set("WebApp", ctor)
	if err != nil {
		panic(err)
	}

	port, err := runtime.RunScript("example", SCRIPT)
	if err != nil {
		panic(err)
	}

	message := req.MustGet("http://localhost:" + port.String())

	fmt.Println(message)

	// output:
	// Hello World!
}

// In this example every log entry come from muxpress runtime will contains a `source` field with value `script`.
func ExampleNewApplicationConstructor_withLogger() {
	runtime := goja.New()

	logger := logrus.StandardLogger().WithField("source", "script")

	ctor, err := muxpress.NewApplicationConstructor(runtime, muxpress.WithLogger(logger))
	if err != nil {
		panic(err)
	}

	err = runtime.Set("WebApp", ctor)
	if err != nil {
		panic(err)
	}

	// output:
}

// In this example custom context will passed to muxpress runtime. All http server will be stopped when context canceled.
func ExampleNewApplicationConstructor_withContext() {
	runtime := goja.New()

	ctx, cancel := context.WithTimeout(context.TODO(), time.Second)
	defer cancel()

	getContext := func() context.Context { return ctx }

	ctor, err := muxpress.NewApplicationConstructor(runtime, muxpress.WithContext(getContext))
	if err != nil {
		panic(err)
	}

	err = runtime.Set("WebApp", ctor)
	if err != nil {
		panic(err)
	}

	// output:
}

func TestDeclaration(t *testing.T) {
	t.Parallel()

	expected, err := os.ReadFile("api/index.d.ts")

	assert.NoError(t, err)
	assert.Equal(t, expected, muxpress.Declarations)
}
