package websvr

import (
	"net/http"
	"net/http/httputil"
	"net/url"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"xxtuitui.com/filesvr/config"
	"xxtuitui.com/filesvr/source"
)

func getCacheUrlHandler(c *gin.Context) {
	reqUrl := c.Param("reqUrl")
	sourceName := c.Param("source")

	s := source.Manager.GetSource(sourceName)
	if s == nil {
		logrus.WithFields(logrus.Fields{
			"sourceName": sourceName,
			"reqUrl":     reqUrl,
		}).Info("SourceNotFound")
		c.Status(http.StatusServiceUnavailable)
		return
	}
	dest, err := s.GetUrl(reqUrl)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"sourceName": sourceName,
			"reqUrl":     reqUrl,
		}).Info("MappingNotFound")
		c.Status(http.StatusNotFound)
		return
	}

	logrus.WithFields(logrus.Fields{
		"sourceName": sourceName,
		"reqUrl":     reqUrl,
		"mappingTo":  dest,
	}).Info("GetMappingUrl")
	c.Redirect(307, dest)
}

func getCacheUrlHandlerByDefault(c *gin.Context) {
	reqUrl := "/library/parts" + c.Param("reqUrl")
	sourceName := config.App.DefaultSource

	s := source.Manager.GetSource(sourceName)
	if s == nil {
		logrus.WithFields(logrus.Fields{
			"sourceName": sourceName,
			"reqUrl":     reqUrl,
		}).Info("SourceNotFound")
		c.Status(http.StatusServiceUnavailable)
		return
	}

	if !s.HasMapping(reqUrl) {
		logrus.WithFields(logrus.Fields{
			"sourceName": sourceName,
			"reqUrl":     reqUrl,
		}).Warn("CacheDegradation")
		remote, err := url.Parse(config.App.PlexHost)
		if err != nil {
			panic(err)
		}

		proxy := httputil.NewSingleHostReverseProxy(remote)
		proxy.Director = func(req *http.Request) {
			req.Header = c.Request.Header
			req.Host = remote.Host
			req.URL.Scheme = remote.Scheme
			req.URL.Host = remote.Host
			req.URL.Path = c.Request.URL.Path
		}
		proxy.ServeHTTP(c.Writer, c.Request)
		return
	}

	dest, err := s.GetUrl(reqUrl)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"sourceName": sourceName,
			"reqUrl":     reqUrl,
		}).Info("MappingNotFound")
		c.Status(http.StatusNotFound)
		return
	}

	logrus.WithFields(logrus.Fields{
		"sourceName": sourceName,
		"reqUrl":     reqUrl,
		"mappingTo":  dest,
	}).Info("GetMappingUrl")
	c.Redirect(307, dest)
}
