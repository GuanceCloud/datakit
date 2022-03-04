package sinkinfluxdb

import (
	"fmt"
	"net/url"
	"time"

	client "github.com/influxdata/influxdb1-client/v2"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/dkstring"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io"
)

const (
	creatorID = "influxdb"

	clientTypeHTTP = 1
	clientTypeUDP  = 2

	defaultTimeout = 10 * time.Second
)

var (
	_             io.ISink = new(SinkInfluxDB)
	initSucceeded          = false
)

type SinkInfluxDB struct {
	ID string // required. sink config identity, unique.

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
	if id, err := dkstring.GetMapAssertString("id", mConf); err != nil {
		return err
	} else {
		s.ID = id
	}

	if addr, err := dkstring.GetMapAssertString("addr", mConf); err != nil {
		return err
	} else {
		s.addr = addr
	}

	if precision, err := dkstring.GetMapAssertString("precision", mConf); err != nil {
		return err
	} else {
		s.precision = precision
	}

	if database, err := dkstring.GetMapAssertString("database", mConf); err != nil {
		return err
	} else {
		s.database = database
	}

	if username, err := dkstring.GetMapAssertString("username", mConf); err != nil {
		return err
	} else {
		s.username = username
	}

	if password, err := dkstring.GetMapAssertString("password", mConf); err != nil {
		return err
	} else {
		s.password = password
	}

	if userAgent, err := dkstring.GetMapAssertString("user_agent", mConf); err != nil {
		return err
	} else {
		s.userAgent = userAgent
	}

	if retentionPolicy, err := dkstring.GetMapAssertString("retention_policy", mConf); err != nil {
		return err
	} else {
		s.retentionPolicy = retentionPolicy
	}

	if writeConsistency, err := dkstring.GetMapAssertString("write_consistency", mConf); err != nil {
		return err
	} else {
		s.writeConsistency = writeConsistency
	}

	if payloadSize, err := dkstring.GetMapAssertInt("payload_size", mConf); err != nil {
		return err
	} else {
		s.payloadSize = payloadSize
	}

	if writeEncoding, err := dkstring.GetMapAssertString("write_encoding", mConf); err != nil {
		return err
	} else {
		if writeEncoding != "" {
			if writeEncoding != string(client.GzipEncoding) {
				return fmt.Errorf("not support encoding")
			}
			s.writeEncoding = writeEncoding
		}
	}

	if timeout, err := dkstring.GetMapAssertString("timeout", mConf); err != nil {
		return err
	} else {
		if timeout != "" {
			td, err := time.ParseDuration(timeout)
			if err != nil {
				return err
			}
			s.timeout = td
		} else {
			s.timeout = defaultTimeout
		}
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
	io.AddImpl(s)
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

func (s *SinkInfluxDB) GetID() string {
	return s.ID
}

func init() {
	io.AddCreator(creatorID, func() io.ISink {
		return &SinkInfluxDB{}
	})
}
