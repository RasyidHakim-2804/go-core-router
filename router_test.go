package core

import (
	"net/http"
	"testing"
)

// mockMiddleware implements MiddlewareInterface
type mockMiddleware struct {
	id int
}

func (m mockMiddleware) Before(w http.ResponseWriter, r *http.Request) bool {
	return true
}

func (m mockMiddleware) After(w http.ResponseWriter, r *http.Request) bool {
	return true
}

func TestRouter_Middlewares(t *testing.T) {
	router := NewRouter()
	m1 := mockMiddleware{id: 1}
	m2 := mockMiddleware{id: 2}

	// Test adding a single middleware
	r1 := router.Middlewares(m1)
	if r1 != router {
		t.Errorf("Expected Middlewares to return the same router instance for chaining")
	}
	if len(router.middlewares) != 1 {
		t.Fatalf("Expected 1 middleware, got %d", len(router.middlewares))
	}
	if router.middlewares[0].Middleware != m1 {
		t.Errorf("Expected middleware to be m1")
	}
	if router.middlewares[0].status != true {
		t.Errorf("Expected middleware status to be true")
	}

	// Test adding multiple middlewares
	router.Middlewares(m2)
	if len(router.middlewares) != 2 {
		t.Fatalf("Expected 2 middlewares, got %d", len(router.middlewares))
	}
	if router.middlewares[1].Middleware != m2 {
		t.Errorf("Expected second middleware to be m2")
	}

	// Test adding a duplicate middleware
	// generateSliceMiddlewareAndStatus should update the status and not append a new one
	router.Middlewares(m1)
	if len(router.middlewares) != 2 {
		t.Fatalf("Expected 2 middlewares after adding a duplicate, got %d", len(router.middlewares))
	}

	// Test adding multiple at once
	m3 := mockMiddleware{id: 3}
	m4 := mockMiddleware{id: 4}
	router.Middlewares(m3, m4)
	if len(router.middlewares) != 4 {
		t.Fatalf("Expected 4 middlewares after adding two at once, got %d", len(router.middlewares))
	}
	if router.middlewares[2].Middleware != m3 {
		t.Errorf("Expected third middleware to be m3")
	}
	if router.middlewares[3].Middleware != m4 {
		t.Errorf("Expected fourth middleware to be m4")
	}
}

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