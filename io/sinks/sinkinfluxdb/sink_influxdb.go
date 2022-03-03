package sinkinfluxdb

import (
	"fmt"
	"net/url"
	"time"

	client "github.com/influxdata/influxdb1-client/v2"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io"
)

const (
	sinkName = "influxdb"

	clientTypeHTTP = 1
	clientTypeUDP  = 2

	defaultTimeout = 10 * time.Second
)

var (
	_             io.ISink = new(SinkInfluxDB)
	initSucceeded          = false
)

type SinkInfluxDB struct {
	addr      string // required. eg. http://172.16.239.130:8086
	precision string // required.
	database  string // required.

	username      string        // option.
	password      string        // option.
	userAgent     string        // option.
	timeout       time.Duration // option.
	writeEncoding string        // option.

	retentionPolicy  string // option.
	writeConsistency string // option.

	// UDP only
	payloadSize int // option.

	// inside usage
	cliType int
}

func (s *SinkInfluxDB) Write(pts []*io.Point) error {
	if !initSucceeded {
		return fmt.Errorf("not_init")
	}

	return s.writeInfluxDB(pts)
}

func (s *SinkInfluxDB) LoadConfig(mConf map[string]interface{}) error {
	s.addr = mConf["addr"].(string)
	s.precision = mConf["precision"].(string)
	s.database = mConf["database"].(string)

	s.username = mConf["username"].(string)
	s.password = mConf["password"].(string)
	s.userAgent = mConf["user_agent"].(string)

	s.retentionPolicy = mConf["retention_policy"].(string)
	s.writeConsistency = mConf["write_consistency"].(string)

	s.payloadSize = mConf["payload_size"].(int)

	s.writeEncoding = mConf["write_encoding"].(string)
	if s.writeEncoding != "" {
		if s.writeEncoding != string(client.GzipEncoding) {
			return fmt.Errorf("not support encoding")
		}
	}

	timeoutStr := mConf["timeout"].(string)
	if timeoutStr != "" {
		td, err := time.ParseDuration(timeoutStr)
		if err != nil {
			return err
		}
		s.timeout = td
	} else {
		s.timeout = defaultTimeout
	}

	ul, err := url.Parse(s.addr)
	if err != nil {
		return err
	}
	switch ul.Scheme {
	case "http":
		s.cliType = clientTypeHTTP
	case "udp":
		s.cliType = clientTypeUDP
	default:
		return fmt.Errorf("invalid addr")
	}

	initSucceeded = true
	return nil
}

func (s *SinkInfluxDB) writeInfluxDB(pts []*io.Point) error {
	var c client.Client
	var err error

	switch s.cliType {
	case clientTypeHTTP:
		c, err = client.NewHTTPClient(client.HTTPConfig{
			Addr:               s.addr,
			Username:           s.username,
			Password:           s.password,
			UserAgent:          s.userAgent,
			Timeout:            s.timeout,
			InsecureSkipVerify: true,
			WriteEncoding:      client.ContentEncoding(s.writeEncoding),
		})
	case clientTypeUDP:
		c, err = client.NewUDPClient(client.UDPConfig{
			Addr:        s.addr,
			PayloadSize: s.payloadSize,
		})
	}
	if err != nil {
		return err
	}
	defer c.Close()

	bp, err := client.NewBatchPoints(client.BatchPointsConfig{
		Database:         s.database,
		Precision:        s.precision,
		RetentionPolicy:  s.retentionPolicy,
		WriteConsistency: s.writeConsistency,
	})
	if err != nil {
		return err
	}

	var ps []*client.Point
	for _, v := range pts {
		ps = append(ps, v.Point)
	}
	bp.AddPoints(ps)

	err = c.Write(bp)
	if err != nil {
		return err
	}
	return nil
}

func (s *SinkInfluxDB) Metrics() map[string]interface{} {
	return nil
}

func (s *SinkInfluxDB) GetName() string {
	return sinkName
}

func init() {
	io.Add(sinkName, func() io.ISink {
		return &SinkInfluxDB{}
	})
}
