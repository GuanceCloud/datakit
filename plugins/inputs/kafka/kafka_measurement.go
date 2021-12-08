package kafka

import (
	"time"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
)

type KafkaMeasurement struct {
	name   string
	tags   map[string]string
	fields map[string]interface{}
	ts     time.Time
}

type KafkaControllerMment struct {
	KafkaMeasurement
}

type KafkaReplicaMment struct {
	KafkaMeasurement
}

type KafkaPurgatoryMment struct {
	KafkaMeasurement
}

type KafkaClientMment struct {
	KafkaMeasurement
}

type KafkaRequestMment struct {
	KafkaMeasurement
}

type KafkaTopicsMment struct {
	KafkaMeasurement
}

type KafkaTopicMment struct {
	KafkaMeasurement
}

type KafkaPartitionMment struct {
	KafkaMeasurement
}

func (j *KafkaMeasurement) LineProto() (*io.Point, error) {
	return io.MakeTypedPoint(j.name, datakit.Metric, j.tags, j.fields, j.ts)
}

func (j *KafkaControllerMment) LineProto() (*io.Point, error) {
	return io.MakeTypedPoint(j.name, datakit.Metric, j.tags, j.fields, j.ts)
}

//nolint:lll
func (j *KafkaControllerMment) Info() *inputs.MeasurementInfo { //nolint:funlen
	return &inputs.MeasurementInfo{
		Name: "kafka_controller",
		Fields: map[string]interface{}{
			"AutoLeaderBalanceRateAndTimeMs.50thPercentile":    &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
			"AutoLeaderBalanceRateAndTimeMs.75thPercentile":    &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
			"AutoLeaderBalanceRateAndTimeMs.95thPercentile":    &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
			"AutoLeaderBalanceRateAndTimeMs.98thPercentile":    &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
			"AutoLeaderBalanceRateAndTimeMs.999thPercentile":   &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
			"AutoLeaderBalanceRateAndTimeMs.99thPercentile":    &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
			"AutoLeaderBalanceRateAndTimeMs.FifteenMinuteRate": &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
			"AutoLeaderBalanceRateAndTimeMs.FiveMinuteRate":    &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
			"AutoLeaderBalanceRateAndTimeMs.Max":               &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
			"AutoLeaderBalanceRateAndTimeMs.Mean":              &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
			"AutoLeaderBalanceRateAndTimeMs.MeanRate":          &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
			"AutoLeaderBalanceRateAndTimeMs.Min":               &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
			"AutoLeaderBalanceRateAndTimeMs.OneMinuteRate":     &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
			"AutoLeaderBalanceRateAndTimeMs.StdDev":            &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
			"AutoLeaderBalanceRateAndTimeMs.Count":             &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount, Desc: ""},
			"AutoLeaderBalanceRateAndTimeMs.LatencyUnit":       &inputs.FieldInfo{DataType: inputs.String, Type: inputs.String, Unit: inputs.UnknownUnit, Desc: ""},
			"AutoLeaderBalanceRateAndTimeMs.EventType":         &inputs.FieldInfo{DataType: inputs.String, Type: inputs.String, Unit: inputs.UnknownUnit, Desc: ""},
			"AutoLeaderBalanceRateAndTimeMs.RateUnit":          &inputs.FieldInfo{DataType: inputs.String, Type: inputs.String, Unit: inputs.UnknownUnit, Desc: ""},

			"ControlledShutdownRateAndTimeMs.50thPercentile":    &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
			"ControlledShutdownRateAndTimeMs.75thPercentile":    &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
			"ControlledShutdownRateAndTimeMs.95thPercentile":    &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
			"ControlledShutdownRateAndTimeMs.98thPercentile":    &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
			"ControlledShutdownRateAndTimeMs.999thPercentile":   &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
			"ControlledShutdownRateAndTimeMs.99thPercentile":    &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
			"ControlledShutdownRateAndTimeMs.FifteenMinuteRate": &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
			"ControlledShutdownRateAndTimeMs.FiveMinuteRate":    &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
			"ControlledShutdownRateAndTimeMs.Max":               &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
			"ControlledShutdownRateAndTimeMs.Mean":              &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
			"ControlledShutdownRateAndTimeMs.MeanRate":          &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
			"ControlledShutdownRateAndTimeMs.Min":               &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
			"ControlledShutdownRateAndTimeMs.OneMinuteRate":     &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
			"ControlledShutdownRateAndTimeMs.StdDev":            &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
			"ControlledShutdownRateAndTimeMs.Count":             &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount, Desc: ""},
			"ControlledShutdownRateAndTimeMs.LatencyUnit":       &inputs.FieldInfo{DataType: inputs.String, Type: inputs.String, Unit: inputs.UnknownUnit, Desc: ""},
			"ControlledShutdownRateAndTimeMs.EventType":         &inputs.FieldInfo{DataType: inputs.String, Type: inputs.String, Unit: inputs.UnknownUnit, Desc: ""},
			"ControlledShutdownRateAndTimeMs.RateUnit":          &inputs.FieldInfo{DataType: inputs.String, Type: inputs.String, Unit: inputs.UnknownUnit, Desc: ""},

			"ControllerChangeRateAndTimeMs.50thPercentile":    &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
			"ControllerChangeRateAndTimeMs.75thPercentile":    &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
			"ControllerChangeRateAndTimeMs.95thPercentile":    &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
			"ControllerChangeRateAndTimeMs.98thPercentile":    &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
			"ControllerChangeRateAndTimeMs.999thPercentile":   &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
			"ControllerChangeRateAndTimeMs.99thPercentile":    &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
			"ControllerChangeRateAndTimeMs.FifteenMinuteRate": &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
			"ControllerChangeRateAndTimeMs.FiveMinuteRate":    &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
			"ControllerChangeRateAndTimeMs.Max":               &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
			"ControllerChangeRateAndTimeMs.Mean":              &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
			"ControllerChangeRateAndTimeMs.MeanRate":          &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
			"ControllerChangeRateAndTimeMs.Min":               &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
			"ControllerChangeRateAndTimeMs.OneMinuteRate":     &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
			"ControllerChangeRateAndTimeMs.StdDev":            &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
			"ControllerChangeRateAndTimeMs.Count":             &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount, Desc: ""},
			"ControllerChangeRateAndTimeMs.LatencyUnit":       &inputs.FieldInfo{DataType: inputs.String, Type: inputs.String, Unit: inputs.UnknownUnit, Desc: ""},
			"ControllerChangeRateAndTimeMs.EventType":         &inputs.FieldInfo{DataType: inputs.String, Type: inputs.String, Unit: inputs.UnknownUnit, Desc: ""},
			"ControllerChangeRateAndTimeMs.RateUnit":          &inputs.FieldInfo{DataType: inputs.String, Type: inputs.String, Unit: inputs.UnknownUnit, Desc: ""},

			"ControllerShutdownRateAndTimeMs.50thPercentile":    &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
			"ControllerShutdownRateAndTimeMs.75thPercentile":    &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
			"ControllerShutdownRateAndTimeMs.95thPercentile":    &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
			"ControllerShutdownRateAndTimeMs.98thPercentile":    &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
			"ControllerShutdownRateAndTimeMs.999thPercentile":   &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
			"ControllerShutdownRateAndTimeMs.99thPercentile":    &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
			"ControllerShutdownRateAndTimeMs.FifteenMinuteRate": &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
			"ControllerShutdownRateAndTimeMs.FiveMinuteRate":    &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
			"ControllerShutdownRateAndTimeMs.Max":               &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
			"ControllerShutdownRateAndTimeMs.Mean":              &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
			"ControllerShutdownRateAndTimeMs.MeanRate":          &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
			"ControllerShutdownRateAndTimeMs.Min":               &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
			"ControllerShutdownRateAndTimeMs.OneMinuteRate":     &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
			"ControllerShutdownRateAndTimeMs.StdDev":            &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
			"ControllerShutdownRateAndTimeMs.Count":             &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount, Desc: ""},
			"ControllerShutdownRateAndTimeMs.LatencyUnit":       &inputs.FieldInfo{DataType: inputs.String, Type: inputs.String, Unit: inputs.UnknownUnit, Desc: ""},
			"ControllerShutdownRateAndTimeMs.EventType":         &inputs.FieldInfo{DataType: inputs.String, Type: inputs.String, Unit: inputs.UnknownUnit, Desc: ""},
			"ControllerShutdownRateAndTimeMs.RateUnit":          &inputs.FieldInfo{DataType: inputs.String, Type: inputs.String, Unit: inputs.UnknownUnit, Desc: ""},

			"IsrChangeRateAndTimeMs.50thPercentile":    &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
			"IsrChangeRateAndTimeMs.75thPercentile":    &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
			"IsrChangeRateAndTimeMs.95thPercentile":    &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
			"IsrChangeRateAndTimeMs.98thPercentile":    &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
			"IsrChangeRateAndTimeMs.999thPercentile":   &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
			"IsrChangeRateAndTimeMs.99thPercentile":    &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
			"IsrChangeRateAndTimeMs.FifteenMinuteRate": &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
			"IsrChangeRateAndTimeMs.FiveMinuteRate":    &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
			"IsrChangeRateAndTimeMs.Max":               &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
			"IsrChangeRateAndTimeMs.Mean":              &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
			"IsrChangeRateAndTimeMs.MeanRate":          &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
			"IsrChangeRateAndTimeMs.Min":               &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
			"IsrChangeRateAndTimeMs.OneMinuteRate":     &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
			"IsrChangeRateAndTimeMs.StdDev":            &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
			"IsrChangeRateAndTimeMs.Count":             &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount, Desc: ""},
			"IsrChangeRateAndTimeMs.LatencyUnit":       &inputs.FieldInfo{DataType: inputs.String, Type: inputs.String, Unit: inputs.UnknownUnit, Desc: ""},
			"IsrChangeRateAndTimeMs.EventType":         &inputs.FieldInfo{DataType: inputs.String, Type: inputs.String, Unit: inputs.UnknownUnit, Desc: ""},
			"IsrChangeRateAndTimeMs.RateUnit":          &inputs.FieldInfo{DataType: inputs.String, Type: inputs.String, Unit: inputs.UnknownUnit, Desc: ""},

			"LeaderAndIsrResponseReceivedRateAndTimeMs.50thPercentile":    &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
			"LeaderAndIsrResponseReceivedRateAndTimeMs.75thPercentile":    &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
			"LeaderAndIsrResponseReceivedRateAndTimeMs.95thPercentile":    &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
			"LeaderAndIsrResponseReceivedRateAndTimeMs.98thPercentile":    &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
			"LeaderAndIsrResponseReceivedRateAndTimeMs.999thPercentile":   &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
			"LeaderAndIsrResponseReceivedRateAndTimeMs.99thPercentile":    &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
			"LeaderAndIsrResponseReceivedRateAndTimeMs.FifteenMinuteRate": &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
			"LeaderAndIsrResponseReceivedRateAndTimeMs.FiveMinuteRate":    &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
			"LeaderAndIsrResponseReceivedRateAndTimeMs.Max":               &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
			"LeaderAndIsrResponseReceivedRateAndTimeMs.Mean":              &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
			"LeaderAndIsrResponseReceivedRateAndTimeMs.MeanRate":          &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
			"LeaderAndIsrResponseReceivedRateAndTimeMs.Min":               &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
			"LeaderAndIsrResponseReceivedRateAndTimeMs.OneMinuteRate":     &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
			"LeaderAndIsrResponseReceivedRateAndTimeMs.StdDev":            &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
			"LeaderAndIsrResponseReceivedRateAndTimeMs.Count":             &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount, Desc: ""},
			"LeaderAndIsrResponseReceivedRateAndTimeMs.LatencyUnit":       &inputs.FieldInfo{DataType: inputs.String, Type: inputs.String, Unit: inputs.UnknownUnit, Desc: ""},
			"LeaderAndIsrResponseReceivedRateAndTimeMs.EventType":         &inputs.FieldInfo{DataType: inputs.String, Type: inputs.String, Unit: inputs.UnknownUnit, Desc: ""},
			"LeaderAndIsrResponseReceivedRateAndTimeMs.RateUnit":          &inputs.FieldInfo{DataType: inputs.String, Type: inputs.String, Unit: inputs.UnknownUnit, Desc: ""},

			"LeaderElectionRateAndTimeMs.50thPercentile":    &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
			"LeaderElectionRateAndTimeMs.75thPercentile":    &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
			"LeaderElectionRateAndTimeMs.95thPercentile":    &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
			"LeaderElectionRateAndTimeMs.98thPercentile":    &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
			"LeaderElectionRateAndTimeMs.999thPercentile":   &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
			"LeaderElectionRateAndTimeMs.99thPercentile":    &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
			"LeaderElectionRateAndTimeMs.FifteenMinuteRate": &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
			"LeaderElectionRateAndTimeMs.FiveMinuteRate":    &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
			"LeaderElectionRateAndTimeMs.Max":               &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
			"LeaderElectionRateAndTimeMs.Mean":              &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
			"LeaderElectionRateAndTimeMs.MeanRate":          &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
			"LeaderElectionRateAndTimeMs.Min":               &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
			"LeaderElectionRateAndTimeMs.OneMinuteRate":     &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
			"LeaderElectionRateAndTimeMs.StdDev":            &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
			"LeaderElectionRateAndTimeMs.Count":             &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount, Desc: ""},
			"LeaderElectionRateAndTimeMs.LatencyUnit":       &inputs.FieldInfo{DataType: inputs.String, Type: inputs.String, Unit: inputs.UnknownUnit, Desc: ""},
			"LeaderElectionRateAndTimeMs.EventType":         &inputs.FieldInfo{DataType: inputs.String, Type: inputs.String, Unit: inputs.UnknownUnit, Desc: ""},
			"LeaderElectionRateAndTimeMs.RateUnit":          &inputs.FieldInfo{DataType: inputs.String, Type: inputs.String, Unit: inputs.UnknownUnit, Desc: ""},

			"ListPartitionReassignmentRateAndTimeMs.50thPercentile":    &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
			"ListPartitionReassignmentRateAndTimeMs.75thPercentile":    &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
			"ListPartitionReassignmentRateAndTimeMs.95thPercentile":    &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
			"ListPartitionReassignmentRateAndTimeMs.98thPercentile":    &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
			"ListPartitionReassignmentRateAndTimeMs.999thPercentile":   &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
			"ListPartitionReassignmentRateAndTimeMs.99thPercentile":    &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
			"ListPartitionReassignmentRateAndTimeMs.FifteenMinuteRate": &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
			"ListPartitionReassignmentRateAndTimeMs.FiveMinuteRate":    &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
			"ListPartitionReassignmentRateAndTimeMs.Max":               &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
			"ListPartitionReassignmentRateAndTimeMs.Mean":              &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
			"ListPartitionReassignmentRateAndTimeMs.MeanRate":          &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
			"ListPartitionReassignmentRateAndTimeMs.Min":               &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
			"ListPartitionReassignmentRateAndTimeMs.OneMinuteRate":     &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
			"ListPartitionReassignmentRateAndTimeMs.StdDev":            &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
			"ListPartitionReassignmentRateAndTimeMs.Count":             &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount, Desc: ""},
			"ListPartitionReassignmentRateAndTimeMs.LatencyUnit":       &inputs.FieldInfo{DataType: inputs.String, Type: inputs.String, Unit: inputs.UnknownUnit, Desc: ""},
			"ListPartitionReassignmentRateAndTimeMs.EventType":         &inputs.FieldInfo{DataType: inputs.String, Type: inputs.String, Unit: inputs.UnknownUnit, Desc: ""},
			"ListPartitionReassignmentRateAndTimeMs.RateUnit":          &inputs.FieldInfo{DataType: inputs.String, Type: inputs.String, Unit: inputs.UnknownUnit, Desc: ""},

			"LogDirChangeRateAndTimeMs.50thPercentile":    &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
			"LogDirChangeRateAndTimeMs.75thPercentile":    &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
			"LogDirChangeRateAndTimeMs.95thPercentile":    &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
			"LogDirChangeRateAndTimeMs.98thPercentile":    &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
			"LogDirChangeRateAndTimeMs.999thPercentile":   &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
			"LogDirChangeRateAndTimeMs.99thPercentile":    &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
			"LogDirChangeRateAndTimeMs.FifteenMinuteRate": &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
			"LogDirChangeRateAndTimeMs.FiveMinuteRate":    &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
			"LogDirChangeRateAndTimeMs.Max":               &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
			"LogDirChangeRateAndTimeMs.Mean":              &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
			"LogDirChangeRateAndTimeMs.MeanRate":          &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
			"LogDirChangeRateAndTimeMs.Min":               &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
			"LogDirChangeRateAndTimeMs.OneMinuteRate":     &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
			"LogDirChangeRateAndTimeMs.StdDev":            &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
			"LogDirChangeRateAndTimeMs.Count":             &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount, Desc: ""},
			"LogDirChangeRateAndTimeMs.LatencyUnit":       &inputs.FieldInfo{DataType: inputs.String, Type: inputs.String, Unit: inputs.UnknownUnit, Desc: ""},
			"LogDirChangeRateAndTimeMs.EventType":         &inputs.FieldInfo{DataType: inputs.String, Type: inputs.String, Unit: inputs.UnknownUnit, Desc: ""},
			"LogDirChangeRateAndTimeMs.RateUnit":          &inputs.FieldInfo{DataType: inputs.String, Type: inputs.String, Unit: inputs.UnknownUnit, Desc: ""},

			"ManualLeaderBalanceRateAndTimeMs.50thPercentile":    &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
			"ManualLeaderBalanceRateAndTimeMs.75thPercentile":    &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
			"ManualLeaderBalanceRateAndTimeMs.95thPercentile":    &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
			"ManualLeaderBalanceRateAndTimeMs.98thPercentile":    &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
			"ManualLeaderBalanceRateAndTimeMs.999thPercentile":   &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
			"ManualLeaderBalanceRateAndTimeMs.99thPercentile":    &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
			"ManualLeaderBalanceRateAndTimeMs.FifteenMinuteRate": &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
			"ManualLeaderBalanceRateAndTimeMs.FiveMinuteRate":    &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
			"ManualLeaderBalanceRateAndTimeMs.Max":               &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
			"ManualLeaderBalanceRateAndTimeMs.Mean":              &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
			"ManualLeaderBalanceRateAndTimeMs.MeanRate":          &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
			"ManualLeaderBalanceRateAndTimeMs.Min":               &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
			"ManualLeaderBalanceRateAndTimeMs.OneMinuteRate":     &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
			"ManualLeaderBalanceRateAndTimeMs.StdDev":            &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
			"ManualLeaderBalanceRateAndTimeMs.Count":             &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount, Desc: ""},
			"ManualLeaderBalanceRateAndTimeMs.LatencyUnit":       &inputs.FieldInfo{DataType: inputs.String, Type: inputs.String, Unit: inputs.UnknownUnit, Desc: ""},
			"ManualLeaderBalanceRateAndTimeMs.EventType":         &inputs.FieldInfo{DataType: inputs.String, Type: inputs.String, Unit: inputs.UnknownUnit, Desc: ""},
			"ManualLeaderBalanceRateAndTimeMs.RateUnit":          &inputs.FieldInfo{DataType: inputs.String, Type: inputs.String, Unit: inputs.UnknownUnit, Desc: ""},

			"PartitionReassignmentRateAndTimeMs.50thPercentile":    &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
			"PartitionReassignmentRateAndTimeMs.75thPercentile":    &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
			"PartitionReassignmentRateAndTimeMs.95thPercentile":    &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
			"PartitionReassignmentRateAndTimeMs.98thPercentile":    &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
			"PartitionReassignmentRateAndTimeMs.999thPercentile":   &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
			"PartitionReassignmentRateAndTimeMs.99thPercentile":    &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
			"PartitionReassignmentRateAndTimeMs.FifteenMinuteRate": &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
			"PartitionReassignmentRateAndTimeMs.FiveMinuteRate":    &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
			"PartitionReassignmentRateAndTimeMs.Max":               &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
			"PartitionReassignmentRateAndTimeMs.Mean":              &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
			"PartitionReassignmentRateAndTimeMs.MeanRate":          &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
			"PartitionReassignmentRateAndTimeMs.Min":               &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
			"PartitionReassignmentRateAndTimeMs.OneMinuteRate":     &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
			"PartitionReassignmentRateAndTimeMs.StdDev":            &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
			"PartitionReassignmentRateAndTimeMs.Count":             &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount, Desc: ""},
			"PartitionReassignmentRateAndTimeMs.LatencyUnit":       &inputs.FieldInfo{DataType: inputs.String, Type: inputs.String, Unit: inputs.UnknownUnit, Desc: ""},
			"PartitionReassignmentRateAndTimeMs.EventType":         &inputs.FieldInfo{DataType: inputs.String, Type: inputs.String, Unit: inputs.UnknownUnit, Desc: ""},
			"PartitionReassignmentRateAndTimeMs.RateUnit":          &inputs.FieldInfo{DataType: inputs.String, Type: inputs.String, Unit: inputs.UnknownUnit, Desc: ""},

			"TopicChangeRateAndTimeMs.50thPercentile":    &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
			"TopicChangeRateAndTimeMs.75thPercentile":    &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
			"TopicChangeRateAndTimeMs.95thPercentile":    &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
			"TopicChangeRateAndTimeMs.98thPercentile":    &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
			"TopicChangeRateAndTimeMs.999thPercentile":   &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
			"TopicChangeRateAndTimeMs.99thPercentile":    &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
			"TopicChangeRateAndTimeMs.FifteenMinuteRate": &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
			"TopicChangeRateAndTimeMs.FiveMinuteRate":    &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
			"TopicChangeRateAndTimeMs.Max":               &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
			"TopicChangeRateAndTimeMs.Mean":              &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
			"TopicChangeRateAndTimeMs.MeanRate":          &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
			"TopicChangeRateAndTimeMs.Min":               &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
			"TopicChangeRateAndTimeMs.OneMinuteRate":     &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
			"TopicChangeRateAndTimeMs.StdDev":            &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
			"TopicChangeRateAndTimeMs.Count":             &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount, Desc: ""},
			"TopicChangeRateAndTimeMs.LatencyUnit":       &inputs.FieldInfo{DataType: inputs.String, Type: inputs.String, Unit: inputs.UnknownUnit, Desc: ""},
			"TopicChangeRateAndTimeMs.EventType":         &inputs.FieldInfo{DataType: inputs.String, Type: inputs.String, Unit: inputs.UnknownUnit, Desc: ""},
			"TopicChangeRateAndTimeMs.RateUnit":          &inputs.FieldInfo{DataType: inputs.String, Type: inputs.String, Unit: inputs.UnknownUnit, Desc: ""},

			"TopicDeletionRateAndTimeMs.50thPercentile":    &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
			"TopicDeletionRateAndTimeMs.75thPercentile":    &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
			"TopicDeletionRateAndTimeMs.95thPercentile":    &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
			"TopicDeletionRateAndTimeMs.98thPercentile":    &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
			"TopicDeletionRateAndTimeMs.999thPercentile":   &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
			"TopicDeletionRateAndTimeMs.99thPercentile":    &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
			"TopicDeletionRateAndTimeMs.FifteenMinuteRate": &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
			"TopicDeletionRateAndTimeMs.FiveMinuteRate":    &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
			"TopicDeletionRateAndTimeMs.Max":               &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
			"TopicDeletionRateAndTimeMs.Mean":              &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
			"TopicDeletionRateAndTimeMs.MeanRate":          &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
			"TopicDeletionRateAndTimeMs.Min":               &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
			"TopicDeletionRateAndTimeMs.OneMinuteRate":     &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
			"TopicDeletionRateAndTimeMs.StdDev":            &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
			"TopicDeletionRateAndTimeMs.Count":             &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount, Desc: ""},
			"TopicDeletionRateAndTimeMs.LatencyUnit":       &inputs.FieldInfo{DataType: inputs.String, Type: inputs.String, Unit: inputs.UnknownUnit, Desc: ""},
			"TopicDeletionRateAndTimeMs.EventType":         &inputs.FieldInfo{DataType: inputs.String, Type: inputs.String, Unit: inputs.UnknownUnit, Desc: ""},
			"TopicDeletionRateAndTimeMs.RateUnit":          &inputs.FieldInfo{DataType: inputs.String, Type: inputs.String, Unit: inputs.UnknownUnit, Desc: ""},

			"TopicUncleanLeaderElectionEnableRateAndTimeMs.50thPercentile":    &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
			"TopicUncleanLeaderElectionEnableRateAndTimeMs.75thPercentile":    &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
			"TopicUncleanLeaderElectionEnableRateAndTimeMs.95thPercentile":    &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
			"TopicUncleanLeaderElectionEnableRateAndTimeMs.98thPercentile":    &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
			"TopicUncleanLeaderElectionEnableRateAndTimeMs.999thPercentile":   &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
			"TopicUncleanLeaderElectionEnableRateAndTimeMs.99thPercentile":    &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
			"TopicUncleanLeaderElectionEnableRateAndTimeMs.FifteenMinuteRate": &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
			"TopicUncleanLeaderElectionEnableRateAndTimeMs.FiveMinuteRate":    &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
			"TopicUncleanLeaderElectionEnableRateAndTimeMs.Max":               &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
			"TopicUncleanLeaderElectionEnableRateAndTimeMs.Mean":              &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
			"TopicUncleanLeaderElectionEnableRateAndTimeMs.MeanRate":          &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
			"TopicUncleanLeaderElectionEnableRateAndTimeMs.Min":               &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
			"TopicUncleanLeaderElectionEnableRateAndTimeMs.OneMinuteRate":     &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
			"TopicUncleanLeaderElectionEnableRateAndTimeMs.StdDev":            &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
			"TopicUncleanLeaderElectionEnableRateAndTimeMs.Count":             &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount, Desc: ""},
			"TopicUncleanLeaderElectionEnableRateAndTimeMs.LatencyUnit":       &inputs.FieldInfo{DataType: inputs.String, Type: inputs.String, Unit: inputs.UnknownUnit, Desc: ""},
			"TopicUncleanLeaderElectionEnableRateAndTimeMs.EventType":         &inputs.FieldInfo{DataType: inputs.String, Type: inputs.String, Unit: inputs.UnknownUnit, Desc: ""},
			"TopicUncleanLeaderElectionEnableRateAndTimeMs.RateUnit":          &inputs.FieldInfo{DataType: inputs.String, Type: inputs.String, Unit: inputs.UnknownUnit, Desc: ""},

			"UncleanLeaderElectionEnableRateAndTimeMs.50thPercentile":    &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
			"UncleanLeaderElectionEnableRateAndTimeMs.75thPercentile":    &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
			"UncleanLeaderElectionEnableRateAndTimeMs.95thPercentile":    &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
			"UncleanLeaderElectionEnableRateAndTimeMs.98thPercentile":    &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
			"UncleanLeaderElectionEnableRateAndTimeMs.999thPercentile":   &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
			"UncleanLeaderElectionEnableRateAndTimeMs.99thPercentile":    &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
			"UncleanLeaderElectionEnableRateAndTimeMs.FifteenMinuteRate": &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
			"UncleanLeaderElectionEnableRateAndTimeMs.FiveMinuteRate":    &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
			"UncleanLeaderElectionEnableRateAndTimeMs.Max":               &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
			"UncleanLeaderElectionEnableRateAndTimeMs.Mean":              &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
			"UncleanLeaderElectionEnableRateAndTimeMs.MeanRate":          &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
			"UncleanLeaderElectionEnableRateAndTimeMs.Min":               &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
			"UncleanLeaderElectionEnableRateAndTimeMs.OneMinuteRate":     &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
			"UncleanLeaderElectionEnableRateAndTimeMs.StdDev":            &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
			"UncleanLeaderElectionEnableRateAndTimeMs.Count":             &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount, Desc: ""},
			"UncleanLeaderElectionEnableRateAndTimeMs.LatencyUnit":       &inputs.FieldInfo{DataType: inputs.String, Type: inputs.String, Unit: inputs.UnknownUnit, Desc: ""},
			"UncleanLeaderElectionEnableRateAndTimeMs.EventType":         &inputs.FieldInfo{DataType: inputs.String, Type: inputs.String, Unit: inputs.UnknownUnit, Desc: ""},
			"UncleanLeaderElectionEnableRateAndTimeMs.RateUnit":          &inputs.FieldInfo{DataType: inputs.String, Type: inputs.String, Unit: inputs.UnknownUnit, Desc: ""},

			"UpdateFeaturesRateAndTimeMs.50thPercentile":    &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
			"UpdateFeaturesRateAndTimeMs.75thPercentile":    &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
			"UpdateFeaturesRateAndTimeMs.95thPercentile":    &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
			"UpdateFeaturesRateAndTimeMs.98thPercentile":    &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
			"UpdateFeaturesRateAndTimeMs.999thPercentile":   &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
			"UpdateFeaturesRateAndTimeMs.99thPercentile":    &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
			"UpdateFeaturesRateAndTimeMs.FifteenMinuteRate": &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
			"UpdateFeaturesRateAndTimeMs.FiveMinuteRate":    &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
			"UpdateFeaturesRateAndTimeMs.Max":               &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
			"UpdateFeaturesRateAndTimeMs.Mean":              &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
			"UpdateFeaturesRateAndTimeMs.MeanRate":          &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
			"UpdateFeaturesRateAndTimeMs.Min":               &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
			"UpdateFeaturesRateAndTimeMs.OneMinuteRate":     &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
			"UpdateFeaturesRateAndTimeMs.StdDev":            &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
			"UpdateFeaturesRateAndTimeMs.Count":             &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount, Desc: ""},
			"UpdateFeaturesRateAndTimeMs.LatencyUnit":       &inputs.FieldInfo{DataType: inputs.String, Type: inputs.String, Unit: inputs.UnknownUnit, Desc: ""},
			"UpdateFeaturesRateAndTimeMs.EventType":         &inputs.FieldInfo{DataType: inputs.String, Type: inputs.String, Unit: inputs.UnknownUnit, Desc: ""},
			"UpdateFeaturesRateAndTimeMs.RateUnit":          &inputs.FieldInfo{DataType: inputs.String, Type: inputs.String, Unit: inputs.UnknownUnit, Desc: ""},

			"UncleanLeaderElectionsPerSec.50thPercentile":    &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
			"UncleanLeaderElectionsPerSec.75thPercentile":    &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
			"UncleanLeaderElectionsPerSec.95thPercentile":    &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
			"UncleanLeaderElectionsPerSec.98thPercentile":    &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
			"UncleanLeaderElectionsPerSec.999thPercentile":   &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
			"UncleanLeaderElectionsPerSec.99thPercentile":    &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
			"UncleanLeaderElectionsPerSec.FifteenMinuteRate": &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
			"UncleanLeaderElectionsPerSec.FiveMinuteRate":    &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
			"UncleanLeaderElectionsPerSec.Max":               &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
			"UncleanLeaderElectionsPerSec.Mean":              &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
			"UncleanLeaderElectionsPerSec.MeanRate":          &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
			"UncleanLeaderElectionsPerSec.Min":               &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
			"UncleanLeaderElectionsPerSec.OneMinuteRate":     &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
			"UncleanLeaderElectionsPerSec.StdDev":            &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
			"UncleanLeaderElectionsPerSec.Count":             &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount, Desc: ""},
			"UncleanLeaderElectionsPerSec.LatencyUnit":       &inputs.FieldInfo{DataType: inputs.String, Type: inputs.String, Unit: inputs.UnknownUnit, Desc: ""},
			"UncleanLeaderElectionsPerSec.EventType":         &inputs.FieldInfo{DataType: inputs.String, Type: inputs.String, Unit: inputs.UnknownUnit, Desc: ""},
			"UncleanLeaderElectionsPerSec.RateUnit":          &inputs.FieldInfo{DataType: inputs.String, Type: inputs.String, Unit: inputs.UnknownUnit, Desc: ""},

			"EventQueueTimeMs.50thPercentile":    &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
			"EventQueueTimeMs.75thPercentile":    &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
			"EventQueueTimeMs.95thPercentile":    &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
			"EventQueueTimeMs.98thPercentile":    &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
			"EventQueueTimeMs.999thPercentile":   &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
			"EventQueueTimeMs.99thPercentile":    &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
			"EventQueueTimeMs.FifteenMinuteRate": &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
			"EventQueueTimeMs.FiveMinuteRate":    &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
			"EventQueueTimeMs.Max":               &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
			"EventQueueTimeMs.Mean":              &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
			"EventQueueTimeMs.MeanRate":          &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
			"EventQueueTimeMs.Min":               &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
			"EventQueueTimeMs.OneMinuteRate":     &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
			"EventQueueTimeMs.StdDev":            &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
			"EventQueueTimeMs.Count":             &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount, Desc: ""},
			"EventQueueTimeMs.LatencyUnit":       &inputs.FieldInfo{DataType: inputs.String, Type: inputs.String, Unit: inputs.UnknownUnit, Desc: ""},
			"EventQueueTimeMs.EventType":         &inputs.FieldInfo{DataType: inputs.String, Type: inputs.String, Unit: inputs.UnknownUnit, Desc: ""},
			"EventQueueTimeMs.RateUnit":          &inputs.FieldInfo{DataType: inputs.String, Type: inputs.String, Unit: inputs.UnknownUnit, Desc: ""},

			"GlobalPartitionCount.Value":            &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount, Desc: ""},
			"GlobalTopicCount.Value":                &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount, Desc: ""},
			"OfflinePartitionsCount.Value":          &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount, Desc: ""},
			"PreferredReplicaImbalanceCount.Value":  &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount, Desc: ""},
			"ReplicasIneligibleToDeleteCount.Value": &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount, Desc: ""},
			"ReplicasToDeleteCount.Value":           &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount, Desc: ""},
			"TopicsIneligibleToDeleteCount.Value":   &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount, Desc: ""},
			"TopicsToDeleteCount.Value":             &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount, Desc: ""},
			"ActiveControllerCount.Value":           &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount, Desc: ""},

			"ControllerState.Value": &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
			"EventQueueSize.Value":  &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
			"TotalQueueSize.Value":  &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
		},
		Tags: map[string]interface{}{
			"jolokia_agent_url": inputs.TagInfo{Desc: "jolokia agent url path"},
		},
	}
}

