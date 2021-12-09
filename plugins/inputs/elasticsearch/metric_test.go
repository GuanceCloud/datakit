package elasticsearch

import (
	"testing"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
)

func buildTags(t *testing.T, ti map[string]interface{}) map[string]string {
	t.Helper()
	x := map[string]string{}
	for k, _ := range ti {
		x[k] = "nodeStatsMeasurement-tags"
	}
	return x
}

func buildFields(t *testing.T, fi map[string]interface{}) map[string]interface{} {
	t.Helper()
	x := map[string]interface{}{}
	for k, v := range fi {
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
}

func TestIndicesStatsShardsMeasurement(t *testing.T) {
	m := &indicesStatsShardsMeasurement{
		elasticsearchMeasurement: elasticsearchMeasurement{
			name:   "elasticsearch_indices_stats_shards",
			tags:   buildTags(t, indicesStatsShardsTags),
			fields: buildFields(t, indicesStatsShardsFields),
		},
	}

	if info := m.Info(); info.Name != "elasticsearch_indices_stats_shards" {
		t.Fatal("Info name invalid")
	}

	if pt, err := m.LineProto(); err != nil {
		t.Fatal(err)
	} else {
		t.Log(pt.String())
	}
}

func TestIndicesStatsMeasurement(t *testing.T) {
	m := &indicesStatsMeasurement{
		elasticsearchMeasurement: elasticsearchMeasurement{
			name:   "elasticsearch_indices_stats",
			tags:   buildTags(t, indicesStatsTags),
			fields: buildFields(t, indicesStatsFields),
		},
	}

	if info := m.Info(); info.Name != "elasticsearch_indices_stats" {
		t.Fatal("Info name invalid")
	}

	if pt, err := m.LineProto(); err != nil {
		t.Fatal(err)
	} else {
		t.Log(pt.String())
	}
}

func TestClusterHealthMeasurement(t *testing.T) {
	m := &clusterHealthMeasurement{
		elasticsearchMeasurement: elasticsearchMeasurement{
			name:   "elasticsearch_cluster_health",
			tags:   buildTags(t, clusterHealthTags),
			fields: buildFields(t, clusterHealthFields),
		},
	}
	if info := m.Info(); info.Name != "elasticsearch_cluster_health" {
		t.Fatal("Info name invalid")
	}

	if pt, err := m.LineProto(); err != nil {
		t.Fatal(err)
	} else {
		t.Log(pt.String())
	}
}

func TestClusterStatsMeasurement(t *testing.T) {
	m := &clusterStatsMeasurement{
		elasticsearchMeasurement: elasticsearchMeasurement{
			name:   "elasticsearch_cluster_stats",
			tags:   buildTags(t, clusterStatsTags),
			fields: buildFields(t, clusterStatsFields),
		},
	}

	if info := m.Info(); info.Name != "elasticsearch_cluster_stats" {
		t.Fatal("Info name invalid")
	}

	if pt, err := m.LineProto(); err != nil {
		t.Fatal(err)
	} else {
		t.Log(pt.String())
	}
}

func TestNodeStatsMeasurement(t *testing.T) {
	m := &nodeStatsMeasurement{
		elasticsearchMeasurement: elasticsearchMeasurement{
			name:   "elasticsearch_node_stats",
			tags:   buildTags(t, nodeStatsTags),
			fields: buildFields(t, nodeStatsFields),
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
	m := &elasticsearchMeasurement{
		name:   "elasticsearch",
		tags:   make(map[string]string),
		fields: buildFields(t, elasticsearchMeasurementFields),
	}

	if info := m.Info(); info.Name != "elasticsearch" {
		t.Fatal("Info name invalid")
	}

	if pt, err := m.LineProto(); err != nil {
		t.Fatal(err)
	} else {
		t.Log(pt.String())
	}
}
