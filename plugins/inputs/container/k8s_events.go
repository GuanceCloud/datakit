package container

import (
	"context"
	"encoding/json"
	"time"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
	kubeapi "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	kubewatch "k8s.io/apimachinery/pkg/watch"
)

const k8sEventName = "kubernetes_events"

var globalPause bool

func watchingEvent(client k8sClientX, extraTags tagsType, stop <-chan interface{}) {
	// Outer loop, for reconnections.
	for {
		select {
		case <-stop:
			l.Infof("Event watching stopped")
			return
		default:
			// nil
		}

		events, err := client.getEvents().List(context.Background(), metaV1ListOption)
		if err != nil {
			l.Errorf("Failed to load events: %v", err)
			time.Sleep(time.Second)
			continue
		}
		// Do not write old events.

		resourceVersion := events.ResourceVersion

		watcher, err := client.getEvents().Watch(context.Background(),
			metav1.ListOptions{Watch: true, ResourceVersion: resourceVersion})
		if err != nil {
			l.Errorf("Failed to start watch for new events: %v", err)
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
					l.Errorf("Event watch channel closed")
					break inner_loop
				}

				if globalPause {
					continue
				}

				if watchUpdate.Type == kubewatch.Error {
					if status, ok := watchUpdate.Object.(*metav1.Status); ok {
						l.Errorf("Error during watch: %#v", status)
						break inner_loop
					}
					l.Errorf("Received unexpected error: %#v", watchUpdate.Object)
					break inner_loop
				}

				if event, ok := watchUpdate.Object.(*kubeapi.Event); ok {
					switch watchUpdate.Type {
					case kubewatch.Added, kubewatch.Modified:
						if err := feedEvent(event, extraTags); err != nil {
							l.Error("Failed to parse event: %v", err)
						}

					case kubewatch.Bookmark, kubewatch.Error:
						l.Info("ignore type %s", watchUpdate.Type)

					case kubewatch.Deleted:
						// Deleted events are silently ignored.
					default:
						l.Warnf("Unknown watchUpdate.Type: %#v", watchUpdate.Type)
					}
				} else {
					l.Errorf("Wrong object received: %v", watchUpdate)
				}

			case <-stop:
				watcher.Stop()
				l.Infof("Event watching stopped")
				return
			}
		}
	}
}

func feedEvent(event *kubeapi.Event, extraTags tagsType) error {
	return inputs.FeedMeasurement("k8s-events",
		datakit.Logging,
		[]inputs.Measurement{buildEventData(event, extraTags)},
		nil,
	)
}

func buildEventData(item *kubeapi.Event, extraTags tagsType) inputs.Measurement {
	obj := newEvent()
	obj.tags["kind"] = item.InvolvedObject.Kind
	obj.tags["name"] = item.Name
	obj.tags["namespace"] = defaultNamespace(item.Namespace)
	obj.tags["node_name"] = item.Source.Host
	obj.tags["type"] = item.Type
	obj.tags["reason"] = item.Reason
	obj.tags["status"] = "info"
	obj.tags.append(extraTags)

	obj.tags["message"] = item.Message
	msg, err := json.Marshal(obj.tags)
	if err != nil {
		l.Errorf("Failed to build event message: %s", err)
	} else {
		obj.fields["message"] = string(msg)
	}
	delete(obj.tags, "message")

	return obj
}

type event struct {
	tags   tagsType
	fields fieldsType
	time   time.Time
}

func newEvent() *event {
	return &event{
		tags:   make(tagsType),
		fields: make(fieldsType),
	}
}

func (e *event) LineProto() (*io.Point, error) {
	return io.NewPoint(k8sEventName, e.tags, e.fields, &io.PointOption{Time: e.time, Category: datakit.Logging})
}

//nolint:lll
func (*event) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name: k8sEventName,
		Desc: "Kubernetes event 日志数据",
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

//nolint:gochecknoinits
func init() {
	registerMeasurement(&event{})
}
