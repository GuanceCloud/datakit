// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package kubernetes

import (
	"context"
	"sync"
	"time"

	"github.com/GuanceCloud/cliutils/point"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/container/typed"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/goroutine"
)

func (k *Kube) gather(category string, feed func([]*point.Point) error) {
	var mu sync.Mutex
	counterWithName := make(map[string]map[string]int)

	wrapFeed := func(pts []*point.Point) error {
		err := feed(pts)
		if err == nil {
			collectPtsVec.WithLabelValues(category).Add(float64(len(pts)))
		}
		return err
	}

	start := time.Now()

	namespaces, err := k.getActiveNamespaces(context.Background())
	if err != nil {
		klog.Warnf("get namespaces err: %s", err)
		return
	}

	g := goroutine.NewGroup(goroutine.Option{Name: "k8s-" + category})

	for typee, constructor := range resources {
		func(typ resourceType, newResource resourceConstructor) {
			g.Go(func(_ context.Context) error {
				namespaces := namespaces
				if !typ.namespaced {
					namespaces = []string{""}
				}

				counts := k.gatherResource(newResource, namespaces, category, wrapFeed)

				mu.Lock()
				counterWithName[typ.name] = counts
				mu.Unlock()

				collectResourcePtsVec.WithLabelValues(category, typ.name).Observe(float64(sum(counts)))
				return nil
			})
		}(typee, constructor)
	}

	if err = g.Wait(); err != nil {
		klog.Error(err.Error())

		return
	}

	if category == "metric" {
		pts := transToNamespacePoint(counterWithName)
		k.addExtraTagsV2(pts)
		if err := wrapFeed(pts); err != nil {
			klog.Warn(err)
		}
	}

	collectCostVec.WithLabelValues(category).Observe(time.Since(start).Seconds())
}

func (k *Kube) gatherResource(
	newResource resourceConstructor,
	namespaces []string,
	category string,
	feed func([]*point.Point) error,
) map[string]int {
	counts := make(map[string]int)

	var opts []point.Option
	var transform func(m metadata) pointKVs

	switch category {
	case "metric":
		opts = point.DefaultMetricOptions()
		transform = func(m metadata) pointKVs {
			return m.transformMetric()
		}

	case "object":
		opts = point.DefaultObjectOptions()
		transform = func(m metadata) pointKVs {
			return m.transformObject()
		}

	default:
		return nil
	}

	processor := func(m metadata) (int, error) {
		pts := transform(m)
		k.addExtraTags(pts)

		points := transToPoint(pts, opts)
		n := len(points)

		if n != 0 {
			return n, feed(points)
		}
		return 0, nil
	}

	r := newResource(k.client)

	for _, ns := range namespaces {
		count := gatherAndProcessResource(r, ns, processor)
		counts[ns] = count
	}

	return counts
}

func (k *Kube) addExtraTags(pts pointKVs) {
	for _, pt := range pts {
		pt.SetTags(k.cfg.ExtraTags)
	}
}

func (k *Kube) addExtraTagsV2(pts []*point.Point) {
	for _, pt := range pts {
		for k, v := range k.cfg.ExtraTags {
			pt.MustAddTag([]byte(k), []byte(v))
		}
	}
}

func gatherAndProcessResource(r resource, ns string, processor func(metadata) (int, error)) (count int) {
	for {
		data, err := r.getMetadata(context.Background(), ns)
		if err != nil {
			fetchErrorVec.WithLabelValues(ns, err.Error()).Set(float64(time.Now().Unix()))
			klog.Warnf("fetch k8s resource err: %s", err)
			break
		}

		num, err := processor(data)
		if err != nil {
			klog.Warnf("process err: %s", err)
			continue
		}
		count += num

		if !r.hasNext() {
			break
		}
	}

	return count
}

func transToPoint(pts pointKVs, opts []point.Option) []*point.Point {
	if len(pts) == 0 {
		return nil
	}

	var res []*point.Point
	for _, pt := range pts {
		r := point.NewPointV2(
			[]byte(pt.Name()),
			append(point.NewTags(pt.Tags()), point.NewKVs(pt.Fields())...),
			opts...,
		)
		res = append(res, r)
	}

	return res
}

func transToNamespacePoint(counterWithName map[string]map[string]int) []*point.Point {
	// counterWithName rotated
	//    e.g. map["kube-system"]["pod"] = 10
	counterWithNamespace := make(map[string]map[string]int)
	var pts pointKVs

	for name, m := range counterWithName {
		for namespace, count := range m {
			if counterWithNamespace[namespace] == nil {
				counterWithNamespace[namespace] = make(map[string]int)
			}
			counterWithNamespace[namespace][name] = count
		}
	}

	for namespace, m := range counterWithNamespace {
		p := typed.NewPointKV("kubernetes")
		p.SetTag("namespace", namespace)
		for name, count := range m {
			p.SetField(name, count)
		}
		pts = append(pts, p)
	}

	return transToPoint(pts, point.DefaultMetricOptions())
}

func sum(m map[string]int) int {
	sum := 0
	for _, v := range m {
		sum += v
	}
	return sum
}
