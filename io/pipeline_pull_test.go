package io

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io/dataway"
)

var debugPipelinePullData *dataway.PullPipelineReturn

type debugPipelinePullMock struct{}

func (*debugPipelinePullMock) getPipelinePull(ts int64) (*dataway.PullPipelineReturn, error) {
	return debugPipelinePullData, nil
}

// go test -v -timeout 30s -run ^TestPullPipeline$ gitlab.jiagouyun.com/cloudcare-tools/datakit/io
func TestPullPipeline(t *testing.T) {
	cases := []struct {
		Name      string
		LocalTS   int64
		Pipelines *dataway.PullPipelineReturn
		Expect    *struct {
			mFiles     map[string]string
			updateTime int64
		}
	}{
		{
			Name:    "has_data",
			LocalTS: 0,
			Pipelines: &dataway.PullPipelineReturn{
				UpdateTime: 1641796675,
				Pipelines: []*dataway.PipelineUnit{
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
			Pipelines: &dataway.PullPipelineReturn{
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
