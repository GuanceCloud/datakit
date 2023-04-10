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
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/config"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/goroutine"
	dkio "gitlab.jiagouyun.com/cloudcare-tools/datakit/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
)

const sampleConfig = `
[[inputs.memcached]]
  ## Servers' addresses.
  servers = ["localhost:11211"]
  # unix_sockets = ["/var/run/memcached.sock"]

  ## Set true to enable election
  election = true

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

type Input struct {
	Servers     []string          `toml:"servers"`
	UnixSockets []string          `toml:"unix_sockets"`
	Election    bool              `toml:"election"`
	Interval    string            `toml:"interval"`
	Tags        map[string]string `toml:"tags"`

	duration     time.Duration
	collectCache []*point.Point

	feeder  dkio.Feeder
	semStop *cliutils.Sem // start stop signal
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
			i.feeder.FeedLastError(inputName, err.Error())
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
			i.feeder.FeedLastError(inputName, err.Error())
			return err
		}
		defer conn.Close() //nolint:errcheck
	}

	if conn == nil {
		return fmt.Errorf("failed to create net connection")
	}

	if err := conn.SetDeadline(time.Now().Add(defaultTimeout)); err != nil {
		l.Errorf("conn.SetDeadline: %s", err)
	}

	rw := bufio.NewReadWriter(bufio.NewReader(conn), bufio.NewWriter(conn))

	if _, err := fmt.Fprint(rw, "stats\r\n"); err != nil {
		return err
	}

	if err := rw.Flush(); err != nil {
		return err
	}

	values, err := parseResponse(rw.Reader)
	if err != nil {
		return err
	}

	tags := map[string]string{"server": address}

	fields := make(map[string]interface{})

	for key := range memFields {
		if value, ok := values[key]; ok {
			if iValue, errParse := strconv.ParseInt(value, 10, 64); errParse == nil {
				fields[key] = iValue
			} else {
				fields[key] = value
			}
		}
	}

	if i.Tags != nil {
		for k, v := range i.Tags {
			tags[k] = v
		}
	}

	// i.collectCache = append(i.collectCache, &inputMeasurement{
	// 	name:   "memcached",
	// 	fields: fields,
	// 	tags:   tags,
	// 	ts:     time.Now(),
	// })
	metric := &inputMeasurement{
		name:     inputName,
		tags:     tags,
		fields:   fields,
		ts:       time.Now(),
		election: i.Election,
	}
	i.collectCache = append(i.collectCache, metric.Point())
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

const (
	maxInterval = 1 * time.Minute
	minInterval = 1 * time.Second
)

func (i *Input) Run() {
	l = logger.SLogger(inputName)

	duration, err := time.ParseDuration(i.Interval)
	if err != nil {
		l.Errorf("invalid interval, %w", err)
	} else if duration <= 0 {
		l.Error("invalid interval, cannot be less than zero")
	}

	i.duration = config.ProtectedInterval(minInterval, maxInterval, duration)

	tick := time.NewTicker(i.duration)

	for {
		start := time.Now()
		if err := i.Collect(); err != nil {
			l.Errorf("Collect: %s", err)
			i.feeder.FeedLastError(inputName, err.Error())
		}

		if len(i.collectCache) > 0 {
			// err := inputs.FeedMeasurement(inputName,
			// 	datakit.Metric,
			// 	i.collectCache,
			// 	&dkio.Option{CollectCost: time.Since(start)})
			// if err != nil {
			// 	l.Errorf("FeedMeasurement: %s", err.Error())
			// }
			if err := i.feeder.Feed(inputName, point.Metric, i.collectCache, &dkio.Option{CollectCost: time.Since(start)}); err != nil {
				l.Errorf("FeedMeasurement: %s", err.Error())
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
