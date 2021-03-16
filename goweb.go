package goweb

import (
	"net/http"
	"strings"
)

// HandlerFunc defines the request handler used by gee
type HandlerFunc func(*Context)

// Engine implement the interface of ServeHTTP
type Launcher struct {
	router *router
	*RouterGroup
	groups []*RouterGroup //store all RouterGroups
}

type RouterGroup struct {
	prefix      string
	middlewares []HandlerFunc
	parant      *RouterGroup
	launcher    *Launcher // all RouterGroups share one launcher
}

func New() *Launcher {
	launcher := &Launcher{router: newRouter()}
	launcher.RouterGroup = &RouterGroup{launcher: launcher}
	launcher.groups = []*RouterGroup{launcher.RouterGroup}
	return launcher
}

//create a new RouterGroup
func (group *RouterGroup) Group(prefix string) *RouterGroup {
	launcher := group.launcher
	newGroup := &RouterGroup{
		launcher: launcher,
		parant:   group,
		prefix:   group.prefix + prefix,
	}
	launcher.groups = append(launcher.groups, newGroup)
	return newGroup
}

func (group *RouterGroup) Use(middlewares ...HandlerFunc) {
	group.middlewares = append(group.middlewares, middlewares...)
}

func (group *RouterGroup) addRoute(method string, part string, handler HandlerFunc) {
	pattern := group.prefix + part
	group.launcher.router.addRoute(method, pattern, handler)
}

func (group *RouterGroup) GET(pattern string, handler HandlerFunc) {
	group.addRoute("GET", pattern, handler)
}

func (group *RouterGroup) POST(pattern string, handler HandlerFunc) {
	group.addRoute("POST", pattern, handler)
}

func (l *Launcher) addRoute(method string, pattern string, handler HandlerFunc) {
	l.router.addRoute(method, pattern, handler)
}

// GET defines the method to add GET request
func (l *Launcher) GET(pattern string, handler HandlerFunc) {
	l.addRoute("GET", pattern, handler)
}

// POST defines the method to add POST request
func (l *Launcher) POST(pattern string, handler HandlerFunc) {
	l.addRoute("POST", pattern, handler)
}

// Run defines the method to start a http server
func (l *Launcher) Run(addr string) (err error) {
	return http.ListenAndServe(addr, l)
}

func (l *Launcher) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	var middlewares []HandlerFunc
	for _, group := range l.groups {
		if strings.HasPrefix(req.URL.Path, group.prefix) {
			middlewares = append(middlewares, group.middlewares...)
		}
	}
	c := newContext(w, req)
	c.handlers = middlewares
	l.router.handle(c)
}
