package httprouter

import (
	"net/http"
	pth "path"
)

type ApplicationGroup struct {
	a    WebApplication
	mws  []Middleware
	path string
}

func (g *ApplicationGroup) Group(path string, mws ...Middleware) WebApplication {
	return &ApplicationGroup{
		a:    g,
		mws:  mws,
		path: path,
	}
}

func (g *ApplicationGroup) Post(path string, handler http.Handler, mws ...Middleware) {
	p, h, m := g.group(path, handler, mws...)
	g.a.Post(p, h, m...)
}

func (g *ApplicationGroup) Get(path string, handler http.Handler, mws ...Middleware) {
	p, h, m := g.group(path, handler, mws...)
	g.a.Get(p, h, m...)
}

func (g *ApplicationGroup) Put(path string, handler http.Handler, mws ...Middleware) {
	p, h, m := g.group(path, handler, mws...)
	g.a.Put(p, h, m...)
}

func (g *ApplicationGroup) Patch(path string, handler http.Handler, mws ...Middleware) {
	p, h, m := g.group(path, handler, mws...)
	g.a.Patch(p, h, m...)
}

func (g *ApplicationGroup) Delete(path string, handler http.Handler, mws ...Middleware) {
	p, h, m := g.group(path, handler, mws...)
	g.a.Delete(p, h, m...)
}

func (g *ApplicationGroup) WithGlobalMiddlewares(mws ...Middleware) WebApplication {
	g.mws = append(g.mws, mws...)
	return g
}

func (g *ApplicationGroup) group(path string, handler http.Handler, mws ...Middleware) (string, http.Handler, []Middleware) {
	return pth.Join(g.path, path), handler, append(g.mws, mws...)
}
