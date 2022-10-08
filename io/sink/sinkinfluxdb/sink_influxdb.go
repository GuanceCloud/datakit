// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

// Package sinkinfluxdb contains influxdb sink implement
package sinkinfluxdb

import (
	"fmt"
	"net/url"
	"strings"
	"time"

	client "github.com/influxdata/influxdb1-client/v2"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/dkstring"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/sinkfuncs"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io/point"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io/sink/sinkcommon"
)

const (
	creatorID = "influxdb"

	clientTypeHTTP = 1
	clientTypeUDP  = 2

	defaultTimeout = 10 * time.Second
)

var (
	_             sinkcommon.ISink = new(SinkInfluxDB)
	initSucceeded                  = false
)

type SinkInfluxDB struct {
	ID    string // sink config identity, unique, automatically generated.
	IDStr string // MD5 origin string.

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

func (s *SinkInfluxDB) Write(category string, pts []*point.Point) error {
	if !initSucceeded {
		return fmt.Errorf("not_init")
	}

	return s.writeInfluxDB(pts)
}

func (s *SinkInfluxDB) LoadConfig(mConf map[string]interface{}) error {
	if id, str, err := sinkfuncs.GetSinkCreatorID(mConf); err != nil {
		return err
	} else {
		s.ID = id
		s.IDStr = str
	}

	if host, err := dkstring.GetMapAssertString("host", mConf); err != nil {
		return err
	} else {
		hostNew, err := dkstring.CheckNotEmpty(host, "host")
		if err != nil {
			return err
		}

		if protocol, err := dkstring.GetMapAssertString("protocol", mConf); err != nil {
			return err
		} else {
			protocolNew, err := dkstring.CheckNotEmpty(protocol, "protocol")
			if err != nil {
				return err
			}

			s.addr = fmt.Sprintf("%s://%s", protocolNew, hostNew)
		}
	}

	if database, err := dkstring.GetMapAssertString("database", mConf); err != nil {
		return err
	} else {
		databaseNew, err := dkstring.CheckNotEmpty(database, "database")
		if err != nil {
			return err
		}
		s.database = databaseNew
	}

	if precision, err := dkstring.GetMapAssertString("precision", mConf); err != nil {
		return err
	} else {
		s.precision = precision
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
	} else if writeEncoding != "" {
		if writeEncoding != string(client.GzipEncoding) {
			return fmt.Errorf("not support encoding")
		}
		s.writeEncoding = writeEncoding
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
		tmpAddr := strings.ReplaceAll(s.addr, "udp://", "")
		s.addr = tmpAddr
	default:
		return fmt.Errorf("invalid addr")
	}

	initSucceeded = true
	sinkcommon.AddImpl(s)
	return nil
}

func (s *SinkInfluxDB) writeInfluxDB(pts []*point.Point) error {
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
	defer c.Close() //nolint:errcheck

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
		ps = append(ps, v.ToPoint())
	}
	bp.AddPoints(ps)

	err = c.Write(bp)
	if err != nil {
		return err
	}
	return nil
}

func (s *SinkInfluxDB) GetInfo() *sinkcommon.SinkInfo {
	return &sinkcommon.SinkInfo{
		ID:         s.ID,
		IDStr:      s.IDStr,
		CreateID:   creatorID,
		Categories: []string{datakit.SinkCategoryMetric},
	}
}

func init() { //nolint:gochecknoinits
	sinkcommon.AddCreator(creatorID, func() sinkcommon.ISink {
		return &SinkInfluxDB{}
	})
}
