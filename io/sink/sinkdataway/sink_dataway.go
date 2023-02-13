// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

// Package sinkdataway contains dataway sink implement
package sinkdataway

import (
	"fmt"
	"reflect"
	"strings"
	"sync"

	"github.com/GuanceCloud/cliutils/logger"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/dkstring"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/sinkfuncs"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io/dataway"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io/filter"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io/parser"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io/point"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io/sink/sinkcommon"
)

const (
	creatorID = "dataway"
	logName   = "sink_dataway"
)

var (
	_             sinkcommon.ISink = new(SinkDataway)
	initSucceeded                  = false
	l                              = logger.DefaultSLogger(logName)
	onceInit      sync.Once
)

type SinkDataway struct {
	ID    string // sink config identity, unique, automatically generated.
	IDStr string // MD5 origin string.

	url string // required. eg. https://openway.guance.com?token=tkn_xxxxx

	proxy      string // option. eg. http://127.0.0.1:1080
	conditions parser.WhereConditions
	dw         *dataway.DataWayDefault
}

func (s *SinkDataway) Write(category string, pts []*point.Point) error {
	if !initSucceeded {
		return fmt.Errorf("not_init")
	}

	if len(s.conditions) > 0 {
		var newPts []*point.Point
		for _, pt := range pts {
			isFiltered, err := filter.CheckPointFiltered(s.conditions, category, pt)
			if err != nil {
				l.Errorf("pt.Fields: %s, ignored", err.Error())
				continue // Exclude it!
			}
			if isFiltered { // Pick those points that matched filter rules.
				// meet the condition
				newPts = append(newPts, pt)
				pt.SetWritten()
			}
		}

		if datakit.LogSinkDetail {
			if s.dw != nil && s.dw.DataWayCfg != nil && s.dw.DataWayCfg.URLs != nil {
				// Checked pointers okay, log ok to go.
				var cond string
				for _, v := range s.conditions {
					cond += v.String()
				}
				for _, v := range newPts {
					l.Infof("(sink_detail) (%v) (%s) writes: %s", s.dw.DataWayCfg.URLs, cond, v.String())
				}
			} else {
				// Log errors.
				if s.dw == nil { // nolint:gocritic
					l.Error("(sink_detail) s.dw == nil")
				} else if s.dw.DataWayCfg == nil {
					l.Error("(sink_detail) s.dw.DataWayCfg == nil")
				} else if s.dw.DataWayCfg.URLs == nil {
					l.Error("(sink_detail) s.dw.DataWayCfg.URLs == nil")
				}
			}
		}

		return s.writeDataway(category, newPts)
	} // if len(s.conditions) > 0

	return s.writeDataway(category, pts)
}

func (s *SinkDataway) LoadConfig(mConf map[string]interface{}) error {
	onceInit.Do(func() {
		l = logger.SLogger(logName)
	})

	s.conditions = make(parser.WhereConditions, 0) // clear

	if id, str, err := sinkfuncs.GetSinkCreatorID(mConf); err != nil {
		return err
	} else {
		s.ID = id
		s.IDStr = str
	}

	if url, err := getURLFromMapConfig(mConf); err != nil {
		return err
	} else {
		s.url = url
	}

	if proxy, err := dkstring.GetMapAssertString("proxy", mConf); err != nil {
		return err
	} else {
		s.proxy = proxy
	}

	if filters, ok := mConf["filters"]; ok {
		var newArr []string

		// When comes from daemonset, the type is []string.
		// When comes from datakit.conf, the type is []interface{}.
		switch filterArr := filters.(type) {
		case []string:
			newArr = filterArr
		case []interface{}:
			for _, v := range filterArr {
				if sv, ok := v.(string); ok {
					newArr = append(newArr, sv)
				} else {
					return fmt.Errorf("%#v not string: %s, %s", v, reflect.TypeOf(v).Name(), reflect.TypeOf(v).String())
				}
			}
		default:
			return fmt.Errorf("filter illegal: %s, %s", reflect.TypeOf(filters).Name(), reflect.TypeOf(filters).String())
		}

		conds, err := filter.GetConds(newArr)
		if err != nil {
			return err
		}
		s.conditions = conds
	}

	// init dataway
	dwCfg := dataway.DataWayCfg{URLs: []string{s.url}}
	if len(s.proxy) > 0 {
		dwCfg.HTTPProxy = s.proxy
	}
	dw := &dataway.DataWayDefault{}
	if err := dw.Init(&dwCfg); err != nil {
		return err
	}
	s.dw = dw

	initSucceeded = true
	sinkcommon.AddImpl(s)
	return nil
}

func getURLFromMapConfig(mConf map[string]interface{}) (string, error) {
	if uRL, err := dkstring.GetMapAssertString("url", mConf); err != nil {
		return "", err
	} else {
		urlNew, err := dkstring.CheckNotEmpty(uRL, "url")
		if err != nil {
			return "", err
		}

		urlTmp := strings.ToLower(urlNew)
		if !strings.Contains(urlTmp, "&token=") && !strings.Contains(urlTmp, "?token=") {
			// if not has token, the get from token
			if token, err := dkstring.GetMapAssertString("token", mConf); err != nil {
				return "", err
			} else {
				urlNew += "?token="
				urlNew += token
			}
		}

		return urlNew, nil
	}
}

func (s *SinkDataway) writeDataway(category string, pts []*point.Point) error {
	_, err := s.dw.Write(category, pts)
	return err
}

func (s *SinkDataway) GetInfo() *sinkcommon.SinkInfo {
	return &sinkcommon.SinkInfo{
		ID:       s.ID,
		IDStr:    s.IDStr,
		CreateID: creatorID,
		Categories: []string{
			datakit.SinkCategoryMetric,
			datakit.SinkCategoryNetwork,
			datakit.SinkCategoryKeyEvent,
			datakit.SinkCategoryObject,
			datakit.SinkCategoryCustomObject,
			datakit.SinkCategoryLogging,
			datakit.SinkCategoryTracing,
			datakit.SinkCategoryRUM,
			datakit.SinkCategorySecurity,
			datakit.SinkCategoryProfiling,
		},
	}
}

func init() { //nolint:gochecknoinits
	sinkcommon.AddCreator(creatorID, func() sinkcommon.ISink {
		return &SinkDataway{}
	})
}
