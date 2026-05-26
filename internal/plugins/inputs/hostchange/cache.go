// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

//go:build !windows
// +build !windows

package hostchange

import (
	"fmt"
	"os"
)

type ConfigCache struct {
	cacheDir string
}

func NewConfigCache(cacheDir string) (*ConfigCache, error) {
	// Ensure cache directory exists
	if err := os.MkdirAll(cacheDir, 0o750); err != nil {
		return nil, fmt.Errorf("failed to create cache directory: %w", err)
	}

	cache := &ConfigCache{
		cacheDir: cacheDir,
	}

	return cache, nil
}
