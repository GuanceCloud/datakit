// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package socket

func (i *input) Resume() error {
	i.pause.Store(false)
	return nil
}

func (i *input) Pause() error {
	i.pause.Store(true)
	return nil
}

func (i *input) ElectionEnabled() bool {
	return i.Election
}
