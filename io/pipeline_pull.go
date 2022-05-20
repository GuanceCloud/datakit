// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package io

import (
	"encoding/base64"
	"fmt"
)

// PullPipeline returns name/text, updateTime, err.
func PullPipeline(ts int64) (mFiles map[string]map[string]string, updateTime int64, err error) {
	defer func() {
		if err := recover(); err != nil {
			if log != nil {
				log.Error(err)
			}
		}
	}()
	pulledStruct, err := defPipelinePullMock.getPipelinePull(ts)
	if err != nil {
		return nil, 0, err
	}
	if pulledStruct == nil {
		err = fmt.Errorf("got nil")
		return nil, 0, err
	}

	mFiles, updateTime, err = parsePipelinePullStruct(pulledStruct)
	return
}

func parsePipelinePullStruct(pulledStruct *pullPipelineReturn) (
	map[string]map[string]string, int64, error) {
	mFiles := make(map[string]map[string]string)
	for _, v := range pulledStruct.Pipelines {
		bys, err := base64.StdEncoding.DecodeString(v.Base64Text)
		if err != nil {
			return nil, 0, err
		}

		if val, ok := mFiles[v.Category]; ok {
			val[v.Name] = string(bys)
		} else {
			mf := make(map[string]string)
			mf[v.Name] = string(bys)
			mFiles[v.Category] = mf
		}
	}
	return mFiles, pulledStruct.UpdateTime, nil
}
