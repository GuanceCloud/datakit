package kubernetes

import (
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
	"time"
)

// daemonset
type daemonsetMeasurement struct {
	name   string
	tags   map[string]string
	fields map[string]interface{}
	ts     time.Time
}

func (m *daemonsetMeasurement) LineProto() (*io.Point, error) {
	return io.MakePoint(m.name, m.tags, m.fields, m.ts)
}

func (m *daemonsetMeasurement) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name: "kube_daemonsets",
		Desc: "daemonsets resource measurement",
		Tags: map[string]interface{}{
			"name":      &inputs.TagInfo{Desc: "name"},
			"namespace": &inputs.TagInfo{Desc: "namespace"},
		},
		Fields: map[string]interface{}{
			"generation": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Count,
				Unit:     inputs.NCount,
				Desc:     "",
			},
			"current_number_scheduled": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Count,
				Unit:     inputs.NCount,
				Desc:     "",
			},
			"desired_number_scheduled": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Count,
				Unit:     inputs.NCount,
				Desc:     "",
			},
			"number_available": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Count,
				Unit:     inputs.NCount,
				Desc:     "",
			},
			"number_misscheduled": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Count,
				Unit:     inputs.NCount,
				Desc:     "",
			},
			"number_ready": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Count,
				Unit:     inputs.NCount,
				Desc:     "",
			},
			"number_unavailable": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Count,
				Unit:     inputs.NCount,
				Desc:     "",
			},
			"updated_number_scheduled": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Count,
				Unit:     inputs.NCount,
				Desc:     "",
			},
			"created": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Count,
				Unit:     inputs.UnknownUnit,
				Desc:     "创建时间(时间戳)",
			},
		},
	}
}

// deployment
type deploymentMeasurement struct {
	name   string
	tags   map[string]string
	fields map[string]interface{}
	ts     time.Time
}

func (m *deploymentMeasurement) LineProto() (*io.Point, error) {
	return io.MakePoint(m.name, m.tags, m.fields, m.ts)
}

func (m *deploymentMeasurement) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name: "kube_deployment",
		Desc: "deployment resource measurement",
		Tags: map[string]interface{}{
			"name":      &inputs.TagInfo{Desc: "name"},
			"namespace": &inputs.TagInfo{Desc: "namespace"},
		},
		Fields: map[string]interface{}{
			"replicas_available": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Count,
				Unit:     inputs.NCount,
				Desc:     "",
			},
			"replicas_unavailable": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Count,
				Unit:     inputs.NCount,
				Desc:     "",
			},
			"created": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Count,
				Unit:     inputs.UnknownUnit,
				Desc:     "创建时间(时间戳)",
			},
		},
	}
}

// endpoint
type endpointMeasurement struct {
	name   string
	tags   map[string]string
	fields map[string]interface{}
	ts     time.Time
}

func (m *endpointMeasurement) LineProto() (*io.Point, error) {
	return io.MakePoint(m.name, m.tags, m.fields, m.ts)
}

func (m *endpointMeasurement) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name: "kube_endpoints",
		Desc: "endpoints resource measurement",
		Tags: map[string]interface{}{
			"name":          &inputs.TagInfo{Desc: "name"},
			"namespace":     &inputs.TagInfo{Desc: "namespace"},
			"hostname":      &inputs.TagInfo{Desc: "hostname"},
			"node_name":     &inputs.TagInfo{Desc: "node_name"},
			"port_name":     &inputs.TagInfo{Desc: "port_name"},
			"port_protocol": &inputs.TagInfo{Desc: "port_protocol"},
		},
		Fields: map[string]interface{}{
			"generation": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Count,
				Unit:     inputs.NCount,
				Desc:     "",
			},
			"ready": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Count,
				Unit:     inputs.NCount,
				Desc:     "",
			},
			"port": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Count,
				Unit:     inputs.NCount,
				Desc:     "",
			},
			"created": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Count,
				Unit:     inputs.UnknownUnit,
				Desc:     "创建时间(时间戳)",
			},
		},
	}
}

// ingress
type ingressMeasurement struct {
	name   string
	tags   map[string]string
	fields map[string]interface{}
	ts     time.Time
}

func (m *ingressMeasurement) LineProto() (*io.Point, error) {
	return io.MakePoint(m.name, m.tags, m.fields, m.ts)
}

func (m *ingressMeasurement) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name: "kube_ingress",
		Desc: "ingress resource measurement",
		Tags: map[string]interface{}{
			"name":          &inputs.TagInfo{Desc: "name"},
			"namespace":     &inputs.TagInfo{Desc: "namespace"},
			"hostname":      &inputs.TagInfo{Desc: "hostname"},
			"node_name":     &inputs.TagInfo{Desc: "node_name"},
			"port_name":     &inputs.TagInfo{Desc: "port_name"},
			"port_protocol": &inputs.TagInfo{Desc: "port_protocol"},
		},
		Fields: map[string]interface{}{
			"generation": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Count,
				Unit:     inputs.NCount,
				Desc:     "",
			},
			"ready": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Count,
				Unit:     inputs.NCount,
				Desc:     "",
			},
			"port": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Count,
				Unit:     inputs.NCount,
				Desc:     "",
			},
			"created": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Count,
				Unit:     inputs.UnknownUnit,
				Desc:     "创建时间(时间戳)",
			},
		},
	}
}

// node
type nodeMeasurement struct {
	name   string
	tags   map[string]string
	fields map[string]interface{}
	ts     time.Time
}