func (j *KafkaReplicaMment) LineProto() (*io.Point, error) {
	return io.MakePoint(j.name, j.tags, j.fields, j.ts)
}

//nolint:lll
func (j *KafkaReplicaMment) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name: "kafka_replica_manager",
		Fields: map[string]interface{}{
			"FailedIsrUpdatesPerSec.Count":             &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount, Desc: ""},
			"FailedIsrUpdatesPerSec.EventType":         &inputs.FieldInfo{DataType: inputs.String, Type: inputs.String, Unit: inputs.UnknownUnit, Desc: ""},
			"FailedIsrUpdatesPerSec.FifteenMinuteRate": &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
			"FailedIsrUpdatesPerSec.FiveMinuteRate":    &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
			"FailedIsrUpdatesPerSec.MeanRate":          &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
			"FailedIsrUpdatesPerSec.OneMinuteRate":     &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
			"FailedIsrUpdatesPerSec.RateUnit":          &inputs.FieldInfo{DataType: inputs.String, Type: inputs.String, Unit: inputs.UnknownUnit, Desc: ""},

			"IsrExpandsPerSec.Count":             &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount, Desc: ""},
			"IsrExpandsPerSec.EventType":         &inputs.FieldInfo{DataType: inputs.String, Type: inputs.String, Unit: inputs.UnknownUnit, Desc: ""},
			"IsrExpandsPerSec.FifteenMinuteRate": &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
			"IsrExpandsPerSec.FiveMinuteRate":    &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
			"IsrExpandsPerSec.MeanRate":          &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
			"IsrExpandsPerSec.OneMinuteRate":     &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
			"IsrExpandsPerSec.RateUnit":          &inputs.FieldInfo{DataType: inputs.String, Type: inputs.String, Unit: inputs.UnknownUnit, Desc: ""},

			"IsrShrinksPerSec.Count":             &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount, Desc: ""},
			"IsrShrinksPerSec.EventType":         &inputs.FieldInfo{DataType: inputs.String, Type: inputs.String, Unit: inputs.UnknownUnit, Desc: ""},
			"IsrShrinksPerSec.FifteenMinuteRate": &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
			"IsrShrinksPerSec.FiveMinuteRate":    &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
			"IsrShrinksPerSec.MeanRate":          &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
			"IsrShrinksPerSec.OneMinuteRate":     &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
			"IsrShrinksPerSec.RateUnit":          &inputs.FieldInfo{DataType: inputs.String, Type: inputs.String, Unit: inputs.UnknownUnit, Desc: ""},

			"AtMinIsrPartitionCount.Value":    &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount, Desc: ""},
			"LeaderCount.Value":               &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount, Desc: ""},
			"OfflineReplicaCount.Value":       &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount, Desc: ""},
			"PartitionCount.Value":            &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount, Desc: ""},
			"ReassigningPartitions.Value":     &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount, Desc: ""},
			"UnderMinIsrPartitionCount.Value": &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount, Desc: ""},
			"UnderReplicatedPartitions.Value": &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount, Desc: ""},
		},
		Tags: map[string]interface{}{
			"jolokia_agent_url": inputs.TagInfo{Desc: "jolokia agent url path"},
		},
	}
}

