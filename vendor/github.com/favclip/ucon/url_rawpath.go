// +build go1.5

package ucon

import (
	"net/http"
)

func encodedPathFromRequest(r *http.Request) string {
	return r.URL.EscapedPath()
}
