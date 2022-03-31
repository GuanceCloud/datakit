// Package sinklogstash contains Logstash sink implement
package sinklogstash

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/dkstring"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io/sink/sinkcommon"
)

const (
	creatorID      = "logstash"
	defaultTimeout = 10 * time.Second
)

var (
	_             sinkcommon.ISink = new(SinkLogstash)
	initSucceeded                  = false
)

type SinkLogstash struct {
	ID      string        // sink config identity, unique, automatically generated.
	addr    string        // required. eg. http://172.16.239.130:8080
	timeout time.Duration // option.
}

func (s *SinkLogstash) LoadConfig(mConf map[string]interface{}) error {
	if id, err := dkstring.GetMapMD5String(mConf); err != nil {
		return err
	} else {
		s.ID = id
	}

	if host, err := dkstring.GetMapAssertString("host", mConf); err != nil {
		return err
	} else {
		hostNew, err := dkstring.CheckNotEmpty(host, "host")
		if err != nil {
			return err
		}

		var protocolNew string
		if protocol, err := dkstring.GetMapAssertString("protocol", mConf); err != nil {
			return err
		} else {
			protocolNew, err = dkstring.CheckNotEmpty(protocol, "protocol")
			if err != nil {
				return err
			}
		}

		requestPath, err := dkstring.GetMapAssertString("request_path", mConf)
		if err != nil {
			return err
		}

		s.addr = fmt.Sprintf("%s://%s%s", protocolNew, hostNew, requestPath)
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

	initSucceeded = true
	sinkcommon.AddImpl(s)
	return nil
}

func (s *SinkLogstash) Write(pts []sinkcommon.ISinkPoint) error {
	if !initSucceeded {
		return fmt.Errorf("not_init")
	}

	return s.writeLogstash(pts)
}

func (s *SinkLogstash) GetInfo() *sinkcommon.SinkInfo {
	return &sinkcommon.SinkInfo{
		ID:         s.ID,
		CreateID:   creatorID,
		Categories: []string{datakit.SinkCategoryLogging},
	}
}

func (s *SinkLogstash) writeLogstash(pts []sinkcommon.ISinkPoint) error {
	for _, v := range pts {
		jp, err := v.ToJSON()
		if err != nil {
			return err
		}

		jsn, err := json.Marshal(jp)
		if err != nil {
			return err
		}

		client := &http.Client{
			Timeout: s.timeout,
		}

		req, err := http.NewRequest(http.MethodPut, s.addr, bytes.NewBuffer(jsn))
		if err != nil {
			return err
		}

		req.Header.Set("Content-Type", "application/json; charset=utf-8")

		resp, err := client.Do(req)
		if err != nil {
			return err
		}

		if resp.StatusCode < 200 || resp.StatusCode >= 400 {
			return fmt.Errorf("invalid status code: %d", resp.StatusCode)
		}
	}

	return nil
}

func init() { //nolint:gochecknoinits
	sinkcommon.AddCreator(creatorID, func() sinkcommon.ISink {
		return &SinkLogstash{}
	})
}
