package aggregate

import (
	"fmt"
	"hash/fnv"
	"math"
	"strconv"
	"strings"
	"time"

	fp "github.com/GuanceCloud/cliutils/filter"
	"github.com/GuanceCloud/cliutils/logger"
	"github.com/GuanceCloud/cliutils/point"
)

const (
	PipelineTypeCondition = "condition"
	PipelineTypeSampling  = "probabilistic"

	PipelineActionKeep = "keep"
	PipelineActionDrop = "drop"
	sampleRange        = 10000
)

type (
	PipelineType   string
	PipelineAction string
)

type DerivedMetric struct {
	Name      string     `toml:"name" json:"name"`
	Type      AlgoMethod `toml:"type" json:"type"`
	Condition string     `toml:"condition" json:"condition"`
	Groupby   []string   `toml:"group_by" json:"group_by"`
}

type SamplingPipeline struct {
	Name      string         `toml:"name" json:"name"`
	Type      PipelineType   `toml:"type" json:"type"`
	Condition string         `toml:"condition,omitempty" json:"condition,omitempty"`
	Action    PipelineAction `toml:"action,omitempty" json:"action,omitempty"`
	Rate      float64        `toml:"rate,omitempty" json:"rate,omitempty"`
	HashKeys  []string       `toml:"hash_keys" json:"hash_keys"`

	conds fp.WhereConditions
}

func (sp *SamplingPipeline) Apply() error {
	if sp == nil {
		return nil
	}
	if sp.Condition == "" {
		sp.conds = nil
		return nil
	}

	if ast, err := fp.GetConds(sp.Condition); err != nil {
		return err
	} else {
		sp.conds = ast
		return nil
	}
}

func (sp *SamplingPipeline) DoAction(td *DataPacket) (bool, *DataPacket) {
	if sp == nil || (sp.conds == nil && !sp.isMatchAllSampling()) {
		return false, td
	}
	if td == nil {
		return false, nil
	}

	matched, packet := evaluatePipelines(td, []*SamplingPipeline{sp})
	if !matched {
		return false, td
	}

	return true, packet
}

func evaluatePipelines(td *DataPacket, pipelines []*SamplingPipeline) (bool, *DataPacket) {
	if td == nil || len(pipelines) == 0 {
		return false, nil
	}

	ptw := &ptWrap{}
	matched := false
	var keptPacket *DataPacket

	walkErr := td.WalkRawPBPoints(func(raw []byte) bool {
		if err := ptw.Reset(raw); err != nil {
			l.Errorf("decode datapacket point failed: %v", err)
			keptPacket = nil
			matched = false
			return false
		}

		for _, pipeline := range pipelines {
			if pipeline == nil {
				continue
			}

			if pipeline.conds == nil && !pipeline.isMatchAllSampling() {
				continue
			}

			if pipeline.conds != nil {
				if x := pipeline.conds.Eval(ptw); x < 0 {
					continue
				}
			}

			if pipeline.Type == PipelineTypeSampling && pipeline.Rate <= 0 {
				continue
			}

			matched = true
			keptPacket = pipelineMatchedPacket(td, pipeline)
			return false
		}

		return true
	})
	if walkErr != nil {
		l.Errorf("walk datapacket payload failed: %v", walkErr)
		return false, nil
	}

	return matched, keptPacket
}

func (sp *SamplingPipeline) isMatchAllSampling() bool {
	return sp != nil && sp.Type == PipelineTypeSampling && sp.Condition == "" && sp.Rate > 0
}

func pipelineMatchedPacket(td *DataPacket, pipeline *SamplingPipeline) *DataPacket {
	if td == nil || pipeline == nil {
		return nil
	}

	switch pipeline.Type {
	case PipelineTypeSampling:
		if pipeline.Rate <= 0.0 {
			return td
		}

		arg := td.GroupIdHash % sampleRange
		threshold := uint64(math.Floor(pipeline.Rate * float64(sampleRange)))
		if arg < threshold {
			return td
		}

		return nil
	case PipelineTypeCondition:
		switch pipeline.Action {
		case PipelineActionDrop:
			return nil
		case PipelineActionKeep:
			return td
		default:
			return td
		}
	default:
		l.Warnf("unsupported pipeline-type %s", pipeline.Type)
		return td
	}
}