func (j *KafkaPurgatoryMment) LineProto() (*io.Point, error) {
	return io.MakePoint(j.name, j.tags, j.fields, j.ts)
}

//nolint:lll
func (j *KafkaPurgatoryMment) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name: "kafka_purgatory",
		Fields: map[string]interface{}{
			"AlterAcls.NumDelayedOperations": &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
			"AlterAcls.PurgatorySize":        &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},

			"DeleteRecords.NumDelayedOperations": &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
			"DeleteRecords.PurgatorySize":        &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},

			"ElectLeader.NumDelayedOperations": &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
			"ElectLeader.PurgatorySize":        &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},

			"Fetch.NumDelayedOperations": &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
			"Fetch.PurgatorySize":        &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},

			"Heartbeat.NumDelayedOperations": &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
			"Heartbeat.PurgatorySize":        &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},

			"Produce.NumDelayedOperations": &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
			"Produce.PurgatorySize":        &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},

			"Rebalance.NumDelayedOperations": &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
			"Rebalance.PurgatorySize":        &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},

			"topic.NumDelayedOperations": &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
			"topic.PurgatorySize":        &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
		},

		Tags: map[string]interface{}{
			"jolokia_agent_url": inputs.TagInfo{Desc: "jolokia agent url path"},
		},
	}
}

