// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

// Package election implements DataFlux central election client.
package election

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs"
)

type fakeElectionInput struct{}

func (inp *fakeElectionInput) Pause() error  { return nil }
func (inp *fakeElectionInput) Resume() error { return nil }

func TestTaskElectionBuildRequest(t *testing.T) {
	t.Skip()

	tim := time.Now()
	timeNow = func() time.Time {
		return tim
	}

	opt := option{
		id:        "id",
		namespace: "ns",
	}
	plugins := map[string][]inputs.ElectionInput{
		"fake_input01": {&fakeElectionInput{}},
		"fake_input02": {&fakeElectionInput{}, &fakeElectionInput{}},
	}

	task := newTaskElection(&opt, plugins)

	expectRequ := &taskElectionRequest{
		Namespace:         "ns",
		ID:                "id",
		Timestamp:         timeNow().UnixMilli(),
		ApplicationInputs: map[string]int{"fake_input01": 1, "fake_input02": 2},
		RunningInputs:     []string{"fake_input01", "fake_input02"},
	}

	actualRequ := task.buildRequest()
	assert.Equal(t, expectRequ, actualRequ)
}
