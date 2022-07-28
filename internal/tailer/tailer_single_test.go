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

func TestGenerateJSONLogs(t *testing.T) {
	str16k := func() string {
		strBytes := []byte{}
		for i := 0; i < 16*1024; i++ {
			strBytes = append(strBytes, 96)
		}

		return string(strBytes)
	}()
	for c, test := range []struct {
		lines    []string
		expected []string
	}{
		{
			lines: []string{
				`{"log":"[WARNING] test json log\n","stream":"stdout","time":"2022-06-30T03:22:20.055429751Z"}`,
				`{"log":"padding\n","stream":"stdout","time":"2022-06-30T03:22:20.055429751Z"}`,
			},
			expected: []string{
				"[WARNING] test json log\n",
			},
		},
		{
			lines: []string{
				`{"log":"` + str16k + `","stream":"stdout","time":"2022-06-30T03:22:20.055429751Z"}`,
				`{"log":"test partial log2\n","stream":"stdout","time":"2022-06-30T03:22:20.055429751Z"}`,
				`{"log":"padding\n","stream":"stdout","time":"2022-06-30T03:22:20.055429751Z"}`,
			},
			expected: []string{
				str16k + "test partial log2\n",
			},
		},
		{
			lines: []string{
				`{"log":"not 16k log log1","stream":"stdout","time":"2022-06-30T03:22:20.055429751Z"}`,
				`{"log":"test partial log2\n","stream":"stdout","time":"2022-06-30T03:22:20.055429751Z"}`,
				`{"log":"padding\n","stream":"stdout","time":"2022-06-30T03:22:20.055429751Z"}`,
			},
			expected: []string{
				"not 16k log log1",
				"test partial log2\n",
			},
		},
	} {
		mult, _ := multiline.New("", 10000)
		tail := &Single{
			mult: mult,
			opt: &Option{
				log: l,
			},
		}
		t.Logf("TestCase #%d: %+v", c, test)
		logs := tail.generateJSONLogs(test.lines)
		assert.Equal(t, test.expected, logs)
	}
}

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
