package cmds

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
)

var defaultMetaInfo = &OutputMetaInfo{
	MetricMetaInfo: make(map[string]metaInfoMeasurement),
	ObjectMetaInfo: make(map[string]metaInfoMeasurement),
}

type OutputMetaInfo struct {
	MetricMetaInfo map[string]metaInfoMeasurement `json:"metric_metaInfo"`
	ObjectMetaInfo map[string]metaInfoMeasurement `json:"object_metaInfo"`
}

type metaInfoMeasurement struct {
	Desc   string                 `json:"desc"`
	Type   string                 `json:"type"`
	Fields map[string]interface{} `json:"fields"`
	Tags   map[string]interface{} `json:"tags"`
	From   string                 `json:"from"`
}

func ExportMetaInfo(path string) error {
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

				switch tmp.Type {
				case "logging":
					continue
				case "object":
					if _, ok := defaultMetaInfo.ObjectMetaInfo[tmp.Name]; ok {
						if defaultMetaInfo.ObjectMetaInfo[tmp.Name].From == k {
							err := errors.New("measurement set already exists in same collector")
							errorf("Original measurement set:%s, Now measurement set:%s, measurement type:%s, measurement name:%s, error:%s\n",
								defaultMetaInfo.ObjectMetaInfo[tmp.Name].From, k, tmp.Type, tmp.Name, err)
							return err
						} else {
							err = errors.New("same measurement set in different collector")
							warnf("Original measurement set:%s,Now measurement set:%s, measurement type:%s, measurement name:%s, error:%s\n",
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

				default:
					if _, ok := defaultMetaInfo.MetricMetaInfo[tmp.Name]; ok {
						if defaultMetaInfo.MetricMetaInfo[tmp.Name].From == k {
							err = errors.New("measurement set already exists in same collector")
							errorf("Original measurement set:%s,Now measurement set:%s, measurement type:%s, measurement name:%s, error:%s\n",
								defaultMetaInfo.MetricMetaInfo[tmp.Name].From, k, tmp.Type, tmp.Name, err)
							return err
						} else {
							err = errors.New("same measurement set in different collector")
							warnf("Original measurement set:%s, Now measurement set:%s, measurement type:%s, measurement name:%s, error:%s\n",
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
			l.Warnf("incomplete input: %s", k)
		}
	}

	data, _ := json.Marshal(defaultMetaInfo)
	err = ioutil.WriteFile(path, data, os.ModePerm)
	if err != nil {
		return err
	}
	return nil
}
