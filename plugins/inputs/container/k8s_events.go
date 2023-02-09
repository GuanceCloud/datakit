// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package container

import (
	"context"
	"encoding/json"
	"sync/atomic"
	"time"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io/point"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
	kubeapi "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	kubewatch "k8s.io/apimachinery/pkg/watch"
)

const k8sEventName = "kubernetes_events"

var globalPause = new(atomBool)

func watchingEvent(client k8sClientX, extraTags tagsType, done <-chan interface{}, election bool) {
	// Outer loop, for reconnections.
	for {
		select {
		case <-datakit.Exit.Wait():
			l.Infof("k8s event watching exit")
			return
		case <-done:
			l.Infof("k8s event watching stopped")
			return
		default:
			// nil
		}

		events, err := client.getEvents().List(context.Background(), metaV1ListOption)
		if err != nil {
			l.Warnf("failed to load events: %s", err)
			time.Sleep(time.Second)
			continue
		}
		// Do not write old events.

		resourceVersion := events.ResourceVersion

		watcher, err := client.getEvents().Watch(context.Background(),
			metav1.ListOptions{Watch: true, ResourceVersion: resourceVersion})
		if err != nil {
			l.Warnf("failed to start watch for new events: %s", err)
			time.Sleep(time.Second)
			continue
		}

		watchChannel := watcher.ResultChan()
		// Inner loop, for update processing.
	inner_loop:
		for {
			select {
			case watchUpdate, ok := <-watchChannel:
				if !ok {
					l.Warnf("event watch channel closed, retry")
					break inner_loop
				}

				if globalPause.get() {
					continue
				}

				if watchUpdate.Type == kubewatch.Error {
					if status, ok := watchUpdate.Object.(*metav1.Status); ok {
						l.Warnf("error during watch: %#v", status)
						break inner_loop
					}
					l.Warnf("received unexpected error: %#v", watchUpdate.Object)
					break inner_loop
				}

				if event, ok := watchUpdate.Object.(*kubeapi.Event); ok {
					switch watchUpdate.Type {
					case kubewatch.Added, kubewatch.Modified:
						if err := feedEvent(event, extraTags, election); err != nil {
							l.Warnf("failed to parse event: %s", err)
						}

					case kubewatch.Bookmark, kubewatch.Error:
						l.Infof("ignore type %s", watchUpdate.Type)

					case kubewatch.Deleted:
						// Deleted events are silently ignored.
					default:
						l.Warnf("unknown watchUpdate.Type: %#v", watchUpdate.Type)
					}
				} else {
					l.Warnf("wrong object received: %v", watchUpdate)
				}

			case <-done:
				watcher.Stop()
				l.Infof("event watching stopped")
				return

			case <-datakit.Exit.Wait():
				watcher.Stop()
				l.Infof("k8s event watching exit")
				return
			}
		}
	}
}

func feedEvent(event *kubeapi.Event, extraTags tagsType, election bool) error {
	return inputs.FeedMeasurement("k8s-events",
		datakit.Logging,
		[]inputs.Measurement{buildEventData(event, extraTags, election)},
		nil,
	)
}

func buildEventData(item *kubeapi.Event, extraTags tagsType, election bool) inputs.Measurement {
	obj := newEvent()
	obj.tags["kind"] = item.InvolvedObject.Kind
	obj.tags["name"] = item.Name
	obj.tags["namespace"] = defaultNamespace(item.Namespace)
	obj.tags["node_name"] = item.Source.Host
	obj.tags["type"] = item.Type
	obj.tags["reason"] = item.Reason
	obj.tags["status"] = "info"
	obj.tags.append(extraTags)
	obj.election = election

	obj.tags["message"] = item.Message
	msg, err := json.Marshal(obj.tags)
	if err != nil {
		l.Warnf("failed to build event message: %s", err)
	} else {
		obj.fields["message"] = string(msg)
	}
	delete(obj.tags, "message")

	return obj
}

type event struct {
	tags     tagsType
	fields   fieldsType
	election bool
}

func newEvent() *event {
	return &event{
		tags:   make(tagsType),
		fields: make(fieldsType),
	}
}

func (e *event) LineProto() (*point.Point, error) {
	return point.NewPoint(k8sEventName, e.tags, e.fields, point.LOptElectionV2(e.election))
}

//nolint:lll
func (*event) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name: k8sEventName,
		Desc: "The logging of the Kubernetes Event.",
		Type: "logging",
		Tags: map[string]interface{}{
			"kind":      inputs.NewTagInfo("Kind of the referent."),
			"status":    inputs.NewTagInfo("log status"),
			"name":      inputs.NewTagInfo("Name must be unique within a namespace."),
			"namespace": inputs.NewTagInfo("Namespace defines the space within which each name must be unique."),
			"node_name": inputs.NewTagInfo("NodeName is a request to schedule this pod onto a specific node."),
			"type":      inputs.NewTagInfo("Type of this event (Normal, Warning), new types could be added in the future."),
			"reason":    inputs.NewTagInfo("This should be a short, machine understandable string that gives the reason, for the transition into the object's current status."),
		},
		Fields: map[string]interface{}{
			"message": &inputs.FieldInfo{DataType: inputs.String, Unit: inputs.UnknownUnit, Desc: "event log details"},
		},
	}
}

type atomBool struct{ flag int32 }

//nolint:unconvert
func (b *atomBool) set(value bool) {
	var i int32 = 0
	if value {
		i = 1
	}
	atomic.StoreInt32(&(b.flag), int32(i))
}

func (b *atomBool) get() bool {
	return atomic.LoadInt32(&(b.flag)) != 0
}

//nolint:gochecknoinits
func init() {
	registerMeasurement(&event{})
}
