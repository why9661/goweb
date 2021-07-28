package ggin

import (
	"github.com/why9661/ggin/middlewares"
	"net/http"
	"strings"
)

type HandlerFunc func(*Context)

// Engine implement the interface of ServeHTTP
type Engine struct {
	router *router
	*RouterGroup
	groups []*RouterGroup //store all RouterGroups
}

func New() *Engine {
	engine := &Engine{router: newRouter()}
	engine.RouterGroup = &RouterGroup{engine: engine}
	engine.groups = []*RouterGroup{engine.RouterGroup}
	return engine
}

func Default() *Engine {
	engine := New()
	engine.Use(middlewares.Logger())
	engine.Use(middlewares.Recovery())
	return engine
}

func (l *Engine) addRoute(method string, pattern string, handler HandlerFunc) {
	l.router.addRoute(method, pattern, handler)
}

// GET defines the method to add GET request
func (l *Engine) GET(pattern string, handler HandlerFunc) {
	l.addRoute("GET", pattern, handler)
}

// POST defines the method to add POST request
func (l *Engine) POST(pattern string, handler HandlerFunc) {
	l.addRoute("POST", pattern, handler)
}

// Run defines the method to start a http server
func (l *Engine) Run(addr string) (err error) {
	return http.ListenAndServe(addr, l)
}

func (l *Engine) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	var middlewares []HandlerFunc
	for _, group := range l.groups {
		if strings.HasPrefix(req.URL.Path, group.prefix) {
			middlewares = append(middlewares, group.middlewares...)
		}
	}
	c := newContext(w, req)
	c.handlers = middlewares
	c.engine = l
	l.router.handle(c)
}
