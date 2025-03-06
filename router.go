package core

import (
	"net/http"
)

// Router struct represents the routing system for handling HTTP requests.
type Router struct {
	middlewares []MiddlewareAndStatus
	prefix      string
	mux         *http.ServeMux
}

// === PRIVATE FUNC ===

func (router *Router) newRequestHandler(handler *Handler) *RequestHandler {
	newRequestHandler := &RequestHandler{
		handler:     *handler,
		middlewares: make([]MiddlewareAndStatus, len(router.middlewares)),
	}

	copy(newRequestHandler.middlewares, router.middlewares)
	return newRequestHandler
}

// func (router *Router) setQueueMiddleware(newMiddleware Middleware) {

// 	for _, middleware := range router.queueMiddlewares {
// 		if middleware == newMiddleware {
// 			return
// 		}
// 	}
// 	router.queueMiddlewares = append(router.queueMiddlewares, newMiddleware)
// }

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
		prefix: "",                 // Initial empty prefix
		mux:    http.NewServeMux(), // Create a new ServeMux instance
	}
}

func (router *Router) GetMux() *http.ServeMux {
	return router.mux
}

// Allows adding multiple middlewares to the Router.
func (router *Router) Middlewares(middlewares ...Middleware) *Router {

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
func (router *Router) ExceptMiddlewares(exceptMiddlewares ...Middleware) *Router {

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
	router.prefix += prefix
	return router
}

// create a new router based on the parent router, the middleware and prefix will be inherited.
func (router *Router) Group(callback func(router *Router)) {
	newRouter := &Router{}
	newRouter.prefix = router.prefix
	newRouter.middlewares = make([]MiddlewareAndStatus, len(router.middlewares))
	newRouter.mux = router.mux

	copy(newRouter.middlewares, router.middlewares)

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
