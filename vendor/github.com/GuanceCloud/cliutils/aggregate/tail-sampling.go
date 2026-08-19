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

	// Fast path: when pipeline conditions can be covered by span predicates
	// (or are match-all probabilistic samplers), the decision needs no payload
	// decompression; otherwise it falls back to the decompressed walk, keeping
	// behavior identical to the legacy logic. Packets with all-zero predicates
	// (legacy data / hand-crafted) or an empty payload also fall back.
	if len(td.PointsPayload) > 0 && packetHasSpanPredicates(td) {
		// An uncompressed payload must have a decodable first point: in the legacy
		// logic a corrupted first point stops the walk and yields no match
		// (see TestTailSamplingBusinessBranches). Compressed payloads are protected
		// by the zstd frame itself (a corrupted frame fails to decompress), so no
		// pre-check is needed.
		if td.PayloadCompression == PayloadCompressionNone && !payloadHasDecodablePoint(td.PointsPayload) {
			return false, nil
		}

		if exprs, ok := compilePipelinesFast(pipelines); ok {
			for idx, expr := range exprs {
				if expr == nil {
					continue
				}
				if expr(td) {
					return true, pipelineMatchedPacket(td, pipelines[idx])
				}
			}
			return false, nil
		}
	}

	ptw := &ptWrap{}
	matchedPipelines := make([]bool, len(pipelines))

	walkErr := td.WalkRawPBPoints(func(raw []byte) bool {
		if err := ptw.Reset(raw); err != nil {
			l.Errorf("decode datapacket point failed: %v", err)
			return false
		}

		for idx, pipeline := range pipelines {
			if matchedPipelines[idx] {
				continue
			}

			matchedPipelines[idx] = pipelineMatchesPoint(ptw, pipeline)
		}

		return true
	})
	if walkErr != nil {
		l.Errorf("walk datapacket payload failed: %v", walkErr)
		return false, nil
	}

	for idx, matched := range matchedPipelines {
		if !matched {
			continue
		}

		return true, pipelineMatchedPacket(td, pipelines[idx])
	}

	return false, nil
}

// compilePipelinesFast compiles predicate decisions for all pipelines.
// It returns false overall (falling back to the decompressed walk) when any
// pipeline cannot be covered.
func compilePipelinesFast(pipelines []*SamplingPipeline) ([]spanPredicateExpr, bool) {
	exprs := make([]spanPredicateExpr, len(pipelines))
	for idx, pipeline := range pipelines {
		expr, ok := compilePipelinePredicate(pipeline)
		if !ok {
			return nil, false
		}
		exprs[idx] = expr
	}
	return exprs, true
}

// payloadHasDecodablePoint checks whether the first point in the payload is decodable.
// It matches the first-point semantics of the original walk (a corrupted first
// point stops the traversal immediately).
func payloadHasDecodablePoint(payload []byte) bool {
	var pb point.PBPoint
	found := false
	_ = point.WalkPBPointsPayload(payload, func(raw []byte) bool {
		pb.Reset()
		if err := pb.Unmarshal(raw); err != nil {
			return false
		}
		found = true
		return false
	})
	return found
}

func (sp *SamplingPipeline) isMatchAllSampling() bool {
	return sp != nil && sp.Type == PipelineTypeSampling && sp.Condition == "" && sp.Rate > 0
}

func pipelineMatchesPoint(ptw *ptWrap, pipeline *SamplingPipeline) bool {
	if ptw == nil || pipeline == nil {
		return false
	}

	if pipeline.conds == nil && !pipeline.isMatchAllSampling() {
		return false
	}

	if pipeline.conds != nil {
		if x := pipeline.conds.Eval(ptw); x < 0 {
			return false
		}
	}

	if pipeline.Type == PipelineTypeSampling && pipeline.Rate <= 0 {
		return false
	}

	return true
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

		applySpanPredicates(traceData, pt)
	}

	compressGroupedPackets(traceDatas)

	return traceDatas
}

func compressGroupedPackets(packets map[uint64]*DataPacket) {
	for _, packet := range packets {
		compressPacketPayload(packet)
	}
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
		return fmt.Errorf("%s", strings.Join(errs, "; "))
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

		applySpanPredicates(traceData, pt)
	}

	compressGroupedPackets(traceDatas)

	return traceDatas, passedThrough
}

func compressPacketPayload(packet *DataPacket) {
	if packet == nil {
		return
	}

	compressed, compression, err := CompressPointsPayload(packet.PointsPayload)
	if err != nil {
		l.Errorf("compress points payload failed: %v", err)
		return
	}

	packet.PointsPayload = compressed
	packet.PayloadCompression = compression
}

