package kubernetes

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
	kubev1core "k8s.io/client-go/kubernetes/typed/core/v1"
)

const kubernetesEventName = "kubernetes_events"

type event struct {
	client    kubev1core.EventInterface
	extraTags map[string]string
}

func (e *event) Gather() { /* nil */ }

func (e *event) Watch() {
	// Outer loop, for reconnections.
	for {
		select {
		case <-datakit.Exit.Wait():
			l.Infof("Event watching stopped")
			return
		default:
			// nil
		}

		events, err := e.client.List(context.Background(), metav1.ListOptions{})
		if err != nil {
			l.Errorf("Failed to load events: %v", err)
			time.Sleep(time.Second)
			continue
		}
		// Do not write old events.

		resourceVersion := events.ResourceVersion

		watcher, err := e.client.Watch(context.Background(),
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
						if err := e.feedEvent(event); err != nil {
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

			case <-datakit.Exit.Wait():
				watcher.Stop()
				l.Infof("Event watching stopped")
				return
			}
		}
	}
}

func (e *event) feedEvent(item *kubeapi.Event) error {
	tags := map[string]string{
		"kind":      item.InvolvedObject.Kind,
		"name":      item.Name,
		"namespace": item.Namespace,
		"node_name": item.Source.Host,
		"type":      item.Type,
		"reason":    item.Reason,
	}

	for k, v := range e.extraTags {
		tags[k] = v
	}

	tags["message"] = item.Message

	msg, err := json.Marshal(tags)
	if err != nil {
		return err
	}
	delete(tags, "message")

	fields := map[string]interface{}{
		"message": string(msg),
	}
	pt, err := io.MakePoint(kubernetesEventName, tags, fields, time.Now())
	if err != nil {
		return err
	}

	return io.Feed(inputName, datakit.Logging, []*io.Point{pt}, nil)
}

func (*event) Stop() {}

func (*event) Resource() { /*empty interface*/ }

func (*event) LineProto() (*io.Point, error) { return nil, nil }

//nolint:lll
func (*event) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name: kubernetesEventName,
		Desc: "Kubernetes event 日志数据",
		Type: "logging",
		Tags: map[string]interface{}{
			"kind":      inputs.NewTagInfo("Kind of the referent."),
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
