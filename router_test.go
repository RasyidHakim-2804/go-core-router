package core

import (
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
)

func probe(log *[]string, name string, reject bool) NamedMiddleware {
	return NamedMiddleware{Name: name, Middleware: func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			*log = append(*log, name+" before")
			if reject {
				return
			}
			defer func() { *log = append(*log, name+" after") }()
			next.ServeHTTP(w, r)
		})
	}}
}

func TestLifecycle(t *testing.T) {
	for _, mode := range []string{"normal", "reject", "panic"} {
		t.Run(mode, func(t *testing.T) {
			var log []string
			r := NewRouter()
			r = r.Middlewares(probe(&log, "a", false), probe(&log, "b", mode == "reject"))
			r.Get("/", func(http.ResponseWriter, *http.Request) {
				log = append(log, "handler")
				if mode == "panic" {
					panic("boom")
				}
			})
			var recovered any
			func() {
				defer func() { recovered = recover() }()
				r.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest("GET", "/", nil))
			}()
			want := []string{"a before", "b before", "handler", "b after", "a after"}
			if mode == "reject" {
				want = []string{"a before", "b before", "a after"}
			}
			if !reflect.DeepEqual(log, want) {
				t.Fatalf("got %v want %v", log, want)
			}
			if (recovered != nil) != (mode == "panic") {
				t.Fatalf("unexpected panic %v", recovered)
			}
		})
	}
}
func TestRegistration(t *testing.T) {
	r := NewRouter()
	m := NamedMiddleware{Name: "m", Middleware: func(next http.Handler) http.Handler { return next }}
	r = r.Middlewares(m).ExceptMiddlewares("m").Middlewares(m)
	r.Group(func(g *Router) {
		g = g.Prefix("api/")
		g.Get("/users", func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(201) })
	})
	if _, err := r.Handle("GET", "/api/users", func(http.ResponseWriter, *http.Request) {}); err == nil {
		t.Fatal("duplicate accepted")
	}
	if _, err := r.Handle("GET", "/{bad", func(http.ResponseWriter, *http.Request) {}); err == nil {
		t.Fatal("invalid pattern accepted")
	}
	w := httptest.NewRecorder()
	r.GetMux().ServeHTTP(w, httptest.NewRequest("GET", "/api/users", nil))
	if w.Code != 201 {
		t.Fatal(w.Code)
	}
	defer func() {
		if recover() == nil {
			t.Fatal("mutation after serving accepted")
		}
	}()
	r.Middlewares(m)
}

func TestDerivedRouterIsolation(t *testing.T) {
	var events []string
	a := probe(&events, "a", false)
	b := probe(&events, "b", false)
	root := NewRouter()
	base := root.Middlewares(a)
	api := base.Prefix("/api").Middlewares(b).ExceptMiddlewares("a")
	sibling := base.Prefix("/other")
	handle := func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(http.StatusNoContent) }
	// Ignoring a returned scope must have no effect on the receiver.
	root.Prefix("/ignored")
	root.Middlewares(b)
	base.ExceptMiddlewares("a")
	root.Get("/user", handle)
	base.Get("/base", handle)
	api.Group(func(g *Router) {
		g.Get("/user", handle)
		g.Prefix("/v1").Middlewares(a).Group(func(nested *Router) {
			nested.Get("/user", handle)
		})
	})
	sibling.Get("/user", handle)
	for _, tc := range []struct {
		path string
		want []string
	}{
		{"/user", nil},
		{"/base", []string{"a before", "a after"}},
		{"/api/user", []string{"b before", "b after"}},
		{"/api/v1/user", []string{"a before", "b before", "b after", "a after"}},
		{"/other/user", []string{"a before", "a after"}},
	} {
		events = nil
		w := httptest.NewRecorder()
		root.ServeHTTP(w, httptest.NewRequest("GET", tc.path, nil))
		if w.Code != http.StatusNoContent || !reflect.DeepEqual(events, tc.want) {
			t.Fatalf("%s: status=%d events=%v want=%v", tc.path, w.Code, events, tc.want)
		}
	}
	// All derived scopes still share the configuration freeze guard.
	defer func() {
		if recover() == nil {
			t.Fatal("derived scope accepted configuration after serving")
		}
	}()
	api.Prefix("/late")
}
