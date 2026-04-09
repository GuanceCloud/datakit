// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package dataway

import (
	"net/url"
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
	safe  bool

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

func (g *ptGrouper) reset() {
	g.pt = nil
	g.cat = point.UnknownCategory
	g.extKVs = g.extKVs[:0]

	g.kvarr = g.kvarr[:0]
	g.safe = false

	for k := range g.groupedPts {
		delete(g.groupedPts, k)
	}
}

func putGrouper(g *ptGrouper) {
	g.reset()
	grouperPool.Put(g)
}

func (g *ptGrouper) setExtraKVs() {
	g.extKVs = append(g.extKVs, [2]string{"category", g.cat.String()})

	switch g.cat {
	case
		point.Logging,
		point.ObjectChange, // Deprecated.
		point.DialTesting,
		point.Network,
		point.KeyEvent,
		point.RUM:

		// set measurement name as tag `source'
		g.extKVs = append(g.extKVs, [2]string{"source", g.pt.Name()})

	case
		point.Tracing,
		point.Security,
		point.Profiling:
		// using measurement name as tag `service'.

	case point.Metric, point.MetricDeprecated:
		// set measurement name as tag `measurement'
		g.extKVs = append(g.extKVs, [2]string{"measurement", g.pt.Name()})

	case point.Object, point.CustomObject:
		// set measurement name as tag `class'
		g.extKVs = append(g.extKVs, [2]string{"class", g.pt.Name()})

	case point.ExecutionLog, point.LLM, point.DynamicDWCategory, point.UnknownCategory:
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
		if x := g.getGroupValue(k, v, globalTags, customerKeys); x != "" {
			arr = append(arr, x)
		}
	}

	if len(arr) == 0 {
		return ""
	}

	return strings.Join(arr, ",")
}

func (g *ptGrouper) getGroupValue(k, v string,
	globalTags map[string]string,
	customerKeys []string,
) string {
	if _, ok := globalTags[k]; ok {
		if g.safe {
			return url.QueryEscape(k) + "=" + url.QueryEscape(v)
		} else {
			return k + "=" + v
		}
	}

	for _, ck := range customerKeys {
		if k == ck { // append customer tag key's value
			if g.safe {
				return url.QueryEscape(k) + "=" + url.QueryEscape(v)
			} else {
				return (k) + "=" + v
			}
		}
	}

	return ""
}

// sinkHeaderValue create X-Global-Tags header value.
func (g *ptGrouper) sinkHeaderValue(globalTags map[string]string, customerKeys []string) string {
	if len(globalTags) == 0 && len(customerKeys) == 0 {
		return ""
	}

	g.setExtraKVs()

	for _, kv := range g.pt.KVs() {
		switch kv.Val.(type) {
		case *point.Field_S: // only accept key-value from string-type KVs
			if x := g.getGroupValue(kv.Key, kv.GetS(), globalTags, customerKeys); x != "" {
				g.kvarr = append(g.kvarr, x)
			}
		default: // ignored
		}
	}

	for _, ekv := range g.extKVs {
		if x := g.getGroupValue(ekv[0], ekv[1], globalTags, customerKeys); x != "" {
			g.kvarr = append(g.kvarr, x)
		}
	}

	if len(g.kvarr) == 0 {
		return ""
	}

	return strings.Join(g.kvarr, ",")
}
