package ucon

import (
	"context"
	"fmt"
	"net/http"
	"strings"
)

// DefaultMux is the default ServeMux in ucon.
var DefaultMux = NewServeMux()

// NewServeMux allocates and returns a new ServeMux.
func NewServeMux() *ServeMux {
	mux := &ServeMux{
		router: &Router{},
	}
	mux.router.mux = mux
	return mux
}

// ServeMux is an HTTP request multiplexer.
type ServeMux struct {
	Debug       bool
	router      *Router
	middlewares []MiddlewareFunc
	plugins     []*pluginContainer
}

// MiddlewareFunc is an adapter to hook middleware processing.
// Middleware works with 1 request.
type MiddlewareFunc func(b *Bubble) error

// Middleware can append Middleware to ServeMux.
func (m *ServeMux) Middleware(f MiddlewareFunc) {
	m.middlewares = append(m.middlewares, f)
}

// Plugin can append Plugin to ServeMux.
func (m *ServeMux) Plugin(plugin interface{}) {
	p, ok := plugin.(*pluginContainer)
	if !ok {
		p = &pluginContainer{base: plugin}
	}
	p.check()
	m.plugins = append(m.plugins, p)
}

// Prepare the ServeMux.
// Plugin is not show affect to anything.
// This method is enabled plugins.
func (m *ServeMux) Prepare() {
	for _, plugin := range m.plugins {
		used := false
		if sc := plugin.HandlersScanner(); sc != nil {
			err := sc.HandlersScannerProcess(m, m.router.handlers)
			if err != nil {
				panic(err)
			}
			used = true
		}
		if !used {
			panic(fmt.Sprintf("unused plugin: %#v", plugin))
		}
	}
}

// ServeHTTP dispatches request to the handler.
func (m *ServeMux) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// NOTE ucon内部のルーティングは一元的にこの関数から行う
	// Handlerを細分化しhttp.ServeMuxに登録すると、OPTIONSのhandleがうまくできなくなる
	// このため、Handlerはucon全体で1つとし、OPTIONSも通常のMethodと同じようにHandlerを設定し利用する
	// OPTIONSを適切にhandleするため、全てのHandlerに特殊なHookを入れるよりマシである

	m.router.ServeHTTP(w, r)
}

// ListenAndServe start accepts the client request.
func (m *ServeMux) ListenAndServe(addr string) error {
	m.Prepare()

	server := &http.Server{Addr: addr, Handler: m}
	return server.ListenAndServe()
}

// Handle register the HandlerContainer for the given method & path to the ServeMux.
func (m *ServeMux) Handle(method string, path string, hc HandlerContainer) {
	CheckFunction(hc.Handler())

	pathTmpl := ParsePathTemplate(path)
	methods := strings.Split(strings.ToUpper(method), ",")
	for _, method := range methods {
		rd := &RouteDefinition{
			Method:           method,
			PathTemplate:     pathTmpl,
			HandlerContainer: hc,
		}
		m.router.addRoute(rd)
	}
}

// HandleFunc register the handler function for the given method & path to the ServeMux.
func (m *ServeMux) HandleFunc(method string, path string, h interface{}) {
	m.Handle(method, path, &handlerContainerImpl{
		handler: h,
		Context: background,
	})
}

func (m *ServeMux) newBubble(c context.Context, w http.ResponseWriter, r *http.Request, rd *RouteDefinition) (*Bubble, error) {
	b := &Bubble{
		R:              r,
		W:              w,
		Context:        c,
		RequestHandler: rd.HandlerContainer,
	}
	err := b.init(m)
	if err != nil {
		return nil, err
	}

	return b, nil
}

// HandlerContainer is handler function container.
// and It has a ucon Context that make it possible communicate to Plugins.
type HandlerContainer interface {
	Handler() interface{}
	Context
}

type handlerContainerImpl struct {
	handler interface{}
	Context
}

func (hc *handlerContainerImpl) Handler() interface{} {
	return hc.handler
}

// Orthodox middlewares enable to DefaultServeMux.
func Orthodox() {
	DefaultMux.Middleware(ResponseMapper())
	DefaultMux.Middleware(HTTPRWDI())
	DefaultMux.Middleware(ContextDI())
	DefaultMux.Middleware(RequestObjectMapper())
}

// Middleware can append Middleware to ServeMux.
func Middleware(f MiddlewareFunc) {
	DefaultMux.Middleware(f)
}

// Plugin can append Plugin to ServeMux.
func Plugin(plugin interface{}) {
	DefaultMux.Plugin(plugin)
}

// ListenAndServe start accepts the client request.
func ListenAndServe(addr string) {
	DefaultMux.ListenAndServe(addr)
}

// Handle register the HandlerContainer for the given method & path to the ServeMux.
func Handle(method string, path string, hc HandlerContainer) {
	DefaultMux.Handle(method, path, hc)
}

// HandleFunc register the handler function for the given method & path to the ServeMux.
func HandleFunc(method string, path string, h interface{}) {
	DefaultMux.HandleFunc(method, path, h)
}
