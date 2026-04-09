// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package upgrader

import (
	"github.com/GuanceCloud/cliutils/logger"
	"go.uber.org/atomic"

	ws "gitlab.jiagouyun.com/cloudcare-tools/datakit/dca/websocket"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/config"
)

var ui = &upgraderImpl{
	upgradeStatus: atomic.NewInt32(0),
}

type RegisterData struct {
	Datakit *ws.DataKit       `json:"datakit"`
	DCA     *config.DCAConfig `json:"dca"`
}

func DebugRun() {
	if err := Cfg.LoadMainTOML(MainConfigFile); err != nil {
		l.Warnf("unable to load main config file: %s", err)
	}
	Cfg.SetLogging()
	l = logger.SLogger("main")
	ui.c = Cfg

	if err := startDCA(&serviceImpl{done: make(chan struct{})}); err != nil {
		l.Errorf("startDCA failed: %s", err.Error())
	}
}
