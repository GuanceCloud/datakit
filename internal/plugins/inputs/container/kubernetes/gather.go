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
	"k8s.io/apimachinery/pkg/api/errors"
)

func (k *Kube) gather(category string, feed func([]*point.Point) error, paused, nodeLocal bool) {
	var mu sync.Mutex

	counterWithName := make(map[string]map[string]int)
	processor := k.composeProcessor(category, feed)

	start := time.Now()

	namespaces, err := k.getActiveNamespaces(context.Background())
	if err != nil {
		klog.Warnf("get namespaces err: %s", err)
		return
	}

	g := goroutine.NewGroup(goroutine.Option{Name: "k8s-" + category})

	for typee, constructor := range resources {
		fieldSelector := ""

		if paused && nodeLocal && !typee.nodeLocal {
			continue
		}
		if nodeLocal && typee.nodeLocal {
			fieldSelector = getFieldSelector(k.nodeName)
		}

		func(typ resourceType, newResource resourceConstructor) {
			g.Go(func(_ context.Context) error {
				namespaces := namespaces
				if !typ.namespaced {
					namespaces = []string{""}
				}

				r := newResource(k.client)
				counts := k.gatherResource(r, typ.name, fieldSelector, namespaces, processor)

				mu.Lock()
				counterWithName[typ.name] = counts
				mu.Unlock()

				collectResourcePtsVec.WithLabelValues(category, typ.name, fieldSelector).Observe(float64(sum(counts)))
				return nil
			})
		}(typee, constructor)
	}

	if err = g.Wait(); err != nil {
		klog.Error("exception error: %s", err)
		return
	}

	if category == "metric" {
		pts := transToNamespacePoint(counterWithName)
		if len(pts) != 0 {
			k.addExtraTagsV2(pts)
			if err := feed(pts); err != nil {
				klog.Warn(err)
			} else {
				collectPtsVec.WithLabelValues(category).Add(float64(len(pts)))
			}
		}
	}

	collectCostVec.WithLabelValues(category).Observe(time.Since(start).Seconds())
}

func (k *Kube) gatherResource(
	r resource,
	resourceName, fieldSelector string,
	namespaces []string,
	processor func(m metadata) (int, error),
) map[string]int {
	counts := make(map[string]int)

	for _, ns := range namespaces {
		count := gatherAndProcessResource(r, resourceName, ns, fieldSelector, processor)
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
			pt.MustAddTag(k, v)
		}
	}
}

func (k *Kube) composeProcessor(category string, feed func([]*point.Point) error) func(m metadata) (int, error) {
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
		// unreachable
	}

	fn := func(m metadata) (int, error) {
		pts := transform(m)
		k.addExtraTags(pts)

		points := transToPoint(pts, opts)
		n := len(points)
		if n == 0 {
			return 0, nil
		}

		if err := feed(points); err != nil {
			return 0, err
		}

		collectPtsVec.WithLabelValues(category).Add(float64(n))
		return n, nil
	}

	return fn
}

func gatherAndProcessResource(
	r resource,
	resourceName, ns, fieldSelector string,
	processor func(metadata) (int, error),
) (count int) {
	for {
		data, err := r.getMetadata(context.Background(), ns, fieldSelector)
		if err != nil {
			if !errors.IsNotFound(err) {
				fetchErrorVec.WithLabelValues(ns, resourceName, err.Error()).Set(float64(time.Now().Unix()))
			}
			break
		}

		num, err := processor(data)
		if err != nil {
			klog.Warnf("resources %s process err: %s", resourceName, err)
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
			pt.Name(),
			append(point.NewTags(pt.Tags()), point.NewKVs(pt.Fields())...),
			opts...,
		)

		res = append(res, r)
	}

	// debugging
	if pts[0].Name() == "kube_pod" {
		klog.Infof("pod-data, time: %s, pointKVs: %s", time.Now().Format(time.RFC3339), pts)
	}

	return res
}

func transToNamespacePoint(counterWithName map[string]map[string]int) []*point.Point {
	if len(counterWithName) == 0 {
		return nil
	}

	// counterWithName rotated
	//    e.g. map["kube-system"]["pod"] = 10
	counterWithNamespace := make(map[string]map[string]int)
	var pts pointKVs

	for name, m := range counterWithName {
		for namespace, count := range m {
			if count == 0 {
				continue
			}
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

func getFieldSelector(nodeName string) string {
	return "spec.nodeName=" + nodeName
}
