package kafka

import (
	"testing"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io"
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

	for _, tc := range cases {
		t.Run("", func(t *testing.T) {
			if pt, err := tc.m.LineProto(); err != nil {
				t.Fatal(err)
			} else {
				t.Log(pt.String())
				fs, err := pt.Fields()
				if err != nil {
					t.Error(err)
				}
				ts := pt.Tags()

				if len(fs) > io.MaxFields {
					t.Errorf("exceed max fields(%d > %d)", len(fs), io.MaxFields)
				}
				if len(ts) > io.MaxTags {
					t.Errorf("exceed max tags(%d > %d)", len(ts), io.MaxTags)
				}

				t.Logf("fields count: %d", len(fs))
			}
		})
	}
}
