// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

//go:build linux
// +build linux

package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"net/url"
	"os"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/logger"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/tailer"
)

const name = "logfwd"

var (
	argJSONConfig = flag.String("json-config", "", "logfwd json-config")
	argConfig     = flag.String("config", "", "logfwd config file")

	podName                  = os.Getenv("LOGFWD_POD_NAME")
	podNamespace             = os.Getenv("LOGFWD_POD_NAMESPACE")
	wsHost                   = os.Getenv("LOGFWD_DATAKIT_HOST")
	wsPort                   = os.Getenv("LOGFWD_DATAKIT_PORT")
	envMainJSONConfig        = os.Getenv("LOGFWD_JSON_CONFIG")
	envAnnotationDataKitLogs = os.Getenv("LOGFWD_ANNOTATION_DATAKIT_LOGS")

	l = logger.DefaultSLogger(name)
)

func main() {
	quitChannel := make(chan struct{})
	flag.Parse()

	optflags := (logger.OPT_DEFAULT | logger.OPT_STDOUT)
	if err := logger.InitRoot(&logger.Option{Level: logger.INFO, Flags: optflags}); err != nil {
		l.Warnf("failed to set root log, err: %w", err)
	}

	cfg, err := getConfig()
	if err != nil {
		l.Error(err)
		l.Info("exit")
		os.Exit(0)
	}

	startLog(cfg, quitChannel)

	<-quitChannel
}

func getConfig() (*config, error) {
	cfg := func() string {
		if envMainJSONConfig != "" {
			return envMainJSONConfig
		}

		if *argJSONConfig != "" {
			return *argJSONConfig
		}

		if *argConfig != "" {
			b, err := ioutil.ReadFile(*argConfig)
			if err != nil {
				l.Errorf("failed to read fwdconfig file, err: %w", err)
				return ""
			}
			return string(b)
		}

		return ""
	}()

	if cfg == "" {
		return nil, fmt.Errorf("not found fwd config")
	}

	return parseConfig(cfg)
}

func startLog(cfg *config, stop <-chan struct{}) {
	u := url.URL{Scheme: "ws", Host: cfg.DataKitAddr, Path: "/logfwd"}

	var wg sync.WaitGroup

	for _, c := range cfg.Loggings {
		wg.Add(1)

		go func(lg *logging) {
			defer wg.Done()

			wscli := newWsclient(&u)
			wscli.tryConnectWebsocketSrv()
			go wscli.start()

			defer func() {
				if err := wscli.close(); err != nil {
					l.Errorf("failed to close websocket client, err: %w", err)
				}
			}()

			startTailing(lg, forwardFunc(lg, wscli.writeMessage), stop)
		}(c)
	}

	wg.Wait()
}

type writeMessage func([]byte) error

func forwardFunc(lg *logging, fn writeMessage) tailer.ForwardFunc {
	return func(filename, text string) error {
		msg := message{
			Type:     "1",
			Source:   lg.Source,
			Pipeline: lg.Pipeline,
			Tags:     lg.Tags,
			Log:      text,
		}
		msg.addTags("filename", filename)
		data, err := msg.json()
		if err != nil {
			return err
		}

		if err := fn(data); err != nil {
			l.Errorf("client write failed: %s", err.Error())
			return err
		}
		return nil
	}
}

func startTailing(lg *logging, fn tailer.ForwardFunc, stop <-chan struct{}) {
	opt := &tailer.Option{
		Source:                lg.Source,
		Pipeline:              lg.Pipeline,
		CharacterEncoding:     lg.CharacterEncoding,
		MultilinePatterns:     []string{lg.MultilineMatch},
		RemoveAnsiEscapeCodes: lg.RemoveAnsiEscapeCodes,
		ForwardFunc:           fn,
		FromBeginning:         true,
	}

	tailer, err := tailer.NewTailer(lg.LogFiles, opt, lg.Ignore)
	if err != nil {
		l.Error(err)
		return
	}

	go tailer.Start()

	<-stop
	tailer.Close()
}

// main config

type config struct {
	DataKitAddr string   `json:"datakit_addr"`
	Loggings    loggings `json:"loggings"`
}