func PickTrace(source string, pts []*point.Point, version int64) map[uint64]*DataPacket {
	traceDatas := make(map[uint64]*DataPacket)
	for _, pt := range pts {
		v := pt.Get("trace_id")
		tid, ok := v.(string)
		if !ok {
			l.Errorf("invalid trace_id:%v", v)
			continue
		}

		id := hashTraceID(tid)
		traceData, ok := traceDatas[id]
		if !ok {
			traceData = &DataPacket{
				GroupIdHash:   id,
				RawGroupId:    tid,
				Token:         "", // 在pick调用处添加。
				DataType:      point.Tracing.String(),
				Source:        source,
				ConfigVersion: version,
				PointsPayload: make([]byte, 0, 256),
				GroupKey:      "",
			}
			traceDatas[id] = traceData
		}

		if !appendPointPayload(traceData, pt) {
			continue
		}

		status := pt.GetTag("status")
		if status == "error" {
			traceData.HasError = true
		}
		start, duration := getTime(pt)
		if traceData.TraceStartTimeUnixNano == 0 {
			traceData.TraceStartTimeUnixNano = start
		}
		if traceData.TraceStartTimeUnixNano > start {
			traceData.TraceStartTimeUnixNano = start
		}
		if traceData.TraceEndTimeUnixNano == 0 {
			traceData.TraceEndTimeUnixNano = start + duration
		}
		if traceData.TraceEndTimeUnixNano < start+duration {
			traceData.TraceEndTimeUnixNano = start + duration
		}
	}

	return traceDatas
}

type TraceTailSampling struct {
	DataTTL        time.Duration       `toml:"data_ttl" json:"data_ttl"`
	DerivedMetrics []*DerivedMetric    `toml:"derived_metrics" json:"derived_metrics"`
	BuiltinMetrics []*BuiltinMetricCfg `toml:"builtin_metrics" json:"builtin_metrics"`
	Pipelines      []*SamplingPipeline `toml:"sampling_pipeline" json:"pipelines"`
	Version        int64               `toml:"version" json:"version"`

	GroupKey string `toml:"group_key" json:"group_key"` // 链路固定为 "trace_id"
}

type TailSamplingConfigs struct {
	Tracing *TraceTailSampling   `toml:"trace" json:"trace"`
	Logging *LoggingTailSampling `toml:"logging" json:"logging"`
	RUM     *RUMTailSampling     `toml:"rum" json:"rum"`
	Version int64                `toml:"version" json:"version"`
}

type BuiltinMetricCfg struct {
	Name    string `toml:"name" json:"name"`
	Enabled bool   `toml:"enabled" json:"enabled"`
}

func (t *TailSamplingConfigs) ToString() string {
	if t == nil {
		return "<nil>"
	}

	var b strings.Builder
	fmt.Fprintf(&b, "TailSamplingConfigs{version=%d", t.Version)

	if t.Tracing != nil {
		fmt.Fprintf(&b, ", trace={ttl=%s, version=%d, group_key=%q, pipelines=%s, derived_metrics=%s}",
			t.Tracing.DataTTL,
			t.Tracing.Version,
			t.Tracing.GroupKey,
			pipelineNames(t.Tracing.Pipelines),
			derivedMetricNames(t.Tracing.DerivedMetrics),
		)
		fmt.Fprintf(&b, ", trace_builtin_metrics=%s", builtinMetricNames(t.Tracing.BuiltinMetrics))
	}

	if t.Logging != nil {
		fmt.Fprintf(&b, ", logging={ttl=%s, version=%d, group_dimensions=%s}",
			t.Logging.DataTTL,
			t.Logging.Version,
			loggingGroupStrings(t.Logging.GroupDimensions),
		)
		fmt.Fprintf(&b, ", logging_builtin_metrics=%s", builtinMetricNames(t.Logging.BuiltinMetrics))
	}

	if t.RUM != nil {
		fmt.Fprintf(&b, ", rum={ttl=%s, version=%d, group_dimensions=%s}",
			t.RUM.DataTTL,
			t.RUM.Version,
			rumGroupStrings(t.RUM.GroupDimensions),
		)
		fmt.Fprintf(&b, ", rum_builtin_metrics=%s", builtinMetricNames(t.RUM.BuiltinMetrics))
	}

	b.WriteString("}")
	return b.String()
}

