package elasticsearch

import (
	"testing"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
)

func TestIndicesStatsShardsMeasurement(t *testing.T) {
	m := &indicesStatsShardsMeasurement{
		elasticsearchMeasurement: elasticsearchMeasurement{
			name: "indicesStatsShardsMeasurement",
			tags: make(map[string]string),
		},
	}

	if info := m.Info(); info.Name != "elasticsearch_indices_stats_shards" {
		t.Fatal("Info name invalid")
	}
}

func TestIndicesStatsMeasurement(t *testing.T) {
	m := &indicesStatsMeasurement{
		elasticsearchMeasurement: elasticsearchMeasurement{
			name: "indicesStatsMeasurement",
			tags: make(map[string]string),
		},
	}

	if info := m.Info(); info.Name != "elasticsearch_indices_stats" {
		t.Fatal("Info name invalid")
	}
}

func TestIndicesStatsShardsTotalMeasurement(t *testing.T) {
	m := &indicesStatsShardsTotalMeasurement{
		elasticsearchMeasurement: elasticsearchMeasurement{
			name: "indicesStatsShardsTotalMeasurement",
			tags: make(map[string]string),
		},
	}

	if info := m.Info(); info.Name != "elasticsearch_indices_stats_shards_total" {
		t.Fatal("Info name invalid")
	}
}

func TestClusterHealthIndicesMeasurement(t *testing.T) {
	m := &clusterHealthIndicesMeasurement{
		elasticsearchMeasurement: elasticsearchMeasurement{
			name: "clusterHealthIndicesMeasurement",
			tags: make(map[string]string),
		},
	}

	if info := m.Info(); info.Name != "elasticsearch_cluster_health_indices" {
		t.Fatal("Info name invalid")
	}
}

func TestClusterHealthMeasurement(t *testing.T) {
	m := &clusterHealthMeasurement{
		elasticsearchMeasurement: elasticsearchMeasurement{
			name: "clusterHealthMeasurement",
			tags: make(map[string]string),
		},
	}
	if info := m.Info(); info.Name != "elasticsearch_cluster_health" {
		t.Fatal("Info name invalid")
	}
}

func TestClusterStatsMeasurement(t *testing.T) {
	m := &clusterStatsMeasurement{
		elasticsearchMeasurement: elasticsearchMeasurement{
			name: "clusterStatsMeasurement",
			tags: make(map[string]string),
		},
	}

	if info := m.Info(); info.Name != "elasticsearch_cluster_stats" {
		t.Fatal("Info name invalid")
	}
}

func TestNodeStatsMeasurement(t *testing.T) {
	m := &nodeStatsMeasurement{
		elasticsearchMeasurement: elasticsearchMeasurement{
			name: "nodeStatsMeasurement",
			tags: make(map[string]string),
			fields: func() map[string]interface{} {
				x := map[string]interface{}{}
				for k, v := range nodeStatsFields {
					switch _v := v.(type) {
					case *inputs.FieldInfo:
						switch _v.DataType {
						case inputs.Float:
							x[k] = 1.23
						case inputs.Int:
							x[k] = 123
						case inputs.String:
							x[k] = "abc123"
						case inputs.Bool:
							x[k] = false
						default:
							t.Errorf("invalid data field for field: %s", k)
						}

					default:
						t.Errorf("expect *inputs.FieldInfo")
					}
				}
				return x
			}(),
		},
	}

	if info := m.Info(); info.Name != "elasticsearch_node_stats" {
		t.Fatal("Info name invalid")
	}

	pt, err := m.LineProto()
	if err != nil {
		t.Fatal(err)
	}

	t.Log(pt.String())
}

func TestElasticsearchMeasurement(t *testing.T) {
	fields := make(map[string]interface{})
	fields["a"] = "11111"
	m := &elasticsearchMeasurement{
		name:   "elasticsearch",
		tags:   make(map[string]string),
		fields: fields,
	}
	if info := m.Info(); info.Name != "elasticsearch" {
		t.Fatal("Info name invalid")
	}

	if _, err := m.LineProto(); err != nil {
		t.Fatal(err)
	}
}
