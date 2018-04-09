package ucon

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
)

// BubbleTestOption is an option for setting a mock request.
type BubbleTestOption struct {
	Method            string
	URL               string
	Body              io.Reader
	MiddlewareContext Context
}

// MakeMiddlewareTestBed returns a Bubble and ServeMux for handling the request made from the option.
func MakeMiddlewareTestBed(t *testing.T, middleware MiddlewareFunc, handler interface{}, opts *BubbleTestOption) (*Bubble, *ServeMux) {
	if opts == nil {
		opts = &BubbleTestOption{
			Method: "GET",
			URL:    "/api/tmp",
		}
	}
	if opts.MiddlewareContext == nil {
		opts.MiddlewareContext = background
	}
	mux := NewServeMux()
	mux.Middleware(middleware)

	r, err := http.NewRequest(opts.Method, opts.URL, opts.Body)
	if err != nil {
		t.Fatal(err)
	}

	if opts.Body != nil {
		r.Header.Add("Content-Type", "application/json")
	}

	w := httptest.NewRecorder()

	u, err := url.Parse(opts.URL)
	if err != nil {
		t.Fatal(err)
	}

	rd := &RouteDefinition{
		Method:       opts.Method,
		PathTemplate: ParsePathTemplate(u.Path),
		HandlerContainer: &handlerContainerImpl{
			handler: handler,
			Context: opts.MiddlewareContext,
		},
	}

	b, err := mux.newBubble(context.Background(), w, r, rd)
	if err != nil {
		t.Fatal(err)
	}

	return b, mux
}

// MakeHandlerTestBed returns a response by the request made from arguments.
// To test some handlers, those must be registered by Handle or HandleFunc before calling this.
func MakeHandlerTestBed(t *testing.T, method string, path string, body io.Reader) *http.Response {
	ts := httptest.NewServer(DefaultMux)
	defer ts.Close()

	reqURL, err := url.Parse(ts.URL)
	if err != nil {
		t.Fatal(err)
	}
	reqURL, err = reqURL.Parse(path)
	if err != nil {
		t.Fatal(err)
	}

	req, err := http.NewRequest(method, reqURL.String(), body)
	if err != nil {
		t.Fatal(err)
	}
	if body != nil {
		req.Header.Add("Content-Type", "application/json")
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}

	return resp
}
