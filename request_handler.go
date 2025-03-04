package core

import (
	"net/http"
)

type Handler = func(w http.ResponseWriter, r *http.Request)

type RequestHandler struct {
	handler          Handler
	queueMiddlewares []Middleware
	middlewareStatus map[Middleware]bool
}

// === PRIVATE FUNC ===

func (requestHandler *RequestHandler) executeMiddlewares(w http.ResponseWriter, r *http.Request) bool {

	for _, middleware := range requestHandler.queueMiddlewares {
		if requestHandler.middlewareStatus[middleware] {
			if !middleware.Next(w, r) {
				return false
			}
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

func (requestHandler *RequestHandler) setQueueMiddleware(newMiddleware Middleware) {

	for _, middleware := range requestHandler.queueMiddlewares {
		if middleware == newMiddleware {
			return
		}
	}
	requestHandler.queueMiddlewares = append(requestHandler.queueMiddlewares, newMiddleware)
}

// === PUBLIC FUNC ===

func (requestHandler *RequestHandler) ExceptMiddlewares(exceptMiddlewares ...Middleware) *RequestHandler {

	for _, middleware := range exceptMiddlewares {
		requestHandler.setQueueMiddleware(middleware)
		requestHandler.middlewareStatus[middleware] = false
	}

	return requestHandler
}

func (requestHandler *RequestHandler) Middlewares(middlewares ...Middleware) *RequestHandler {

	for _, middleware := range middlewares {
		requestHandler.setQueueMiddleware(middleware)
		requestHandler.middlewareStatus[middleware] = true
	}

	return requestHandler
}
