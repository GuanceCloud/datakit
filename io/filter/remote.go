// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package filter

import (
	"time"
)

type Filters struct {
	Filters map[string]FilterConditions `json:"filters"`
	// other fields ignored
	PullInterval time.Duration `json:"pull_interval"`
}

func (f *filter) remotePull(what string) ([]byte, error) {
	var (
		start = time.Now()
		body  []byte
		err   error
	)

	defer func() {
		if err != nil {
			filterPullLatencyVec.WithLabelValues("failed").Observe(float64(time.Since(start) / time.Millisecond))
		} else {
			filterPullLatencyVec.WithLabelValues("ok").Observe(float64(time.Since(start) / time.Millisecond))
		}
	}()

	body, err = f.puller.Pull(what)

	f.mtx.Lock()
	defer f.mtx.Unlock()

	if err != nil {
		return nil, err
	}

	l.Debugf("filter condition body: %s", string(body))

	return body, nil
}
