// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

// Package diskcache wrap cache/storage functions
package diskcache

type Option func(*config)

type Handler func(*PBData) error

type config struct {
	path        string
	capacity    int
	consumerNum int
	handle      Handler
}

func defaultConfig() *config {
	return &config{
		path:        "./logging",
		capacity:    1024 * 5, // MiB
		consumerNum: 8,
		handle:      HandleFeedIO,
	}
}

func WithPath(s string) Option     { return func(c *config) { c.path = s } }
func WithCapacity(n int) Option    { return func(c *config) { c.capacity = n } }
func WithHandle(fn Handler) Option { return func(c *config) { c.handle = fn } }
func WithConsumerNum(n int) Option { return func(c *config) { c.consumerNum = n } }
