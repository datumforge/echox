package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/datumforge/echox"
)

func TestRequestID(t *testing.T) {
	e := echox.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	handler := func(c echox.Context) error {
		return c.String(http.StatusOK, "test")
	}

	rid := RequestID()
	h := rid(handler)
	err := h(c)
	assert.NoError(t, err)
	assert.Len(t, rec.Header().Get(echox.HeaderXRequestID), 32)
}

func TestMustRequestIDWithConfig_skipper(t *testing.T) {
	e := echox.New()
	e.GET("/", func(c echox.Context) error {
		return c.String(http.StatusTeapot, "test")
	})

	generatorCalled := false

	e.Use(RequestIDWithConfig(RequestIDConfig{
		Skipper: func(c echox.Context) bool {
			return true
		},
		Generator: func() string {
			generatorCalled = true
			return "customGenerator"
		},
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	res := httptest.NewRecorder()
	e.ServeHTTP(res, req)

	assert.Equal(t, http.StatusTeapot, res.Code)
	assert.Equal(t, "test", res.Body.String())

	assert.Equal(t, res.Header().Get(echox.HeaderXRequestID), "")
	assert.False(t, generatorCalled)
}

func TestMustRequestIDWithConfig_customGenerator(t *testing.T) {
	e := echox.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	handler := func(c echox.Context) error {
		return c.String(http.StatusOK, "test")
	}

	rid := RequestIDWithConfig(RequestIDConfig{
		Generator: func() string { return "customGenerator" },
	})
	h := rid(handler)
	err := h(c)
	assert.NoError(t, err)
	assert.Equal(t, rec.Header().Get(echox.HeaderXRequestID), "customGenerator")
}

func TestMustRequestIDWithConfig_RequestIDHandler(t *testing.T) {
	e := echox.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	handler := func(c echox.Context) error {
		return c.String(http.StatusOK, "test")
	}

	called := false
	rid := RequestIDWithConfig(RequestIDConfig{
		Generator: func() string { return "customGenerator" },
		RequestIDHandler: func(c echox.Context, s string) {
			called = true
		},
	})
	h := rid(handler)
	err := h(c)
	assert.NoError(t, err)
	assert.Equal(t, rec.Header().Get(echox.HeaderXRequestID), "customGenerator")
	assert.True(t, called)
}

func TestRequestIDWithConfig(t *testing.T) {
	e := echox.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	handler := func(c echox.Context) error {
		return c.String(http.StatusOK, "test")
	}

	rid, err := RequestIDConfig{}.ToMiddleware()
	assert.NoError(t, err)

	h := rid(handler)
	h(c)
	assert.Len(t, rec.Header().Get(echox.HeaderXRequestID), 32)

	// Custom generator
	rid = RequestIDWithConfig(RequestIDConfig{
		Generator: func() string { return "customGenerator" },
	})
	h = rid(handler)
	h(c)
	assert.Equal(t, rec.Header().Get(echox.HeaderXRequestID), "customGenerator")
}

func TestRequestID_IDNotAltered(t *testing.T) {
	e := echox.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Add(echox.HeaderXRequestID, "<sample-request-id>")

	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	handler := func(c echox.Context) error {
		return c.String(http.StatusOK, "test")
	}

	rid := RequestIDWithConfig(RequestIDConfig{})
	h := rid(handler)
	_ = h(c)

	assert.Equal(t, rec.Header().Get(echox.HeaderXRequestID), "<sample-request-id>")
}

func TestRequestIDConfigDifferentHeader(t *testing.T) {
	e := echox.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	handler := func(c echox.Context) error {
		return c.String(http.StatusOK, "test")
	}

	rid := RequestIDWithConfig(RequestIDConfig{TargetHeader: echox.HeaderXCorrelationID})
	h := rid(handler)
	h(c)
	assert.Len(t, rec.Header().Get(echox.HeaderXCorrelationID), 32)

	// Custom generator and handler
	customID := "customGenerator"
	calledHandler := false
	rid = RequestIDWithConfig(RequestIDConfig{
		Generator:    func() string { return customID },
		TargetHeader: echox.HeaderXCorrelationID,
		RequestIDHandler: func(_ echox.Context, id string) {
			calledHandler = true
			assert.Equal(t, customID, id)
		},
	})
	h = rid(handler)
	h(c)
	assert.Equal(t, rec.Header().Get(echox.HeaderXCorrelationID), "customGenerator")
	assert.True(t, calledHandler)
}