func (m *nodeMeasurement) LineProto() (*io.Point, error) {
	return io.MakePoint(m.name, m.tags, m.fields, m.ts)
}

func (m *nodeMeasurement) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name: "kube_ingress",
		Desc: "ingress resource measurement",
		Tags: map[string]interface{}{
			"name": &inputs.TagInfo{Desc: "name"},
		},
		Fields: map[string]interface{}{
			"capacity_cpu_cores": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Count,
				Unit:     inputs.NCount,
				Desc:     "",
			},
			"capacity_millicpu_cores": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Count,
				Unit:     inputs.NCount,
				Desc:     "",
			},
			"capacity_memory_bytes": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Count,
				Unit:     inputs.NCount,
				Desc:     "",
			},
			"capacity_pods": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Count,
				Unit:     inputs.NCount,
				Desc:     "",
			},
			"allocatable_cpu_cores": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Count,
				Unit:     inputs.NCount,
				Desc:     "",
			},
			"allocatable_millicpu_cores": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Count,
				Unit:     inputs.NCount,
				Desc:     "",
			},
			"allocatable_memory_bytes": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Count,
				Unit:     inputs.NCount,
				Desc:     "",
			},
			"allocatable_pods": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Count,
				Unit:     inputs.NCount,
				Desc:     "",
			},
		},
	}
}

// pod
type podMeasurement struct {
	name   string
	tags   map[string]string
	fields map[string]interface{}
	ts     time.Time
}

func (m *podMeasurement) LineProto() (*io.Point, error) {
	return io.MakePoint(m.name, m.tags, m.fields, m.ts)
}

func (m *podMeasurement) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name: "kube_pod",
		Desc: "pod resource measurement",
		Tags: map[string]interface{}{
			"namespace": &inputs.TagInfo{Desc: "namespace"},
			"node_name": &inputs.TagInfo{Desc: "node name"},
			"pod_name":  &inputs.TagInfo{Desc: "pod name"},
		},
		Fields: map[string]interface{}{
			"last_transition_time": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Count,
				Unit:     inputs.NCount,
				Desc:     "",
			},
			"ready": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Count,
				Unit:     inputs.NCount,
				Desc:     "",
			},
		},
	}
}

// pod_container
type podContainerMeasurement struct {
	name   string
	tags   map[string]string
	fields map[string]interface{}
	ts     time.Time
}

func (m *podContainerMeasurement) LineProto() (*io.Point, error) {
	return io.MakePoint(m.name, m.tags, m.fields, m.ts)
}

func (m *podContainerMeasurement) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name: "kube_pod_container",
		Desc: "pod container resource measurement",
		Tags: map[string]interface{}{
			"namespace":      &inputs.TagInfo{Desc: "namespace"},
			"node_name":      &inputs.TagInfo{Desc: "node name"},
			"pod_name":       &inputs.TagInfo{Desc: "pod name"},
			"container_name": &inputs.TagInfo{Desc: "container name"},
			"phase":          &inputs.TagInfo{Desc: "container name"},
			"state":          &inputs.TagInfo{Desc: "container name"},
			"readiness":      &inputs.TagInfo{Desc: "container name"},
		},
		Fields: map[string]interface{}{
			"restarts_total": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Count,
				Unit:     inputs.NCount,
				Desc:     "",
			},
			"state_code": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Count,
				Unit:     inputs.NCount,
				Desc:     "",
			},
			"state_reason": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Count,
				Unit:     inputs.NCount,
				Desc:     "",
			},
			"phase_reason": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Count,
				Unit:     inputs.NCount,
				Desc:     "",
			},
			"terminated_reason": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Count,
				Unit:     inputs.NCount,
				Desc:     "",
			},
			"resource_requests_millicpu_units": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Count,
				Unit:     inputs.NCount,
				Desc:     "",
			},
			"resource_requests_memory_bytes": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Count,
				Unit:     inputs.NCount,
				Desc:     "",
			},
			"resource_limits_millicpu_units": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Count,
				Unit:     inputs.NCount,
				Desc:     "",
			},
			"resource_limits_memory_bytes": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Count,
				Unit:     inputs.NCount,
				Desc:     "",
			},
		},
	}
}

// service
type serviceMeasurement struct {
	name   string
	tags   map[string]string
	fields map[string]interface{}
	ts     time.Time
}

func (m *serviceMeasurement) LineProto() (*io.Point, error) {
	return io.MakePoint(m.name, m.tags, m.fields, m.ts)
}

func (m *serviceMeasurement) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name: "kube_service",
		Desc: "service resource measurement",
		Tags: map[string]interface{}{
			"namespace":     &inputs.TagInfo{Desc: "namespace"},
			"service_name":  &inputs.TagInfo{Desc: "node name"},
			"port_name":     &inputs.TagInfo{Desc: "pod name"},
			"port_protocol": &inputs.TagInfo{Desc: "container name"},
			"external_name": &inputs.TagInfo{Desc: "container name"},
			"cluster_ip":    &inputs.TagInfo{Desc: "container name"},
		},
		Fields: map[string]interface{}{
			"created": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Count,
				Unit:     inputs.NCount,
				Desc:     "",
			},
			"generation": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Count,
				Unit:     inputs.NCount,
				Desc:     "",
			},
			"port": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Count,
				Unit:     inputs.NCount,
				Desc:     "",
			},
			"target_port": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Count,
				Unit:     inputs.NCount,
				Desc:     "",
			},
		},
	}
}
