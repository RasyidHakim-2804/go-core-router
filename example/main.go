// Run: go run ./example
// To serve the underlying mux instead: go run ./example -mux
//
// Try:
//
//	curl localhost:8000/index
//	curl localhost:8000/health
//	curl localhost:8000/files/images/logo.png
//	curl localhost:8000/api/v1/users/42
//	curl -H 'X-Demo-Key: secret' localhost:8000/api/v1/users/42
//	curl -X POST -H 'X-Demo-Key: secret' localhost:8000/api/v1/users
//	curl -X PUT -H 'X-Demo-Key: secret' localhost:8000/api/v1/users/42
//	curl -X PATCH -H 'X-Demo-Key: secret' localhost:8000/api/v1/users/42
//	curl -X DELETE -H 'X-Demo-Key: secret' localhost:8000/api/v1/users/42
//	curl -I -H 'X-Demo-Key: secret' localhost:8000/api/v1/users/42
//	curl -i -X OPTIONS -H 'X-Demo-Key: secret' localhost:8000/api/v1/users/42
//	curl localhost:8000/api/v1/public
//	curl localhost:8000/api/public/info
//	curl localhost:8000/api/public/logged
//	curl localhost:8000/api/v1/status
//	curl localhost:8000/custom
//	curl localhost:8000/panic
package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"time"

	core "github.com/RasyidHakim-2804/go-core-router"
)

// Each endpoint builds its wrapper chain once. Keep request state inside ServeHTTP.
func Trace(label string) core.Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			log.Printf("%s before: %s %s", label, r.Method, r.URL.Path)
			defer log.Printf("%s after: %s %s", label, r.Method, r.URL.Path)
			next.ServeHTTP(w, r)
		})
	}
}

type userContextKey struct{}

// A fixed key is a demonstration only, not production authentication.
// The handler receives a new request with a user stored in its context.
func DemoGate(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("X-Demo-Key") != "secret" {
			http.Error(w, "send X-Demo-Key: secret", http.StatusUnauthorized)
			return
		}
		ctx := context.WithValue(r.Context(), userContextKey{}, "demo-user")
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// core.Handler is an alias for func(http.ResponseWriter, *http.Request).
func message(text string) core.Handler {
	return func(w http.ResponseWriter, r *http.Request) { fmt.Fprintln(w, text) }
}

func main() {
	useMux := flag.Bool("mux", false, "serve using GetMux instead of Router.ServeHTTP")
	flag.Parse()
	router := core.NewRouter()

	// Distinct names allow independent control of instances of the same type.
	access := core.NamedMiddleware{Name: "access", Middleware: Trace("access")}
	audit := core.NamedMiddleware{Name: "audit", Middleware: Trace("audit")}
	router = router.Middlewares(access, audit)
	// Router-level exclusion; re-registering an identity re-enables its existing slot.
	router = router.ExceptMiddlewares("audit")
	router = router.Middlewares(audit)
	router.Get("/index", message("Hello world"))
	router.Get("/health", message("ok")).ExceptMiddlewares("access", "audit")

	// Wildcard captures the remaining path; this example does not read files from disk.
	router.Get("/files/{path...}", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "path=%s\n", r.PathValue("path"))
	})
	// {$} matches only the root; / alone would be a subtree pattern.
	router.Get("/{$}", message("Go Core Router examples: start with /index"))

	router.Prefix("/api/").Middlewares(Trace("api")).Group(func(api *core.Router) {
		api.Prefix("/v1").Middlewares(core.NamedMiddleware{Name: "gate", Middleware: DemoGate}).Group(func(v1 *core.Router) {
			// The endpoint snapshots middleware at registration time.
			v1.Get("/users/{id}", func(w http.ResponseWriter, r *http.Request) {
				user, _ := r.Context().Value(userContextKey{}).(string)
				fmt.Fprintf(w, "user=%s, requested-by=%s\n", r.PathValue("id"), user)
			}).Middlewares(Trace("endpoint"))
			v1.Post("/users", func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusCreated)
				fmt.Fprintln(w, "demo user created")
			})
			v1.Put("/users/{id}", message("demo user replaced"))
			v1.Patch("/users/{id}", message("demo user updated"))
			v1.Delete("/users/{id}", func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusNoContent)
			})
			v1.Head("/users/{id}", func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("X-Demo-User", r.PathValue("id"))
				w.WriteHeader(http.StatusOK)
			})
			v1.Options("/users/{id}", func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Allow", "GET, HEAD, PUT, PATCH, DELETE, OPTIONS")
				w.WriteHeader(http.StatusNoContent)
			})
			v1.Get("/public", message("endpoint without gate")).ExceptMiddlewares("gate")
			// This derived router excludes the gate without changing v1.
			v1.ExceptMiddlewares("gate").Get("//status", message("v1 ready")) // Repeated slash normalized.
		})
		api.Prefix("/public").ExceptMiddlewares("access", "audit").Group(func(public *core.Router) {
			public.Get("/info", message("public group"))
			public.Get("/logged", message("access re-enabled for this endpoint")).Middlewares(access)
		})
	})
	// Group changes leave the parent configuration unchanged.
	endpoint, err := router.Handle(http.MethodGet, "/custom", message("registered with Handle"))
	if err != nil {
		log.Fatal(err)
	}
	endpoint.Middlewares(Trace("custom"))

	// Handle returns registration errors; helpers such as Get panic on these errors.
	if _, err := router.Handle(http.MethodGet, "/custom", message("duplicate")); err != nil {
		log.Printf("expected duplicate registration error: %v", err)
	}

	// Trace uses defer for cleanup on normal returns, rejection, and panic.
	// The router does not recover panics; a recovery middleware can wrap next.ServeHTTP.
	router.Get("/panic", func(http.ResponseWriter, *http.Request) { panic("intentional demo panic") })

	// Configure everything before serving. The first request freezes the router,
	// groups and endpoints; later configuration calls panic.
	var handler http.Handler = router
	if *useMux {
		handler = router.GetMux()
		// Serving the mux directly freezes configuration only once a registered route runs.
		// Direct registration on the returned mux bypasses the router's configuration guard.
	}
	server := &http.Server{
		Addr: ":8000", Handler: handler,
		ReadHeaderTimeout: 3 * time.Second,
		ReadTimeout:       5 * time.Second,
		WriteTimeout:      10 * time.Second,
		IdleTimeout:       120 * time.Second,
	}
	log.Println("starting demo server at http://localhost:8000/")
	if err := server.ListenAndServe(); err != nil {
		log.Fatal(err)
	}
}
