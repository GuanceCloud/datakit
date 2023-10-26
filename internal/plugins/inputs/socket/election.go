// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package socket

import (
	"fmt"
	"time"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs"
)

func (i *input) Resume() error {
	return i.trigger(false)
}

func (i *input) Pause() error {
	return i.trigger(true)
}

func (i *input) trigger(off bool) error {
	tick := time.NewTicker(inputs.ElectionPauseTimeout)
	defer tick.Stop()

	select {
	case i.pauseCh <- off:
		return nil
	case <-tick.C:
		return fmt.Errorf("off(%v) %q timeout", off, inputName)
	}
}
