package http

import (
	"bytes"
	"io"
	"io/ioutil"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

// https://github.com/gin-gonic/gin/issues/961#issuecomment-557931409
func RequestLoggerMiddleware(c *gin.Context) {
	var buf bytes.Buffer
	tee := io.TeeReader(c.Request.Body, &buf)
	body, _ := ioutil.ReadAll(tee)
	c.Request.Body = ioutil.NopCloser(&buf)
	log.Debugf("request url: %s", c.Request.RequestURI)
	log.Debugf("request header: %s", c.Request.Header)
	log.Debugf("request body: %s", body)
	c.Next()
}
