# Go Core Router

A lightweight Go HTTP router with Laravel-inspired route groups, prefixes, and middleware configuration. It builds on `net/http.ServeMux` and uses standard `http.ResponseWriter` and `*http.Request` handlers.

## Features

- Helpers for `GET`, `POST`, `PUT`, `PATCH`, `DELETE`, `HEAD`, and `OPTIONS`.
- `Handle` for explicit method registration and registration errors.
- Path parameters, catch-all wildcards, and exact-root patterns through `ServeMux`.
- Nested groups with inherited prefixes and middleware configuration.
- Middleware additions, exclusions, and re-enabling at router, group, and endpoint level.
- Named middleware instances and `Before`/`After` hooks.
- Direct `http.Handler` compatibility and access to the underlying mux.

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

The snippets below are configuration examples to place before starting the server. Handler names such as `listUsers` represent functions supplied by your application.

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
endpoint.Middlewares(TraceMiddleware{Label: "status"})
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

Middleware implements both hooks:

```go
type MiddlewareInterface interface {
    Before(http.ResponseWriter, *http.Request) bool
    After(http.ResponseWriter, *http.Request) bool
}
```

Embed `core.Middleware` to inherit default hooks returning `true`, then override only what you need:

```go
type TraceMiddleware struct {
    core.Middleware
    Label string
}

func (m TraceMiddleware) Before(w http.ResponseWriter, r *http.Request) bool {
    log.Printf("%s Before: %s %s", m.Label, r.Method, r.URL.Path)
    return true
}

func (m TraceMiddleware) After(w http.ResponseWriter, r *http.Request) bool {
    log.Printf("%s After: %s %s", m.Label, r.Method, r.URL.Path)
    return true
}
```

### Router, group, and endpoint middleware

```go
access := TraceMiddleware{Label: "access"}
router = router.Middlewares(access)

router.Get("/health", healthHandler).ExceptMiddlewares(access)
router.Get("/details", detailsHandler).
    Middlewares(TraceMiddleware{Label: "details"})

router.Prefix("/api").Middlewares(TraceMiddleware{Label: "api"}).Group(func(api *core.Router) {
    api.Get("/users", listUsers) // Runs inherited access + group api middleware.
})
```

Both `Middlewares` and `ExceptMiddlewares` accept multiple arguments. On a router, these methods return a derived router without changing the receiver; chain the result or save it. An endpoint captures a snapshot when registered; its own middleware methods change only that endpoint.

### Exclusions and re-enabling

```go
quiet := router.ExceptMiddlewares(access)
quiet.Get("/quiet", quietHandler)
quiet.Middlewares(access).Get("/logged", loggedHandler) // Re-enable in a new scope.

router.Prefix("/public").ExceptMiddlewares(access).Group(func(public *core.Router) {
    public.Get("/info", infoHandler)
    public.Get("/logged", loggedHandler).Middlewares(access)
})
```

Exclusions are retained as disabled entries, even if that middleware was not previously present. Re-enabling an identity updates its instance without moving its position in the list.

### Named middleware and identity

```go
access := core.NamedMiddleware{
    Name: "access",
    MiddlewareInterface: TraceMiddleware{Label: "access"},
}
audit := core.NamedMiddleware{
    Name: "audit",
    MiddlewareInterface: TraceMiddleware{Label: "audit"},
}
router = router.Middlewares(access, audit)
router.Get("/health", healthHandler).ExceptMiddlewares(audit)
```

Use `NamedMiddleware` values as above when you want explicit identity. Different names create separate entries; the same name updates/re-enables an existing entry. Empty names are invalid. Use distinct names for instances that should run separately.

Without a name, comparable values use value equality, pointers use pointer identity, and non-comparable values use their concrete type as identity. For a pointer middleware, keep and reuse the same pointer when excluding it. Nil middleware, including typed nil pointers, is rejected with a panic.

### Execution order, rejection, and panic

For middleware A followed by B:

