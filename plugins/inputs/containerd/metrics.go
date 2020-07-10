// +build linux

package containerd

import (
	"bytes"
	"context"
	"errors"
	"time"

	"github.com/containerd/containerd"
	"github.com/containerd/containerd/cio"
	v1 "github.com/containerd/containerd/metrics/types/v1"
	"github.com/containerd/containerd/namespaces"
	"github.com/containerd/typeurl"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io"
)

var attach = func(*cio.FIFOSet) (cio.IO, error) { return cio.NullIO("") }

func (con *Containerd) collectContainerd() ([]byte, error) {

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	ctx = namespaces.WithNamespace(ctx, con.Namespace)

	client, err := containerd.New(con.HostPath)
	if err != nil {
		return nil, err
	}
	defer client.Close()

	containers, err := client.Containers(ctx)
	if err != nil {
		return nil, err
	}

	var pts bytes.Buffer
	var tags = make(map[string]string)
	tags["namespace"] = con.Namespace

	for _, container := range containers {

		if _, ok := con.ids[container.ID()]; !con.isAll && !ok {
			continue
		}

		metrics, err := getMetrics(ctx, container)
		if err != nil {
			return nil, err
		}

		tags["id"] = container.ID()
		for k, v := range con.Tags {
			tags[k] = v
		}
		fields := parseMetrics(metrics)
		pt, err := io.MakeMetric(defaultMeasurement, tags, fields, time.Now())
		if err != nil {
			return nil, err
		}

		if err := appendData(&pts, pt); err != nil {
			return nil, err
		}
	}

	return pts.Bytes(), nil
}

func appendData(pts *bytes.Buffer, pt []byte) error {
	if _, err := pts.Write(pt); err != nil {
		return err
	}
	if _, err := pts.WriteString("\n"); err != nil {
		return err
	}
	return nil
}

func getMetrics(ctx context.Context, c containerd.Container) (*v1.Metrics, error) {
	task, err := c.Task(ctx, attach)
	if err != nil {
		return nil, err
	}

	mt, err := task.Metrics(ctx)
	if err != nil {
		return nil, err
	}

	data, err := typeurl.UnmarshalAny(mt.Data)
	if err != nil {
		return nil, err
	}

	meta, ok := data.(*v1.Metrics)
	if !ok {
		return nil, errors.New("invalid metrics data")
	}
	return meta, nil
}

