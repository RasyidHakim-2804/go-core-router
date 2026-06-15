package core

import (
	"net/http"
	"testing"
)

type mockMiddleware struct {
	id string
}

func (m mockMiddleware) Before(w http.ResponseWriter, r *http.Request) bool {
	return true
}

func (m mockMiddleware) After(w http.ResponseWriter, r *http.Request) bool {
	return true
}

func TestRouter_ExceptMiddlewares(t *testing.T) {
	router := NewRouter()

	m1 := mockMiddleware{id: "m1"}
	m2 := mockMiddleware{id: "m2"}
	m3 := mockMiddleware{id: "m3"}

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
	m4 := mockMiddleware{id: "m4"}
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
