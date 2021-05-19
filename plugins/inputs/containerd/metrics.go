// +build linux

package containerd

import (
	"bytes"
	"context"
	"errors"
	"reflect"
	"strings"
	"time"

	"github.com/containerd/containerd"
	"github.com/containerd/containerd/cio"
	v1 "github.com/containerd/containerd/metrics/types/v1"
	"github.com/containerd/containerd/namespaces"
	"github.com/containerd/typeurl"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io"
)

var attach = func(*cio.FIFOSet) (cio.IO, error) { return cio.NullIO("") }

func (con *Containerd) collectContainerd() ([]byte, error) {

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	ctx = namespaces.WithNamespace(ctx, con.Namespace)

	client, err := containerd.New(con.Location)
	if err != nil {
		return nil, err
	}
	defer client.Close()

	containers, err := client.Containers(ctx)
	if err != nil {
		return nil, err
	}

	var pts bytes.Buffer
	var tags = make(map[string]string)
	tags["namespace"] = con.Namespace

	for _, container := range containers {

		if _, ok := con.ids[container.ID()]; !con.isAll && !ok {
			continue
		}

		metrics, err := getMetrics(ctx, container)
		if err != nil {
			return nil, err
		}

		tags["id"] = container.ID()
		for k, v := range con.Tags {
			tags[k] = v
		}
		fields := parseMetrics(metrics)
		pt, err := io.MakeMetric(defaultMeasurement, tags, fields, time.Now())
		if err != nil {
			return nil, err
		}

		if err := appendData(&pts, pt); err != nil {
			return nil, err
		}
	}

	return pts.Bytes(), nil
}

func appendData(pts *bytes.Buffer, pt []byte) error {
	if _, err := pts.Write(pt); err != nil {
		return err
	}
	if _, err := pts.WriteString("\n"); err != nil {
		return err
	}
	return nil
}

func getMetrics(ctx context.Context, c containerd.Container) (*v1.Metrics, error) {
	task, err := c.Task(ctx, attach)
	if err != nil {
		return nil, err
	}

	mt, err := task.Metrics(ctx)
	if err != nil {
		return nil, err
	}

	data, err := typeurl.UnmarshalAny(mt.Data)
	if err != nil {
		return nil, err
	}

	meta, ok := data.(*v1.Metrics)
	if !ok {
		return nil, errors.New("invalid metrics data")
	}
	return meta, nil
}

func parseMetrics(mt *v1.Metrics) map[string]interface{} {
	var fields = make(map[string]interface{})
	deepHit(mt, "", fields)
	return fields
}

func deepHit(data interface{}, prefix string, m map[string]interface{}) {

	if reflect.ValueOf(data).IsNil() {
		return
	}

	t := reflect.TypeOf(data).Elem()
	v := reflect.ValueOf(data).Elem()

	for i := 0; i < v.NumField(); i++ {
		if v.Field(i).CanInterface() {
			key := strings.ToLower(t.Field(i).Name)

			// filter 'XXX_', example:
			//   type PidsStat struct {
			//       Current              uint64
			//       Limit                uint64
			//       XXX_NoUnkeyedLiteral struct{}
			//       XXX_unrecognized     []byte
			//       XXX_sizecache        int32
			//   }
			if strings.HasPrefix(key, "XXX_") {
				continue
			}

			switch v.Field(i).Kind() {

			case reflect.Ptr:
				deepHit(v.Field(i).Interface(), prefix+key+"_", m)

			case reflect.Uint64:
				// integer
				m[prefix+key] = int64(v.Field(i).Uint())

			default:
				// nil
			}
		}
	}
}
