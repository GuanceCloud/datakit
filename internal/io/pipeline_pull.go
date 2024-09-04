// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package io

import (
	"encoding/base64"
	"fmt"

	"github.com/GuanceCloud/cliutils/point"
)

// PullPipeline returns name/text, updateTime, err.
func PullPipeline(ts, relaTS int64) (mFiles, plRelation map[point.Category]map[string]string,
	defaultPl map[point.Category]string, updateTime int64, relationTS int64, err error,
) {
	defer func() {
		if err := recover(); err != nil {
			if log != nil {
				log.Error(err)
			}
		}
	}()
	pulledStruct, err := defPipelinePullMock.getPipelinePull(ts, relaTS)
	if err != nil {
		return nil, nil, nil, 0, 0, err
	}
	if pulledStruct == nil {
		err = fmt.Errorf("got nil")
		return nil, nil, nil, 0, 0, err
	}

	mFiles, plRelation, defaultPl, updateTime, relationTS, err = parsePipelinePullStruct(pulledStruct)
	return
}

func parsePipelinePullStruct(pulledStruct *pullPipelineReturn) (
	map[point.Category]map[string]string, map[point.Category]map[string]string,
	map[point.Category]string, int64, int64, error,
) {
	mFiles := make(map[point.Category]map[string]string)
	defaultPl := make(map[point.Category]string)
	for _, v := range pulledStruct.Pipelines {
		bys, err := base64.StdEncoding.DecodeString(v.Base64Text)
		if err != nil {
			return nil, nil, nil, 0, 0, err
		}

		cat := point.CatString(v.Category)
		if v.AsDefault {
			defaultPl[cat] = v.Name
		}

		if val, ok := mFiles[cat]; ok {
			val[v.Name] = string(bys)
		} else {
			mf := make(map[string]string)
			mf[v.Name] = string(bys)
			mFiles[cat] = mf
		}
	}

	plRelation := make(map[point.Category]map[string]string)
	for _, v := range pulledStruct.Relation {
		cat := point.CatString(v.Category)
		if m, ok := plRelation[cat]; !ok {
			plRelation[cat] = map[string]string{
				v.Source: v.Name,
			}
		} else {
			m[v.Source] = v.Name
		}
	}

	return mFiles, plRelation, defaultPl, pulledStruct.UpdateTime,
		pulledStruct.RelationUpdateTime, nil
}
