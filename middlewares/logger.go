package middlewares

import (
	"github.com/why9661/goweb"
	"log"
	"time"
)

func Logger() goweb.HandlerFunc {
	return func(c *goweb.Context) {
		t := time.Now()
		c.Next()
		log.Printf("[%d] %s in %v", c.StatusCode, c.Req.RequestURI, time.Since(t))
	}
}
