// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package rabbitmq

import (
	"time"

	"github.com/GuanceCloud/cliutils/point"
	dkio "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/io"
)

// objectTotals ...
type objectTotals struct {
	Channels    int64
	Connections int64
	Consumers   int64
	Exchanges   int64
	Queues      int64
}

type queueTotals struct {
	Messages       int64
	MessagesDetail details `json:"messages_details"`

	MessagesReady       int64   `json:"messages_ready"`
	MessagesReadyDetail details `json:"messages_ready_details"`

	MessagesUnacknowledged       int64   `json:"messages_unacknowledged"`
	MessagesUnacknowledgedDetail details `json:"messages_unacknowledged_details"`
}

// details ...
type details struct {
	Rate float64 `json:"rate"`
}

// messageStats ...
type messageStats struct {
	Ack                     int64
	AckDetails              details `json:"ack_details"`
	Confirm                 int64   `json:"confirm"`
	ConfirmDetail           details `json:"ack_details_details"`
	Deliver                 int64
	DeliverDetails          details `json:"deliver_details"`
	DeliverGet              int64   `json:"deliver_get"`
	DeliverGetDetails       details `json:"deliver_get_details"`
	Publish                 int64
	PublishDetails          details `json:"publish_details"`
	Redeliver               int64
	RedeliverDetails        details `json:"redeliver_details"`
	PublishIn               int64   `json:"publish_in"`
	PublishInDetails        details `json:"publish_in_details"`
	PublishOut              int64   `json:"publish_out"`
	PublishOutDetails       details `json:"publish_out_details"`
	ReturnUnroutable        int64   `json:"return_unroutable"`
	ReturnUnroutableDetails details `json:"return_unroutable_details"`
}

type listeners struct {
	Protocol string `json:"protocol"`
}

type overviewResponse struct {
	Version      string        `json:"rabbitmq_version"`
	ClusterName  string        `json:"cluster_name"`
	MessageStats *messageStats `json:"message_stats"`
	ObjectTotals *objectTotals `json:"object_totals"`
	QueueTotals  *queueTotals  `json:"queue_totals"`
	Listeners    []listeners   `json:"listeners"`
}

func getOverview(n *Input) {
	var (
		collectStart = time.Now()
		overview     = &overviewResponse{}
	)

	err := n.requestJSON("/api/overview", &overview)
	if err != nil {
		l.Errorf(err.Error())
		n.lastErr = err
		return
	}

	if overview.QueueTotals == nil || overview.ObjectTotals == nil || overview.MessageStats == nil {
		l.Errorf("Wrong answer from rabbitmq. Probably auth issue")
		return
	}

	kvs := point.NewTags(n.mergedTags)

	kvs = kvs.AddTag("url", n.URL).
		AddTag("rabbitmq_version", overview.Version).
		AddTag("cluster_name", overview.ClusterName).
		Set("object_totals_channels", overview.ObjectTotals.Channels).
		Set("object_totals_connections", overview.ObjectTotals.Connections).
		Set("object_totals_consumers", overview.ObjectTotals.Consumers).
		Set("object_totals_queues", overview.ObjectTotals.Queues).
		Set("message_ack_count", overview.MessageStats.Ack).
		Set("message_ack_rate", overview.MessageStats.AckDetails.Rate).
		Set("message_confirm_count", overview.MessageStats.Confirm).
		Set("message_confirm_rate", overview.MessageStats.ConfirmDetail.Rate).
		Set("message_deliver_get_count", overview.MessageStats.DeliverGet).
		Set("message_deliver_get_rate", overview.MessageStats.DeliverGetDetails.Rate).
		Set("message_publish_count", overview.MessageStats.Publish).
		Set("message_publish_rate", overview.MessageStats.PublishDetails.Rate).
		Set("message_publish_in_count", overview.MessageStats.PublishIn).
		Set("message_publish_in_rate", overview.MessageStats.PublishInDetails.Rate).
		Set("message_publish_out_count", overview.MessageStats.PublishOut).
		Set("message_publish_out_rate", overview.MessageStats.PublishOutDetails.Rate).
		Set("message_redeliver_count", overview.MessageStats.Redeliver).
		Set("message_redeliver_rate", overview.MessageStats.RedeliverDetails.Rate).
		Set("message_return_unroutable_count", overview.MessageStats.ReturnUnroutable).
		Set("message_return_unroutable_count_rate", overview.MessageStats.ReturnUnroutableDetails.Rate).
		Set("queue_totals_messages_count", overview.QueueTotals.Messages).
		Set("queue_totals_messages_rate", overview.QueueTotals.MessagesDetail.Rate).
		Set("queue_totals_messages_ready_count", overview.QueueTotals.MessagesReady).
		Set("queue_totals_messages_ready_rate", overview.QueueTotals.MessagesReadyDetail.Rate).
		Set("queue_totals_messages_unacknowledged_count", overview.QueueTotals.MessagesUnacknowledged).
		Set("queue_totals_messages_unacknowledged_rate", overview.QueueTotals.MessagesUnacknowledgedDetail.Rate)

	opts := append(point.DefaultMetricOptions(), point.WithTime(n.start))
	pt := point.NewPoint(overviewMeasurementName, kvs, opts...)

	if err := n.feeder.Feed(point.Metric, []*point.Point{pt},
		dkio.WithCollectCost(time.Since(collectStart)),
		dkio.WithElection(n.Election),
		dkio.WithSource(inputName),
	); err != nil {
		l.Errorf("FeedMeasurement: %s", err.Error())
	}
}
