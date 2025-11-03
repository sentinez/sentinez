package httpxbase

import (
	"context"
	"time"
)

// Server is the interface that provides the basic methods for an HTTP server.
type Server interface {
	ListenAndServe(addr string) error
	Shutdown(ctx context.Context) error
}

// Context is the interface that wraps the basic methods for an HTTP context.
// It provides methods to handle HTTP requests and responses.
type Context interface {
	Release()
	Path() string
	Time() time.Time
	Context() context.Context
	String(statusCode int, body string) error
	JSON(statusCode int, body []byte) error
}

type RequestHandler func(ctx Context) error
