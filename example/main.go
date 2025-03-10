package main

import (
	"fmt"
	"net/http"

	"github.com/RasyidHakim-2804/go-core-router"
)

type GlobalMiddleware struct {
	core.Middleware
}

func (gm GlobalMiddleware) Before(w http.ResponseWriter, r *http.Request) bool {
	fmt.Println("this is global middleware")
	if contentType := r.Header.Get("Accept"); contentType != "application/json" {

		http.Error(w, "header 'Accept' must 'application/json'", http.StatusUnsupportedMediaType)

		return false
	}
	return true
}

// <--- END GLOBAL MIDDLEWARE -->

type FirstMiddleware struct {
	core.Middleware
}

func (fm FirstMiddleware) Before(w http.ResponseWriter, r *http.Request) bool {
	fmt.Println("this is first Middleware")
	return true
}

// <-- END FIRST MIDDLEWARE -->

func handler(w http.ResponseWriter, r *http.Request) {
	mesage := "Hello world"
	w.Write([]byte(mesage))
}

func homeHandler(w http.ResponseWriter, r *http.Request) {
	message := "This is from home"
	w.Write([]byte(message))
}

func aboutHandler(w http.ResponseWriter, r *http.Request) {
	message := "This is from about"
	w.Write([]byte(message))
}

func main() {
	router := core.NewRouter()

	router.Middlewares(GlobalMiddleware{}, FirstMiddleware{})
	router.Get("/index", handler) // /index

	router.Group(func(router2 *core.Router) {
		router2.Prefix("/home")

		router2.Get("/index", homeHandler).ExceptMiddlewares(FirstMiddleware{})   // /home/index
		router2.Get("/about", aboutHandler).ExceptMiddlewares(GlobalMiddleware{}) // /home/about
	})

	fmt.Println("starting web server at http://localhost:8000/")
	http.ListenAndServe(":8000", router.GetMux())
}