func (j *KafkaClientMment) LineProto() (*io.Point, error) {
	return io.MakePoint(j.name, j.tags, j.fields, j.ts)
}

//nolint:lll
func (j *KafkaClientMment) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name:   "kafka_client",
		Fields: nil,
		Tags:   nil,
	}
}

func (j *KafkaRequestMment) LineProto() (*io.Point, error) {
	return io.MakePoint(j.name, j.tags, j.fields, j.ts)
}

//nolint:lll
func (j *KafkaRequestMment) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name: "kafka_request",
		Fields: map[string]interface{}{
			"LocalTimeMs.50thPercentile":  &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS, Desc: ""},
			"LocalTimeMs.75thPercentile":  &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS, Desc: ""},
			"LocalTimeMs.95thPercentile":  &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS, Desc: ""},
			"LocalTimeMs.98thPercentile":  &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS, Desc: ""},
			"LocalTimeMs.999thPercentile": &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS, Desc: ""},
			"LocalTimeMs.99thPercentile":  &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS, Desc: ""},
			"LocalTimeMs.Max":             &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS, Desc: ""},
			"LocalTimeMs.Mean":            &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS, Desc: ""},
			"LocalTimeMs.Min":             &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS, Desc: ""},
			"LocalTimeMs.StdDev":          &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS, Desc: ""},
			"LocalTimeMs.Count":           &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount, Desc: ""},

			"RemoteTimeMs.50thPercentile":  &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS, Desc: ""},
			"RemoteTimeMs.75thPercentile":  &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS, Desc: ""},
			"RemoteTimeMs.95thPercentile":  &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS, Desc: ""},
			"RemoteTimeMs.98thPercentile":  &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS, Desc: ""},
			"RemoteTimeMs.999thPercentile": &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS, Desc: ""},
			"RemoteTimeMs.99thPercentile":  &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS, Desc: ""},
			"RemoteTimeMs.Max":             &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS, Desc: ""},
			"RemoteTimeMs.Mean":            &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS, Desc: ""},
			"RemoteTimeMs.Min":             &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS, Desc: ""},
			"RemoteTimeMs.StdDev":          &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS, Desc: ""},
			"RemoteTimeMs.Count":           &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount, Desc: ""},

			"RequestBytes.50thPercentile":  &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS, Desc: ""},
			"RequestBytes.75thPercentile":  &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS, Desc: ""},
			"RequestBytes.95thPercentile":  &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS, Desc: ""},
			"RequestBytes.98thPercentile":  &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS, Desc: ""},
			"RequestBytes.999thPercentile": &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS, Desc: ""},
			"RequestBytes.99thPercentile":  &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS, Desc: ""},
			"RequestBytes.Max":             &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS, Desc: ""},
			"RequestBytes.Mean":            &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS, Desc: ""},
			"RequestBytes.Min":             &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS, Desc: ""},
			"RequestBytes.StdDev":          &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS, Desc: ""},
			"RequestBytes.Count":           &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount, Desc: ""},

			"RequestQueueTimeMs.50thPercentile":  &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS, Desc: ""},
			"RequestQueueTimeMs.75thPercentile":  &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS, Desc: ""},
			"RequestQueueTimeMs.95thPercentile":  &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS, Desc: ""},
			"RequestQueueTimeMs.98thPercentile":  &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS, Desc: ""},
			"RequestQueueTimeMs.999thPercentile": &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS, Desc: ""},
			"RequestQueueTimeMs.99thPercentile":  &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS, Desc: ""},
			"RequestQueueTimeMs.Max":             &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS, Desc: ""},
			"RequestQueueTimeMs.Mean":            &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS, Desc: ""},
			"RequestQueueTimeMs.Min":             &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS, Desc: ""},
			"RequestQueueTimeMs.StdDev":          &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS, Desc: ""},
			"RequestQueueTimeMs.Count":           &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount, Desc: ""},

			"ResponseQueueTimeMs.50thPercentile":  &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS, Desc: ""},
			"ResponseQueueTimeMs.75thPercentile":  &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS, Desc: ""},
			"ResponseQueueTimeMs.95thPercentile":  &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS, Desc: ""},
			"ResponseQueueTimeMs.98thPercentile":  &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS, Desc: ""},
			"ResponseQueueTimeMs.999thPercentile": &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS, Desc: ""},
			"ResponseQueueTimeMs.99thPercentile":  &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS, Desc: ""},
			"ResponseQueueTimeMs.Max":             &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS, Desc: ""},
			"ResponseQueueTimeMs.Mean":            &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS, Desc: ""},
			"ResponseQueueTimeMs.Min":             &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS, Desc: ""},
			"ResponseQueueTimeMs.StdDev":          &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS, Desc: ""},
			"ResponseQueueTimeMs.Count":           &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount, Desc: ""},

			"ResponseSendTimeMs.50thPercentile":  &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS, Desc: ""},
			"ResponseSendTimeMs.75thPercentile":  &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS, Desc: ""},
			"ResponseSendTimeMs.95thPercentile":  &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS, Desc: ""},
			"ResponseSendTimeMs.98thPercentile":  &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS, Desc: ""},
			"ResponseSendTimeMs.999thPercentile": &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS, Desc: ""},
			"ResponseSendTimeMs.99thPercentile":  &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS, Desc: ""},
			"ResponseSendTimeMs.Max":             &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS, Desc: ""},
			"ResponseSendTimeMs.Mean":            &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS, Desc: ""},
			"ResponseSendTimeMs.Min":             &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS, Desc: ""},
			"ResponseSendTimeMs.StdDev":          &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS, Desc: ""},
			"ResponseSendTimeMs.Count":           &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount, Desc: ""},

			"ThrottleTimeMs.50thPercentile":  &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS, Desc: ""},
			"ThrottleTimeMs.75thPercentile":  &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS, Desc: ""},
			"ThrottleTimeMs.95thPercentile":  &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS, Desc: ""},
			"ThrottleTimeMs.98thPercentile":  &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS, Desc: ""},
			"ThrottleTimeMs.999thPercentile": &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS, Desc: ""},
			"ThrottleTimeMs.99thPercentile":  &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS, Desc: ""},
			"ThrottleTimeMs.Max":             &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS, Desc: ""},
			"ThrottleTimeMs.Mean":            &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS, Desc: ""},
			"ThrottleTimeMs.Min":             &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS, Desc: ""},
			"ThrottleTimeMs.StdDev":          &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS, Desc: ""},
			"ThrottleTimeMs.Count":           &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount, Desc: ""},

			"TotalTimeMs.50thPercentile":  &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS, Desc: ""},
			"TotalTimeMs.75thPercentile":  &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS, Desc: ""},
			"TotalTimeMs.95thPercentile":  &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS, Desc: ""},
			"TotalTimeMs.98thPercentile":  &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS, Desc: ""},
			"TotalTimeMs.999thPercentile": &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS, Desc: ""},
			"TotalTimeMs.99thPercentile":  &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS, Desc: ""},
			"TotalTimeMs.Max":             &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS, Desc: ""},
			"TotalTimeMs.Mean":            &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS, Desc: ""},
			"TotalTimeMs.Min":             &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS, Desc: ""},
			"TotalTimeMs.StdDev":          &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS, Desc: ""},
			"TotalTimeMs.Count":           &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount, Desc: ""},
		},
		Tags: map[string]interface{}{
			"jolokia_agent_url": inputs.TagInfo{Desc: "jolokia agent url path"},
		},
	}
}