func parseMetrics(mt *v1.Metrics) map[string]interface{} {
	var fields = make(map[string]interface{})

	if mt.Pids != nil {
		fields["pids_limit"] = int64(mt.Pids.Limit)
		fields["pids_current"] = int64(mt.Pids.Current)
	}

	if mt.CPU != nil {
		if mt.CPU.Usage != nil {
			fields["cpu_usage_total"] = int64(mt.CPU.Usage.Total)
			fields["cpu_usage_kernal"] = int64(mt.CPU.Usage.Kernel)
			fields["cpu_usage_user"] = int64(mt.CPU.Usage.User)
		}

		if mt.CPU.Throttling != nil {
			fields["cpu_thorttling_periods"] = int64(mt.CPU.Throttling.Periods)
			fields["cpu_thorttling_throttled_eriods"] = int64(mt.CPU.Throttling.ThrottledPeriods)
			fields["cpu_thorttling_throttled_ime"] = int64(mt.CPU.Throttling.ThrottledTime)
		}
	}

	if mt.Memory != nil {
		fields["memory_cache"] = int64(mt.Memory.Cache)
		fields["memory_rss"] = int64(mt.Memory.RSS)
		fields["memory_rss_huge"] = int64(mt.Memory.RSSHuge)
		fields["memory_mapped_file"] = int64(mt.Memory.MappedFile)
		fields["memory_dirty"] = int64(mt.Memory.Dirty)
		fields["memory_writeback"] = int64(mt.Memory.Writeback)
		fields["memory_pg_pg_in"] = int64(mt.Memory.PgPgIn)
		fields["memory_pg_pg_out"] = int64(mt.Memory.PgPgOut)
		fields["memory_pg_fault"] = int64(mt.Memory.PgFault)
		fields["memory_pg_maj_fault"] = int64(mt.Memory.PgMajFault)
		fields["memory_inactive_anon"] = int64(mt.Memory.InactiveAnon)
		fields["memory_active_anon"] = int64(mt.Memory.ActiveAnon)
		fields["memory_inactive_file"] = int64(mt.Memory.InactiveFile)
		fields["memory_active_file"] = int64(mt.Memory.ActiveFile)
		fields["memory_unevictable"] = int64(mt.Memory.Unevictable)
		fields["memory_hierarchical_memory_limit"] = int64(mt.Memory.HierarchicalMemoryLimit)
		fields["memory_hierarchical_swap_limit"] = int64(mt.Memory.HierarchicalSwapLimit)
		fields["memory_total_cache"] = int64(mt.Memory.TotalCache)
		fields["memory_total_rss"] = int64(mt.Memory.TotalRSS)
		fields["memory_total_rss_huge"] = int64(mt.Memory.TotalRSSHuge)
		fields["memory_total_mapped_file"] = int64(mt.Memory.TotalMappedFile)
		fields["memory_total_dirty"] = int64(mt.Memory.TotalDirty)
		fields["memory_total_writeback"] = int64(mt.Memory.TotalWriteback)
		fields["memory_total_pg_pg_in"] = int64(mt.Memory.TotalPgPgIn)
		fields["memory_total_pg_pg_out"] = int64(mt.Memory.TotalPgPgOut)
		fields["memory_total_pg_fault"] = int64(mt.Memory.TotalPgFault)
		fields["memory_total_pg_maj_fault"] = int64(mt.Memory.TotalPgMajFault)
		fields["memory_total_inactive_anon"] = int64(mt.Memory.TotalInactiveAnon)
		fields["memory_total_active_anon"] = int64(mt.Memory.TotalActiveAnon)
		fields["memory_total_inactive_file"] = int64(mt.Memory.TotalInactiveFile)
		fields["memory_total_active_file"] = int64(mt.Memory.TotalActiveFile)
		fields["memory_total_unevictable"] = int64(mt.Memory.TotalUnevictable)

		if mt.Memory.Usage != nil {
			fields["memory_usage_limit"] = int64(mt.Memory.Usage.Limit)
			fields["memory_usage_usage"] = int64(mt.Memory.Usage.Usage)
			fields["memory_usage_max"] = int64(mt.Memory.Usage.Max)
			fields["memory_usage_failcnt"] = int64(mt.Memory.Usage.Failcnt)
		}

		if mt.Memory.Swap != nil {
			fields["memory_swap_limit"] = int64(mt.Memory.Swap.Limit)
			fields["memory_swap_usage"] = int64(mt.Memory.Swap.Usage)
			fields["memory_swap_max"] = int64(mt.Memory.Swap.Max)
			fields["memory_swap_failcnt"] = int64(mt.Memory.Swap.Failcnt)
		}

		if mt.Memory.Kernel != nil {
			fields["memory_kernel_limit"] = int64(mt.Memory.Kernel.Limit)
			fields["memory_kernel_usage"] = int64(mt.Memory.Kernel.Usage)
			fields["memory_kernel_max"] = int64(mt.Memory.Kernel.Max)
			fields["memory_kernel_failcnt"] = int64(mt.Memory.Kernel.Failcnt)
		}
		if mt.Memory.KernelTCP != nil {

			fields["memory_kernel_tcp_limit"] = int64(mt.Memory.KernelTCP.Limit)
			fields["memory_kernel_tcp_usage"] = int64(mt.Memory.KernelTCP.Usage)
			fields["memory_kernel_tcp_max"] = int64(mt.Memory.KernelTCP.Max)
			fields["memory_kernel_tcp_failcnt"] = int64(mt.Memory.KernelTCP.Failcnt)
		}
	}

	return fields
}
