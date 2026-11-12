package core

import (
	"net/http"
)

type Handler = func(w http.ResponseWriter, r *http.Request)

type RequestHandler struct {
	config      *configuration
	handler     Handler
	middlewares []MiddlewareAndStatus
}

func (requestHandler *RequestHandler) run(w http.ResponseWriter, r *http.Request) {
	requestHandler.config.mu.Lock()
	requestHandler.config.frozen = true
	requestHandler.config.mu.Unlock()
	requestHandler.runAt(0, w, r)
}

func (requestHandler *RequestHandler) runAt(index int, w http.ResponseWriter, r *http.Request) {
	for index < len(requestHandler.middlewares) {
		entry := requestHandler.middlewares[index]
		index++
		if !entry.status {
			continue
		}
		if !entry.Middleware.Before(w, r) {
			return
		}
		defer entry.Middleware.After(w, r)
		requestHandler.runAt(index, w, r)
		return
	}
	requestHandler.handler(w, r)
}

func (requestHandler *RequestHandler) setMiddelware(middleware MiddlewareInterface, status bool) {
	requestHandler.middlewares = generateSliceMiddlewareAndStatus(requestHandler.middlewares, middleware, status)
}

// === PUBLIC FUNC ===

func (requestHandler *RequestHandler) ExceptMiddlewares(exceptMiddlewares ...MiddlewareInterface) *RequestHandler {
	requestHandler.config.edit()
	defer requestHandler.config.mu.Unlock()

	for _, middleware := range exceptMiddlewares {
		requestHandler.setMiddelware(middleware, false)
	}

	return requestHandler
}

func (requestHandler *RequestHandler) Middlewares(middlewares ...MiddlewareInterface) *RequestHandler {
	requestHandler.config.edit()
	defer requestHandler.config.mu.Unlock()

	for _, middleware := range middlewares {
		requestHandler.setMiddelware(middleware, true)
	}

	return requestHandler
}
