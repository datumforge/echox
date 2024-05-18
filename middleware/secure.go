package middleware

import (
	"fmt"

	"github.com/datumforge/echox"
)

// SecureConfig defines the config for Secure middleware.
type SecureConfig struct {
	// Skipper defines a function to skip middleware.
	Skipper Skipper

	// XSSProtection provides protection against cross-site scripting attack (XSS)
	// by setting the `X-XSS-Protection` header.
	// Optional. Default value "1; mode=block".
	XSSProtection string

	// ContentTypeNosniff provides protection against overriding Content-Type
	// header by setting the `X-Content-Type-Options` header.
	// Optional. Default value "nosniff".
	ContentTypeNosniff string

	// XFrameOptions can be used to indicate whether or not a browser should
	// be allowed to render a page in a <frame>, <iframe> or <object> .
	// Sites can use this to avoid clickjacking attacks, by ensuring that their
	// content is not embedded into other sites.provides protection against
	// clickjacking.
	// Optional. Default value "SAMEORIGIN".
	// Possible values:
	// - "SAMEORIGIN" - The page can only be displayed in a frame on the same origin as the page itself.
	// - "DENY" - The page cannot be displayed in a frame, regardless of the site attempting to do so.
	// - "ALLOW-FROM uri" - The page can only be displayed in a frame on the specified origin.
	XFrameOptions string

	// HSTSMaxAge sets the `Strict-Transport-Security` header to indicate how
	// long (in seconds) browsers should remember that this site is only to
	// be accessed using HTTPS. This reduces your exposure to some SSL-stripping
	// man-in-the-middle (MITM) attacks.
	// Optional. Default value 0.
	HSTSMaxAge int

	// HSTSExcludeSubdomains won't include subdomains tag in the `Strict Transport Security`
	// header, excluding all subdomains from security policy. It has no effect
	// unless HSTSMaxAge is set to a non-zero value.
	// Optional. Default value false.
	HSTSExcludeSubdomains bool

	// ContentSecurityPolicy sets the `Content-Security-Policy` header providing
	// security against cross-site scripting (XSS), clickjacking and other code
	// injection attacks resulting from execution of malicious content in the
	// trusted web page context.
	// Optional. Default value "".
	ContentSecurityPolicy string

	// CSPReportOnly would use the `Content-Security-Policy-Report-Only` header instead
	// of the `Content-Security-Policy` header. This allows iterative updates of the
	// content security policy by only reporting the violations that would
	// have occurred instead of blocking the resource.
	// Optional. Default value false.
	CSPReportOnly bool

	// HSTSPreloadEnabled will add the preload tag in the `Strict Transport Security`
	// header, which enables the domain to be included in the HSTS preload list
	// maintained by Chrome (and used by Firefox and Safari): https://hstspreload.org/
	// Optional.  Default value false.
	HSTSPreloadEnabled bool

	// ReferrerPolicy sets the `Referrer-Policy` header providing security against
	// leaking potentially sensitive request paths to third parties.
	// Optional. Default value "".
	ReferrerPolicy string
}

// DefaultSecureConfig is the default Secure middleware config.
var DefaultSecureConfig = SecureConfig{
	Skipper:            DefaultSkipper,
	XSSProtection:      "1; mode=block",
	ContentTypeNosniff: "nosniff",
	XFrameOptions:      "SAMEORIGIN",
	HSTSPreloadEnabled: false,
}

// Secure returns a Secure middleware.
// Secure middleware provides protection against cross-site scripting (XSS) attack,
// content type sniffing, clickjacking, insecure connection and other code injection
// attacks.
func Secure() echox.MiddlewareFunc {
	return SecureWithConfig(DefaultSecureConfig)
}

// SecureWithConfig returns a Secure middleware with config or panics on invalid configuration.
func SecureWithConfig(config SecureConfig) echox.MiddlewareFunc {
	return toMiddlewareOrPanic(config)
}

// ToMiddleware converts SecureConfig to middleware or returns an error for invalid configuration
func (config SecureConfig) ToMiddleware() (echox.MiddlewareFunc, error) {
	// Defaults
	if config.Skipper == nil {
		config.Skipper = DefaultSecureConfig.Skipper
	}

	return func(next echox.HandlerFunc) echox.HandlerFunc {
		return func(c echox.Context) error {
			if config.Skipper(c) {
				return next(c)
			}

			req := c.Request()
			res := c.Response()

			if config.XSSProtection != "" {
				res.Header().Set(echox.HeaderXXSSProtection, config.XSSProtection)
			}

			if config.ContentTypeNosniff != "" {
				res.Header().Set(echox.HeaderXContentTypeOptions, config.ContentTypeNosniff)
			}

			if config.XFrameOptions != "" {
				res.Header().Set(echox.HeaderXFrameOptions, config.XFrameOptions)
			}

			if (c.IsTLS() || (req.Header.Get(echox.HeaderXForwardedProto) == "https")) && config.HSTSMaxAge != 0 {
				subdomains := ""
				if !config.HSTSExcludeSubdomains {
					subdomains = "; includeSubdomains"
				}

				if config.HSTSPreloadEnabled {
					subdomains = fmt.Sprintf("%s; preload", subdomains)
				}

				res.Header().Set(echox.HeaderStrictTransportSecurity, fmt.Sprintf("max-age=%d%s", config.HSTSMaxAge, subdomains))
			}

			if config.ContentSecurityPolicy != "" {
				if config.CSPReportOnly {
					res.Header().Set(echox.HeaderContentSecurityPolicyReportOnly, config.ContentSecurityPolicy)
				} else {
					res.Header().Set(echox.HeaderContentSecurityPolicy, config.ContentSecurityPolicy)
				}
			}

			if config.ReferrerPolicy != "" {
				res.Header().Set(echox.HeaderReferrerPolicy, config.ReferrerPolicy)
			}

			return next(c)
		}
	}, nil
}
