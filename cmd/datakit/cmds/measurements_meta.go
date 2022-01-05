package cmds

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/git"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
)

var defaultMetaInfo = &OutputMetaInfo{
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

type OutputMetaInfo struct {
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

func ExportMetaInfo(path string) error {
	infof("set TODO as '%s'\n", inputs.TODO)

	f, err := os.Stat(path)
	if err == nil {
		if f.IsDir() {
			err := errors.New("the specified path is a directory")
			errorf("%s", err)
			return err
		}
	} else if err = os.MkdirAll(filepath.Dir(path), os.ModePerm); err != nil {
		return err
	}

	for k := range inputs.Inputs {
		c, ok := inputs.Inputs[k]
		if !ok {
			return fmt.Errorf("input %s not found", k)
		}

		input := c()
		switch i := input.(type) {
		case inputs.InputV2:
			sampleMeasurements := i.SampleMeasurement()
			for _, m := range sampleMeasurements {
				tmp := m.Info()
				if tmp.Name == "" {
					warnf("[W] ignore measurement from %s: empty measurement name\n", k)
					continue
				}

				if len(tmp.Fields) == 0 {
					warnf("[W] ignore measurement from %s: no fields\n", k)
					continue
				}

				switch tmp.Type {
				case "logging":
					continue

				case "object":
					if _, ok := defaultMetaInfo.ObjectMetaInfo[tmp.Name]; ok {
						if defaultMetaInfo.ObjectMetaInfo[tmp.Name].From == k {
							err := errors.New("object measurement set already exists in same collector")
							errorf("[E] original object measurement set:%s, Now measurement set:%s, measurement type:%s, measurement name:%s, error:%s\n",
								defaultMetaInfo.ObjectMetaInfo[tmp.Name].From, k, tmp.Type, tmp.Name, err)
							return err
						} else {
							err = errors.New("same measurement set in different collector")
							warnf("[E] original object measurement set: %s, current measurement set: %s, measurement type: %s, measurement name: %s, error:%s\n",
								defaultMetaInfo.ObjectMetaInfo[tmp.Name].From, k, tmp.Type, tmp.Name, err)
						}
					}

					defaultMetaInfo.ObjectMetaInfo[tmp.Name] = metaInfoMeasurement{
						Desc:   tmp.Desc,
						Type:   tmp.Type,
						Tags:   tmp.Tags,
						Fields: tmp.Fields,
						From:   k,
					}

				default: // they are metrics
					if _, ok := defaultMetaInfo.MetricMetaInfo[tmp.Name]; ok {
						if defaultMetaInfo.MetricMetaInfo[tmp.Name].From == k {
							err = errors.New("metric measurement set already exists in same collector")
							errorf("[E] original metric measurement set: %s, current measurement set: %s, measurement type: %s, measurement name: %s, error:%s\n",
								defaultMetaInfo.MetricMetaInfo[tmp.Name].From, k, tmp.Type, tmp.Name, err)
							return err
						} else {
							err = errors.New("same metric measurement set in different collector")
							warnf("[E] original metric measurement set:%s, current measurement set: %s, measurement type: %s, measurement name: %s, error:%s\n",
								defaultMetaInfo.MetricMetaInfo[tmp.Name].From, k, tmp.Type, tmp.Name, err)
						}
					}

					defaultMetaInfo.MetricMetaInfo[tmp.Name] = metaInfoMeasurement{
						Desc:   tmp.Desc,
						Type:   tmp.Type,
						Tags:   tmp.Tags,
						Fields: tmp.Fields,
						From:   k,
					}
				}
			}

		default:
			l.Warnf("unknown input: %s, not implements inputs.InputV2", k)
		}
	}

	data, _ := json.MarshalIndent(defaultMetaInfo, "", "  ")
	err = ioutil.WriteFile(path, data, os.ModePerm)
	if err != nil {
		return err
	}
	return nil
}
