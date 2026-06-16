package core

import (
	"net/http"
	"testing"
)

// mockMiddleware implements MiddlewareInterface (used by TestRouter_Middlewares, TestRouterPrefix)
type mockMiddleware struct {
	id int
}

func (m mockMiddleware) Before(w http.ResponseWriter, r *http.Request) bool {
	return true
}

func (m mockMiddleware) After(w http.ResponseWriter, r *http.Request) bool {
	return true
}

// mockMiddlewareStr implements MiddlewareInterface (used by TestRouter_ExceptMiddlewares)
type mockMiddlewareStr struct {
	id string
}

func (m mockMiddlewareStr) Before(w http.ResponseWriter, r *http.Request) bool {
	return true
}

func (m mockMiddlewareStr) After(w http.ResponseWriter, r *http.Request) bool {
	return true
}

func TestRouter_ExceptMiddlewares(t *testing.T) {
	router := NewRouter()
	m1 := mockMiddlewareStr{id: "m1"}
	m2 := mockMiddlewareStr{id: "m2"}
	m3 := mockMiddlewareStr{id: "m3"}
	router.Middlewares(m1, m2, m3)

	// Exclude m2
	router.ExceptMiddlewares(m2)
	if len(router.middlewares) != 3 {
		t.Fatalf("expected 3 middlewares in slice, got %d", len(router.middlewares))
	}
	for _, ms := range router.middlewares {
		if ms.Middleware == m2 {
			if ms.status != false {
				t.Errorf("expected m2 status to be false, got %v", ms.status)
			}
		} else if ms.Middleware == m1 || ms.Middleware == m3 {
			if ms.status != true {
				t.Errorf("expected m1/m3 status to be true, got %v", ms.status)
			}
		} else {
			t.Errorf("unexpected middleware in slice: %v", ms.Middleware)
		}
	}

	// Exclude non-existent middleware (m4), should be added with status=false
	m4 := mockMiddlewareStr{id: "m4"}
	router.ExceptMiddlewares(m4)
	if len(router.middlewares) != 4 {
		t.Fatalf("expected 4 middlewares in slice, got %d", len(router.middlewares))
	}
	foundM4 := false
	for _, ms := range router.middlewares {
		if ms.Middleware == m4 {
			foundM4 = true
			if ms.status != false {
				t.Errorf("expected m4 status to be false, got %v", ms.status)
			}
		}
	}
	if !foundM4 {
		t.Errorf("expected to find m4 in middlewares slice")
	}
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