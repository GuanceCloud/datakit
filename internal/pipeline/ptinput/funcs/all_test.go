// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package funcs

import (
	"fmt"
	"time"

	"github.com/GuanceCloud/cliutils/point"
	"github.com/GuanceCloud/platypus/pkg/engine"
	"github.com/GuanceCloud/platypus/pkg/engine/runtime"
	"github.com/GuanceCloud/platypus/pkg/errchain"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/pipeline/ptinput"
)

func NewTestingRunner(script string) (*runtime.Script, error) {
	name := "default.p"
	ret1, ret2 := engine.ParseScript(map[string]string{
		"default.p": script,
	},
		FuncsMap, FuncsCheckMap,
	)
	if len(ret1) > 0 {
		return ret1[name], nil
	}
	if len(ret2) > 0 {
		return nil, ret2[name]
	}
	return nil, fmt.Errorf("parser func error")
}

func NewTestingRunner2(scripts map[string]string) (map[string]*runtime.Script, map[string]error) {
	return engine.ParseScript(scripts, FuncsMap, FuncsCheckMap)
}

func runScript(proc *runtime.Script, pt ptinput.PlInputPt) *errchain.PlError {
	return engine.RunScriptWithRMapIn(proc, pt, nil)
}

func newPoint(cat point.Category, name string, tags map[string]string, fields map[string]any,
	ts ...time.Time,
) *point.Point {
	var t time.Time
	if len(ts) > 0 {
		t = ts[0]
	} else {
		t = time.Now()
	}

	var opt []point.Option
	switch cat { //nolint:exhaustive
	case point.Metric, point.MetricDeprecated:
		opt = point.DefaultMetricOptions()
	case point.Object, point.CustomObject:
		opt = point.DefaultObjectOptions()
	case point.Logging:
		opt = point.DefaultLoggingOptions()
	default:
		opt = []point.Option{
			point.WithDisabledKeys(point.KeySource, point.KeyDate),
			point.WithMaxFieldValLen(32 * 1024 * 1024),
			point.WithDotInKey(false),
		}
	}

	opt = append(opt, point.WithTime(t))

	kCount := len(tags) + len(fields)
	kvs := make(point.KVs, 0, kCount)
	kvs = append(kvs, point.NewTags(tags)...)
	kvs = append(kvs, point.NewKVs(fields)...)
	return point.NewPointV2([]byte(name), kvs, opt...)
}
