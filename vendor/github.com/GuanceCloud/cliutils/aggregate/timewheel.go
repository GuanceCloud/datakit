package aggregate

import (
	"container/list"
	"sync"
	"time"

	"github.com/GuanceCloud/cliutils"
	"github.com/GuanceCloud/cliutils/point"
	"github.com/cespare/xxhash/v2"
)

var dataGroupPool = sync.Pool{
	New: func() interface{} {
		return &DataGroup{}
	},
}

// DataGroup 是时间轮中的 entry， 仅持有原始 point 二进制切片，避免长期持有展开对象图。
type DataGroup struct {
	dataType    string
	packet      *DataPacket
	FirstSeen   time.Time
	ExpiredTime int64
	slotIndex   int
	element     *list.Element
	spillKey    string // key of the spilled payload on disk (empty means in memory)
}

// Reset 清理函数.
func (dg *DataGroup) Reset() {
	dg.dataType = ""
	dg.packet = nil
	dg.FirstSeen = time.Time{}
	dg.element = nil
	dg.slotIndex = 0
	dg.ExpiredTime = 0
	dg.spillKey = ""
}

// Shard 定义分段桶.
type Shard struct {
	mu        sync.Mutex
	activeMap map[uint64]*DataGroup

	// 时间轮：本质是一个环形数组
	// 假设最大支持 3600 秒（1小时）的过期时间
	slots      [3600]*list.List
	currentPos int // 当前指针指向的槽位下标
}

// GlobalSampler 全局管理器.
type GlobalSampler struct {
	shards     []*Shard
	shardCount int
	waitTime   time.Duration // 5分钟
	configMap  map[string]*TailSamplingConfigs
	lock       sync.RWMutex

	// Oversized-packet tiered spilling (optional, disabled by default):
	// payloads larger than spillThreshold are written to the spiller while the
	// time wheel keeps metadata only; data is lazily read back when a decision
	// needs it and released right after the decision.
	spiller        PayloadSpiller
	spillThreshold int
}

type TailSamplingOutcome struct {
	Packet       *DataPacket
	SourcePacket *DataPacket
	Decision     DerivedMetricDecision
}

func NewGlobalSampler(shardCount int, waitTime time.Duration) *GlobalSampler {
	sampler := &GlobalSampler{
		shards:     make([]*Shard, shardCount),
		shardCount: shardCount,
		waitTime:   waitTime,
		configMap:  make(map[string]*TailSamplingConfigs),
	}

	for i := 0; i < shardCount; i++ {
		// 1. 初始化 Shard 结构体
		sampler.shards[i] = &Shard{
			activeMap: make(map[uint64]*DataGroup),
			// currentPos 默认为 0
		}

		// 2. 初始化时间轮的 3600 个槽位
		// 必须为每个槽位创建一个新的 list.List
		for j := 0; j < 3600; j++ {
			sampler.shards[i].slots[j] = list.New()
		}
	}

	return sampler
}

// SetPayloadSpiller enables tiered spilling of oversized payloads.
// threshold is the spill threshold in bytes; <=0 or a nil spiller disables it.
func (s *GlobalSampler) SetPayloadSpiller(spiller PayloadSpiller, threshold int) {
	if s == nil {
		return
	}

	if spiller == nil || threshold <= 0 {
		s.spiller = nil
		s.spillThreshold = 0
		return
	}

	s.spiller = spiller
	s.spillThreshold = threshold
}

// spillIfNeeded spills the payload when it exceeds the threshold and clears
// the in-memory bytes; it returns the spill key and whether it spilled.
func (s *GlobalSampler) spillIfNeeded(packet *DataPacket) (string, bool) {
	if s == nil || s.spiller == nil || packet == nil {
		return "", false
	}
	if len(packet.PointsPayload) <= s.spillThreshold {
		return "", false
	}

	key, err := s.spiller.Put(packet.PointsPayload)
	if err != nil {
		l.Errorf("spill packet payload failed: %v", err)
		return "", false
	}

	packet.PointsPayload = nil
	return key, true
}

