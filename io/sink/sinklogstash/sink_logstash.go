// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

// Package sinklogstash contains Logstash sink implement
package sinklogstash

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"reflect"
	"strings"
	"time"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/dkstring"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/sinkfuncs"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io/point"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io/sink/sinkcommon"
)

const (
	creatorID      = "logstash"
	defaultTimeout = 10 * time.Second

	writeTypeJSON  = 0
	writeTypePlain = 1
)

var (
	_             sinkcommon.ISink = new(SinkLogstash)
	initSucceeded                  = false
)

type SinkLogstash struct {
	ID    string // sink config identity, unique, automatically generated.
	IDStr string // MD5 origin string.

	addr      string // required. eg. http://172.16.239.130:8080
	writeType int    // required. json or plain

	timeout time.Duration // option.
}

func (s *SinkLogstash) LoadConfig(mConf map[string]interface{}) error {
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

	if writeType, err := dkstring.GetMapAssertString("write_type", mConf); err != nil {
		return err
	} else {
		writeTypeNew, err := dkstring.CheckNotEmpty(writeType, "write_type")
		if err != nil {
			return err
		}

		switch writeTypeNew {
		case "json":
			s.writeType = writeTypeJSON
		case "plain":
			s.writeType = writeTypePlain
		default:
			return fmt.Errorf("not support write type: %s", writeTypeNew)
		}
	}

	initSucceeded = true
	sinkcommon.AddImpl(s)
	return nil
}

func (s *SinkLogstash) Write(category string, pts []*point.Point) error {
	if !initSucceeded {
		return fmt.Errorf("not_init")
	}

	// NOTE: if failed data need to cache, we have to create the point.Failed return
	return s.writeLogstash(pts)
}

func (s *SinkLogstash) GetInfo() *sinkcommon.SinkInfo {
	return &sinkcommon.SinkInfo{
		ID:         s.ID,
		IDStr:      s.IDStr,
		CreateID:   creatorID,
		Categories: []string{datakit.SinkCategoryLogging},
	}
}

func (s *SinkLogstash) writeLogstash(pts []*point.Point) error {
	var jsn []byte
	var err error

	switch s.writeType {
	case writeTypeJSON:
		var jps []*point.JSONPoint
		for _, v := range pts {
			jp, err := v.ToJSON()
			if err != nil {
				return err
			}
			jps = append(jps, jp)
		}

		jsn, err = json.Marshal(jps)
		if err != nil {
			return err
		}
	case writeTypePlain:
		var msgs []string
		for _, v := range pts {
			fields, err := v.ToPoint().Fields()
			if err != nil {
				return err
			}

			msg, ok := fields["message"]
			if !ok {
				return fmt.Errorf("not found plain message")
			}
			mmsg, ok := msg.(string)
			if !ok {
				return fmt.Errorf("not string: %s, %s", reflect.TypeOf(msg).Name(), reflect.TypeOf(msg).String())
			}
			msgs = append(msgs, mmsg)
		}
		jsn = []byte(strings.Join(msgs, "\n"))
	}

	client := &http.Client{
		Timeout: s.timeout,
	}

	req, err := http.NewRequest(http.MethodPut, s.addr, bytes.NewBuffer(jsn))
	if err != nil {
		return err
	}

	switch s.writeType {
	case writeTypeJSON:
		req.Header.Set("Content-Type", "application/json; charset=utf-8")
	case writeTypePlain:
	default:
	}

	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close() //nolint:errcheck

	if resp.StatusCode < 200 || resp.StatusCode >= 400 {
		return fmt.Errorf("invalid status code: %d", resp.StatusCode)
	}

	return nil
}

func init() { //nolint:gochecknoinits
	sinkcommon.AddCreator(creatorID, func() sinkcommon.ISink {
		return &SinkLogstash{}
	})
}
