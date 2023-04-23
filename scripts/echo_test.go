// SPDX-FileCopyrightText: 2023 Iván Szkiba
//
// SPDX-License-Identifier: MIT

package scripts_test

import "testing"

func TestEcho(t *testing.T) {
	t.Parallel()
	js(t, `
// js
const app = new Application()

app.get('/echo', (req, res) => {
	const data = {
		cookies: req.cookies,
		query: req.query,
	}

	res.json(data)
})

app.post('/echo', (req, res) => {
	const data = {
		cookies: req.cookies,
		query: req.query,
		body: req.body,
	}

	res.json(data)
})

app.listen(() => {
	client.SetBaseURL('http://' + app.host)
})

test('query', () => {
	const resp = client.R().Get('/echo?foo=bar&answer=42')
	assert.Equal(200, resp.GetStatusCode())
	const data = JSON.parse(resp.ToString())
	assert.Equal('bar', data.query.foo)
	assert.Equal("42", data.query.answer)
})

test('body', () => {
	const resp = client.R().SetBodyJsonMarshal({foo:"bar", answer:42}).Post('/echo')
	assert.Equal(200, resp.GetStatusCode())
	const data = JSON.parse(resp.ToString())
	assert.Equal('bar', data.body.foo)
	assert.Equal(42, data.body.answer)
})

// !js
`)
}