// applySpanPredicates extracts the span predicate summary of seen spans into
// a DataPacket. Predicate semantics strictly match cliutils/filter condition
// evaluation (missing fields never match), enabling decompression-free dataway
// decisions; across DataKits the Ingest merge path OR-aggregates the summary.
func applySpanPredicates(packet *DataPacket, pt *point.Point) {
	if packet == nil || pt == nil {
		return
	}

	applySpanPredicatesWith(packet, func(key string) (any, bool) {
		v := pt.Get(key)
		if v == nil {
			return nil, false
		}
		return v, true
	})
}

// applySpanPredicatesWith extracts the span predicate summary via a field
// getter, shared by the Point path (grouping) and the PBPoint path
// (dataway backfill).
func applySpanPredicatesWith(packet *DataPacket, get func(string) (any, bool)) {
	if packet == nil || get == nil {
		return
	}

	if v, ok := get("status"); ok {
		if s, ok := v.(string); ok && s == "error" {
			packet.PredError = true
		}
	}

	if v, ok := get("http_status_code"); ok {
		if s, ok := v.(string); ok && isHTTPErrorStatus(s) {
			packet.PredHttpError = true
		}
	}

	if v, ok := get("body_code"); ok {
		if s, ok := v.(string); ok && s != "0" && s != "200" {
			packet.PredBizError = true
		}
	}

	if v, ok := get("trace_keep"); ok {
		if b, ok := v.(bool); ok && b {
			packet.PredTraceKeep = true
		}
	}

	var duration int64
	switch v, ok := get("duration"); {
	case !ok:
		return
	case v == nil:
		return
	default:
		switch d := v.(type) {
		case int64:
			duration = d
		case float64:
			duration = int64(d)
		default:
			return
		}
	}
	if duration > packet.MaxSpanDurationUs {
		packet.MaxSpanDurationUs = duration
	}

	if v, ok := get("parent_id"); ok {
		pid, ok := v.(string)
		if !ok {
			return
		}

		switch pid {
		case "0":
			if duration > packet.RootDurationUs {
				packet.RootDurationUs = duration
			}
		default:
			if duration > packet.MaxNonrootDurationUs {
				packet.MaxNonrootDurationUs = duration
			}
		}
	}
}

// ComputeSpanPredicates backfills span predicates for DataPackets missing the
// summary. It runs at dataway ingestion for legacy DataKit data (no predicates),
// so the decision fast path works for any client version and removes the
// datakit/dataway version coupling. It is a no-op when predicates already exist
// (non-zero); when the payload is empty or any span fails to decode it returns
// an error without writing partial predicates (the caller keeps the
// all-zero → decompressed-walk fallback semantics).
func ComputeSpanPredicates(packet *DataPacket) error {
	if packet == nil || packetHasSpanPredicates(packet) {
		return nil
	}

	payload, err := DecompressPointsPayload(packet.PointsPayload, packet.PayloadCompression)
	if err != nil {
		return err
	}

	// Compute into a temporary struct first: any span decode failure fails the
	// whole computation, preventing partial predicates from being trusted by the
	// fast path.
	tmp := &DataPacket{}
	ptw := &ptWrap{}
	var decodeErr error
	walkErr := point.WalkPBPointsPayload(payload, func(raw []byte) bool {
		if err := ptw.Reset(raw); err != nil {
			decodeErr = err
			return false
		}
		applySpanPredicatesWith(tmp, ptw.Get)
		return true
	})
	if walkErr != nil {
		return walkErr
	}
	if decodeErr != nil {
		return decodeErr
	}

	packet.PredError = tmp.PredError
	packet.PredHttpError = tmp.PredHttpError
	packet.PredBizError = tmp.PredBizError
	packet.PredTraceKeep = tmp.PredTraceKeep
	packet.MaxSpanDurationUs = tmp.MaxSpanDurationUs
	packet.RootDurationUs = tmp.RootDurationUs
	packet.MaxNonrootDurationUs = tmp.MaxNonrootDurationUs

	return nil
}

// isHTTPErrorStatus reports whether http_status_code is 4xx/5xx.
func isHTTPErrorStatus(code string) bool {
	if len(code) != 3 {
		return false
	}
	switch code[0] {
	case '4', '5':
		return code[1] >= '0' && code[1] <= '9' && code[2] >= '0' && code[2] <= '9'
	default:
		return false
	}
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
