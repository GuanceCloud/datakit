// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package election

import (
	"github.com/GuanceCloud/cliutils/metrics"
	p8s "github.com/prometheus/client_golang/prometheus"
	dto "github.com/prometheus/client_model/go"
)

var (
	inputsPauseVec,
	electionStatusSwitched,
	inputsResumeVec *p8s.CounterVec

	electionInputs,
	electionStatusVec *p8s.GaugeVec
)

func metricsSetup() {
	inputsPauseVec = p8s.NewCounterVec(
		p8s.CounterOpts{
			Namespace: "datakit",
			Subsystem: "election",
			Name:      "pause_total",
			Help:      "Input paused count when election failed",
		},
		[]string{
			"id",
			"namespace",
		},
	)

	inputsResumeVec = p8s.NewCounterVec(
		p8s.CounterOpts{
			Namespace: "datakit",
			Subsystem: "election",
			Name:      "resume_total",
			Help:      "Input resume count when election OK",
		},
		[]string{
			"id",
			"namespace",
		},
	)

	electionStatusVec = p8s.NewGaugeVec(
		p8s.GaugeOpts{
			Namespace: "datakit",
			Subsystem: "election",
			Name:      "status",
			Help:      "DataKit election status, if metric = 0, meas not elected, or the elected time(unix timestamp second)",
		},
		[]string{
			"elected_id", // who elected ok in current namespace
			"id",
			"namespace",
			"status",
		},
	)

	electionInputs = p8s.NewGaugeVec(
		p8s.GaugeOpts{
			Namespace: "datakit",
			Subsystem: "election",
			Name:      "inputs",
			Help:      "DataKit election input count",
		},
		[]string{
			"namespace",
		},
	)

	electionStatusSwitched = p8s.NewCounterVec(
		p8s.CounterOpts{
			Namespace: "datakit",
			Name:      "election_switched_total",
			Help:      "Election status switch count",
		}, []string{
			"namespace",
			"status",
		},
	)

	metrics.MustRegister(
		inputsPauseVec,
		inputsResumeVec,
		electionStatusSwitched,
		electionStatusVec,
		electionInputs,
	)
}

//nolint:gochecknoinits
func init() {
	metricsSetup()
}

// A ElectionInfo is the election info of current datakit.
type ElectionInfo struct {
	Namespace  string `json:"namespace"`
	WhoElected string `json:"elected_hostname"`

	// elected ok duration
	UpdateTime int64  `json:"elected_time"`
	Status     string `json:"status"`
}

func MetricElectionInfo(mf *dto.MetricFamily) (res *ElectionInfo) {
	for _, m := range mf.Metric {
		lps := m.GetLabel()
		if len(lps) != 4 {
			// should not been here
			log.Warn("invalid labels, got %+#v", lps)
			continue
		}

		ei := &ElectionInfo{
			WhoElected: lps[0].GetValue(),
			Namespace:  lps[2].GetValue(),
			Status:     lps[3].GetValue(),
			UpdateTime: int64(m.GetGauge().GetValue()),
		}

		if res == nil {
			res = ei
		} else if ei.UpdateTime > res.UpdateTime { // use latest updated value
			res = ei
		}

		log.Debugf("election info: %+#v", res)
	}

	return res
}

func GetElectionInfo(mfs []*dto.MetricFamily) *ElectionInfo {
	for _, mf := range mfs {
		if mf.GetName() == "datakit_election_status" {
			return MetricElectionInfo(mf)
		}
	}

	return nil
}
