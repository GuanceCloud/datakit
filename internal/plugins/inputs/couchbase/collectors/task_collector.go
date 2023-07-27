// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package collectors

import "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs/couchbase/objects"

const (
	taskRebalance                      = "rebalance"
	taskBucketCompaction               = "bucket_compaction"
	taskXdcr                           = "xdcr"
	taskClusterLogCollection           = "clusterLogsCollection"
	metricRebalancePerNode             = "rebalancePerNode"
	metricCompacting                   = "compacting"
	metricXdcrChangesLeft              = "xdcrChangesLeft"
	metricXdcrDocsChecked              = "xdcrDocsChecked"
	metricXdcrDocsWritten              = "xdcrDocsWritten"
	metricXdcrPaused                   = "xdcrPaused"
	metricXdcrErrors                   = "xdcrErrors"
	metricDocsTotal                    = "progressDocsTotal"
	metricDocsTransferred              = "progressDocsTransferred"
	metricDocsActiveVbucketsLeft       = "progressActiveVBucketsLeft"
	metricDocsTotalReplicaVBucketsLeft = "progressReplicaVBucketsLeft"
)

func (c *Client) taskCollect() error {
	defer func() { c.config = nil }()
	if c.config == nil {
		c.config = objects.GetTaskCollectorDefaultConfig()
	}

	var tasks []objects.Task
	err := c.get(c.url("pools/default/tasks"), &tasks)
	if err != nil {
		return err
	}

	var buckets []objects.BucketInfo
	err = c.get(c.url("pools/default/buckets"), &buckets)
	if err != nil {
		return err
	}

	compactsReported := c.collectTasks(tasks)
	// always report the compacting task, even if it is not happening
	// this is to not break dashboards and make it easier to test alert rule
	// and etc.
	compact := c.config.Metrics[metricCompacting]

	for _, bucket := range buckets {
		c.Ctx.BucketName = bucket.Name
		if _, ok := compactsReported[bucket.Name]; !ok {
			c.addPoint(c.config.Namespace, getFieldName(compact), 0, compact.Labels)
		}

		compactsReported[bucket.Name] = true
	}

	return nil
}

func (c *Client) collectTasks(tasks []objects.Task) map[string]bool {
	compactsReported := map[string]bool{}

	for _, task := range tasks {
		c.Ctx.BucketName = task.Bucket
		c.Ctx.Source = task.Source
		c.Ctx.Target = task.Target

		switch task.Type {
		case taskRebalance:
			c.addRebalance(task)
		case taskBucketCompaction:
			// XXX: there can be more than one compacting tasks for the same
			// bucket for now, let's report just the first.
			c.addBucketCompaction(task, compactsReported[task.Bucket])
			compactsReported[task.Bucket] = true
		case taskXdcr:
			c.addXdcr(task)
		case taskClusterLogCollection:
			c.addClusterLogCollection(task)
		default:
		}
	}

	return compactsReported
}

func (c *Client) addRebalance(task objects.Task) {
	if rb, ok := c.config.Metrics[taskRebalance]; ok && rb.Enabled {
		c.addPoint(c.config.Namespace, getFieldName(rb), task.Progress, rb.Labels)
	}

	if rbPN, ok := c.config.Metrics[metricRebalancePerNode]; ok && rbPN.Enabled {
		for node, progress := range task.PerNode {
			c.Ctx.NodeHostname = node
			c.addPoint(c.config.Namespace, getFieldName(rbPN), progress.Progress, rbPN.Labels)
		}
	}
}

func (c *Client) addBucketCompaction(task objects.Task, compactsReported bool) {
	if cp, ok := c.config.Metrics[metricCompacting]; ok && cp.Enabled && !compactsReported {
		c.addPoint(c.config.Namespace, getFieldName(cp), task.Progress, cp.Labels)
	}
}

func (c *Client) addXdcr(task objects.Task) {
	if xcl, ok := c.config.Metrics[metricXdcrChangesLeft]; ok && xcl.Enabled {
		c.addPoint(c.config.Namespace, getFieldName(xcl), float64(task.ChangesLeft), xcl.Labels)
	}

	if xdc, ok := c.config.Metrics[metricXdcrDocsChecked]; ok && xdc.Enabled {
		c.addPoint(c.config.Namespace, getFieldName(xdc), float64(task.DocsChecked), xdc.Labels)
	}

	if xdw, ok := c.config.Metrics[metricXdcrDocsWritten]; ok && xdw.Enabled {
		c.addPoint(c.config.Namespace, getFieldName(xdw), float64(task.DocsWritten), xdw.Labels)
	}

	if xp, ok := c.config.Metrics[metricXdcrPaused]; ok && xp.Enabled {
		c.addPoint(c.config.Namespace, getFieldName(xp), boolToFloat64(task.PauseRequested), xp.Labels)
	}

	if xe, ok := c.config.Metrics[metricXdcrErrors]; ok && xe.Enabled {
		c.addPoint(c.config.Namespace, getFieldName(xe), float64(len(task.Errors)), xe.Labels)
	}

	for _, data := range task.DetailedProgress.PerNode {
		// for each node grab these specific metrics from the config (if they exist)
		// then grab their data from the request and dump it into prometheus.
		c.Ctx.BucketName = task.DetailedProgress.Bucket

		if dt, ok := c.config.Metrics[metricDocsTotal]; ok && dt.Enabled {
			c.addPoint(c.config.Namespace, getFieldName(dt), float64(data.Ingoing.DocsTotal), dt.Labels)
		}

		if dtrans, ok := c.config.Metrics[metricDocsTransferred]; ok && dtrans.Enabled {
			c.addPoint(c.config.Namespace, getFieldName(dtrans), float64(data.Ingoing.DocsTransferred), dtrans.Labels)
		}

		if avbl, ok := c.config.Metrics[metricDocsActiveVbucketsLeft]; ok && avbl.Enabled {
			c.addPoint(c.config.Namespace, getFieldName(avbl), float64(data.Ingoing.ActiveVBucketsLeft), avbl.Labels)
		}

		if rvbl, ok := c.config.Metrics[metricDocsTotalReplicaVBucketsLeft]; ok && rvbl.Enabled {
			c.addPoint(c.config.Namespace, getFieldName(rvbl), float64(data.Ingoing.ReplicaVBucketsLeft), rvbl.Labels)
		}
	}
}

func (c *Client) addClusterLogCollection(task objects.Task) {
	if clc, ok := c.config.Metrics[taskClusterLogCollection]; ok && clc.Enabled {
		c.addPoint(c.config.Namespace, getFieldName(clc), task.Progress, clc.Labels)
	}
}
