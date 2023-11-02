// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package export

import (
	"encoding/json"
	"fmt"

	cp "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/colorprint"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/git"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs"
)

type outputMetaInfo struct {
	Version        string                         `json:"version"`
	ReleaseDate    string                         `json:"release_date"`
	Doc            []string                       `json:"doc"`
	MetricMetaInfo map[string]metaInfoMeasurement `json:"metrics"`
	ObjectMetaInfo map[string]metaInfoMeasurement `json:"objects"`
}

type metaInfoMeasurement struct {
	Desc   string                 `json:"desc"`
	Type   string                 `json:"type"`
	Fields map[string]interface{} `json:"fields"`
	Tags   map[string]interface{} `json:"tags"`
	From   string                 `json:"from"`
}

//nolint:lll
func exportMetaInfo(ipts map[string]inputs.Creator) ([]byte, error) {
	cp.Infof("set TODO as '%s'\n", inputs.TODO)

	defaultMetaInfo := &outputMetaInfo{
		Version:     git.Version,
		ReleaseDate: git.BuildAt,
		Doc: []string{
			"本文档主要用来描述 DataKit 所采集到的各种指标集数据",
			"-----------------------------",
			"version 为当前 json 对应的 datakit 版本号",
			"release_date 为当前 json 的发布日期",
			"其中 metrics.xxx 为指标集名称",
			"metrics.xxx.desc 为指标集描述",
			"metrics.xxx.type 为指标集类型，目前只有指标(M::)和对象(O::)",
			"metrics.xxx.fields 为单个指标集中的指标列表",
			"metrics.xxx.fields.xxx.data_type 为具体指标的数据类型，目前只要有 int/float/int/string 四种类型",
			"metrics.xxx.fields.xxx.unit 为具体指标的单位，可执行 cat measurements-meta.json  | grep unit | sort  | uniq 查看当前的单位列表（注意，这个列表中的单位可以再调整）",
			"metrics.xxx.fields.xxx.desc 为具体指标描述",
			"metrics.xxx.fields.xxx.yyy 其余字段暂时应该没用",
			"metrics.xxx.tags 为单个指标集中的标签列表",
			"-----------------------------",
			"objects.xxx 下为对象指标集，其余字段类似",
		},
		MetricMetaInfo: make(map[string]metaInfoMeasurement),
		ObjectMetaInfo: make(map[string]metaInfoMeasurement),
	}

	for k := range ipts {
		l.Debugf("export measurement info for %q...", k)

		c, ok := ipts[k]
		if !ok {
			return nil, fmt.Errorf("input %s not found", k)
		}

		input := c()
		switch i := input.(type) {
		// nolint:gocritic
		case inputs.InputV2:
			sampleMeasurements := i.SampleMeasurement()
			for _, m := range sampleMeasurements {
				tmp := m.Info()
				if tmp == nil {
					l.Warnf("ignore measurement info from %q...", k)
					continue
				}

				if tmp.Name == "" {
					l.Warnf("ignore measurement from %s: empty measurement name: %+#v", k, tmp)
					continue
				}

				if len(tmp.Fields) == 0 {
					l.Warnf("ignore measurement from %s: no fields", k)
					continue
				}

				switch tmp.Type {
				case "logging", "tracing":
					l.Warnf("ignore %s from %s", tmp.Type, k)
					continue

				case "object":
					if _, ok := defaultMetaInfo.ObjectMetaInfo[tmp.Name]; ok {
						if defaultMetaInfo.ObjectMetaInfo[tmp.Name].From == k {
							l.Warnf("original object measurement %q, current measurement %q, measurement type: %q, measurement name: %q",
								defaultMetaInfo.ObjectMetaInfo[tmp.Name].From, k, tmp.Type, tmp.Name)
						} else {
							l.Warnf("original object measurement %q, current measurement %q, measurement type: %q, measurement name: %q",
								defaultMetaInfo.ObjectMetaInfo[tmp.Name].From, k, tmp.Type, tmp.Name)
						}
					}

					l.Debugf("add measurement info for %q from %q...", tmp.Name, k)
					defaultMetaInfo.ObjectMetaInfo[tmp.Name] = metaInfoMeasurement{
						Desc:   tmp.Desc,
						Type:   tmp.Type,
						Tags:   tmp.Tags,
						Fields: tmp.Fields,
						From:   k,
					}

				case "metric", "":
					if _, ok := defaultMetaInfo.MetricMetaInfo[tmp.Name]; ok {
						if defaultMetaInfo.MetricMetaInfo[tmp.Name].From == k {
							l.Warnf("original metric measurement %q, current measurement %q, measurement type: %q, measurement name: %q",
								defaultMetaInfo.MetricMetaInfo[tmp.Name].From, k, tmp.Type, tmp.Name)
						} else {
							l.Warnf("original metric measurement %q, current measurement %q, measurement type: %q, measurement name: %q",
								defaultMetaInfo.MetricMetaInfo[tmp.Name].From, k, tmp.Type, tmp.Name)
						}
					}

					l.Debugf("add measurement info for %q from %q...", tmp.Name, k)
					defaultMetaInfo.MetricMetaInfo[tmp.Name] = metaInfoMeasurement{
						Desc:   tmp.Desc,
						Type:   tmp.Type,
						Tags:   tmp.Tags,
						Fields: tmp.Fields,
						From:   k,
					}

				default:
					l.Warnf("error measurement type %s from %s", tmp.Type, k)
					return nil, fmt.Errorf("error measurement type %s from %s", tmp.Type, k)
				}
			}

		default:
			l.Warnf("unknown input: %s, not implements inputs.InputV2", k)
		}
	}

	return json.MarshalIndent(defaultMetaInfo, "", "  ")
}
