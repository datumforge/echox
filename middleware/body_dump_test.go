package middleware

import (
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/datumforge/echox"
)

func TestBodyDump(t *testing.T) {
	e := echox.New()
	hw := "Hello, World!"
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(hw))
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	h := func(c echox.Context) error {
		body, err := io.ReadAll(c.Request().Body)
		if err != nil {
			return err
		}

		return c.String(http.StatusOK, string(body))
	}

	requestBody := ""
	responseBody := ""
	mw, err := BodyDumpConfig{Handler: func(c echox.Context, reqBody, resBody []byte) {
		requestBody = string(reqBody)
		responseBody = string(resBody)
	}}.ToMiddleware()
	assert.NoError(t, err)

	if assert.NoError(t, mw(h)(c)) {
		assert.Equal(t, requestBody, hw)
		assert.Equal(t, responseBody, hw)
		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Equal(t, hw, rec.Body.String())
	}
}

func TestBodyDump_skipper(t *testing.T) {
	e := echox.New()

	isCalled := false
	mw, err := BodyDumpConfig{
		Skipper: func(c echox.Context) bool {
			return true
		},
		Handler: func(c echox.Context, reqBody, resBody []byte) {
			isCalled = true
		},
	}.ToMiddleware()
	assert.NoError(t, err)

	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader("{}"))
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	h := func(c echox.Context) error {
		return errors.New("some error")
	}

	err = mw(h)(c)
	assert.EqualError(t, err, "some error")
	assert.False(t, isCalled)
}

func TestBodyDump_fails(t *testing.T) {
	e := echox.New()
	hw := "Hello, World!"
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(hw))
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	h := func(c echox.Context) error {
		return errors.New("some error")
	}

	mw, err := BodyDumpConfig{Handler: func(c echox.Context, reqBody, resBody []byte) {}}.ToMiddleware()
	assert.NoError(t, err)

	err = mw(h)(c)
	assert.EqualError(t, err, "some error")
	assert.Equal(t, http.StatusOK, rec.Code)
}

func TestBodyDumpWithConfig_panic(t *testing.T) {
	assert.Panics(t, func() {
		mw := BodyDumpWithConfig(BodyDumpConfig{
			Skipper: nil,
			Handler: nil,
		})
		assert.NotNil(t, mw)
	})

	assert.NotPanics(t, func() {
		mw := BodyDumpWithConfig(BodyDumpConfig{Handler: func(c echox.Context, reqBody, resBody []byte) {}})
		assert.NotNil(t, mw)
	})
}

func TestBodyDump_panic(t *testing.T) {
	assert.Panics(t, func() {
		mw := BodyDump(nil)
		assert.NotNil(t, mw)
	})

	assert.NotPanics(t, func() {
		BodyDump(func(c echox.Context, reqBody, resBody []byte) {})
	})
}
