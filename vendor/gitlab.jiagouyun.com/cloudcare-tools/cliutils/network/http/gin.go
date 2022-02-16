package http

import (
	"bytes"
	"compress/gzip"
	"crypto/md5"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"gitlab.jiagouyun.com/cloudcare-tools/cliutils"
	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/logger"
)

const (
	XAgentIP       = "X-Agent-Ip"
	XAgentUID      = "X-Agent-Uid"
	XCQRP          = "X-CQ-RP"
	XDatakitInfo   = "X-Datakit-Info"
	XDatakitUUID   = "X-Datakit-UUID" // deprecated
	XDBUUID        = "X-DB-UUID"
	XDomainName    = "X-Domain-Name"
	XLua           = "X-Lua"
	XPrecision     = "X-Precision"
	XRP            = "X-RP"
	XSource        = "X-Source"
	XTableName     = "X-Table-Name"
	XToken         = "X-Token"
	XTraceId       = "X-Trace-Id"
	XVersion       = "X-Version"
	XWorkspaceUUID = "X-Workspace-UUID"
)

var (
	allowHeaders = strings.Join(
		[]string{
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
			XToken,
			XDatakitUUID,
			XRP,
			XPrecision,
			XLua,
		}, ", ")
	realIPHeader      = []string{"X-Forwarded-For", "X-Real-IP", "RemoteAddr"}
	MaxRequestBodyLen = 128

	l = logger.DefaultSLogger("gin")
)

func Init() {
	l = logger.SLogger("gin")

	if v, ok := os.LookupEnv("MAX_REQUEST_BODY_LEN"); ok {
		if i, err := strconv.ParseInt(v, 10, 64); err != nil {
			l.Warnf("invalid MAX_REQUEST_BODY_LEN, expect int, got %s, ignored", v)
		} else {
			MaxRequestBodyLen = int(i)
		}
	}
}

func GinLogFormmatter(param gin.LogFormatterParams) string {
	realIP := param.ClientIP
	for _, h := range realIPHeader {
		if v := param.Request.Header.Get(h); v != "" {
			realIP = v
		}
	}

	if param.ErrorMessage != "" {
		return fmt.Sprintf("[GIN] %v | %3d | %8v | %15s | %-7s %#v -> %s\n",
			param.TimeStamp.Format("2006/01/02 - 15:04:05"),
			param.StatusCode,
			param.Latency,
			net.ParseIP(realIP),
			param.Method,
			param.Path,
			param.ErrorMessage)
	} else {
		return fmt.Sprintf("[GIN] %v | %3d | %8v | %15s | %-7s %#v\n",
			param.TimeStamp.Format("2006/01/02 - 15:04:05"),
			param.StatusCode,
			param.Latency,
			net.ParseIP(realIP),
			param.Method,
			param.Path)
	}
}

func CORSMiddleware(c *gin.Context) {
	allowOrigin := c.GetHeader("origin")
	if allowOrigin == "" {
		allowOrigin = "*"
	}

	c.Writer.Header().Set("Access-Control-Allow-Origin", allowOrigin)
	c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
	c.Writer.Header().Set("Access-Control-Allow-Headers", allowHeaders)
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
		tid := c.Request.Header.Get(XTraceId)
		if tid == "" {
			tid = cliutils.XID(`trace_`)
			c.Request.Header.Set(XTraceId, tid)
		}

		c.Writer.Header().Set(XTraceId, tid)
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

	body := w.body.String()

	l.Infof("%s %s %d, RemoteAddr: %s, Request: [%s], Body: %s",
		c.Request.Method,
		c.Request.URL,
		c.Writer.Status(),
		c.Request.RemoteAddr,
		FormatRequest(c.Request),
		body[:len(body)%MaxRequestBodyLen]+"...")
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
