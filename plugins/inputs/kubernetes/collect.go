package kubernetes

import (
	"context"
	v1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
)

func collectDaemonSets(ctx context.Context, i *Input) {
	list, err := i.client.getDaemonSets(ctx)
	if err != nil {
		i.mu.Lock()
		i.err = err
		i.mu.Unlock()
		return
	}

	for _, d := range list.Items {
		i.gatherDaemonSet(d)
	}
}

func (i *Input) gatherDaemonSet(d v1.DaemonSet) {
	m := &daemonsetMeasurement{
		name:   "kube_daemonsets",
		tags:   make(map[string]string),
		fields: make(map[string]interface{}),
	}

	for key, value := range i.Tags {
		m.tags[key] = value
	}

	m.tags["name"] = d.Name
	m.tags["namespace"] = d.Namespace

	for key, val := range d.Spec.Selector.MatchLabels {
		if i.selectorFilter.Match(key) {
			m.tags["selector_"+key] = val
		}
	}

	m.fields["generation"] = d.Generation
	m.fields["current_number_scheduled"] = d.Status.CurrentNumberScheduled
	m.fields["desired_number_scheduled"] = d.Status.DesiredNumberScheduled
	m.fields["number_available"] = d.Status.NumberAvailable
	m.fields["number_misscheduled"] = d.Status.NumberMisscheduled
	m.fields["number_ready"] = d.Status.NumberReady
	m.fields["number_unavailable"] = d.Status.NumberUnavailable
	m.fields["updated_number_scheduled"] = d.Status.UpdatedNumberScheduled

	if d.GetCreationTimestamp().Second() != 0 {
		m.fields["created"] = d.GetCreationTimestamp().UnixNano()
	}

	i.mu.Lock()
	i.collectCache = append(i.collectCache, m)
	i.mu.Unlock()
}

func collectDeployments(ctx context.Context, i *Input) {
	list, err := i.client.getDeployments(ctx)
	if err != nil {
		i.mu.Lock()
		i.err = err
		i.mu.Unlock()
		return
	}
	for _, d := range list.Items {
		i.gatherDeployment(d)
	}
}

func (i *Input) gatherDeployment(d v1.Deployment) {
	m := &deploymentMeasurement{
		name:   "kube_deployment",
		tags:   make(map[string]string),
		fields: make(map[string]interface{}),
	}

	for key, value := range i.Tags {
		m.tags[key] = value
	}

	m.tags["name"] = d.Name
	m.tags["namespace"] = d.Namespace

	for key, val := range d.Spec.Selector.MatchLabels {
		if i.selectorFilter.Match(key) {
			m.tags["selector_"+key] = val
		}
	}

	m.fields["replicas_available"] = d.Status.AvailableReplicas
	m.fields["replicas_unavailable"] = d.Status.UnavailableReplicas
	m.fields["created"] = d.GetCreationTimestamp().UnixNano()

	i.mu.Lock()
	i.collectCache = append(i.collectCache, m)
	i.mu.Unlock()
}

func collectNodes(ctx context.Context, i *Input) {
	list, err := i.client.getNodes(ctx)
	if err != nil {
		i.mu.Lock()
		i.err = err
		i.mu.Unlock()
		return
	}
	for _, n := range list.Items {
		i.gatherNode(n)
	}
}

func (i *Input) gatherNode(n corev1.Node) {
	m := &deploymentMeasurement{
		name:   "kube_deployment",
		tags:   make(map[string]string),
		fields: make(map[string]interface{}),
	}

	for key, value := range i.Tags {
		m.tags[key] = value
	}

	m.tags["name"] = n.Name

	m.fields["replicas_available"] = n.Status.AvailableReplicas
	m.fields["replicas_unavailable"] = n.Status.UnavailableReplicas
	m.fields["created"] = n.GetCreationTimestamp().UnixNano()

	for resourceName, val := range n.Status.Capacity {
		switch resourceName {
		case "cpu":
			m.fields["capacity_cpu_cores"] = convertQuantity(string(val.Format), 1)
			m.fields["capacity_millicpu_cores"] = convertQuantity(string(val.Format), 1000)
		case "memory":
			m.fields["capacity_memory_bytes"] = convertQuantity(string(val.Format), 1)
		case "pods":
			m.fields["capacity_pods"] = atoi(string(val.Format))
		}
	}

	for resourceName, val := range n.Status.Allocatable {
		switch resourceName {
		case "cpu":
			m.fields["allocatable_cpu_cores"] = convertQuantity(string(val.Format), 1)
			m.fields["allocatable_millicpu_cores"] = convertQuantity(string(val.Format), 1000)
		case "memory":
			m.fields["allocatable_memory_bytes"] = convertQuantity(string(val.Format), 1)
		case "pods":
			m.fields["allocatable_pods"] = atoi(string(val.Format))
		}
	}

	i.mu.Lock()
	i.collectCache = append(i.collectCache, m)
	i.mu.Unlock()
}
