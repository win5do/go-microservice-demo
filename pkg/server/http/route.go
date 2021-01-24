package http

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

func Register(mux *gin.Engine) {
	// list all api
	mux.GET("/apis", func(c *gin.Context) {
		list := ""
		for _, v := range mux.Routes() {
			list += fmt.Sprintf("%s %s\n", v.Method, v.Path)
		}
		c.String(http.StatusOK, list)
	})

	mux.GET("/ping", func(c *gin.Context) {
		c.String(200, "pong")
	})
}
