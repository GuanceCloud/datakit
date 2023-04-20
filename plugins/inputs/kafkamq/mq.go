// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

// Package kafkamq  mq
package kafkamq

import (
	"os"

	"github.com/Shopify/sarama"
)

var (
	// kafka 分区分配策略.
	assignors = map[string]sarama.BalanceStrategy{
		"range":      sarama.BalanceStrategyRange,
		"roundrobin": sarama.BalanceStrategyRoundRobin,
		"sticky":     sarama.BalanceStrategySticky,
	}
	defaultAssignors = sarama.BalanceStrategyRange // 轮训模式最适合 datakit 的工作模式.
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

type option func(con *sarama.Config)

func withVersion(version string) option {
	v, err := sarama.ParseKafkaVersion(version)
	if err != nil {
		log.Infof("can not get version from conf:[%s], use default version:%s", version, sarama.DefaultVersion.String())
		v = sarama.DefaultVersion
	}
	return func(con *sarama.Config) {
		con.Version = v
	}
}

func withAssignors(balance string) option {
	var bt sarama.BalanceStrategy
	if assignor, ok := assignors[balance]; ok {
		bt = assignor
	} else {
		log.Infof("can not find assignor, use default `roundrobin`")
		bt = defaultAssignors
	}

	return func(con *sarama.Config) {
		con.Consumer.Group.Rebalance.GroupStrategies = []sarama.BalanceStrategy{bt}
	}
}

func withOffset(offset int64) option {
	return func(con *sarama.Config) {
		con.Consumer.Offsets.Initial = sarama.OffsetOldest
		if offset == sarama.OffsetNewest {
			con.Consumer.Offsets.Initial = sarama.OffsetNewest
		}
	}
}

func withSASL(enable bool, mechanism, username, pw string) option {
	return func(config *sarama.Config) {
		if enable {
			config.Net.SASL.Enable = true
			config.Net.SASL.User = username
			config.Net.SASL.Password = pw
			config.Net.SASL.Mechanism = sarama.SASLMechanism(mechanism)
			config.Net.SASL.Version = sarama.SASLHandshakeV1
		}
	}
}

func newSaramaConfig(opts ...option) *sarama.Config {
	conf := sarama.NewConfig()
	conf.Consumer.Return.Errors = false
	conf.Consumer.Offsets.Initial = sarama.OffsetOldest // 未找到组消费位移的时候从哪边开始消费

	conf.Consumer.Offsets.Retry.Max = 10
	name, _ := os.Hostname()
	conf.ClientID = name

	for _, opt := range opts {
		opt(conf)
	}
	return conf
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
