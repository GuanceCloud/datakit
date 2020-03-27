package internal

import (
	"encoding/json"
	"strings"
	"sync"
	"sync/atomic"

	"github.com/influxdata/telegraf"
)

type PluginStat interface {
	StatMetric() telegraf.Metric
	IsRunning() bool
}

type MetricGatherStat struct {
	TotalMetric       uint64
	SuccessSendMetric uint64
	DroppedMetric     uint64
}

type ContextErr struct {
	ID      string
	Context string
	Content string
}

func (e *ContextErr) Error() string {
	j := map[string]string{
		"id":      e.ID,
		"context": e.Context,
		"content": e.Content,
	}
	data, err := json.Marshal(j)
	if err != nil {
		return e.ID + "\n" + e.Context + "\n" + e.Content
	}
	return string(data)
}

type InputStat struct {
	MetricGatherStat
	Stat   int32
	ErrIDs []string

	mutex sync.Mutex
}

func (s *InputStat) Fields() map[string]interface{} {
	fields := map[string]interface{}{
		"running":             s.Stat > 0,
		"total_metric":        s.TotalMetric,
		"success_send_metric": s.SuccessSendMetric,
		"dropped_metric":      s.DroppedMetric,
	}
	if len(s.ErrIDs) > 0 {
		fields["error_ids"] = strings.Join(s.ErrIDs, ",")
	}

	return fields
}

func (s *MetricGatherStat) IncrTotal() {
	atomic.AddUint64(&s.TotalMetric, 1)
}

func (s *MetricGatherStat) AddTotal(v uint64) {
	atomic.AddUint64(&s.TotalMetric, v)
}

func (s *MetricGatherStat) SetTotal(v uint64) {
	atomic.StoreUint64(&s.TotalMetric, v)
}

func (s *InputStat) SetStat(v int) {
	atomic.StoreInt32(&s.Stat, int32(v))
}

func (s *InputStat) AddErrorID(id string) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	s.ErrIDs = append(s.ErrIDs, id)
}

func (s *InputStat) ClearErrorID() {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	s.ErrIDs = s.ErrIDs[:0]
}