// hydrateDataGroup reads the spilled payload back into memory.
func (s *GlobalSampler) hydrateDataGroup(dg *DataGroup) error {
	if dg == nil || dg.spillKey == "" || s.spiller == nil {
		return nil
	}

	payload, err := s.spiller.Get(dg.spillKey)
	if err != nil {
		return err
	}

	dg.packet.PointsPayload = payload
	dg.spillKey = ""
	return nil
}

// releaseSpill deletes spilled data (idempotent).
func (s *GlobalSampler) releaseSpill(key string) {
	if s == nil || s.spiller == nil || key == "" {
		return
	}

	if err := s.spiller.Delete(key); err != nil {
		l.Errorf("release spill payload failed: %v", err)
	}
}

// needsPayloadForDecision reports whether the decision for this token/type/
// group requires payload content.
// No payload is needed when all pipelines can be covered by predicates
// (or are match-all probabilistic samplers).
func (s *GlobalSampler) needsPayloadForDecision(dataType, token, groupKey string) bool {
	if s == nil {
		return true
	}

	pipelines := s.pipelinesFor(dataType, token, groupKey)
	_, ok := compilePipelinesFast(pipelines)
	return !ok
}

func tailSamplingGroupMapKey(packet *DataPacket) uint64 {
	if packet == nil {
		return 0
	}

	return tailSamplingGroupMapKeyByFields(packet.Token, packet.GroupIdHash, packet.DataType, packet.GroupKey)
}

func tailSamplingGroupMapKeyByFields(token string, groupIDHash uint64, dataType, groupKey string) uint64 {
	key := HashToken(token, groupIDHash)
	key = HashCombine(key, xxhash.Sum64(cliutils.ToUnsafeBytes(dataType)))
	key = HashCombine(key, xxhash.Sum64(cliutils.ToUnsafeBytes(groupKey)))
	return key
}

// mergePacketPayload merges the src points payload into dst.
// When both sides are uncompressed it keeps the direct-append path; when either
// side is compressed it decompresses, merges, and re-compresses.
func mergePacketPayload(dst, src *DataPacket) error {
	if dst == nil || src == nil {
		return nil
	}

	if dst.PayloadCompression == PayloadCompressionNone && src.PayloadCompression == PayloadCompressionNone {
		dst.PointsPayload = append(dst.PointsPayload, src.PointsPayload...)
		return nil
	}

	dstPayload, err := DecompressPointsPayload(dst.PointsPayload, dst.PayloadCompression)
	if err != nil {
		return err
	}
	srcPayload, err := DecompressPointsPayload(src.PointsPayload, src.PayloadCompression)
	if err != nil {
		return err
	}

	merged := make([]byte, 0, len(dstPayload)+len(srcPayload))
	merged = append(merged, dstPayload...)
	merged = append(merged, srcPayload...)

	compressed, compression, err := CompressPointsPayload(merged)
	if err != nil {
		return err
	}

	dst.PointsPayload = compressed
	dst.PayloadCompression = compression
	return nil
}

// mergeSpanPredicates OR-merges the src span predicate summary into dst.
// Boolean predicates are OR-ed; duration summaries take the maximum.
// Cross-DataKit scenario: spans of the same trace are spread across multiple
// DataKits, so the merged predicates are the union of all seen DataKits,
// consistent with walking the full merged payload at decision time.
func mergeSpanPredicates(dst, src *DataPacket) {
	if dst == nil || src == nil {
		return
	}

	dst.PredError = dst.PredError || src.PredError
	dst.PredHttpError = dst.PredHttpError || src.PredHttpError
	dst.PredBizError = dst.PredBizError || src.PredBizError
	dst.PredTraceKeep = dst.PredTraceKeep || src.PredTraceKeep

	if src.MaxSpanDurationUs > dst.MaxSpanDurationUs {
		dst.MaxSpanDurationUs = src.MaxSpanDurationUs
	}
	if src.RootDurationUs > dst.RootDurationUs {
		dst.RootDurationUs = src.RootDurationUs
	}
	if src.MaxNonrootDurationUs > dst.MaxNonrootDurationUs {
		dst.MaxNonrootDurationUs = src.MaxNonrootDurationUs
	}
}

