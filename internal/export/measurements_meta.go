// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package export

import (
	"encoding/json"
	"fmt"
	"path/filepath"

	"github.com/GuanceCloud/cliutils/point"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/git"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs"
)

type outputMetaInfo struct {
	Version     string                         `json:"version"`
	ReleaseDate string                         `json:"release_date"`
	Doc         []string                       `json:"doc"`
	MMetaInfo   map[string]metaInfoMeasurement `json:"metrics"`
	OMetaInfo   map[string]metaInfoMeasurement `json:"objects"`
	COMetaInfo  map[string]metaInfoMeasurement `json:"custom_object"`
	OCHMetaInfo map[string]metaInfoMeasurement `json:"object_change"`
	LMetaInfo   map[string]metaInfoMeasurement `json:"logging"`
	NMetaInfo   map[string]metaInfoMeasurement `json:"networking"`
	EMetaInfo   map[string]metaInfoMeasurement `json:"keyevent"`
	TMetaInfo   map[string]metaInfoMeasurement `json:"tracing"`
	RMetaInfo   map[string]metaInfoMeasurement `json:"rum"`
	SMetaInfo   map[string]metaInfoMeasurement `json:"security"`
	PMetaInfo   map[string]metaInfoMeasurement `json:"profiling"`
	DTMetaInfo  map[string]metaInfoMeasurement `json:"dialtesting"`
}

type metaInfoMeasurement struct {
	*inputs.MeasurementInfo
	From string `json:"from"`
}

func (m *metaInfoMeasurement) String() string {
	return fmt.Sprintf("From: %s, Type: %s, Desc: %s", m.From, m.Cat, m.Desc)
}

