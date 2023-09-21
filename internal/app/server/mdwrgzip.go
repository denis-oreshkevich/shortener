package server

import (
	"compress/gzip"
	"github.com/gin-gonic/gin"
	"strings"
)

type gzipWriter struct {
	gin.ResponseWriter
	gzw *gzip.Writer
}

func newGzipWriter(w gin.ResponseWriter, gzw *gzip.Writer) *gzipWriter {
	return &gzipWriter{ResponseWriter: w, gzw: gzw}
}

func (g *gzipWriter) WriteString(s string) (int, error) {
	if shouldCompress(g) {
		g.ResponseWriter.Header().Set("Content-Encoding", "gzip")
		g.ResponseWriter.Header().Del("Content-Length")
		return g.gzw.Write([]byte(s))
	}
	return g.ResponseWriter.WriteString(s)
}

func (g *gzipWriter) Write(data []byte) (int, error) {
	if shouldCompress(g) {
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

func shouldCompress(g *gzipWriter) bool {
	ct := g.Header().Get("Content-Type")
	return strings.Contains(ct, "application/json") || strings.Contains(ct, "text/html")
}
func Gzip(c *gin.Context) {
	ceh := c.GetHeader("Content-Encoding")
	if strings.Contains(ceh, "gzip") {
		reader, err := gzip.NewReader(c.Request.Body)
		defer reader.Close()
		if err != nil {
			c.AbortWithError(500, err)
		}
		c.Request.Body = reader
	}
	aeh := c.GetHeader("Accept-Encoding")
	if strings.Contains(aeh, "gzip") {
		wr := gzip.NewWriter(c.Writer)
		defer wr.Close()
		cw := newGzipWriter(c.Writer, wr)
		c.Writer = cw
	}
	c.Next()
}