```text
A.Before → B.Before → handler → B.After → A.After
```

Return `false` from `Before` to stop the request. The middleware must write its own rejection response if one is needed:

```go
type DemoGate struct { core.Middleware }

func (DemoGate) Before(w http.ResponseWriter, r *http.Request) bool {
    if r.Header.Get("X-Demo-Key") != "secret" {
        http.Error(w, "send X-Demo-Key: secret", http.StatusUnauthorized)
        return false
    }
    return true
}
```

This fixed key is only an example of rejection, not production authentication.

If B rejects, the sequence is `A.Before → B.Before → A.After`. B's `After` is not called because B did not pass `Before`. If `Before` panics, its own `After` likewise has not been scheduled.

`After` runs in reverse order for middleware whose `Before` returned `true`, including when downstream code panics. The router does not recover or turn panics into HTTP error responses. With the standard HTTP server, a propagated handler panic is handled by `net/http`.

**The boolean returned by `After` is currently ignored.** Returning `false` does not stop other `After` hooks. The return type is retained for compatibility. Headers already sent cannot be changed in `After`.

## Chaining without modifying the original router

**Breaking change:** router-level `Prefix`, `Middlewares`, and `ExceptMiddlewares` now return a derived router. They never change the receiver's prefix or middleware configuration. Use the returned value:

```go
router.Prefix("/api").Middlewares(DemoGate{}).Group(func(api *core.Router) {
    api.Get("/users", listUsers) // /api/users, with DemoGate.
})
router.Get("/health", healthHandler) // /health, without DemoGate.

// Reuse a configured scope for several routes.
api := router.Prefix("/api")
api.Get("/status", statusHandler) // /api/status.
```

The old standalone usage has no configuration effect:

```go
router.Prefix("/api") // Returned scope is discarded.
router.Get("/user", getUser) // /user, not /api/user.
```

To migrate `router.Middlewares(m)` followed by route registrations, use `router = router.Middlewares(m)` or chain directly into `Get`/`Group`. The same rule applies inside group callbacks. `Group` ends the chain because it returns no value.

All derived scopes share the mux, so routes registered through them are served by the original router. Registration itself still changes the shared route table. Endpoint methods (`router.Get(...).Middlewares(...)` and `.ExceptMiddlewares(...)`) continue configuring that registered endpoint directly.

## Groups and prefixes

Groups inherit a snapshot of their parent's prefix and middleware list. Nested groups accumulate prefixes:

```go
router.Prefix("/api").Middlewares(DemoGate{}).Group(func(api *core.Router) {
    api.Prefix("/v1").Group(func(v1 *core.Router) {
        v1.Get("/users/{id}", getUser) // /api/v1/users/{id}
        v1.Get("/public", publicHandler).ExceptMiddlewares(DemoGate{})
    })
})
```

Group configuration does not change the parent, and new scopes derived from the parent do not change existing groups. All groups share the underlying mux and configuration freeze state, so conflicts are checked across all registered routes.

`Prefix` appends to the prefix in the returned scope, leaving the receiver unchanged. Paths receive a leading slash and repeated slashes are collapsed; prefix trailing slashes are removed, while endpoint trailing slashes are preserved for subtree routing.

Copies are shallow: referenced middleware state can remain shared between groups, endpoints, and concurrent requests. Middleware must synchronize mutable shared state or keep it request-local.

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

[`example/main.go`](example/main.go) demonstrates all route helpers, parameters, wildcard matching, nested groups, middleware scopes and exclusions, named middleware, `Handle`, and panic cleanup.

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

The request without the demo key to `/api/v1/users/42` returns 401. The example logs an intentional duplicate registration error at startup and continues serving. `/panic` intentionally panics to demonstrate `After` cleanup; it may produce an empty response/closed connection, not a formatted 500 response. More commands are listed at the top of the example file. Stop the server with `Ctrl+C`.

## License

[MIT](LICENSE).
