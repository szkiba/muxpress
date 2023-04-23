// SPDX-FileCopyrightText: 2023 Iván Szkiba
//
// SPDX-License-Identifier: MIT

package muxpress_test

import (
	"fmt"

	"github.com/dop251/goja"
	req "github.com/imroc/req/v3"
	"github.com/szkiba/muxpress"
)

// This example starts a server and listens for connections on a randomly assigned free port. The application responds with "Hello World!" for requests to the root URL. All other routes are answered with a 404 not found message.
// The allocated port is accessed via the `app.port`` property. The example takes advantage of the fact that goja returns the value of the last expression evaluated.
//
//
// In this example, the name of the constructor is `Application`, but the name you use is up to you.
func Example_random() {
	const SCRIPT = `
	const app = new Application()

	app.get("/", (req, res) => {
		res.text("Hello World!")
	})

	app.listen()

	app.port // goja runtime returns the last evaluated expression
  `
	runtime := goja.New()
	ctor, _ := muxpress.NewApplicationConstructor(runtime)

	runtime.Set("Application", ctor)

	port, _ := runtime.RunScript("example", SCRIPT)
	location := fmt.Sprintf("http://127.0.0.1:%d", port.ToInteger())

	message := req.MustGet(location).String()

	fmt.Println(message)

	// output: Hello World!
}
