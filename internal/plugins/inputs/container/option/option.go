// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

// Package option wraps collect options for container and kubernetes.
package option

type CollectOption func(*Option)

type Option struct {
	OnlyElection bool
	Paused       bool
	NodeLocal    bool
}

func DefaultOption() *Option {
	return &Option{}
}

func WithOnlyElection(b bool) CollectOption { return func(c *Option) { c.OnlyElection = b } }
func WithPaused(b bool) CollectOption       { return func(c *Option) { c.Paused = b } }
func WithNodeLocal(b bool) CollectOption    { return func(c *Option) { c.NodeLocal = b } }
