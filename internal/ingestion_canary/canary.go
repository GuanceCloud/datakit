// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

// Package ingestioncanary implements the ingestion canary.
package ingestioncanary

type Canary struct {
	Name          string // The name of the canary.
	TestType      string // The test type of the canary.
	EnableMetrics bool   // Whether to record Prometheus metrics.
}

// New creates a new canary.
func New(opts ...Option) *Canary {
	c := &Canary{}

	for _, opt := range opts {
		opt(c)
	}
	return c
}

type Option func(*Canary)

func WithName(name string) Option {
	return func(c *Canary) {
		c.Name = name
	}
}

func WithTestType(testType string) Option {
	return func(c *Canary) {
		c.TestType = testType
	}
}

func WithEnableMetrics(enable bool) Option {
	return func(c *Canary) {
		c.EnableMetrics = enable
	}
}
