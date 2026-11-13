# Go Core Router

Go Core Router is a small HTTP routing library built on Go's `net/http.ServeMux`. It adds method-based route helpers, route groups with prefixes, and configurable middleware while keeping handlers compatible with `net/http`.

## Features

- Register `GET`, `POST`, `PUT`, `DELETE`, `PATCH`, `HEAD`, and `OPTIONS` routes.
- Register routes with `Handle` and receive errors for invalid or conflicting patterns.
- Group routes with a shared prefix and a snapshot of the parent's middleware configuration.
- Add or exclude middleware at router, group, or endpoint level.
- Use Go `ServeMux` patterns, including wildcard patterns supported by the Go version in use.

## Install

```bash
go get github.com/RasyidHakim-2804/go-core-router
```

## Quick start

```go
package main

import (
    "log"
    "net/http"

    core "github.com/RasyidHakim-2804/go-core-router"
)

func main() {
    router := core.NewRouter()
    router.Get("/hello", func(w http.ResponseWriter, r *http.Request) {
        _, _ = w.Write([]byte("Hello world"))
    })

    log.Fatal(http.ListenAndServe(":8000", router))
}
```

Then visit `http://localhost:8000/hello`. Passing `router` directly as the handler freezes router configuration on the first request, including requests that do not match a route.

`GetMux()` is also available when integrating with an existing `ServeMux` or server setup:

```go
server := &http.Server{
    Addr:    ":8000",
    Handler: router.GetMux(),
}
log.Fatal(server.ListenAndServe())
```

When serving the mux directly, configuration freezes when a registered route is reached. Prefer finishing all router and endpoint configuration before starting the server.

## Middleware

A middleware implements `Before` and `After`. Embedding `core.Middleware` supplies default methods that return `true`:

```go
type AuthMiddleware struct {
    core.Middleware
}

func (m AuthMiddleware) Before(w http.ResponseWriter, r *http.Request) bool {
    if r.Header.Get("Authorization") == "" {
        http.Error(w, "unauthorized", http.StatusUnauthorized)
        return false
    }
    return true
}
```

Register middleware on the router, then exclude it from a specific endpoint:

```go
router.Middlewares(AuthMiddleware{})
router.Get("/private", privateHandler)
router.Get("/health", healthHandler).ExceptMiddlewares(AuthMiddleware{})
```

`Before` runs in registration order. Once a middleware's `Before` returns `true`, its `After` runs after downstream middleware and the handler, in reverse order. If a later `Before` returns `false`, earlier middleware still receive `After`; the rejecting middleware does not. `After`'s boolean result is currently ignored. Panics propagate, while deferred `After` calls still run for middleware that passed `Before`.

Comparable middleware values are identified by value; pointers use pointer identity. Non-comparable middleware values are identified by concrete type. Registering an existing identity updates/re-enables its entry rather than adding a second execution. Wrap middleware in `NamedMiddleware` to distinguish multiple instances or when excluding a specific instance:

```go
logging := core.NamedMiddleware{Name: "logging", MiddlewareInterface: LoggingMiddleware{}}
metrics := core.NamedMiddleware{Name: "metrics", MiddlewareInterface: LoggingMiddleware{}}
router.Middlewares(logging, metrics)
router.Get("/health", healthHandler).ExceptMiddlewares(logging)
```

Names must be non-empty and unique within a router's middleware configuration. Nil middleware values, including typed nil pointers, are invalid and panic during configuration.

Middleware instances are shallow-copied into groups and registered endpoints. Pointer instances are therefore shared across requests and must be safe for concurrent use.

## Groups and prefixes

A group inherits a snapshot of its parent's prefix and middleware configuration. Later changes to the parent do not alter the group, and group configuration does not alter the parent. Routes registered through a group use the same underlying mux.

```go
router.Group(func(api *core.Router) {
    api.Prefix("/api")
    api.Middlewares(AuthMiddleware{})
    api.Get("/users", usersHandler) // GET /api/users
})
```

Paths are joined with a leading slash and repeated slashes are collapsed. A trailing slash on an endpoint path is preserved because it has special subtree meaning in `ServeMux`. Prefixes are normalized without a trailing slash. Prefix changes apply to routes registered afterward. Route variables and wildcard syntax follow `http.ServeMux`; this library does not add Laravel-style `:id` parameters.

## Registration errors

Convenience methods such as `Get` panic with registration context when a route is invalid or conflicts with an existing pattern. Use `Handle` to handle registration errors explicitly:

```go
endpoint, err := router.Handle("GET", "/users/{id}", userHandler)
if err != nil {
    log.Fatal(err)
}
_ = endpoint
```

`Handle` rejects nil handlers and invalid methods. Middleware/configuration changes after serving starts panic with a descriptive message. Configure the router before starting the HTTP server.

## Example application

A complete example is available in [`example/main.go`](example/main.go). Run it from the repository root with:

```bash
go run ./example
```
