// Package beats_output receive and process multiple elastic beats output data.
package beats_output

import (
	"encoding/json"
	"fmt"
	"net"
	"net/url"
	"strings"
	"time"

	"gitlab.jiagouyun.com/cloudcare-tools/cliutils"
	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/logger"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"

	v2 "github.com/elastic/go-lumber/server/v2"
)

//------------------------------------------------------------------------------

func (ipt *Input) Run() {
	l = logger.SLogger(inputName)
	l.Debug("Run entry")

	if ipt.Source == "" {
		ipt.Source = "default"
	}
	if ipt.Service == "" {
		ipt.Service = ipt.Source
	}

	opServer, err := NewOutputServerTCP(ipt.Listen)
	if err != nil {
		l.Errorf("NewOutputServerTCP failed: %v", err)
		l.Warn(inputName + ".Run() exit")
		return
	}
	server, _ := v2.NewWithListener(opServer.Listener)
	defer server.Close()

	l.Debug("listening...")
	ipt.stopped = false

	go func() {
		for {
			if ipt.stopped {
				return
			}

			// try to receive event from server
			batch := server.Receive()
			if batch == nil {
				continue
			}

			// host.name
			// log.file.path
			// message
			for _, v := range batch.Events {
				// hostName := eventGet(v, "host.name").(string)
				// logFilePath := eventGet(v, "log.file.path").(string)
				// message := eventGet(v, "message").(string)

				l.Infof("host.name = %s", eventGet(v, "host.name").(string))
				l.Infof("log.file.path = %s", eventGet(v, "log.file.path").(string))
				l.Infof("message = %s", eventGet(v, "message").(string))
			}

			batch.ACK()
		}
	}()

	for {
		select {
		case <-datakit.Exit.Wait():
			ipt.exit()
			l.Info(inputName + " exit")
			return

		case <-ipt.semStop.Wait():
			ipt.exit()
			l.Infof(inputName + " return")
			return
		}
	}
}

func (ipt *Input) exit() {
	ipt.stopped = true
}

func (ipt *Input) Terminate() {
	if ipt.semStop != nil {
		ipt.semStop.Close()
	}
}

//------------------------------------------------------------------------------

func NewOutputServerTCP(listen string) (*OutputServer, error) {
	mVal, err := parseListen(listen)
	if err != nil {
		return nil, fmt.Errorf("parseListen failed: %v", err)
	}

	tcpListener, err := net.Listen(mVal["scheme"], mVal["host"])
	if err != nil {
		return nil, fmt.Errorf("net.Listen failed: %v", err)
	}

	return &OutputServer{Listener: tcpListener}, nil
}

type OutputServer struct {
	net.Listener
	Timeout time.Duration
	Err     error
}

func (m *OutputServer) Addr() string {
	return m.Listener.Addr().String()
}

func (m *OutputServer) Accept() net.Conn {
	if m.Err != nil {
		return nil
	}

	client, err := m.Listener.Accept()
	m.Err = err
	return client
}

//------------------------------------------------------------------------------

func debugPrettyPrintMap(x map[string]interface{}) string {
	b, err := json.MarshalIndent(x, "", "  ")
	if err != nil {
		fmt.Println("error:", err)
	}
	return string(b)
}

func parseListen(listen string) (map[string]string, error) {
	uurl, err := url.Parse(listen)
	if err != nil {
		return nil, err
	}

	mVal := make(map[string]string)
	mVal["scheme"] = uurl.Scheme
	mVal["host"] = uurl.Host

	return mVal, nil
}

func eventGet(event interface{}, path string) interface{} {
	doc := event.(map[string]interface{})
	elems := strings.Split(path, ".")
	for i := 0; i < len(elems)-1; i++ {
		doc = doc[elems[i]].(map[string]interface{})
	}
	return doc[elems[len(elems)-1]]
}

//------------------------------------------------------------------------------

const (
	inputName = "beats_output"

	sampleCfg = `
[[inputs.beats_output]]
  # listen address, with protocol scheme and port
  listen = "tcp://0.0.0.0:5044"

  ## source, if it's empty, use 'default'
  source = ""

  ## add service tag, if it's empty, use $source.
  service = ""

  ## grok pipeline script name
  pipeline = ""

  [inputs.beats_output.tags]
  # some_tag = "some_value"
  # more_tag = "some_other_value"
`
)

type Input struct {
	Listen   string            `toml:"listen"`
	Source   string            `toml:"source"`
	Service  string            `toml:"service"`
	Pipeline string            `toml:"pipeline"`
	Tags     map[string]string `toml:"tags"`

	semStop *cliutils.Sem // start stop signal
	stopped bool
}

var _ inputs.InputV2 = &Input{}

var l = logger.DefaultSLogger(inputName)

func (*Input) Catalog() string { return inputName }

func (*Input) SampleConfig() string { return sampleCfg }

func (*Input) AvailableArchs() []string { return datakit.AllArch }

//------------------------------------------------------------------------------

type loggingMeasurement struct {
	name   string
	tags   map[string]string
	fields map[string]interface{}
	ts     time.Time
}

func (*Input) SampleMeasurement() []inputs.Measurement {
	return []inputs.Measurement{
		&loggingMeasurement{},
	}
}

func (ipt *loggingMeasurement) LineProto() (*io.Point, error) {
	return io.MakePoint(ipt.name, ipt.tags, ipt.fields, ipt.ts)
}

//nolint:lll
func (*loggingMeasurement) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name: "接收器",
		Type: "logging",
		Desc: "使用配置文件中的 `source` 字段值，如果该值为空，则默认为 `default`",
		Tags: map[string]interface{}{
			"filepath": inputs.NewTagInfo(`此条记录来源的文件名，全路径`), // log.file.path
			"host":     inputs.NewTagInfo(`主机名`),            // host.name
		},
		Fields: map[string]interface{}{
			"message": &inputs.FieldInfo{DataType: inputs.String, Unit: inputs.UnknownUnit, Desc: "记录正文，默认存在，可以使用 pipeline 删除此字段"}, // message
		},
	}
}

//------------------------------------------------------------------------------

func init() { //nolint:gochecknoinits
	inputs.Add(inputName, func() inputs.Input {
		return &Input{
			Tags:    make(map[string]string),
			semStop: cliutils.NewSem(),
		}
	})
}
