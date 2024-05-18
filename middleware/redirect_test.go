package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/datumforge/echox"
)

type middlewareGenerator func() echox.MiddlewareFunc

func TestRedirectHTTPSRedirect(t *testing.T) {
	var testCases = []struct {
		whenHost         string
		whenHeader       http.Header
		expectLocation   string
		expectStatusCode int
	}{
		{
			whenHost:         "labstack.com",
			expectLocation:   "https://labstack.com/",
			expectStatusCode: http.StatusMovedPermanently,
		},
		{
			whenHost:         "labstack.com",
			whenHeader:       map[string][]string{echox.HeaderXForwardedProto: {"https"}},
			expectLocation:   "",
			expectStatusCode: http.StatusOK,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.whenHost, func(t *testing.T) {
			res := redirectTest(HTTPSRedirect, tc.whenHost, tc.whenHeader)

			assert.Equal(t, tc.expectStatusCode, res.Code)
			assert.Equal(t, tc.expectLocation, res.Header().Get(echox.HeaderLocation))
		})
	}
}

func TestRedirectHTTPSWWWRedirect(t *testing.T) {
	var testCases = []struct {
		whenHost         string
		whenHeader       http.Header
		expectLocation   string
		expectStatusCode int
	}{
		{
			whenHost:         "labstack.com",
			expectLocation:   "https://www.labstack.com/",
			expectStatusCode: http.StatusMovedPermanently,
		},
		{
			whenHost:         "www.labstack.com",
			expectLocation:   "",
			expectStatusCode: http.StatusOK,
		},
		{
			whenHost:         "a.com",
			expectLocation:   "https://www.a.com/",
			expectStatusCode: http.StatusMovedPermanently,
		},
		{
			whenHost:         "ip",
			expectLocation:   "https://www.ip/",
			expectStatusCode: http.StatusMovedPermanently,
		},
		{
			whenHost:         "labstack.com",
			whenHeader:       map[string][]string{echox.HeaderXForwardedProto: {"https"}},
			expectLocation:   "",
			expectStatusCode: http.StatusOK,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.whenHost, func(t *testing.T) {
			res := redirectTest(HTTPSWWWRedirect, tc.whenHost, tc.whenHeader)

			assert.Equal(t, tc.expectStatusCode, res.Code)
			assert.Equal(t, tc.expectLocation, res.Header().Get(echox.HeaderLocation))
		})
	}
}

func TestRedirectHTTPSNonWWWRedirect(t *testing.T) {
	var testCases = []struct {
		whenHost         string
		whenHeader       http.Header
		expectLocation   string
		expectStatusCode int
	}{
		{
			whenHost:         "www.labstack.com",
			expectLocation:   "https://labstack.com/",
			expectStatusCode: http.StatusMovedPermanently,
		},
		{
			whenHost:         "a.com",
			expectLocation:   "https://a.com/",
			expectStatusCode: http.StatusMovedPermanently,
		},
		{
			whenHost:         "ip",
			expectLocation:   "https://ip/",
			expectStatusCode: http.StatusMovedPermanently,
		},
		{
			whenHost:         "www.labstack.com",
			whenHeader:       map[string][]string{echox.HeaderXForwardedProto: {"https"}},
			expectLocation:   "",
			expectStatusCode: http.StatusOK,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.whenHost, func(t *testing.T) {
			res := redirectTest(HTTPSNonWWWRedirect, tc.whenHost, tc.whenHeader)

			assert.Equal(t, tc.expectStatusCode, res.Code)
			assert.Equal(t, tc.expectLocation, res.Header().Get(echox.HeaderLocation))
		})
	}
}

