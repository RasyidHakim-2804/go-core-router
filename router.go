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
	middlewares []MiddlewareAndStatus
	prefix      string
	mux         *http.ServeMux
}

// === PRIVATE FUNC ===

func (router *Router) newRequestHandler(handler *Handler) *RequestHandler {
	newRequestHandler := &RequestHandler{
		handler:     *handler,
		config:      router.config,
		middlewares: make([]MiddlewareAndStatus, len(router.middlewares)),
	}

	copy(newRequestHandler.middlewares, router.middlewares)
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

// Allows adding multiple middlewares to the Router.
func (router *Router) Middlewares(middlewares ...MiddlewareInterface) *Router {
	router.config.edit()
	defer router.config.mu.Unlock()

	for _, middleware := range middlewares {
		router.middlewares = generateSliceMiddlewareAndStatus(
			router.middlewares,
			middleware,
			true,
		)
	}

	return router
}

// Allows excluding multiple middlewares from being applied to the Router.
func (router *Router) ExceptMiddlewares(exceptMiddlewares ...MiddlewareInterface) *Router {
	router.config.edit()
	defer router.config.mu.Unlock()

	for _, middleware := range exceptMiddlewares {
		router.middlewares = generateSliceMiddlewareAndStatus(
			router.middlewares,
			middleware,
			false,
		)
	}

	return router
}

// Adds a prefix to all routes in the Router.
func (router *Router) Prefix(prefix string) *Router {
	router.config.edit()
	defer router.config.mu.Unlock()
	router.prefix = strings.TrimRight(joinRoute(router.prefix, prefix), "/")
	return router
}

// create a new router based on the parent router, the middleware and prefix will be inherited.
func (router *Router) Group(callback func(router *Router)) {
	router.config.edit()
	newRouter := &Router{prefix: router.prefix, mux: router.mux, config: router.config,
		middlewares: append([]MiddlewareAndStatus(nil), router.middlewares...)}
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
