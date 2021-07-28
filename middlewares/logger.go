package middlewares

import (
	"github.com/why9661/ggin"
	"log"
	"time"
)

func Logger() ggin.HandlerFunc {
	return func(c *ggin.Context) {
		t := time.Now()
		c.Next()
		log.Printf("[%d] %s in %v", c.StatusCode, c.Req.RequestURI, time.Since(t))
	}
}
