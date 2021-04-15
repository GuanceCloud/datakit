package kafka

import (
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
	"time"
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
	data, err := io.MakePoint(j.name, j.tags, j.fields, j.ts)
	return data, err
}

func (j *KafkaControllerMment) LineProto() (*io.Point, error) {
	return j.LineProto()
}

func (j *KafkaControllerMment) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name:   "controller",
		Fields: nil,
		Tags:   nil,
	}
}

func (j *KafkaReplicaMment) LineProto() (*io.Point, error) {
	return j.LineProto()
}

func (j *KafkaReplicaMment) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name:   "replica_manager",
		Fields: nil,
		Tags:   nil,
	}
}

func (j *KafkaPurgatoryMment) LineProto() (*io.Point, error) {
	return j.LineProto()
}

func (j *KafkaPurgatoryMment) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name:   "purgatory",
		Fields: nil,
		Tags:   nil,
	}
}

func (j *KafkaClientMment) LineProto() (*io.Point, error) {
	return j.LineProto()
}

func (j *KafkaClientMment) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name:   "client",
		Fields: nil,
		Tags:   nil,
	}
}

func (j *KafkaRequestMment) LineProto() (*io.Point, error) {
	return j.LineProto()
}

func (j *KafkaRequestMment) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name:   "request",
		Fields: nil,
		Tags:   nil,
	}
}

func (j *KafkaTopicsMment) LineProto() (*io.Point, error) {
	return j.LineProto()
}

func (j *KafkaTopicsMment) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name:   "topics",
		Fields: nil,
		Tags:   nil,
	}
}

func (j *KafkaTopicMment) LineProto() (*io.Point, error) {
	return j.LineProto()
}

func (j *KafkaTopicMment) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name:   "topic",
		Fields: nil,
		Tags:   nil,
	}
}

func (j *KafkaPartitionMment) LineProto() (*io.Point, error) {
	return j.LineProto()
}

func (j *KafkaPartitionMment) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name:   "partition",
		Fields: nil,
		Tags:   nil,
	}
}
