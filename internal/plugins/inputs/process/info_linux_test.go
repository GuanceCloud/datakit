// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

//go:build linux
// +build linux

package process

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseScopePathForCgroup(t *testing.T) {
	cases := []struct {
		in   string
		out  string
		fail bool
	}{
		{
			in:  "0::/system.slice/containerd.service",
			out: "",
		},
		{
			in:  "0::/system.slice/docker-89ac90181a86fdea23d87ab993ceb698aa61528123c23135ef43fe5246a1b11e.scope/system.slice/containerd.service",
			out: "89ac90181a86fdea23d87ab993ceb698aa61528123c23135ef43fe5246a1b11e",
		},
		{
			in:  "0::/kubelet.slice/kubelet-kubepods.slice/kubelet-kubepods-burstable.slice/kubelet-kubepods-burstable-podb1e89a2fcc688644de3eb95aee8025ce.slice/cri-containerd-ee3fc60554cd9eefdd745a10dda84e041a7883c0381d067e76185d342de64759.scope",
			out: "ee3fc60554cd9eefdd745a10dda84e041a7883c0381d067e76185d342de64759",
		},
		{
			in:  "0::/system.slice/docker-b362fa680d463b43611d956fcaab31640a49de2d11bf2604c2af55570865e46b.scope",
			out: "b362fa680d463b43611d956fcaab31640a49de2d11bf2604c2af55570865e46b",
		},
		{
			in:  "0::/system.slice/docker-b362fa680d463b43611d956fcaab31640a49de2d11bf2604c2af55570865e46b.scope\n",
			out: "b362fa680d463b43611d956fcaab31640a49de2d11bf2604c2af55570865e46b",
		},
		{
			in:   "0::/system.slice/docker-b362fa680d463b43611d956fcaab31640a49de2d11bf2604c2af55570865e46b-.scope",
			fail: true,
		},
		{
			in:   "0::/system.slice/docker#####################b362fa680d463b43611d956fcaab31640a49de2d11bf2604c2af55570865e46b.scope",
			fail: true,
		},
		{
			in:   "0:",
			fail: true,
		},
		{
			in:  "1::/system.slice/containerd.service\n0::/system.slice/containerd.service",
			out: "",
		},
		{
			in:  "1::/faker.slice/docker-b362fa680d463b43611d956fcaab31640a49de2d11bf2604c2af55570865e46b.scope\n0::/system.slice/docker-b362fa680d463b43611d956fcaab31640a49de2d11bf2604c2af55570865e46b.scope",
			out: "b362fa680d463b43611d956fcaab31640a49de2d11bf2604c2af55570865e46b",
		},
		{
			in:  "1::/faker.slice/docker-b362fa680d463b43611d956fcaab31640a49de2d11bf2604c2af55570865e46b.scope\n0::/system.slice/docker-b362fa680d463b43611d956fcaab31640a49de2d11bf2604c2af55570865e46b.scope\n",
			out: "b362fa680d463b43611d956fcaab31640a49de2d11bf2604c2af55570865e46b",
		},
	}

	for _, tc := range cases {
		res, err := parseScopePathForCgroup(tc.in)
		if tc.fail && assert.Error(t, err) {
			continue
		}

		assert.NoError(t, err)
		assert.Equal(t, tc.out, res)
	}
}
