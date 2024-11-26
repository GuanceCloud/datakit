// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

// Package graphite collector.
package graphite

import (
	"bufio"
	"bytes"
	"context"
	"io"
	"net"
	"strconv"
	"strings"
	"time"

	"github.com/GuanceCloud/cliutils"
	"github.com/GuanceCloud/cliutils/logger"
	"github.com/GuanceCloud/cliutils/point"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/config"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/datakit"
	dkio "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/metrics"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs/graphite/mapper"
)

type Input struct {
	Address string            `toml:"address"` // 使用的UDP/TCP端口
	Tags    map[string]string `toml:"tags"`

	semStop *cliutils.Sem // start stop signal
	feeder  dkio.Feeder
	tagger  datakit.GlobalTagger

	MetricMapper mapper.MetricMapper `toml:"metric_mapper"`
	StrictMatch  bool                `toml:"strict_match"`
	Interval     datakit.Duration    `toml:"interval"`
	BufferSize   int                 `toml:"buffer_size"`

	logger      *logger.Logger
	tcpListener net.Listener
	udpListener *net.UDPConn

	lineCh chan string
	outCh  chan *graphiteMetric
	quitCh chan struct{}
}

func (*Input) Catalog() string {
	return inputName
}

func (ipt *Input) Terminate() {
	if ipt.semStop != nil {
		ipt.semStop.Close()
	}
}

func (*Input) SampleConfig() string {
	return sampleConfig
}

func (*Input) SampleMeasurement() []inputs.Measurement {
	return []inputs.Measurement{&Measurement{}}
}

func (*Input) AvailableArchs() []string {
	return datakit.AllOS
}

func (ipt *Input) Run() {
	ipt.Interval.Duration = config.ProtectedInterval(minInterval, maxInterval, ipt.Interval.Duration)

	if err := ipt.MetricMapper.MapperSetup(); err != nil {
		ipt.logger.Infof("error occur: %v", err)
		return
	}

	g.Go(func(_ context.Context) error {
		ipt.reportMetric()
		return nil
	})
	g.Go(func(_ context.Context) error {
		ipt.doTCP(ipt.Address)
		return nil
	})
	g.Go(func(_ context.Context) error {
		ipt.doUDP(ipt.Address)
		return nil
	})
	g.Go(func(_ context.Context) error {
		ipt.processLines()
		return nil
	})

	select {
	case <-datakit.Exit.Wait():
		ipt.exit()
		ipt.logger.Info("graphite exit")
		return
	case <-ipt.semStop.Wait():
		ipt.exit()
		ipt.logger.Info("graphite return")
		return
	}
}

func (ipt *Input) doTCP(addr string) {
	listen, err := net.Listen(TCP, addr)
	if err != nil {
		ipt.logger.Errorf("Error binding to TCP socket, error=%v", err)
		return
	}
	ipt.tcpListener = listen
	//nolint:errcheck
	defer listen.Close()
	for {
		conn, err := listen.Accept()
		if err != nil {
			select {
			case <-ipt.quitCh:
				return
			default:
				ipt.logger.Errorf("Error accepting TCP connection, error=%v", err)
				continue
			}
		}
		g.Go(func(_ context.Context) error {
			//nolint:errcheck
			defer conn.Close()
			ipt.processReader(conn)
			return nil
		})
	}
}

func (ipt *Input) doUDP(addr string) {
	udpAddress, err := net.ResolveUDPAddr(UDP, addr)
	if err != nil {
		ipt.logger.Errorf("Error resolving UDP address, error=%v", err)
		return
	}
	listen, err := net.ListenUDP(UDP, udpAddress)
	if err != nil {
		ipt.logger.Errorf("Error listening to UDP address, error=%v", err)
		return
	}
	ipt.udpListener = listen
	//nolint:errcheck
	defer listen.Close()
	for {
		buf := make([]byte, 65536)
		chars, srcAddr, err := listen.ReadFromUDP(buf)
		if err != nil {
			select {
			case <-ipt.quitCh:
				return
			default:
				ipt.logger.Errorf("Error reading UDP packet from %v, error=%v", srcAddr, err)
				continue
			}
		}
		g.Go(func(_ context.Context) error {
			ipt.processReader(bytes.NewReader(buf[0:chars]))
			return nil
		})
	}
}

func (ipt *Input) processReader(reader io.Reader) {
	ls := bufio.NewScanner(reader)
	for {
		if ok := ls.Scan(); !ok {
			break
		}
		ipt.lineCh <- ls.Text()
	}
}

func (ipt *Input) processLines() {
	for line := range ipt.lineCh {
		ipt.processLine(line)
	}
}

