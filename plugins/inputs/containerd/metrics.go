// +build linux

package containerd

import (
	"context"
	"errors"
	"fmt"
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

func (s *stream) processMetrics() error {

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	ctx = namespaces.WithNamespace(ctx, s.sub.Namespace)

	client, err := containerd.New(s.sub.HostPath)
	if err != nil {
		return err
	}
	defer client.Close()

	cs, err := client.Containers(ctx)
	if err != nil {
		return err
	}

	for _, c := range cs {
		fmt.Println(c.ID())
		if _, ok := s.ids[c.ID()]; !s.isAll && !ok {
			continue
		}

		task, err := c.Task(ctx, attach)
		if err != nil {
			return err
		}

		mt, err := task.Metrics(ctx)
		if err != nil {
			return err
		}

		data, err := typeurl.UnmarshalAny(mt.Data)
		if err != nil {
			return err
		}

		meta, ok := data.(*v1.Metrics)
		if !ok {
			return errors.New("invalid metrics data")
		}

		point, err := parseMetrics(s.sub.Measurement, s.sub.Namespace, c.ID(), meta)
		if err != nil {
			return err
		}

		s.points = append(s.points, point)

	}
	return nil
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
