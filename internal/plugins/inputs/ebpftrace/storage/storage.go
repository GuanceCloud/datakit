// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

// Package storage cache espan meta and original data
package storage

import (
	"github.com/GuanceCloud/cliutils/point"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs/ebpftrace/espan"
)

type DB interface {
	Put(pts []*point.Point) error
	Get() ([]*point.Point, error)

	PutMetaList(*espan.SpanMetaList) error
	GetMetaList() (*espan.SpanMetaList, error)

	Sync() error

	Drop() error

	Close() error
}
