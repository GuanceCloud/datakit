// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

// Package pipeline implement pipeline.
package pipeline

import (
	"github.com/GuanceCloud/cliutils/pipeline/manager"
	"github.com/GuanceCloud/cliutils/pipeline/manager/relation"
	"github.com/GuanceCloud/cliutils/pipeline/offload"
	"github.com/GuanceCloud/cliutils/pipeline/ptinput/funcs"
	"github.com/GuanceCloud/cliutils/pipeline/ptinput/ipdb/geoip"
	"github.com/GuanceCloud/cliutils/pipeline/ptinput/ipdb/iploc"
	"github.com/GuanceCloud/cliutils/pipeline/ptinput/plmap"
	"github.com/GuanceCloud/cliutils/pipeline/ptinput/refertable"
	"github.com/GuanceCloud/cliutils/pipeline/stats"
)

func InitLog() {
	// pipeline scripts manager
	manager.InitLog()
	// scripts relation
	relation.InitLog()

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
