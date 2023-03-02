// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

// Package beats_output receive and process multiple Elastic Beats output data.
package beats_output //nolint:stylecheck

import (
	"context"
	"fmt"
	"net"
	"net/url"
	"reflect"
	"strings"
	"time"

	"github.com/GuanceCloud/cliutils"
	"github.com/GuanceCloud/cliutils/logger"
	"github.com/GuanceCloud/cliutils/point"
	v2 "github.com/elastic/go-lumber/server/v2"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io"
	dkpt "gitlab.jiagouyun.com/cloudcare-tools/datakit/io/point"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/pipeline"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
)

//------------------------------------------------------------------------------

const (
	inputName       = "beats_output"
	measurementName = "Elastic Beats Collector"

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

  ## would replaced by origin fields if repeated
  [inputs.beats_output.tags]
  # some_tag = "some_value"
  # more_tag = "some_other_value"
`
)

var g = datakit.G("inputs_beats_output")

type Input struct {
	Listen        string            `toml:"listen"`
	Source        string            `toml:"source"`
	Service       string            `toml:"service"`
	Pipeline      string            `toml:"pipeline"`
	MaximumLength int               `toml:"maximum_length,omitempty"`
	Tags          map[string]string `toml:"tags"`

	feeder io.Feeder

	semStop *cliutils.Sem // start stop signal
	stopped bool
}

// Make sure Input implements the inputs.InputV2 interface.
var _ inputs.InputV2 = &Input{}

var l = logger.DefaultSLogger(inputName)

func (*Input) Catalog() string { return inputName }

func (*Input) SampleConfig() string { return sampleCfg }

func (*Input) AvailableArchs() []string { return datakit.AllOS }

type DataStruct struct {
	HostName    string
	LogFilePath string
	Message     string
	Fields      map[string]interface{}
}

//------------------------------------------------------------------------------

type loggingMeasurement struct {
	name   string
	tags   map[string]string
	fields map[string]interface{}
}

func (*Input) SampleMeasurement() []inputs.Measurement {
	return []inputs.Measurement{
		&loggingMeasurement{},
	}
}

// Point implement MeasurementV2.
func (ipt *loggingMeasurement) Point() *point.Point {
	opts := point.DefaultLoggingOptions()

	return point.NewPointV2([]byte(ipt.name),
		append(point.NewTags(ipt.tags), point.NewKVs(ipt.fields)...),
		opts...)
}

func (ipt *loggingMeasurement) LineProto() (*dkpt.Point, error) {
	// return point.NewPoint(ipt.name, ipt.tags, ipt.fields, point.LOpt())
	return nil, fmt.Errorf("not implement")
}

//nolint:lll
func (*loggingMeasurement) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name: measurementName,
		Type: "logging",
		Desc: "Using `source` field in the config file, default is `default`.",
		Tags: map[string]interface{}{
			"filepath": inputs.NewTagInfo(`This item source file, full path.`), // log.file.path
			"host":     inputs.NewTagInfo(`Host name.`),                        // host.name
			"service":  inputs.NewTagInfo("Service name, equal to `service` field in the config file."),
		},
		Fields: map[string]interface{}{
			"message": &inputs.FieldInfo{DataType: inputs.String, Unit: inputs.UnknownUnit, Desc: "Message text, existed when default. Could use pipeline to delete this field."}, // message
			"status":  &inputs.FieldInfo{DataType: inputs.String, Unit: inputs.UnknownUnit, Desc: "Log status."},
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

	g.Go(func(ctx context.Context) error {
		for {
			if ipt.stopped {
				return nil
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
				dataPiece := getDataPieceFromEvent(v)
				pending = append(pending, dataPiece)
			}
			ipt.feed(pending)

			batch.ACK()
		}
	})

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

func (ipt *Input) getNewTags(dataPiece *DataStruct) map[string]string {
	// merge map
	newTags := make(map[string]string)
	for kk, vv := range ipt.Tags {
		newTags[kk] = vv
	}
	if len(dataPiece.HostName) > 0 {
		newTags["host"] = dataPiece.HostName // host.name
	}
	if len(dataPiece.LogFilePath) > 0 {
		newTags["filepath"] = dataPiece.LogFilePath // log.file.path
	}
	if len(ipt.Service) > 0 {
		newTags["service"] = ipt.Service
	}
	for kk, vv := range dataPiece.Fields {
		if str, ok := vv.(string); ok {
			newTags[kk] = str
		} else {
			// TODO: 这里有 bug。这里的 Tag 目前只有 map[string]string 形式，
			// 没有 map[string]interface{} 形式，后面 PL 改良了，这边需要再改一下。
			// 现在是强行 format 为 string
			switch n := vv.(type) {
			case int, int64, int32:
				newTags[kk] = fmt.Sprintf("%d", n)
			default:
				l.Warnf("ignore fields: %s, type name = %s, string = %s",
					kk, reflect.TypeOf(n).Name(), reflect.TypeOf(n).String())
			}
		}
	}
	return newTags
}

func (ipt *Input) feed(pending []*DataStruct) {
	pts := []*point.Point{}
	for _, v := range pending {
		if len(v.Message) == 0 {
			continue
		}

		newTags := ipt.getNewTags(v)
		l.Debugf("newTags = %#v", newTags)

		// pt, err := point.NewPoint(ipt.Source, newTags,
		// 	map[string]interface{}{
		// 		pipeline.FieldMessage: v.Message,
		// 		pipeline.FieldStatus:  pipeline.DefaultStatus,
		// 	}, point.LOpt())
		// if err != nil {
		// 	l.Error(err)
		// 	continue
		// }

		logging := &loggingMeasurement{
			name:   measurementName,
			tags:   newTags,
			fields: v.Fields,
		}

		// map[string]interface{}{
		// 	pipeline.FieldMessage: v.Message,
		// 	pipeline.FieldStatus:  pipeline.DefaultStatus,
		// }
		if logging.fields == nil {
			logging.fields = make(map[string]interface{})
		}
		logging.fields[pipeline.FieldMessage] = v.Message
		logging.fields[pipeline.FieldStatus] = pipeline.DefaultStatus

		pts = append(pts, logging.Point())
	}
	if len(pts) > 0 {
		if err := ipt.feeder.Feed(inputName+"/"+ipt.Listen, point.Logging, pts, &io.Option{
			PlScript: map[string]string{
				ipt.Source: ipt.Pipeline,
			},
		}); err != nil {
			l.Error(err)
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

func getEventPrint(event interface{}) map[string]interface{} {
	eventMap, ok := event.(map[string]interface{})
	if !ok {
		return nil
	}
	return eventMap
}

func getEventPathStringValue(event interface{}, path string) string {
	val, ok := eventGet(event, path).(string)
	if !ok {
		l.Warnf("cannot find %s, event = %#v", path, event)
	}
	return val
}

func getDataPieceFromEvent(event interface{}) *DataStruct {
	// debug print
	eventMap := getEventPrint(event)
	if eventMap != nil {
		l.Debugf("event = %#v", eventMap)
	}

	hostName := getEventPathStringValue(event, "host.name")
	logFilePath := getEventPathStringValue(event, "log.file.path")
	message := getEventPathStringValue(event, "message")

	dataPiece := &DataStruct{
		HostName:    hostName,
		LogFilePath: logFilePath,
		Message:     message,
	}
	fields, ok := eventGet(event, "fields").(map[string]interface{})
	if ok {
		dataPiece.Fields = fields
	}
	return dataPiece
}

//------------------------------------------------------------------------------

func defaultInput() *Input {
	return &Input{
		Tags:    make(map[string]string),
		feeder:  io.DefaultFeeder(),
		semStop: cliutils.NewSem(),
	}
}

func init() { //nolint:gochecknoinits
	inputs.Add(inputName, func() inputs.Input {
		return defaultInput()
	})
}
