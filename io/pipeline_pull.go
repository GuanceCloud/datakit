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
func PullPipeline(ts int64) (mFiles map[string]string, updateTime int64, err error) {
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

	updateTime = pulledStruct.UpdateTime
	mFiles = make(map[string]string)
	for _, v := range pulledStruct.Pipelines {
		bys, e := base64.StdEncoding.DecodeString(v.Base64Text)
		if e != nil {
			err = e
			return nil, 0, err
		}
		mFiles[v.Name] = string(bys)
	}

	return
}
