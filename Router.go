package patriot_router

import (
	"fmt"
	"net/http"
	"strings"
)

type Router struct {
	routes          map[*Route]func(http.ResponseWriter, *http.Request)
	prefixes        map[string]*http.Handler
	cors_middleware func(func(http.ResponseWriter, *http.Request)) http.HandlerFunc // by default will just return the handlerFunc
}

func (self *Router) RegisterRoute(route *Route, handler func(http.ResponseWriter, *http.Request)) {
	self.routes[route] = handler
}

func (self *Router) RedirectIfPrefix(prefix string, handler http.Handler) {
	/*

		when serving the router will check if a request matches the prefix after checking for fullmatches and if
		the request url has the given prefix it will pass the request to the handler. this function registers a handler to a
		prefix. if the prefix is already registered it will overwrite the handler for that prefix.

	*/
	if prefix != "" {
		self.prefixes[prefix] = &handler
	} else {
		// log a warnings
		fmt.Println("Warning: prefix was empty, it was not registred")
	}
}

func (self *Router) SetCorsHandler(handler func(func(http.ResponseWriter, *http.Request)) http.HandlerFunc) {
	self.cors_middleware = handler
}

func (self *Router) ServeHTTP(response http.ResponseWriter, request *http.Request) {
	request_path := request.URL.Path
	fmt.Printf("Serving '%s' for %s by %s\n", request.Method, request_path, request.RemoteAddr)
	for route, handler := range self.routes {
		if route.Match(request_path) {
			self.cors_middleware(handler)(response, request)
			return
		}
	}

	for prefix, handler := range self.prefixes {
		if strings.HasPrefix(request_path, prefix) {
			self.cors_middleware((*handler).ServeHTTP)(response, request)
			return
		}
	}
	fmt.Println("Route had not handler set")

	response.WriteHeader(404)
	response.Write([]byte("not found"))
}

func CorsAllowAll(handler func(http.ResponseWriter, *http.Request)) http.HandlerFunc {
	return func(response http.ResponseWriter, request *http.Request) {
		response.Header().Set("Access-Control-Allow-Origin", "*")
		response.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
		response.Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")
		handler(response, request)
	}
}

func CreateRouter() *Router {
	var new_router *Router = new(Router)
	new_router.routes = make(map[*Route]func(http.ResponseWriter, *http.Request))
	new_router.prefixes = make(map[string]*http.Handler)
	new_router.cors_middleware = func(handler func(http.ResponseWriter, *http.Request)) http.HandlerFunc {
		return handler
	}
	return new_router
}
