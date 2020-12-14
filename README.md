# An evolutive multiplexer for HTTP routes

Read HTTP requests, tests matchers then route the good handler.

## Add in your project

Load module:

```sh
go get github.com/karkael64/golang-mux
```

## Use in your project file

Here is an exemple for adding a route in your HttpServer.

```go
import (
  "fmt"
  mux "github.com/karkael64/golang-mux"
)

func main() {
  var address = "localhost:4242"
  var srv = mux.New()

  srv.AddRoute(getWelcomeRoute())

  fmt.Printf("Starting server at https://%s/\n", address)
  err := srv.Listen(address, "keys/cert.pem", "keys/priv.key")
  if err != nil {
    fmt.Println(err)
    os.Exit(1)
  }
}

func getWelcomeRoute() *mux.Route {
	return mux.CreateRoute(
    mux.MatchPathExact([]string{"GET"}, `/welcome`),
    func (w http.ResponseWriter, r *http.Request) error {
      w.Write("Welcome!")
    },
  )
}
```

## Create a route

A route is composed by a matcher and a handler:

* The matcher is a function which tests the request. It returns a success boolean and an error. If it returns a success, the handler is executed, else it test next route matcher. If any matcher returns an error and none matchs the request, then the error is thrown to client. It is helpfull for forbidden requests (code 403).
* The handler is a function which write the response HTTP for client. It expect to write a code, headers and a body for client and return `nil`. The handler can return error about to be thrown to client.

```go
type Handler func(w http.ResponseWriter, r *http.Request) error
type Matcher func(r *http.Request) (bool, error)
type Route struct {
	handler Handler
	match   Matcher
}
```

You can write a route with the function `mux.CreateRoute`:

```go
func matcher(r *http.Request) (bool, error) {
  if !len(r.Header()["Token"][0]) {
    return false, mux.NewHttpError(403, "Forbidden")
  }
  return true
}

func handler(w http.ResponseWriter, r *http.Request) error {
  w.Write("Allowed")
}

func init() {
  var route = mux.CreateRoute(matcher, handler)
}
```

## Create a simple file handler

Create a handler for a file with `GET` method. The example below does match exactly `/file` path in request and returns the file content at path `./file`, with header

```go
func init() {
  var fileRoute = mux.CreateFileRoute(
    "/file",
    "./file",
    http.Header{"Field":{"Value"}}
  )
}
```

## Create a route matching regexp

```go
func init() {
  var regexpMatcher = mux.MatchPathRegexp([]string{"GET"}, `/regexp/\w+`)
}
```

The matcher can return a boolean and an error. The error could be:
* The method is not allowed, return HttpError(403)
* The regexp is not well formatted, return the regexp error

## Notes

* File `.env` can set `HTTP_DEBUG` status:
  * if not set, does not display any debug texts
  * if set by `HTTP_DEBUG=full`, display every errors found
  * if set with empty value `HTTP_DEBUG` (or any value not `full`), display only working directory errors
