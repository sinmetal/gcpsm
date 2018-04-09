package ucon

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"reflect"
)

// ErrInvalidRequestHandler is the error that Bubble.RequestHandler is not a function.
var ErrInvalidRequestHandler = errors.New("invalid request handler. not function")

// ErrInvalidArgumentLength is the error that length of Bubble.Arguments does not match to RequestHandler arguments.
var ErrInvalidArgumentLength = errors.New("invalid arguments")

// ErrInvalidArgumentValue is the error that value in Bubble.Arguments is invalid.
var ErrInvalidArgumentValue = errors.New("invalid argument value")

// Bubble is a context of data processing that will be passed to a request handler at last.
// The name `Bubble` means that the processing flow is a event-bubbling.
// Processors, called `middleware`, are executed in order with same context, and at last the RequestHandler will be called.
type Bubble struct {
	R              *http.Request
	W              http.ResponseWriter
	Context        context.Context
	RequestHandler HandlerContainer

	Debug bool

	Handled       bool
	ArgumentTypes []reflect.Type
	Arguments     []reflect.Value
	Returns       []reflect.Value

	queueIndex int
	mux        *ServeMux
}

func (b *Bubble) checkHandlerType() error {
	if _, ok := b.RequestHandler.(HandlerContainer); ok {
		return nil
	}
	hv := reflect.ValueOf(b.RequestHandler)
	if hv.Type().Kind() == reflect.Func {
		return nil
	}

	return ErrInvalidRequestHandler
}

func (b *Bubble) handler() interface{} {
	if hv, ok := b.RequestHandler.(HandlerContainer); ok {
		return hv.Handler()
	}
	hv := reflect.ValueOf(b.RequestHandler)
	if hv.Type().Kind() == reflect.Func {
		return b.RequestHandler
	}

	return nil
}

func (b *Bubble) init(m *ServeMux) error {
	err := b.checkHandlerType()
	if err != nil {
		return err
	}

	hv := reflect.ValueOf(b.handler())
	numIn := hv.Type().NumIn()
	b.ArgumentTypes = make([]reflect.Type, numIn)
	for i := 0; i < numIn; i++ {
		b.ArgumentTypes[i] = hv.Type().In(i)
	}
	b.Arguments = make([]reflect.Value, numIn)

	b.mux = m
	b.Debug = m.Debug

	return nil
}

// Next passes the bubble to next middleware.
// If the bubble reaches at last, RequestHandler will be called.
func (b *Bubble) Next() error {
	if b.queueIndex < len(b.mux.middlewares) {
		qi := b.queueIndex
		b.queueIndex++
		m := b.mux.middlewares[qi]
		err := m(b)
		return err
	}

	return b.do()
}

func (b *Bubble) do() error {
	hv := reflect.ValueOf(b.handler())

	if len(b.Arguments) != len(b.ArgumentTypes) || len(b.Arguments) != hv.Type().NumIn() {
		return ErrInvalidArgumentLength
	}
	for idx, arg := range b.Arguments {
		if !arg.IsValid() {
			fmt.Printf("ArgumentInvalid %d\n", idx)
			return ErrInvalidArgumentValue
		}
		if !arg.Type().AssignableTo(hv.Type().In(idx)) {
			fmt.Printf("TypeMismatch %d, %+v, %+v\n", idx, b.Arguments[idx], hv.Type().In(idx))
			return ErrInvalidArgumentValue
		}
	}

	b.Returns = hv.Call(b.Arguments)

	b.Handled = true

	return nil
}