func (t *TailSamplingConfigs) Init() error {
	var errs []string

	if t.Tracing != nil {
		if t.Tracing.DataTTL == 0 {
			t.Tracing.DataTTL = 5 * time.Minute
		}
		t.Tracing.Pipelines = normalizeSamplingPipelines(t.Tracing.Pipelines)
		if t.Tracing.GroupKey == "" {
			t.Tracing.GroupKey = "trace_id"
		} else if t.Tracing.GroupKey != "trace_id" {
			errs = append(errs, fmt.Sprintf("invalid trace group key %q: trace tail sampling only supports \"trace_id\"", t.Tracing.GroupKey))
		}
		for _, pipeline := range t.Tracing.Pipelines {
			if err := validateSamplingPipeline(pipeline); err != nil {
				errs = append(errs, fmt.Sprintf("trace pipeline: %s", err))
			}
		}
		if len(t.Tracing.DerivedMetrics) > 0 {
			errs = append(errs, "trace derived_metrics is not supported yet")
		}
		t.Tracing.BuiltinMetrics = initBuiltinMetricCfgs(t.Tracing.BuiltinMetrics, traceBuiltinMetricNames())
	}

	if t.Logging != nil {
		if t.Logging.DataTTL == 0 {
			t.Logging.DataTTL = 1 * time.Minute
		}
		for idx, group := range t.Logging.GroupDimensions {
			if group == nil {
				errs = append(errs, fmt.Sprintf("logging group_dimensions[%d] is nil", idx))
				continue
			}
			if group.GroupKey == "" {
				errs = append(errs, fmt.Sprintf("logging group_dimensions[%d] missing group_key", idx))
			}
			group.Pipelines = normalizeSamplingPipelines(group.Pipelines)
			for _, pipeline := range group.Pipelines {
				if err := validateSamplingPipeline(pipeline); err != nil {
					errs = append(errs, fmt.Sprintf("logging group %q pipeline: %s", group.GroupKey, err))
				}
			}
			if len(group.DerivedMetrics) > 0 {
				errs = append(errs, fmt.Sprintf("logging group %q derived_metrics is not supported yet", group.GroupKey))
			}
		}
		t.Logging.BuiltinMetrics = initBuiltinMetricCfgs(t.Logging.BuiltinMetrics, loggingBuiltinMetricNames())
	}

	if t.RUM != nil {
		if t.RUM.DataTTL == 0 {
			t.RUM.DataTTL = 1 * time.Minute
		}
		for idx, group := range t.RUM.GroupDimensions {
			if group == nil {
				errs = append(errs, fmt.Sprintf("rum group_dimensions[%d] is nil", idx))
				continue
			}
			if group.GroupKey == "" {
				errs = append(errs, fmt.Sprintf("rum group_dimensions[%d] missing group_key", idx))
			}
			group.Pipelines = normalizeSamplingPipelines(group.Pipelines)
			for _, pipeline := range group.Pipelines {
				if err := validateSamplingPipeline(pipeline); err != nil {
					errs = append(errs, fmt.Sprintf("rum group %q pipeline: %s", group.GroupKey, err))
				}
			}
			if len(group.DerivedMetrics) > 0 {
				errs = append(errs, fmt.Sprintf("rum group %q derived_metrics is not supported yet", group.GroupKey))
			}
		}
		t.RUM.BuiltinMetrics = initBuiltinMetricCfgs(t.RUM.BuiltinMetrics, rumBuiltinMetricNames())
	}

	if len(errs) > 0 {
		return fmt.Errorf(strings.Join(errs, "; "))
	}

	return nil
}

func normalizeSamplingPipelines(pipelines []*SamplingPipeline) []*SamplingPipeline {
	if len(pipelines) == 0 {
		return pipelines
	}

	normalized := make([]*SamplingPipeline, 0, len(pipelines))
	for _, pipeline := range pipelines {
		if pipeline != nil && pipeline.Type == PipelineTypeSampling && pipeline.Rate <= 0 {
			l.Warnf("drop probabilistic pipeline %q: rate must be greater than zero", pipeline.Name)
			continue
		}
		normalized = append(normalized, pipeline)
	}

	return normalized
}

