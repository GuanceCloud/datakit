// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

// Package cache pinpoint data cache.
package cache

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/datakit"

	"github.com/GuanceCloud/cliutils/logger"
	ppv1 "github.com/GuanceCloud/tracing-protos/pinpoint-gen-go/v1"
	"google.golang.org/grpc/metadata"
	"google.golang.org/protobuf/proto"
)

var log = logger.DefaultSLogger("pinpoint")

var (
	cacheName               = "resource.json"
	defaultResourceCacheDir = filepath.Join(datakit.CacheDir, "pp_resources")
	cacheFullPath           = filepath.Join(defaultResourceCacheDir, cacheName)
	eventDefaultTime        = time.Minute
)

type MetaData struct {
	mu           sync.RWMutex
	PSqlDatas    map[int32]*ppv1.PSqlMetaData    `json:"p_sql_datas"`
	PApiDatas    map[int32]*ppv1.PApiMetaData    `json:"p_api_datas"`
	PStringDatas map[int32]*ppv1.PStringMetaData `json:"p_string_datas"`
	connectStat  bool
}

func NewMetaData() *MetaData {
	return &MetaData{
		PSqlDatas:    make(map[int32]*ppv1.PSqlMetaData),
		PApiDatas:    make(map[int32]*ppv1.PApiMetaData),
		PStringDatas: make(map[int32]*ppv1.PStringMetaData),
		connectStat:  true,
	}
}

type EventItem struct {
	events     []*ppv1.PSpanEvent
	spanChunks []*ppv1.PSpanChunk
	expiredAt  time.Time
}

// SpanItem 当收到 span 中 parentId 不为 -1 则证明有父级 span
// 缓存当前span和mate信息，等最终的 parentId == -1 时，依次取出追加到当前span。
// 如果过期，应当组装成顶层 span 立即发送。
type SpanItem struct {
	Span      *ppv1.PSpan
	Meta      metadata.MD
	expiredAt time.Time
}

type ProcessSpan func(span *ppv1.PSpan, md metadata.MD)

// AgentCache pinpoint cache，must be not nil !!!
// metaData 缓存：API SQL 缓存表，dk重启需要从文件读取。定期删除链接不在的agent防止oom。
// agentInfo 缓存：按 agent id 区分，用于填充链路和指标中的tag：hostname ip port 等。
// span   缓存：将链路父子级关系归拢的重要一步。
// event  缓存：span到来之后，将作为span补充的chunk event缓存的取出填充之。
type AgentCache struct {
	reWriteFile bool                 // meta to
	metaLock    sync.RWMutex         // meta lock
	Metas       map[string]*MetaData // key is agentID

	mu          sync.RWMutex
	Agents      map[string]*ppv1.PAgentInfo // key is agentID
	EventData   map[string]*EventItem       // key is traceID
	SpanData    map[int64]*SpanItem         // key spanId
	ProcessFunc ProcessSpan                 // 回调函数
	timeout     time.Duration
}

func NewAgentCache(pf ProcessSpan) *AgentCache {
	log = logger.SLogger("pinpoint-cache")
	ac := &AgentCache{
		Agents:      make(map[string]*ppv1.PAgentInfo),
		Metas:       make(map[string]*MetaData),
		EventData:   make(map[string]*EventItem),
		SpanData:    make(map[int64]*SpanItem),
		ProcessFunc: pf,
		timeout:     time.Minute,
	}

	ac.Metas = readFromFile()
	go ac.deleteExpiredSpan()
	go ac.deleteExpiredEvent()
	go ac.writeToFile()
	go ac.deleteMeta()
	return ac
}

func (ac *AgentCache) GetSpan(key int64) (*SpanItem, bool) {
	ac.mu.Lock()
	defer ac.mu.Unlock()

	spanItme, ok := ac.SpanData[key]
	if !ok {
		return nil, false
	}

	if time.Now().After(spanItme.expiredAt) {
		return nil, false
	}

	delete(ac.SpanData, key)
	return spanItme, true
}

func (ac *AgentCache) SetSpan(key int64, span *ppv1.PSpan, meta metadata.MD, duration time.Duration) {
	if duration == 0 {
		duration = eventDefaultTime
	}
	ac.mu.Lock()
	defer ac.mu.Unlock()
	if item, ok := ac.SpanData[key]; ok {
		// item.spans = append(item.spans, span)
		log.Debugf("itme is exist %v", item)
	} else {
		ac.SpanData[key] = &SpanItem{
			Span:      span,
			Meta:      meta,
			expiredAt: time.Now().Add(duration),
		}
	}
}

