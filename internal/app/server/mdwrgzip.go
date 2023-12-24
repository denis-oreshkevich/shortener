package server

import (
	"compress/gzip"
	"strings"

	"github.com/gin-gonic/gin"
)

type gzipWriter struct {
	gin.ResponseWriter
	gzw *gzip.Writer
}

func newGzipWriter(w gin.ResponseWriter) *gzipWriter {
	return &gzipWriter{ResponseWriter: w}
}

func (g *gzipWriter) WriteString(s string) (int, error) {
	if shouldCompress(g) {
		g.initGzip()
		g.ResponseWriter.Header().Set("Content-Encoding", "gzip")
		g.ResponseWriter.Header().Del("Content-Length")
		return g.gzw.Write([]byte(s))
	}
	return g.ResponseWriter.WriteString(s)
}

func (g *gzipWriter) initGzip() {
	if g.gzw == nil {
		g.gzw = gzip.NewWriter(g.ResponseWriter)
	}
}

func (g *gzipWriter) Write(data []byte) (int, error) {
	if shouldCompress(g) {
		g.initGzip()
		g.ResponseWriter.Header().Set("Content-Encoding", "gzip")
		g.ResponseWriter.Header().Del("Content-Length")
		return g.gzw.Write(data)
	}
	return g.ResponseWriter.Write(data)
}

func (g *gzipWriter) WriteHeader(code int) {
	g.Header().Del("Content-Length")
	g.ResponseWriter.WriteHeader(code)
}

func shouldCompress(w gin.ResponseWriter) bool {
	ct := w.Header().Get("Content-Type")
	return strings.Contains(ct, "application/json") || strings.Contains(ct, "text/html")
}
func Gzip(c *gin.Context) {
	ceh := c.GetHeader("Content-Encoding")
	if strings.Contains(ceh, "gzip") {
		reader, err := gzip.NewReader(c.Request.Body)
		if err != nil {
			c.AbortWithError(500, err)
			return
		}
		defer reader.Close()
		c.Request.Body = reader
	}
	aeh := c.GetHeader("Accept-Encoding")
	var cw *gzipWriter
	if strings.Contains(aeh, "gzip") {
		cw = newGzipWriter(c.Writer)
		c.Writer = cw

	}
	c.Next()
	if cw != nil && cw.gzw != nil {
		cw.gzw.Close()
	}
}
