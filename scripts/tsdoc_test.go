// SPDX-FileCopyrightText: 2023 Iván Szkiba
//
// SPDX-License-Identifier: MIT

package scripts_test

import "testing"

func TestTsDoc(t *testing.T) {
	t.Parallel()
	js(t, `
// js
const app = new Application()
const port = 0

app.get('/', (req, res) => {
	res.json({message:"Hello World!"})
})

app.listen(port)

client.SetBaseURL('http://' + app.host)

test('hello', () => {
	const resp = client.R().Get('/')
	assert.Equal(200, resp.GetStatusCode())
	const data = JSON.parse(resp.ToString())
	assert.Equal('Hello World!', data.message)
})

// !js
`)
}
