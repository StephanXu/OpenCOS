package websvr

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"xxtuitui.com/filesvr/source"
)

type MappingFileReq struct {
	ReqUrl    string            `json:"reqUrl"`
	LocalName string            `json:"localName"`
	Hashes    map[string]string `json:"hashes"`
}

func MappingFile(c *gin.Context) {
	req := MappingFileReq{}
	if err := c.BindJSON(&req); err != nil {
		c.AbortWithError(http.StatusBadRequest, err)
		return
	}
	if err := source.Manager.MappingFile(req.ReqUrl, req.LocalName, req.Hashes); err != nil {
		c.AbortWithError(http.StatusBadRequest, err)
	}
	c.Status(http.StatusOK)
}
