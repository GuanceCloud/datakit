// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package kubernetes

import (
	"context"
	"fmt"
	"time"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/container/typed"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/io/point"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs"
	kubeapi "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	kubewatch "k8s.io/apimachinery/pkg/watch"
)

//nolint:gochecknoinits
func init() {
	registerMeasurement(&event{})
}

func (k *Kube) watchingEvent() {
	var watcher kubewatch.Interface
	var watchChannel <-chan kubewatch.Event

	defer func() {
		if watcher != nil {
			watcher.Stop()
		}
		klog.Infof("event exit")
	}()

	for {
		select {
		case <-datakit.Exit.Wait():
			return
		case <-k.done:
			klog.Infof("event terminated")
			return
		default:
			// nil
		}

	inner_loop:
		var err error

		watcher, err = k.reconnectingEvent()
		if err != nil {
			time.Sleep(time.Second)
			klog.Warnf("reconnect k8s event err: %s", err)
			continue
		} else {
			klog.Info("reconnect k8s event ok")
			watchChannel = watcher.ResultChan()
		}

		for {
			select {
			case <-datakit.Exit.Wait():
				return

			case <-k.done:
				klog.Infof("event terminated")
				return

			case item, ok := <-watchChannel:
				if k.paused() {
					klog.Debug("not leader for election, skip")
					time.Sleep(time.Second)
					continue
				}

				if !ok {
					klog.Warnf("event watch channel closed, retry")
					goto inner_loop
				}

				if item.Type == kubewatch.Error {
					if status, ok := item.Object.(*metav1.Status); ok {
						klog.Warnf("error during watch: %#v", status)
					} else {
						klog.Warnf("received unexpected error: %#v", item.Object)
					}
					time.Sleep(time.Second)
					goto inner_loop
				}

				if meas, err := k.gatherEvent(&item); err != nil {
					klog.Warnf("failed to get event log, err: %s", err)
				} else {
					if len(meas) == 0 {
						klog.Debug("nopoint of k8s-event")
					} else {
						err = inputs.FeedMeasurement("k8s-events", datakit.Logging, meas, &io.Option{})
						if err != nil {
							klog.Warnf("failed to feed event, err: %s", err)
						}
					}
				}
			}
		}
	}
}

func (k *Kube) reconnectingEvent() (kubewatch.Interface, error) {
	list, err := k.client.GetEvents().List(context.Background(), metaV1ListOption)
	if err != nil {
		return nil, fmt.Errorf("failed to load events, err: %w", err)
	}

	// Do not write old events.
	resourceVersion := list.ResourceVersion

	watcher, err := k.client.GetEvents().Watch(context.Background(), metav1.ListOptions{Watch: true, ResourceVersion: resourceVersion})
	if err != nil {
		return nil, fmt.Errorf("failed to start watch for new events, err: %w", err)
	}

	return watcher, nil
}

func (k *Kube) gatherEvent(item *kubewatch.Event) ([]inputs.Measurement, error) {
	e, ok := item.Object.(*kubeapi.Event)
	if !ok {
		return nil, fmt.Errorf("wrong object received: %s", item)
	}

	//nolint:exhaustive
	switch item.Type {
	case kubewatch.Bookmark:
		// Bookmark events are silently ignored.
		return nil, nil
	default:
		// next
	}

	p := typed.NewPointKV()
	p.SetTag("uid", string(e.UID))
	p.SetTag("type", e.Type)
	p.SetTag("reason", e.Reason)
	p.SetTags(k.cfg.ExtraTags)

	p.SetField("involved_kind", e.InvolvedObject.Kind)
	p.SetField("involved_uid", string(e.InvolvedObject.UID))
	p.SetField("involved_name", e.InvolvedObject.Name)
	p.SetField("involved_namespace", e.InvolvedObject.Namespace)
	p.SetField("message", e.Message)

	ev := event{p}
	return []inputs.Measurement{&ev}, nil
}

type event struct{ typed.PointKV }

func (e *event) LineProto() (*point.Point, error) {
	return point.NewPoint("kubernetes_events", e.Tags(), e.Fields(), loggingOpt)
}

//nolint:lll
func (*event) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name: "kubernetes_events",
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
