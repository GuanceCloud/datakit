// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package tdengine

import (
	"time"

	"github.com/GuanceCloud/cliutils/point"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs"
)

/*
tdPlugIn:
当通过 sql 当从数据库中查询数据并不能直接获取指标时
需要通过运算或者二次查询才能计算指标时，可使用定制插件形式完成 Measurement 组装。
最终返回 []inputs.Measurement.
*/
type tdPlugIn interface {
	resToMeasurement(subMetricName string, res restResult, sql selectSQL, ipt *Input) []*point.Point
}

func metricName(subMetricName, sqlTitle string) string {
	if sqlTitle == "" {
		return subMetricName
	} else {
		return subMetricName + "_" + sqlTitle
	}
}

type tablesCount struct{}

func (*tablesCount) resToMeasurement(subMetricName string, res restResult, sql selectSQL, ipt *Input) []*point.Point {
	// 获取 ntables index
	var nodeIndex int
	for i := 0; i < len(res.ColumnMeta); i++ {
		l.Debug(res.ColumnMeta[i][0])
		if res.ColumnMeta[i][0].(string) == "ntables" {
			nodeIndex = i
			break
		}
	}

	counts := 0
	for i := 0; i < len(res.Data); i++ {
		switch res.Data[i][nodeIndex].(type) {
		case float32, float64:
			f := res.Data[i][nodeIndex].(float64)
			counts += int(f)
		case int, int64:
			c, ok := res.Data[i][nodeIndex].(int)
			if ok {
				counts += c
			}
		default:
		}
	}
	name := metricName(subMetricName, sql.title)

	msm := &Measurement{
		name: name,
		tags: map[string]string{},
		fields: map[string]interface{}{
			"table_count": counts,
		},
		ts: time.Now(),
	}
	if host := getHostTagIfNotLoopback(ipt.AdapterEndpoint); host != "" {
		msm.tags["host"] = host
	}

	opts := point.DefaultMetricOptions()
	opts = append(opts, point.WithTime(msm.ts))

	if ipt.Election {
		msm.tags = inputs.MergeTagsWrapper(msm.tags, ipt.Tagger.ElectionTags(), ipt.Tags, ipt.AdapterEndpoint)
	} else {
		msm.tags = inputs.MergeTagsWrapper(msm.tags, ipt.Tagger.HostTags(), ipt.Tags, ipt.AdapterEndpoint)
	}

	pt := point.NewPointV2(msm.name,
		append(point.NewTags(msm.tags), point.NewKVs(msm.fields)...),
		opts...)

	return []*point.Point{pt}
}

type databaseCount struct{}

func (d *databaseCount) resToMeasurement(subMetricName string, res restResult, sql selectSQL, ipt *Input) []*point.Point {
	counts := res.Rows
	name := metricName(subMetricName, sql.title)
	msm := &Measurement{
		name: name,
		tags: map[string]string{},
		fields: map[string]interface{}{
			"database_count": counts,
		},
		ts: time.Now(),
	}
	if host := getHostTagIfNotLoopback(ipt.AdapterEndpoint); host != "" {
		msm.tags["host"] = host
	}

	opts := point.DefaultMetricOptions()
	opts = append(opts, point.WithTime(msm.ts))

	if ipt.Election {
		msm.tags = inputs.MergeTagsWrapper(msm.tags, ipt.Tagger.ElectionTags(), ipt.Tags, ipt.AdapterEndpoint)
	} else {
		msm.tags = inputs.MergeTagsWrapper(msm.tags, ipt.Tagger.HostTags(), ipt.Tags, ipt.AdapterEndpoint)
	}

	pt := point.NewPointV2(msm.name,
		append(point.NewTags(msm.tags), point.NewKVs(msm.fields)...),
		opts...)

	return []*point.Point{pt}
}
