// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package discovery

import metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

const (
	annotationPromExport  = "datakit/prom.instances"
	annotationPromIPIndex = "datakit/prom.instances.ip_index"

	annotationPrometheusioScrape           = "prometheus.io/scrape"
	annotationPrometheusioPort             = "prometheus.io/port"
	annotationPrometheusioPath             = "prometheus.io/path"
	annotationPrometheusioScheme           = "prometheus.io/scheme"
	annotationPrometheusioParamMeasurement = "prometheus.io/param_measurement"
)

var (
	defaultPromScheme = "http"
	defaultPromPath   = "/metrics"

	metaV1ListOption = metav1.ListOptions{ResourceVersion: "0"}
	metaV1GetOption  = metav1.GetOptions{ResourceVersion: "0"}
)
