package ucon

import "fmt"

type pluginContainer struct {
	base interface{}
}

// HandlersScannerPlugin is an interface to make a plugin for scanning request handlers.
type HandlersScannerPlugin interface {
	HandlersScannerProcess(m *ServeMux, rds []*RouteDefinition) error
}

func (p *pluginContainer) check() {
	if p.HandlersScanner() != nil {
		return
	}

	panic(fmt.Sprintf("unused plugin: %#v", p.base))
}

// HandlersScanner returns itself if it implements HandlersScannerPlugin.
func (p *pluginContainer) HandlersScanner() HandlersScannerPlugin {
	if v, ok := p.base.(HandlersScannerPlugin); ok {
		return v
	}

	return nil
}

type emptyCtx int

var background = new(emptyCtx)

func (*emptyCtx) Value(key interface{}) interface{} {
	return nil
}

// Context is a key-value store.
type Context interface {
	Value(key interface{}) interface{}
}

// WithValue returns a new context containing the value.
// Values contained by parent context are inherited.
func WithValue(parent Context, key interface{}, val interface{}) Context {
	return &valueCtx{parent, key, val}
}

type valueCtx struct {
	Context
	key interface{}
	val interface{}
}

func (c *valueCtx) Value(key interface{}) interface{} {
	if c.key == key {
		return c.val
	}
	return c.Context.Value(key)
}
