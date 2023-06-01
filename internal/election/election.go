// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

// Package election implements DataFlux central election client.
package election

import (
	"context"
	"io"

	"github.com/GuanceCloud/cliutils/logger"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
)

var (
	log                     = logger.DefaultSLogger("dk-election")
	electionIntervalDefault = 4
	CurrentElected          = "<checking...>"
)

type Puller interface {
	Election(namespace, id string, reqBody io.Reader) ([]byte, error)
	ElectionHeartbeat(namespace, id string, reqBody io.Reader) ([]byte, error)
}

type Election interface {
	Run()
}

func Start(opts ...ElectionOption) {
	log = logger.SLogger("dk-election")

	opt := option{}
	for idx := range opts {
		opts[idx](&opt)
	}

	if !opt.enabled {
		status := statusDisabled
		electionStatusVec.WithLabelValues(CurrentElected, opt.id, opt.namespace, status.String()).Set(float64(status))
		log.Info("election not enabled.")
		return
	}

	var electionInstance Election

	switch opt.mode {
	case modeDataway:
		electionInstance = newLeaderElection(&opt, inputs.GetElectionInputs())
		log.Info("election mode with Dataway")
	case modeOperator:
		electionInstance = newTaskElection(&opt, inputs.GetElectionInputs())
		opt.namespace = "N/A"
		log.Info("election mode with Operator")
	default:
		log.Info("invalid election mode, election not enabled")
		return
	}

	log.Infof("namespace: %s, id: %s", opt.namespace, opt.id)
	// log.Infof("get %d election inputs", len(x.plugins))

	g := datakit.G("election")
	g.Go(func(ctx context.Context) error {
		electionInstance.Run()
		return nil
	})
}
