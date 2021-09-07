package nsq

import (
	"encoding/json"
	"fmt"
	"sync"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io"
)

type stats struct {
	mu sync.Mutex

	// map["topic"]["channle"]*ChannelStats
	topicCache map[string]map[string]*ChannelStats

	tags map[string]string

	//
	//nodeCache map[string]
}

func newStats(tags map[string]string) *stats {
	return &stats{
		topicCache: make(map[string]map[string]*ChannelStats),
		tags:       tags,
	}
}

func (s *stats) add(body []byte) error {
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
		data = &wrapper.Data
	}

	s.feedCache(data)
	return nil
}

func (s *stats) feedCache(data *DataStats) {
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, topic := range data.Topics {
		if _, ok := s.topicCache[topic.Name]; !ok {
			s.topicCache[topic.Name] = make(map[string]*ChannelStats)
		}
		for idx, channel := range topic.Channels {
			c, ok := s.topicCache[topic.Name][channel.Name]
			if !ok {
				s.topicCache[topic.Name][channel.Name] = &topic.Channels[idx]
			} else {
				c.Merge(&topic.Channels[idx])
			}
		}
	}
}

func (s *stats) makePoint() ([]*io.Point, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	var pts []*io.Point
	var lastErr error

	for topic, c := range s.topicCache {
		for channel, channelStats := range c {
			tags := map[string]string{
				"topic":   topic,
				"channel": channel,
			}
			for k, v := range s.tags {
				tags[k] = v
			}
			fields := channelStats.ToMap()

			pt, err := io.MakePoint("nsq_topics", tags, fields)
			if err != nil {
				lastErr = err
				continue
			}
			pts = append(pts, pt)
		}
	}

	return pts, lastErr
}

type GlobalStats struct {
	Code int64     `json:"status_code"`
	Txt  string    `json:"status_txt"`
	Data DataStats `json:"data"`
}

type DataStats struct {
	Version   string       `json:"version"`
	Health    string       `json:"health"`
	StartTime int64        `json:"start_time"`
	Topics    []TopicStats `json:"topics"`
}

// e2e_processing_latency is not modeled
type TopicStats struct {
	Name         string         `json:"topic_name"`
	Depth        int64          `json:"depth"`
	BackendDepth int64          `json:"backend_depth"`
	MessageCount int64          `json:"message_count"`
	Paused       bool           `json:"paused"`
	Channels     []ChannelStats `json:"channels"`
}

// e2e_processing_latency is not modeled
type ChannelStats struct {
	Name          string `json:"channel_name"`
	Depth         int64  `json:"depth"`
	BackendDepth  int64  `json:"backend_depth"`
	InFlightCount int64  `json:"in_flight_count"`
	DeferredCount int64  `json:"deferred_count"`
	MessageCount  int64  `json:"message_count"`
	RequeueCount  int64  `json:"requeue_count"`
	TimeoutCount  int64  `json:"timeout_count"`
	Paused        bool   `json:"paused"`
	//Clients       []ClientStats `json:"clients"`
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

type ClientStats struct {
	// DEPRECATED 1.x+, still here as the structs are currently being shared for parsing v3.x and 1.x
	Name                          string `json:"name"`
	ID                            string `json:"client_id"`
	Hostname                      string `json:"hostname"`
	Version                       string `json:"version"`
	RemoteAddress                 string `json:"remote_address"`
	State                         int64  `json:"state"`
	ReadyCount                    int64  `json:"ready_count"`
	InFlightCount                 int64  `json:"in_flight_count"`
	MessageCount                  int64  `json:"message_count"`
	FinishCount                   int64  `json:"finish_count"`
	RequeueCount                  int64  `json:"requeue_count"`
	ConnectTime                   int64  `json:"connect_ts"`
	SampleRate                    int64  `json:"sample_rate"`
	Deflate                       bool   `json:"deflate"`
	Snappy                        bool   `json:"snappy"`
	UserAgent                     string `json:"user_agent"`
	TLS                           bool   `json:"tls"`
	TLSCipherSuite                string `json:"tls_cipher_suite"`
	TLSVersion                    string `json:"tls_version"`
	TLSNegotiatedProtocol         string `json:"tls_negotiated_protocol"`
	TLSNegotiatedProtocolIsMutual bool   `json:"tls_negotiated_protocol_is_mutual"`
}