func doExportMetaInfo(ipts map[string]inputs.Creator) ([]byte, error) {
	l.Infof("set TODO as '%s'\n", inputs.TODO)

	defaultMetaInfo := &outputMetaInfo{
		Version:     git.Version,
		ReleaseDate: git.BuildAt,
		Doc: []string{
			"本文档主要用来描述 DataKit 所采集到的各种指标集数据",
			"-----------------------------",
			"version 为当前 json 对应的 datakit 版本号",
			"release_date 为当前 json 的发布日期",
			"以下以 metrics 为例说明各个字段的意义",
			"其中 metrics.xxx 为指标集名称",
			"metrics.xxx.desc 为指标集描述",
			"metrics.xxx.type 为指标集类型，目前只有指标(M::)和对象(O::)",
			"metrics.xxx.fields 为单个指标集中的指标列表",
			"metrics.xxx.fields.xxx.data_type 为具体指标的数据类型，目前只要有 int/float/int/string 四种类型",
			"metrics.xxx.fields.xxx.unit 为具体指标的单位，可执行 cat measurements-meta.json | grep unit | sort | uniq 查看当前的单位列表（注意，这个列表中的单位可以再调整）",
			"metrics.xxx.fields.xxx.desc 为具体指标描述",
			"metrics.xxx.fields.xxx.yyy 其余字段暂时应该没用",
			"metrics.xxx.tags 为单个指标集中的标签列表",
			"-----------------------------",
			"objects.xxx 下为对象指标集，其余字段类似",
		},
		MMetaInfo:   make(map[string]metaInfoMeasurement),
		OMetaInfo:   make(map[string]metaInfoMeasurement),
		COMetaInfo:  make(map[string]metaInfoMeasurement),
		OCHMetaInfo: make(map[string]metaInfoMeasurement),
		LMetaInfo:   make(map[string]metaInfoMeasurement),
		NMetaInfo:   make(map[string]metaInfoMeasurement),
		EMetaInfo:   make(map[string]metaInfoMeasurement),
		TMetaInfo:   make(map[string]metaInfoMeasurement),
		RMetaInfo:   make(map[string]metaInfoMeasurement),
		SMetaInfo:   make(map[string]metaInfoMeasurement),
		PMetaInfo:   make(map[string]metaInfoMeasurement),
		DTMetaInfo:  make(map[string]metaInfoMeasurement),
	}

	// check if exported measurement duplicated.
	dup := func(info *inputs.MeasurementInfo, exist map[string]metaInfoMeasurement) bool {
		if info.MetaDuplicated { // duplicated allowed
			return false
		}

		if x, ok := exist[info.Name]; ok {
			l.Warnf("measurement %+#v exist(%s)", info, x.String())
			return true
		}
		return false
	}

	for inputName := range ipts {
		l.Debugf("export measurement info for %q...", inputName)

		c, ok := ipts[inputName]
		if !ok {
			return nil, fmt.Errorf("input %s not found", inputName)
		}

		input := c()
		switch ipt := input.(type) {
		// nolint:gocritic
		case inputs.InputV2:
			sampleMeasurements := ipt.SampleMeasurement()
			for _, m := range sampleMeasurements {
				if m == inputs.DefaultEmptyMeasurement {
					l.Warnf("%q got empty measurement info exported.", inputName)
					continue
				}

				measurement := m.Info()
				if measurement == nil {
					l.Warnf("ignore measurement info from %q, no measurement exported...", inputName)
					continue
				}

				if measurement.ExportSkip {
					l.Warnf("skip measurement %+#v", measurement)
					continue
				}

				if measurement.Name == "" {
					l.Warnf("ignore measurement from %s: empty measurement name: %+#v", inputName, measurement)
					return nil, fmt.Errorf("measurement name not set")
				}

				if len(measurement.Fields) == 0 {
					l.Warnf("ignore measurement from %s: no fields", inputName)
					continue
				}

				im := metaInfoMeasurement{measurement, inputName}

				l.Debugf("add measurement info for %q from %q...", measurement.Name, inputName)

				switch measurement.Cat {
				case point.Metric:
					if !dup(measurement, defaultMetaInfo.MMetaInfo) {
						defaultMetaInfo.MMetaInfo[measurement.Name] = im
					}
				case point.Network:
					if !dup(measurement, defaultMetaInfo.NMetaInfo) {
						defaultMetaInfo.NMetaInfo[measurement.Name] = im
					}
				case point.KeyEvent:
					if !dup(measurement, defaultMetaInfo.EMetaInfo) {
						defaultMetaInfo.EMetaInfo[measurement.Name] = im
					}
				case point.Object:
					if !dup(measurement, defaultMetaInfo.OMetaInfo) {
						defaultMetaInfo.OMetaInfo[measurement.Name] = im
					}
				case point.CustomObject:
					if !dup(measurement, defaultMetaInfo.COMetaInfo) {
						defaultMetaInfo.COMetaInfo[measurement.Name] = im
					}
				case point.ObjectChange:
					if !dup(measurement, defaultMetaInfo.OCHMetaInfo) {
						defaultMetaInfo.OCHMetaInfo[measurement.Name] = im
					}
				case point.Logging:
					if !dup(measurement, defaultMetaInfo.LMetaInfo) {
						defaultMetaInfo.LMetaInfo[measurement.Name] = im
					}
				case point.Tracing:
					if !dup(measurement, defaultMetaInfo.TMetaInfo) {
						defaultMetaInfo.TMetaInfo[measurement.Name] = im
					}
				case point.RUM:
					if !dup(measurement, defaultMetaInfo.RMetaInfo) {
						defaultMetaInfo.RMetaInfo[measurement.Name] = im
					}
				case point.Security:
					if !dup(measurement, defaultMetaInfo.SMetaInfo) {
						defaultMetaInfo.SMetaInfo[measurement.Name] = im
					}
				case point.Profiling:
					if !dup(measurement, defaultMetaInfo.PMetaInfo) {
						defaultMetaInfo.PMetaInfo[measurement.Name] = im
					}
				case point.DialTesting:
					if !dup(measurement, defaultMetaInfo.DTMetaInfo) {
						defaultMetaInfo.DTMetaInfo[measurement.Name] = im
					}

				case point.DynamicDWCategory, point.MetricDeprecated, point.UnknownCategory:
					l.Warnf("should not use category(%s) from %q, name: %q", measurement.Cat, inputName, measurement.Name)
					return nil, fmt.Errorf("invalid category")

				default:
					l.Warnf("no category set from %q, name: %q", inputName, measurement.Name)
					return nil, fmt.Errorf("invalid category")
				}
			}

		default:
			l.Warnf("unknown input: %s, not implements inputs.InputV2", inputName)
		}
	}

	return json.MarshalIndent(defaultMetaInfo, "", "  ")
}

//nolint:lll
func (i *Integration) exportMetaInfo(ipts map[string]inputs.Creator, lang inputs.I18n) error {
	if j, err := doExportMetaInfo(ipts); err != nil {
		return err
	} else {
		i.docs[filepath.Join(i.opt.topDir,
			"datakit",
			lang.String(),
			"measurements-meta.json")] = j
		return nil
	}
}
