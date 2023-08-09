// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

// Package memcached collects memcached metrics.
package memcached

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"net"
	"strconv"
	"time"

	"github.com/GuanceCloud/cliutils"
	"github.com/GuanceCloud/cliutils/logger"
	"github.com/GuanceCloud/cliutils/point"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/config"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/goroutine"
	dkio "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/io"
	dkpt "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/io/point"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs"
)

const sampleConfig = `
[[inputs.memcached]]
  ## Servers' addresses.
  servers = ["localhost:11211"]
  # unix_sockets = ["/var/run/memcached.sock"]

  ## Set true to enable election
  election = true

  ## Collect extra stats
  # extra_stats = ["slabs", "items"]

  ## Collect interval.
  # 单位 "ns", "us" (or "µs"), "ms", "s", "m", "h"
  interval = "10s"

[inputs.memcached.tags]
  # some_tag = "some_value"
  # more_tag = "some_other_value"
  # ...
`

var (
	inputName      = "memcached"
	catalogName    = "db"
	l              = logger.DefaultSLogger("memcached")
	defaultTimeout = 5 * time.Second
)

const emptySlabID = "empty_slab_id"

type (
	getValueFunc func(key, value []byte) (slabID string, fields map[string]interface{}, err error)
	collectFunc  func(net.Conn, *inputs.MeasurementInfo, map[string]string) ([]*point.Point, error)
	collectItem  struct {
		metricInfo *inputs.MeasurementInfo
		collector  collectFunc
	}
)

type Input struct {
	Servers     []string          `toml:"servers"`
	UnixSockets []string          `toml:"unix_sockets"`
	Election    bool              `toml:"election"`
	ExtraStats  []string          `toml:"extra_stats"` // may be slabs or items
	Interval    string            `toml:"interval"`
	Tags        map[string]string `toml:"tags"`

	duration           time.Duration
	collectCache       []*point.Point
	metricCollectorMap map[string]*collectItem

	feeder  dkio.Feeder
	semStop *cliutils.Sem // start stop signal
	opt     point.Option
}

func (*Input) Catalog() string {
	return catalogName
}

func (*Input) SampleConfig() string {
	return sampleConfig
}

func (*Input) AvailableArchs() []string {
	return datakit.AllOSWithElection
}

func (*Input) SampleMeasurement() []inputs.Measurement {
	return []inputs.Measurement{
		&inputMeasurement{},
		&itemsMeasurement{},
		&slabsMeasurement{},
	}
}

func (i *Input) ElectionEnabled() bool {
	return i.Election
}

func (i *Input) Collect() error {
	if len(i.Servers) == 0 && len(i.UnixSockets) == 0 {
		return i.gatherServer(":11211", false)
	}

	g := goroutine.NewGroup(goroutine.Option{Name: goroutine.GetInputName(inputName)})
	for _, serverAddress := range i.Servers {
		func(addr string) {
			g.Go(func(ctx context.Context) error {
				return i.gatherServer(addr, false)
			})
		}(serverAddress)
	}

	for _, unixAddress := range i.UnixSockets {
		func(unixAddr string) {
			g.Go(func(ctx context.Context) error {
				return i.gatherServer(unixAddr, true)
			})
		}(unixAddress)
	}

	return g.Wait()
}

func (i *Input) gatherServer(address string, unix bool) error {
	var conn net.Conn
	var err error
	if unix {
		conn, err = net.DialTimeout("unix", address, defaultTimeout)
		if err != nil {
			i.feeder.FeedLastError(err.Error(),
				dkio.WithLastErrorInput(inputName),
			)
			return err
		}
		defer conn.Close() //nolint:errcheck
	} else {
		_, _, err = net.SplitHostPort(address)
		if err != nil {
			address += ":11211"
		}

		conn, err = net.DialTimeout("tcp", address, defaultTimeout)
		if err != nil {
			i.feeder.FeedLastError(err.Error(),
				dkio.WithLastErrorInput(inputName),
			)
			return err
		}
		defer conn.Close() //nolint:errcheck
	}

	if conn == nil {
		return fmt.Errorf("failed to create net connection")
	}

	tags := map[string]string{"server": address}
	if i.Tags != nil {
		for k, v := range i.Tags {
			tags[k] = v
		}
	}

	for k, v := range i.metricCollectorMap {
		if points, err := v.collector(conn, v.metricInfo, tags); err != nil {
			l.Warnf("collect stats %s failed: %s", k, err.Error())
		} else {
			i.collectCache = append(i.collectCache, points...)
		}
	}

	return nil
}

