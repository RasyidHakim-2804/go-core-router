package core

import (
	"fmt"
	"net/http"
	"strings"
	"sync"
)

// Router struct represents the routing system for handling HTTP requests.
type configuration struct {
	mu     sync.Mutex
	frozen bool
}

func (c *configuration) edit() {
	c.mu.Lock()
	if c.frozen {
		c.mu.Unlock()
		panic("core: router configuration is frozen after serving begins")
	}
}

type Router struct {
	config      *configuration
	middlewares []middlewareEntry
	prefix      string
	mux         *http.ServeMux
}

// === PRIVATE FUNC ===

// clone copies configuration; callers hold the shared configuration lock.
// Middleware instances and the mux remain shared.
func (router *Router) clone() *Router {
	return &Router{config: router.config, mux: router.mux, prefix: router.prefix,
		middlewares: append([]middlewareEntry(nil), router.middlewares...)}
}

func (router *Router) newRequestHandler(handler *Handler) *RequestHandler {
	newRequestHandler := &RequestHandler{
		handler:     *handler,
		config:      router.config,
		middlewares: make([]middlewareEntry, len(router.middlewares)),
	}

	copy(newRequestHandler.middlewares, router.middlewares)
	newRequestHandler.compiled = sync.OnceValue(newRequestHandler.buildChain)
	return newRequestHandler
}

// handles the handler logic.
func (router *Router) next(method string, route string, requestHandler *RequestHandler) {

	urlWithMethod := method + " " + joinRoute(router.prefix, route)

	router.mux.HandleFunc(urlWithMethod, func(w http.ResponseWriter, r *http.Request) {
		requestHandler.run(w, r)
	})
}

// === PUBLIC FUNC ===

func NewRouter() *Router {
	return &Router{
		config: &configuration{},
		prefix: "",                 // Initial empty prefix
		mux:    http.NewServeMux(), // Create a new ServeMux instance
	}
}

func (router *Router) GetMux() *http.ServeMux {
	return router.mux
}

// Middlewares returns a derived router with middleware added/enabled.
// Accepts func(http.Handler) http.Handler and NamedMiddleware values.
// Anonymous functions append; named entries replace by name in the same position.
func (router *Router) Middlewares(middlewares ...any) *Router {
	router.config.edit()
	defer router.config.mu.Unlock()
	derived := router.clone()
	derived.middlewares = addMiddlewares(derived.middlewares, middlewares)
	return derived
}

// ExceptMiddlewares returns a derived router with the given names disabled.
func (router *Router) ExceptMiddlewares(names ...string) *Router {
	router.config.edit()
	defer router.config.mu.Unlock()
	derived := router.clone()
	derived.middlewares = excludeMiddlewares(derived.middlewares, names)
	return derived
}

// Prefix returns a derived router with an appended prefix; the receiver is unchanged.
func (router *Router) Prefix(prefix string) *Router {
	router.config.edit()
	defer router.config.mu.Unlock()
	derived := router.clone()
	derived.prefix = strings.TrimRight(joinRoute(derived.prefix, prefix), "/")
	return derived
}

// create a new router based on the parent router, the middleware and prefix will be inherited.
func (router *Router) Group(callback func(router *Router)) {
	router.config.edit()
	newRouter := router.clone()
	router.config.mu.Unlock()

	callback(newRouter)
}

// Handle registers a method and path. Invalid or conflicting patterns return errors.
func (router *Router) Handle(method, route string, handler Handler) (result *RequestHandler, err error) {
	router.config.edit()
	defer router.config.mu.Unlock()
	if handler == nil {
		return nil, fmt.Errorf("core: nil handler")
	}
	if method == "" || strings.ContainsAny(method, " \t\r\n") {
		return nil, fmt.Errorf("core: invalid method %q", method)
	}
	defer func() {
		if problem := recover(); problem != nil {
			result = nil
			err = fmt.Errorf("core: registering %s %q: %v", method, route, problem)
		}
	}()
	result = router.newRequestHandler(&handler)
	router.next(method, route, result)
	return result, nil
}

func (router *Router) mustHandle(method, route string, handler Handler) *RequestHandler {
	result, err := router.Handle(method, route, handler)
	if err != nil {
		panic(err)
	}
	return result
}

// ServeHTTP allows using Router directly as an http.Handler.
func (router *Router) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	router.config.mu.Lock()
	router.config.frozen = true
	router.config.mu.Unlock()
	router.mux.ServeHTTP(w, r)
}

// Preserve trailing slashes: ServeMux uses them to identify subtree patterns.
func joinRoute(prefix, route string) string {
	path := strings.TrimRight(prefix, "/") + "/" + strings.TrimLeft(route, "/")
	for strings.Contains(path, "//") {
		path = strings.ReplaceAll(path, "//", "/")
	}
	return path
}

func (router *Router) Get(route string, handler Handler) *RequestHandler {
	return router.mustHandle("GET", route, handler)
}

func (router *Router) Post(route string, handler Handler) *RequestHandler {
	return router.mustHandle("POST", route, handler)
}

func (router *Router) Put(route string, handler Handler) *RequestHandler {
	return router.mustHandle("PUT", route, handler)
}

func (router *Router) Delete(route string, handler Handler) *RequestHandler {
	return router.mustHandle("DELETE", route, handler)
}

func (router *Router) Patch(route string, handler Handler) *RequestHandler {
	return router.mustHandle("PATCH", route, handler)
}

func (router *Router) Head(route string, handler Handler) *RequestHandler {
	return router.mustHandle("HEAD", route, handler)
}

func (router *Router) Options(route string, handler Handler) *RequestHandler {
	return router.mustHandle("OPTIONS", route, handler)
}
