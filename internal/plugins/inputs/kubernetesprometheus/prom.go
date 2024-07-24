// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package kubernetesprometheus

import (
	"context"
	"os"
	"time"

	"github.com/GuanceCloud/cliutils/point"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/datakit"
	dkio "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/io"
	iprom "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/prom"
)

func runPromCollect(ctx context.Context, interval time.Duration, urlstr string, opts []iprom.PromOption) error {
	pm, err := iprom.NewProm(opts...)
	if err != nil {
		return err
	}

	tick := time.NewTicker(interval)
	defer tick.Stop()

	for {
		if _, err := pm.CollectFromHTTPV2(urlstr); err != nil {
			klog.Warn(err)
		} else {
			klog.Debugf("collect once %s", urlstr)
		}

		select {
		case <-datakit.Exit.Wait():
			klog.Infof("prom %s exit", urlstr)
			return nil

		case <-ctx.Done():
			klog.Debugf("prom %s stop", urlstr)
			return nil

		case <-tick.C:
			// next
		}
	}
}

func buildPromOptions(role Role, key string, feeder dkio.Feeder, opts ...iprom.PromOption) []iprom.PromOption {
	callbackFn := func(pts []*point.Point) error {
		if len(pts) == 0 {
			return nil
		}

		if err := feeder.FeedV2(
			point.Metric,
			pts,
			dkio.WithInputName(inputName),
			dkio.DisableGlobalTags(true),
			dkio.WithElection(true),
			dkio.WithBlocking(true),
		); err != nil {
			klog.Warnf("failed to feed prom metrics: %s, ignored", err)
		}

		collectPtsCounter.WithLabelValues(string(role), key).Add(float64(len(pts)))
		return nil
	}

	res := []iprom.PromOption{
		iprom.WithLogger(klog), // WithLogger must in the first
		iprom.WithSource(string(role) + "/" + key),
		iprom.WithMaxBatchCallback(1, callbackFn),
	}
	res = append(res, opts...)
	return res
}

func buildPromOptionsWithAuth(auth *Auth) ([]iprom.PromOption, error) {
	var opts []iprom.PromOption

	if auth.BearerTokenFile != "" {
		token, err := os.ReadFile(auth.BearerTokenFile)
		if err != nil {
			return nil, err
		}
		opts = append(opts, iprom.WithBearerToken(string(token)))
	}

	if auth.TLSConfig != nil {
		opts = append(
			opts,
			iprom.WithTLSOpen(true),
			iprom.WithCacertFiles(auth.TLSConfig.CaCerts),
			iprom.WithCertFile(auth.TLSConfig.Cert),
			iprom.WithKeyFile(auth.TLSConfig.CertKey),
			iprom.WithInsecureSkipVerify(auth.TLSConfig.InsecureSkipVerify),
		)
	}

	return opts, nil
}
