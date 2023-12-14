// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package monitor

import (
	"crypto/tls"
	"net/http"

	dto "github.com/prometheus/client_model/go"
	"github.com/prometheus/common/expfmt"
)

func requestMetrics(url string) (map[string]*dto.MetricFamily, error) {
	http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true} // nolint:gosec
	resp, err := http.Get(url)                                                                      // nolint:gosec
	if err != nil {
		return nil, err
	}

	var psr expfmt.TextParser

	mfs, err := psr.TextToMetricFamilies(resp.Body)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close() //nolint:errcheck

	return mfs, err
}
