package memcached

import (
	"bufio"
	"bytes"
	"fmt"
	"net"
	"strconv"
	"sync"
	"time"

	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/logger"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/config"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
)

const sampleConfig = `
[[inputs.memcached]]
  ## 服务器地址，可支持多个
  servers = ["localhost:11211"]
  # unix_sockets = ["/var/run/memcached.sock"]

  ## 采集间隔
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

type inputMeasurement struct {
	name   string
	tags   map[string]string
	fields map[string]interface{}
	ts     time.Time
}

func (m inputMeasurement) LineProto() (*io.Point, error) {
	return io.MakePoint(m.name, m.tags, m.fields, m.ts)
}

func (m inputMeasurement) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name:   inputName,
		Fields: memFields,
		Tags:   map[string]interface{}{"server": inputs.NewTagInfo("The host name from which metrics are gathered")},
	}
}

type Input struct {
	Servers     []string          `toml:"servers"`
	UnixSockets []string          `toml:"unix_sockets"`
	Interval    string            `toml:"interval"`
	Tags        map[string]string `toml:"tags"`

	duration     time.Duration
	collectCache []inputs.Measurement
}

func (*Input) Catalog() string {
	return catalogName
}

func (*Input) SampleConfig() string {
	return sampleConfig
}

func (*Input) AvailableArchs() []string {
	return datakit.AllArch
}

func (i *Input) SampleMeasurement() []inputs.Measurement {
	return []inputs.Measurement{
		&inputMeasurement{},
	}
}

func (i *Input) Collect() error {
	if len(i.Servers) == 0 && len(i.UnixSockets) == 0 {
		return i.gatherServer(":11211", false)
	}
	var wg sync.WaitGroup
	for _, serverAddress := range i.Servers {
		wg.Add(1)
		go func(addr string) {
			defer wg.Done()
			if err := i.gatherServer(addr, false); err != nil {
				l.Error(err)
			}
		}(serverAddress)
	}

	for _, unixAddress := range i.UnixSockets {
		wg.Add(1)
		go func(unixAddr string) {
			defer wg.Done()
			if err := i.gatherServer(unixAddr, true); err != nil {
				l.Error(err)
			}
		}(unixAddress)
	}

	wg.Wait()

	return nil
}

func (i *Input) gatherServer(address string, unix bool) error {
	var conn net.Conn
	var err error
	if unix {
		conn, err = net.DialTimeout("unix", address, defaultTimeout)
		if err != nil {
			return err
		}
		defer conn.Close()
	} else {
		_, _, err = net.SplitHostPort(address)
		if err != nil {
			address = address + ":11211"
		}

		conn, err = net.DialTimeout("tcp", address, defaultTimeout)
		if err != nil {
			return err
		}
		defer conn.Close()
	}

	if conn == nil {
		return fmt.Errorf("failed to create net connection")
	}

	conn.SetDeadline(time.Now().Add(defaultTimeout))

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

	i.collectCache = append(i.collectCache, &inputMeasurement{
		name:   "memcached",
		fields: fields,
		tags:   tags,
		ts:     time.Now(),
	})
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
		l.Error(fmt.Errorf("invalid interval, %s", err.Error()))
	} else if duration <= 0 {
		l.Error(fmt.Errorf("invalid interval, cannot be less than zero"))
	}

	i.duration = config.ProtectedInterval(minInterval, maxInterval, duration)

	tick := time.NewTicker(i.duration)

	for {
		select {
		case <-datakit.Exit.Wait():
			l.Info("memcached exit")
			return
		case <-tick.C:
			start := time.Now()
			if err := i.Collect(); err != nil {
				io.FeedLastError(inputName, err.Error())
				l.Error(err)
			} else {
				if len(i.collectCache) > 0 {
					err := inputs.FeedMeasurement(inputName, datakit.Metric, i.collectCache, &io.Option{CollectCost: time.Since(start)})
					if err != nil {
						io.FeedLastError(inputName, err.Error())
						l.Error(err.Error())
					}
					i.collectCache = i.collectCache[:0]
				}
			}
		}
	}
}

func init() {
	inputs.Add(inputName, func() inputs.Input {
		return &Input{}
	})
}