func parseResponse(r *bufio.Reader) (map[string]string, error) {
	values := make(map[string]string)

	for {
		line, _, errRead := r.ReadLine()
		if errRead != nil {
			return values, errRead
		}

		if bytes.Equal(line, []byte("END")) {
			break
		}

		s := bytes.SplitN(line, []byte(" "), 3)
		if len(s) != 3 || !bytes.Equal(s[0], []byte("STAT")) {
			return values, fmt.Errorf("unexpected line in stats response: %q", line)
		}

		values[string(s[1])] = string(s[2])
	}
	return values, nil
}

func getResponseReader(conn net.Conn, command string) (reader *bufio.Reader, err error) {
	err = conn.SetDeadline(time.Now().Add(defaultTimeout))
	if err != nil {
		err = fmt.Errorf("conn.SetDeadline: %w", err)
		return
	}

	rw := bufio.NewReadWriter(bufio.NewReader(conn), bufio.NewWriter(conn))

	_, err = fmt.Fprintf(rw, "%s\r\n", command)
	if err != nil {
		return
	}

	if err = rw.Flush(); err != nil {
		return
	}

	reader = rw.Reader
	return
}

// collectStatsItems collect items stats.
func (i *Input) collectStatsItems(conn net.Conn, info *inputs.MeasurementInfo, extraTags map[string]string) (pts []*point.Point, err error) {
	reader, err := getResponseReader(conn, "stats items")
	if err != nil {
		return
	}
	return i.collectPoints(reader, info, extraTags,
		func(key, value []byte) (slabID string, fields map[string]interface{}, err error) {
			fields = make(map[string]interface{})
			keyParts := bytes.SplitN(key, []byte(":"), 3)
			if len(keyParts) != 3 || !bytes.Equal(keyParts[0], []byte("items")) {
				err = fmt.Errorf("unexpected key: %q", key)
				return
			}

			slabID = string(keyParts[1])
			fieldName := string(keyParts[2])
			if iValue, errParse := strconv.ParseInt(string(value), 10, 64); errParse == nil {
				fields[fieldName] = iValue
			} else {
				l.Warnf("invalid value: %q, expect number", value)
			}
			return
		})
}

// collectStatsSlabs collect slabs stats.
func (i *Input) collectStatsSlabs(conn net.Conn, info *inputs.MeasurementInfo, extraTags map[string]string) (pts []*point.Point, err error) {
	reader, err := getResponseReader(conn, "stats slabs")
	if err != nil {
		return
	}
	return i.collectPoints(reader, info, extraTags,
		func(key, value []byte) (slabID string, fields map[string]interface{}, err error) {
			fields = make(map[string]interface{})
			keyParts := bytes.SplitN(key, []byte(":"), 2)
			fieldName := ""
			switch {
			case len(keyParts) == 1:
				slabID = emptySlabID
				fieldName = string(keyParts[0])
			case len(keyParts) == 2:
				slabID = string(keyParts[0])
				fieldName = string(keyParts[1])
			default:
				err = fmt.Errorf("unexpected key: %q", key)
				return
			}

			if iValue, errParse := strconv.ParseInt(string(value), 10, 64); errParse == nil {
				fields[fieldName] = iValue
			} else {
				l.Warnf("invalid value: %q, expect number", value)
			}
			return
		})
}

