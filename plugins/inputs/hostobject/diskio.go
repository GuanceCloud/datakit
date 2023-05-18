// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package hostobject

import (
	"time"
)

type diskIOInfo struct {
	readBytes        uint64
	writeBytes       uint64
	readBytesPerSec  int64
	writeBytesPerSec int64
	ts               time.Time
}

// getDiskIORate get disIO R/W rates.
// ipt.lastDiskIOInfo store last loop data.
// - rate = (nowData - lastData) / duration.
// - If has some error, will send last loop rate.
func (ipt *Input) getDiskIORate() {
	var readBytes uint64
	var writeBytes uint64
	ts := time.Now()

	diskio, err := ipt.diskIOCounters([]string{}...)
	if err != nil {
		l.Errorf("error getting disk io info: %w", err)
		return
	}

	for _, stat := range diskio {
		readBytes += stat.ReadBytes
		writeBytes += stat.WriteBytes
	}

	if ipt.lastDiskIOInfo.readBytes > 0 || ipt.lastDiskIOInfo.writeBytes > 0 {
		deltaTime := ts.Unix() - ipt.lastDiskIOInfo.ts.Unix()
		ipt.lastDiskIOInfo.readBytesPerSec = int64(readBytes-ipt.lastDiskIOInfo.readBytes) / deltaTime
		ipt.lastDiskIOInfo.writeBytesPerSec = int64(writeBytes-ipt.lastDiskIOInfo.writeBytes) / deltaTime
	}
	ipt.lastDiskIOInfo.ts = ts
	ipt.lastDiskIOInfo.readBytes = readBytes
	ipt.lastDiskIOInfo.writeBytes = writeBytes
}
