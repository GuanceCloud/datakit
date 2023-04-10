// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package election

type ElectionOption func(c *candidate)

func WithElectionEnabled(on bool) ElectionOption {
	return func(c *candidate) {
		c.enabled = on
	}
}

func WithID(id string) ElectionOption {
	return func(c *candidate) {
		c.id = id
	}
}

func WithNamespace(ns string) ElectionOption {
	return func(c *candidate) {
		c.namespace = ns
	}
}

func WithPuller(p Puller) ElectionOption {
	return func(c *candidate) {
		c.puller = p
	}
}
