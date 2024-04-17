// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package rum

import (
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

const promDurationUnit = float64(time.Second)

const (
	StatusOK           = "success"
	StatusLackField    = "lack-field"
	StatusError        = "error"
	StatusUnknown      = "unknown"
	StatusToolNotFound = "no-tool"
	StatusZipNotFound  = "no-zip"
)

const (
	IPStatusRemoteAddr = "remote-addr"
	IPStatusIllegal    = "illegal"
	IPStatusPrivate    = "private"
	IPStatusPublic     = "public"
)

const (
	LocateStatusGEOSuccess = "success"
	LocateStatusGEOFailure = "failure"
	LocateStatusGEONil     = "nil"
)

const (
	namespace = "datakit"
	subSystem = "input_rum"
)

var ClientRealIPCounter = prometheus.NewCounterVec(prometheus.CounterOpts{
	Namespace: namespace,
	Subsystem: subSystem,
	Name:      "locate_statistics_total",
	Help:      "locate by ip addr statistics",
}, []string{"app_id", "ip_status", "locate_status"})

var sourceMapCount = prometheus.NewCounterVec(prometheus.CounterOpts{
	Namespace: namespace,
	Subsystem: subSystem,
	Name:      "source_map_total",
	Help:      "source map result statistics",
},
	[]string{"app_id", "sdk_name", "status", "remark"},
)

var loadedZipGauge = prometheus.NewGaugeVec(prometheus.GaugeOpts{
	Namespace: namespace,
	Subsystem: subSystem,
	Name:      "loaded_zips",
	Help:      "RUM source map currently loaded zip archive count",
},
	[]string{"platform"},
)

var sourceMapDurationSummary = prometheus.NewSummaryVec(prometheus.SummaryOpts{
	Namespace:  namespace,
	Subsystem:  subSystem,
	Name:       "source_map_duration_seconds",
	Help:       "statistics elapsed time in RUM source map(unit: second)",
	Objectives: map[float64]float64{0.5: 0.05, 0.9: 0.01, 0.99: 0.001},
},
	[]string{"sdk_name", "app_id", "env", "version"},
)

var replayUploadingDurationSummary = prometheus.NewSummaryVec(prometheus.SummaryOpts{
	Namespace:  namespace,
	Subsystem:  subSystem,
	Name:       "session_replay_upload_latency_seconds",
	Help:       "statistics elapsed time in session replay uploading",
	Objectives: map[float64]float64{0.5: 0.05, 0.9: 0.01, 0.99: 0.001},
},
	[]string{"app_id", "env", "version", "service", "status_code"},
)

var replayFailureTotalCount = prometheus.NewCounterVec(prometheus.CounterOpts{
	Namespace: namespace,
	Subsystem: subSystem,
	Name:      "session_replay_upload_failure_total",
	Help:      "statistics count of session replay points which which have unsuccessfully uploaded",
}, []string{"app_id", "env", "version", "service", "status_code"})

var replayFailureTotalBytes = prometheus.NewCounterVec(prometheus.CounterOpts{
	Namespace: namespace,
	Subsystem: subSystem,
	Name:      "session_replay_upload_failure_bytes_total",
	Help:      "statistics the total bytes of session replay points which have unsuccessfully uploaded",
}, []string{"app_id", "env", "version", "service", "status_code"})

var replayReadBodyDelaySeconds = prometheus.NewSummaryVec(prometheus.SummaryOpts{
	Namespace:  namespace,
	Subsystem:  subSystem,
	Name:       "session_replay_read_body_delay_seconds",
	Help:       "statistics the duration of reading session replay body",
	Objectives: map[float64]float64{0.5: 0.05, 0.9: 0.01, 0.99: 0.001},
}, []string{"app_id", "env", "version", "service"})

var replayFilteredTotalCount = prometheus.NewCounterVec(prometheus.CounterOpts{
	Namespace: namespace,
	Subsystem: subSystem,
	Name:      "session_replay_drop_total",
	Help:      "statistics the total count of session replay points which have been filtered by rules",
}, []string{"app_id", "env", "version", "service"})

var replayFilteredTotalBytes = prometheus.NewCounterVec(prometheus.CounterOpts{
	Namespace: namespace,
	Subsystem: subSystem,
	Name:      "session_replay_drop_bytes_total",
	Help:      "statistics the total bytes of session replay points which have been filtered by rules",
}, []string{"app_id", "env", "version", "service"})

type sourceMapStatus struct {
	sdkName string
	appid   string
	status  string
	remark  string
	reason  string
}

type ipLocationStatus struct {
	appid        string
	ipStatus     string
	locateStatus string
}

func utf8SubStr(s string, cnt int) string {
	if cnt > len(s) {
		cnt = len(s)
	}
	runes := make([]rune, 0, cnt/4)

	for _, r := range s {
		runes = append(runes, r)
		cnt--
		if cnt <= 0 {
			break
		}
	}
	return string(runes)
}