func parseConfig(s string) (*config, error) {
	if s == "" {
		return nil, fmt.Errorf("invalid logfwd config")
	}

	var configs []*config

	if err := json.Unmarshal([]byte(s), &configs); err != nil {
		return nil, err
	}

	if len(configs) == 0 {
		return nil, fmt.Errorf("not found fwd config")
	}

	cfg := configs[0]
	if cfg == nil {
		return nil, fmt.Errorf("unreachable, invalid config pointer")
	}

	if wsHost != "" && wsPort != "" {
		cfg.DataKitAddr = fmt.Sprintf("%s:%s", wsHost, wsPort)
	}

	var annotationLoggings loggings
	if envAnnotationDataKitLogs != "" {
		_ = json.Unmarshal([]byte(envAnnotationDataKitLogs), &annotationLoggings)
	}
	for _, logging := range cfg.Loggings {
		logging.merge(annotationLoggings)
		logging.setup()
	}

	return cfg, nil
}

// logging config

type loggings []*logging

type logging struct {
	LogFiles              []string          `json:"logfiles"`
	Ignore                []string          `json:"ignore"`
	Source                string            `json:"source"`
	Service               string            `json:"service"`
	Pipeline              string            `json:"pipeline"`
	CharacterEncoding     string            `json:"character_encoding"`
	MultilineMatch        string            `json:"multiline_match"`
	RemoveAnsiEscapeCodes bool              `json:"remove_ansi_escape_codes"`
	Tags                  map[string]string `json:"tags"`
}

func (lg *logging) merge(cfgs loggings) {
	if len(cfgs) == 0 {
		return
	}
	for _, c := range cfgs {
		if lg.Source != c.Source {
			continue
		}
		lg.Service = c.Service
		lg.Pipeline = c.Pipeline
		lg.MultilineMatch = c.MultilineMatch
	}
}

func (lg *logging) setup() {
	if lg.Source == "" {
		lg.Source = "default"
	}
	if lg.Service == "" {
		lg.Service = lg.Source
	}

	lg.addTags("service", lg.Service)
	if podName != "" {
		lg.addTags("pod_name", podName)
	}
	if podNamespace != "" {
		lg.addTags("pod_namespace", podNamespace)
	}
}

func (lg *logging) addTags(key, value string) {
	if lg.Tags == nil {
		lg.Tags = make(map[string]string)
	}
	lg.Tags[key] = value
}

// message

type message struct {
	Type     string            `json:"type"`
	Source   string            `json:"source"`
	Pipeline string            `json:"pipeline,omitempty"`
	Tags     map[string]string `json:"tags,omitempty"`
	Log      string            `json:"log"`
}

func (m *message) addTags(key, value string) {
	if m.Tags == nil {
		m.Tags = make(map[string]string)
	}
	m.Tags[key] = value
}

func (m *message) json() ([]byte, error) {
	j, err := json.Marshal(m)
	if err != nil {
		return nil, err
	}
	return j, nil
}

// wsclient

type wsclient struct {
	u      *url.URL
	conn   *websocket.Conn
	dataCh chan []byte
}

func newWsclient(u *url.URL) *wsclient {
	return &wsclient{
		u:      u,
		dataCh: make(chan []byte, 64),
	}
}

func (w *wsclient) start() {
	for {
		data := <-w.dataCh
		err := w.conn.WriteMessage(websocket.TextMessage, data)
		if err != nil {
			l.Errorf("client write failed: %s", err.Error())
			w.tryConnectWebsocketSrv()
		}
	}
}

func (w *wsclient) close() error {
	if w.conn == nil {
		return nil
	}
	return w.conn.Close()
}

func (w *wsclient) tryConnectWebsocketSrv() {
	for {
		wscli, _, err := websocket.DefaultDialer.Dial(w.u.String(), nil)
		if err != nil {
			l.Errorf("failed to connect: %s", err.Error())
			time.Sleep(time.Second)
			continue
		}
		w.conn = wscli
		return
	}
}

func (w *wsclient) writeMessage(data []byte) error {
	// abstraction
	w.dataCh <- data
	// fmt.Errorf("failed to write channel")
	return nil
}
