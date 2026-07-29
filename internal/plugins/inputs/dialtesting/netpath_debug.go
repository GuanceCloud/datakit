// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package dialtesting

import (
	"fmt"
	"net"
	"sync"

	dt "github.com/GuanceCloud/cliutils/dialtesting"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/httpapi"
)

var sharedNetPathConcurrency = make(chan struct{}, defaultNetPathMaxConcurrency)

var netPathDebugInput struct {
	sync.RWMutex
	ipt *Input
}

func setNetPathDebugInput(ipt *Input) {
	netPathDebugInput.Lock()
	netPathDebugInput.ipt = ipt
	netPathDebugInput.Unlock()
}

func clearNetPathDebugInput(ipt *Input) {
	netPathDebugInput.Lock()
	if netPathDebugInput.ipt == ipt {
		netPathDebugInput.ipt = nil
	}
	netPathDebugInput.Unlock()
}

func currentNetPathConcurrency() chan struct{} {
	netPathDebugInput.RLock()
	ipt := netPathDebugInput.ipt
	netPathDebugInput.RUnlock()
	if ipt != nil && ipt.netPathConcurrency != nil {
		return ipt.netPathConcurrency
	}
	return sharedNetPathConcurrency
}

func makeNetPathResolvedIPValidator(blockInternal bool, cidrs []string) func(net.IP) error {
	if !blockInternal {
		return nil
	}
	cidrs = append([]string(nil), cidrs...)
	return func(ip net.IP) error {
		if ip == nil {
			return fmt.Errorf("resolved IP is empty")
		}
		internal, err := httpapi.IsInternalHost(ip.String(), cidrs)
		if err != nil {
			return err
		}
		if internal {
			return fmt.Errorf("internal address %s is blocked", ip.String())
		}
		return nil
	}
}

func (ipt *Input) netPathResolvedIPValidator() func(net.IP) error {
	if ipt == nil {
		return nil
	}
	return makeNetPathResolvedIPValidator(
		ipt.DisableInternalNetworkTask,
		ipt.DisabledInternalNetworkCIDRList,
	)
}

func enrichNetPathDebugResult(
	task *dt.NetPathTask,
	tags map[string]string,
	fields map[string]interface{},
) {
	netPathDebugInput.RLock()
	ipt := netPathDebugInput.ipt
	netPathDebugInput.RUnlock()
	if ipt == nil {
		return
	}

	d := newDialer(task, ipt)
	d.seqNumber = 1
	d.enrichPointResult(tags, fields, d.regionName())
}
