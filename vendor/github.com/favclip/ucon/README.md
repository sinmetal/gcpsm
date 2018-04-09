# ucon

ucon is a web application framework, which is pluggable by Middleware and Plugin.

_ucon_ is the name of turmeric in Japanese. ucon knocks down any alcohol. :)

## Install

```
go get -u github.com/favclip/ucon
```

## Get Start

Getting start using ucon, you should setup the http server. 
If you decided to using ucon, it is very simple. Let's take a look at the following code.

```go
package main

import (
	"net/http"
	
	"github.com/favclip/ucon"
)

func main() {
	ucon.Orthodox()

	ucon.HandleFunc("GET", "/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Hello World!"))
	})

	ucon.ListenAndServe(":8080")
}
```

Next, execute `go run` to run the server.

```
go run main.go
```

Then, you can get `Hello World!` on `localhost:8080`.

You can find more examples in `/sample` directory if you want.

## Features

* Compatible interface with [net/http](https://golang.org/pkg/net/http/)
* Flexible routing configuration
* **Middleware** - Extendable request handler
  * Powerful DI (Dependency Injection) system
* **Plugin** - Easy to customize the server
* `Orthodox()` - Standard features provider 
* Helpful utilities for testing
* Run on Google App Engine and more platform
* Opt-in plugin for [Swagger(Open API Initiative)](https://openapis.org/)

### Run a server
`ucon.ListenAndServe` function starts new http server.
This function is fully compatible with [`http.ListenAndServe`](https://golang.org/pkg/net/http/#ListenAndServe) except which doesn't have second argument.

If you want to use ucon on existing server, see ["With existing server"](#with-existing-server).

### Routing
Routing of ucon is set by the `Handle` function or `HandleFunc` function.
`HandleFunc` registers the function object as a request handler.
If you want to make a complex request handler, 
you can use `Handle` function with a object which implements [HandlerContainer](http://godoc.org/github.com/favclip/ucon#HandlerContainer) interface.

The different from `http` package is that ucon requires HTTP request method to configure routing.
(This is a necessary approach to improve friendliness for some platforms like [Google Cloud Endpoints](https://cloud.google.com/endpoints/?hl=en) or Swagger.) 
Even if paths are same, different request handler is required for each request method.
However, the request handler for the wildcard (`*`) is valid for all of the request method.

* `HandleFunc("GET", "/a/", ...)` will match a GET request for `/a/b` or `/a/b/c`, but won't match a POST request for `/a/b` or a GET request for `/a`.
* `HandleFunc("*", "/", ...)` can match any requests.
* If there are two routing, (A)`HandleFunc("GET", "/a", ...)` and (B)`HandleFunc("GET", "/a/", ...)`, 
 a request for `/a` will match (A), but a request for `/a/b` will match (B).
* `HandleFunc("GET", "/users/{id}", ...)` will match requests like `/users/1` or `/users/foo/bar` but won't match `/users`.

### Middleware
Middleware is a preprocessor which is executed in between server and request handler.
Some of Middleware are provided as standard, and when you run the `ucon.Orthodox()`, the following Middleware will be loaded.

* [ResponseMapper](http://godoc.org/github.com/favclip/ucon#ResponseMapper) - Converts the return value of the request handler to JSON.
* [HTTPRWDI](http://godoc.org/github.com/favclip/ucon#HTTPRWDI) - Injects dependencies of `http.Request` and `http.ResponseWriter`.
* [NetContextDI](http://godoc.org/github.com/favclip/ucon#NetContextDI) - Injects `context.Context` dependency.
* [RequestObjectMapper](http://godoc.org/github.com/favclip/ucon#RequestObjectMapper) - Convert the parameters and data in the request to the argument of the request handler.

Of course, you can create your own Middleware. 
As an example, let's create a Middleware to write the log to stdout each time it receives a request.

Middleware is a function expressed as `func(b *ucon.Bubble) error`.
Write `Logger` function into `main.go`.

```go
func Logger(b *ucon.Bubble) error {
	fmt.Printf("Received: %s %s\n", b.R.Method, b.R.URL.String())
	return b.Next()
}
```

Next, Register `Logger` as Middleware. Call `ucon.Middleware()` to register a Middleware.

```go
package main

import (
	"fmt"
	"net/http"
	
	"github.com/favclip/ucon"
)

func main() {
	ucon.Orthodox()

	ucon.Middleware(Logger)

	ucon.HandleFunc("GET", "/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Hello World!"))
	})

	ucon.ListenAndServe(":8080")
}

func Logger(b *ucon.Bubble) error {
	fmt.Printf("Received: %s %s\n", b.R.Method, b.R.URL.String())
	return b.Next()
}
```

OK! The server will output a log each time it receives the request.

Bubble given to Middleware will carry data until the request reaches the appropriate request handler.
`Bubble.Next()` passes the processing to next Middleware. When All of Middleware has been called, the request handler will be executed.

### DI (Dependency Injection)
DI system of ucon is solved by that the Middleware provides the data to `Bubble.Arguments`.
The types of request handler's arguments are contained in `Bubble.ArgumentTypes` and so Middleware can give a value to each type.

For example, when you add the argument of `time.Time` type in the request handler, you can add the DI in the following Middleware.

```go
package main

import (
	"fmt"
	"net/http"
	"reflect"
	"time"

	"github.com/favclip/ucon"
)

func main() {
	ucon.Orthodox()

	ucon.Middleware(NowInJST)

	ucon.HandleFunc("GET", "/", func(w http.ResponseWriter, r *http.Request, now time.Time) {
		w.Write([]byte(
		    fmt.Sprintf("Hello World! : %s", now.Format("2006/01/02 15:04:05")))
		)
	})

	ucon.ListenAndServe(":8080")
}

func NowInJST(b *ucon.Bubble) error {
	for idx, argT := range b.ArgumentTypes {
		if argT == reflect.TypeOf(time.Time{}) {
			b.Arguments[idx] = reflect.ValueOf(time.Now())
			break
		}
	}
	return b.Next()
}
```

### Plugin
Plugin is a preprocessor to customize the server.
Plugin is not executed each time that comes request like Middleware. It will be executed only once when the server prepares.

sample of `swagger` will be a help to know how to use a Plugin.

To create a Plugin, you can register the object implements the interface of the Plugin in `ucon.Plugin` function.
Currently, the following Plugin interfaces are provided.

- [HandlersScannerPlugin](http://godoc.org/github.com/favclip/ucon#HandlersScannerPlugin) - Get a list of request handlers registered

Because the Plugin is given `*ServeMux`, it is also possible to add the request handler and Middleware by Plugin.

### Testing Helper
ucon provides a useful utility to make the unit tests.

#### [MakeMiddlewareTestBed](http://godoc.org/github.com/favclip/ucon#MakeMiddlewareTestBed)
`MakeMiddlewareTestBed` provides a test bed for testing the Middleware.
For example, test of `NetContextDI` Middleware is described as follows.

```go
func TestNetContextDI(t *testing.T) {
	b, _ := MakeMiddlewareTestBed(t, NetContextDI, func(c context.Context) {
		if c == nil {
			t.Errorf("unexpected: %v", c)
		}
	}, nil)
	err := b.Next()
	if err != nil {
		t.Fatal(err)
	}
}
```

#### [MakeHandlerTestBed](http://godoc.org/github.com/favclip/ucon#MakeHandlerTestBed)
`MakeHandlerTestBed` provides a test bed for testing the request handler.
Request handler must have been registered before calling this function.

In routing_test.go this function is used to describe a test such as the following.

```go
func TestRouterServeHTTP1(t *testing.T) {
	DefaultMux = NewServeMux()
	Orthodox()

	HandleFunc("PUT", "/api/test/{id}", func(req *RequestOfRoutingInfoAddHandlers) (*ResponseOfRoutingInfoAddHandlers, error) {
		if v := req.ID; v != 1 {
			t.Errorf("unexpected: %v", v)
		}
		if v := req.Offset; v != 100 {
			t.Errorf("unexpected: %v", v)
		}
		if v := req.Text; v != "Hi!" {
			t.Errorf("unexpected: %v", v)
		}
		return &ResponseOfRoutingInfoAddHandlers{Text: req.Text + "!"}, nil
	})

	DefaultMux.Prepare()

	resp := MakeHandlerTestBed(t, "PUT", "/api/test/1?offset=100", strings.NewReader("{\"text\":\"Hi!\"}"))

	if v := resp.StatusCode; v != 200 {
		t.Errorf("unexpected: %v", v)
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		t.Fatal(err)
	}
	if v := string(body); v != "{\"text\":\"Hi!!\"}" {
		t.Errorf("unexpected: %v", v)
	}
}
```

### With existing server
Because ucon's `ServeMux` implements `http.Handler` interface, it can be easily integrated into existing Golang server by passing the `http.Handle` function.
However before passing it to `Handle` function, you need to explicitly call `ServeMux#Prepare` function instead of `ucon.ListenAndServe`.

You can get default reference of `ServeMux` by `ucon.DefaultMux`.

```go
func init() {
	ucon.Orthodox()

    ...
    
	ucon.DefaultMux.Prepare()
	http.Handle("/", ucon.DefaultMux)
}
```

## LICENSE
MIT

