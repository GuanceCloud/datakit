// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package dialtesting

import (
	"testing"
	"time"

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

	child := &dialtesting.HTTPTask{
		Method: "GET",
		Task: &dialtesting.Task{
			Name:      "test",
			Frequency: "1ms",
		},
		URL: "http://127.0.0.1:9529",
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

	task, err := dialtesting.NewTask("", child)
	assert.NoError(t, err)

	dialer := newDialer(task, ipt)
	assert.Error(t, dialer.run())

	child = &dialtesting.HTTPTask{
		Method: "GET",
		Task: &dialtesting.Task{
			Name:      "test",
			PostURL:   "http://xxxxx?token=xxxxxx",
			Frequency: "1ms",
		},
		URL: "http://8.8.8.8",
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

	task, err = dialtesting.NewTask("", child)

	assert.NoError(t, err)

	dialer = newDialer(task, ipt)
	go func() {
		time.Sleep(100 * time.Millisecond)
		task.SetStatus(dialtesting.StatusStop)
		dialer.updateCh <- task
	}()
	assert.NoError(t, dialer.run())
}

func TestPopulateDFLabelTags(t *testing.T) {
	cases := []struct {
		Title  string
		Label  string
		Expect map[string]string
	}{
		{
			Title:  "no need to extract tags",
			Label:  "test",
			Expect: map[string]string{LabelDF: `["test"]`},
		},
		{
			Title:  "empty label",
			Label:  "",
			Expect: map[string]string{LabelDF: `[]`},
		},
		{
			Title:  "extract tags",
			Label:  "test,f1:2,f2:3:3",
			Expect: map[string]string{LabelDF: `["test","f1:2","f2:3:3"]`, "f1": "2", "f2": "3:3"},
		},
		{
			Title:  "new label format",
			Label:  "[\"tag1:value1\",\"tag2:value2\",\"tag3:value3\"]",
			Expect: map[string]string{LabelDF: "[\"tag1:value1\",\"tag2:value2\",\"tag3:value3\"]", "tag1": "value1", "tag2": "value2", "tag3": "value3"},
		},
	}
	for _, tc := range cases {
		tags := make(map[string]string)
		populateDFLabelTags(tc.Label, tags)

		assert.EqualValues(t, tc.Expect, tags)
	}
}
