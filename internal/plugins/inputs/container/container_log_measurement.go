// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package container

import (
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/io/point"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs"
)

type containerLog struct{}

func (c *containerLog) LineProto() (*point.Point, error) { return nil, nil }

//nolint:lll
func (c *containerLog) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name: "Use Logging Source",
		Desc: "The logging of the container.",
		Type: "logging",
		Tags: map[string]interface{}{
			"container_id":           inputs.NewTagInfo(`Container ID`),
			"container_name":         inputs.NewTagInfo("Container name from k8s (label `io.kubernetes.container.name`). If empty then use $container_runtime_name."),
			"container_runtime_name": inputs.NewTagInfo(`Container name from runtime (like 'docker ps'). If empty then use 'unknown' ([:octicons-tag-24: Version-1.4.6](changelog.md#cl-1.4.6)).`),
			"container_type":         inputs.NewTagInfo(`The type of the container (this container is created by Kubernetes/Docker/containerd).`),
			// "stream":                 inputs.NewTagInfo(`数据流方式，stdout/stderr/tty（containerd 日志缺少此字段）`),
			"pod_name":    inputs.NewTagInfo("The pod name of the container (label `io.kubernetes.pod.name`)."),
			"namespace":   inputs.NewTagInfo("The pod namespace of the container (label `io.kubernetes.pod.namespace`)."),
			"deployment":  inputs.NewTagInfo(`The deployment name of the container's pod (unsupported containerd).`),
			"service":     inputs.NewTagInfo("The name of the service, if `service` is empty then use `source`."),
			"[POD_LABEL]": inputs.NewTagInfo("The pod labels will be extracted as tags if `extract_k8s_label_as_tags` is enabled."),
		},
		Fields: map[string]interface{}{
			"status":          &inputs.FieldInfo{DataType: inputs.String, Unit: inputs.UnknownUnit, Desc: "The status of the logging, only supported `info/emerg/alert/critical/error/warning/debug/OK/unknown`."},
			"log_read_lines":  &inputs.FieldInfo{DataType: inputs.Int, Unit: inputs.NCount, Desc: "The lines of the read file ([:octicons-tag-24: Version-1.4.6](changelog.md#cl-1.4.6))."},
			"log_read_offset": &inputs.FieldInfo{DataType: inputs.Int, Unit: inputs.UnknownUnit, Desc: "The offset of the read file ([:octicons-tag-24: Version-1.4.8](changelog.md#cl-1.4.8) · [:octicons-beaker-24: Experimental](index.md#experimental))."},
			"log_read_time":   &inputs.FieldInfo{DataType: inputs.DurationSecond, Unit: inputs.UnknownUnit, Desc: "The timestamp of the read file."},
			"message_length":  &inputs.FieldInfo{DataType: inputs.SizeByte, Unit: inputs.NCount, Desc: "The length of the message content."},
			"message":         &inputs.FieldInfo{DataType: inputs.String, Unit: inputs.UnknownUnit, Desc: "The text of the logging."},
		},
	}
}

//nolint:gochecknoinits
func init() {
	registerMeasurement(&containerLog{})
}
