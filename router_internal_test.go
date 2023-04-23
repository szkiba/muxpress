// SPDX-FileCopyrightText: 2023 Iván Szkiba
//
// SPDX-License-Identifier: MIT

package muxpress

import (
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/dop251/goja"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
)

func Test_router_fixpath(t *testing.T) {
	t.Parallel()

	router := new(router)

	assert.Equal(t, "/*filepath", router.fixpath(""))
	assert.Equal(t, "/*filepath", router.fixpath("/"))
	assert.Equal(t, "/*filepath", router.fixpath("/*filepath"))
	assert.Equal(t, "/foo/*filepath", router.fixpath("/foo"))
	assert.Equal(t, "/foo/*filepath", router.fixpath("/foo/"))
	assert.Equal(t, "/foo/*filepath", router.fixpath("/foo/*filepath"))
}

func Test_newRouter(t *testing.T) {
	t.Parallel()

	runner := syncRunner()
	filesystem := afero.NewOsFs()
	router := newRouter(runner, filesystem)

	assertRunnerFuncEqual(t, runner, router.runner)
	assert.Equal(t, filesystem, router.filesystem)
	assert.NotNil(t, router.Router)
	assert.NotNil(t, router.middlewares)
}

func Test_router_runSync(t *testing.T) {
	t.Parallel()

	t.Run("asyncRunner", func(t *testing.T) {
		t.Parallel()

		runner := func(fn func() error) {
			go func() {
				time.Sleep(time.Millisecond)

				fn() //nolint:errcheck
			}()
		}

		router := newRouter(runner, afero.NewOsFs())

		var result string

		router.runSync(func() error {
			result = "foo"

			return nil
		})

		assert.Equal(t, "foo", result)
	})

	t.Run("syncRunner", func(t *testing.T) {
		t.Parallel()

		router := newRouter(syncRunner(), afero.NewOsFs())

		var result string

		router.runSync(func() error {
			result = "foo"

			return nil
		})

		assert.Equal(t, "foo", result)
	})
}

func newStaticFs(t *testing.T) (afero.Fs, func()) {
	t.Helper()

	dir, err := os.MkdirTemp("", "*")

	assert.NoError(t, err)

	filesystem := afero.NewBasePathFs(afero.NewOsFs(), dir)

	const mode = 0o755

	assert.NoError(t, os.WriteFile(filepath.Join(dir, "foo.html"), []byte("<html></html>"), mode))
	assert.NoError(t, os.Mkdir(filepath.Join(dir, "foo"), mode))
	assert.NoError(t, os.WriteFile(filepath.Join(dir, "foo", "foo.txt"), []byte("Hello, World!"), mode))

	cleanup := func() {
		assert.NoError(t, os.RemoveAll(dir))
	}

	return filesystem, cleanup
}

func Test_router_static_subdir(t *testing.T) {
	t.Parallel()

	fs, cleanup := newStaticFs(t)
	defer cleanup()

	router := newRouter(syncRunner(), fs)

	router.static("/sub", "/foo") // subdir to path

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/sub/foo.txt", nil)

	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)

	res := rec.Result()
	body, err := io.ReadAll(res.Body)

	defer res.Body.Close()

	assert.NoError(t, err)
	assert.Equal(t, "Hello, World!", string(body))
}

func Test_router_static_root(t *testing.T) {
	t.Parallel()

	fs, cleanup := newStaticFs(t)
	defer cleanup()

	router := newRouter(syncRunner(), fs)

	router.static("/bar", "/")

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/bar/foo.html", nil)

	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)

	res := rec.Result()
	body, err := io.ReadAll(res.Body)

	defer res.Body.Close()

	assert.NoError(t, err)
	assert.Equal(t, "<html></html>", string(body))

	rec = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodGet, "/bar/foo/foo.txt", nil)

	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)

	res = rec.Result()
	body, err = io.ReadAll(res.Body)

	defer res.Body.Close()

	assert.NoError(t, err)
	assert.Equal(t, "Hello, World!", string(body))
}

func newEcho(t *testing.T, runtime *goja.Runtime) middleware {
	t.Helper()

	return func(req *goja.Object, res *goja.Object, next goja.Callable) {
		query, isObject := req.Get("query").(*goja.Object)

		assert.True(t, isObject)

		msg := query.Get("message").String()

		callMethod(t, res, "text", runtime.ToValue(msg))
	}
}

func newAddMagicHeader(t *testing.T, runtime *goja.Runtime) middleware {
	t.Helper()

	return func(req *goja.Object, res *goja.Object, next goja.Callable) {
		callMethod(t, res, "set", runtime.ToValue("magic"), runtime.ToValue("42"))

		_, err := next(runtime.GlobalObject())

		assert.NoError(t, err)
	}
}

func Test_router_handleMethod(t *testing.T) {
	t.Parallel()

	runtime := goja.New()
	router := newRouter(syncRunner(), nil)

	echo := newEcho(t, runtime)

	router.handleMethod(runtime, http.MethodGet, "/echo", echo)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/echo?message=Hello", nil)

	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "text/plain; charset=utf-8", rec.Header().Get("content-type"))

	res := rec.Result()
	body, err := io.ReadAll(res.Body)

	defer res.Body.Close()

	assert.NoError(t, err)
	assert.Equal(t, "Hello", string(body))
}

func Test_router_use(t *testing.T) {
	t.Parallel()

	runtime := goja.New()
	router := newRouter(syncRunner(), nil)

	echo := newEcho(t, runtime)

	router.handleMethod(runtime, http.MethodGet, "/echo", echo)
	router.use(newAddMagicHeader(t, runtime))

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/echo?message=Hello", nil)

	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "text/plain; charset=utf-8", rec.Header().Get("content-type"))
	assert.Equal(t, "42", rec.Header().Get("magic"))
}
