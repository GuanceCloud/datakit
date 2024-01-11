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
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/goroutine"
	"k8s.io/apimachinery/pkg/api/errors"
)

func (k *Kube) gather(category string, feed func([]*point.Point) error, paused bool) {
	var mu sync.Mutex

	countPts := []pointV2{}
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

		if paused && k.cfg.NodeLocal && !typee.nodeLocal {
			continue
		}
		if k.cfg.NodeLocal && typee.nodeLocal {
			fieldSelector = getFieldSelector(k.nodeName)
		}

		func(typ resourceType, newResource resourceConstructor) {
			g.Go(func(_ context.Context) error {
				startCollect := time.Now()

				namespaces := namespaces
				if !typ.namespaced {
					namespaces = []string{""}
				}

				r := newResource(k.client)
				k.gatherResource(r, typ.name, fieldSelector, namespaces, processor)

				mu.Lock()
				countPts = append(countPts, r.count()...)
				mu.Unlock()

				collectResourceCostVec.WithLabelValues(category, typ.name, fieldSelector).Observe(time.Since(startCollect).Seconds())
				return nil
			})
		}(typee, constructor)
	}

	if err = g.Wait(); err != nil {
		klog.Errorf("unexpected error: %s", err)
		return
	}

	if category == "metric" && len(countPts) != 0 {
		k.addExtraTagsV2(countPts)
		if err := feed(countPts); err != nil {
			klog.Warn(err)
		} else {
			collectPtsVec.WithLabelValues(category).Add(float64(len(countPts)))
		}
	}

	collectCostVec.WithLabelValues(category).Observe(time.Since(start).Seconds())
}

func (k *Kube) gatherResource(r resource, resourceName, fieldSelector string, namespaces []string, processor func(m metadata) error) {
	for _, ns := range namespaces {
		gatherAndProcessResource(r, resourceName, ns, fieldSelector, processor)
	}
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

func (k *Kube) composeProcessor(category string, feed func([]*point.Point) error) func(m metadata) error {
	var opts []point.Option
	var builder func(m metadata) pointKVs

	switch category {
	case "metric":
		opts = point.DefaultMetricOptions()
		builder = func(m metadata) pointKVs {
			return m.newMetric(k.cfg)
		}

	case "object":
		opts = point.DefaultObjectOptions()
		builder = func(m metadata) pointKVs {
			return m.newObject(k.cfg)
		}
	default:
		// unreachable
	}

	fn := func(m metadata) error {
		pts := builder(m)
		k.addExtraTags(pts)

		points := transToPoint(pts, opts)
		if len(points) == 0 {
			return nil
		}

		if err := feed(points); err != nil {
			return err
		}

		collectPtsVec.WithLabelValues(category).Add(float64(len(points)))
		return nil
	}

	return fn
}

func gatherAndProcessResource(r resource, resourceName, ns, fieldSelector string, processor func(metadata) error) {
	for {
		data, err := r.getMetadata(context.Background(), ns, fieldSelector)
		if err != nil {
			if !errors.IsNotFound(err) {
				fetchErrorVec.WithLabelValues(ns, resourceName, err.Error()).Set(float64(time.Now().Unix()))
			}
			break
		}

		if err := processor(data); err != nil {
			klog.Warnf("resources %s process err: %s", resourceName, err)
			continue
		}

		if !r.hasNext() {
			break
		}
	}
}

func transToPoint(pts pointKVs, opts []point.Option) []*point.Point {
	if len(pts) == 0 {
		return nil
	}

	var res []*point.Point
	t := time.Now()

	for _, pt := range pts {
		r := point.NewPointV2(
			pt.Name(),
			append(point.NewTags(pt.Tags()), point.NewKVs(pt.Fields())...),
			append(opts, point.WithTime(t))...,
		)

		res = append(res, r)
	}

	return res
}

func getFieldSelector(nodeName string) string {
	return "spec.nodeName=" + nodeName
}

func buildCountPoints(name string, counter map[string]int) []pointV2 {
	var pts []pointV2
	t := time.Now()

	for ns, count := range counter {
		pt := point.NewPointV2(
			"kubernetes",
			point.KVs{
				point.NewKV("namespace", ns, point.WithKVTagSet(true)),
				point.NewKV(name, count),
			},
			append(point.DefaultMetricOptions(), point.WithTime(t))...,
		)
		pts = append(pts, pt)
	}
	return pts
}