func (j *KafkaTopicsMment) LineProto() (*io.Point, error) {
	return io.MakePoint(j.name, j.tags, j.fields, j.ts)
}

//nolint:lll
func (j *KafkaTopicsMment) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name: "kafka_topics",
		Fields: map[string]interface{}{
			"BytesInPerSec.Count":             &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount, Desc: ""},
			"BytesInPerSec.EventType":         &inputs.FieldInfo{DataType: inputs.String, Type: inputs.String, Unit: inputs.UnknownUnit, Desc: ""},
			"BytesInPerSec.FifteenMinuteRate": &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
			"BytesInPerSec.FiveMinuteRate":    &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
			"BytesInPerSec.MeanRate":          &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
			"BytesInPerSec.OneMinuteRate":     &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
			"BytesInPerSec.RateUnit":          &inputs.FieldInfo{DataType: inputs.String, Type: inputs.String, Unit: inputs.UnknownUnit, Desc: ""},

			"BytesOutPerSec.Count":             &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount, Desc: ""},
			"BytesOutPerSec.EventType":         &inputs.FieldInfo{DataType: inputs.String, Type: inputs.String, Unit: inputs.UnknownUnit, Desc: ""},
			"BytesOutPerSec.FifteenMinuteRate": &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
			"BytesOutPerSec.FiveMinuteRate":    &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
			"BytesOutPerSec.MeanRate":          &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
			"BytesOutPerSec.OneMinuteRate":     &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
			"BytesOutPerSec.RateUnit":          &inputs.FieldInfo{DataType: inputs.String, Type: inputs.String, Unit: inputs.UnknownUnit, Desc: ""},

			"BytesRejectedPerSec.Count":             &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount, Desc: ""},
			"BytesRejectedPerSec.EventType":         &inputs.FieldInfo{DataType: inputs.String, Type: inputs.String, Unit: inputs.UnknownUnit, Desc: ""},
			"BytesRejectedPerSec.FifteenMinuteRate": &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
			"BytesRejectedPerSec.FiveMinuteRate":    &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
			"BytesRejectedPerSec.MeanRate":          &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
			"BytesRejectedPerSec.OneMinuteRate":     &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
			"BytesRejectedPerSec.RateUnit":          &inputs.FieldInfo{DataType: inputs.String, Type: inputs.String, Unit: inputs.UnknownUnit, Desc: ""},

			"FailedFetchRequestsPerSec.Count":             &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount, Desc: ""},
			"FailedFetchRequestsPerSec.EventType":         &inputs.FieldInfo{DataType: inputs.String, Type: inputs.String, Unit: inputs.UnknownUnit, Desc: ""},
			"FailedFetchRequestsPerSec.FifteenMinuteRate": &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
			"FailedFetchRequestsPerSec.FiveMinuteRate":    &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
			"FailedFetchRequestsPerSec.MeanRate":          &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
			"FailedFetchRequestsPerSec.OneMinuteRate":     &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
			"FailedFetchRequestsPerSec.RateUnit":          &inputs.FieldInfo{DataType: inputs.String, Type: inputs.String, Unit: inputs.UnknownUnit, Desc: ""},

			"FailedProduceRequestsPerSec.Count":             &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount, Desc: ""},
			"FailedProduceRequestsPerSec.EventType":         &inputs.FieldInfo{DataType: inputs.String, Type: inputs.String, Unit: inputs.UnknownUnit, Desc: ""},
			"FailedProduceRequestsPerSec.FifteenMinuteRate": &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
			"FailedProduceRequestsPerSec.FiveMinuteRate":    &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
			"FailedProduceRequestsPerSec.MeanRate":          &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
			"FailedProduceRequestsPerSec.OneMinuteRate":     &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
			"FailedProduceRequestsPerSec.RateUnit":          &inputs.FieldInfo{DataType: inputs.String, Type: inputs.String, Unit: inputs.UnknownUnit, Desc: ""},

			"FetchMessageConversionsPerSec.Count":             &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount, Desc: ""},
			"FetchMessageConversionsPerSec.EventType":         &inputs.FieldInfo{DataType: inputs.String, Type: inputs.String, Unit: inputs.UnknownUnit, Desc: ""},
			"FetchMessageConversionsPerSec.FifteenMinuteRate": &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
			"FetchMessageConversionsPerSec.FiveMinuteRate":    &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
			"FetchMessageConversionsPerSec.MeanRate":          &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
			"FetchMessageConversionsPerSec.OneMinuteRate":     &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
			"FetchMessageConversionsPerSec.RateUnit":          &inputs.FieldInfo{DataType: inputs.String, Type: inputs.String, Unit: inputs.UnknownUnit, Desc: ""},

			"InvalidMagicNumberRecordsPerSec.Count":             &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount, Desc: ""},
			"InvalidMagicNumberRecordsPerSec.EventType":         &inputs.FieldInfo{DataType: inputs.String, Type: inputs.String, Unit: inputs.UnknownUnit, Desc: ""},
			"InvalidMagicNumberRecordsPerSec.FifteenMinuteRate": &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
			"InvalidMagicNumberRecordsPerSec.FiveMinuteRate":    &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
			"InvalidMagicNumberRecordsPerSec.MeanRate":          &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
			"InvalidMagicNumberRecordsPerSec.OneMinuteRate":     &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
			"InvalidMagicNumberRecordsPerSec.RateUnit":          &inputs.FieldInfo{DataType: inputs.String, Type: inputs.String, Unit: inputs.UnknownUnit, Desc: ""},

			"InvalidMessageCrcRecordsPerSec.Count":             &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount, Desc: ""},
			"InvalidMessageCrcRecordsPerSec.EventType":         &inputs.FieldInfo{DataType: inputs.String, Type: inputs.String, Unit: inputs.UnknownUnit, Desc: ""},
			"InvalidMessageCrcRecordsPerSec.FifteenMinuteRate": &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
			"InvalidMessageCrcRecordsPerSec.FiveMinuteRate":    &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
			"InvalidMessageCrcRecordsPerSec.MeanRate":          &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
			"InvalidMessageCrcRecordsPerSec.OneMinuteRate":     &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
			"InvalidMessageCrcRecordsPerSec.RateUnit":          &inputs.FieldInfo{DataType: inputs.String, Type: inputs.String, Unit: inputs.UnknownUnit, Desc: ""},

			"InvalidOffsetOrSequenceRecordsPerSec.Count":             &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount, Desc: ""},
			"InvalidOffsetOrSequenceRecordsPerSec.EventType":         &inputs.FieldInfo{DataType: inputs.String, Type: inputs.String, Unit: inputs.UnknownUnit, Desc: ""},
			"InvalidOffsetOrSequenceRecordsPerSec.FifteenMinuteRate": &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
			"InvalidOffsetOrSequenceRecordsPerSec.FiveMinuteRate":    &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
			"InvalidOffsetOrSequenceRecordsPerSec.MeanRate":          &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
			"InvalidOffsetOrSequenceRecordsPerSec.OneMinuteRate":     &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
			"InvalidOffsetOrSequenceRecordsPerSec.RateUnit":          &inputs.FieldInfo{DataType: inputs.String, Type: inputs.String, Unit: inputs.UnknownUnit, Desc: ""},

			"MessagesInPerSec.Count":             &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount, Desc: ""},
			"MessagesInPerSec.EventType":         &inputs.FieldInfo{DataType: inputs.String, Type: inputs.String, Unit: inputs.UnknownUnit, Desc: ""},
			"MessagesInPerSec.FifteenMinuteRate": &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
			"MessagesInPerSec.FiveMinuteRate":    &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
			"MessagesInPerSec.MeanRate":          &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
			"MessagesInPerSec.OneMinuteRate":     &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
			"MessagesInPerSec.RateUnit":          &inputs.FieldInfo{DataType: inputs.String, Type: inputs.String, Unit: inputs.UnknownUnit, Desc: ""},

			"NoKeyCompactedTopicRecordsPerSec.Count":             &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount, Desc: ""},
			"NoKeyCompactedTopicRecordsPerSec.EventType":         &inputs.FieldInfo{DataType: inputs.String, Type: inputs.String, Unit: inputs.UnknownUnit, Desc: ""},
			"NoKeyCompactedTopicRecordsPerSec.FifteenMinuteRate": &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
			"NoKeyCompactedTopicRecordsPerSec.FiveMinuteRate":    &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
			"NoKeyCompactedTopicRecordsPerSec.MeanRate":          &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
			"NoKeyCompactedTopicRecordsPerSec.OneMinuteRate":     &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
			"NoKeyCompactedTopicRecordsPerSec.RateUnit":          &inputs.FieldInfo{DataType: inputs.String, Type: inputs.String, Unit: inputs.UnknownUnit, Desc: ""},

			"ProduceMessageConversionsPerSec.Count":             &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount, Desc: ""},
			"ProduceMessageConversionsPerSec.EventType":         &inputs.FieldInfo{DataType: inputs.String, Type: inputs.String, Unit: inputs.UnknownUnit, Desc: ""},
			"ProduceMessageConversionsPerSec.FifteenMinuteRate": &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
			"ProduceMessageConversionsPerSec.FiveMinuteRate":    &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
			"ProduceMessageConversionsPerSec.MeanRate":          &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
			"ProduceMessageConversionsPerSec.OneMinuteRate":     &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
			"ProduceMessageConversionsPerSec.RateUnit":          &inputs.FieldInfo{DataType: inputs.String, Type: inputs.String, Unit: inputs.UnknownUnit, Desc: ""},

			"ReassignmentBytesInPerSec.Count":             &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount, Desc: ""},
			"ReassignmentBytesInPerSec.EventType":         &inputs.FieldInfo{DataType: inputs.String, Type: inputs.String, Unit: inputs.UnknownUnit, Desc: ""},
			"ReassignmentBytesInPerSec.FifteenMinuteRate": &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
			"ReassignmentBytesInPerSec.FiveMinuteRate":    &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
			"ReassignmentBytesInPerSec.MeanRate":          &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
			"ReassignmentBytesInPerSec.OneMinuteRate":     &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
			"ReassignmentBytesInPerSec.RateUnit":          &inputs.FieldInfo{DataType: inputs.String, Type: inputs.String, Unit: inputs.UnknownUnit, Desc: ""},

			"ReassignmentBytesOutPerSec.Count":             &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount, Desc: ""},
			"ReassignmentBytesOutPerSec.EventType":         &inputs.FieldInfo{DataType: inputs.String, Type: inputs.String, Unit: inputs.UnknownUnit, Desc: ""},
			"ReassignmentBytesOutPerSec.FifteenMinuteRate": &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
			"ReassignmentBytesOutPerSec.FiveMinuteRate":    &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
			"ReassignmentBytesOutPerSec.MeanRate":          &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
			"ReassignmentBytesOutPerSec.OneMinuteRate":     &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
			"ReassignmentBytesOutPerSec.RateUnit":          &inputs.FieldInfo{DataType: inputs.String, Type: inputs.String, Unit: inputs.UnknownUnit, Desc: ""},

			"ReplicationBytesInPerSec.Count":             &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount, Desc: ""},
			"ReplicationBytesInPerSec.EventType":         &inputs.FieldInfo{DataType: inputs.String, Type: inputs.String, Unit: inputs.UnknownUnit, Desc: ""},
			"ReplicationBytesInPerSec.FifteenMinuteRate": &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
			"ReplicationBytesInPerSec.FiveMinuteRate":    &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
			"ReplicationBytesInPerSec.MeanRate":          &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
			"ReplicationBytesInPerSec.OneMinuteRate":     &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
			"ReplicationBytesInPerSec.RateUnit":          &inputs.FieldInfo{DataType: inputs.String, Type: inputs.String, Unit: inputs.UnknownUnit, Desc: ""},

			"ReplicationBytesOutPerSec.Count":             &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount, Desc: ""},
			"ReplicationBytesOutPerSec.EventType":         &inputs.FieldInfo{DataType: inputs.String, Type: inputs.String, Unit: inputs.UnknownUnit, Desc: ""},
			"ReplicationBytesOutPerSec.FifteenMinuteRate": &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
			"ReplicationBytesOutPerSec.FiveMinuteRate":    &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
			"ReplicationBytesOutPerSec.MeanRate":          &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
			"ReplicationBytesOutPerSec.OneMinuteRate":     &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
			"ReplicationBytesOutPerSec.RateUnit":          &inputs.FieldInfo{DataType: inputs.String, Type: inputs.String, Unit: inputs.UnknownUnit, Desc: ""},

			"TotalFetchRequestsPerSec.Count":             &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount, Desc: ""},
			"TotalFetchRequestsPerSec.EventType":         &inputs.FieldInfo{DataType: inputs.String, Type: inputs.String, Unit: inputs.UnknownUnit, Desc: ""},
			"TotalFetchRequestsPerSec.FifteenMinuteRate": &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
			"TotalFetchRequestsPerSec.FiveMinuteRate":    &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
			"TotalFetchRequestsPerSec.MeanRate":          &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
			"TotalFetchRequestsPerSec.OneMinuteRate":     &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
			"TotalFetchRequestsPerSec.RateUnit":          &inputs.FieldInfo{DataType: inputs.String, Type: inputs.String, Unit: inputs.UnknownUnit, Desc: ""},

			"TotalProduceRequestsPerSec.Count":             &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount, Desc: ""},
			"TotalProduceRequestsPerSec.EventType":         &inputs.FieldInfo{DataType: inputs.String, Type: inputs.String, Unit: inputs.UnknownUnit, Desc: ""},
			"TotalProduceRequestsPerSec.FifteenMinuteRate": &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
			"TotalProduceRequestsPerSec.FiveMinuteRate":    &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
			"TotalProduceRequestsPerSec.MeanRate":          &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
			"TotalProduceRequestsPerSec.OneMinuteRate":     &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
			"TotalProduceRequestsPerSec.RateUnit":          &inputs.FieldInfo{DataType: inputs.String, Type: inputs.String, Unit: inputs.UnknownUnit, Desc: ""},
		},

		Tags: map[string]interface{}{
			"jolokia_agent_url": inputs.TagInfo{Desc: "jolokia agent url path"},
		},
	}
}

