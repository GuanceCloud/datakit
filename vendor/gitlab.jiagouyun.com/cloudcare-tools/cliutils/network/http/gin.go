package http

import (
	"bytes"
	"compress/gzip"
	"crypto/md5"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"gitlab.jiagouyun.com/cloudcare-tools/cliutils"
)

func CORSMiddleware(c *gin.Context) {
	allowHeaders := []string{
		"Content-Type",
		"Content-Length",
		"Accept-Encoding",
		"X-CSRF-Token",
		"Authorization",
		"accept",
		"origin",
		"Cache-Control",
		"X-Requested-With",

		// dataflux headers
		"X-Token",
		"X-Datakit-UUID",
		"X-RP",
		"X-Precision",
		"X-Lua",
	}

	c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
	c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
	c.Writer.Header().Set("Access-Control-Allow-Headers", strings.Join(allowHeaders, ", "))
	c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT")

	if c.Request.Method == "OPTIONS" {
		c.AbortWithStatus(http.StatusNoContent)
		return
	}

	c.Next()
}

func TraceIDMiddleware(c *gin.Context) {
	if c.Request.Method == `OPTIONS` {
		c.Next()
	} else {
		tid := c.Request.Header.Get("X-Trace-ID")
		if tid == "" {
			tid = cliutils.XID(`trace_`)
			c.Request.Header.Set("X-Trace-ID", tid)
		}

		c.Writer.Header().Set("X-Trace-ID", tid)
		c.Next()
	}
}

func FormatRequest(r *http.Request) string {

	// Add the request string
	url := fmt.Sprintf("%v %v %v", r.Method, r.URL, r.Proto)
	request := []string{url}

	// Add the host
	request = append(request, fmt.Sprintf("Host: %v", r.Host))
	// Loop through headers

	for name, headers := range r.Header {
		for _, h := range headers {
			request = append(request, fmt.Sprintf("%v: %v", name, h))
		}
	}

	// Return the request as a string
	return strings.Join(request, "|")
}

type bodyLoggerWriter struct {
	gin.ResponseWriter
	body *bytes.Buffer
}

func (w bodyLoggerWriter) Write(b []byte) (int, error) {
	w.body.Write(b)
	return w.ResponseWriter.Write(b)
}

func RequestLoggerMiddleware(c *gin.Context) {

	w := &bodyLoggerWriter{
		ResponseWriter: c.Writer,
		body:           bytes.NewBufferString(``),
	}

	c.Writer = w
	c.Next()

	code := c.Writer.Status()
	switch code / 200 {
	case 1:
		tid := c.Writer.Header().Get("X-Trace-ID")
		log.Printf("[debug][%s] %s %s %d", tid, c.Request.Method, c.Request.URL, code)

	default:

		log.Printf("[warn] %s %s %d, RemoteAddr: %s, Request: [%s], Body: %s",
			c.Request.Method, c.Request.URL, code, c.Request.RemoteAddr, FormatRequest(c.Request), w.body.String())
	}
}

func GinReadWithMD5(c *gin.Context) (buf []byte, md5str string, err error) {
	buf, err = readBody(c)
	if err != nil {
		return
	}

	md5str = fmt.Sprintf("%x", md5.Sum(buf))

	if c.Request.Header.Get("Content-Encoding") == "gzip" {
		buf, err = Unzip(buf)
	}
	return
}

func GinRead(c *gin.Context) (buf []byte, err error) {
	buf, err = readBody(c)
	if err != nil {
		return
	}

	if c.Request.Header.Get("Content-Encoding") == "gzip" {
		buf, err = Unzip(buf)
	}
	return
}

func GinGetArg(c *gin.Context, hdr, param string) (v string, err error) {
	v = c.Request.Header.Get(hdr)
	if v == "" {
		v = c.Query(param)
		if v == "" {
			err = fmt.Errorf("HTTP header %s and query param %s missing", hdr, param)
		}
	}
	return
}

func Unzip(in []byte) (out []byte, err error) {
	gzr, err := gzip.NewReader(bytes.NewBuffer(in))
	if err != nil {
		return
	}

	out, err = ioutil.ReadAll(gzr)
	if err != nil {
		return
	}
	gzr.Close()
	return
}

func readBody(c *gin.Context) ([]byte, error) {
	body, err := ioutil.ReadAll(c.Request.Body)
	if err != nil {
		return nil, err
	}

	defer c.Request.Body.Close()
	return body, nil
}
