// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

// Package pipeline implement pipeline.
package pipeline

import (
	"github.com/GuanceCloud/pipeline-go/manager"
	"github.com/GuanceCloud/pipeline-go/offload"
	"github.com/GuanceCloud/pipeline-go/ptinput/funcs"
	"github.com/GuanceCloud/pipeline-go/ptinput/ipdb/geoip"
	"github.com/GuanceCloud/pipeline-go/ptinput/ipdb/iploc"
	"github.com/GuanceCloud/pipeline-go/ptinput/plmap"
	"github.com/GuanceCloud/pipeline-go/ptinput/refertable"
	"github.com/GuanceCloud/pipeline-go/stats"
)

func InitLog() {
	// pipeline scripts manager
	manager.InitLog()

	// pipeline offload
	offload.InitLog()

	// all ptinputs's package

	// inner pl functions
	funcs.InitLog()
	// ip db
	iploc.InitLog()
	geoip.InitLog()
	// pipeline map
	plmap.InitLog()
	// refertable
	refertable.InitLog()

	// stats
	stats.InitLog()
}
