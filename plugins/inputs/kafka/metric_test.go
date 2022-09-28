// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package kafka

import (
	"testing"

	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/lineproto"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io/point"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
)

func TestMeasurement(t *testing.T) {
	cases := []struct {
		m inputs.Measurement
	}{
		{
			m: &KafkaPartitionMment{
				KafkaMeasurement{
					name:   "kafka_partition",
					tags:   inputs.BuildTags(t, partitionTags),
					fields: inputs.BuildFields(t, partitionFields),
				},
			},
		},
		{
			m: &KafkaTopicMment{
				KafkaMeasurement{
					name:   "kafka_topic",
					tags:   inputs.BuildTags(t, topicTags),
					fields: inputs.BuildFields(t, topicFields),
				},
			},
		},

		{
			m: &KafkaTopicsMment{
				KafkaMeasurement{
					name:   "kafka_topics",
					tags:   inputs.BuildTags(t, topicsTags),
					fields: inputs.BuildFields(t, topicsFields),
				},
			},
		},

		{
			m: &KafkaRequestMment{
				KafkaMeasurement{
					name:   "kafka_request",
					tags:   inputs.BuildTags(t, requestTags),
					fields: inputs.BuildFields(t, requestFields),
				},
			},
		},

		{
			m: &KafkaPurgatoryMment{
				KafkaMeasurement{
					name:   "kafka_purgatory",
					tags:   inputs.BuildTags(t, purgatoryTags),
					fields: inputs.BuildFields(t, purgatoryFields),
				},
			},
		},

		{
			m: &KafkaReplicaMment{
				KafkaMeasurement{
					name:   "kafka_replica_manager",
					tags:   inputs.BuildTags(t, replicationTags),
					fields: inputs.BuildFields(t, replicationFields),
				},
			},
		},

		{
			m: &KafkaControllerMment{
				KafkaMeasurement{
					name:   "kafka_controller",
					tags:   inputs.BuildTags(t, controllerTags),
					fields: inputs.BuildFields(t, controllerFields),
				},
			},
		},
	}

	encoder := lineproto.NewLineEncoder()
	for _, tc := range cases {
		t.Run("", func(t *testing.T) {
			if pt, err := tc.m.LineProto(); err != nil {
				t.Fatal(err)
			} else {
				encoder.Reset()
				if err := encoder.AppendPoint(pt.Point); err != nil {
					t.Fatal(err)
				}
				line, err := encoder.UnsafeStringWithoutLn()
				if err != nil {
					t.Fatal(err)
				}
				t.Log(line)
				fs := pt.Fields
				ts := pt.Tags

				if len(fs) > point.MaxFields {
					t.Errorf("exceed max fields(%d > %d)", len(fs), point.MaxFields)
				}
				if len(ts) > point.MaxTags {
					t.Errorf("exceed max tags(%d > %d)", len(ts), point.MaxTags)
				}

				t.Logf("fields count: %d", len(fs))
			}
		})
	}
}
