package core

import (
	"net/http"
)

// Middleware defines the interface for handling middleware logic.
type Middleware interface {
	Next(w http.ResponseWriter, r *http.Request) bool
}