func (s *GlobalSampler) Ingest(packet *DataPacket) {
	if s == nil || packet == nil || s.shardCount == 0 {
		return
	}

	// 1. 路由到对应的 Shard
	shard := s.shards[packet.GroupIdHash%uint64(s.shardCount)]

	shard.mu.Lock()
	defer shard.mu.Unlock()

	// 懒加载初始化
	if shard.activeMap == nil {
		shard.activeMap = make(map[uint64]*DataGroup)
		for i := 0; i < 3600; i++ {
			shard.slots[i] = list.New()
		}
	}

	// 2. 获取配置
	var ttlSec int

	switch packet.DataType {
	case point.STracing:
		traceConfig := s.GetTraceConfig(packet.Token)
		if traceConfig == nil {
			l.Errorf("no tail sampling config for token: %s, data type: %s", packet.Token, packet.DataType)
			return
		}

		ttlSec = int(traceConfig.DataTTL.Seconds())
	case point.SLogging:
		loggingConfig := s.GetLoggingConfig(packet.Token)
		if loggingConfig == nil {
			l.Errorf("no tail sampling config for token: %s, data type: %s", packet.Token, packet.DataType)
			return
		}

		ttlSec = int(loggingConfig.DataTTL.Seconds())
	case point.SRUM:
		rumConfig := s.GetRUMConfig(packet.Token)
		if rumConfig == nil {
			l.Errorf("no tail sampling config for token: %s, data type: %s", packet.Token, packet.DataType)
			return
		}
		ttlSec = int(rumConfig.DataTTL.Seconds())
	default:
		l.Errorf("unsupported data type: %s", packet.DataType)
		return
	}

	if ttlSec <= 0 {
		l.Errorf("invalid ttl for data type: %s", packet.DataType)
		return
	}
	if ttlSec >= 3600 {
		ttlSec = 3599
	}
	// 计算时间轮槽位
	expirePos := (shard.currentPos + ttlSec) % 3600

	// 创建组合键
	key := tailSamplingGroupMapKey(packet)

	pointCount := packetPointCount(packet)

	if old, exists := shard.activeMap[key]; exists {
		// --- 场景 A：老分组更新 ---
		oldSpillKey := old.spillKey
		if oldSpillKey != "" {
			if err := s.hydrateDataGroup(old); err != nil {
				// Hydrate failure: skip this merge and drop the incoming packet.
				// Continuing would leave the payload with only the new packet while
				// PointCount keeps accumulating, causing data misalignment and
				// wrong decisions.
				l.Errorf("hydrate spill payload for merge failed, skip merge: %v", err)
				return
			}
		}

		if err := mergePacketPayload(old.packet, packet); err != nil {
			l.Errorf("merge packet payload failed: %v", err)
			return
		}
		old.packet.HasError = old.packet.HasError || packet.HasError
		mergeSpanPredicates(old.packet, packet)
		old.packet.PointCount += pointCount

		if packet.TraceStartTimeUnixNano > 0 {
			if old.packet.TraceStartTimeUnixNano == 0 || packet.TraceStartTimeUnixNano < old.packet.TraceStartTimeUnixNano {
				old.packet.TraceStartTimeUnixNano = packet.TraceStartTimeUnixNano
			}
		}
		if packet.TraceEndTimeUnixNano > old.packet.TraceEndTimeUnixNano {
			old.packet.TraceEndTimeUnixNano = packet.TraceEndTimeUnixNano
		}
		if packet.MaxPointTimeUnixNano > old.packet.MaxPointTimeUnixNano {
			old.packet.MaxPointTimeUnixNano = packet.MaxPointTimeUnixNano
		}
		if old.packet.RawGroupId == "" {
			old.packet.RawGroupId = packet.RawGroupId
		}
		if packet.ConfigVersion > old.packet.ConfigVersion {
			old.packet.ConfigVersion = packet.ConfigVersion
		}
		if old.packet.Source == "" {
			old.packet.Source = packet.Source
		}

		// Re-decide whether to spill after the merge, then release the disk
		// data used before the merge.
		if key, spilled := s.spillIfNeeded(old.packet); spilled {
			old.spillKey = key
		} else {
			old.spillKey = ""
		}
		s.releaseSpill(oldSpillKey)

		// 时间轮迁移：从旧格子移到新格子
		if old.element != nil {
			shard.slots[old.slotIndex].Remove(old.element)
		}
		old.slotIndex = expirePos
		old.element = shard.slots[expirePos].PushBack(key)
		old.ExpiredTime = time.Now().Unix() + int64(ttlSec)
	} else {
		// --- 场景 B：新数据到达 ---
		// 从 Pool 中获取对象
		dg := dataGroupPool.Get().(*DataGroup) //nolint:forcetypeassert
		dg.dataType = packet.DataType
		dg.packet = packet
		if dg.packet.PointCount == 0 {
			dg.packet.PointCount = pointCount
		}
		dg.FirstSeen = time.Now()

		dg.slotIndex = expirePos
		dg.ExpiredTime = time.Now().Unix() + int64(ttlSec)
		// Spill oversized payloads: the time wheel keeps metadata only
		if key, spilled := s.spillIfNeeded(packet); spilled {
			dg.spillKey = key
		}
		// 挂载到时间轮
		dg.element = shard.slots[expirePos].PushBack(key)

		shard.activeMap[key] = dg
	}
}

