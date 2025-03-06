package core

import (
	"net/http"
)

type Handler = func(w http.ResponseWriter, r *http.Request)

type RequestHandler struct {
	handler     Handler
	middlewares []MiddlewareAndStatus
}

// === PRIVATE FUNC ===

func (requestHandler *RequestHandler) executeMiddlewares(w http.ResponseWriter, r *http.Request) bool {

	for _, middlewareStatus := range requestHandler.middlewares {
		if !middlewareStatus.status {
			continue
		}

		if !middlewareStatus.Middleware.Next(w, r) {
			return false
		}
	}

	return true
}

func (requestHandler *RequestHandler) run(w http.ResponseWriter, r *http.Request) {

	if !requestHandler.executeMiddlewares(w, r) {
		return
	}

	requestHandler.handler(w, r)
}

func (requestHandler *RequestHandler) setMiddelware(newMiddleware Middleware, status bool) {
	requestHandler.middlewares = generateSliceMiddlewareAndStatus(
		requestHandler.middlewares,
		newMiddleware,
		status,
	)
}

// === PUBLIC FUNC ===

func (requestHandler *RequestHandler) ExceptMiddlewares(exceptMiddlewares ...Middleware) *RequestHandler {

	for _, middleware := range exceptMiddlewares {
		requestHandler.setMiddelware(middleware, false)
	}

	return requestHandler
}

func (requestHandler *RequestHandler) Middlewares(middlewares ...Middleware) *RequestHandler {

	for _, middleware := range middlewares {
		requestHandler.setMiddelware(middleware, true)
	}

	return requestHandler
}