func TestRedirectWWWRedirect(t *testing.T) {
	var testCases = []struct {
		whenHost         string
		whenHeader       http.Header
		expectLocation   string
		expectStatusCode int
	}{
		{
			whenHost:         "labstack.com",
			expectLocation:   "http://www.labstack.com/",
			expectStatusCode: http.StatusMovedPermanently,
		},
		{
			whenHost:         "a.com",
			expectLocation:   "http://www.a.com/",
			expectStatusCode: http.StatusMovedPermanently,
		},
		{
			whenHost:         "ip",
			expectLocation:   "http://www.ip/",
			expectStatusCode: http.StatusMovedPermanently,
		},
		{
			whenHost:         "a.com",
			whenHeader:       map[string][]string{echox.HeaderXForwardedProto: {"https"}},
			expectLocation:   "https://www.a.com/",
			expectStatusCode: http.StatusMovedPermanently,
		},
		{
			whenHost:         "www.ip",
			expectLocation:   "",
			expectStatusCode: http.StatusOK,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.whenHost, func(t *testing.T) {
			res := redirectTest(WWWRedirect, tc.whenHost, tc.whenHeader)

			assert.Equal(t, tc.expectStatusCode, res.Code)
			assert.Equal(t, tc.expectLocation, res.Header().Get(echox.HeaderLocation))
		})
	}
}

func TestRedirectNonWWWRedirect(t *testing.T) {
	var testCases = []struct {
		whenHost         string
		whenHeader       http.Header
		expectLocation   string
		expectStatusCode int
	}{
		{
			whenHost:         "www.labstack.com",
			expectLocation:   "http://labstack.com/",
			expectStatusCode: http.StatusMovedPermanently,
		},
		{
			whenHost:         "www.a.com",
			expectLocation:   "http://a.com/",
			expectStatusCode: http.StatusMovedPermanently,
		},
		{
			whenHost:         "www.a.com",
			whenHeader:       map[string][]string{echox.HeaderXForwardedProto: {"https"}},
			expectLocation:   "https://a.com/",
			expectStatusCode: http.StatusMovedPermanently,
		},
		{
			whenHost:         "ip",
			expectLocation:   "",
			expectStatusCode: http.StatusOK,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.whenHost, func(t *testing.T) {
			res := redirectTest(NonWWWRedirect, tc.whenHost, tc.whenHeader)

			assert.Equal(t, tc.expectStatusCode, res.Code)
			assert.Equal(t, tc.expectLocation, res.Header().Get(echox.HeaderLocation))
		})
	}
}

func TestNonWWWRedirectWithConfig(t *testing.T) {
	var testCases = []struct {
		name             string
		givenCode        int
		givenSkipFunc    func(c echox.Context) bool
		whenHost         string
		whenHeader       http.Header
		expectLocation   string
		expectStatusCode int
	}{
		{
			name:             "usual redirect",
			whenHost:         "www.labstack.com",
			expectLocation:   "http://labstack.com/",
			expectStatusCode: http.StatusMovedPermanently,
		},
		{
			name: "redirect is skipped",
			givenSkipFunc: func(c echox.Context) bool {
				return true // skip always
			},
			whenHost:         "www.labstack.com",
			expectLocation:   "",
			expectStatusCode: http.StatusOK,
		},
		{
			name:             "redirect with custom status code",
			givenCode:        http.StatusSeeOther,
			whenHost:         "www.labstack.com",
			expectLocation:   "http://labstack.com/",
			expectStatusCode: http.StatusSeeOther,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.whenHost, func(t *testing.T) {
			middleware := func() echox.MiddlewareFunc {
				return NonWWWRedirectWithConfig(RedirectConfig{
					Skipper: tc.givenSkipFunc,
					Code:    tc.givenCode,
				})
			}
			res := redirectTest(middleware, tc.whenHost, tc.whenHeader)

			assert.Equal(t, tc.expectStatusCode, res.Code)
			assert.Equal(t, tc.expectLocation, res.Header().Get(echox.HeaderLocation))
		})
	}
}

func redirectTest(fn middlewareGenerator, host string, header http.Header) *httptest.ResponseRecorder {
	e := echox.New()
	next := func(c echox.Context) (err error) {
		return c.NoContent(http.StatusOK)
	}
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Host = host

	if header != nil {
		req.Header = header
	}

	res := httptest.NewRecorder()
	c := e.NewContext(req, res)

	fn()(next)(c)

	return res
}
