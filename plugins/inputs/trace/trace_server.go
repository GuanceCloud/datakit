package trace

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
	"runtime/debug"
	"log"

	"github.com/gin-gonic/gin"
	"github.com/influxdata/telegraf/metric"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/config"
	"gitlab.jiagouyun.com/cloudcare-tools/ftagent/utils"
)

type TraceDecoder interface {
	Decode(octets []byte) error
}

type TraceReqInfo struct {
	Source      string
	Version     string
	ContentType string
}

type ZipkinTracer struct {
	TraceReqInfo
}

type JaegerTracer struct {
	TraceReqInfo
}

type TraceAdapter struct {
	source string

	duration    int64
	timestampUs int64
	content     string

	class         string
	serviceName   string
	operationName string
	parentID      string
	traceID       string
	spanID        string
	isError       string
}

const US_PER_SECOND int64= 1000000

func (tAdpt *TraceAdapter) mkLineProto() {
	tags   := make(map[string]string)
	fields := make(map[string]interface{})

	tags["__class"]         = tAdpt.class
	tags["__operationName"] = tAdpt.operationName
	tags["__serviceName"]   = tAdpt.serviceName
	if tAdpt.parentID != "" {
		tags["__parentID"]      = tAdpt.parentID
	}

	tags["__traceID"]       = tAdpt.traceID
	tags["__spanID"]        = tAdpt.spanID
	if tAdpt.isError == "true" {
		tags["__isError"] = "true"
	} else {
		tags["__isError"] = "false"
	}

	fields["__duration"]    = tAdpt.duration
	fields["__content"]     = tAdpt.content

	ts := time.Unix(tAdpt.timestampUs/US_PER_SECOND, (tAdpt.timestampUs%US_PER_SECOND)*1000)
	pointMetric, err := metric.New(tAdpt.source, tags, fields, ts)
	if err != nil {
		log.Printf("W! [trace] build metric %s", err)
		return
	}
	acc.AddMetric(pointMetric)
}

func (t *TraceReqInfo) Decode(octets []byte) error {
	var decoder TraceDecoder
	source := strings.ToLower(t.Source)

	switch source {
	case "zipkin":
		decoder = &ZipkinTracer{*t}
	case "jaeger":
		decoder = &JaegerTracer{*t}
	default:
		return fmt.Errorf("Unsupported trace source %s", t.Source)
	}

	return decoder.Decode(octets)
}

func (t *Trace) Serve() {
	initLog()

	path := strings.TrimSpace(t.Path)
	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}

	router := gin.Default()
	router.POST(path, writeTracing)
	err := router.Run(t.Host)
	if err != nil {
		log.Printf("W! [trace] start server %s", err)
	}
}

func initLog() {
	gin.DisableConsoleColor()

	ginLogDir := filepath.Join(config.InstallDir, "data", "trace")
	if _, err := os.Stat(ginLogDir); err != nil {
		if os.IsNotExist(err) {
			err := os.MkdirAll(ginLogDir, os.ModePerm)
			if err != nil {
				log.Printf("W! [trace] create gin log dir %s", err)
				return
			}
		}
	}

	ginLogFile := filepath.Join(ginLogDir, "gin.log")
	if f, err := os.Create(ginLogFile); err != nil {
		return
	} else {
		gin.DefaultWriter = io.MultiWriter(f)
	}
}

func writeTracing(c *gin.Context) {
	defer func(){
		if r := recover(); r != nil {
			log.Printf("W! [trace] %v", r)
			log.Printf("W! [trace] %s", string(debug.Stack()))
		}
	}()

	if err := handleTrace(c); err != nil {
		log.Printf("W! [trace] %v", err)
	}
}
func handleTrace(c *gin.Context) error {
	source := c.Query("source")
	version := c.Query("version")
	contentType := c.Request.Header.Get("Content-Type")
	contentEncoding := c.Request.Header.Get("Content-Encoding")

	body, err := ioutil.ReadAll(c.Request.Body)
	if err != nil {
		JsonReply(c, http.StatusBadRequest, "Read body err: %s", err)
		return err
	}
	defer c.Request.Body.Close()

	if contentEncoding == "gzip" {
		body, err = utils.ReadCompressed(bytes.NewReader(body), true)
		if err != nil {
			JsonReply(c, http.StatusBadRequest, "Uncompress body err: %s", err)
			return err
		}
	}
	
	tInfo := TraceReqInfo{source, version, contentType,}
	err = tInfo.Decode(body)
	if err != nil {
		JsonReply(c, http.StatusBadRequest, "Parse trace err: %s", err)
		return err
	}

	JsonReply(c, http.StatusOK, "ok")
	return nil
}

func JsonReply(c *gin.Context, code int, strfmt string, args ...interface{}) {
	msg := fmt.Sprintf(strfmt, args...)
	c.JSON(code, gin.H{
		"code":    code,
		"message": msg,
	})
}