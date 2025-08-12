// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

// Package exporter collect RealTime data.
package exporter

import (
	"encoding/json"
	"strconv"
	"strings"
	"time"

	"github.com/GuanceCloud/cliutils/logger"
	"github.com/GuanceCloud/cliutils/point"
)

type TriggerEvents struct {
	Hosts    []string `json:"hosts"`
	Groups   []string `json:"groups"`
	Tags     []*Tag   `json:"tags"`
	Name     string   `json:"name"`
	Clock    int64    `json:"clock"`
	Ns       int64    `json:"ns"`
	EventID  int      `json:"eventid"`
	PEventID int      `json:"p_eventid"`
	Value    int      `json:"value"`
}

func triggerEventsToPoints(lines [][]byte) []*point.Point { //nolint
	pts := make([]*point.Point, 0)
	var err error

	for _, line := range lines {
		ps := &TriggerEvents{}
		err = json.Unmarshal(line, ps)
		if err != nil {
			log.Warnf("unmarshal err=%v", err)
			continue
		}
		t := time.Unix(ps.Clock, ps.Ns)
		groups := strings.Join(ps.Groups, ",")
		// 判断是事件告警 还是恢复事件。
		if ps.PEventID != 0 {
			var kvs point.KVs
			kvs = kvs.
				Add("date", ps.Clock).
				Add("df_date_range", 0).
				Add("df_check_range_start", ps.Clock).
				Add("df_check_range_end", ps.Clock).
				Add("df_issue_duration", 1).
				Add("df_source", "custom").
				Add("df_status", ProblemStatus[0]).
				Add("df_event_id", ps.PEventID)

			opts := point.CommonLoggingOptions()
			opts = append(opts, point.WithTime(t))
			pt := point.NewPoint("zabbix_server", kvs, opts...)
			pts = append(pts, pt)
		}

		for _, host := range ps.Hosts {
			var kvs point.KVs
			for _, tag := range ps.Tags {
				kvs = kvs.AddTag(tag.Tag, tag.Value)
			}
			kvs = kvs.AddTag("host", host).
				AddTag("hostnane", host).
				AddTag("groups", groups).
				Add("date", ps.Clock).
				Add("df_date_range", 0).
				Add("df_check_range_start", ps.Clock).
				Add("df_check_range_end", ps.Clock).
				Add("df_issue_duration", 1).
				AddTag("df_source", "custom").
				Add("df_event_id", ps.EventID).
				AddTag("df_title", ps.Name).
				AddTag("df_message", ps.Name)

			opts := point.CommonLoggingOptions()
			opts = append(opts, point.WithTime(t))
			pt := point.NewPoint("zabbix_server", kvs, opts...)

			pts = append(pts, pt)
		}
	}
	return pts
}

type ItemValues struct {
	Host         string      `json:"host"`
	Groups       []string    `json:"groups"`
	Applications []string    `json:"applications"`
	ItemID       int64       `json:"itemid"`
	Name         string      `json:"name"`
	Clock        int64       `json:"clock"`
	Ns           int64       `json:"ns"`
	Timestamp    int64       `json:"timestamp"` // log only
	Source       string      `json:"source"`    // log only
	Severity     int         `json:"severity"`  // log only
	EventID      int         `json:"eventid"`   // log only
	Value        interface{} `json:"value"`
}

func ItemValuesToPoint(lines []string, tags map[string]string, log *logger.Logger, cd *CacheData) []*point.Point {
	pts := make([]*point.Point, 0)
	var err error

	for _, line := range lines {
		item := &ItemValues{}
		err = json.Unmarshal([]byte(line), item)
		if err != nil {
			log.Warnf("unmarshal err=%v ,read line =%s", err, line)
			continue
		}

		host := item.Host
		hostName := item.Host
		group := strings.Join(item.Groups, ",")
		apps := strings.Join(item.Applications, ",")
		t := time.Unix(item.Clock, item.Ns)

		var value float64
		switch v := item.Value.(type) {
		case float64:
			value = v
		case float32, int32, int, int8, int64, int16, uint, uint64, uint16, uint32:
			if f, ok := item.Value.(float64); ok {
				value = f
			}
		case string:
			if f, err := strconv.ParseFloat(v, 64); err == nil {
				value = f
			} else {
				continue
			}
		default:
			continue
		}

		var kvs point.KVs
		keyName := item.Name
		if cd != nil {
			keyName = cd.getKeyName(keyName, item.ItemID)
			iTags := cd.getTagsByItemID(item.ItemID)
			for k, v := range iTags {
				kvs = kvs.AddTag(k, v)
			}
		}
		keyName = strings.ReplaceAll(keyName, " ", "_")
		kvs = kvs.AddTag("host", host).
			AddTag("hostname", hostName).
			AddTag("groups", group).
			AddTag("applications", apps).
			AddTag("resource", "zabbix_server").
			AddTag("data_source", "history").
			Add(keyName, value)
		for k, v := range tags {
			kvs = kvs.AddTag(k, v)
		}

		pt := point.NewPoint("zabbix_server", kvs, opts...)
		pt.SetTime(t)
		pts = append(pts, pt)
	}
	return pts
}

type TrendsValue struct {
	Host         string   `json:"host"`
	Group        []string `json:"groups"`
	Applications []string `json:"applications"`
	ItemID       int      `json:"itemid"`
	Name         string   `json:"name"`
	Clock        int64    `json:"clock"`
	Count        int      `json:"count"`
	Avg          float64  `json:"avg"`
	Min          float64  `json:"min"`
	Max          float64  `json:"max"`
}

func trendsValueToPoints(lines []string, tags map[string]string, log *logger.Logger, cd *CacheData) []*point.Point {
	pts := make([]*point.Point, 0)
	var err error

	for _, line := range lines {
		trends := &TrendsValue{}
		err = json.Unmarshal([]byte(line), trends)
		if err != nil {
			log.Warnf("unmarshal err=%v", err)
			continue
		}
		host := trends.Host
		hostName := trends.Host
		group := strings.Join(trends.Group, ",")
		apps := strings.Join(trends.Applications, ",")
		t := time.Unix(trends.Clock, 0)

		// todo cacheData
		var kvs point.KVs
		kvs = kvs.AddTag("host", host).
			AddTag("hostname", hostName).
			AddTag("groups", group).
			AddTag("applications", apps).
			AddTag("resource", "zabbix_server").
			AddTag("item_id", strconv.Itoa(trends.ItemID)).
			Add(trends.Name+"_avg", trends.Avg).
			Add(trends.Name+"_max", trends.Max).
			Add(trends.Name+"_min", trends.Min)
		for k, v := range tags {
			kvs = kvs.AddTag(k, v)
		}

		pt := point.NewPoint("zabbix_server", kvs, opts...)
		pt.SetTime(t)
		pts = append(pts, pt)
	}
	return pts
}
