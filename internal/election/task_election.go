// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

// Package election implements DataFlux central election client.
package election

import (
	"bytes"
	"encoding/json"
	"io"
	"time"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
)

type taskElection struct {
	*option
	status            electionStatus
	applicationInputs map[string][]inputs.ElectionInput
	runningInputs     map[string][]inputs.ElectionInput
}

func newTaskElection(opt *option, plugins map[string][]inputs.ElectionInput) *taskElection {
	return &taskElection{
		option:            opt,
		status:            statusFail,
		applicationInputs: plugins,
		runningInputs:     plugins,
	}
}

func (x *taskElection) Run() {
	x.pausePlugins()
	tick := time.NewTicker(time.Second * time.Duration(electionIntervalDefault))
	defer tick.Stop()

	for {
		select {
		case <-datakit.Exit.Wait():
			return
		case <-tick.C:
			if err := x.runOnce(); err != nil {
				log.Error(err)
			}
		}
	}

	// electionInputs.WithLabelValues(x.namespace).Set(float64(len(x.plugins)))
}

func (x *taskElection) runOnce() error {
	defer func() {
		inputsPauseVec.WithLabelValues(x.id, x.namespace).Add(float64(len(x.runningInputs)))
	}()

	requ := x.buildRequest()
	b, err := requ.ToBytes()
	if err != nil {
		log.Errorf("failed to build request, err: %s", err)
		return err
	}

	body, err := x.puller.Election(x.namespace, x.id, b)
	if err != nil {
		log.Errorf("puller.Election: %s", err)
		return err
	}

	result := taskElectionResult{}
	if err := json.Unmarshal(body, &result); err != nil {
		log.Errorf("puller.Result: %s", err)
		return nil
	}

	log.Debugf("allowed plugins: %v", result.AllowedInputs)

	matched := x.match(result.AllowedInputs)

	if !matched {
		log.Info("pause all plugins for task election")
		x.pausePlugins()
		x.runningInputs = make(map[string][]inputs.ElectionInput)
		for _, name := range result.AllowedInputs {
			x.runningInputs[name] = append(x.runningInputs[name], x.applicationInputs[name]...)
			log.Debugf("add new plugins: %s, len(%d)", name, len(x.runningInputs[name]))
		}
		x.resumePlugins()
		log.Info("resume all plugins for task election")
	}

	return nil
}

var timeNow = time.Now

func (x *taskElection) buildRequest() *taskElectionRequest {
	requ := &taskElectionRequest{
		Namespace:         x.namespace,
		ID:                x.id,
		Timestamp:         timeNow().UnixMilli(),
		ApplicationInputs: make(map[string]int),
	}

	for name, v := range x.applicationInputs {
		requ.ApplicationInputs[name] = len(v)
	}
	for name := range x.runningInputs {
		requ.RunningInputs = append(requ.RunningInputs, name)
	}

	return requ
}

func (x *taskElection) pausePlugins() {
	for name, plugins := range x.runningInputs {
		for idx, p := range plugins {
			log.Debugf("pause %s %dth inputs...", name, idx)
			if err := p.Pause(); err != nil {
				log.Warn(err)
			}
		}
	}
}

func (x *taskElection) resumePlugins() {
	for name, plugins := range x.runningInputs {
		for idx, p := range plugins {
			log.Debugf("resume %s %dth inputs...", name, idx)
			if err := p.Resume(); err != nil {
				log.Warn(err)
			}
		}
	}
}

func (x *taskElection) match(allowedInputs []string) bool {
	if len(allowedInputs) != len(x.runningInputs) {
		return false
	}
	for name := range x.runningInputs {
		found := contains(allowedInputs, name)
		if !found {
			return false
		}
	}
	return true
}

type taskElectionRequest struct {
	Namespace         string         `json:"namespace,omitempty"`
	ID                string         `json:"id"`
	Timestamp         int64          `json:"timestamp"`
	ApplicationInputs map[string]int `json:"application_inputs"`
	RunningInputs     []string       `json:"running_inputs"`
}

func (req *taskElectionRequest) ToBytes() (io.Reader, error) {
	var buff bytes.Buffer
	b, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}

	_, err = buff.Write(b)
	if err != nil {
		return nil, err
	}

	return &buff, nil
}

type taskElectionResult struct {
	AllowedInputs []string `json:"allowed_inputs"`
}

func contains(array []string, s string) bool {
	for _, a := range array {
		if a == s {
			return true
		}
	}
	return false
}
