// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package kubernetes

import (
	"context"
	"strconv"

	"github.com/GuanceCloud/cliutils/point"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/config"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs"
	kubeapi "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	kubewatch "k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/watch"
)

const (
	eventLoggingMeasurement = "kubernetes_events"
)

func (k *Kube) gatherEvent(ctx context.Context) {
	list, err := k.client.GetEvents(allNamespaces).List(context.Background(), metav1.ListOptions{Limit: 1})
	if err != nil {
		klog.Warnf("query events failed, err: %s", err)
		return
	}

	// Do not care old events.
	resourceVersion := list.ResourceVersion
	resourceVersion = latestResourveVersion(k.lastEventResourceVersion, resourceVersion)
	klog.Infof("use event resourceVersion %s", resourceVersion)

	watchFunc := func(opt metav1.ListOptions) (kubewatch.Interface, error) {
		return k.client.GetEvents("").Watch(context.Background(), opt)
	}

	w, err := watch.NewRetryWatcher(resourceVersion, &cache.ListWatch{WatchFunc: watchFunc})
	if err != nil {
		klog.Warnf("watch events failed, err: %s", err)
		return
	}
	defer w.Stop()

	for {
		select {
		case <-ctx.Done():
			klog.Infof("k8s event exit")
			return

		case event, ok := <-w.ResultChan():
			if !ok {
				klog.Warnf("event channel is closed, exit")
				return
			}
			pts := k.buildEventPoints(&event)
			feedLogging("k8s-event", k.cfg.Feeder, pts)
		}
	}
}

func (k *Kube) buildEventPoints(event *kubewatch.Event) []*point.Point {
	//nolint:exhaustive
	switch event.Type {
	case kubewatch.Bookmark, kubewatch.Deleted:
		// Bookmark events are silently ignored.
		return nil
	default:
		// next
	}

	item, ok := event.Object.(*kubeapi.Event)
	if !ok {
		klog.Warnf("event type mismatch")
		return nil
	}

	var kvs point.KVs
	kvs = kvs.AddTag("uid", string(item.UID))
	kvs = kvs.AddTag("type", item.Type)
	kvs = kvs.AddTag("reason", item.Reason)
	kvs = kvs.AddTag("from_node", config.RenameNode(k.cfg.NodeName))

	kvs = kvs.AddV2("involved_kind", item.InvolvedObject.Kind, false)
	kvs = kvs.AddV2("involved_uid", string(item.InvolvedObject.UID), false)
	kvs = kvs.AddV2("involved_name", item.InvolvedObject.Name, false)
	kvs = kvs.AddV2("involved_namespace", item.InvolvedObject.Namespace, false)
	kvs = kvs.AddV2("source_component", item.Source.Component, false)
	kvs = kvs.AddV2("source_host", item.Source.Host, false)

	kvs = kvs.AddV2("resource_version", item.ResourceVersion, false)
	kvs = kvs.AddV2("message", item.Message, false)

	status := "unknown"
	switch item.Type {
	case "Normal":
		status = "info"
	case "Warning":
		status = "warn"
	default:
		// nil
	}

	kvs = kvs.AddV2("status", status, false)
	kvs = append(kvs, point.NewTags(k.cfg.ExtraTags)...)
	pt := point.NewPointV2(eventLoggingMeasurement, kvs,
		append(point.DefaultLoggingOptions(), point.WithTime(item.LastTimestamp.Time))...)

	// record resourceVersion
	k.lastEventResourceVersion = item.ResourceVersion
	return []*point.Point{pt}
}

// nolint
func latestResourveVersion(v1, v2 string) string {
	int1, _ := strconv.Atoi(v1)
	int2, _ := strconv.Atoi(v2)
	if int1 < int2 {
		return v2
	}
	return v1
}

type K8sEventLog struct{}

//nolint:lll
func (*K8sEventLog) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name: eventLoggingMeasurement,
		Desc: "The logging of the Kubernetes Event.",
		Cat:  point.Logging,
		Tags: map[string]interface{}{
			"uid":    inputs.NewTagInfo("The UID of event."),
			"type":   inputs.NewTagInfo("Type of this event."),
			"reason": inputs.NewTagInfo("This should be a short, machine understandable string that gives the reason, for the transition into the object's current status."),
		},
		Fields: map[string]interface{}{
			"involved_kind":      &inputs.FieldInfo{DataType: inputs.String, Unit: inputs.UnknownUnit, Desc: "Kind of the referent for involved object."},
			"involved_uid":       &inputs.FieldInfo{DataType: inputs.String, Unit: inputs.UnknownUnit, Desc: "The UID of involved object."},
			"involved_name":      &inputs.FieldInfo{DataType: inputs.String, Unit: inputs.UnknownUnit, Desc: "Name must be unique within a namespace for involved object."},
			"involved_namespace": &inputs.FieldInfo{DataType: inputs.String, Unit: inputs.UnknownUnit, Desc: "Namespace defines the space within which each name must be unique for involved object."},
			"source_component":   &inputs.FieldInfo{DataType: inputs.String, Unit: inputs.UnknownUnit, Desc: "Component from which the event is generated."},
			"source_host":        &inputs.FieldInfo{DataType: inputs.String, Unit: inputs.UnknownUnit, Desc: "Node name on which the event is generated."},
			"message":            &inputs.FieldInfo{DataType: inputs.String, Unit: inputs.UnknownUnit, Desc: "Details of event log"},
		},
	}
}
