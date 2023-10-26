// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package kubernetes

import (
	"context"

	"github.com/GuanceCloud/cliutils/point"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/container/typed"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/datakit"
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

//nolint:gochecknoinits
func init() {
	registerMeasurements(&event{})
}

func (k *Kube) gatherEvent(feed func([]*point.Point) error) {
	list, err := k.client.GetEvents("").List(context.Background(), metav1.ListOptions{Limit: 1})
	if err != nil {
		klog.Warnf("failed to load events, err: %s", err)
		return
	}
	// Do not care old events.
	resourceVersion := list.ResourceVersion

	watchFunc := func(opt metav1.ListOptions) (kubewatch.Interface, error) {
		return k.client.GetEvents("").Watch(context.Background(), opt)
	}

	w, err := watch.NewRetryWatcher(resourceVersion, &cache.ListWatch{WatchFunc: watchFunc})
	if err != nil {
		klog.Warnf("failed to retry watch events, err: %s", err)
		return
	}
	defer w.Stop()

	for {
		select {
		case <-datakit.Exit.Wait():
			return

		case <-k.done:
			klog.Infof("event terminated")
			return

		case event, ok := <-w.ResultChan():
			if k.paused() {
				klog.Info("not leader for election, exit")
				return
			}

			if !ok {
				klog.Warnf("event channel is closed, exit")
				return
			}

			pts := k.transformEvent(&event)
			if err := feed(pts); err != nil {
				klog.Warn(err)
			} else {
				collectPtsVec.WithLabelValues("events").Add(float64(len(pts)))
			}
		}
	}
}

func (k *Kube) transformEvent(event *kubewatch.Event) []*point.Point {
	//nolint:exhaustive
	switch event.Type {
	case kubewatch.Bookmark:
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

	pt := typed.NewPointKV(eventLoggingMeasurement)
	pt.SetTag("uid", string(item.UID))
	pt.SetTag("type", item.Type)
	pt.SetTag("reason", item.Reason)
	pt.SetTag("from_node", k.cfg.NodeName)
	pt.SetTags(k.cfg.ExtraTags)

	pt.SetField("involved_kind", item.InvolvedObject.Kind)
	pt.SetField("involved_uid", string(item.InvolvedObject.UID))
	pt.SetField("involved_name", item.InvolvedObject.Name)
	pt.SetField("involved_namespace", item.InvolvedObject.Namespace)
	pt.SetField("message", item.Message)

	status := "unknown"
	switch item.Type {
	case "Normal":
		status = "info"
	case "Warning":
		status = "warn"
	default:
		// nil
	}
	pt.SetField("status", status)

	pt.SetCustomerTags(item.Labels, getGlobalCustomerKeys())

	pts := pointKVs{pt}

	return transToPoint(pts, append(point.DefaultLoggingOptions(), point.WithTime(item.CreationTimestamp.Time)))
}

type event struct{ *typed.PointKV }

//nolint:lll
func (*event) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name: eventLoggingMeasurement,
		Desc: "The logging of the Kubernetes Event.",
		Type: "logging",
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
			"message":            &inputs.FieldInfo{DataType: inputs.String, Unit: inputs.UnknownUnit, Desc: "Details of event log"},
		},
	}
}
