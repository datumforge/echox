package middleware

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/datumforge/echox"
)

func TestMethodOverride(t *testing.T) {
	e := echox.New()
	m := MethodOverride()
	h := func(c echox.Context) error {
		return c.String(http.StatusOK, "test")
	}

	// Override with http header
	req := httptest.NewRequest(http.MethodPost, "/", nil)
	rec := httptest.NewRecorder()

	req.Header.Set(echox.HeaderXHTTPMethodOverride, http.MethodDelete)
	c := e.NewContext(req, rec)

	err := m(h)(c)
	assert.NoError(t, err)

	assert.Equal(t, http.MethodDelete, req.Method)
}

func TestMethodOverride_formParam(t *testing.T) {
	e := echox.New()
	h := func(c echox.Context) error {
		return c.String(http.StatusOK, "test")
	}

	// Override with form parameter
	m, err := MethodOverrideConfig{Getter: MethodFromForm("_method")}.ToMiddleware()
	assert.NoError(t, err)

	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader([]byte("_method="+http.MethodDelete)))
	rec := httptest.NewRecorder()

	req.Header.Set(echox.HeaderContentType, echox.MIMEApplicationForm)
	c := e.NewContext(req, rec)

	err = m(h)(c)
	assert.NoError(t, err)

	assert.Equal(t, http.MethodDelete, req.Method)
}

func TestMethodOverride_queryParam(t *testing.T) {
	e := echox.New()
	h := func(c echox.Context) error {
		return c.String(http.StatusOK, "test")
	}

	// Override with query parameter
	m, err := MethodOverrideConfig{Getter: MethodFromQuery("_method")}.ToMiddleware()
	assert.NoError(t, err)

	req := httptest.NewRequest(http.MethodPost, "/?_method="+http.MethodDelete, nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err = m(h)(c)
	assert.NoError(t, err)

	assert.Equal(t, http.MethodDelete, req.Method)
}

func TestMethodOverride_ignoreGet(t *testing.T) {
	e := echox.New()
	m := MethodOverride()
	h := func(c echox.Context) error {
		return c.String(http.StatusOK, "test")
	}

	// Ignore `GET`
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set(echox.HeaderXHTTPMethodOverride, http.MethodDelete)

	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := m(h)(c)
	assert.NoError(t, err)

	assert.Equal(t, http.MethodGet, req.Method)
}
