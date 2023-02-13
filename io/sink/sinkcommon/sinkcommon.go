// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

// Package sinkcommon contains sink common declaration
package sinkcommon

import (
	"time"

	"github.com/GuanceCloud/cliutils/logger"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io/point"
)

//----------------------------------------------------------------------

type JSONPoint struct {
	Measurement string                 `json:"measurement"`    // measurement name of the point.
	Tags        map[string]string      `json:"tags,omitempty"` // tags associated with the point.
	Fields      map[string]interface{} `json:"fields"`         // the fields for the point.
	Time        time.Time              `json:"time,omitempty"` // timestamp for the point.
}

type SinkInfo struct {
	ID         string // MD5.
	IDStr      string // MD5 origin string.
	CreateID   string
	Categories []string
}

type ISink interface {
	GetInfo() *SinkInfo
	LoadConfig(mConf map[string]interface{}) error
	Write(category string, pts []*point.Point) error
}

type SinkUnsupportError struct {
	err string
}

func (e *SinkUnsupportError) Error() string {
	return e.err
}

//----------------------------------------------------------------------

type SinkCreator func() ISink

func AddCreator(creatorID string, creator SinkCreator) {
	if _, ok := SinkImplCreator[creatorID]; ok {
		l.Fatalf("sink %s exist(from datakit)", creatorID)
	}
	SinkImplCreator[creatorID] = creator
}

func AddImpl(sk ISink) {
	// check errors
	if len(sk.GetInfo().CreateID) == 0 ||
		len(sk.GetInfo().ID) == 0 ||
		len(sk.GetInfo().IDStr) == 0 ||
		len(sk.GetInfo().Categories) == 0 {
		l.Error("sink check failed: should not be empty.")
	}

	SinkImpls = append(SinkImpls, sk)
}

//----------------------------------------------------------------------

const packageName = "sinkcommon"

var (
	SinkImplCreator = make(map[string]SinkCreator)
	SinkImpls       = []ISink{}
	SinkCategoryMap = make(map[string][]ISink)

	l = logger.DefaultSLogger(packageName)
)

func init() { //nolint:gochecknoinits
	l = logger.SLogger(packageName)
}

//----------------------------------------------------------------------
