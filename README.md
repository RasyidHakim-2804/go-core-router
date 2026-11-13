# Go Core Router

A lightweight Go HTTP router with Laravel-inspired route groups, prefixes, and middleware configuration. It builds on `net/http.ServeMux` and uses standard `http.ResponseWriter` and `*http.Request` handlers.

## Why use this router?

Go's `http.ServeMux` already handles method and path matching, including wildcards. This library keeps that standard router and adds a fluent way to derive route scopes with prefixes and middleware. It is useful when you want grouped route configuration while keeping ordinary `net/http` handlers. If you only need method/path matching, plain `ServeMux` may be enough; this package does not aim to replace full web frameworks.

## Features

- Helpers for `GET`, `POST`, `PUT`, `PATCH`, `DELETE`, `HEAD`, and `OPTIONS`.
- `Handle` for explicit method registration and registration errors.
- Path parameters, catch-all wildcards, and exact-root patterns through `ServeMux`.
- Nested groups with inherited prefixes and middleware configuration.
- Middleware additions, exclusions, and re-enabling at router, group, and endpoint level.
- Named middleware instances and standard HTTP handler wrappers with context propagation.
- Direct `http.Handler` compatibility and access to the underlying mux.

## Contents

- [Why use this router?](#why-use-this-router)
- [Installation and quick start](#installation)
- [Router scopes and chaining](#router-scopes-and-chaining)
- [Registering routes](#registering-routes)
- [Path parameters and wildcards](#path-parameters-and-wildcards)
- [Middleware](#middleware)
- [Groups and prefixes](#groups-and-prefixes)
- [Serving and configuration lifecycle](#serving-and-configuration-lifecycle)
- [Complete example](#running-the-complete-example)
- [License](#license)

## Installation

Requires Go **1.23.1 or newer**, as declared in `go.mod`.

Inside your application's Go module:

```bash
go get github.com/RasyidHakim-2804/go-core-router
```

If starting a new application, first run `go mod init example.com/myapp`.

## Quick start

Save this as `main.go`:

```go
package main

import (
    "fmt"
    "log"
    "net/http"

    core "github.com/RasyidHakim-2804/go-core-router"
)

func main() {
    router := core.NewRouter()
    router.Get("/hello", func(w http.ResponseWriter, r *http.Request) {
        fmt.Fprintln(w, "Hello world")
    })

    log.Fatal(http.ListenAndServe(":8000", router))
}
```

Run `go run .`, then open `http://localhost:8000/hello`.

The snippets below show configuration to complete before starting the server. Handler names such as `listUsers` represent functions supplied by your application.

## Router scopes and chaining

`Prefix`, `Middlewares`, and `ExceptMiddlewares` return a **derived router** with copied prefix and middleware configuration. They do not change the router they were called on. Use the returned router by continuing the chain or assigning it to a variable:

```go
api := router.Prefix("/api").Middlewares(Logging)
api.Get("/users", listUsers) // /api/users; Logging runs first.

router.Get("/health", healthHandler) // /health; the original router is unchanged.
```

Discarding the return value discards that configuration:

```go
router.Prefix("/api")
router.Get("/users", listUsers) // /users, not /api/users.
```

Derived routers and groups share the same route table, so routes registered through any scope are served by the original router. Middleware lists are copied, but middleware functions may close over shared state. Make mutable captured state safe for concurrent requests.

## Registering routes

Handlers have this signature; `core.Handler` is an alias for the same function type:

```go
func handler(w http.ResponseWriter, r *http.Request) {
    fmt.Fprintln(w, "ok")
}
```

All helpers return a `*core.RequestHandler`, allowing endpoint middleware chaining:

```go
router.Get("/users", listUsers)
router.Post("/users", createUser)
router.Put("/users/{id}", replaceUser)
router.Patch("/users/{id}", updateUser)
router.Delete("/users/{id}", deleteUser)
router.Head("/users/{id}", userHeaders)
router.Options("/users/{id}", userOptions)
```

These register handlers only: your handler implements the operation and response. `ServeMux` also lets a GET route match HEAD requests unless a more specific HEAD route is registered. An OPTIONS handler must supply its own response, such as an `Allow` header; registering it does not configure CORS.

Use `Handle` when you want to handle registration errors explicitly or register another HTTP method:

```go
endpoint, err := router.Handle(http.MethodGet, "/status", statusHandler)
if err != nil {
    log.Fatal(err)
}
endpoint.Middlewares(Logging)
```

`Handle` returns errors for nil handlers, invalid methods, and invalid or conflicting patterns. The convenience helpers panic on these registration errors. Configuration after serving starts also panics, including calls to `Handle`.

## Path parameters and wildcards

### One path segment

```go
router.Get("/users/{id}", func(w http.ResponseWriter, r *http.Request) {
    fmt.Fprintf(w, "user=%s\n", r.PathValue("id"))
})
```

`/users/42` gives `id == "42"`. Parameters are strings; parse or validate them in your handler.

### Remaining path

```go
router.Get("/files/{path...}", func(w http.ResponseWriter, r *http.Request) {
    fmt.Fprintf(w, "path=%s\n", r.PathValue("path"))
})
```

`/files/images/logo.png` gives `path == "images/logo.png"`. A catch-all wildcard must be at the end of the pattern. This example only reports the path; it does not serve files.

### Exact root and subtrees

```go
router.Get("/{$}", homeHandler)       // Only /.
router.Get("/docs/", docsHandler)     // /docs/ and paths beneath it.
router.Get("/docs/{$}", indexHandler) // Only /docs/; more specific than /docs/.
```

A trailing slash has subtree meaning, so `"/"` alone is a fallback subtree pattern rather than an exact-root pattern. Routing uses Go `ServeMux` syntax, not Laravel-style `:id` or `*path` syntax. Matching, precedence, redirects, and default 404/405 responses come from `ServeMux`.

## Middleware

Middleware uses the standard Go wrapper signature:

```go
func(next http.Handler) http.Handler
```

`core.Middleware` is an alias for this function type. `Middlewares` accepts plain functions and `core.NamedMiddleware` values. Unsupported values and nil middleware are rejected with a panic during configuration.

```go
func Logging(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        log.Printf("before: %s", r.URL.Path)
        defer log.Printf("after: %s", r.URL.Path)
        next.ServeHTTP(w, r)
    })
}

router.Middlewares(Logging).Get("/users", listUsers)
```

Call `next.ServeHTTP` to continue. Return without calling it to reject a request. Middleware controls cleanup with `defer` and may recover panics around `next.ServeHTTP`. The router does not recover handler panics or automatically run cleanup hooks.

### Passing context to a handler

Use a private key type to avoid context key collisions. This snippet needs `import "context"`:

```go
type userKey struct{}

func DemoGate(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        if r.Header.Get("X-Demo-Key") != "secret" {
            http.Error(w, "unauthorized", http.StatusUnauthorized)
            return
        }
        ctx := context.WithValue(r.Context(), userKey{}, "demo-user")
        next.ServeHTTP(w, r.WithContext(ctx))
    })
}

router.Middlewares(DemoGate).Get("/profile", func(w http.ResponseWriter, r *http.Request) {
    user, ok := r.Context().Value(userKey{}).(string)
    if !ok {
        http.Error(w, "missing user", http.StatusUnauthorized)
        return
    }
    fmt.Fprintln(w, user)
})
```

The fixed key is only a demonstration. Real authentication must validate credentials. Wrappers can also forward a modified request or a wrapped `http.ResponseWriter`, for example to observe status codes or transform responses. A custom writer should preserve any optional HTTP interfaces needed by downstream handlers (such as streaming or hijacking).

### Router, group, and endpoint middleware

`Middlewares` accepts one or more middleware functions or `core.NamedMiddleware` values. It uses `...any` to accept both forms, so unsupported values are rejected with a panic at configuration time rather than by the compiler. A separately defined named function type may need explicit conversion to `core.Middleware`.

```go
api := router.Middlewares(Logging, RequestMetrics)
api.Get("/users", listUsers) // Logging, then RequestMetrics, then handler.

router.Get("/details", detailsHandler).Middlewares(Logging) // Endpoint only.
```

Router middleware methods return derived routers. Endpoint middleware methods modify the endpoint returned by `Get`, `Post`, or `Handle`.

### Grouping middleware for several endpoints

Create a derived router scope with middleware shared by several routes. Here `RequestMetrics` and `RateLimit` are middleware functions supplied by the application. Then add middleware or exclude named middleware on a narrower scope or endpoint:

```go
auth := core.NamedMiddleware{Name: "auth", Middleware: DemoGate}
metrics := core.NamedMiddleware{Name: "metrics", Middleware: RequestMetrics}

api := router.Prefix("/api").Middlewares(Logging, auth, metrics)
api.Get("/users", listUsers) // Logging + auth + metrics.
api.Get("/health", healthHandler).ExceptMiddlewares("auth")

// This subgroup inherits api's middleware, adds rate limiting,
// and excludes metrics for every endpoint in the subgroup.
admin := api.Prefix("/admin").Middlewares(RateLimit).ExceptMiddlewares("metrics")
admin.Get("/reports", reportsHandler) // Logging + auth + RateLimit.
admin.Get("/status", adminStatusHandler).ExceptMiddlewares("auth")

// Creating admin did not change api.
api.Get("/settings", settingsHandler) // Still Logging + auth + metrics.
```

Middleware that needs to be excluded must be registered as `NamedMiddleware`, then excluded by its name. Unnamed middleware such as `Logging` or `RateLimit` can be added but cannot be selected for exclusion.

### Named middleware, exclusions, and re-enabling

Functions are not compared by identity. Every unnamed middleware registration appends a new wrapper, even if the same function was already registered. A name lets you replace, exclude, or re-enable a specific middleware instance. This matters when one factory creates middleware with different configuration:

```go
func LoggingWithLevel(level string) core.Middleware {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            log.Printf("[%s] %s %s", level, r.Method, r.URL.Path)
            next.ServeHTTP(w, r)
        })
    }
}

info := core.NamedMiddleware{
    Name:       "request-info",
    Middleware: LoggingWithLevel("info"),
}
debug := core.NamedMiddleware{
    Name:       "request-debug",
    Middleware: LoggingWithLevel("debug"),
}

router = router.Middlewares(info, debug)
router.Get("/health", healthHandler).ExceptMiddlewares("request-info")
```

Both entries come from `LoggingWithLevel`, but each closure keeps its own level. The health endpoint excludes only the info logger; debug logging remains active. The names are identifiers for configuration, not middleware code. Without names, function identity cannot reliably distinguish closures with different captured settings.

The same mechanism works across router, group, and endpoint scopes:

```go
auth := core.NamedMiddleware{Name: "auth", Middleware: DemoGate}
router = router.Middlewares(auth)
router.Get("/health", healthHandler).ExceptMiddlewares("auth")
router.Prefix("/public").ExceptMiddlewares("auth").Group(func(public *core.Router) {
    public.Get("/info", infoHandler)
    public.Get("/private", privateHandler).Middlewares(auth) // Re-enable for this endpoint.
})

quiet := router.ExceptMiddlewares("auth")
quiet.Get("/quiet", quietHandler)
quiet.Middlewares(auth).Get("/logged", loggedHandler) // Re-enable in a derived scope.
```

The same name replaces/re-enables an entry in its original position; distinct names are independent. Excluding an unknown name retains a disabled placeholder. Later registration with that name re-enables its slot. Names are scoped to each copied middleware list, not a global registry. Anonymous wrappers cannot be excluded by name.

Names must be non-empty. Nil functions and unsupported values passed to `Middlewares` panic during configuration. Invalid batches do not partially update an endpoint. Use `NamedMiddleware` values, not pointers. A separately defined function type can be explicitly converted with `core.Middleware(fn)`.

### Execution order and chain construction

For registration `Middlewares(A, B)`, requests enter A, then B, then the handler. Code after `next.ServeHTTP` unwinds in reverse order:

```text
A before → B before → handler → B after → A after
```

If B rejects, A resumes after its call to B; B's own code determines which cleanup runs. Only deferred cleanup is guaranteed during panic unwinding. Code after `next.ServeHTTP` without `defer` is skipped when downstream panics.

Each endpoint builds its chain once, lazily on its first request, after configuration freezes. Constructors run from the last registered middleware to the first. Keep per-request state inside the returned handler, not in the constructor. Constructors must return a non-nil handler. Constructor panics (including nil-handler validation) are cached and repeated on subsequent requests to that endpoint; they are configuration bugs, not request-level rejection.

## Groups and prefixes

`Group` creates a child router that inherits a snapshot of its parent's prefix and middleware list. Prefixes accumulate across nested groups:

```go
auth := core.NamedMiddleware{Name: "auth", Middleware: DemoGate}
api := router.Prefix("/api").Middlewares(auth)

api.Group(func(v1 *core.Router) {
    v1.Prefix("/v1").Group(func(routes *core.Router) {
        routes.Get("/users/{id}", getUser) // /api/v1/users/{id}, protected by auth.
        routes.Get("/health", healthHandler).ExceptMiddlewares("auth")
    })
})
```

Creating or configuring a child does not change its parent. Groups share the underlying mux and configuration freeze state, so route conflicts apply across all groups. Prefix changes affect routes registered through that derived router afterward. Paths are joined with a leading slash and repeated slashes are collapsed; prefix trailing slashes are removed, while endpoint trailing slashes are preserved for subtree routing.

## Serving and configuration lifecycle

Create routers with `core.NewRouter()`. Complete all route and middleware setup before serving requests.

Pass the router directly to an HTTP server; it implements `http.Handler` through `ServeHTTP`:

```go
server := &http.Server{
    Addr:              ":8000",
    Handler:           router,
    ReadHeaderTimeout: 3 * time.Second,
    ReadTimeout:       5 * time.Second,
    WriteTimeout:      10 * time.Second,
    IdleTimeout:       120 * time.Second,
}
log.Fatal(server.ListenAndServe())
```

This snippet also needs `import "time"`. Adjust server timeouts for your application's needs.

The first request, even an unmatched request, freezes configuration for the router, all its groups, and registered endpoints. Later route/prefix/middleware configuration calls panic.

`router.GetMux()` returns the underlying `*http.ServeMux`. It can be used as the server's handler when needed, but configuration then freezes only when a registered router endpoint runs. Direct registrations on the returned mux bypass the router's middleware and configuration guard.

## Running the complete example

[The complete example in `example/main.go`](https://github.com/RasyidHakim-2804/go-core-router/blob/main/example/main.go) demonstrates all route helpers, parameters, wildcard matching, nested groups, middleware scopes and exclusions, named middleware, `Handle`, and panic cleanup.

From this repository:

```bash
go run ./example
```

To serve through `GetMux()` instead:

```bash
go run ./example -mux
```

Try these requests in another terminal:

```bash
curl localhost:8000/index
curl localhost:8000/health
curl localhost:8000/files/images/logo.png
curl localhost:8000/api/v1/users/42
curl -H 'X-Demo-Key: secret' localhost:8000/api/v1/users/42
curl -X POST -H 'X-Demo-Key: secret' localhost:8000/api/v1/users
curl -X PATCH -H 'X-Demo-Key: secret' localhost:8000/api/v1/users/42
curl -I -H 'X-Demo-Key: secret' localhost:8000/api/v1/users/42
curl localhost:8000/api/v1/public
curl localhost:8000/api/public/info
curl localhost:8000/custom
```

The request without the demo key to `/api/v1/users/42` returns 401. The example logs an intentional duplicate registration error at startup and continues serving. `/panic` intentionally panics to demonstrate deferred middleware cleanup; it may produce an empty response/closed connection, not a formatted 500 response. More commands are listed at the top of the example file. Stop the server with `Ctrl+C`.

## License

[MIT](LICENSE).