func validateSamplingPipeline(pipeline *SamplingPipeline) error {
	if pipeline == nil {
		return fmt.Errorf("pipeline is nil")
	}

	switch pipeline.Type {
	case PipelineTypeCondition:
		if pipeline.Action != PipelineActionKeep && pipeline.Action != PipelineActionDrop {
			return fmt.Errorf("pipeline %q has invalid action %q", pipeline.Name, pipeline.Action)
		}
	case PipelineTypeSampling:
		if pipeline.Rate < 0 || pipeline.Rate > 1 {
			return fmt.Errorf("pipeline %q has invalid sampling rate %v", pipeline.Name, pipeline.Rate)
		}
	default:
		return fmt.Errorf("pipeline %q has invalid type %q", pipeline.Name, pipeline.Type)
	}

	if pipeline.Condition == "" {
		pipeline.conds = nil
		return nil
	}

	if err := pipeline.Apply(); err != nil {
		return fmt.Errorf("pipeline %q invalid condition: %w", pipeline.Name, err)
	}

	return nil
}

type LoggingTailSampling struct {
	DataTTL time.Duration `toml:"data_ttl" json:"data_ttl"`
	Version int64         `toml:"version" json:"version"`
	// 内置指标配置，默认全开
	BuiltinMetrics []*BuiltinMetricCfg `toml:"builtin_metrics" json:"builtin_metrics"`

	// 按分组维度配置（不再是全局管道）
	GroupDimensions []*LoggingGroupDimension `toml:"group_dimensions" json:"group_dimensions"`
}

type LoggingGroupDimension struct {
	// 分组键（如 user_id, order_id, session_id）
	GroupKey string `toml:"group_key" json:"group_key"`

	// 该分组维度下的采样管道
	Pipelines []*SamplingPipeline `toml:"pipelines" json:"pipelines"`

	// 该分组特有的派生指标
	DerivedMetrics []*DerivedMetric `toml:"derived_metrics" json:"derived_metrics"`
}

func (logGroup *LoggingGroupDimension) PickLogging(source string, pts []*point.Point) (map[uint64]*DataPacket, []*point.Point) {
	return pickByGroupKey(logGroup.GroupKey, source, pts, point.Logging)
}

func pickByGroupKey(groupKey string, source string, pts []*point.Point, category point.Category) (map[uint64]*DataPacket, []*point.Point) {
	traceDatas := make(map[uint64]*DataPacket)
	passedThrough := make([]*point.Point, 0)
	for _, pt := range pts {
		v := pt.Get(groupKey) // string float int64...
		if v == nil {
			passedThrough = append(passedThrough, pt)
			continue
		}

		tid := fieldToString(v)
		if tid == "" {
			passedThrough = append(passedThrough, pt)
			continue
		}
		id := hashTraceID(tid)
		traceData, ok := traceDatas[id]
		if !ok {
			traceData = &DataPacket{
				GroupIdHash:   id,
				RawGroupId:    tid,
				Token:         "",
				Source:        source,
				DataType:      category.String(),
				PointsPayload: make([]byte, 0, 256),
				GroupKey:      groupKey,
			}
			traceDatas[id] = traceData
		}

		if !appendPointPayload(traceData, pt) {
			continue
		}

		status := pt.GetTag("status")
		if status == "error" {
			traceData.HasError = true
		}
	}

	return traceDatas, passedThrough
}

// RUMTailSampling holds the RUM tail-sampling configuration.
type RUMTailSampling struct {
	DataTTL time.Duration `toml:"data_ttl" json:"data_ttl"`
	Version int64         `toml:"version" json:"version"`
	// 内置指标配置，默认全开
	BuiltinMetrics []*BuiltinMetricCfg `toml:"builtin_metrics" json:"builtin_metrics"`

	// RUM可能也有多个分组维度
	GroupDimensions []*RUMGroupDimension `toml:"group_dimensions" json:"group_dimensions"`
}

type RUMGroupDimension struct {
	GroupKey       string              `toml:"group_key" json:"group_key"` // session_id, user_id, page_id
	Pipelines      []*SamplingPipeline `toml:"pipelines" json:"pipelines"`
	DerivedMetrics []*DerivedMetric    `toml:"derived_metrics" json:"derived_metrics"`
}

func (rumGroup *RUMGroupDimension) PickRUM(source string, pts []*point.Point) (map[uint64]*DataPacket, []*point.Point) {
	return pickByGroupKey(rumGroup.GroupKey, source, pts, point.RUM)
}

func SetLogging(log *logger.Logger) {
	l = log
}

// hashTraceID 将字符串 TraceID 转换为 uint64.
func hashTraceID(s string) uint64 {
	h := fnv.New64a()
	_, _ = h.Write([]byte(s))
	return h.Sum64()
}

