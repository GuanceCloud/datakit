// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

// Package ccommon contains common collect code.
package ccommon

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/GuanceCloud/cliutils/logger"
	"github.com/GuanceCloud/cliutils/point"
	"golang.org/x/net/context/ctxhttp"
)

const (
	CategoryMetric  = "metric"
	CategoryEvent   = "keyevent"
	CategoryLogging = "logging"
)

////////////////////////////////////////////////////////////////////////////////

type IInput interface {
	Run()
}

type DBMetricsCollector interface {
	Collect() (*point.Point, error)
}

////////////////////////////////////////////////////////////////////////////////

type Option struct {
	Interval        string `long:"interval" description:"gather interval" default:"10s"`
	InstanceDesc    string `long:"instance-desc" description:"description"`
	Host            string `long:"host" description:"db host"`
	Port            string `long:"port" description:"db port" default:"1521"`
	Username        string `long:"username" description:"db username"`
	Password        string `long:"password" description:"db password"`
	ServiceName     string `long:"service-name" description:"oracle service name"`
	Tags            string `long:"tags" description:"additional tags in 'a=b,c=d,...' format"`
	DatakitHTTPHost string `long:"datakit-http-host" description:"DataKit HTTP server host" default:"localhost"`
	DatakitHTTPPort int    `long:"datakit-http-port" description:"DataKit HTTP server port" default:"9529"`
	Election        bool   `long:"election" description:"whether election of this input is enabled"`
	Inputs          string `long:"inputs" description:"collectors should be enabled" default:"oracle"`
	Database        string `long:"database" description:"database name"`
	SlowQueryTime   string `long:"slow-query-time" description:"Slow query time defined" default:""`

	Log      string   `long:"log" description:"log path"`
	LogLevel string   `long:"log-level" description:"log file" default:"info"`
	Query    []string `long:"query" description:"custom query array"`
}

////////////////////////////////////////////////////////////////////////////////

func GetPostURL(election bool, category, inputName, host string, port int) string {
	var postURL string

	var (
		ignoreGlobalHostTags = "ignore_global_host_tags=true"
		globalEnvTags        = "global_env_tags=true"
	)
	if election {
		postURL = fmt.Sprintf(
			"http://%s/v1/write/%s?input=%s&%s&%s",
			net.JoinHostPort(host, strconv.Itoa(port)),
			category, inputName,
			ignoreGlobalHostTags, globalEnvTags)
	} else {
		postURL = fmt.Sprintf(
			"http://%s/v1/write/%s?input=%s",
			net.JoinHostPort(host, strconv.Itoa(port)),
			category, inputName,
		)
	}

	return postURL
}

func GetLastErrorURL(host string, port int) string {
	return fmt.Sprintf(
		"http://%s/v1/lasterror",
		net.JoinHostPort(host, strconv.Itoa(port)),
	)
}

////////////////////////////////////////////////////////////////////////////////

type BuildPointOpt struct {
	TF         *TagField
	MetricName string
	Tags       map[string]string
	Host       string
}

func BuildPoint(l *logger.Logger, opt *BuildPointOpt) *point.Point {
	var err error
	l.Debugf("got %d fields from metric %s", len(opt.TF.Fields), opt.MetricName)

	opts := point.DefaultMetricOptions()
	opts = append(opts, point.WithTime(opt.TF.TS))

	setHost := false
	host := strings.ToLower(opt.Host)
	switch host {
	case "", "localhost":
		setHost = true
	default:
		if net.ParseIP(host).IsLoopback() {
			setHost = true
		}
	}
	if setHost {
		host, err = os.Hostname()
		if err != nil {
			l.Errorf("os.Hostname failed: %v", err)
		}
	}

	newTags := MergeTags(opt.Tags, opt.TF.Tags, host)

	return point.NewPointV2([]byte(opt.MetricName),
		append(point.NewTags(newTags), point.NewKVs(opt.TF.Fields)...),
		opts...)
}

func BuildPointLogging(l *logger.Logger, opt *BuildPointOpt) *point.Point {
	var err error
	l.Debugf("got %d fields from logging %s", len(opt.TF.Fields), opt.MetricName)

	opts := point.DefaultLoggingOptions()
	opts = append(opts, point.WithTime(opt.TF.TS))

	setHost := false
	host := strings.ToLower(opt.Host)
	switch host {
	case "", "localhost":
		setHost = true
	default:
		if net.ParseIP(host).IsLoopback() {
			setHost = true
		}
	}
	if setHost {
		host, err = os.Hostname()
		if err != nil {
			l.Errorf("os.Hostname failed: %v", err)
		}
	}

	newTags := MergeTags(opt.Tags, opt.TF.Tags, host)

	return point.NewPointV2([]byte(opt.MetricName),
		append(point.NewTags(newTags), point.NewKVs(opt.TF.Fields)...),
		opts...)
}

