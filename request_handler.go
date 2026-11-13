package core

import (
	"net/http"
	"reflect"
)

type Handler = func(w http.ResponseWriter, r *http.Request)

type RequestHandler struct {
	config      *configuration
	handler     Handler
	middlewares []middlewareEntry
	compiled    func() http.Handler
}

func (requestHandler *RequestHandler) run(w http.ResponseWriter, r *http.Request) {
	requestHandler.config.mu.Lock()
	requestHandler.config.frozen = true
	requestHandler.config.mu.Unlock()
	requestHandler.compiled().ServeHTTP(w, r)
}

// Compile once, after configuration freezes. First registered middleware is
// outermost. Constructors are invoked once per endpoint, in reverse order.
func (requestHandler *RequestHandler) buildChain() http.Handler {
	var next http.Handler = http.HandlerFunc(requestHandler.handler)
	for i := len(requestHandler.middlewares) - 1; i >= 0; i-- {
		entry := requestHandler.middlewares[i]
		if !entry.enabled {
			continue
		}
		next = entry.wrap(next)
		if next == nil {
			panic("core: middleware returned nil handler")
		}
		value := reflect.ValueOf(next)
		switch value.Kind() {
		case reflect.Pointer, reflect.Func, reflect.Map, reflect.Slice, reflect.Chan, reflect.Interface:
			if value.IsNil() {
				panic("core: middleware returned typed nil handler")
			}
		}
	}
	return next
}

// ExceptMiddlewares disables named middleware on this endpoint only.
func (requestHandler *RequestHandler) ExceptMiddlewares(names ...string) *RequestHandler {
	requestHandler.config.edit()
	defer requestHandler.config.mu.Unlock()
	entries := append([]middlewareEntry(nil), requestHandler.middlewares...)
	requestHandler.middlewares = excludeMiddlewares(entries, names)
	return requestHandler
}

// Middlewares accepts middleware functions or NamedMiddleware values. It changes
// this registered endpoint only, unlike the immutable Router method.
func (requestHandler *RequestHandler) Middlewares(middlewares ...any) *RequestHandler {
	requestHandler.config.edit()
	defer requestHandler.config.mu.Unlock()
	entries := append([]middlewareEntry(nil), requestHandler.middlewares...)
	requestHandler.middlewares = addMiddlewares(entries, middlewares)
	return requestHandler
}