func (j *KafkaTopicMment) LineProto() (*io.Point, error) {
	return io.MakePoint(j.name, j.tags, j.fields, j.ts)
}

//nolint:lll
func (j *KafkaTopicMment) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name: "kafka_topic",
		Fields: map[string]interface{}{
			"BytesInPerSec.Count":             &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount, Desc: ""},
			"BytesInPerSec.EventType":         &inputs.FieldInfo{DataType: inputs.String, Type: inputs.String, Unit: inputs.UnknownUnit, Desc: ""},
			"BytesInPerSec.FifteenMinuteRate": &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
			"BytesInPerSec.FiveMinuteRate":    &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
			"BytesInPerSec.MeanRate":          &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
			"BytesInPerSec.OneMinuteRate":     &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
			"BytesInPerSec.RateUnit":          &inputs.FieldInfo{DataType: inputs.String, Type: inputs.String, Unit: inputs.UnknownUnit, Desc: ""},

			"BytesOutPerSec.Count":             &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount, Desc: ""},
			"BytesOutPerSec.EventType":         &inputs.FieldInfo{DataType: inputs.String, Type: inputs.String, Unit: inputs.UnknownUnit, Desc: ""},
			"BytesOutPerSec.FifteenMinuteRate": &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
			"BytesOutPerSec.FiveMinuteRate":    &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
			"BytesOutPerSec.MeanRate":          &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
			"BytesOutPerSec.OneMinuteRate":     &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
			"BytesOutPerSec.RateUnit":          &inputs.FieldInfo{DataType: inputs.String, Type: inputs.String, Unit: inputs.UnknownUnit, Desc: ""},

			"MessagesInPerSec.Count":             &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount, Desc: ""},
			"MessagesInPerSec.EventType":         &inputs.FieldInfo{DataType: inputs.String, Type: inputs.String, Unit: inputs.UnknownUnit, Desc: ""},
			"MessagesInPerSec.FifteenMinuteRate": &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
			"MessagesInPerSec.FiveMinuteRate":    &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
			"MessagesInPerSec.MeanRate":          &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
			"MessagesInPerSec.OneMinuteRate":     &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
			"MessagesInPerSec.RateUnit":          &inputs.FieldInfo{DataType: inputs.String, Type: inputs.String, Unit: inputs.UnknownUnit, Desc: ""},

			"TotalFetchRequestsPerSec.Count":             &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount, Desc: ""},
			"TotalFetchRequestsPerSec.EventType":         &inputs.FieldInfo{DataType: inputs.String, Type: inputs.String, Unit: inputs.UnknownUnit, Desc: ""},
			"TotalFetchRequestsPerSec.FifteenMinuteRate": &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
			"TotalFetchRequestsPerSec.FiveMinuteRate":    &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
			"TotalFetchRequestsPerSec.MeanRate":          &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
			"TotalFetchRequestsPerSec.OneMinuteRate":     &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
			"TotalFetchRequestsPerSec.RateUnit":          &inputs.FieldInfo{DataType: inputs.String, Type: inputs.String, Unit: inputs.UnknownUnit, Desc: ""},

			"TotalProduceRequestsPerSec.Count":             &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount, Desc: ""},
			"TotalProduceRequestsPerSec.EventType":         &inputs.FieldInfo{DataType: inputs.String, Type: inputs.String, Unit: inputs.UnknownUnit, Desc: ""},
			"TotalProduceRequestsPerSec.FifteenMinuteRate": &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
			"TotalProduceRequestsPerSec.FiveMinuteRate":    &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
			"TotalProduceRequestsPerSec.MeanRate":          &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
			"TotalProduceRequestsPerSec.OneMinuteRate":     &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
			"TotalProduceRequestsPerSec.RateUnit":          &inputs.FieldInfo{DataType: inputs.String, Type: inputs.String, Unit: inputs.UnknownUnit, Desc: ""},
		},

		Tags: map[string]interface{}{
			"jolokia_agent_url": inputs.TagInfo{Desc: "jolokia agent url path"},
			"topic":             inputs.TagInfo{Desc: "topic name"},
		},
	}
}

func (j *KafkaPartitionMment) LineProto() (*io.Point, error) {
	return io.MakePoint(j.name, j.tags, j.fields, j.ts)
}

//nolint:lll
func (j *KafkaPartitionMment) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name: "kafka_partition",
		Fields: map[string]interface{}{
			"LogEndOffset":              &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
			"LogStartOffset":            &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
			"NumLogSegments":            &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
			"Size":                      &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
			"UnderReplicatedPartitions": &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
		},

		Tags: map[string]interface{}{
			"jolokia_agent_url": inputs.TagInfo{Desc: "jolokia agent url path"},
			"partition":         inputs.TagInfo{Desc: "partition number"},
			"topic":             inputs.TagInfo{Desc: "topic name"},
		},
	}
}
