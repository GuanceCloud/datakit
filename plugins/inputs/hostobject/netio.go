// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package hostobject

import (
	"time"
)

type netIOInfo struct {
	recvBytes       uint64
	sendBytes       uint64
	recvBytesPerSec int64
	sendBytesPerSec int64
	ts              time.Time
}

// getNetIORate get netIO R/W rates.
// ipt.lastNetIOInfo store last loop data.
// - rate = (nowData - lastData) / duration.
// - If has some error, will send last loop rate.
func (ipt *Input) getNetIORate(infos []*NetInfo) {
	var recvBytes uint64
	var sendBytes uint64
	ts := time.Now()
	netio, err := ipt.netIOCounters(true)
	if err != nil {
		l.Errorf("error getting net io info: %w", err)
	}
	for _, stat := range netio {
		if filterInterface(stat.Name, infos) {
			recvBytes += stat.BytesRecv
			sendBytes += stat.BytesSent
		}
	}

	if ipt.lastNetIOInfo.recvBytes > 0 || ipt.lastNetIOInfo.sendBytes > 0 {
		deltaTime := ts.Unix() - ipt.lastNetIOInfo.ts.Unix()
		ipt.lastNetIOInfo.recvBytesPerSec = int64(recvBytes-ipt.lastNetIOInfo.recvBytes) / deltaTime
		ipt.lastNetIOInfo.sendBytesPerSec = int64(sendBytes-ipt.lastNetIOInfo.sendBytes) / deltaTime
	}
	ipt.lastNetIOInfo.ts = ts
	ipt.lastNetIOInfo.recvBytes = recvBytes
	ipt.lastNetIOInfo.sendBytes = sendBytes
}

func filterInterface(name string, infos []*NetInfo) bool {
	for _, info := range infos {
		if name == info.Name {
			return true
		}
	}
	return false
}
