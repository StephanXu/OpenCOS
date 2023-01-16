package websvr

import (
	"bytes"
	"compress/flate"
	"compress/gzip"
	"encoding/json"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"strconv"

	"github.com/Jeffail/gabs/v2"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"xxtuitui.com/filesvr/config"
	"xxtuitui.com/filesvr/source"
)

type (
	FileHashItem struct {
		Filename string            `json:"filename"`
		Hashes   map[string]string `json:"hashes"`
	}
)

var fileHashes map[string]FileHashItem

func refreshFileHashes(filename string) error {
	if fileHashes == nil {
		fileHashes = make(map[string]FileHashItem)
	}
	content, err := os.ReadFile(filename)
	if err != nil {
		return err
	}

	data := []FileHashItem{}
	if err := json.Unmarshal(content, &data); err != nil {
		return err
	}
	for _, item := range data {
		fileHashes[item.Filename] = item
	}
	return nil
}

func extendsLibrary(content []byte) []byte {
	if fileHashes == nil || len(fileHashes) == 0 {
		refreshFileHashes(config.App.LocalHash)
	}
	document, err := gabs.ParseJSON(content)
	if err != nil {
		logrus.WithField("err", err).Error("ParseLibraryJsonFailed")
		return content
	}
	for _, metaData := range document.Path("MediaContainer.Metadata").Children() {
		items, _ := gabs.New().Array()
		for _, media := range metaData.S("Media").Children() {
			if len(media.S("Part").Children()) != 1 {
				continue
			}
			filename, ok := media.Path("Part.0.file").Data().(string)
			if !ok {
				continue
			}
			requrl, ok := media.Path("Part.0.key").Data().(string)
			if !ok {
				continue
			}

			hashes, ok := fileHashes[filename]
			if !ok {
				continue
			}
			if source.Manager.HasMapping(requrl) {
				continue
			}
			err := source.Manager.MappingFile(requrl, filename, hashes.Hashes)
			if err != nil {
				logrus.WithFields(logrus.Fields{
					"reqUrl":   requrl,
					"filename": filename,
					"hashes":   hashes.Hashes,
				}).Error("MappingFileFailed")
			} else {
				logrus.WithFields(logrus.Fields{
					"reqUrl":   requrl,
					"filename": filename,
					"hashes":   hashes.Hashes,
				}).Info("MappingFileSuccessFromRequest")
			}
		}
		for _, item := range items.Children() {
			metaData.ArrayAppend(item, "Media")
		}
	}
	return document.Bytes()
}

func rewriteBody(resp *http.Response) (err error) {
	var bodyReader io.Reader
	switch resp.Header.Get("Content-Encoding") {
	case "gzip":
		bodyReader, _ = gzip.NewReader(resp.Body)
	case "deflate":
		bodyReader = flate.NewReader(resp.Body)
	default:
		bodyReader = resp.Body
	}
	b, err := ioutil.ReadAll(bodyReader)
	if err != nil {
		return err
	}
	if err := resp.Body.Close(); err != nil {
		return err
	}
	b = extendsLibrary(b)

	var compressedBuffer bytes.Buffer
	switch resp.Header.Get("Content-Encoding") {
	case "gzip":
		bodyWriter := gzip.NewWriter(&compressedBuffer)
		bodyWriter.Write(b)
		bodyWriter.Close()
		b = compressedBuffer.Bytes()
	case "deflate":
		bodyWriter, _ := flate.NewWriter(&compressedBuffer, 1)
		bodyWriter.Write(b)
		bodyWriter.Close()
		b = compressedBuffer.Bytes()
	}

	resp.Body = ioutil.NopCloser(bytes.NewReader(b))
	resp.ContentLength = int64(len(b))
	resp.Header.Set("Content-Length", strconv.Itoa(len(b)))
	return nil
}

func proxy(c *gin.Context) {
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
	proxy.ModifyResponse = rewriteBody

	proxy.ServeHTTP(c.Writer, c.Request)
}
