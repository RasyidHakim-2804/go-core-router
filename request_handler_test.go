package core

import (
	"testing"
)

func TestRequestHandler_ExceptMiddlewares(t *testing.T) {
	rh := &RequestHandler{}
	m1 := mockMiddleware{id: 1}
	m2 := mockMiddleware{id: 2}

	// 1. Adding an excepted middleware to an empty RequestHandler
	r1 := rh.ExceptMiddlewares(m1)
	if r1 != rh {
		t.Errorf("Expected ExceptMiddlewares to return the same RequestHandler instance for chaining")
	}
	if len(rh.middlewares) != 1 {
		t.Fatalf("Expected 1 middleware, got %d", len(rh.middlewares))
	}
	if rh.middlewares[0].Middleware != m1 {
		t.Errorf("Expected middleware to be m1")
	}
	if rh.middlewares[0].status != false {
		t.Errorf("Expected middleware status to be false")
	}

	// 2. Setting an existing middleware to excepted (status=false)
	// First add m2 as an active middleware
	rh.Middlewares(m2)
	if len(rh.middlewares) != 2 {
		t.Fatalf("Expected 2 middlewares, got %d", len(rh.middlewares))
	}
	if rh.middlewares[1].status != true {
		t.Errorf("Expected m2 status to be true initially")
	}

	// Then apply ExceptMiddlewares
	rh.ExceptMiddlewares(m2)
	if len(rh.middlewares) != 2 {
		t.Fatalf("Expected still 2 middlewares, got %d", len(rh.middlewares))
	}
	if rh.middlewares[1].status != false {
		t.Errorf("Expected m2 status to be updated to false")
	}

	// 3. Applying ExceptMiddlewares with multiple arguments
	m3 := mockMiddleware{id: 3}
	m4 := mockMiddleware{id: 4}
	rh.ExceptMiddlewares(m3, m4)
	if len(rh.middlewares) != 4 {
		t.Fatalf("Expected 4 middlewares, got %d", len(rh.middlewares))
	}
	if rh.middlewares[2].Middleware != m3 || rh.middlewares[2].status != false {
		t.Errorf("Expected m3 to be added with status false")
	}
	if rh.middlewares[3].Middleware != m4 || rh.middlewares[3].status != false {
		t.Errorf("Expected m4 to be added with status false")
	}
}
