// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2022-present Guance, Inc.

// Package nvidiasmi collects host nvidiasmi metrics.
package nvidiasmi

import (
	"context"

	"github.com/GuanceCloud/cliutils/logger"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/datakit"
)

// type dataStruct struct {
// 	server        string
// 	data          []byte
// }

func (ipt *Input) collectRemote() error {
	// use goroutine, send data through dataCh, with timeout, with tag
	dataCh := make(chan datakit.SSHData, 1)
	l = logger.SLogger("gpu_smi")
	g := datakit.G("gpu_smi")

	g.Go(func(ctx context.Context) error {
		return datakit.SSHGetData(dataCh, &ipt.SSHServers, ipt.Timeout.Duration)
	})

	// to receive data through dataCh
	ipt.handleDatas(dataCh)

	return nil
}

// handle ssh remote server data. need not ctx, because SSHGetData() has timeout close ch.
func (ipt *Input) handleDatas(ch chan datakit.SSHData) {
	for datas := range ch {
		// fmt.Println(datas.Server)
		// fmt.Println(string(datas.Data), datas.Server)

		// convert and calculate GPU metrics
		_ = ipt.convert(datas.Data, datas.Server)
	}
}
