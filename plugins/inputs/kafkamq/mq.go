// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

// Package kafkamq  mq
package kafkamq

import (
	"github.com/Shopify/sarama"
)

var (
	// kafka 分区分配策略.
	assignors = map[string]sarama.BalanceStrategy{
		"range":      sarama.BalanceStrategyRange,
		"roundrobin": sarama.BalanceStrategyRoundRobin,
		"sticky":     sarama.BalanceStrategySticky,
	}
	defaultAssignors = sarama.BalanceStrategyRoundRobin // 轮训模式最适合 datakit 的工作模式.
)

func getKafkaVersion(ver string) sarama.KafkaVersion {
	version, err := sarama.ParseKafkaVersion(ver)
	if err != nil {
		log.Infof("can not get version from conf:[%s], use default version:%s", ver, sarama.DefaultVersion.String())
		return sarama.DefaultVersion
	}
	log.Infof("use version:%s", version.String())
	return version
}

func getAddrs(addr string, addrs []string) []string {
	kafkaAddress := make([]string, 0)
	if addr != "" {
		kafkaAddress = append(kafkaAddress, addr)
	}
	if addrs != nil {
		kafkaAddress = append(kafkaAddress, addrs...)
	}
	return kafkaAddress
}

func getAssignors(balance string) sarama.BalanceStrategy {
	if assignor, ok := assignors[balance]; ok {
		return assignor
	}
	log.Infof("can not find assignor, use default `roundrobin`")
	return defaultAssignors
}
