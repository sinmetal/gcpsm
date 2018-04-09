package ucon

import (
	"context"
	"net/http"
	"net/url"
	"regexp"
	"strings"
)

type methodMatchRate int

const (
	noMethodMatch methodMatchRate = iota
	starMethodMatch
	overloadMethodMatch
	exactMethodMatch
)

const (
	noPathMatch int = iota
	starPathMatch
	exactPathMatch
)

// Router is a handler to pass requests to the best-matched route definition.
// For the router decides the best route definition, there are 3 rules.
//
// 1. Methods must match.
//    a. If the method of request is `HEAD`, exceptionally `GET` definition is also allowed.
//    b. If the method of definition is `*`, the definition matches on all method.
// 2. Paths must match as longer as possible.
//    a. The path of definition must match to the request path completely.
//    b. Select the longest match.
//      * Against Request[/api/foo/hi/comments/1], Definition[/api/foo/{bar}/] is stronger than Definition[/api/foo/].
// 3. If there are multiple options after 1 and 2 rules, select the earliest one which have been added to router.
//
type Router struct {
	mux      *ServeMux
	handlers []*RouteDefinition
}

func (ro *Router) addRoute(rd *RouteDefinition) {
	ro.handlers = append(ro.handlers, rd)
}

// ServeHTTP routes a request to the handler and creates new bubble.
func (ro *Router) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	rd := ro.pickupBestRouteDefinition(r)

	if rd == nil {
		http.NotFound(w, r)
		return
	}

	ctx := getDefaultContext(r)

	match, params := rd.PathTemplate.Match(encodedPathFromRequest(r))
	if !match {
		http.Error(w, "[ucon] invalid handler picked", http.StatusInternalServerError)
		return
	}
	ctx = context.WithValue(ctx, PathParameterKey, params)

	b, err := ro.mux.newBubble(ctx, w, r, rd)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	err = b.Next()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (ro *Router) pickupBestRouteDefinition(r *http.Request) *RouteDefinition {
	// NOTE 動作概要
	// 設定されているHandler群から適切なものを選択しそれにroutingする
	// "適切なもの"の選び方は以下の通り
	// 1. Methodが一致する
	//   a. 指定MethodがHEADの場合、GETも探索する(明示的にHEADのものが優先される)
	//   b. 指定Methodが*の場合、すべてのMethodに対して一致するものとする
	// 2. RequestPathが一致する
	//   a. Handler側のパスは全長一致しなければならない
	//   b. より長いパス長のものを優先する /api/foo/bar なら3節 という数え方
	//      /api/foo/ と /api/foo/{bar}/ というHandlerがあったら、 /api/foo/hi/comments/1 は /api/foo/{bar}/ に割り当てられる
	// 3. 1,2での評価が最も高いものが複数ある場合は、より早くServeMuxに追加されたHandlerを選ぶ
	//
	// ルーティングのサンプル
	// Handler: OPTIONS / , POST /api/todo & Request: OPTIONS /api/todo -> OPTIONS / が選択される
	// Handler: * /api/todo/ , POST /api/todo/{id} & Request: GET /api/todo/1  -> * /api/todo/ が選択される

	var bestRoute *RouteDefinition
	var bestMethodMatchRate methodMatchRate
	var bestPathMatchRate int

	methodMatchRate := func(rd *RouteDefinition) methodMatchRate {
		if rd.Method == "*" {
			return starMethodMatch
		}

		if rd.Method == "GET" && r.Method == "HEAD" {
			return overloadMethodMatch
		}

		if rd.Method == r.Method {
			return exactMethodMatch
		}

		return noMethodMatch
	}

	pathMatchRate := func(rd *RouteDefinition) int {
		match, _ := rd.PathTemplate.Match(r.URL.Path)
		if !match {

			return noPathMatch
		}

		tempPathTokens := rd.PathTemplate.splittedPathTemplate
		reqPathTokens := strings.Split(r.URL.Path, "/")

		if len(reqPathTokens) < len(tempPathTokens) {
			// tempPath must not be longer than reqPath
			return noPathMatch
		}

		var rate int
		for i, token := range tempPathTokens {
			if i == 0 {
				// first token is always ""
				continue
			}
			if rd.PathTemplate.isVariables[i] {
				// variable token matches to everything.
				rate += exactPathMatch
				continue
			}
			if token == "" {
				// "/a/" can match to "/a/c", but it's weaker than exact match.
				rate += starPathMatch
				continue
			}
			if token == reqPathTokens[i] {
				rate += exactPathMatch
			}
		}

		return rate
	}

	for _, rd := range ro.handlers {
		mRate := methodMatchRate(rd)
		if mRate == noMethodMatch {
			continue
		} else if mRate < bestMethodMatchRate {
			continue
		}

		pRate := pathMatchRate(rd)
		if pRate == noPathMatch {
			continue
		} else if pRate < bestPathMatchRate {
			continue
		}

		if bestMethodMatchRate == mRate && bestPathMatchRate == pRate {
			continue
		}

		bestMethodMatchRate = mRate
		bestPathMatchRate = pRate
		bestRoute = rd
	}

	return bestRoute
}

