package websvr

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"xxtuitui.com/filesvr/source"
)

func GetCacheUrlHandler(c *gin.Context) {
	reqUrl := c.Param("reqUrl")
	sourceName := c.Param("source")
	fmt.Printf("sourceName: %s, reqUrl: %s\n", sourceName, reqUrl)
	s := source.Manager.GetSource(sourceName)
	if s == nil {
		c.Status(http.StatusServiceUnavailable)
		return
	}
	dest, err := (*s).GetUrl(reqUrl)
	if err != nil {
		c.Status(http.StatusNotFound)
		return
	}
	c.Redirect(307, dest)
}

func GetCacheUrlHandlerByDefault(c *gin.Context) {
	reqUrl := "/library" + c.Param("reqUrl")
	sourceName := "aliyunpan"
	fmt.Printf("sourceName: %s, reqUrl: %s\n", sourceName, reqUrl)
	s := source.Manager.GetSource(sourceName)
	if s == nil {
		c.Status(http.StatusServiceUnavailable)
		return
	}
	dest, err := (*s).GetUrl(reqUrl)
	if err != nil {
		c.Status(http.StatusNotFound)
		return
	}
	c.Redirect(307, dest)
}