////////////////////////////////////////////////////////////////////////////////

// MergeTags merge all optional tags from global tags/inputs config tags and host tags
// from remote URL.
// NOTE: This function needs to synchronize with the same name function in the file
//
//	internal/plugins/inputs/inputs.go of project Datakit.
func MergeTags(global, origin map[string]string, remote string) map[string]string {
	out := map[string]string{}

	for k, v := range origin {
		out[k] = v
	}

	host := remote
	if remote == "" {
		goto end
	}

	// if 'host' exist in origin tags, ignore 'host' tag within remote
	if _, ok := origin["host"]; ok {
		goto end
	}

	// try get 'host' tag from remote URL.
	if u, err := url.Parse(remote); err == nil && u.Host != "" { // like scheme://host:[port]/...
		host = u.Host
		if ip, _, err := net.SplitHostPort(u.Host); err == nil {
			host = ip
		}
	} else { // not URL, only IP:Port
		if ip, _, err := net.SplitHostPort(remote); err == nil {
			host = ip
		}
	}

	if host != "localhost" && !net.ParseIP(host).IsLoopback() {
		out["host"] = host
	}

end: // global tags(host/election tags) got the lowest priority.
	for k, v := range global {
		if _, ok := out[k]; !ok {
			out[k] = v
		}
	}

	return out
}

////////////////////////////////////////////////////////////////////////////////

func WriteData(l *logger.Logger, data []byte, urlPath string) error {
	// dataway path
	ctx, ctxCancel := context.WithCancel(context.Background())
	defer ctxCancel()
	httpReq, err := http.NewRequest("POST", urlPath, bytes.NewBuffer(data))
	if err != nil {
		l.Errorf(err.Error())
		return err
	}

	httpReq = httpReq.WithContext(ctx)
	tmctx, timeoutCancel := context.WithTimeout(context.Background(), time.Second*10)
	defer timeoutCancel()

	resp, err := ctxhttp.Do(tmctx, http.DefaultClient, httpReq)
	if err != nil {
		l.Errorf(err.Error())
		return err
	}
	defer resp.Body.Close() //nolint:errcheck

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		l.Error(err)
		return err
	}

	switch resp.StatusCode / 100 {
	case 2:
		l.Debugf("post to %s ok", urlPath)
		return nil
	default:
		l.Errorf("post to %s failed(HTTP: %d): %s", urlPath, resp.StatusCode, string(body))
		return fmt.Errorf("post datakit failed")
	}
}

////////////////////////////////////////////////////////////////////////////////

type TagField struct {
	Tags   map[string]string      `json:"tags"`
	Fields map[string]interface{} `json:"fields"`
	TS     time.Time
}

func NewTagField() *TagField {
	return &TagField{
		Tags:   make(map[string]string),
		Fields: make(map[string]interface{}),
	}
}

func (tf *TagField) SetTS(t time.Time) {
	if tf.TS.IsZero() {
		tf.TS = t
	}
}

func (tf *TagField) AddTag(key, val string) {
	if _, ok := tf.Tags[key]; !ok {
		tf.Tags[key] = val
	}
}

func (tf *TagField) AddField(key string, val interface{}, dic map[string]string) {
	if _, ok := tf.Fields[key]; !ok {
		tf.Fields[key] = val

		if dic != nil {
			alias, ok := dic[key]
			if ok {
				tf.Fields[key] = alias // replace with dic.
			}
		}
	}
}

func (tf *TagField) IsEmpty() bool {
	return len(tf.Fields) == 0
}

////////////////////////////////////////////////////////////////////////////////

type ByteArray struct {
	Data [][]byte
}

func NewByteArray() *ByteArray {
	return &ByteArray{}
}

func (ba *ByteArray) Add(s string) {
	ba.Data = append(ba.Data, []byte(s))
}

func (ba *ByteArray) Get() [][]byte {
	newArr := make([][]byte, len(ba.Data))
	copy(newArr, ba.Data)

	return newArr
}

func (ba *ByteArray) Len() int {
	return len(ba.Data)
}

////////////////////////////////////////////////////////////////////////////////

type SafeByteArray struct {
	Data   [][]byte
	locker sync.RWMutex
}

func NewSafeArray() *SafeByteArray {
	return &SafeByteArray{}
}

func (sa *SafeByteArray) Add(s string) {
	sa.locker.Lock()
	defer sa.locker.Unlock()

	sa.Data = append(sa.Data, []byte(s))
}

func (sa *SafeByteArray) Get() [][]byte {
	sa.locker.RLock()
	defer sa.locker.RUnlock()

	newArr := make([][]byte, len(sa.Data))
	copy(newArr, sa.Data)

	return newArr
}

func (sa *SafeByteArray) Len() int {
	sa.locker.RLock()
	defer sa.locker.RUnlock()

	return len(sa.Data)
}

////////////////////////////////////////////////////////////////////////////////
