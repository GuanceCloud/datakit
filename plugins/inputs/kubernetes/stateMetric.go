package kubernetes

import (
	"fmt"
	"github.com/prometheus/common/expfmt"
	"strings"
	"time"
	// dto "github.com/prometheus/client_model/go"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
)

type stateMetric struct {
	name   string
	tags   map[string]string
	fields map[string]interface{}
	ts     time.Time
}

func (o *stateMetric) LineProto() (*io.Point, error) {
	return io.MakePoint(o.name, o.tags, o.fields, o.ts)
}

func (o *stateMetric) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name: objectName,
		Desc: "kubernet pod 对象",
		Tags: map[string]interface{}{
			"name":      &inputs.TagInfo{Desc: "pod name"},
			"namespace": &inputs.TagInfo{Desc: "namespace"},
			"nodeName":  &inputs.TagInfo{Desc: "node name"},
		},
		Fields: map[string]interface{}{
			"ready": &inputs.FieldInfo{
				DataType: inputs.String,
				Type:     inputs.Gauge,
				Unit:     inputs.UnknownUnit,
				Desc:     "容器ready数/总数",
			},
			"status": &inputs.FieldInfo{
				DataType: inputs.String,
				Type:     inputs.Gauge,
				Unit:     inputs.UnknownUnit,
				Desc:     "pod 状态",
			},
			"restarts": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.UnknownUnit,
				Desc:     "重启次数",
			},
			"age": &inputs.FieldInfo{
				DataType: inputs.String,
				Type:     inputs.Gauge,
				Unit:     inputs.UnknownUnit,
				Desc:     "pod存活时长",
			},
			"podIp": &inputs.FieldInfo{
				DataType: inputs.String,
				Type:     inputs.Gauge,
				Unit:     inputs.UnknownUnit,
				Desc:     "pod ip",
			},
			"createTime": &inputs.FieldInfo{
				DataType: inputs.String,
				Type:     inputs.Gauge,
				Unit:     inputs.UnknownUnit,
				Desc:     "pod 创建时间",
			},
			"label_xxx": &inputs.FieldInfo{
				DataType: inputs.String,
				Type:     inputs.Gauge,
				Unit:     inputs.UnknownUnit,
				Desc:     "pod lable",
			},
		},
	}
}

func (i *Input) collectStateMetric() error {
	resp, err := i.client.promMetrics(i.StateUrl)
	if err != nil {
		i.lastErr = err
		return err
	}
	defer resp.Body.Close()

	var parser expfmt.TextParser

	metrics, err := parser.TextToMetricFamilies(resp.Body)
	if err != nil {
		return err
	}

	for name, metric := range metrics {
		// name匹配, 提取前半部分做指标集名, 后半部分为指标名
		measurement, _ := getMeasurement(name)

		// tags
		var tags = make(map[string]string)
		var fields = make(map[string]interface{})

		var tm = time.Now()
		for _, m := range metric.GetMetric() {
			for _, lab := range m.GetLabel() {
				tags[lab.GetName()] = lab.GetValue()
			}

			tm = time.Unix(m.GetTimestampMs(), 0)

			// value
			switch metric.GetType() {
			case dto.MetricType_GAUGE:
				fields[metricName] = m.GetGauge()
			case dto.MetricType_COUNTER:
				fields[metricName] = m.GetCounter()
			}
		}

		stateM := &stateMetric{
			name:   measurement,
			tags:   tags,
			fields: fields,
			ts:     tm,
		}

		// fmt.Println("name =====>", stateM.name)
		// fmt.Println("tags =====>", stateM.tags)
		// fmt.Println("filed =====>", stateM.fields)
		// fmt.Println("ts =====>", stateM.ts)

		i.collectCache = append(i.collectCache, stateM)
	}

	return nil
}

func getMeasurement(name string) (measurement string, metric string) {
	itemArr := strings.Split(name, "_")
	if len(itemArr) > 2 {
		measurement = strings.Join(itemArr[:2], "_")
		metric = strings.Join(itemArr[2:], "_")
	}

	return measurement, metric
}
