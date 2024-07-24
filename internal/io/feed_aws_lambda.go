// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2024-present Guance, Inc.

package io

import (
	"context"

	"github.com/GuanceCloud/cliutils/point"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/datakit"
)

type awsLambdaOutput struct {
	cache map[point.Category]*SafeSlice[[]*point.Point, *point.Point]
}

func (a *awsLambdaOutput) Write(fo *feedOption) error {
	if fo.syncSend {
		defer a.flush()
	}
	if len(fo.pts) == 0 {
		return nil
	}
	a.cache[fo.cat].Append(fo.pts)

	return nil
}

func (a *awsLambdaOutput) WriteLastError(err string, opts ...LastErrorOption) {
	writeLastError(err, opts...)
}

func (a *awsLambdaOutput) Reader(_ point.Category) <-chan *feedOption {
	panic("unsupported")
}

func (a *awsLambdaOutput) flush() {
	for i := 0; i < 2; i++ {
		for cat, arr := range a.cache {
			pts := arr.Reset(true)
			if len(pts) == 0 {
				continue
			}
			err := defIO.doFlush(pts, cat, nil)
			if err != nil {
				log.Warnf("post %d points to %s failed: %s, ignored", len(pts), cat, err)
			}
			datakit.PutbackPoints(pts...)
		}
	}
}

func NewAwsLambdaOutput() FeederOutputer {
	fo := &awsLambdaOutput{
		cache: map[point.Category]*SafeSlice[[]*point.Point, *point.Point]{},
	}
	for _, category := range point.AllCategories() {
		if category == point.Metric || category == point.Logging || category == point.Tracing {
			fo.cache[category] = NewSafeSlice[[]*point.Point](256)
			continue
		}
		fo.cache[category] = NewSafeSlice[[]*point.Point](0)
	}
	g := datakit.G("io/aws_lambda_output")
	g.Go(func(ctx context.Context) error {
		select {
		case <-ctx.Done():
		case <-datakit.Exit.Wait():
		}
		fo.flush()
		return nil
	})
	return fo
}
