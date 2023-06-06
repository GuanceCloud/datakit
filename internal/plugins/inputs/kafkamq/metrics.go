// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

// Package kafkamq prom metrics
package kafkamq

import (
	"github.com/GuanceCloud/cliutils/metrics"
	"github.com/prometheus/client_golang/prometheus"
)

/*
   类型: sky, jaeger, custom
   指标:
   	1 消费数量 ： tag:分区，topic，类型\状态  val:数量

   	1 group选举次数 ： topic，类型  val:数量

	当dk运行时，访问 localhost:9529/metrics
*/
var kafkaConsumeMessages,
	kafkaGroupElection *prometheus.CounterVec

func metricsSetup() {
	kafkaConsumeMessages = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "datakit",
			Subsystem: "kafkamq",
			Name:      "consumer_message_total",
			Help:      "Kafka consumer message numbers from Datakit start",
		},
		[]string{
			"topic",
			"partition",
			"status",
		},
	)

	kafkaGroupElection = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "datakit",
			Subsystem: "kafkamq",
			Name:      "group_election_total",
			Help:      "Kafka group election count",
		},
		[]string{},
	)
}

func init() { //nolint:gochecknoinits
	metricsSetup()
	metrics.MustRegister(kafkaGroupElection, kafkaConsumeMessages)
}
