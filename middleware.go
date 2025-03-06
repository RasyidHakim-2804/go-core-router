package core

import (
	"net/http"
)

// Middleware defines the interface for handling middleware logic.
type Middleware interface {
	Next(w http.ResponseWriter, r *http.Request) bool
}

type MiddlewareAndStatus struct {
	Middleware Middleware
	status     bool
}

func generateSliceMiddlewareAndStatus(slice []MiddlewareAndStatus, middleware Middleware, status bool) []MiddlewareAndStatus {
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
