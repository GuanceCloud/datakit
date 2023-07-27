// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package testutils

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
