// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package election

import (
	"context"
	"encoding/json"
	"time"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
)

func (x *candidate) run() {
	defer func() {
		electionStatusVec.WithLabelValues(
			CurrentElected,
			x.id,
			x.namespace,
			x.status.String(),
		).Set(float64(x.status))
	}()

	if !x.enabled {
		x.status = statusDisabled
		log.Info("election not enabled.")
		return
	}

	x.plugins = inputs.GetElectionInputs()

	electionInputs.WithLabelValues(x.namespace).Set(float64(len(x.plugins)))

	log.Infof("namespace: %s, id: %s", x.namespace, x.id)
	log.Infof("get %d election inputs", len(x.plugins))

	x.startElection()
}

func (x *candidate) startElection() {
	g := datakit.G("election")
	g.Go(func(ctx context.Context) error {
		x.pausePlugins()
		tick := time.NewTicker(time.Second * time.Duration(electionIntervalDefault))
		defer tick.Stop()

		for {
			select {
			case <-datakit.Exit.Wait():
				return nil
			case <-tick.C:
				electionInterval := x.runOnce()
				if electionInterval != electionIntervalDefault {
					tick.Reset(time.Second * time.Duration(electionInterval))
					electionIntervalDefault = electionInterval
				}
			}
		}
	})
}

func (x *candidate) runOnce() int {
	var (
		elecIntv int
		err      error
	)

	switch x.status {
	case statusSuccess:
		elecIntv, err = x.keepalive()
	case statusFail:
		elecIntv, err = x.tryElection()
	case statusDisabled: // pass
		return electionIntervalDefault
	}

	if err != nil {
		io.FeedLastError("election", err.Error())
	}

	return elecIntv
}

func (x *candidate) tryElection() (int, error) {
	var (
		electedTime int64
		start       = time.Now()
	)

	body, err := x.puller.Election(x.namespace, x.id)
	if err != nil {
		log.Errorf("puller.Election: %s", err)

		return electionIntervalDefault, err
	}

	defer func() {
		electionVec.WithLabelValues(
			x.namespace,
			x.status.String(),
		).Observe(float64(time.Since(start) / time.Millisecond))

		electionStatusVec.WithLabelValues(
			CurrentElected,
			x.id,
			x.namespace,
			x.status.String(),
		).Set(float64(electedTime))
	}()

	e := electionResult{}
	if err := json.Unmarshal(body, &e); err != nil {
		log.Error(err)

		return electionIntervalDefault, nil
	}

	log.Debugf("result body: %s", body)

	if CurrentElected != e.Content.IncumbencyID {
		CurrentElected = e.Content.IncumbencyID
		electionStatusVec.Reset() // cleanup election status metrics
	}

	switch e.Content.Status {
	case statusFail.String():
		electionStatusVec.Reset() // cleanup election status if election failed

		x.status = statusFail

	case statusSuccess.String():
		electionStatusVec.Reset() // cleanup election status if election ok

		x.status = statusSuccess
		x.resumePlugins()
		electedTime = time.Now().Unix()

	default:
		log.Warnf("unknown election status: %s", e.Content.Status)
	}

	return e.Content.Interval, nil
}
