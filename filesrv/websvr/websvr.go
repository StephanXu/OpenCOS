package websvr

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

func logMiddleWare() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()
		method := c.Request.Method
		reqUrl := c.Request.RequestURI
		statusCode := c.Writer.Status()
		clientIP := c.ClientIP()
		logrus.WithFields(logrus.Fields{
			"status_code": statusCode,
			"client_ip":   clientIP,
			"req_method":  method,
			"req_uri":     reqUrl,
		}).Info()
	}
}

func Run() {
	r := gin.Default()
	r.Use(logMiddleWare())
	r.GET("/cache/:source/*reqUrl", GetCacheUrlHandler)
	r.GET("/library/parts/*reqUrl", GetCacheUrlHandlerByDefault)
	r.POST("/cache/mapping", MappingFile)
	r.GET("/library/sections/:id/*proxyPath", proxy)
	if err := r.Run(":3000"); err != nil {
		fmt.Printf("startup service failed, err: %v\n", err)
		return
	}
}
