// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package nsq

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/GuanceCloud/cliutils/point"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs"
)

type topicChannels map[string]*ChannelStats

type stats struct {
	topicCache map[string]topicChannels
	nodeCache  map[string]*nodeStats
	ipt        *Input
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

func newStats(ipt *Input) *stats {
	return &stats{
		topicCache: make(map[string]topicChannels),
		nodeCache:  make(map[string]*nodeStats),
		ipt:        ipt,
	}
}

func (s *stats) add(nodeHost string, body []byte) error {
	data := &DataStats{}
	if err := json.Unmarshal(body, data); err != nil {
		return fmt.Errorf("error parsing response: %w", err)
	}

	// Data was not parsed correctly attempt to use old format.
	if len(data.Version) < 1 {
		wrapper := &GlobalStats{}
		if err := json.Unmarshal(body, wrapper); err != nil {
			return fmt.Errorf("error parsing response: %w", err)
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

func (s *stats) makePoint(addTags map[string]string, ptsTime time.Time) ([]*point.Point, error) {
	var pts []*point.Point
	var lastErr error

	opts := append(point.DefaultMetricOptions(), point.WithTime(ptsTime))

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

			pt := point.NewPoint(nsqTopics,
				append(point.NewTags(tags), point.NewKVs(fields)...),
				opts...,
			)
			pts = append(pts, pt)
		}
	}

	for nodeHost, n := range s.nodeCache {
		tags := map[string]string{
			"server_host": nodeHost,
		}

		remote := getURLHost(nodeHost)
		if remote == unknownHost {
			remote = ""
		}
		if s.ipt.Election {
			tags = inputs.MergeTagsWrapper(tags, s.ipt.Tagger.ElectionTags(), s.ipt.Tags, remote)
		} else {
			tags = inputs.MergeTagsWrapper(tags, s.ipt.Tagger.HostTags(), s.ipt.Tags, remote)
		}

		for k, v := range addTags {
			tags[k] = v
		}

		fields := n.ToMap()

		pt := point.NewPoint(nsqNodes,
			append(point.NewTags(tags), point.NewKVs(fields)...),
			opts...,
		)
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

type TopicStats struct {
	Name         string          `json:"topic_name"`
	Depth        int64           `json:"depth"`
	BackendDepth int64           `json:"backend_depth"`
	MessageCount int64           `json:"message_count"`
	Paused       bool            `json:"paused"`
	Channels     []*ChannelStats `json:"channels"`
}

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
