// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package election

type option struct {
	enabled       bool
	namespace, id string
	puller        Puller
	mode          electionMode
}

type ElectionOption func(opt *option)

func WithElectionEnabled(on bool) ElectionOption {
	return func(opt *option) {
		opt.enabled = on
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
		opt.mode = modeDataway
	}
}

func WithOperatorPuller(p Puller) ElectionOption {
	return func(opt *option) {
		opt.puller = p
		opt.mode = modeOperator
	}
}

type electionMode int

const (
	modeDataway electionMode = iota + 1
	modeOperator
)
