// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

// Package parameter defines the parsing parameters.
package parameter

import (
	"errors"
	"fmt"

	"github.com/GuanceCloud/cliutils/pprofparser/domain/languages"
	"github.com/GuanceCloud/cliutils/pprofparser/tools/jsontoolkit"
)

const (
	FromTrace   = "trace"
	FromProfile = "profile"
)

const (
	FilterBySpanTime = "spanTime"
	FilterByFull     = "full"
)

const (
	MinTimestampMicro = 1640966400000000
	MaxTimestampMicro = 2147483647000000
	MinTimestampNano  = 1640966400000000000
)

type BaseRequestParam struct {
	WorkspaceUUID string `json:"workspace_uuid" binding:"required"`
}

type WithTypeRequestParam struct {
	BaseRequestParam
	Type string `json:"type" binding:"required"`
}

type Profile struct {
	Language          languages.Lang `json:"language" binding:"required"`
	EsDocID           string         `json:"__docid"`
	ProfileID         string         `json:"profile_id" binding:"required"`
	ProfileStart      interface{}    `json:"profile_start" binding:"required"` // , min=1640966400000000000
	ProfileEnd        interface{}    `json:"profile_end" binding:"required"`   // , min=1640966400000000000
	internalProfStart int64
	internalProfEnd   int64
}

func (p *Profile) StartTime() (int64, error) {
	if p.internalProfStart > 0 {
		return p.internalProfStart, nil
	}
	start, err := jsontoolkit.IFaceCast2Int64(p.ProfileStart)
	if err != nil {
		return 0, err
	}
	if start <= 0 {
		return 0, errors.New("illegal profile_start parameter")
	}
	p.internalProfStart = start
	return start, err
}

func (p *Profile) EndTime() (int64, error) {
	if p.internalProfEnd > 0 {
		return p.internalProfEnd, nil
	}
	end, err := jsontoolkit.IFaceCast2Int64(p.ProfileEnd)
	if err != nil {
		return 0, err
	}
	if end <= 0 {
		return 0, errors.New("illegal profile_end parameter")
	}
	p.internalProfEnd = end
	return end, err
}

type Span struct {
	TraceID   string `json:"trace_id"`
	SpanID    string `json:"span_id"`
	SpanStart int64  `json:"span_start"`
	SpanEnd   int64  `json:"span_end"`
}

type SummaryParam struct {
	BaseRequestParam
	Span
	From     string     `json:"from"`
	FilterBy string     `json:"filter_by"`
	Profiles []*Profile `json:"profiles" binding:"required"`
}

type ParseParam struct {
	WithTypeRequestParam
	Profile
}

type LookupParam struct {
	WithTypeRequestParam
	Span
	FilterBy string     `json:"filter_by"`
	Profiles []*Profile `json:"profiles" binding:"required"`
}

type DownloadParam struct {
	BaseRequestParam
	Profiles []*Profile `json:"profiles" binding:"required"`
}

// VerifyLanguage 校验多个profiles中的language是否相同
func VerifyLanguage(profiles []*Profile) (languages.Lang, error) {

	if len(profiles) == 0 {
		return languages.Unknown, fmt.Errorf("empty profiles param")
	}

	//lang := profiles[0].Language

	// 对比是否是同一个language
	//for i := 1; i < len(profiles); i++ {
	//	if !lang.Is(profiles[i].Language) {
	//		return languages.Unknown, fmt.Errorf("the languages are not same")
	//	}
	//}

	return profiles[0].Language, nil
}