func (ac *AgentCache) deleteExpiredSpan() {
	ticker := time.NewTicker(ac.timeout)
	for range ticker.C {
		for key, item := range ac.SpanData {
			if time.Now().After(item.expiredAt) {
				ac.ProcessFunc(item.Span, item.Meta)
				ac.mu.Lock()
				delete(ac.SpanData, key)
				ac.mu.Unlock()
			}
		}
	}
}

type PSpanEventList []*ppv1.PSpanEvent

func (x PSpanEventList) Len() int { return len(x) }

func (x PSpanEventList) Less(i, j int) bool {
	return x[i].Sequence < x[j].Sequence
}

func (x PSpanEventList) Swap(i, j int) { x[i], x[j] = x[j], x[i] }

// GetEvent key is traceId.
func (ac *AgentCache) GetEvent(key string, spanStartTime int64) ([]*ppv1.PSpanEvent, bool) {
	ac.mu.Lock()
	defer ac.mu.Unlock()

	item, ok := ac.EventData[key]
	if !ok {
		return nil, false
	}

	if time.Now().After(item.expiredAt) {
		return nil, false
	}
	for _, chunk := range ac.EventData[key].spanChunks {
		eventTimeEl := int64(0)
		if keyTime := chunk.GetKeyTime(); keyTime > 0 {
			eventTimeEl = keyTime - (spanStartTime / int64(time.Millisecond)) // 单位毫秒
		}

		for i := 0; i < len(chunk.GetSpanEvent())-1; i++ {
			chunk.SpanEvent[i].StartElapsed += int32(eventTimeEl)
			item.events = append(item.events, chunk.SpanEvent[i])
		}
	}
	list := PSpanEventList(item.events)
	sort.Sort(list)
	delete(ac.EventData, key)
	return list, true
}

func (ac *AgentCache) SetEvent(key string, events []*ppv1.PSpanEvent, duration time.Duration) {
	if duration == 0 {
		duration = eventDefaultTime
	}
	ac.mu.Lock()
	defer ac.mu.Unlock()
	if _, ok := ac.EventData[key]; ok {
		ac.EventData[key].events = append(ac.EventData[key].events, events...)
	} else {
		ac.EventData[key] = &EventItem{
			events:     events,
			spanChunks: make([]*ppv1.PSpanChunk, 0),
			expiredAt:  time.Now().Add(duration),
		}
	}
}

func (ac *AgentCache) SetSpanChunk(key string, chunk *ppv1.PSpanChunk, duration time.Duration) {
	if duration == 0 {
		duration = eventDefaultTime
	}
	ac.mu.Lock()
	defer ac.mu.Unlock()

	if _, ok := ac.EventData[key]; ok {
		ac.EventData[key].spanChunks = append(ac.EventData[key].spanChunks, chunk)
	} else {
		ac.EventData[key] = &EventItem{
			events:     make([]*ppv1.PSpanEvent, 0),
			spanChunks: []*ppv1.PSpanChunk{chunk},
			expiredAt:  time.Now().Add(duration),
		}
	}
}

func (ac *AgentCache) deleteExpiredEvent() {
	ticker := time.NewTicker(ac.timeout)
	for range ticker.C {
		ac.mu.Lock()
		for key, item := range ac.EventData {
			if time.Now().After(item.expiredAt) {
				delete(ac.EventData, key)
			}
		}
		ac.mu.Unlock()
	}
}

func (ac *AgentCache) StoreMeta(agentID string, meta proto.Message) {
	ac.metaLock.Lock()
	defer ac.metaLock.Unlock()
	switch ps := meta.(type) {
	case *ppv1.PSqlMetaData:
		log.Debugf("store sql ip=%s id=%d v=%s", agentID, ps.SqlId, ps.Sql)
		if _, ok := ac.Metas[agentID]; ok {
			ac.Metas[agentID].mu.Lock()
			defer ac.Metas[agentID].mu.Unlock()
			ac.Metas[agentID].PSqlDatas[ps.SqlId] = ps
		} else {
			md := NewMetaData()
			md.PSqlDatas[ps.SqlId] = ps
			ac.Metas[agentID] = md
		}

		ac.reWriteFile = true
	case *ppv1.PApiMetaData:
		log.Debugf("store api ip=%s id=%d  v=%v", agentID, ps.ApiId, ps.ApiInfo)
		ps.ApiInfo = strings.ReplaceAll(ps.ApiInfo, "\n", " ")
		if _, ok := ac.Metas[agentID]; ok {
			ac.Metas[agentID].PApiDatas[ps.ApiId] = ps
		} else {
			md := NewMetaData()
			md.PApiDatas[ps.ApiId] = ps
			ac.Metas[agentID] = md
		}

		ac.reWriteFile = true
	case *ppv1.PStringMetaData:
		log.Debugf("store string id=%d v =%s", ps.StringId, ps.StringValue)
		if _, ok := ac.Metas[agentID]; ok {
			ac.Metas[agentID].PStringDatas[ps.StringId] = ps
		} else {
			md := NewMetaData()
			md.PStringDatas[ps.StringId] = ps
			ac.Metas[agentID] = md
		}

		ac.reWriteFile = true
	default:
		log.Infof("unknown type %v", meta)
	}
}

