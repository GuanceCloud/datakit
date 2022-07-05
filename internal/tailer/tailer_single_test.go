// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package tailer

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/multiline"
)

func TestGenerateCRILogs(t *testing.T) {
	for c, test := range []struct {
		lines    []string
		expected []string
	}{
		{
			lines: []string{
				"2016-10-20T18:39:20.57606443Z stdout F cri stdout test log",
				"2016-10-20T18:39:20.57606443Z stderr F cri stderr test log",
				"2016-10-20T18:39:20.57606443Z stderr F cri stderr test log padding",
			},
			expected: []string{
				"cri stdout test log",
				"cri stderr test log",
			},
		},
		{
			lines: []string{
				"2016-10-20T18:39:20.57606443Z stdout P cri stdout test log part 1\n",
				"2016-10-20T18:39:20.57606443Z stderr F cri stderr test log part 2",
				"2016-10-20T18:39:20.57606443Z stderr F cri stderr test log padding",
			},
			expected: []string{
				"cri stdout test log part 1cri stderr test log part 2",
			},
		},
	} {
		mult, _ := multiline.New("", 10000)
		tail := &Single{
			mult: mult,
			opt:  &Option{},
		}
		t.Logf("TestCase #%d: %+v", c, test)
		logs := tail.generateCRILogs(test.lines)
		assert.Equal(t, test.expected, logs)
	}
}