// AdvanceTime 拨动时间轮，返回当前槽位到期的数据.
func (s *GlobalSampler) AdvanceTime() map[uint64]*DataGroup {
	frozenMap := make(map[uint64]*DataGroup)

	for _, shard := range s.shards {
		shard.mu.Lock()

		// 1. 指针向前跳一格
		shard.currentPos = (shard.currentPos + 1) % 3600

		// 2. 获取当前格子的链表
		currList := shard.slots[shard.currentPos]

		// 3. 遍历链表，这里面的全是这一秒该过期的
		for e := currList.Front(); e != nil; {
			next := e.Next()
			key := e.Value.(uint64) //nolint:forcetypeassert

			if dg, ok := shard.activeMap[key]; ok {
				// 提取数据
				frozenMap[key] = dg
				// 从 Map 中删除
				delete(shard.activeMap, key)
			}

			// 从链表删除
			currList.Remove(e)
			e = next
		}

		shard.mu.Unlock()
	}
	return frozenMap
}

func (s *GlobalSampler) TailSamplingOutcomes(dataGroups map[uint64]*DataGroup) map[uint64]*TailSamplingOutcome {
	outcomes := make(map[uint64]*TailSamplingOutcome, len(dataGroups))
	for key, dg := range dataGroups {
		if dg == nil || dg.packet == nil {
			outcomes[key] = &TailSamplingOutcome{Decision: DerivedMetricDecisionDropped}
			continue
		}

		sourcePacket := dg.packet

		// Spilled packets are read back only when the decision needs payload
		// content (pipelines not coverable by predicates) or when the predicate
		// summary is untrusted (all-zero: legacy data / hand-crafted). Disk data
		// is released uniformly after the decision.
		spillKey := dg.spillKey
		if spillKey != "" && (s.needsPayloadForDecision(dg.dataType, sourcePacket.Token, sourcePacket.GroupKey) ||
			!packetHasSpanPredicates(sourcePacket)) {
			if err := s.hydrateDataGroup(dg); err != nil {
				l.Errorf("hydrate spill payload for decision failed: %v", err)
			}
		}

		decision := DerivedMetricDecisionDropped
		var keptPacket *DataPacket

		token := dg.packet.Token
		groupKey := dg.packet.GroupKey

		pipelines := s.pipelinesFor(dg.dataType, token, groupKey)
		if pipelines != nil && sourcePacket != nil {
			// For spilled packets with coverable pipelines and trusted predicates,
			// decide directly from predicates to avoid disk reads; otherwise go
			// through evaluatePipelines (internal fast path or decompressed walk).
			usePredicates := spillKey != "" &&
				!s.needsPayloadForDecision(dg.dataType, token, groupKey) &&
				packetHasSpanPredicates(sourcePacket)
			if usePredicates {
				if match, packet := decideWithPredicates(sourcePacket, pipelines); match && packet != nil {
					keptPacket = packet
					decision = DerivedMetricDecisionKept
				}
			} else {
				if match, packet := evaluatePipelines(sourcePacket, pipelines); match && packet != nil {
					keptPacket = packet
					decision = DerivedMetricDecisionKept
				}
			}
		}

		// Kept packets are sent downstream and must hold the full payload:
		// hydrate here when the decision did not read it back.
		if spillKey != "" && keptPacket != nil && len(keptPacket.PointsPayload) == 0 {
			if err := s.hydrateDataGroup(dg); err != nil {
				l.Errorf("hydrate spill payload for kept packet failed: %v", err)
			}
		}
		// Decision done: release disk data (payload was either read back or
		// discarded, regardless of kept/dropped).
		s.releaseSpill(spillKey)

		outcomes[key] = &TailSamplingOutcome{
			Packet:       keptPacket,
			SourcePacket: sourcePacket,
			Decision:     decision,
		}

		dg.Reset()
		dataGroupPool.Put(dg)
	}
	return outcomes
}

