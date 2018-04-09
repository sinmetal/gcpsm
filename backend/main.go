package backend

import (
	"context"
	"fmt"
	"net/http"
	"reflect"
	"strings"

	"github.com/favclip/ucon"
	"github.com/favclip/ucon/swagger"
	"go.mercari.io/datastore"
	"go.mercari.io/datastore/aedatastore"
	"google.golang.org/appengine"
)

func init() {
	ucon.Middleware(UseAppengineContext)
	ucon.Orthodox()
	ucon.Middleware(swagger.RequestValidator())

	swPlugin := swagger.NewPlugin(&swagger.Options{
		Object: &swagger.Object{
			Info: &swagger.Info{
				Title:   "GCPSM",
				Version: "v1",
			},
			Schemes: []string{"http", "https"},
		},
		DefinitionNameModifier: func(refT reflect.Type, defName string) string {
			if strings.HasSuffix(defName, "JSON") {
				return defName[:len(defName)-4]
			}
			return defName
		},
	})
	ucon.Plugin(swPlugin)

	setupSecretAPI(swPlugin)

	ucon.DefaultMux.Prepare()
	http.Handle("/api/", ucon.DefaultMux)
}

// UseAppengineContext is UseAppengineContext
func UseAppengineContext(b *ucon.Bubble) error {
	if b.Context == nil {
		b.Context = appengine.NewContext(b.R)
	} else {
		b.Context = appengine.WithContext(b.Context, b.R)
	}

	return b.Next()
}

// FromContext is Create Datastore Client from Context
func FromContext(ctx context.Context) (datastore.Client, error) {
	return aedatastore.FromContext(ctx)
}

// HTTPError is API Resposeとして返すError
type HTTPError struct {
	Code    int         `json:"code"`
	Message interface{} `json:"message"`
}

// StatusCode is Http Response Status Codeを返す
func (he *HTTPError) StatusCode() int {
	return he.Code
}

// ErrorMessage is Clientに返すErrorMessageを返す
func (he *HTTPError) ErrorMessage() interface{} {
	return he
}

// Error is error interfaceを実装
func (he *HTTPError) Error() string {
	return fmt.Sprintf("status code %d: %s", he.StatusCode(), he.ErrorMessage())
}
