# ucon

uconはMiddlewareとPluginによる柔軟な拡張が可能な、Golangのウェブアプリケーションフレームワークです。

## インストール

```
go get -u github.com/favclip/ucon
```

## 使い方

uconを始めるには、サーバーを起動するためのgoファイルが必要です。まずは`main.go`を作成し、次のようにmain関数を実装します。

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

次に`go run`コマンドでサーバーを起動してみましょう。

```
go run main.go
```

Webブラウザで`localhost:8080`にアクセスすると、uconによって返された`Hello World!`という文字列が表示されるでしょう！

もっとサンプルが見たい方は`/sample`ディレクトリの中にあるいくつかの例を見てみるといいかもしれません。

## 特徴

* 標準の[net/http](https://golang.org/pkg/net/http/)との互換性
* 柔軟なルーティング設定
* Middlewareによるリクエストハンドラの拡張
  * 強力なDI（依存性注入）機構
* Pluginによるサーバー機能の拡張
* `Orthodox()`による標準的な機能の提供
* テスト支援のための便利なユーティリティ
* Google App Engineなどの様々なプラットフォームで利用可能
* [Swagger(Open API Initiative)](https://openapis.org/)に対応する`swagger`Plugin

### サーバーの起動
uconを使ったサーバーを起動するには`ucon.ListenAndServe`関数を実行します。
この関数は引数がアドレスだけであるという点を除いて、[`http.ListenAndServe`](https://golang.org/pkg/net/http/#ListenAndServe)と完全に互換性があります。

既存のサーバー上でuconを使用する場合は[「既存のサーバーへの組み込み」](#既存のサーバーへの組み込み)を参照してください。

### ルーティング
uconのルーティングは`Handle`関数、または`HandleFunc`関数によって設定されます。
`HandleFunc`関数はリクエストハンドラとして関数オブジェクトを登録します。
関数だけでは表現できない複雑なリクエストハンドリングを行いたい場合は、[HandlerContainer](http://godoc.org/github.com/favclip/ucon#HandlerContainer)インタフェースを実装したオブジェクトを`Handle`関数で登録することもできます。

uconが標準の`http`パッケージと違うのは、uconではルーティングの定義にHTTPのリクエストメソッドが必須であることです。
（これは[Google Cloud Endpoints](https://cloud.google.com/endpoints/?hl=en)やSwaggerといったプラットフォームとの親和性を高める上で必要なアプローチでした。）
同じパスへのリクエストでも、リクエストメソッドが違う場合はそれぞれにリクエストハンドラが個別に必要です。
ただし、ルーティング定義のリクエストメソッドをワイルドカード(`*`)に指定した場合はすべてのリクエストメソッドに対して有効なリクエストハンドラになります。

* `HandleFunc("GET", "/a/", ...)`は、GETリクエストの`/a/b`や`/a/b/c`などにはマッチしますが、
POSTリクエストの`/a/b`にはマッチしませんし、GETリクエストの`/a`にもマッチしません。
* `HandleFunc("*", "/", ...)`は、すべてのリクエストにマッチします。
* (1)`HandleFunc("GET", "/a", ...)`と(2)`HandleFunc("GET", "/a/", ...)`の2つがある場合、
GETリクエストの`/a`は(1)にマッチしますが、`/a/b`は(2)にマッチします。
* `HandleFunc("GET", "/users/{id}", ...)`は、`/users/1`や`/users/foo/bar`にマッチしますが、`/users`にはマッチしません。

### Middleware機能
uconにおけるMiddlewareとは、サーバーがリクエストを受け取ってからリクエストハンドラに届けられるまでの間に実行されるプリプロセッサのことです。
いくつかのMiddlewareがuconには標準で用意されており、`ucon.Orthodox()`を実行すると次のMiddlewareが読み込まれます。

* [ResponseMapper](http://godoc.org/github.com/favclip/ucon#ResponseMapper) - リクエストハンドラの戻り値をJSONに変換する
* [HTTPRWDI](http://godoc.org/github.com/favclip/ucon#HTTPRWDI) - `http.Request`と`http.ResponseWriter`のDIを行う
* [NetContextDI](http://godoc.org/github.com/favclip/ucon#NetContextDI) - `context.Context`のDIを行う
* [RequestObjectMapper](http://godoc.org/github.com/favclip/ucon#RequestObjectMapper) - リクエストに含まれるパラメータやデータをリクエストハンドラの引数にある型のオブジェクトに変換する

もちろん独自のMiddlewareを作成することもできます。ここでは例としてリクエストを受け取るたびに標準出力にログを書き込むMiddlewareを作ってみましょう。

Middlewareの実体は、`func(b *ucon.Bubble) error`で表現される関数です。
`main.go`に次のような`Logger`関数を定義します。

```go
func Logger(b *ucon.Bubble) error {
	fmt.Printf("Received: %s %s\n", b.R.Method, b.R.URL.String())
	return b.Next()
}
```

あとは`Logger`をMiddlewareとして登録するだけです。Middlewareを登録するには`ucon.Middleware()`関数を呼び出します。

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

これで`main.go`を実行してサーバーを立ち上げると、リクエストを受け取るたびにログを出力するようになりました。

Middlewareに与えられる`Bubble`は、リクエストがサーバーに届いてから、適切なリクエストハンドラに到達までの間のデータの運搬を行っています。
`Bubble.Next()`を呼び出すと、次のMiddlewareに処理が移り、すべてのMiddlewareが処理を終えたら`ucon.Handle`か`ucon.HandleFunc`で登録された
リクエストハンドラの処理が実行されます。

### DI機能
uconのDI機構はMiddlewareが`Bubble.Arguments`に値を格納することで解決しています。
リクエストハンドラとして登録された関数の引数は`Bubble.ArgumentTypes`に格納されるため、
Middlewareで`Bubble.ArgumentTypes`を見て、任意の型に対して値を与えることができます。
例えば、リクエストハンドラに`time.Time`型の引数を追加する際は、次のようなMiddlewareでDIを追加することができます。

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

### Plugin機能
uconにおけるPluginとは、ucon全体の機能を拡張するためのプリプロセッサです。
Middlewareのようにリクエストが来るたびに実行されるのではなく、サーバーが起動する際に一度だけ実行されます。
sampleの`swagger`は、具体的なPluginの使い方を知るための手助けになるでしょう。

Pluginを実装するには、Pluginのインタフェースを実装した構造体のオブジェクトを`ucon.Plugin()`関数で登録します。
現在、uconでは以下のPluginインタフェースが提供されています。

- [HandlersScannerPlugin](http://godoc.org/github.com/favclip/ucon#HandlersScannerPlugin) - uconに登録されたリクエストハンドラの一覧を取得できる

Pluginには`*ServeMux`が引数で与えられるので、PluginによってリクエストハンドラやMiddlewareを追加することも可能です。

### テスト支援機能
uconではuconを使ったアプリケーションの単体テストを行うのに便利なユーティリティも提供しています。

#### [MakeMiddlewareTestBed](http://godoc.org/github.com/favclip/ucon#MakeMiddlewareTestBed)
`MakeMiddlewareTestBed` はMiddlewareのテストを行うためのテストベッドを提供します。
例えば、`NetContextDI`Middlewareのテストは次のように記述されています。

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
`MakeHandlerTestBed`はリクエストハンドラのテストを行うためのテストベッドを提供します。
この関数を呼び出す前にリクエストハンドラが登録されている必要があります。

routing_test.goではこの関数を使って次のようなテストが記述されています。

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

### 既存のサーバーへの組み込み
uconの`ServeMux`は`http.Handler`を実装しているので、`http.Handle`関数に渡すことで
既存のGolangのサーバーに簡単に組み込むことができます。
ただし`Handle`関数に渡す前に、通常は`ucon.ListenAndServe`の内部で実行されている`ServeMux#Prepare`関数を明示的に実行する必要があります。

デフォルトの`ServeMux`の参照は、`ucon.DefaultMux`で取得することができます。

```go
func init() {
	ucon.Orthodox()

    ...
    
	ucon.DefaultMux.Prepare()
	http.Handle("/", ucon.DefaultMux)
}
```

## ライセンス
MIT

