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

// This example starts a server and listens for connections on port 3000. The application responds with "Hello World!" for requests to the root URL. All other routes are answered with a 404 not found message.
//
// In this example, the name of the constructor is `Application`, but the name you use is up to you.
func Example_fixed() {
	const SCRIPT = `
	const app = new Application()

	app.get("/", (req, res) => {
		res.text("Hello World!")
	})

	app.listen(3000)
  `
	runtime := goja.New()
	ctor, _ := muxpress.NewApplicationConstructor(runtime)

	runtime.Set("Application", ctor)

	runtime.RunScript("example", SCRIPT)

	message := req.MustGet("http://127.0.0.1:3000").String()

	fmt.Println(message)

	// output: Hello World!
}
