// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

//go:build !windows
// +build !windows

package dialtesting

import (
	"testing"

	"github.com/GuanceCloud/cliutils/dialtesting"
	"github.com/stretchr/testify/assert"
)

func TestInternalNetwork(t *testing.T) {
	dialWorker = &worker{
		sender: &mockSender{},
	}
	ipt := &Input{
		DisableInternalNetworkTask: true,
	}

	task := &dialtesting.HTTPTask{
		Name:      "test",
		URL:       "http://127.0.0.1:9529",
		Method:    "GET",
		Frequency: "1ms",
		SuccessWhen: []*dialtesting.HTTPSuccess{
			{
				StatusCode: []*dialtesting.SuccessOption{
					{
						Is: "200",
					},
				},
			},
		},
	}

	dialer := newDialer(task, ipt)
	assert.Error(t, dialer.run())

	task = &dialtesting.HTTPTask{
		Name:      "test",
		URL:       "http://8.8.8.8",
		PostURL:   "http://xxxxx?token=xxxxxx",
		Method:    "GET",
		Frequency: "1ms",
		AdvanceOptions: &dialtesting.HTTPAdvanceOption{
			RequestTimeout: "1s",
		},
		SuccessWhen: []*dialtesting.HTTPSuccess{
			{
				StatusCode: []*dialtesting.SuccessOption{
					{
						Is: "200",
					},
				},
			},
		},
	}

	dialer = newDialer(task, ipt)
	go func() {
		task.CurStatus = dialtesting.StatusStop
		dialer.updateCh <- task
	}()
	assert.NoError(t, dialer.run())
}
