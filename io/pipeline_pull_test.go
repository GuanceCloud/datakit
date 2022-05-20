// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package io

import (
	"encoding/base64"
	"testing"

	"github.com/stretchr/testify/assert"
)

var debugPipelinePullData *PullPipelineReturn

type debugPipelinePullMock struct{}

func (*debugPipelinePullMock) getPipelinePull(ts int64) (*PullPipelineReturn, error) {
	return debugPipelinePullData, nil
}

// go test -v -timeout 30s -run ^TestPullPipeline$ gitlab.jiagouyun.com/cloudcare-tools/datakit/io
func TestPullPipeline(t *testing.T) {
	cases := []struct {
		Name      string
		LocalTS   int64
		Pipelines *PullPipelineReturn
		Expect    *struct {
			mFiles     map[string]string
			updateTime int64
		}
	}{
		{
			Name:    "has_data",
			LocalTS: 0,
			Pipelines: &PullPipelineReturn{
				UpdateTime: 1641796675,
				Pipelines: []*PipelineUnit{
					{
						Name:       "123.p",
						Base64Text: "dGV4dDE=",
					},
					{
						Name:       "456.p",
						Base64Text: "dGV4dDI=",
					},
				},
			},
			Expect: &struct {
				mFiles     map[string]string
				updateTime int64
			}{
				mFiles: map[string]string{
					"123.p": "text1",
					"456.p": "text2",
				},
				updateTime: 1641796675,
			},
		},
		{
			Name:    "no_data",
			LocalTS: 1641796675,
			Pipelines: &PullPipelineReturn{
				UpdateTime: -1,
			},
			Expect: &struct {
				mFiles     map[string]string
				updateTime int64
			}{
				mFiles:     map[string]string{},
				updateTime: -1,
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.Name, func(t *testing.T) {
			debugPipelinePullData = tc.Pipelines
			mFiles, updateTime, err := PullPipeline(tc.LocalTS)
			assert.NoError(t, err)
			assert.Equal(t, tc.Expect.mFiles, mFiles)
			assert.Equal(t, tc.Expect.updateTime, updateTime)
		})
	}
}

func init() { //nolint:gochecknoinits
	defPipelinePullMock = &debugPipelinePullMock{}
}

// go test -v -timeout 30s -run ^TestParsePipelinePullStruct$ gitlab.jiagouyun.com/cloudcare-tools/datakit/io
func TestParsePipelinePullStruct(t *testing.T) {
	cases := []struct {
		name      string
		pipelines *PullPipelineReturn
		expect    *struct {
			mfiles     map[string]map[string]string
			updateTime int64
			err        error
		}
	}{
		{
			name: "normal",
			pipelines: &PullPipelineReturn{
				UpdateTime: 1653020819,
				Pipelines: []*PipelineUnit{
					{
						Name:       "123.p",
						Base64Text: base64.StdEncoding.EncodeToString([]byte("text123")),
						Category:   "logging",
					},
					{
						Name:       "1234.p",
						Base64Text: base64.StdEncoding.EncodeToString([]byte("text1234")),
						Category:   "logging",
					},
					{
						Name:       "456.p",
						Base64Text: base64.StdEncoding.EncodeToString([]byte("text456")),
						Category:   "metric",
					},
				},
			},
			expect: &struct {
				mfiles     map[string]map[string]string
				updateTime int64
				err        error
			}{
				mfiles: map[string]map[string]string{
					"logging": {
						"123.p":  "text123",
						"1234.p": "text1234",
					},
					"metric": {
						"456.p": "text456",
					},
				},
				updateTime: 1653020819,
			},
		},
		{
			name: "repeat",
			pipelines: &PullPipelineReturn{
				UpdateTime: 1653020819,
				Pipelines: []*PipelineUnit{
					{
						Name:       "123.p",
						Base64Text: base64.StdEncoding.EncodeToString([]byte("text123")),
						Category:   "logging",
					},
					{
						Name:       "123.p",
						Base64Text: base64.StdEncoding.EncodeToString([]byte("text1234")),
						Category:   "logging",
					},
					{
						Name:       "456.p",
						Base64Text: base64.StdEncoding.EncodeToString([]byte("text456")),
						Category:   "metric",
					},
				},
			},
			expect: &struct {
				mfiles     map[string]map[string]string
				updateTime int64
				err        error
			}{
				mfiles: map[string]map[string]string{
					"logging": {
						"123.p": "text1234",
					},
					"metric": {
						"456.p": "text456",
					},
				},
				updateTime: 1653020819,
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			mFiles, updateTime, err := parsePipelinePullStruct(tc.pipelines)
			assert.Equal(t, tc.expect.mfiles, mFiles)
			assert.Equal(t, tc.expect.updateTime, updateTime)
			assert.Equal(t, tc.expect.err, err)
		})
	}
}