// pipelinesFor returns the sampling pipeline config by data type / group dimension.
func (s *GlobalSampler) pipelinesFor(dataType, token, groupKey string) []*SamplingPipeline {
	switch dataType {
	case point.STracing:
		if cfg := s.GetTraceConfig(token); cfg != nil {
			return cfg.Pipelines
		}
	case point.SLogging:
		if cfg := s.GetLoggingConfig(token); cfg != nil {
			for _, group := range cfg.GroupDimensions {
				if group != nil && group.GroupKey == groupKey {
					return group.Pipelines
				}
			}
		}
	case point.SRUM:
		if cfg := s.GetRUMConfig(token); cfg != nil {
			for _, group := range cfg.GroupDimensions {
				if group != nil && group.GroupKey == groupKey {
					return group.Pipelines
				}
			}
		}
	}
	return nil
}

// decideWithPredicates decides pipelines via span predicates without reading
// the payload. All pipelines must be compilable; otherwise no match is returned.
func decideWithPredicates(sourcePacket *DataPacket, pipelines []*SamplingPipeline) (bool, *DataPacket) {
	exprs, ok := compilePipelinesFast(pipelines)
	if !ok {
		return false, nil
	}

	for idx, expr := range exprs {
		if expr == nil {
			continue
		}
		if expr(sourcePacket) {
			return true, pipelineMatchedPacket(sourcePacket, pipelines[idx])
		}
	}

	return false, nil
}

func (s *GlobalSampler) TailSamplingData(dataGroups map[uint64]*DataGroup) map[uint64]*DataPacket {
	outcomes := s.TailSamplingOutcomes(dataGroups)
	packets := make(map[uint64]*DataPacket)

	for key, outcome := range outcomes {
		if outcome == nil || outcome.Packet == nil {
			continue
		}
		packets[key] = outcome.Packet
	}

	return packets
}

func (s *GlobalSampler) UpdateConfig(token string, ts *TailSamplingConfigs) error {
	s.lock.Lock()
	defer s.lock.Unlock()

	if ts == nil {
		return nil
	}
	// 设置各数据类型的默认 TTL
	if ts.Tracing != nil && ts.Tracing.DataTTL == 0 {
		ts.Tracing.DataTTL = s.waitTime
	}
	if ts.Logging != nil && ts.Logging.DataTTL == 0 {
		ts.Logging.DataTTL = s.waitTime
	}
	if ts.RUM != nil && ts.RUM.DataTTL == 0 {
		ts.RUM.DataTTL = s.waitTime
	}

	if tsO, ok := s.configMap[token]; !ok {
		if err := ts.Init(); err != nil {
			return err
		}
		s.configMap[token] = ts
	} else if tsO.Version != ts.Version {
		if err := ts.Init(); err != nil {
			return err
		}
		s.configMap[token] = ts
	}

	return nil
}

func (s *GlobalSampler) GetTraceConfig(token string) *TraceTailSampling {
	s.lock.RLock()
	defer s.lock.RUnlock()
	config, ok := s.configMap[token]
	if !ok {
		return nil
	}
	return config.Tracing
}

func (s *GlobalSampler) GetLoggingConfig(token string) *LoggingTailSampling {
	s.lock.RLock()
	defer s.lock.RUnlock()
	config, ok := s.configMap[token]
	if !ok {
		return nil
	}
	return config.Logging
}

func (s *GlobalSampler) GetRUMConfig(token string) *RUMTailSampling {
	s.lock.RLock()
	defer s.lock.RUnlock()
	config, ok := s.configMap[token]
	if !ok {
		return nil
	}
	return config.RUM
}
