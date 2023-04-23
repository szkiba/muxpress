// SPDX-FileCopyrightText: 2023 Iván Szkiba
//
// SPDX-License-Identifier: MIT

package muxpress

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"runtime"
	"strconv"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func newHelloHandler(t *testing.T) http.Handler {
	t.Helper()

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("content-type", "text/plain")
		w.Write([]byte("Hello, World!")) // nolint:errcheck
	})
}

func serverRequest(t *testing.T, addr *net.TCPAddr, path string) (*http.Response, error) { //nolint:unparam
	t.Helper()

	a := net.JoinHostPort(addr.IP.String(), strconv.Itoa(addr.Port))
	req, err := http.NewRequestWithContext(context.TODO(), http.MethodGet, fmt.Sprintf("http://%s/%s", a, path), nil)

	assert.NoError(t, err)

	return http.DefaultClient.Do(req)
}

func Test_server_listenAndServe(t *testing.T) {
	t.Parallel()

	srv := newServer(context.TODO, logrus.StandardLogger())

	addr, err := srv.listenAndServe("", newHelloHandler(t))

	assert.NoError(t, err)
	assert.Greater(t, addr.Port, 0)

	res, err := serverRequest(t, addr, "/")
	defer func() {
		if err == nil {
			res.Body.Close()
		}
	}()

	assert.NoError(t, err)

	assert.Equal(t, "text/plain", res.Header.Get("content-type"))
}

func Test_server_shutdown(t *testing.T) {
	t.Parallel()

	srv := newServer(context.TODO, logrus.StandardLogger())

	addr, err := srv.listenAndServe("", newHelloHandler(t))

	assert.NoError(t, err)

	res, err := serverRequest(t, addr, "/")

	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, res.StatusCode)

	srv.shutdown()

	runtime.Gosched()

	time.Sleep(100 * time.Microsecond) // XXX: should find a better solution

	res, err = serverRequest(t, addr, "/")
	defer func() {
		if err == nil {
			res.Body.Close()
		}
	}()

	assert.Error(t, err)
}

func Test_server_context_done(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithCancel(context.TODO())

	srv := newServer(func() context.Context { return ctx }, logrus.StandardLogger())

	addr, err := srv.listenAndServe("", newHelloHandler(t))

	assert.NoError(t, err)

	res, err := serverRequest(t, addr, "/")
	defer func() {
		if err == nil {
			res.Body.Close()
		}
	}()

	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, res.StatusCode)

	cancel()

	runtime.Gosched()

	time.Sleep(100 * time.Microsecond) // XXX: should find a better solution

	res, err = serverRequest(t, addr, "/")
	defer func() {
		if err == nil {
			res.Body.Close()
		}
	}()

	assert.Error(t, err)
}

func Test_server_serve_used_port(t *testing.T) {
	t.Parallel()

	srv := newServer(context.TODO, logrus.StandardLogger())

	addr, err := srv.listenAndServe("", newHelloHandler(t))

	assert.NoError(t, err)

	_, err = srv.listenAndServe(":"+strconv.Itoa(addr.Port), newHelloHandler(t))

	assert.Error(t, err)
}
