package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/datumforge/echox"
)

func TestSecure(t *testing.T) {
	e := echox.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	h := func(c echox.Context) error {
		return c.String(http.StatusOK, "test")
	}

	// Default
	err := Secure()(h)(c)
	assert.NoError(t, err)

	assert.Equal(t, "1; mode=block", rec.Header().Get(echox.HeaderXXSSProtection))
	assert.Equal(t, "nosniff", rec.Header().Get(echox.HeaderXContentTypeOptions))
	assert.Equal(t, "SAMEORIGIN", rec.Header().Get(echox.HeaderXFrameOptions))
	assert.Equal(t, "", rec.Header().Get(echox.HeaderStrictTransportSecurity))
	assert.Equal(t, "", rec.Header().Get(echox.HeaderContentSecurityPolicy))
	assert.Equal(t, "", rec.Header().Get(echox.HeaderReferrerPolicy))
}

func TestSecureWithConfig(t *testing.T) {
	e := echox.New()
	h := func(c echox.Context) error {
		return c.String(http.StatusOK, "test")
	}

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set(echox.HeaderXForwardedProto, "https")

	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	mw, err := SecureConfig{
		XSSProtection:         "",
		ContentTypeNosniff:    "",
		XFrameOptions:         "",
		HSTSMaxAge:            3600,
		ContentSecurityPolicy: "default-src 'self'",
		ReferrerPolicy:        "origin",
	}.ToMiddleware()
	assert.NoError(t, err)

	err = mw(h)(c)
	assert.NoError(t, err)

	assert.Equal(t, "", rec.Header().Get(echox.HeaderXXSSProtection))
	assert.Equal(t, "", rec.Header().Get(echox.HeaderXContentTypeOptions))
	assert.Equal(t, "", rec.Header().Get(echox.HeaderXFrameOptions))
	assert.Equal(t, "max-age=3600; includeSubdomains", rec.Header().Get(echox.HeaderStrictTransportSecurity))
	assert.Equal(t, "default-src 'self'", rec.Header().Get(echox.HeaderContentSecurityPolicy))
	assert.Equal(t, "", rec.Header().Get(echox.HeaderContentSecurityPolicyReportOnly))
	assert.Equal(t, "origin", rec.Header().Get(echox.HeaderReferrerPolicy))
}

func TestSecureWithConfig_CSPReportOnly(t *testing.T) {
	// Custom with CSPReportOnly flag
	e := echox.New()
	h := func(c echox.Context) error {
		return c.String(http.StatusOK, "test")
	}

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set(echox.HeaderXForwardedProto, "https")

	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := SecureWithConfig(SecureConfig{
		XSSProtection:         "",
		ContentTypeNosniff:    "",
		XFrameOptions:         "",
		HSTSMaxAge:            3600,
		ContentSecurityPolicy: "default-src 'self'",
		CSPReportOnly:         true,
		ReferrerPolicy:        "origin",
	})(h)(c)
	assert.NoError(t, err)

	assert.Equal(t, "", rec.Header().Get(echox.HeaderXXSSProtection))
	assert.Equal(t, "", rec.Header().Get(echox.HeaderXContentTypeOptions))
	assert.Equal(t, "", rec.Header().Get(echox.HeaderXFrameOptions))
	assert.Equal(t, "max-age=3600; includeSubdomains", rec.Header().Get(echox.HeaderStrictTransportSecurity))
	assert.Equal(t, "default-src 'self'", rec.Header().Get(echox.HeaderContentSecurityPolicyReportOnly))
	assert.Equal(t, "", rec.Header().Get(echox.HeaderContentSecurityPolicy))
	assert.Equal(t, "origin", rec.Header().Get(echox.HeaderReferrerPolicy))
}

func TestSecureWithConfig_HSTSPreloadEnabled(t *testing.T) {
	// Custom with CSPReportOnly flag
	e := echox.New()
	h := func(c echox.Context) error {
		return c.String(http.StatusOK, "test")
	}

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	// Custom, with preload option enabled
	req.Header.Set(echox.HeaderXForwardedProto, "https")

	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := SecureWithConfig(SecureConfig{
		HSTSMaxAge:         3600,
		HSTSPreloadEnabled: true,
	})(h)(c)
	assert.NoError(t, err)

	assert.Equal(t, "max-age=3600; includeSubdomains; preload", rec.Header().Get(echox.HeaderStrictTransportSecurity))
}

func TestSecureWithConfig_HSTSExcludeSubdomains(t *testing.T) {
	// Custom with CSPReportOnly flag
	e := echox.New()
	h := func(c echox.Context) error {
		return c.String(http.StatusOK, "test")
	}

	req := httptest.NewRequest(http.MethodGet, "/", nil)

	// Custom, with preload option enabled and subdomains excluded
	req.Header.Set(echox.HeaderXForwardedProto, "https")

	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := SecureWithConfig(SecureConfig{
		HSTSMaxAge:            3600,
		HSTSPreloadEnabled:    true,
		HSTSExcludeSubdomains: true,
	})(h)(c)
	assert.NoError(t, err)

	assert.Equal(t, "max-age=3600; preload", rec.Header().Get(echox.HeaderStrictTransportSecurity))
}
