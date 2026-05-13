package httprouter

import (
	"net/http"
)

type WebApplication interface {
	Post(path string, handler http.Handler, mws ...Middleware)
	Get(path string, handler http.Handler, mws ...Middleware)
	Put(path string, handler http.Handler, mws ...Middleware)
	Patch(path string, handler http.Handler, mws ...Middleware)
	Delete(path string, handler http.Handler, mws ...Middleware)
	WithGlobalMiddlewares(mws ...Middleware) WebApplication
}

type Middleware func(Handler http.Handler) http.HandlerFunc
