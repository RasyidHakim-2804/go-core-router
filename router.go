package core

import (
	"maps"
	"net/http"
)

// Router struct represents the routing system for handling HTTP requests.
type Router struct {
	queueMiddlewares []Middleware
	middlewareStatus map[Middleware]bool
	prefix           string
	mux              *http.ServeMux
}

// === PRIVATE FUNC ===

func (router *Router) newRequestHandler(handler *Handler) *RequestHandler {
	newRequestHandler := &RequestHandler{
		handler:          *handler,
		queueMiddlewares: make([]Middleware, len(router.queueMiddlewares)),
		middlewareStatus: maps.Clone(router.middlewareStatus),
	}

	copy(newRequestHandler.queueMiddlewares, router.queueMiddlewares)
	return newRequestHandler
}

func (router *Router) setQueueMiddleware(newMiddleware Middleware) {

	for _, middleware := range router.queueMiddlewares {
		if middleware == newMiddleware {
			return
		}
	}
	router.queueMiddlewares = append(router.queueMiddlewares, newMiddleware)
}

// handles the handler logic.
func (router *Router) next(method string, route string, requestHandler *RequestHandler) {

	urlWithMethod := method + " " + router.prefix + route

	router.mux.HandleFunc(urlWithMethod, func(w http.ResponseWriter, r *http.Request) {
		requestHandler.run(w, r)
	})
}

// === PUBLIC FUNC ===

func NewRouter() *Router {
	return &Router{
		middlewareStatus: make(map[Middleware]bool), // Initialize an empty map for middlewares
		prefix:           "",                        // Initial empty prefix
		mux:              http.NewServeMux(),        // Create a new ServeMux instance
	}
}

func (router *Router) GetMux() *http.ServeMux {
	return router.mux
}

// Allows adding multiple middlewares to the Router.
func (router *Router) Middlewares(middlewares ...Middleware) *Router {

	for _, middleware := range middlewares {
		router.setQueueMiddleware(middleware)
		router.middlewareStatus[middleware] = true
	}

	return router
}

// Allows excluding multiple middlewares from being applied to the Router.
func (router *Router) ExceptMiddlewares(exceptMiddlewares ...Middleware) *Router {

	for _, middleware := range exceptMiddlewares {
		router.setQueueMiddleware(middleware)
		router.middlewareStatus[middleware] = false
	}

	return router
}

// Adds a prefix to all routes in the Router.
func (router *Router) Prefix(prefix string) *Router {
	router.prefix += prefix
	return router
}

// create a new router based on the parent router, the middleware and prefix will be inherited.
func (router *Router) Group(callback func(router *Router)) {
	newRouter := &Router{}
	newRouter.prefix = router.prefix
	newRouter.middlewareStatus = maps.Clone(router.middlewareStatus)
	newRouter.queueMiddlewares = make([]Middleware, len(router.queueMiddlewares))
	newRouter.mux = router.mux

	copy(newRouter.queueMiddlewares, router.queueMiddlewares)

	callback(newRouter)
}

func (router *Router) Get(route string, handler Handler) *RequestHandler {
	requestHandler := router.newRequestHandler(&handler)
	router.next("GET", route, requestHandler)
	return requestHandler
}

func (router *Router) Post(route string, handler Handler) *RequestHandler {
	requestHandler := router.newRequestHandler(&handler)
	router.next("POST", route, requestHandler)
	return requestHandler
}

func (router *Router) Put(route string, handler Handler) *RequestHandler {
	requestHandler := router.newRequestHandler(&handler)
	router.next("PUT", route, requestHandler)
	return requestHandler
}

func (router *Router) Delete(route string, handler Handler) *RequestHandler {
	requestHandler := router.newRequestHandler(&handler)
	router.next("DELETE", route, requestHandler)
	return requestHandler
}
