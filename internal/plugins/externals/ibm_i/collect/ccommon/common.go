// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

// Package ccommon contains the external collector protocol helpers.
package ccommon

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"strconv"
	"time"

	"github.com/GuanceCloud/cliutils/logger"
	"golang.org/x/net/context/ctxhttp"
)

const (
	CategoryMetric  = "metric"
	CategoryLogging = "logging"
)

type Option struct {
	DSN                  string   `long:"dsn" description:"raw IBM i Access ODBC connection string"`
	Host                 string   `long:"host" description:"IBM i system host or tunneled endpoint"`
	Username             string   `long:"username" description:"IBM i username"`
	Password             string   `long:"password" description:"IBM i password"`
	Driver               string   `long:"driver" description:"IBM i Access ODBC driver name" default:"IBM i Access ODBC Driver 64-bit"`
	Interval             string   `long:"interval" description:"metric interval" default:"60s"`
	QueryTimeout         string   `long:"query-timeout" description:"default query timeout" default:"30s"`
	JobQueryTimeout      string   `long:"job-query-timeout" description:"job query timeout" default:"240s"`
	SystemMQQueryTimeout string   `long:"system-mq-query-timeout" description:"message queue query timeout" default:"80s"`
	Queries              []string `long:"query" description:"IBM i query name to collect"`
	SeverityThreshold    int      `long:"severity-threshold" description:"minimum severity counted as critical" default:"50"`
	MessageQueues        []string `long:"message-queue" description:"message queue name filter for message_queue_info"`
	Tags                 string   `long:"tags" description:"additional tags in 'a=b;c=d' format"`
	DatakitHTTPHost      string   `long:"datakit-http-host" description:"DataKit HTTP server host" default:"localhost"`
	DatakitHTTPPort      int      `long:"datakit-http-port" description:"DataKit HTTP server port" default:"9529"`
	Election             bool     `long:"election" description:"whether election is enabled"`
	MetricEnabled        string   `long:"metric-enabled" description:"enable metrics" default:"true"`
	Log                  string   `long:"log" description:"collector log path"`
	LogLevel             string   `long:"log-level" description:"collector log level" default:"info"`
}

var DatakitLastErrURL string

func GetPostURL(election bool, category, inputName, host string, port int) string {
	target := fmt.Sprintf("http://%s/v1/write/%s?input=%s",
		net.JoinHostPort(host, strconv.Itoa(port)), category, inputName)
	if election {
		target += "&ignore_global_host_tags=true&global_env_tags=true"
	}
	return target
}

func GetLastErrorURL(host string, port int) string {
	return fmt.Sprintf("http://%s/v1/lasterror", net.JoinHostPort(host, strconv.Itoa(port)))
}

func WriteData(l *logger.Logger, data []byte, urlPath string) error {
	req, err := http.NewRequest("POST", urlPath, bytes.NewBuffer(data))
	if err != nil {
		return err
	}

	timeoutCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	resp, err := ctxhttp.Do(timeoutCtx, http.DefaultClient, req)
	if err != nil {
		return err
	}
	defer resp.Body.Close() //nolint:errcheck

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	if resp.StatusCode/100 != 2 {
		l.Errorf("post to %s failed(HTTP: %d): %s", urlPath, resp.StatusCode, string(body))
		return fmt.Errorf("post datakit failed with HTTP %d", resp.StatusCode)
	}
	return nil
}

type externalLastErr struct {
	Input      string `json:"input"`
	Source     string `json:"source"`
	ErrContent string `json:"err_content"`
}

func FeedLastError(inputName string, l *logger.Logger, errString string) error {
	data, err := json.Marshal(&externalLastErr{
		Input: inputName, Source: inputName, ErrContent: errString,
	})
	if err != nil {
		return err
	}
	return WriteData(l, data, DatakitLastErrURL)
}

func ReportErrorf(inputName string, l *logger.Logger, format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	l.Error(msg)
	if DatakitLastErrURL == "" {
		return
	}
	go func() {
		if err := FeedLastError(inputName, l, msg); err != nil {
			l.Errorf("FeedLastError failed: %v", err)
		}
	}()
}
