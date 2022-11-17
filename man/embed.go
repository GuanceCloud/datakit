// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package man

import (
	"embed"
)

//go:embed manuals/*.md
var docs embed.FS

var OtherDocs = map[string]bool{
	"confd":                              true,
	"datakit-sink-guide":                 true,
	"datakit-sink-dev":                   true,
	"datakit-sink-influxdb":              true,
	"datakit-sink-logstash":              true,
	"datakit-sink-m3db":                  true,
	"datakit-sink-otel-jaeger":           true,
	"datakit-sink-dataway":               true,
	"doc-logging":                        true,
	"apis":                               true,
	"changelog":                          true,
	"datakit-batch-deploy":               true,
	"datakit-conf":                       true,
	"datakit-input-conf":                 true,
	"datakit-daemonset-deploy":           true,
	"datakit-dql-how-to":                 true,
	"datakit-filter":                     true,
	"datakit-logging-how":                true,
	"datakit-install":                    true,
	"datakit-logging":                    true,
	"datakit-monitor":                    true,
	"datakit-offline-install":            true,
	"datakit-on-public":                  true,
	"datakit-pl-how-to":                  true,
	"datakit-pl-global":                  true,
	"datakit-refer-table":                true,
	"datakit-service-how-to":             true,
	"datakit-tools-how-to":               true,
	"datakit-tracing":                    true,
	"datakit-tracing-struct":             true,
	"datakit-tracing-introduction":       true,
	"datakit-update":                     true,
	"datatypes":                          true,
	"dataway":                            true,
	"dca":                                true,
	"profile-ddtrace":                    true,
	"profile-java-async-profiler":        true,
	"ddtrace-golang":                     true,
	"ddtrace-java":                       true,
	"ddtrace-python":                     true,
	"ddtrace-php":                        true,
	"ddtrace-nodejs":                     true,
	"ddtrace-cpp":                        true,
	"ddtrace-ruby":                       true,
	"development":                        true,
	"dialtesting_json":                   true,
	"election":                           true,
	"k8s-config-how-to":                  true,
	"kubernetes-prom":                    true,
	"kubernetes-crd":                     true,
	"kubernetes-prometheus-operator-crd": true,
	"logfwd":                             true,
	"logging-pipeline-bench":             true,
	"logging_socket":                     true,
	"opentelemetry-go":                   true,
	"opentelemetry-java":                 true,
	"pipeline":                           true,
	"prometheus":                         true,
	"rum":                                true,
	"sec-checker":                        true,
	"snmp":                               true,
	"telegraf":                           true,
	"why-no-data":                        true,
	"git-config-how-to":                  true,
}
