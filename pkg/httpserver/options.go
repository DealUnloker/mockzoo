package httpserver

import (
	"net"
	"time"

	"github.com/gofiber/fiber/v2"
)

// Option -.
type Option func(*Server)

// Port -.
func Port(port string) Option {
	return func(s *Server) {
		s.address = net.JoinHostPort("", port)
	}
}

// Prefork -.
func Prefork(prefork bool) Option {
	return func(s *Server) {
		s.prefork = prefork
	}
}

// ReadTimeout -.
func ReadTimeout(timeout time.Duration) Option {
	return func(s *Server) {
		s.readTimeout = timeout
	}
}

// WriteTimeout -.
func WriteTimeout(timeout time.Duration) Option {
	return func(s *Server) {
		s.writeTimeout = timeout
	}
}

// ShutdownTimeout -.
func ShutdownTimeout(timeout time.Duration) Option {
	return func(s *Server) {
		s.shutdownTimeout = timeout
	}
}

// ErrorHandler sets a custom fiber.ErrorHandler, e.g. to render errors that
// escape route handlers (unmatched routes, wrong methods) in an application-
// specific envelope instead of Fiber's plain-text defaults.
func ErrorHandler(handler fiber.ErrorHandler) Option {
	return func(s *Server) {
		s.errorHandler = handler
	}
}
