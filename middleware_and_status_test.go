package core

import (
	"net/http"
	"testing"
)

type MockMiddleware struct {
	id int
}

func (m MockMiddleware) Before(w http.ResponseWriter, r *http.Request) bool {
	return true
}

func (m MockMiddleware) After(w http.ResponseWriter, r *http.Request) bool {
	return true
}

func TestGenerateSliceMiddlewareAndStatus(t *testing.T) {
	m1 := MockMiddleware{id: 1}
	m2 := MockMiddleware{id: 2}

	t.Run("Append (Not Exists) - Empty Slice", func(t *testing.T) {
		var slice []MiddlewareAndStatus
		result := generateSliceMiddlewareAndStatus(slice, m1, true)

		if len(result) != 1 {
			t.Fatalf("expected length 1, got %d", len(result))
		}

		if result[0].Middleware != m1 || result[0].status != true {
			t.Errorf("unexpected middleware/status appended")
		}
	})

	t.Run("Append (Not Exists) - Multiple Items", func(t *testing.T) {
		slice := []MiddlewareAndStatus{
			{Middleware: m1, status: true},
		}
		result := generateSliceMiddlewareAndStatus(slice, m2, false)

		if len(result) != 2 {
			t.Fatalf("expected length 2, got %d", len(result))
		}

		if result[1].Middleware != m2 || result[1].status != false {
			t.Errorf("unexpected middleware/status appended")
		}
	})

	t.Run("Update (Exists)", func(t *testing.T) {
		slice := []MiddlewareAndStatus{
			{Middleware: m1, status: true},
			{Middleware: m2, status: true},
		}
		// m2 already exists, change its status to false
		result := generateSliceMiddlewareAndStatus(slice, m2, false)

		if len(result) != 2 {
			t.Fatalf("expected length 2, got %d", len(result))
		}

		if result[1].Middleware != m2 || result[1].status != false {
			t.Errorf("expected status to be updated to false, got %v", result[1].status)
		}

        // ensure m1 is unaffected
        if result[0].Middleware != m1 || result[0].status != true {
			t.Errorf("first element unexpectedly modified")
		}
	})
}
