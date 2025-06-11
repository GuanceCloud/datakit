// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package kubernetes

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/GuanceCloud/cliutils/point"
	"github.com/google/uuid"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/diff"
	dkio "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/io"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/yaml"
)

func feedMetric(name string, feeder dkio.Feeder, pts []*point.Point, election bool) {
	if feeder == nil || len(pts) == 0 {
		return
	}

	collectPtsVec.WithLabelValues(name).Add(float64(len(pts)))

	if err := feeder.FeedV2(
		point.Metric,
		pts,
		dkio.WithElection(election),
		dkio.WithInputName("k8s-metric"),
	); err != nil {
		klog.Warnf("%s feed failed, err: %s", name, err)
	}
}

func feedObject(name string, feeder dkio.Feeder, pts []*point.Point, election bool) {
	if feeder == nil || len(pts) == 0 {
		return
	}

	collectPtsVec.WithLabelValues(name).Add(float64(len(pts)))

	if err := feeder.FeedV2(
		point.Object,
		pts,
		dkio.WithElection(election),
		dkio.WithInputName("k8s-object"),
	); err != nil {
		klog.Warnf("%s feed failed, err: %s", name, err)
	}
}

func feedLogging(name string, feeder dkio.Feeder, pts []*point.Point) {
	if feeder == nil || len(pts) == 0 {
		return
	}

	collectPtsVec.WithLabelValues(name).Add(float64(len(pts)))

	if err := feeder.FeedV2(
		point.Logging,
		pts,
		dkio.WithElection(true),
		dkio.WithInputName("k8s-event"),
	); err != nil {
		klog.Warnf("%s feed failed, err: %s", name, err)
	}
}

func processChange(cfg *Config, class, sourceName, sourceType, difftext string, obj metav1.Object) {
	var kvs point.KVs
	kvs = append(kvs, buildDefaultChangeEventKVs()...)

	kvs = kvs.AddTag("class", class)
	kvs = kvs.AddTag("uid", string(obj.GetUID()))
	kvs = kvs.AddTag("namespace", obj.GetNamespace())

	name := obj.GetName()
	kvs = kvs.AddTag(sourceName, name)

	content := fmt.Sprintf("[%s] %s configuration changed", sourceType, name)
	kvs = kvs.AddV2("df_title", content, false)
	kvs = kvs.AddV2("df_detail", content, false)
	kvs = kvs.AddV2("df_message", difftext, false)

	kvs = append(kvs, point.NewTags(cfg.ExtraTags)...)

	pt := point.NewPointV2("event", kvs, point.WithTimestamp(time.Now().UnixNano()))
	collectPtsVec.WithLabelValues("k8s-object-change-event").Add(1)

	if err := cfg.Feeder.FeedV2(
		point.KeyEvent,
		[]*point.Point{pt},
		dkio.WithElection(true),
		dkio.WithInputName("k8s-object-change-event"),
	); err != nil {
		klog.Warnf("feed failed, err: %s", err)
	}
}

func processCounter(cfg *Config, name string, counter map[string]int, timestamp int64 /*nanoseconds*/) {
	var pts []*point.Point
	opts := point.DefaultMetricOptions()

	for ns, count := range counter {
		var kvs point.KVs
		kvs = kvs.AddTag("namespace", ns)
		kvs = kvs.AddV2(name, count, false)
		kvs = append(kvs, point.NewTags(cfg.ExtraTags)...)

		pt := point.NewPointV2("kubernetes", kvs, append(opts, point.WithTimestamp(timestamp))...)
		pts = append(pts, pt)
	}

	feedMetric("k8s-counter", cfg.Feeder, pts, true)
}

func diffObject(oldObj, newObj interface{}) (difftext string, err error) {
	const contextLines = 4

	oldText, err := yaml.Marshal(oldObj)
	if err != nil {
		return "", err
	}
	newText, err := yaml.Marshal(newObj)
	if err != nil {
		return "", err
	}
	return diff.LineDiffWithContextLines(string(oldText), string(newText), contextLines), nil
}

func buildDefaultChangeEventKVs() (kvs point.KVs) {
	const (
		defaultStatus = "info"
		defaultSource = "change"
	)

	var uid string
	if u, err := uuid.NewRandom(); err == nil {
		uid = "event-" + strings.ToLower(u.String())
	} else {
		klog.Warnf("cannot generate UUIDv4, err: %s", err)
	}
	kvs = kvs.AddTag("df_event_id", uid)
	kvs = kvs.AddTag("df_source", defaultSource)
	kvs = kvs.AddTag("df_status", defaultStatus)
	kvs = kvs.AddTag("df_sub_status", defaultStatus)

	return
}

func getLocalNodeName() (string, error) {
	var e string
	if os.Getenv("NODE_NAME") != "" {
		e = os.Getenv("NODE_NAME")
	}
	if os.Getenv("ENV_K8S_NODE_NAME") != "" {
		e = os.Getenv("ENV_K8S_NODE_NAME")
	}
	if e != "" {
		return e, nil
	}
	return "", fmt.Errorf("invalid ENV_K8S_NODE_NAME environment, cannot be empty")
}
