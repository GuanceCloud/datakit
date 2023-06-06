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
	getPipelinePull(ts, relationTS int64) (*pullPipelineReturn, error)
}

type prodPipelinePullMock struct{}

//------------------------------------------------------------------------------
// copied from kodo project

type pipelineUnit struct {
	Name       string `json:"name"`
	Base64Text string `json:"base64text"`
	Category   string `json:"category"`
	AsDefault  bool   `json:"asDefault"`
}

type pipelineRelation struct {
	Category string `json:"category"`
	Source   string `json:"source"`
	Name     string `json:"name"`
}

type pullPipelineReturn struct {
	UpdateTime int64           `json:"update_time"`
	Pipelines  []*pipelineUnit `json:"pipelines"`

	RelationUpdateTime int64               `json:"relation_update_time"`
	Relation           []*pipelineRelation `json:"relation"`
}

type pullResp struct {
	RemotePL *pullPipelineReturn `json:"remote_pipelines"`
}

//------------------------------------------------------------------------------

func (*prodPipelinePullMock) getPipelinePull(ts int64, relationTS int64) (*pullPipelineReturn, error) {
	body, err := defIO.dw.Pull(fmt.Sprintf("remote_pipelines=true&ts=%d&relation_ts=%d", ts, relationTS))
	if err != nil {
		return nil, err
	}
	var br pullResp
	err = json.Unmarshal(body, &br)
	if err != nil {
		log.Error(err.Error())
		return nil, err
	}

	return br.RemotePL, err
}
