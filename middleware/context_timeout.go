package middleware

import (
	"context"
	"errors"
	"time"

	"github.com/datumforge/echox"
)

// ContextTimeoutConfig defines the config for ContextTimeout middleware.
type ContextTimeoutConfig struct {
	// Skipper defines a function to skip middleware.
	Skipper Skipper

	// ErrorHandler is a function when error aries in middeware execution.
	ErrorHandler func(c echox.Context, err error) error

	// Timeout configures a timeout for the middleware
	Timeout time.Duration
}

// ContextTimeout returns a middleware which returns error (503 Service Unavailable error) to client
// when underlying method returns context.DeadlineExceeded error.
func ContextTimeout(timeout time.Duration) echox.MiddlewareFunc {
	return ContextTimeoutWithConfig(ContextTimeoutConfig{Timeout: timeout})
}

// ContextTimeoutWithConfig returns a Timeout middleware with config.
func ContextTimeoutWithConfig(config ContextTimeoutConfig) echox.MiddlewareFunc {
	return toMiddlewareOrPanic(config)
}

// ToMiddleware converts Config to middleware.
func (config ContextTimeoutConfig) ToMiddleware() (echox.MiddlewareFunc, error) {
	if config.Timeout == 0 {
		return nil, errors.New("timeout must be set")
	}

	if config.Skipper == nil {
		config.Skipper = DefaultSkipper
	}

	if config.ErrorHandler == nil {
		config.ErrorHandler = func(c echox.Context, err error) error {
			if err != nil && errors.Is(err, context.DeadlineExceeded) {
				return echox.ErrServiceUnavailable.WithInternal(err)
			}

			return err
		}
	}

	return func(next echox.HandlerFunc) echox.HandlerFunc {
		return func(c echox.Context) error {
			if config.Skipper(c) {
				return next(c)
			}

			timeoutContext, cancel := context.WithTimeout(c.Request().Context(), config.Timeout)
			defer cancel()

			c.SetRequest(c.Request().WithContext(timeoutContext))

			if err := next(c); err != nil {
				return config.ErrorHandler(c, err)
			}

			return nil
		}
	}, nil
}
