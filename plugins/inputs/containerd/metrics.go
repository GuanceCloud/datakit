// +build linux

package containerd

import (
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
	influxdb "github.com/influxdata/influxdb1-client/v2"
)

var attach = func(*cio.FIFOSet) (cio.IO, error) { return cio.NullIO("") }

func (i *Impl) collectContainerd() ([]*influxdb.Point, error) {

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	ctx = namespaces.WithNamespace(ctx, i.Namespace)

	client, err := containerd.New(i.HostPath)
	if err != nil {
		return nil, err
	}
	defer client.Close()

	cs, err := client.Containers(ctx)
	if err != nil {
		return nil, err
	}

	var pts []*influxdb.Point

	for _, c := range cs {
		if _, ok := i.ids[c.ID()]; !i.isAll && !ok {
			continue
		}

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

		pt, err := parseMetrics(inputName, i.Namespace, c.ID(), meta)
		if err != nil {
			return nil, err
		}

		pts = append(pts, pt)

	}
	return pts, nil
}

func parseMetrics(mensurement, namespace, id string, mt *v1.Metrics) (*influxdb.Point, error) {
	var fields = make(map[string]interface{})

	rematch(mt, "", fields)

	return influxdb.NewPoint(
		mensurement,
		map[string]string{"namespace": namespace, "id": id},
		fields,
		time.Now(),
	)
}

func rematch(data interface{}, succkey string, m map[string]interface{}) {

	if reflect.ValueOf(data).IsNil() {
		return
	}

	t := reflect.TypeOf(data).Elem()
	v := reflect.ValueOf(data).Elem()

	for i := 0; i < v.NumField(); i++ {

		key := t.Field(i).Name

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
			rematch(v.Field(i).Interface(), succkey+key+"_", m)

		case reflect.Uint64:
			// integer
			m[succkey+key] = int64(v.Field(i).Uint())

		default:
			// nil

		}
	}
}
