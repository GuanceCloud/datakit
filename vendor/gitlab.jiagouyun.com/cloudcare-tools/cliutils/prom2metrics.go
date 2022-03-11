package cliutils

import (
	"fmt"
	"io"
	"strings"
	"time"

	ifxcli "github.com/influxdata/influxdb1-client/v2"
	dto "github.com/prometheus/client_model/go"
	"github.com/prometheus/common/expfmt"
)

/// prometheus 数据转行协议 metrics
///
/// 转换规则
/// prometheus 数据是Key/Value 格式，以 `cpu_usage_user{cpu="cpu0"} 1.4112903225816156` 为例
///     - measurement:
///          1. 取 Key 字符串的第一个下划线，左右临近字符串。示例 measurement 为 `cpu_usage`
///          2. 允许手动添加 measurement 前缀，如果前缀为空字符串，则不添加。例如 measurementPrefix 为 `cloudcare`，measurement 为 `cloudcare_cpu_usage`
///          3. 前缀不会重复，不会出现 `cloudcare_cloudcare_cpu_usage` 的情况
///          4. 允许设置默认 measurement，当 tags 为空时，使用默认 measurement。默认 measurement 不允许为空字符串。
///     - tags:
///          1. 大括号中的所有键值对，全部转换成 tags。例如 `cpu=cpu0`
///     - fields:
///          1. 大括号以外的 Key/Value 转换成 fields。例如 `cpu_usage_user=1.4112903225816156`
///          2. 所有 fields 值都是 float64 类型
///     - time:
///          1. 允许设置默认时间，当无法解析 prometheus 数据的 timestamp 时，使用默认时间
///
///     如果遇到空数据，则跳过执行下一条。丢弃原有的直方图数据。具体输出，参照测试用例 prom2metrics_test.go

type prom struct {
	metricName         string
	measurement        string
	defaultMeasurement string
	metrics            []*dto.Metric
	t                  time.Time
	pts                []*ifxcli.Point
}

func PromTextToMetrics(data io.Reader, measurementPrefix, defaultMeasurement string, t time.Time) ([]*ifxcli.Point, error) {
	if defaultMeasurement == "" {
		return nil, fmt.Errorf("invalid defaultMeasurement, it is empty")
	}

	var parser expfmt.TextParser
	metrics, err := parser.TextToMetricFamilies(data)
	if err != nil {
		return nil, err
	}

	var pts []*ifxcli.Point

	for name, metric := range metrics {
		measurement := getMeasurement(name, measurementPrefix)
		p := prom{
			metricName:         name,
			measurement:        measurement,
			defaultMeasurement: defaultMeasurement,
			metrics:            metric.GetMetric(),
			t:                  t,
			pts:                []*ifxcli.Point{},
		}

		switch metric.GetType() {
		case dto.MetricType_GAUGE:
			p.gauge()
		case dto.MetricType_UNTYPED:
			p.untyped()
		case dto.MetricType_COUNTER:
			p.counter()
		case dto.MetricType_SUMMARY:
			p.summary()
		case dto.MetricType_HISTOGRAM:
			p.histogram()
		}

		pts = append(pts, p.pts...)
	}
	return pts, nil
}

func (p *prom) gauge() {
	for _, m := range p.metrics {
		p.getValue(m, m.GetGauge())
	}
}

func (p *prom) untyped() {
	for _, m := range p.metrics {
		p.getValue(m, m.GetUntyped())
	}
}

func (p *prom) counter() {
	for _, m := range p.metrics {
		p.getValue(m, m.GetCounter())
	}
}

func (p *prom) summary() {
	for _, m := range p.metrics {
		p.getCountAndSum(m, m.GetSummary())
	}
}

func (p *prom) histogram() {
	for _, m := range p.metrics {
		p.getCountAndSum(m, m.GetHistogram())
	}
}

type value interface {
	GetValue() float64
}

func (p *prom) getValue(m *dto.Metric, v value) {
	if v == nil {
		return
	}

	tags := labelToTags(m.GetLabel())
	fields := map[string]interface{}{p.metricName: v.GetValue()}

	pt, err := p.newPoint(tags, fields, m.GetTimestampMs())
	if err != nil {
		return
	}
	p.pts = append(p.pts, pt)
}

type countAndSum interface {
	GetSampleCount() uint64
	GetSampleSum() float64
}

func (p *prom) getCountAndSum(m *dto.Metric, c countAndSum) {
	if c == nil {
		return
	}

	tags := labelToTags(m.GetLabel())
	fields := map[string]interface{}{
		p.metricName + "_count": float64(c.GetSampleCount()),
		p.metricName + "_sum":   c.GetSampleSum(),
	}

	pt, err := p.newPoint(tags, fields, m.GetTimestampMs())
	if err != nil {
		return
	}
	p.pts = append(p.pts, pt)
}

func (p *prom) newPoint(tags map[string]string, fields map[string]interface{}, ts int64) (*ifxcli.Point, error) {
	if ts > 0 {
		p.t = time.Unix(0, ts*int64(time.Millisecond))
	}
	var measurement string
	if tags == nil {
		measurement = p.defaultMeasurement
	} else {
		measurement = p.measurement
	}
	return ifxcli.NewPoint(measurement, tags, fields, p.t)
}

func getMeasurement(name, measurementPrefix string) string {
	nameBlocks := strings.Split(name, "_")
	if len(nameBlocks) > 2 {
		name = strings.Join(nameBlocks[:2], "_")
	}
	if measurementPrefix != "" {
		if !strings.HasPrefix(name, measurementPrefix) {
			name = measurementPrefix + "_" + name
		}
	}
	return name
}

func labelToTags(label []*dto.LabelPair) map[string]string {
	if len(label) == 0 {
		return nil
	}
	tags := make(map[string]string, len(label))
	for _, lab := range label {
		tags[lab.GetName()] = lab.GetValue()
	}
	return tags
}
