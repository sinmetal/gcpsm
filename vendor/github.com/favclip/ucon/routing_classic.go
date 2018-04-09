// +build !go1.7

package ucon

import (
	"context"
	"net/http"
)

func getDefaultContext(r *http.Request) context.Context {
	return context.Background()
}
