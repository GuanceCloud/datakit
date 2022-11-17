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

var debugPipelinePullData *pullPipelineReturn

type debugPipelinePullMock struct{}

// Make sure debugPipelinePullMock implements the pipelinePullMock interface
var _ pipelinePullMock = new(debugPipelinePullMock)

func (*debugPipelinePullMock) getPipelinePull(ts, relationTS int64) (*pullPipelineReturn, error) {
	return debugPipelinePullData, nil
}

// go test -v -timeout 30s -run ^TestPullPipeline$ gitlab.jiagouyun.com/cloudcare-tools/datakit/io
func TestPullPipeline(t *testing.T) {
	cases := []struct {
		Name       string
		LocalTS    int64
		RelationTS int64
		Pipelines  *pullPipelineReturn
		Expect     *struct {
			mFiles, plRelation map[string]map[string]string
			defaultPl          map[string]string
			updateTime         int64
			relationUpdateTime int64
		}
	}{
		{
			Name:    "has_data",
			LocalTS: 0,
			Pipelines: &pullPipelineReturn{
				UpdateTime: 1641796675,
				Pipelines: []*pipelineUnit{
					{
						Name:       "123.p",
						Base64Text: base64.StdEncoding.EncodeToString([]byte("text1")),
					},
					{
						Name:       "456.p",
						Base64Text: base64.StdEncoding.EncodeToString([]byte("text2")),
						AsDefault:  true,
					},
				},
				Relation: []*pipelineRelation{
					{
						Source: "a1",
						Name:   "abc",
					},
					{
						Source: "a2",
						Name:   "abc",
					},
				},
			},
			Expect: &struct {
				mFiles, plRelation map[string]map[string]string
				defaultPl          map[string]string
				updateTime         int64
				relationUpdateTime int64
			}{
				mFiles: map[string]map[string]string{
					"": {
						"123.p": "text1",
						"456.p": "text2",
					},
				},
				plRelation: map[string]map[string]string{
					"": {
						"a1": "abc",
						"a2": "abc",
					},
				},
				defaultPl: map[string]string{
					"": "456.p",
				},
				updateTime: 1641796675,
			},
		},
		{
			Name:    "no_data",
			LocalTS: 1641796675,
			Pipelines: &pullPipelineReturn{
				UpdateTime: -1,
			},
			Expect: &struct {
				mFiles, plRelation map[string]map[string]string
				defaultPl          map[string]string
				updateTime         int64
				relationUpdateTime int64
			}{
				mFiles:     map[string]map[string]string{},
				plRelation: map[string]map[string]string{},
				defaultPl:  map[string]string{},
				updateTime: -1,
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.Name, func(t *testing.T) {
			debugPipelinePullData = tc.Pipelines
			mFiles, plRelation, defaultPl, updateTime, relationUpdateTime, err := PullPipeline(tc.LocalTS, tc.RelationTS)
			assert.NoError(t, err)
			assert.Equal(t, tc.Expect.mFiles, mFiles)
			assert.Equal(t, tc.Expect.updateTime, updateTime)
			assert.Equal(t, tc.Expect.defaultPl, defaultPl)
			assert.Equal(t, tc.Expect.plRelation, plRelation)
			assert.Equal(t, tc.Expect.relationUpdateTime, relationUpdateTime)
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
		pipelines *pullPipelineReturn
		expect    *struct {
			mfiles, relation   map[string]map[string]string
			defaultPl          map[string]string
			updateTime         int64
			relationUpdateTime int64
			err                error
		}
	}{
		{
			name: "normal",
			pipelines: &pullPipelineReturn{
				UpdateTime: 1653020819,
				Pipelines: []*pipelineUnit{
					{
						Name:       "123.p",
						Base64Text: base64.StdEncoding.EncodeToString([]byte("text123")),
						Category:   "logging",
					},
					{
						Name:       "1234.p",
						Base64Text: base64.StdEncoding.EncodeToString([]byte("text1234")),
						Category:   "logging",
						AsDefault:  true,
					},
					{
						Name:       "456.p",
						Base64Text: base64.StdEncoding.EncodeToString([]byte("text456")),
						Category:   "metric",
					},
				},
				Relation: []*pipelineRelation{
					{
						Source: "a1",
						Name:   "abc",
					},
					{
						Source: "a2",
						Name:   "abc",
					},
				},
			},
			expect: &struct {
				mfiles, relation   map[string]map[string]string
				defaultPl          map[string]string
				updateTime         int64
				relationUpdateTime int64
				err                error
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
				relation: map[string]map[string]string{
					"": {
						"a1": "abc",
						"a2": "abc",
					},
				},
				defaultPl: map[string]string{
					"logging": "1234.p",
				},
				updateTime: 1653020819,
			},
		},
		{
			name: "repeat",
			pipelines: &pullPipelineReturn{
				UpdateTime: 1653020819,
				Pipelines: []*pipelineUnit{
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
				mfiles, relation   map[string]map[string]string
				defaultPl          map[string]string
				updateTime         int64
				relationUpdateTime int64
				err                error
			}{
				mfiles: map[string]map[string]string{
					"logging": {
						"123.p": "text1234",
					},
					"metric": {
						"456.p": "text456",
					},
				},
				relation:   map[string]map[string]string{},
				defaultPl:  map[string]string{},
				updateTime: 1653020819,
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			mFiles, plRelation, defaultPl, updateTime, relationUpdateTime, err := parsePipelinePullStruct(tc.pipelines)
			assert.Equal(t, tc.expect.mfiles, mFiles)
			assert.Equal(t, tc.expect.updateTime, updateTime)
			assert.Equal(t, tc.expect.err, err)
			assert.Equal(t, tc.expect.defaultPl, defaultPl)
			assert.Equal(t, tc.expect.relation, plRelation)
			assert.Equal(t, tc.expect.relationUpdateTime, relationUpdateTime)
		})
	}
}
