// Package beats_output receive and process multiple Elastic Beats output data.
package beats_output //nolint:stylecheck

import (
	"fmt"
	"net"
	"net/url"
	"strings"
	"time"

	v2 "github.com/elastic/go-lumber/server/v2"
	"gitlab.jiagouyun.com/cloudcare-tools/cliutils"
	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/logger"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/pipeline/worker"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
)

//------------------------------------------------------------------------------

const (
	inputName            = "beats_output"
	defaultMaximumLength = 262144

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

  ## datakit read text from Files or Socket , default max_textline is 256k
  ## If your log text line exceeds 256Kb, please configure the length of your text,
  ## but the maximum length cannot exceed 256Mb
  maximum_length = 262144

  [inputs.beats_output.tags]
  # some_tag = "some_value"
  # more_tag = "some_other_value"
`
)

type Input struct {
	Listen        string            `toml:"listen"`
	Source        string            `toml:"source"`
	Service       string            `toml:"service"`
	Pipeline      string            `toml:"pipeline"`
	MaximumLength int               `toml:"maximum_length,omitempty"`
	Tags          map[string]string `toml:"tags"`

	semStop *cliutils.Sem // start stop signal
	stopped bool
}

var _ inputs.InputV2 = &Input{}

var l = logger.DefaultSLogger(inputName)

func (*Input) Catalog() string { return inputName }

func (*Input) SampleConfig() string { return sampleCfg }

func (*Input) AvailableArchs() []string { return datakit.AllArch }

type DataStruct struct {
	HostName    string
	LogFilePath string
	Message     string
}

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
		Name: "Elastic Beats 接收器",
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

func (ipt *Input) Run() {
	l = logger.SLogger(inputName)
	l.Debug("Run entry")

	if ipt.Source == "" {
		ipt.Source = "default"
	}
	if ipt.Service == "" {
		ipt.Service = ipt.Source
	}
	if ipt.MaximumLength == 0 {
		ipt.MaximumLength = defaultMaximumLength
	}

	opServer, err := NewOutputServerTCP(ipt.Listen)
	if err != nil {
		l.Errorf("NewOutputServerTCP failed: %v", err)
		l.Warn(inputName + ".Run() exit")
		return
	}
	server, _ := v2.NewWithListener(opServer.Listener)
	defer server.Close() //nolint:errcheck

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
			var pending []*DataStruct
			for _, v := range batch.Events {
				pending = append(pending, &DataStruct{
					HostName:    eventGet(v, "host.name").(string),
					LogFilePath: eventGet(v, "log.file.path").(string),
					Message:     eventGet(v, "message").(string),
				})
			}
			ipt.sendToPipeline(pending)

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

func (ipt *Input) sendToPipeline(pending []*DataStruct) {
	for _, v := range pending {
		if len(v.Message) == 0 {
			continue
		}

		// merge map
		newTags := make(map[string]string)
		for kk, vv := range ipt.Tags {
			newTags[kk] = vv
		}
		newTags["host.name"] = v.HostName
		newTags["log.file.path"] = v.LogFilePath

		task := &worker.TaskTemplate{
			TaskName:        inputName + "/" + ipt.Listen,
			ContentDataType: worker.ContentString,
			Tags:            newTags,
			ScriptName:      ipt.Pipeline,
			Source:          ipt.Source,
			Content:         []string{v.Message},
			Category:        datakit.Logging,
			TS:              time.Now(),
			MaxMessageLen:   ipt.MaximumLength,
		}
		// 阻塞型channel
		_ = worker.FeedPipelineTaskBlock(task)
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
		return nil, fmt.Errorf("parseListen failed: %w", err)
	}

	tcpListener, err := net.Listen(mVal["scheme"], mVal["host"])
	if err != nil {
		return nil, fmt.Errorf("net.Listen failed: %w", err)
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
	doc, ok := event.(map[string]interface{})
	if !ok {
		return nil
	}
	elems := strings.Split(path, ".")
	for i := 0; i < len(elems)-1; i++ {
		doc, ok = doc[elems[i]].(map[string]interface{})
		if !ok {
			return nil
		}
	}
	return doc[elems[len(elems)-1]]
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
