// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package rabbitmq

import (
	"fmt"
	"net/url"
	"time"

	"github.com/GuanceCloud/cliutils/point"
	dkio "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/io"
)

type queue struct {
	queueTotals          // just to not repeat the same code
	messageStats         `json:"message_stats"`
	Memory               int64   `json:"memory"`
	Consumers            int64   `json:"consumers"`
	ConsumerUtilisation  float64 `json:"consumer_utilisation"` //nolint:misspell
	HeadMessageTimestamp int64   `json:"head_message_timestamp"`
	Name                 string
	Node                 string
	Vhost                string
	Durable              bool
	AutoDelete           bool   `json:"auto_delete"`
	IdleSince            string `json:"idle_since"`
}

func getQueues(n *Input) {
	var (
		Queues       []queue
		collectStart = time.Now()
		pts          []*point.Point
	)

	if err := n.requestJSON("/api/queues", &Queues); err != nil {
		l.Error(err.Error())
		n.lastErr = err
		return
	}

	opts := append(point.DefaultMetricOptions(), point.WithTime(n.start))
	for _, queue := range Queues {
		kvs := point.NewTags(n.mergedTags)

		kvs = kvs.AddTag("url", n.URL).
			AddTag("queue_name", queue.Name).
			AddTag("node_name", queue.Node).
			AddTag("vhost", queue.Vhost).
			AddV2("consumers", queue.Consumers, true).
			AddV2("consumer_utilization", queue.ConsumerUtilisation, true).
			AddV2("memory", queue.Memory, true).
			AddV2("head_message_timestamp", queue.HeadMessageTimestamp, true).
			AddV2("messages", queue.Messages, true).
			AddV2("messages_rate", queue.MessagesDetail.Rate, true).
			AddV2("messages_ready", queue.MessagesReady, true).
			AddV2("messages_ready_rate", queue.MessagesReadyDetail.Rate, true).
			AddV2("messages_unacknowledged", queue.MessagesUnacknowledged, true).
			AddV2("messages_unacknowledged_rate", queue.MessagesUnacknowledgedDetail.Rate, true).
			AddV2("message_ack_count", queue.messageStats.Ack, true).
			AddV2("message_ack_rate", queue.messageStats.AckDetails.Rate, true).
			AddV2("message_deliver_count", queue.messageStats.Deliver, true).
			AddV2("message_deliver_rate", queue.messageStats.DeliverDetails.Rate, true).
			AddV2("message_deliver_get_count", queue.messageStats.DeliverGet, true).
			AddV2("message_deliver_get_rate", queue.messageStats.DeliverGetDetails.Rate, true).
			AddV2("message_publish_count", queue.messageStats.Publish, true).
			AddV2("message_publish_rate", queue.messageStats.PublishDetails.Rate, true).
			AddV2("message_redeliver_count", queue.messageStats.Redeliver, true).
			AddV2("message_redeliver_rate", queue.messageStats.RedeliverDetails.Rate, true)

		bindings, err := n.getBindingCount(queue.Vhost, queue.Name)
		if err != nil {
			l.Errorf("get bindings err:%s", err.Error())
		} else {
			kvs = kvs.AddV2("bindings_count", bindings, true)
		}

		pts = append(pts, point.NewPointV2(queueMeasurementName, kvs, opts...))
	}

	if err := n.feeder.Feed(point.Metric, pts,
		dkio.WithCollectCost(time.Since(collectStart)),
		dkio.WithElection(n.Election),
		dkio.WithSource(inputName),
	); err != nil {
		l.Errorf("FeedMeasurement: %s", err.Error())
	}
}

func (ipt *Input) getBindingCount(vHost, queueName string) (int, error) {
	var binds []interface{}
	// 此处 vhost 可能是 / 需 encode
	err := ipt.requestJSON(fmt.Sprintf("/api/queues/%s/%s/bindings", url.QueryEscape(vHost), queueName), &binds)
	if err != nil {
		return 0, err
	}
	return len(binds), nil
}
