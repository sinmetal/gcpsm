// +build !go1.5

package ucon

import (
	"net/http"
)

func encodedPathFromRequest(r *http.Request) string {
	// url.EscapedPath() が欲しいがappengine環境下ではgo1.4で元データが存在しないのでごまかす /page/foo%2Fbar みたいな構造がうまく処理できない 解決は不可能という認識…
	// r.RequestURI から自力で頑張ればイケる…？？
	return r.URL.Path
}
