package ggin

import (
	"encoding/json"
	"fmt"
	"net/http"
)

type H map[string]interface{}

type Context struct {
	Writer     http.ResponseWriter
	Req        *http.Request
	StatusCode int

	Path   string
	Method string

	Params map[string]string

	handlers []HandlerFunc
	index    int

	engine *Engine
}

func newContext(w http.ResponseWriter, req *http.Request) *Context {
	return &Context{
		Writer: w,
		Req:    req,
		Path:   req.URL.Path,
		Method: req.Method,
		index:  -1,
	}
}

func (c *Context) Param(key string) string {
	value, _ := c.Params[key]
	return value
}

func (c *Context) Next() {
	c.index++
	s := len(c.handlers)
	for ; c.index < s; c.index++ {
		c.handlers[c.index](c)
	}
}

//PostForm returns the first value for the named component of the query.
func (c *Context) PostForm(key string) string {
	return c.Req.FormValue(key)
}

//Query gets the first value associated with the given key.
//If there are no values associated with the key, Get returns the empty string.
func (c *Context) Query(key string) string {
	return c.Req.URL.Query().Get(key)
}

//Status sends an HTTP response header with the provided status code.
func (c *Context) Status(code int) {
	c.StatusCode = code
	c.Writer.WriteHeader(code)
}

//SetHeader sets the header entries associated with key to the single element value.
func (c *Context) SetHeader(key string, value string) {
	c.Writer.Header().Set(key, value)
}

//String writes the given string into the response body and sets the "Content-Type" as "text/plain".
func (c *Context) String(code int, format string, values ...interface{}) {
	c.SetHeader("Content-Type", "text/plain")
	c.Status(code)
	c.Writer.Write([]byte(fmt.Sprintf(format, values...)))
}

//JSON serializes the given struct as JSON into the response body.
//It also sets the Content-Type as "application/json".
func (c *Context) JSON(code int, obj interface{}) {
	c.SetHeader("Content-Type", "application/json")
	c.Status(code)
	encoder := json.NewEncoder(c.Writer)
	if err := encoder.Encode(obj); err != nil {
		http.Error(c.Writer, err.Error(), 500)
	}
}

func (c *Context) Data(code int, data []byte) {
	c.Status(code)
	c.Writer.Write(data)
}

func (c *Context) Fail(code int, err string) {
	c.index = len(c.handlers)
	c.JSON(code, H{"message": err})
}

func (c *Context) HTML(code int, html string) {
	c.SetHeader("Content-Type", "text/html")
	c.Status(code)
	c.Writer.Write([]byte(html))
}