func (i *Input) collectPoints(
	reader *bufio.Reader,
	info *inputs.MeasurementInfo,
	extraTags map[string]string,
	getValue getValueFunc,
) (pts []*point.Point, err error) {
	slabFieldsMap := map[string]map[string]interface{}{}

	for {
		line, _, errRead := reader.ReadLine()
		if errRead != nil {
			err = errRead
			break
		}

		if bytes.Equal(line, []byte("END")) {
			break
		}

		s := bytes.SplitN(line, []byte(" "), 3)
		if len(s) != 3 || !bytes.Equal(s[0], []byte("STAT")) {
			err = fmt.Errorf("unexpected line in stats response: %q", line)
			return
		}

		if slabID, fields, err := getValue(s[1], s[2]); err != nil {
			l.Warnf("get value error: %s", err.Error())
		} else {
			if _, ok := slabFieldsMap[slabID]; !ok {
				slabFieldsMap[slabID] = map[string]interface{}{}
			}
			for k, v := range fields {
				slabFieldsMap[slabID][k] = v
			}
		}
	}

	for slabID, slabfields := range slabFieldsMap {
		fields := make(map[string]interface{})
		tags := make(map[string]string)

		for field, value := range slabfields {
			if _, ok := info.Fields[field]; ok {
				fields[field] = value
			}
		}

		if slabID != emptySlabID {
			tags["slab_id"] = slabID
		}

		for k, v := range extraTags {
			tags[k] = v
		}

		metric := &inputMeasurement{
			name:   info.Name,
			tags:   tags,
			fields: fields,
			ts:     time.Now(),
			ipt:    i,
		}

		pts = append(pts, metric.Point())
	}

	return pts, err
}

func (i *Input) collectStats(conn net.Conn, info *inputs.MeasurementInfo, extraTags map[string]string) (pts []*point.Point, err error) {
	reader, err := getResponseReader(conn, "stats")
	if err != nil {
		return
	}

	return i.collectPoints(reader, info, extraTags,
		func(key, value []byte) (slabID string, fields map[string]interface{}, err error) {
			fields = make(map[string]interface{})
			slabID = emptySlabID
			fieldName := string(key)
			if iValue, errParse := strconv.ParseInt(string(value), 10, 64); errParse == nil {
				fields[fieldName] = iValue
			} else {
				l.Debugf("invalid value: %q, expect number", value)
			}
			return
		})
}

const (
	maxInterval = 1 * time.Minute
	minInterval = 1 * time.Second
)

func (i *Input) init() {
	l = logger.SLogger(inputName)

	duration, err := time.ParseDuration(i.Interval)
	if err != nil {
		l.Errorf("invalid interval, %w", err)
	} else if duration <= 0 {
		l.Error("invalid interval, cannot be less than zero")
	}

	i.duration = config.ProtectedInterval(minInterval, maxInterval, duration)

	i.metricCollectorMap = map[string]*collectItem{
		"memcache": {
			metricInfo: (&inputMeasurement{}).Info(),
			collector:  i.collectStats,
		},
	}

	for _, statType := range i.ExtraStats {
		switch statType {
		case "items":
			i.metricCollectorMap["items"] = &collectItem{
				metricInfo: (&itemsMeasurement{}).Info(),
				collector:  i.collectStatsItems,
			}
		case "slabs":
			i.metricCollectorMap["slabs"] = &collectItem{
				metricInfo: (&slabsMeasurement{}).Info(),
				collector:  i.collectStatsSlabs,
			}

		default:
			l.Warnf("Invalid extra stats type: %s, items or slabs expected", statType)
		}
	}

	if i.Election {
		i.opt = point.WithExtraTags(dkpt.GlobalElectionTags())
	} else {
		i.opt = point.WithExtraTags(dkpt.GlobalHostTags())
	}
}

func (i *Input) Run() {
	i.init()

	tick := time.NewTicker(i.duration)

	for {
		start := time.Now()
		if err := i.Collect(); err != nil {
			l.Errorf("Collect: %s", err)
			i.feeder.FeedLastError(err.Error(),
				dkio.WithLastErrorInput(inputName),
			)
		}

		if len(i.collectCache) > 0 {
			if err := i.feeder.Feed(inputName, point.Metric, i.collectCache, &dkio.Option{CollectCost: time.Since(start)}); err != nil {
				l.Errorf("FeedMeasurement: %s", err.Error())
				i.feeder.FeedLastError(err.Error(),
					dkio.WithLastErrorInput(inputName),
				)
			}
			i.collectCache = i.collectCache[:0]
		}

		select {
		case <-datakit.Exit.Wait():
			l.Info("memcached exit")
			return

		case <-i.semStop.Wait():
			l.Info("memcached return")
			return

		case <-tick.C:
		}
	}
}

func (i *Input) Terminate() {
	if i.semStop != nil {
		i.semStop.Close()
	}
}

func defaultInput() *Input {
	return &Input{
		feeder:  dkio.DefaultFeeder(),
		semStop: cliutils.NewSem(),
	}
}

func init() { //nolint:gochecknoinits
	inputs.Add(inputName, func() inputs.Input {
		return defaultInput()
	})
}
