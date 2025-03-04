# Go Core Router

Go Core Router adalah pustaka Go untuk menangani routing HTTP dengan middleware yang dapat disesuaikan dan mendukung grup rute dengan prefix. Pustaka ini memberikan kemudahan untuk mengatur rute dan middleware dalam aplikasi Go, memisahkan logika pengolahan rute, serta menjaga aplikasi tetap terstruktur dengan baik.

## Fitur Utama

- **Group Route**: Dengan group route, maka setiap endpoint pada objek router yang sama akan memiliki prefix, group middleware, dan pengecualian grup middleware dari objek router parentnya. Perubahan pada child tidak mempengaruhi parent, dan perubahan pada parent setelah baris dipanggilnya ftur ini tidak akan mengubah child sebelumnya. 
- **Middlewares Route**: Setiap objek rute bisa menerapkan beberapa middleware sekaligus.
- **Except Middlewares Route**: Dengan menerapkan ini maka router bisa mendaftarkan beberapa middleware yang tidak ingin dijalankan. 
- **Middlewares Spesifik Endpoint**: Beberapa Middleware bisa diterapkan pada spesifik endpoint tertentu, dan ini tidak mempengaruhi endpoint lainnya dan objek routernya.
- **Except Middlewares Spesifik Endpoint**: Dengan menerapkan ini, maka endpoint bisa mendaftarkan beberapa middlweare yang tidak ingin dijalankan pada spesifik endpoint tertentu. Fitur ini tidak akan mempengaruhi endpoint lainnya dan objek routernya.
- **Prefix Routw**: Semua endpoint pada objek router ini dan anaknya akan memiliki prefix yang sama.

## Instalasi

Untuk menginstal pustaka ini, jalankan perintah berikut di terminal:

```bash
go get github.com/RasyidHakim-2804/go-core-router
```

Berikut adalah contoh aplikasi Go yang menggunakan `go-core-router`:

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

func (gm GlobalMiddleware) Next(w http.ResponseWriter, r *http.Request) bool {
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
}

type FirstMiddleware struct {
}

func (gm GlobalMiddleware) Next(w http.ResponseWriter, r *http.Request) bool {
	fmt.Println("this is global middleware")
	if contentType := r.Header.Get("Accept"); contentType != "application/json" {

		http.Error(w, "header 'Accept' must 'application/json'", http.StatusUnsupportedMediaType)

		return false
	}
	return true
}

func (fm FirstMiddleware) Next(w http.ResponseWriter, r *http.Request) bool {
	fmt.Println("this is first Middleware")
	return true
}

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