// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package rabbitmq

import (
	"time"

	"github.com/GuanceCloud/cliutils/point"
	dkio "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/io"
)

type node struct {
	Name              string
	DiskFreeAlarm     bool  `json:"disk_free_alarm"`
	MemAlarm          bool  `json:"mem_alarm"`
	Running           bool  `json:"running"`
	DiskFree          int64 `json:"disk_free"`
	DiskFreeLimit     int64 `json:"disk_free_limit"`
	FdTotal           int64 `json:"fd_total"`
	FdUsed            int64 `json:"fd_used"`
	MemLimit          int64 `json:"mem_limit"`
	MemUsed           int64 `json:"mem_used"`
	ProcTotal         int64 `json:"proc_total"`
	ProcUsed          int64 `json:"proc_used"`
	RunQueue          int64 `json:"run_queue"`
	SocketsTotal      int64 `json:"sockets_total"`
	SocketsUsed       int64 `json:"sockets_used"`
	Uptime            int64 `json:"uptime"`
	MnesiaDiskTxCount int64 `json:"mnesia_disk_tx_count"`
	MnesiaRAMTxCount  int64 `json:"mnesia_ram_tx_count"`
	GcNum             int64 `json:"gc_num"`
	IoWriteBytes      int64 `json:"io_write_bytes"`
	IoReadBytes       int64 `json:"io_read_bytes"`
	GcBytesReclaimed  int64 `json:"gc_bytes_reclaimed"`

	IoWriteAvgTime float64 `json:"io_write_avg_time"`
	IoReadAvgTime  float64 `json:"io_read_avg_time"`
	IoSeekAvgTime  float64 `json:"io_seek_avg_time"`
	IoSyncAvgTime  float64 `json:"io_sync_avg_time"`

	GcNumDetails             details `json:"gc_num_details"`
	MnesiaRAMTxCountDetails  details `json:"mnesia_ram_tx_count_details"`
	MnesiaDiskTxCountDetails details `json:"mnesia_disk_tx_count_details"`
	GcBytesReclaimedDetails  details `json:"gc_bytes_reclaimed_details"`
	IoReadAvgTimeDetails     details `json:"io_read_avg_time_details"`
	IoReadBytesDetails       details `json:"io_read_bytes_details"`
	IoWriteAvgTimeDetails    details `json:"io_write_avg_time_details"`
	IoWriteBytesDetails      details `json:"io_write_bytes_details"`
}

func getNode(n *Input) {
	var (
		Nodes        []node
		collectStart = time.Now()
		pts          []*point.Point
		opts         = append(point.DefaultMetricOptions(), point.WithTime(n.start))
	)

	if err := n.requestJSON("/api/nodes", &Nodes); err != nil {
		l.Error(err.Error())
		n.lastErr = err
		return
	}

	for _, node := range Nodes {
		kvs := point.NewTags(n.mergedTags)

		kvs = kvs.AddTag("url", n.URL).
			AddTag("node_name", node.Name).
			AddV2("disk_free_alarm", node.DiskFreeAlarm, true).
			AddV2("disk_free", node.DiskFree, true).
			AddV2("fd_used", node.FdUsed, true).
			AddV2("mem_alarm", node.MemAlarm, true).
			AddV2("mem_limit", node.MemLimit, true).
			AddV2("mem_used", node.MemUsed, true).
			AddV2("run_queue", node.RunQueue, true).
			AddV2("running", node.Running, true).
			AddV2("sockets_used", node.SocketsUsed, true).
			AddV2("io_write_avg_time", node.IoWriteAvgTime, true).
			AddV2("io_read_avg_time", node.IoReadAvgTime, true).
			AddV2("io_sync_avg_time", node.IoSyncAvgTime, true).
			AddV2("io_seek_avg_time", node.IoSeekAvgTime, true)

		pts = append(pts, point.NewPointV2(nodeMeasurementName, kvs, opts...))
	}

	if err := n.feeder.Feed(point.Metric, pts,
		dkio.WithCollectCost(time.Since(collectStart)),
		dkio.WithElection(n.Election),
		dkio.WithSource(inputName),
	); err != nil {
		l.Errorf("FeedMeasurement: %s", err.Error())
	}
}
