// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

// Package failcache implements local cache for datakit.
package failcache

import (
	"github.com/GuanceCloud/cliutils/diskcache"
)

type Cache interface {
	// NOTE: reuse callback in diskcache to keep interface ok
	// it's better to define Get as
	//   Get() []byte
	Get(diskcache.Fn) error
	Put([]byte) error
	Close() error
}
