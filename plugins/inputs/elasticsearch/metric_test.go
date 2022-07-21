// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package elasticsearch

import (
	"testing"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io/point"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
)

func TestMeasurement(t *testing.T) {
	cases := []struct {
		m inputs.Measurement
	}{
		{
			m: &indicesStatsShardsMeasurement{
				elasticsearchMeasurement: elasticsearchMeasurement{
					name:   "elasticsearch_indices_stats_shards",
					tags:   inputs.BuildTags(t, indicesStatsShardsTags),
					fields: inputs.BuildFields(t, indicesStatsShardsFields),
				},
			},
		},
		{
			m: &indicesStatsMeasurement{
				elasticsearchMeasurement: elasticsearchMeasurement{
					name:   "elasticsearch_indices_stats",
					tags:   inputs.BuildTags(t, indicesStatsTags),
					fields: inputs.BuildFields(t, indicesStatsFields),
				},
			},
		},

		{
			m: &clusterHealthMeasurement{
				elasticsearchMeasurement: elasticsearchMeasurement{
					name:   "elasticsearch_cluster_health",
					tags:   inputs.BuildTags(t, clusterHealthTags),
					fields: inputs.BuildFields(t, clusterHealthFields),
				},
			},
		},
		{
			m: &clusterStatsMeasurement{
				elasticsearchMeasurement: elasticsearchMeasurement{
					name:   "elasticsearch_cluster_stats",
					tags:   inputs.BuildTags(t, clusterStatsTags),
					fields: inputs.BuildFields(t, clusterStatsFields),
				},
			},
		},

		{
			m: &nodeStatsMeasurement{
				elasticsearchMeasurement: elasticsearchMeasurement{
					name:   "elasticsearch_node_stats",
					tags:   inputs.BuildTags(t, nodeStatsTags),
					fields: inputs.BuildFields(t, nodeStatsFields),
				},
			},
		},

		{
			m: &elasticsearchMeasurement{
				name:   "elasticsearch",
				tags:   make(map[string]string),
				fields: inputs.BuildFields(t, elasticsearchMeasurementFields),
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

				if len(fs) > point.MaxFields {
					t.Errorf("exceed max fields(%d > %d)", len(fs), point.MaxFields)
				}
				if len(ts) > point.MaxTags {
					t.Errorf("exceed max tags(%d > %d)", len(ts), point.MaxTags)
				}

				t.Log(pt.String())
			}
		})
	}
}
