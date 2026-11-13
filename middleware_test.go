package core

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"reflect"
	"sync"
	"sync/atomic"
	"testing"
)

type testUserKey struct{}
type statusWriter struct {
	http.ResponseWriter
	status int
}

func (w *statusWriter) WriteHeader(status int) {
	w.status = status
	w.ResponseWriter.WriteHeader(status)
}

func TestMiddlewareForwardsContextRequestAndWriter(t *testing.T) {
	original := httptest.NewRequest("GET", "/profile", nil)
	status := 0
	inject := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			cloned := r.Clone(context.WithValue(r.Context(), testUserKey{}, "alice"))
			cloned.Header.Set("X-Injected", "yes")
			writer := &statusWriter{ResponseWriter: w}
			next.ServeHTTP(writer, cloned)
			status = writer.status
		})
	}
	r := NewRouter()
	r.Middlewares(inject).Get("/profile", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusCreated)
		fmt.Fprintf(w, "%s:%s", r.Context().Value(testUserKey{}), r.Header.Get("X-Injected"))
	})
	w := httptest.NewRecorder()
	r.ServeHTTP(w, original)
	if w.Body.String() != "alice:yes" || status != 201 || w.Code != 201 {
		t.Fatalf("response=%v status=%d", w, status)
	}
	if original.Context().Value(testUserKey{}) != nil || original.Header.Get("X-Injected") != "" {
		t.Fatal("original request changed")
	}
}

func TestMiddlewareRecovery(t *testing.T) {
	recoverer := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if recover() != nil {
					http.Error(w, "recovered", 500)
				}
			}()
			next.ServeHTTP(w, r)
		})
	}
	r := NewRouter()
	r.Middlewares(recoverer).Get("/", func(http.ResponseWriter, *http.Request) { panic("boom") })
	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest("GET", "/", nil))
	if w.Code != 500 {
		t.Fatal(w.Code)
	}
}

func TestAnonymousClosuresAndNamedReplacement(t *testing.T) {
	var events []string
	factory := func(label string) Middleware {
		return func(next http.Handler) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { events = append(events, label); next.ServeHTTP(w, r) })
		}
	}
	r := NewRouter().Middlewares(factory("one"), factory("two"), NamedMiddleware{Name: "named", Middleware: factory("old")})
	r.Get("/replace", func(http.ResponseWriter, *http.Request) {}).Middlewares(NamedMiddleware{Name: "named", Middleware: factory("new")})
	r.Get("/exclude", func(http.ResponseWriter, *http.Request) {}).ExceptMiddlewares("named")
	r.ExceptMiddlewares("future").Middlewares(NamedMiddleware{Name: "future", Middleware: factory("future")}).Get("/enable", func(http.ResponseWriter, *http.Request) {})
	for _, tc := range []struct {
		path string
		want []string
	}{{"/replace", []string{"one", "two", "new"}}, {"/exclude", []string{"one", "two"}}, {"/enable", []string{"one", "two", "old", "future"}}} {
		events = nil
		r.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest("GET", tc.path, nil))
		if !reflect.DeepEqual(events, tc.want) {
			t.Fatalf("%s: %v", tc.path, events)
		}
	}
}

func TestChainConstructedOnceConcurrently(t *testing.T) {
	var builds, calls atomic.Int32
	wrap := func(next http.Handler) http.Handler {
		builds.Add(1)
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { calls.Add(1); next.ServeHTTP(w, r) })
	}
	r := NewRouter()
	endpoint := r.Get("/", func(http.ResponseWriter, *http.Request) {})
	endpoint.Middlewares(wrap)
	var wg sync.WaitGroup
	for i := 0; i < 32; i++ {
		wg.Add(1)
		go func() { defer wg.Done(); r.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest("GET", "/", nil)) }()
	}
	wg.Wait()
	if builds.Load() != 1 || calls.Load() != 32 {
		t.Fatalf("builds=%d calls=%d", builds.Load(), calls.Load())
	}
}

func mustPanic(t *testing.T, f func()) {
	t.Helper()
	defer func() {
		if recover() == nil {
			t.Error("expected panic")
		}
	}()
	f()
}

func TestInvalidMiddleware(t *testing.T) {
	for _, value := range []any{nil, Middleware(nil), 42, NamedMiddleware{Name: "x"}, NamedMiddleware{Middleware: func(n http.Handler) http.Handler { return n }}} {
		mustPanic(t, func() { NewRouter().Middlewares(value) })
	}
	mustPanic(t, func() { NewRouter().ExceptMiddlewares("") })
	for _, bad := range []Middleware{
		func(http.Handler) http.Handler { return nil },
		func(http.Handler) http.Handler { var h http.HandlerFunc; return h },
		func(http.Handler) http.Handler { panic("constructor failure") },
	} {
		r := NewRouter().Middlewares(bad)
		r.Get("/", func(http.ResponseWriter, *http.Request) {})
		for i := 0; i < 2; i++ {
			mustPanic(t, func() { r.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest("GET", "/", nil)) })
		}
	}
}

func TestEndpointInvalidBatchIsAtomic(t *testing.T) {
	var events []string
	r := NewRouter()
	ep := r.Get("/", func(http.ResponseWriter, *http.Request) {})
	ep.Middlewares(probe(&events, "a", false))
	mustPanic(t, func() { ep.ExceptMiddlewares("a", "") })
	mustPanic(t, func() { ep.Middlewares(probe(&events, "b", false), nil) })
	r.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest("GET", "/", nil))
	if !reflect.DeepEqual(events, []string{"a before", "a after"}) {
		t.Fatal(events)
	}
}
