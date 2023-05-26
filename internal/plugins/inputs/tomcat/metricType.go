// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package tomcat

var TomcatMetricType = map[string]string{
	"requestCount":       "int",
	"bytesReceived":      "int",
	"bytesSent":          "int",
	"processingTime":     "int",
	"errorCount":         "int",
	"jspCount":           "int",
	"jspReloadCount":     "int",
	"jspUnloadCount":     "int",
	"maxTHreads":         "int",
	"currentThreadCount": "int",
	"currentThreadsBusy": "int",
	"hitCount":           "int",
	"lookupCount":        "int",
}
