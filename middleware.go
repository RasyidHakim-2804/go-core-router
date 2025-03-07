package core

import (
	"net/http"
)

type MiddlewareInterface interface {
	Before(w http.ResponseWriter, r *http.Request) bool
	After(w http.ResponseWriter, r *http.Request) bool
}

type Middleware struct {
}

func (m Middleware) Before(w http.ResponseWriter, r *http.Request) bool {
	return true
}

func (m Middleware) After(w http.ResponseWriter, r *http.Request) bool {
	return true
}
