// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package textparser

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseDockerJsonLog(t *testing.T) {
	cases := []struct {
		in   string
		out  *LogMessage
		fail bool
	}{
		{
			in: `{"log":"[INFO] test log, is partial","stream":"stdout","time":"2024-04-28T03:22:20.055429751Z"}`,
			out: &LogMessage{
				Stream:    Stdout,
				Log:       []byte("[INFO] test log, is partial"),
				IsPartial: true,
			},
		},
		{
			in: `{"log":"[INFO] test log, is full\n","stream":"stderr","time":"2024-04-28T03:22:20.055429751Z"}`,
			out: &LogMessage{
				Stream:    Stderr,
				Log:       []byte("[INFO] test log, is full\n"),
				IsPartial: false,
			},
		},
		{
			in: `{"log":"","stream":"stderr","time":"2024-04-28T03:22:20.055429751Z"}`,
			out: &LogMessage{
				Stream:    Stderr,
				Log:       nil,
				IsPartial: false,
			},
		},
		{
			in: `{"stream":"stderr","time":"2024-04-28T03:22:20.055429751Z"}`,
			out: &LogMessage{
				Stream:    Stderr,
				Log:       nil,
				IsPartial: false,
			},
		},
		{
			in:   `{"log":"[WARN] test log, unexpected stream\n","stream":"unexpected-stream","time":"2024-04-28T03:22:20.055429751Z"}`,
			fail: true,
		},
		// {
		// 	// parse timestamp fail
		// 	in:   `{"log":"[WARN] test log, unexpected timestamp\n","stream":"stderr","time":"Sun Apr 28 06:19:06 PM CST 2024"}`,
		// 	fail: true,
		// },
	}

	for _, tc := range cases {
		msg := new(LogMessage)

		err := ParseDockerJSONLog([]byte(tc.in), msg)
		if tc.fail && assert.Error(t, err) {
			return
		}
		assert.Nil(t, err)
		assert.Equal(t, tc.out, msg)
	}
}

func TestParseCRILog(t *testing.T) {
	cases := []struct {
		in   string
		out  *LogMessage
		fail bool
	}{
		{
			in: `2024-04-20T18:39:20.57606443Z stdout P [INFO] test log, is partial`,
			out: &LogMessage{
				Stream:    Stdout,
				Log:       []byte("[INFO] test log, is partial"),
				IsPartial: true,
			},
		},
		{
			in: `2024-04-20T18:39:20.57606443Z stderr F [INFO] test log, is full`,
			out: &LogMessage{
				Stream:    Stderr,
				Log:       []byte("[INFO] test log, is full"),
				IsPartial: false,
			},
		},
		{
			in: `2024-04-20T18:39:20.57606443Z stderr F `,
			out: &LogMessage{
				Stream:    Stderr,
				Log:       []byte{},
				IsPartial: false,
			},
		},
		{
			in: `2024-04-20T18:39:20.57606443Z stderr F`,
			out: &LogMessage{
				Stream:    Stderr,
				Log:       nil,
				IsPartial: false,
			},
		},
		{
			in:   `2024-04-20T18:39:20.57606443Z unexpected-stream F [INFO] test log, is full`,
			fail: true,
		},
		{
			in:   ``,
			fail: true,
		},
		// {
		// 	// parse timestamp fail
		// 	in:   `Sun Apr 28 06:19:06 PM CST 2024 stdout F [INFO] test log, is full`,
		// 	fail: true,
		// },
	}

	for _, tc := range cases {
		msg := new(LogMessage)

		err := ParseCRILog([]byte(tc.in), msg)
		if tc.fail && assert.Error(t, err) {
			return
		}
		assert.Nil(t, err)
		assert.Equal(t, tc.out, msg)
	}
}
