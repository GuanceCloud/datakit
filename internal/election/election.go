// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

// Package election implements DataFlux central election client.
package election

import (
	"context"
	"fmt"
	"io"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/goroutine"

	"github.com/GuanceCloud/cliutils/logger"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs"
)

var (
	log                     = logger.DefaultSLogger("dk-election")
	electionIntervalDefault = 4
	CurrentElected          = "<checking...>"

	chStatus = make(chan ElectionStatus) // blocking
)

type Puller interface {
	Election(namespace, id string, reqBody io.Reader) ([]byte, error)
	ElectionHeartbeat(namespace, id string, reqBody io.Reader) ([]byte, error)
}

type Election interface {
	Run()
}

func SetLog() {
	log = logger.SLogger("dk-election")
}

func Start(opts ...ElectionOption) {
	SetLog()

	opt := option{}
	for idx := range opts {
		opts[idx](&opt)
	}

	if !opt.enabled {
		status := StatusDisabled
		electionStatusVec.WithLabelValues(CurrentElected, opt.id, opt.namespace, status.String()).Set(float64(status))
		log.Info("election not enabled.")
		return
	}

	isBanned := len(opt.nodeWhitelist) != 0
	for _, v := range opt.nodeWhitelist {
		if v == opt.id {
			isBanned = false
		}
	}

	if isBanned {
		status := StatusBanned
		electionStatusVec.WithLabelValues(CurrentElected, opt.id, opt.namespace, status.String()).Set(float64(status))
		log.Info("node is not whitelisted.")
		return
	}

	instance := newLeaderElection(&opt, inputs.GetElectionInputs())
	log.Infof("election mode with Dataway ,namespace: %s, id: %s", opt.namespace, opt.id)

	g := goroutine.G("election")
	g.Go(func(ctx context.Context) error {
		instance.Run()
		return nil
	})
}

func SetStatus(s ElectionStatus) error {
	select {
	case chStatus <- s:
		return nil
	default:
		return fmt.Errorf("busy or election not enabled")
	}
}
