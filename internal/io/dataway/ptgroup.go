// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package dataway

import (
	"strings"
	sync "sync"

	"github.com/GuanceCloud/cliutils/point"
)

var grouperPool sync.Pool

type groupedPoints []*point.Point

type ptGrouper struct {
	pt  *point.Point
	cat point.Category

	kvarr []string

	extKVs     [][2]string
	groupedPts map[string]groupedPoints
}

func getGrouper() *ptGrouper {
	if x := grouperPool.Get(); x == nil {
		return &ptGrouper{
			groupedPts: map[string]groupedPoints{},
		}
	} else {
		return x.(*ptGrouper)
	}
}

func putGrouper(g *ptGrouper) {
	g.pt = nil
	g.cat = point.UnknownCategory
	g.extKVs = g.extKVs[:0]

	g.kvarr = g.kvarr[:0]

	for k := range g.groupedPts {
		delete(g.groupedPts, k)
	}

	grouperPool.Put(g)
}

func (ptg *ptGrouper) setExtKVs() {
	ptg.extKVs = append(ptg.extKVs, [2]string{"category", ptg.cat.String()})

	switch ptg.cat {
	case
		point.Logging,
		point.Network,
		point.KeyEvent,
		point.RUM:

		// set measurement name as tag `source'
		ptg.extKVs = append(ptg.extKVs, [2]string{"source", ptg.pt.Name()})

	case
		point.Tracing,
		point.Security,
		point.Profiling:
		// using measurement name as tag `service'.

	case point.Metric, point.MetricDeprecated:
		// set measurement name as tag `measurement'
		ptg.extKVs = append(ptg.extKVs, [2]string{"measurement", ptg.pt.Name()})

	case point.Object, point.CustomObject:
		// set measurement name as tag `class'
		ptg.extKVs = append(ptg.extKVs, [2]string{"class", ptg.pt.Name()})

	case point.DynamicDWCategory, point.UnknownCategory:
		// pass
	}
}

// SinkHeaderValueFromTags generate HTTP header value of key X-Global-Tags from tags.
func SinkHeaderValueFromTags(tags, globalTags map[string]string, customerKeys []string) string {
	if len(globalTags) == 0 && len(customerKeys) == 0 {
		return ""
	}

	if len(tags) == 0 {
		return ""
	}

	g := getGrouper()
	defer putGrouper(g)

	var arr []string

	for k, v := range tags {
		if x := getGroupValue(k, v, globalTags, customerKeys); x != "" {
			arr = append(arr, x)
		}
	}

	if len(arr) == 0 {
		return ""
	}

	return strings.Join(arr, ",")
}

func getGroupValue(k, v string,
	globalTags map[string]string,
	customerKeys []string,
) string {
	if _, ok := globalTags[k]; ok {
		return k + "=" + v
	}

	for _, ck := range customerKeys {
		if k == ck { // append customer tag key's value
			return k + "=" + v
		}
	}

	return ""
}

// sinkHeaderValue create X-Global-Tags header value.
func (ptg *ptGrouper) sinkHeaderValue(globalTags map[string]string, customerKeys []string) string {
	if len(globalTags) == 0 && len(customerKeys) == 0 {
		return ""
	}

	ptg.setExtKVs()

	for _, kv := range ptg.pt.KVs() {
		switch kv.Val.(type) {
		case *point.Field_S: // only accept key-value from string-type KVs
			if x := getGroupValue(kv.Key, kv.GetS(), globalTags, customerKeys); x != "" {
				ptg.kvarr = append(ptg.kvarr, x)
			}
		default: // ignored
		}
	}

	for _, ekv := range ptg.extKVs {
		if x := getGroupValue(ekv[0], ekv[1], globalTags, customerKeys); x != "" {
			ptg.kvarr = append(ptg.kvarr, x)
		}
	}

	if len(ptg.kvarr) == 0 {
		return ""
	}

	return strings.Join(ptg.kvarr, ",")
}
