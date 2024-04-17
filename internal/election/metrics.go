// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package election

import (
	"time"

	"github.com/GuanceCloud/cliutils/metrics"
	p8s "github.com/prometheus/client_golang/prometheus"
	dto "github.com/prometheus/client_model/go"
)

var (
	inputsPauseVec,
	inputsResumeVec *p8s.CounterVec

	electionVec *p8s.SummaryVec

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
			Help:      "Datakit election status, if metric = 0, meas not elected, or the elected time(unix timestamp second)",
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
			Help:      "Datakit election input count",
		},
		[]string{
			"namespace",
		},
	)

	electionVec = p8s.NewSummaryVec(
		p8s.SummaryOpts{
			Namespace: "datakit",
			Name:      "election_seconds",
			Help:      "Election latency",

			Objectives: map[float64]float64{
				0.5:  0.05,
				0.9:  0.01,
				0.99: 0.001,
			},
		}, []string{
			"namespace",
			"status",
		},
	)

	metrics.MustRegister(
		inputsPauseVec,
		inputsResumeVec,
		electionVec,
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
	ElectedTime time.Duration `json:"elected_time"`
	Status      string        `json:"status"`
}

func MetricElectionInfo(mf *dto.MetricFamily) *ElectionInfo {
	if len(mf.Metric) != 1 {
		return nil
	}

	m := mf.Metric[0]
	lps := m.GetLabel()
	if len(lps) != 4 {
		return nil
	}

	res := &ElectionInfo{
		WhoElected: lps[0].GetValue(),
		Namespace:  lps[2].GetValue(),
		Status:     lps[3].GetValue(),
	}

	x := int64(m.GetGauge().GetValue())
	if x > int64(statusFail) { // elected ok: if elect ok, there is a unix timestamp
		res.ElectedTime = time.Since(time.Unix(x, 0))
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
