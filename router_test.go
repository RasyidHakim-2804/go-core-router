package core

import (
	"testing"
)

func TestRouterPrefix(t *testing.T) {
	router := &Router{}

	router.Prefix("/api/v1")
	if router.prefix != "/api/v1" {
		t.Errorf("Expected prefix '/api/v1', got '%s'", router.prefix)
	}

	router.Prefix("/users")
	if router.prefix != "/api/v1/users" {
		t.Errorf("Expected prefix '/api/v1/users', got '%s'", router.prefix)
	}
}
