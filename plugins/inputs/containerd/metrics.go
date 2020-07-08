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

	client, err := containerd.New(con.HostPath)
	if err != nil {
		return nil, err
	}
	defer client.Close()

	cs, err := client.Containers(ctx)
	if err != nil {
		return nil, err
	}

	var tags = make(map[string]string)
	tags["namespace"] = con.Namespace
	var pts bytes.Buffer

	for _, c := range cs {
		// IS UGLY !!

		if _, ok := con.ids[c.ID()]; !con.isAll && !ok {
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

		tags["id"] = c.ID()
		for k, v := range con.Tags {
			tags[k] = v
		}

		pt, err := parseMetrics(defaultMeasurement, tags, meta)
		if err != nil {
			return nil, err
		}

		if _, err := pts.Write(pt); err != nil {
			return nil, err
		}

		if _, err := pts.WriteString("\n"); err != nil {
			return nil, err
		}
	}

	return pts.Bytes(), nil
}

func parseMetrics(mensurement string, tags map[string]string, mt *v1.Metrics) ([]byte, error) {
	var fields = make(map[string]interface{})

	rematch(mt, "", fields)

	return io.MakeMetric(defaultMeasurement, tags, fields, time.Now())
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