func fieldToString(field any) string {
	switch x := field.(type) {
	case float64:
		return strconv.FormatFloat(x, 'f', -1, 64)
	case int64:
		return strconv.FormatInt(x, 10)
	case uint64:
		return strconv.FormatUint(x, 10)
	case string:
		return x
	case []byte:
		return string(x)
	case bool:
		return strconv.FormatBool(x)
	default: // other types are ignored
		return ""
	}
}

func appendPointPayload(packet *DataPacket, pt *point.Point) bool {
	if packet == nil || pt == nil {
		return false
	}

	packet.PointsPayload = point.AppendPointToPBPointsPayload(packet.PointsPayload, pt)
	packet.PointCount++

	if ts := pt.Time().UnixNano(); ts > packet.MaxPointTimeUnixNano {
		packet.MaxPointTimeUnixNano = ts
	}

	return true
}

func pipelineNames(pipelines []*SamplingPipeline) string {
	if len(pipelines) == 0 {
		return "[]"
	}

	items := make([]string, 0, len(pipelines))
	for _, pipeline := range pipelines {
		if pipeline == nil {
			items = append(items, "<nil>")
			continue
		}
		items = append(items, fmt.Sprintf("{name=%q,type=%q,condition=%q,action=%q,rate=%v,hash_keys=%v}",
			pipeline.Name, pipeline.Type, pipeline.Condition, pipeline.Action, pipeline.Rate, pipeline.HashKeys))
	}

	return "[" + strings.Join(items, ", ") + "]"
}

func derivedMetricNames(metrics []*DerivedMetric) string {
	if len(metrics) == 0 {
		return "[]"
	}

	items := make([]string, 0, len(metrics))
	for _, metric := range metrics {
		if metric == nil {
			items = append(items, "<nil>")
			continue
		}
		items = append(items, fmt.Sprintf("{name=%q,type=%q,condition=%q,group_by=%v}",
			metric.Name, metric.Type.String(), metric.Condition, metric.Groupby))
	}

	return "[" + strings.Join(items, ", ") + "]"
}

func builtinMetricNames(metrics []*BuiltinMetricCfg) string {
	if len(metrics) == 0 {
		return "[]"
	}

	items := make([]string, 0, len(metrics))
	for _, metric := range metrics {
		if metric == nil {
			items = append(items, "<nil>")
			continue
		}
		items = append(items, fmt.Sprintf("{name=%q,enabled=%v}", metric.Name, metric.Enabled))
	}

	return "[" + strings.Join(items, ", ") + "]"
}

func initBuiltinMetricCfgs(cfgs []*BuiltinMetricCfg, defaults []string) []*BuiltinMetricCfg {
	if len(defaults) == 0 {
		return cfgs
	}

	if len(cfgs) == 0 {
		res := make([]*BuiltinMetricCfg, 0, len(defaults))
		for _, name := range defaults {
			res = append(res, &BuiltinMetricCfg{
				Name:    name,
				Enabled: true,
			})
		}
		return res
	}

	set := make(map[string]struct{}, len(cfgs))
	for _, cfg := range cfgs {
		if cfg == nil || cfg.Name == "" {
			continue
		}
		set[cfg.Name] = struct{}{}
	}

	for _, name := range defaults {
		if _, ok := set[name]; ok {
			continue
		}
		cfgs = append(cfgs, &BuiltinMetricCfg{
			Name:    name,
			Enabled: true,
		})
	}

	return cfgs
}

func loggingGroupStrings(groups []*LoggingGroupDimension) string {
	if len(groups) == 0 {
		return "[]"
	}

	items := make([]string, 0, len(groups))
	for _, group := range groups {
		if group == nil {
			items = append(items, "<nil>")
			continue
		}
		items = append(items, fmt.Sprintf("{group_key=%q,pipelines=%s,derived_metrics=%s}",
			group.GroupKey, pipelineNames(group.Pipelines), derivedMetricNames(group.DerivedMetrics)))
	}

	return "[" + strings.Join(items, ", ") + "]"
}

func rumGroupStrings(groups []*RUMGroupDimension) string {
	if len(groups) == 0 {
		return "[]"
	}

	items := make([]string, 0, len(groups))
	for _, group := range groups {
		if group == nil {
			items = append(items, "<nil>")
			continue
		}
		items = append(items, fmt.Sprintf("{group_key=%q,pipelines=%s,derived_metrics=%s}",
			group.GroupKey, pipelineNames(group.Pipelines), derivedMetricNames(group.DerivedMetrics)))
	}

	return "[" + strings.Join(items, ", ") + "]"
}
