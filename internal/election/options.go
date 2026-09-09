// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package election

import "time"

type Clock interface {
	Now() time.Time
}

type realClock struct{}

func (realClock) Now() time.Time { return time.Now() }

type option struct {
	enabled       bool
	namespace, id string
	nodeWhitelist []string
	puller        Puller
	provider      Provider
	clock         Clock
}

type ElectionOption func(opt *option)

func WithElectionEnabled(on bool) ElectionOption {
	return func(opt *option) {
		opt.enabled = on
	}
}

func WithElectionWhitelist(whitelist []string) ElectionOption {
	return func(opt *option) {
		opt.nodeWhitelist = whitelist
	}
}

func WithID(id string) ElectionOption {
	return func(opt *option) {
		opt.id = id
	}
}

func WithNamespace(ns string) ElectionOption {
	return func(opt *option) {
		opt.namespace = ns
	}
}

func WithDatawayPuller(p Puller) ElectionOption {
	return func(opt *option) {
		opt.puller = p
		opt.provider = ProviderDataway
	}
}

func WithPuller(provider Provider, p Puller) ElectionOption {
	return func(opt *option) {
		opt.puller = p
		opt.provider = provider
	}
}
