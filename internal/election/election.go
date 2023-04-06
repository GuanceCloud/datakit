// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

// Package election implements DataFlux central election client.
package election

import (
	"time"

	"github.com/GuanceCloud/cliutils/logger"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
)

/*
 * DataKit 选举说明文档
 *
 * 流程：
 *      1. DataKit 开启 cfg.EnableElection（booler）配置
 *      2. 当运行对应的采集器（采集器列表在 config/inputcfg.go）时，程序会创建一个 goroutine 向 DataWay 发送选举请求，并携带 token 和 namespace（若存在）以及 id
 *      3. 选举成功担任 leader 后会持续发送心跳，心跳间隔过长或选举失败，会恢复 candidate 状态并继续发送选举请求
 *      4. 采集器端只要在采集数据时，判断当前是否为 leader 状态，具体使用见下
 *
 * 使用方式：
 *      1. 在 config/inputcfg.go 的 electionInputs 中添加需要选举的采集器（目前使用此方式后续会优化）
 *      2. 采集器中 import "gitlab.jiagouyun.com/cloudcare-tools/datakit/election"
 *      4. 详见 demo 采集器
 */

var (
	defaultCandidate        = &candidate{status: statusFail} // default set defeated
	log                     = logger.DefaultSLogger("dk-election")
	HTTPTimeout             = time.Second * 3
	electionIntervalDefault = 4
	CurrentElected          = "<checking...>"
)

type Puller interface {
	Election(namespace, id string) ([]byte, error)
	ElectionHeartbeat(namespace, id string) ([]byte, error)
}

type candidate struct {
	enabled       bool
	status        electionStatus
	id, namespace string

	puller Puller

	plugins []inputs.ElectionInput
}

func Start(opts ...ElectionOption) {
	log = logger.SLogger("dk-election")
	for _, opt := range opts {
		if opt != nil {
			opt(defaultCandidate)
		}
	}

	defaultCandidate.run()
}

type electionResult struct {
	Content struct {
		Status       string `json:"status"`
		Namespace    string `json:"namespace,omitempty"`
		ID           string `json:"id"`
		IncumbencyID string `json:"incumbency_id,omitempty"`
		ErrorMsg     string `json:"error_msg,omitempty"`
		Interval     int    `json:"interval"`
	} `json:"content"`
}
