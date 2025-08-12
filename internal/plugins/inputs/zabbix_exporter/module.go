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

type ExportType int

const (
	Items ExportType = iota
	Trends
	Trigger
	Unknown
)

var modules = map[ExportType]string{
	Items:   "item",
	Trends:  "trends",
	Trigger: "trigger",
	Unknown: "unknown",
}

var ProblemStatus = map[int]string{
	0: "ok",
	1: "info",
	2: "warning",
	3: "error",
	4: "critical",
	5: "critical",
}

var opts = point.DefaultMetricOptions()

type Host struct {
	Host string `json:"host"`
	Name string `json:"name"`
}

type Tag struct {
	Tag   string
	Value string
}

type HistorySyncer struct {
	Host         *Host       `json:"host"`
	Group        []string    `json:"groups"`
	Applications []string    `json:"applications"`
	Itemid       int64       `json:"itemid"`
	Name         string      `json:"name"`
	Clock        int64       `json:"clock"`
	Ns           int64       `json:"ns"`
	Value        interface{} `json:"value"`
	Itype        int         `json:"type"`
}

func itemsToPoints(lines []string, tags map[string]string, log *logger.Logger, cd *CacheData) []*point.Point {
	pts := make([]*point.Point, 0)
	var err error

	for _, line := range lines {
		item := &HistorySyncer{}
		err = json.Unmarshal([]byte(line), item)
		if err != nil {
			log.Warnf("unmarshal err=%v ,read line =%s", err, line)
			continue
		}

		host := item.Host.Host
		hostName := item.Host.Name
		group := strings.Join(item.Group, ",")
		apps := strings.Join(item.Applications, ",")
		t := time.Unix(item.Clock, item.Ns)
		var value float64
		switch item.Itype {
		case 0, 3:
			if f, ok := item.Value.(float64); ok {
				value = f
			}
		case 1, 2, 4:
			// string 不再处理.
			continue
		}
		var kvs point.KVs
		keyName := item.Name
		if cd != nil {
			keyName = cd.getKeyName(keyName, item.Itemid)
			iTags := cd.getTagsByItemID(item.Itemid)
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

type TriggerSyncer struct { // nolint
	Clock    int64    `json:"clock"`
	Ns       int64    `json:"ns"`
	Value    int      `json:"value"`
	EventID  int      `json:"eventid"`
	PEventID int      `json:"p_eventid"`
	Name     string   `json:"name"`
	Severity int      `json:"severity"`
	Hosts    []*Host  `json:"hosts"`
	Groups   []string `json:"groups"`
	Tags     []*Tag   `json:"tags"`
}

func triggerToPoints(lines [][]byte) []*point.Point { //nolint
	pts := make([]*point.Point, 0)
	var err error

	for _, line := range lines {
		ps := &TriggerSyncer{}
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
			kvs = kvs.AddTag("host", host.Host).
				AddTag("hostnane", host.Name).
				AddTag("groups", groups).
				Add("date", ps.Clock).
				Add("df_date_range", 0).
				Add("df_check_range_start", ps.Clock).
				Add("df_check_range_end", ps.Clock).
				Add("df_issue_duration", 1).
				Add("df_sub_status", ps.Severity).
				Add("df_event_id", ps.EventID).
				AddTag("df_source", "custom").
				AddTag("df_status", ProblemStatus[ps.Severity]).
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

type TrendsSyncs struct {
	Host         *Host    `json:"host"`
	Group        []string `json:"groups"`
	Applications []string `json:"applications"`
	ItemID       int      `json:"itemid"`
	Name         string   `json:"name"`
	Clock        int64    `json:"clock"`
	Count        int      `json:"count"`
	Avg          float64  `json:"avg"`
	Min          float64  `json:"min"`
	Max          float64  `json:"max"`
	Itype        int      `json:"type"` // 0:float 1:int
}

func trendsToPoints(lines []string, tags map[string]string, log *logger.Logger, cd *CacheData) []*point.Point {
	pts := make([]*point.Point, 0)
	var err error

	for _, line := range lines {
		trends := &TrendsSyncs{}
		err = json.Unmarshal([]byte(line), trends)
		if err != nil {
			log.Warnf("unmarshal err=%v", err)
			continue
		}
		host := trends.Host.Host
		hostName := trends.Host.Name
		group := strings.Join(trends.Group, ",")
		apps := strings.Join(trends.Applications, ",")
		t := time.Unix(trends.Clock, 0)

		keyName := strings.ReplaceAll(trends.Name, " ", "_")
		// todo cacheData
		var kvs point.KVs
		kvs = kvs.AddTag("host", host).
			AddTag("hostname", hostName).
			AddTag("groups", group).
			AddTag("applications", apps).
			AddTag("resource", "zabbix_server").
			AddTag("item_id", strconv.Itoa(trends.ItemID)).
			Add(keyName+"_avg", trends.Avg).
			Add(keyName+"_max", trends.Max).
			Add(keyName+"_min", trends.Min)
		for k, v := range tags {
			kvs = kvs.AddTag(k, v)
		}

		pt := point.NewPoint("zabbix_server", kvs, opts...)
		pt.SetTime(t)
		pts = append(pts, pt)
	}
	return pts
}

func CheckModel(name string) ExportType {
	// 校验文件类型 必须是 ndjson类型。
	if !strings.HasSuffix(name, "ndjson") {
		return Unknown
	}
	strs := strings.Split(name, "-")
	if len(strs) > 0 {
		switch strs[0] {
		case "history":
			return Items
		case "trends":
			return Trends
		}
	}
	return Unknown
}
