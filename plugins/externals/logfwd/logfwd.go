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

const (
	name = "logfwd"

	envPodNameKey      = "LOGFWD_POD_NAME"
	envPodNamespaceKey = "LOGFWD_POD_NAMESPACE"

	envLogConfigKey = "LOGFWD_ANNOTATION_DATAKIT_LOGS"
	envFwdConfigKey = "LOGFWD_JSON_CONFIG"

	envWsHostKey = "LOGFWD_DATAKIT_HOST"
	envWsPortKey = "LOGFWD_DATAKIT_PORT"
)

var (
	argConfigJSON = flag.String("json-config", "", "logfwd json-config")
	argConfig     = flag.String("config", "", "logfwd config file")
	l             = logger.DefaultSLogger(name)
)

func main() {
	quitChannel := make(chan struct{})

	flag.Parse()
	l = logger.SLogger(name)

	cfg, err := getFwdConfig()
	if err != nil {
		l.Error(err)
		l.Info("exit")
		os.Exit(0)
	}

	startLog(cfg, quitChannel)

	<-quitChannel
}

func getFwdConfig() (*fwdConfig, error) {
	cfg := func() string {
		if c := os.Getenv(envFwdConfigKey); c != "" {
			return c
		}
		if *argConfigJSON != "" {
			return *argConfigJSON
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

	return parseFwdConfig(cfg)
}

func startLog(cfg *fwdConfig, stop <-chan struct{}) {
	u := url.URL{Scheme: "ws", Host: cfg.DataKitAddr, Path: "/logfwd"}

	var wg sync.WaitGroup

	for _, c := range cfg.LogConfigs {
		wg.Add(1)

		go func(lc *logConfig) {
			defer wg.Done()

			wscli := newWsclient(&u)
			wscli.tryConnectWebsocketSrv()
			go wscli.start()

			defer func() {
				if err := wscli.close(); err != nil {
					l.Errorf("failed to close websocket client, err: %w", err)
				}
			}()

			startTailing(lc, forwardFunc(lc, wscli.writeMessage), stop)
		}(c)
	}

	wg.Wait()
}

type writeMessage func([]byte) error

func forwardFunc(lc *logConfig, fn writeMessage) tailer.ForwardFunc {
	return func(filename, text string) error {
		msg := message{
			Type:     "1",
			Source:   lc.Source,
			Pipeline: lc.Pipeline,
			TagsStr:  lc.TagsStr,
			Log:      text,
		}
		_ = msg.appendToTagsStr("filename", filename)

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

func startTailing(lc *logConfig, fn tailer.ForwardFunc, stop <-chan struct{}) {
	opt := &tailer.Option{
		Source:                lc.Source,
		Pipeline:              lc.Pipeline,
		CharacterEncoding:     lc.CharacterEncoding,
		MultilineMatch:        lc.MultilineMatch,
		RemoveAnsiEscapeCodes: lc.RemoveAnsiEscapeCodes,
		ForwardFunc:           fn,
		DisableSendEvent:      true,
	}

	tailer, err := tailer.NewTailer(lc.LogFiles, opt, lc.Ignore)
	if err != nil {
		l.Error(err)
		return
	}

	go tailer.Start()

	<-stop
	tailer.Close()
}

type fwdConfig struct {
	DataKitAddr string `json:"datakit_addr"`
	LogPath     string `json:"log_path,omitempty"`
	LogLevel    string `json:"log_level,omitempty"`

	LogConfigs logConfigs `json:"loggings"`
}

func parseFwdConfig(configStr string) (*fwdConfig, error) {
	if configStr == "" {
		return nil, fmt.Errorf("invalid fwd config")
	}

	var configs []*fwdConfig

	if err := json.Unmarshal([]byte(configStr), &configs); err != nil {
		return nil, err
	}

	if len(configs) == 0 {
		return nil, fmt.Errorf("not found fwd config")
	}

	cfg := configs[0]
	if cfg == nil {
		return nil, fmt.Errorf("unreachable, invalid config pointer")
	}

	if os.Getenv(envWsHostKey) != "" && os.Getenv(envWsPortKey) != "" {
		addr := fmt.Sprintf("%s:%s", os.Getenv(envWsHostKey), os.Getenv(envWsPortKey))
		cfg.DataKitAddr = addr
		l.Infof("use env host and port, datakit address '%s'", addr)
	}

	envLogConfigs := getEnvLogConfigs(envLogConfigKey)
	for _, lc := range cfg.LogConfigs {
		lc.setup()
		lc.merge(envLogConfigs)
	}

	return cfg, nil
}

type logConfigs []*logConfig

func getEnvLogConfigs(env string) logConfigs {
	s := os.Getenv(env)
	if s == "" {
		return nil
	}
	var c logConfigs
	if err := json.Unmarshal([]byte(s), &c); err != nil {
		// l.Error(err)
		return nil
	}
	return c
}

// logConfig

type logConfig struct {
	LogFiles              []string `json:"logfiles"`
	Ignore                []string `json:"ignore"`
	Source                string   `json:"source"`
	Service               string   `json:"service"`
	Pipeline              string   `json:"pipeline"`
	CharacterEncoding     string   `json:"character_encoding"`
	MultilineMatch        string   `json:"multiline_match"`
	RemoveAnsiEscapeCodes bool     `json:"remove_ansi_escape_codes"`
	TagsStr               string   `json:"tags_str"`
}

func (lc *logConfig) merge(cfgs logConfigs) {
	for _, c := range cfgs {
		if lc.Source != c.Source {
			continue
		}
		lc.Service = c.Service
		lc.Pipeline = c.Pipeline
		lc.MultilineMatch = c.MultilineMatch
	}
}

func (lc *logConfig) setup() {
	if lc.Source == "" {
		lc.Source = "default"
	}
	if lc.Service == "" {
		lc.Service = lc.Source
	}

	lc.appendToTagsStr("service", lc.Service)
	if podName := os.Getenv(envPodNameKey); podName != "" {
		lc.appendToTagsStr("pod_name", podName)
	}
	if podNamespace := os.Getenv(envPodNamespaceKey); podNamespace != "" {
		lc.appendToTagsStr("pod_namespace", podNamespace)
	}
}

func (lc *logConfig) appendToTagsStr(key, value string) {
	if len(lc.TagsStr) == 0 {
		lc.TagsStr = fmt.Sprintf("%s=%s", key, value)
		return
	}
	lc.TagsStr = fmt.Sprintf("%s=%s,", key, value) + lc.TagsStr
}

// message

type message struct {
	Type     string `json:"type"`
	Source   string `json:"source"`
	Pipeline string `json:"pipeline,omitempty"`
	TagsStr  string `json:"tags_str,omitempty"`
	Log      string `json:"log"`
}

func (m *message) appendToTagsStr(key, value string) string {
	if len(m.TagsStr) == 0 {
		m.TagsStr = fmt.Sprintf("%s=%s", key, value)
	} else {
		m.TagsStr = fmt.Sprintf("%s=%s,", key, value) + m.TagsStr
	}
	return m.TagsStr
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
	select {
	case w.dataCh <- data:
		// nil
	default:
		return fmt.Errorf("failed to write channel")
	}
	return nil
}
