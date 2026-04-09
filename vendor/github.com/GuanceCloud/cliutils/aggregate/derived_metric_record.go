package aggregate

import (
	"time"

	"github.com/GuanceCloud/cliutils/point"
)

const TailSamplingDerivedMeasurement = "tail_sampling"

type DerivedMetricStage string

const (
	DerivedMetricStageIngest      DerivedMetricStage = "ingest"
	DerivedMetricStagePreDecision DerivedMetricStage = "pre_decision"
	DerivedMetricStageDecision    DerivedMetricStage = "decision"
)

type DerivedMetricKind string

const (
	DerivedMetricKindSum       DerivedMetricKind = "sum"
	DerivedMetricKindHistogram DerivedMetricKind = "histogram"
)

type DerivedMetricDecision string

const (
	DerivedMetricDecisionUnknown DerivedMetricDecision = ""
	DerivedMetricDecisionKept    DerivedMetricDecision = "kept"
	DerivedMetricDecisionDropped DerivedMetricDecision = "dropped"
)

// DerivedMetricRecord is the lightweight event produced by tail-sampling hooks.
type DerivedMetricRecord struct {
	Token       string
	DataType    string
	MetricName  string
	Kind        DerivedMetricKind
	Stage       DerivedMetricStage
	Decision    DerivedMetricDecision
	Measurement string
	Tags        map[string]string
	Value       float64
	Buckets     []float64
	Time        time.Time
}

func (r DerivedMetricRecord) measurement() string {
	if r.Measurement != "" {
		return r.Measurement
	}

	return TailSamplingDerivedMeasurement
}

type DerivedMetricPoints struct {
	Token string
	PTS   []*point.Point
}
