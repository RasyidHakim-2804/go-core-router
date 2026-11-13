package core

import "net/http"

// Middleware wraps the next HTTP handler. Call next.ServeHTTP to continue,
// or return without calling next to reject the request. Use defer for cleanup.
type Middleware = func(http.Handler) http.Handler

// NamedMiddleware gives a middleware stable identity for exclusion/replacement.
// Use values of this type in Middlewares; exclude them by Name.
type NamedMiddleware struct {
	Name       string
	Middleware Middleware
}