// RouteDefinition is a definition of route handling.
// If a request matches on both the method and the path, the handler runs.
type RouteDefinition struct {
	Method           string
	PathTemplate     *PathTemplate
	HandlerContainer HandlerContainer
}

// PathTemplate is a path with parameters template.
type PathTemplate struct {
	PathTemplate         string
	httpHandlePath       string
	isVariables          []bool
	splittedPathTemplate []string
	PathParameters       []string
}

// Match checks whether PathTemplate matches the request path.
// If the path contains parameter templates, those key-value map is also returned.
func (pt *PathTemplate) Match(requestPath string) (bool, map[string]string) {
	if pt.PathTemplate == pt.httpHandlePath {
		// TODO ざっくりした実装なので後でリファクタリングすること
		if strings.HasPrefix(requestPath, pt.httpHandlePath) {
			return true, nil
		}
	}

	requestPathSplitted := strings.Split(requestPath, "/")
	if requiredLen, requestLen := len(pt.splittedPathTemplate), len(requestPathSplitted); requiredLen < requestLen {
		// I want to match /js/index.js to / :)
		requestPathSplitted = requestPathSplitted[0:requiredLen]
	} else if requiredLen != requestLen {
		return false, nil
	}

	params := make(map[string]string)
	for idx, s := range pt.splittedPathTemplate {
		reqPart := requestPathSplitted[idx]
		if pt.isVariables[idx] {
			if reqPart == "" {
				return false, nil
			}
			v, err := url.QueryUnescape(reqPart)
			if err != nil {
				v = reqPart
			}
			params[s[1:len(s)-1]] = v
		} else if s != reqPart {
			return false, nil
		} else {
			// match
		}
	}

	return true, params
}

// ParsePathTemplate parses path string to PathTemplate.
func ParsePathTemplate(pathTmpl string) *PathTemplate {
	tmpl := &PathTemplate{}
	tmpl.PathTemplate = pathTmpl
	vIndex := strings.Index(pathTmpl, "{")
	if vIndex == -1 {
		tmpl.httpHandlePath = pathTmpl
	} else {
		tmpl.httpHandlePath = pathTmpl[:vIndex]
	}

	tmpl.splittedPathTemplate = strings.Split(pathTmpl, "/")
	tmpl.isVariables = make([]bool, len(tmpl.splittedPathTemplate))

	re := regexp.MustCompile("^\\{(.+)\\}$")
	for idx, param := range tmpl.splittedPathTemplate {
		if re.MatchString(param) {
			key := re.FindStringSubmatch(param)[1]
			tmpl.PathParameters = append(tmpl.PathParameters, key)
			tmpl.isVariables[idx] = true
			continue
		}
	}

	return tmpl
}
