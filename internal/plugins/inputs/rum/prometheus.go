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

var (
	namespace = "datakit"
	subSystem = "rum"
)

var ClientRealIPCounter = prometheus.NewCounterVec(prometheus.CounterOpts{
	Namespace: namespace,
	Subsystem: subSystem,
	Name:      "locate_statistics_total",
	Help:      "locate by ip addr statistics",
}, []string{"ip_status", "locate_status"})

var sourceMapCount = prometheus.NewCounterVec(prometheus.CounterOpts{
	Namespace: namespace,
	Subsystem: subSystem,
	Name:      "source_map_total",
	Help:      "source map result statistics",
},
	[]string{"sdk_name", "status", "remark"},
)

var loadedZipGauge = prometheus.NewGaugeVec(prometheus.GaugeOpts{
	Namespace: namespace,
	Subsystem: subSystem,
	Name:      "loaded_zip_cnt",
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

type sourceMapStatus struct {
	sdkName string
	status  string
	remark  string
}

type ipLocationStatus struct {
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