func (ipt *Input) processLine(line string) {
	line = strings.TrimSpace(line)

	ipt.logger.Infof("line %s", line)

	parts := strings.Split(line, " ")
	if len(parts) != 3 {
		ipt.logger.Infof("invalid part count, get %d parts, line: %s\n", len(parts), line)
		return
	}

	originalName := parts[0]

	parsedName, labels, err := ipt.parseMetricNameAndTags(originalName)
	if err != nil {
		ipt.logger.Infof("invalid tags, line: %s\n, err: %v", line, err)
	}

	mapping, mappingLabels, mappingPresent := ipt.MetricMapper.GetMapping(parsedName, mapper.MetricTypeGauge)

	// add mapping labels to parsed labels
	for k, v := range mappingLabels {
		labels[k] = v
	}

	if (mappingPresent && mapping.Action == mapper.ActionTypeDrop) || (!mappingPresent && ipt.StrictMatch) {
		return
	}

	var name string
	if mappingPresent {
		name = invalidMetricChars.ReplaceAllString(mapping.Name, "_")
	} else {
		name = invalidMetricChars.ReplaceAllString(parsedName, "_")
	}

	value, err := strconv.ParseFloat(parts[1], 64)
	if err != nil {
		ipt.logger.Infof("invalid value, line: %s", line)
		return
	}
	timestamp, err := strconv.ParseFloat(parts[2], 64)
	if err != nil {
		ipt.logger.Infof("invalid timestamp, line: %s", line)
		return
	}

	var measurementName string
	if mappingPresent {
		measurementName = mapping.MeasurementName
	} else {
		measurementName = defaultMeasurement
	}

	m := graphiteMetric{
		OriginalName:    originalName,
		MeasurementName: measurementName,
		Name:            name,
		Value:           value,
		Labels:          labels,
		Timestamp:       int64(timestamp * 1e9),
	}
	lastProcessed.Set(float64(time.Now().UnixNano()) / 1e9)
	ipt.outCh <- &m
}

func (ipt *Input) parseMetricNameAndTags(name string) (string, mapper.Labels, error) {
	var err error

	labels := make(mapper.Labels)

	parts := strings.Split(name, ";")
	parsedName := parts[0]

	tags := parts[1:]
	for _, tag := range tags {
		kv := strings.SplitN(tag, "=", 2)
		if len(kv) != 2 {
			tagParseFailures.Inc()
			continue
		}

		k := kv[0]
		v := kv[1]
		labels[k] = v
	}

	return parsedName, labels, err
}

func (ipt *Input) reportMetric() {
	ticker := time.NewTicker(ipt.Interval.Duration)
	defer ticker.Stop()

	var buffer []*graphiteMetric

	for {
		start := time.Now()
		select {
		case metric := <-ipt.outCh:
			buffer = append(buffer, metric)

			// 超过缓冲区大小，立即上报
			if len(buffer) >= ipt.BufferSize {
				ipt.sendMetric(buffer, start)
				buffer = buffer[:0]
			}
		case <-ticker.C:
			// 定时上报
			if len(buffer) > 0 {
				ipt.sendMetric(buffer, start)
				buffer = buffer[:0]
			}
		case <-ipt.semStop.Wait():
			if len(buffer) > 0 {
				ipt.sendMetric(buffer, start)
			}
			return
		}
	}
}

func (ipt *Input) sendMetric(measurements []*graphiteMetric, start time.Time) {
	var pts []*point.Point

	opts := point.DefaultMetricOptions()

	for i := range measurements {
		metric := measurements[i]
		kvs := make(point.KVs, 0)

		for k, v := range metric.Labels {
			kvs = kvs.MustAddTag(k, v)
		}

		kvs = kvs.Add(metric.Name, metric.Value, false, true)
		kvs = kvs.Add("timestamp", metric.Timestamp, false, true)

		pts = append(pts, point.NewPointV2(metric.MeasurementName, kvs, opts...))
	}

	if len(pts) > 0 {
		if err := ipt.feeder.FeedV2(point.Metric, pts,
			dkio.WithCollectCost(time.Since(start)),
			dkio.WithInputName(inputName),
		); err != nil {
			ipt.feeder.FeedLastError(err.Error(),
				metrics.WithLastErrorInput(inputName),
				metrics.WithLastErrorCategory(point.Metric),
			)
			ipt.logger.Errorf("feed: %v", err)
		}
	}
}

func (ipt *Input) exit() {
	ipt.logger.Infof("graphite shutting down.")

	close(ipt.quitCh)

	if ipt.tcpListener != nil {
		if err := ipt.tcpListener.Close(); err != nil {
			ipt.logger.Errorf("error closing tcp listener: %v", err)
		}
		ipt.tcpListener = nil
	}

	if ipt.udpListener != nil {
		if err := ipt.udpListener.Close(); err != nil {
			ipt.logger.Errorf("error closing tcp listener: %v", err)
		}
		ipt.udpListener = nil
	}

	close(ipt.lineCh)
	close(ipt.outCh)

	if err := g.Wait(); err != nil {
		ipt.logger.Errorf("error waiting goroutine: %v", err)
	}

	ipt.logger.Infof("graphite shutting down complete.")
}

//------------------------------------------------------------------------------

func defaultInput() *Input {
	return &Input{
		feeder:       dkio.DefaultFeeder(),
		semStop:      cliutils.NewSem(),
		tagger:       datakit.DefaultGlobalTagger(),
		logger:       logger.DefaultSLogger("graphite"),
		BufferSize:   defaultBufferSize,
		lineCh:       make(chan string),
		outCh:        make(chan *graphiteMetric),
		quitCh:       make(chan struct{}),
		MetricMapper: mapper.MetricMapper{},
		Address:      defaultPort,
	}
}

func init() { //nolint:gochecknoinits
	inputs.Add(inputName, func() inputs.Input {
		return defaultInput()
	})
}
