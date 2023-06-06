// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

// Package confd implements various configure daemon clients.
package confd

import (
	"github.com/GuanceCloud/cliutils/logger"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/config"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs"
)

func Run(arr []*config.ConfdCfg) error {
	confds = arr

	l = logger.SLogger("confd")

	// First need RunInputs. lots of start in this func
	// must befor StartConfd()
	if err := inputs.RunInputs(); err != nil {
		l.Error("error running inputs: %v", err)
		return err
	}

	// if use config source from confd, like etcd zookeeper concul tredis ...
	if err := startConfd(); err != nil {
		l.Errorf("config.StartConfd failed: %v", err)
		return err
	}

	return nil
}
