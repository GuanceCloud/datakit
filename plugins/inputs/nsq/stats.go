package nsq

import (
	"encoding/json"
	"fmt"
	"time"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io"
)

type topicChannels map[string]*ChannelStats

type stats struct {
	topicCache map[string]topicChannels
	nodeCache  map[string]*nodeStats
}

type nodeStats struct {
	Depth        int64
	BackendDepth int64
	MessageCount int64
}

func (n *nodeStats) ToMap() map[string]interface{} {
	return map[string]interface{}{
		"depth":         n.Depth,
		"backend_depth": n.BackendDepth,
		"message_count": n.MessageCount,
	}
}

func newStats() *stats {
	return &stats{
		topicCache: make(map[string]topicChannels),
		nodeCache:  make(map[string]*nodeStats),
	}
}

func (s *stats) add(nodeHost string, body []byte) error {
	data := &DataStats{}
	if err := json.Unmarshal(body, data); err != nil {
		return fmt.Errorf("error parsing response: %s", err)
	}

	// Data was not parsed correctly attempt to use old format.
	if len(data.Version) < 1 {
		wrapper := &GlobalStats{}
		if err := json.Unmarshal(body, wrapper); err != nil {
			return fmt.Errorf("error parsing response: %s", err)
		}
		data = wrapper.Data
	}

	// 为什么需要此处进行 feedCache，而不是直接 make point？
	// 因为需要对多份数据进行整合
	s.feedCache(nodeHost, data)
	return nil
}

func (s *stats) feedCache(host string, data *DataStats) {
	for _, topic := range data.Topics {
		if _, ok := s.topicCache[topic.Name]; !ok {
			s.topicCache[topic.Name] = make(topicChannels)
		}

		for _, channel := range topic.Channels {
			if c, ok := s.topicCache[topic.Name][channel.Name]; !ok {
				s.topicCache[topic.Name][channel.Name] = channel
			} else {
				c.Merge(channel)
			}

			if n, ok := s.nodeCache[host]; !ok {
				s.nodeCache[host] = &nodeStats{}
			} else {
				n.Depth += topic.Depth
				n.BackendDepth += topic.BackendDepth
				n.MessageCount += topic.MessageCount
			}
		}
	}
}

func (s *stats) makePoint(addTags map[string]string, ts ...time.Time) ([]*io.Point, error) {
	var pts []*io.Point
	var lastErr error

	tim := time.Now()
	if len(ts) != 0 {
		tim = ts[0]
	}

	for topic, c := range s.topicCache {
		for channel, channelStats := range c {
			tags := map[string]string{
				"topic":   topic,
				"channel": channel,
			}
			for k, v := range addTags {
				tags[k] = v
			}
			fields := channelStats.ToMap()

			pt, err := io.MakePoint("nsq_topics", tags, fields, tim)
			if err != nil {
				lastErr = err
				continue
			}
			pts = append(pts, pt)
		}
	}

	for nodeHost, n := range s.nodeCache {
		tags := map[string]string{
			"server_host": nodeHost,
		}
		for k, v := range addTags {
			tags[k] = v
		}
		fields := n.ToMap()

		pt, err := io.MakePoint("nsq_nodes", tags, fields, tim)
		if err != nil {
			lastErr = err
			continue
		}
		pts = append(pts, pt)
	}

	return pts, lastErr
}

type GlobalStats struct {
	Code int64      `json:"status_code"`
	Txt  string     `json:"status_txt"`
	Data *DataStats `json:"data"`
}

type DataStats struct {
	Version   string        `json:"version"`
	Health    string        `json:"health"`
	StartTime int64         `json:"start_time"`
	Topics    []*TopicStats `json:"topics"`
}

// e2e_processing_latency is not modeled.
type TopicStats struct {
	Name         string          `json:"topic_name"`
	Depth        int64           `json:"depth"`
	BackendDepth int64           `json:"backend_depth"`
	MessageCount int64           `json:"message_count"`
	Paused       bool            `json:"paused"`
	Channels     []*ChannelStats `json:"channels"`
}

// e2e_processing_latency is not modeled.
type ChannelStats struct {
	Name          string `json:"channel_name"`
	Depth         int64  `json:"depth"`
	BackendDepth  int64  `json:"backend_depth"`
	InFlightCount int64  `json:"in_flight_count"`
	DeferredCount int64  `json:"deferred_count"`
	MessageCount  int64  `json:"message_count"`
	RequeueCount  int64  `json:"requeue_count"`
	TimeoutCount  int64  `json:"timeout_count"`
}

func (c *ChannelStats) Merge(n *ChannelStats) {
	c.Depth += n.Depth
	c.BackendDepth += n.BackendDepth
	c.InFlightCount += n.InFlightCount
	c.DeferredCount += n.DeferredCount
	c.MessageCount += n.MessageCount
	c.RequeueCount += n.RequeueCount
	c.TimeoutCount += n.TimeoutCount
}

func (c *ChannelStats) ToMap() map[string]interface{} {
	return map[string]interface{}{
		"depth":           c.Depth,
		"backend_depth":   c.BackendDepth,
		"in_flight_count": c.InFlightCount,
		"deferred_count":  c.DeferredCount,
		"message_count":   c.MessageCount,
		"requeue_count":   c.RequeueCount,
		"timeout_count":   c.TimeoutCount,
	}
}
