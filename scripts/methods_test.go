// SPDX-FileCopyrightText: 2023 Iván Szkiba
//
// SPDX-License-Identifier: MIT

package scripts_test

import "testing"

func TestMethods(t *testing.T) {
	t.Parallel()
	js(t, `
// js
const app = new Application()

const ping = (req, res)  => res.status(200)

app.get('/',ping)
app.head('/',ping)
app.post('/',ping)
app.put('/',ping)
app.patch('/',ping)
app.delete('/',ping)
app.options('/',ping)

app.listen(() => {
	client.SetBaseURL('http://' + app.host)
})

const R = () => client.R()

test('get', () => {	assert.Equal(200, R().Get('/').GetStatusCode()) })
test('head', () => { assert.Equal(200, R().Head('/').GetStatusCode()) })
test('post', () => { assert.Equal(200, R().Post('/').GetStatusCode()) })
test('put', () => { assert.Equal(200, R().Post('/').GetStatusCode()) })
test('patch', () => {assert.Equal(200, R().Post('/').GetStatusCode()) })
test('delete', () => { assert.Equal(200, R().Post('/').GetStatusCode()) })
test('options', () => { assert.Equal(200, R().Post('/').GetStatusCode()) })

// !js
`)
}
