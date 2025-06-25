// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package rabbitmq

import (
	"strconv"
	"time"

	"github.com/GuanceCloud/cliutils/point"
	dkio "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/io"
)

type exchange struct {
	Name         string
	messageStats `json:"message_stats"`
	Type         string
	Internal     bool
	Vhost        string
	Durable      bool
	AutoDelete   bool `json:"auto_delete"`
}

func getExchange(n *Input) {
	var (
		exchanges    []exchange
		collectStart = time.Now()
	)

	if err := n.requestJSON("/api/exchanges", &exchanges); err != nil {
		l.Error(err.Error())
		n.lastErr = err
		return
	}

	var pts []*point.Point

	opts := append(point.DefaultMetricOptions(), point.WithTime(n.start))

	for _, exchange := range exchanges {
		kvs := point.NewTags(n.mergedTags)

		if exchange.Name == "" {
			exchange.Name = "(AMQP default)"
		}

		kvs = kvs.AddTag("exchange_name", exchange.Name).
			AddTag("url", n.URL).
			AddTag("exchange_name", exchange.Name).
			AddTag("type", exchange.Type).
			AddTag("vhost", exchange.Vhost).
			AddTag("internal", strconv.FormatBool(exchange.Internal)).
			AddTag("durable", strconv.FormatBool(exchange.Durable)).
			AddTag("auto_delete", strconv.FormatBool(exchange.AutoDelete)).
			AddV2("message_ack_count", exchange.messageStats.Ack, true).
			AddV2("message_ack_rate", exchange.messageStats.AckDetails.Rate, true).
			AddV2("message_confirm_count", exchange.messageStats.Confirm, true).
			AddV2("message_confirm_rate", exchange.messageStats.ConfirmDetail.Rate, true).
			AddV2("message_deliver_get_count", exchange.messageStats.DeliverGet, true).
			AddV2("message_deliver_get_rate", exchange.messageStats.DeliverGetDetails.Rate, true).
			AddV2("message_publish_count", exchange.messageStats.Publish, true).
			AddV2("message_publish_rate", exchange.messageStats.PublishDetails.Rate, true).
			AddV2("message_publish_in_count", exchange.messageStats.PublishIn, true).
			AddV2("message_publish_in_rate", exchange.messageStats.PublishInDetails.Rate, true).
			AddV2("message_publish_out_count", exchange.messageStats.PublishOut, true).
			AddV2("message_publish_out_rate", exchange.messageStats.PublishOutDetails.Rate, true).
			AddV2("message_redeliver_count", exchange.messageStats.Redeliver, true).
			AddV2("message_redeliver_rate", exchange.messageStats.RedeliverDetails.Rate, true).
			AddV2("message_return_unroutable_count", exchange.messageStats.ReturnUnroutable, true).
			AddV2("message_return_unroutable_count_rate", exchange.messageStats.ReturnUnroutableDetails.Rate, true)

		pts = append(pts, point.NewPointV2(exchangeMeasurementName, kvs, opts...))
	}

	if err := n.feeder.Feed(point.Metric, pts,
		dkio.WithCollectCost(time.Since(collectStart)),
		dkio.WithElection(n.Election),
		dkio.WithSource(inputName),
	); err != nil {
		l.Errorf("FeedMeasurement: %s", err.Error())
	}
}
