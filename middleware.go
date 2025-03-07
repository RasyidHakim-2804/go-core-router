package core

import (
	"net/http"
)

type MiddlewareInterface interface {
	Next(w http.ResponseWriter, r *http.Request) bool
	After(w http.ResponseWriter, r *http.Request) bool
}

type Middleware struct {
}

func (m Middleware) Next(w http.ResponseWriter, r *http.Request) bool {
	return true
}

func (m Middleware) After(w http.ResponseWriter, r *http.Request) bool {
	return true
}

type MiddlewareAndStatus struct {
	Middleware MiddlewareInterface
	status     bool
}

func generateSliceMiddlewareAndStatus(slice []MiddlewareAndStatus, middleware MiddlewareInterface, status bool) []MiddlewareAndStatus {
	notExists := true

	for i := range slice {
		if slice[i].Middleware == middleware {
			slice[i].status = status
			notExists = false
			break
		}
	}

	if notExists {
		slice = append(slice, MiddlewareAndStatus{middleware, status})
	}

	return slice
}
