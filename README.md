# Go Core Router

Go Core Router is a Go library for handling HTTP routing with customizable middleware and support for route groups with prefixes. This library makes it easier to organize routes and middleware in Go applications, separating route handling logic, and keeping the application well-structured.

## Key Features

- **Group Route**: With group routes, each endpoint on the same router object will have a prefix, group middleware, and exceptions for middleware groups from its parent router object. Changes to the child do not affect the parent, and changes to the parent after this feature is invoked will not modify the child.
- **Middlewares Route**: Each route object can apply multiple middlewares simultaneously.
- **Except Middlewares Route**: By applying this, the router can register several middlewares that should not be executed.
- **Specific Endpoint Middlewares**: Multiple middlewares can be applied to specific endpoints, and this does not affect other endpoints or the router object.
- **Except Specific Endpoint Middlewares**: By applying this, an endpoint can register several middlewares that should not be executed on specific endpoints. This feature will not affect other endpoints or the router object.
- **Prefix Routes**: All endpoints on this router object and its children will share the same prefix.

## Installation

To install this library, run the following command in your terminal:

```bash
go get github.com/RasyidHakim-2804/go-core-router
```

Here's an example of a Go application that uses `go-core-router`:

```go
package main

import (
    "fmt"
    "net/http"
    "github.com/RasyidHakim-2804/go-core-router"
)

func handler(w http.ResponseWriter, r *http.Request) {
    message := "Hello world"
    w.Write([]byte(message))
}

func main() {
    router := core.NewRouter()
    router.Get("/index", handler)
    http.ListenAndServe(":8000", router.GetMux())
}
```
## Groupe Route

```go

router.Group(func(router2 *core.Router) {
		router2.Get("/home2", handler2)
	})

```

## Making Middleware
```go
type GlobalMiddleware struct {
}

func (gm GlobalMiddleware) Before(w http.ResponseWriter, r *http.Request) bool {
	fmt.Println("this is global middleware")
	if contentType := r.Header.Get("Accept"); contentType != "application/json" {

		http.Error(w, "header 'Accept' must 'application/json'", http.StatusUnsupportedMediaType)

		return false
	}
	return true
}
```
## Prefix

```go
router.Group(func(router2 *core.Router) {
		router2.Prefix("/home")

		router2.Get("/index", homeHandler).ExceptMiddlewares(FirstMiddleware{})   // /home/index
		router2.Get("/about", aboutHandler).ExceptMiddlewares(GlobalMiddleware{}) // /home/about
	})
```

## Full Example

```go
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


```