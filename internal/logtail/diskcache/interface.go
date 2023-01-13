// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

// Package diskcache wrap cache/storage functions
package diskcache

import "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/storage"

type Cache interface {
	RegisterConsumer(uint8, storage.ConsumerFunc)
	RunConsumeWorker() error
	Put(uint8, []byte) error
	Close() error
}
