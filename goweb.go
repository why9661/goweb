package goweb

import (
	"html/template"
	"net/http"
	"path"
	"strings"
)

type HandlerFunc func(*Context)

// Launcher implement the interface of ServeHTTP
type Launcher struct {
	router *router
	*RouterGroup
	groups        []*RouterGroup //store all RouterGroups
	htmlTemplates *template.Template
	funcMap       template.FuncMap
}

type RouterGroup struct {
	prefix      string
	middlewares []HandlerFunc
	launcher    *Launcher // all RouterGroups share one launcher
}

func New() *Launcher {
	launcher := &Launcher{router: newRouter()}
	launcher.RouterGroup = &RouterGroup{launcher: launcher}
	launcher.groups = []*RouterGroup{launcher.RouterGroup}
	return launcher
}

func (l *Launcher) SetFuncMap(funcMap template.FuncMap) {
	l.funcMap = funcMap
}

func (l *Launcher) LoadHTMLGlob(pattern string) {
	l.htmlTemplates = template.Must(template.New("").Funcs(l.funcMap).ParseGlob(pattern))
}

//create a new RouterGroup
func (group *RouterGroup) Group(prefix string) *RouterGroup {
	launcher := group.launcher
	newGroup := &RouterGroup{
		launcher: launcher,
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

// create static handler
func (group *RouterGroup) createStaticHandler(relativePath string, fs http.FileSystem) HandlerFunc {
	absolutePath := path.Join(group.prefix, relativePath)
	fileServer := http.StripPrefix(absolutePath, http.FileServer(fs))
	return func(c *Context) {
		file := c.Param("filepath")
		if _, err := fs.Open(file); err != nil {
			c.Status(http.StatusNotFound)
			return
		}

		fileServer.ServeHTTP(c.Writer, c.Req)
	}
}

// serve static files
func (group *RouterGroup) Static(relativePath string, root string) {
	handler := group.createStaticHandler(relativePath, http.Dir(root))
	urlPattern := path.Join(relativePath, "/*filepath")
	// Register GET handlers
	group.GET(urlPattern, handler)
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
	c.launcher = l
	l.router.handle(c)
}