func (ac *AgentCache) SyncMetaConn(md metadata.MD, isConn bool) {
	agentID := ""
	if vals := md.Get("agentid"); len(vals) > 0 {
		agentID = vals[0]
	}
	for key := range ac.Metas {
		if agentID == key {
			ac.metaLock.RLock()
			ac.Metas[agentID].connectStat = isConn
			ac.metaLock.RUnlock()
		}
	}
}

func (ac *AgentCache) deleteMeta() {
	// 5分钟：每次agent发送 agentInfo 信息都会刷新链接信息
	// 如果长时间没有发送，就应该删除，而不是 链接断开就删
	// 这是因为 agentID 可能多服务公用。
	ticker := time.NewTicker(time.Minute * 3)
	for range ticker.C {
		for agentID, meta := range ac.Metas {
			if !meta.connectStat {
				ac.metaLock.Lock()
				log.Infof("delete agent id = %s MetaData", agentID)
				delete(ac.Metas, agentID)
				ac.reWriteFile = true
				ac.metaLock.Unlock()
			}
		}
	}
}

func readFromFile() map[string]*MetaData {
	// 读取 meta
	metas := make(map[string]*MetaData)
	_, err := os.Stat(cacheFullPath)
	if err != nil {
		if os.IsNotExist(err) {
			_ = os.MkdirAll(defaultResourceCacheDir, 0o600)
		}
		return metas
	}

	bts, err := os.ReadFile(filepath.Clean(cacheFullPath))
	if err != nil {
		log.Warnf("readFromFile err=%v", err)
		return metas
	}
	err = json.Unmarshal(bts, &metas)
	if err != nil {
		log.Warnf("json unMarshal err=%v", err)
	}

	return metas
}

func (ac *AgentCache) writeToFile() {
	ticker := time.NewTicker(time.Second * 31)
	for range ticker.C {
		log.Debugf("mc= %+v", ac.reWriteFile)
		if ac.reWriteFile {
			f, err := os.OpenFile(filepath.Clean(cacheFullPath), os.O_CREATE|os.O_RDWR|os.O_TRUNC, 0o600)
			if err != nil {
				log.Infof("openFile err=%v", err)
				continue
			}
			bts, err := json.MarshalIndent(ac.Metas, "", "	")
			if err != nil {
				log.Errorf("err = %v", err)
			} else {
				if _, err = f.Write(bts); err != nil {
					log.Errorf("write err=%v", err)
				}
			}
			ac.reWriteFile = false
			_ = f.Sync()
			_ = f.Close()
		}
	}
}

func (ac *AgentCache) FindAPIInfo(agentID string, apiID int32) (res, opt string, find bool) {
	ac.metaLock.RLock()
	defer ac.metaLock.RUnlock()
	metas, ok := ac.Metas[agentID]
	if !ok {
		return
	}

	if data, ok := metas.PApiDatas[apiID]; ok {
		res = data.ApiInfo
		opt = fmt.Sprintf("id:%d line:%d %s:%s", apiID, data.Line, data.Location, data.ApiInfo)
		find = true
		return
	}

	if data, ok := metas.PStringDatas[apiID]; ok {
		res = data.StringValue
		opt = fmt.Sprintf("id:%d res:%s", data.StringId, res)
		find = true
		return
	}
	return
}

func (ac *AgentCache) FindSQLInfo(agentID string, sqlID int32) (res, opt string, find bool) {
	ac.metaLock.RLock()
	defer ac.metaLock.RUnlock()
	metas, ok := ac.Metas[agentID]
	if !ok {
		return
	}

	if data, ok := metas.PSqlDatas[sqlID]; ok {
		res = data.Sql
		opt = fmt.Sprintf("id:%d :%s", sqlID, data.Sql)
		find = true
		return
	}

	return
}

func (ac *AgentCache) SetAgentInfo(agentID string, info *ppv1.PAgentInfo) {
	ac.mu.Lock()
	defer ac.mu.Unlock()
	ac.Agents[agentID] = info
}

func (ac *AgentCache) GetAgentInfo(agentID string) *ppv1.PAgentInfo {
	ac.mu.Lock()
	defer ac.mu.Unlock()
	return ac.Agents[agentID]
}
