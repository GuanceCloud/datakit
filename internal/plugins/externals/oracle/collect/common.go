// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

// Package collect contains Oracle collect implement.
package collect

import (
	"bytes"
	"context"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"net/url"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/GuanceCloud/cliutils/logger"
	"github.com/GuanceCloud/cliutils/point"
	"golang.org/x/net/context/ctxhttp"
)

type Option struct {
	Interval        string `long:"interval" description:"gather interval" default:"10s"`
	InstanceDesc    string `long:"instance-desc" description:"oracle description"`
	Host            string `long:"host" description:"oracle host"`
	Port            string `long:"port" description:"oracle port" default:"1521"`
	Username        string `long:"username" description:"oracle username"`
	Password        string `long:"password" description:"oracle password"`
	ServiceName     string `long:"service-name" description:"oracle service name"`
	Tags            string `long:"tags" description:"additional tags in 'a=b,c=d,...' format"`
	DatakitHTTPHost string `long:"datakit-http-host" description:"DataKit HTTP server host" default:"localhost"`
	DatakitHTTPPort int    `long:"datakit-http-port" description:"DataKit HTTP server port" default:"9529"`
	Election        bool   `long:"election" description:"whether election of this input is enabled"`

	Log      string   `long:"log" description:"log path"`
	LogLevel string   `long:"log-level" description:"log file" default:"info"`
	Query    []string `long:"query" description:"custom query array"`
}

var (
	opt            *Option
	l              *logger.Logger
	SignaIterrrupt = make(chan os.Signal, 1)

	dic = map[string]string{
		"buffer_cache_hit_ratio":       "buffer_cachehit_ratio",
		"cursor_cache_hit_ratio":       "cursor_cachehit_ratio",
		"library_cache_hit_ratio":      "library_cachehit_ratio",
		"shared_pool_free_%":           "shared_pool_free",
		"physical_read_bytes_per_sec":  "physical_reads",
		"physical_write_bytes_per_sec": "physical_writes",
		"enqueue_timeouts_per_sec":     "enqueue_timeouts",

		"gc_cr_block_received_per_second": "gc_cr_block_received",
		"global_cache_blocks_corrupted":   "cache_blocks_corrupt",
		"global_cache_blocks_lost":        "cache_blocks_lost",
		"average_active_sessions":         "active_sessions",
		"sql_service_response_time":       "service_response_time",
		"user_rollbacks_per_sec":          "user_rollbacks",
		"total_sorts_per_user_call":       "sorts_per_user_call",
		"rows_per_sort":                   "rows_per_sort",
		"disk_sort_per_sec":               "disk_sorts",
		"memory_sorts_ratio":              "memory_sorts_ratio",
		"database_wait_time_ratio":        "database_wait_time_ratio",
		"session_limit_%":                 "session_limit_usage",
		"session_count":                   "session_count",
		"temp_space_used":                 "temp_space_used",
	}
)

func Set(op *Option, log *logger.Logger) {
	opt = op
	l = log
}

////////////////////////////////////////////////////////////////////////////////

const (
	metricNameProcess    = "oracle_process"
	metricNameTablespace = "oracle_tablespace"
	metricNameSystem     = "oracle_system"

	pdbName        = "pdb_name"
	tablespaceName = "tablespace_name"
	programName    = "program"
)

////////////////////////////////////////////////////////////////////////////////

type collectParameters struct {
	m *Monitor
}

// collectOption used to add various options to collect.
type collectOption func(x *collectParameters)

func withMonitor(m *Monitor) collectOption {
	return func(x *collectParameters) {
		x.m = m
	}
}

type dbMetricsCollector interface {
	collect() (*point.Point, error)
}

////////////////////////////////////////////////////////////////////////////////

type buildPointOpt struct {
	tf         *tagField
	metricName string
	m          *Monitor
}

func buildPoint(opt *buildPointOpt) *point.Point {
	l.Debugf("got %d fields from metric %s", len(opt.tf.Fields), opt.metricName)

	opts := point.DefaultMetricOptions()
	opts = append(opts, point.WithTime(opt.tf.TS))

	newTags := MergeTags(opt.m.tags, opt.tf.Tags, opt.m.host)

	return point.NewPointV2([]byte(opt.metricName),
		append(point.NewTags(newTags), point.NewKVs(opt.tf.Fields)...),
		opts...)
}

////////////////////////////////////////////////////////////////////////////////

func selectWrapper[T any](m *Monitor, s T, sql string) error {
	now := time.Now()

	err := m.db.Select(s, sql)
	if err != nil && (strings.Contains(err.Error(), "ORA-01012") || strings.Contains(err.Error(), "database is closed")) {
		if err := m.ConnectDB(); err != nil {
			m.CloseDB()
			return err
		}
	}

	l.Debugf("executed sql: %s, cost: %v\n", sql, time.Since(now))
	return err
}

func writeData(data []byte, urlPath string) error {
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

	body, err := ioutil.ReadAll(resp.Body)
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

// MergeTags merge all optional tags from global tags/inputs config tags and host tags
// from remote URL.
// NOTE: This function needs to synchronize with the same name function in the file
//       internal/plugins/inputs/inputs.go of project Datakit.
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

type safeByteArray struct {
	lock sync.RWMutex
	data [][]byte
}

func newSafeArray() *safeByteArray {
	return &safeByteArray{}
}

func (sa *safeByteArray) add(s string) {
	sa.lock.Lock()
	defer sa.lock.Unlock()

	sa.data = append(sa.data, []byte(s))
}

func (sa *safeByteArray) get() [][]byte {
	sa.lock.RLock()
	defer sa.lock.RUnlock()

	newArr := make([][]byte, len(sa.data))
	copy(newArr, sa.data)

	return newArr
}

func (sa *safeByteArray) len() int {
	sa.lock.RLock()
	defer sa.lock.RUnlock()

	return len(sa.data)
}

////////////////////////////////////////////////////////////////////////////////

type tagField struct {
	Tags   map[string]string      `json:"tags"`
	Fields map[string]interface{} `json:"fields"`
	TS     time.Time
}

func newTagField() *tagField {
	return &tagField{
		Tags:   make(map[string]string),
		Fields: make(map[string]interface{}),
	}
}

func (tf *tagField) setTS(t time.Time) {
	if tf.TS.IsZero() {
		tf.TS = t
	}
}

func (tf *tagField) addTag(key, val string) {
	if _, ok := tf.Tags[key]; !ok {
		tf.Tags[key] = val
	}
}

func (tf *tagField) addField(key string, val interface{}) {
	if _, ok := tf.Fields[key]; !ok {
		alias, ok := dic[key]
		if ok {
			tf.Fields[key] = alias // replace with dic.
		} else {
			tf.Fields[key] = val
		}
	}
}

func (tf *tagField) isEmpty() bool {
	return len(tf.Fields) == 0
}

////////////////////////////////////////////////////////////////////////////////
