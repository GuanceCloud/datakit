// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package testutils

import "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/datakit"

type taggerMock struct {
	hostTags, electionTags map[string]string
}

// NewTaggerHost referred from function "setupGlobalTags" in file internal/config/conf.go.
func NewTaggerHost() *taggerMock {
	return &taggerMock{
		hostTags: map[string]string{
			"host": "HOST",
		},
		electionTags: map[string]string{
			"host": "HOST",
		},
	}
}

// NewTaggerElection referred from function "setupGlobalTags" in file internal/config/conf.go.
func NewTaggerElection() *taggerMock {
	return &taggerMock{
		hostTags: map[string]string{
			"host": "HOST",
		},
		electionTags: map[string]string{
			"election": "1",
		},
	}
}

func (m *taggerMock) HostTags() map[string]string {
	return m.hostTags
}

func (m *taggerMock) ElectionTags() map[string]string {
	return m.electionTags
}

// DefaultMockTagger How to use?
//
//	return &Input{
//		tagger:          testutils.DefaultMockTagger(),
//	}
//
//	func (ipt *Input) setup() {
//		if ipt.Election {
//	        // got: map[string]string{"election": "TRUE"}
//			ipt.mergedTags = inputs.MergeTags(ipt.tagger.ElectionTags(), ipt.Tags, "")
//		} else {
//	        // got: map[string]string{"host": "HOST"}
//			ipt.mergedTags = inputs.MergeTags(ipt.tagger.HostTags(), ipt.Tags, "")
//		}
//	}
func DefaultMockTagger() datakit.GlobalTagger {
	return &mockTaggerImpl{}
}

type mockTaggerImpl struct{}

func (g *mockTaggerImpl) HostTags() map[string]string {
	return map[string]string{"host": "HOST"}
}

func (g *mockTaggerImpl) ElectionTags() map[string]string {
	return map[string]string{"election": "TRUE"}
}
