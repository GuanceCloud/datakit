// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package io

import (
	"encoding/json"
	"fmt"
)

var defPipelinePullMock pipelinePullMock = &prodPipelinePullMock{}

type pipelinePullMock interface {
	getPipelinePull(ts int64) (*PullPipelineReturn, error)
}

type prodPipelinePullMock struct{}

type HTTPError struct {
	ErrCode  string `json:"error_code,omitempty"`
	Err      error  `json:"-"`
	HTTPCode int    `json:"-"`
}

type bodyResp struct {
	*HTTPError
	Message string              `json:"message,omitempty"`
	Content *PullPipelineReturn `json:"content,omitempty"` // 注意与 kodo 中的不一样
}

type PipelineUnit struct {
	Name       string `json:"name"`
	Base64Text string `json:"base64text"`
	Category   string `json:"category"`
}

type PullPipelineReturn struct {
	UpdateTime int64           `json:"update_time"`
	Pipelines  []*PipelineUnit `json:"pipelines"`
}

func (*prodPipelinePullMock) getPipelinePull(ts int64) (*PullPipelineReturn, error) {
	body, err := defaultIO.dw.DatakitPull(fmt.Sprintf("ts=%d", ts))
	if err != nil {
		return nil, err
	}
	var br bodyResp
	err = json.Unmarshal(body, &br)
	if err != nil {
		log.Error(err.Error())
		return nil, err
	}

	return br.Content, err
}
