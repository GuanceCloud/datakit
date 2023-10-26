// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package elasticsearch

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/GuanceCloud/cliutils/point"
	"github.com/stretchr/testify/assert"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/datakit"
)

var (
	uu          = "http://example.com:9200"
	clusterName = "elasticsearch_cluster"
)

type transportMock struct {
	statusCode int
	body       string
}

func (t *transportMock) RoundTrip(r *http.Request) (*http.Response, error) {
	res := &http.Response{
		Header:     make(http.Header),
		Request:    r,
		StatusCode: t.statusCode,
	}
	res.Header.Set("Content-Type", "application/json")
	res.Body = ioutil.NopCloser(strings.NewReader(t.body))
	return res, nil
}

func (t *transportMock) CancelRequest(_ *http.Request) {
}

func newTransportMock(body string) http.RoundTripper {
	return &transportMock{
		statusCode: http.StatusOK,
		body:       body,
	}
}

//nolint:deadcode,unused
func defaultTags() map[string]string {
	return map[string]string{
		"cluster_name":          "es-testcluster",
		"node_attribute_master": "true",
		"node_id":               "SDFsfSDFsdfFSDSDfSFDSDF",
		"node_name":             "test.host.com",
		"node_host":             "test",
		"node_roles":            "data,ingest,master",
	}
}

func defaultServerInfo() serverInfo {
	return serverInfo{nodeID: "", masterID: "SDFsfSDFsdfFSDSDfSFDSDF"}
}

func TestGatherNodeStats(t *testing.T) {
	es := newElasticsearchWithClient()
	es.Servers = []string{uu}
	es.client.Transport = newTransportMock(nodeStatsResponse)
	es.serverInfo = make(map[string]serverInfo)
	es.serverInfo[uu] = defaultServerInfo()

	if _, err := es.gatherNodeStats(""); err != nil {
		t.Fatal(err)
	}

	tags := defaultTags()

	checkIsMaster(t, es, es.Servers[0], false)

	nodestatsExpected := nodestatsExpectedFunc()

	for field, value := range nodestatsExpected {
		fields := `fs_total_available_in_bytes,fs_total_free_in_bytes,fs_total_total_in_bytes,fs_data_0_available_in_bytes,fs_data_0_free_in_bytes,fs_data_0_total_in_bytes`

		if strings.Contains(fields, field) {
			if value, ok := value.(float64); ok {
				val := value / (1024 * 1024 * 1024)
				filedName := strings.ReplaceAll(field, "in_bytes", "in_gigabytes")
				nodestatsExpected[filedName] = val
			}
			delete(nodestatsExpected, field)
			continue
		}
		_, ok := nodeStatsFields[field]
		if !ok {
			delete(nodestatsExpected, field)
		}
	}

	AssertContainsTaggedFields(t, "elasticsearch_node_stats", nodestatsExpected, tags, es.collectCache)
}

func TestUtilDuration(t *testing.T) {
	d := Duration{Duration: time.Second}
	err := d.UnmarshalTOML([]byte("1s"))
	if err != nil {
		t.Fatal(err)
	}

	err = d.UnmarshalTOML([]byte("1619059234299"))
	if err != nil {
		t.Fatal(err)
	}

	err = d.UnmarshalTOML([]byte("1619059234299.001"))
	if err != nil {
		t.Fatal(err)
	}
}

func TestCollect(t *testing.T) {
	es := newElasticsearchWithClient()
	es.Servers = []string{uu}
	es.ClusterHealth = true
	es.ClusterStats = true
	es.ClusterHealthLevel = ""
	es.client.Transport = newTransportMock(clusterHealthResponse)
	es.serverInfo = make(map[string]serverInfo)
	es.serverInfo[uu] = defaultServerInfo()

	if err := es.Collect(); err != nil {
		t.Fatal(err)
	}
}

func TestGatherClusterHealthEmptyClusterHealth(t *testing.T) {
	es := newElasticsearchWithClient()
	es.Servers = []string{uu}
	es.ClusterHealth = true
	es.ClusterHealthLevel = ""
	es.client.Transport = newTransportMock(clusterHealthResponse)
	es.serverInfo = make(map[string]serverInfo)
	es.serverInfo[uu] = defaultServerInfo()

	if err := es.gatherClusterHealth("", ""); err != nil {
		t.Fatal(err)
	}

	checkIsMaster(t, es, es.Servers[0], false)

	AssertContainsTaggedFields(t, "elasticsearch_cluster_health",
		clusterHealthExpected,
		map[string]string{"name": clusterName, "cluster_name": clusterName, "cluster_status": "green"}, es.collectCache)
}

func TestGatherClusterHealthSpecificClusterHealth(t *testing.T) {
	es := newElasticsearchWithClient()
	es.Servers = []string{uu}
	es.ClusterHealth = true
	es.ClusterHealthLevel = "cluster"
	es.client.Transport = newTransportMock(clusterHealthResponse)
	es.serverInfo = make(map[string]serverInfo)
	es.serverInfo[uu] = defaultServerInfo()

	if err := es.gatherClusterHealth("", ""); err != nil {
		t.Fatal(err)
	}

	checkIsMaster(t, es, es.Servers[0], false)

	AssertContainsTaggedFields(t, "elasticsearch_cluster_health",
		clusterHealthExpected,
		map[string]string{"name": clusterName, "cluster_name": clusterName, "cluster_status": "green"}, es.collectCache)

	AssertDoesNotContainsTaggedFields(t, "elasticsearch_cluster_health_indices",
		v1IndexExpected,
		map[string]string{"index": "v1"}, es.collectCache)

	AssertDoesNotContainsTaggedFields(t, "elasticsearch_cluster_health_indices",
		v2IndexExpected,
		map[string]string{"index": "v2"}, es.collectCache)
}

func TestGatherClusterHealthAlsoIndicesHealth(t *testing.T) {
	es := newElasticsearchWithClient()
	es.Servers = []string{uu}
	es.ClusterHealth = true
	es.ClusterHealthLevel = "indices"
	es.client.Transport = newTransportMock(clusterHealthResponseWithIndices)
	es.serverInfo = make(map[string]serverInfo)
	es.serverInfo[uu] = defaultServerInfo()

	if err := es.gatherClusterHealth("", ""); err != nil {
		t.Fatal(err)
	}

	checkIsMaster(t, es, es.Servers[0], false)

	clusterHealthExpected["indices_lifecycle_error_count"] = int64(2)
	AssertContainsTaggedFields(t, "elasticsearch_cluster_health",
		clusterHealthExpected,
		map[string]string{"name": clusterName, "cluster_name": clusterName, "cluster_status": "green"}, es.collectCache)
}

func TestGatherClusterIndicesStats(t *testing.T) {
	es := newElasticsearchWithClient()
	es.IndicesInclude = []string{"_all"}
	es.Servers = []string{uu}
	es.client.Transport = newTransportMock(clusterIndicesResponse)
	es.serverInfo = make(map[string]serverInfo)
	es.serverInfo[uu] = defaultServerInfo()

	if err := es.gatherIndicesStats("", ""); err != nil {
		t.Fatal(err)
	}

	AssertContainsTaggedFields(t, "elasticsearch_indices_stats",
		clusterIndicesTotalExpected,
		map[string]string{"index_name": "es", "cluster_name": ""}, es.collectCache)

	AssertContainsTaggedFields(t, "elasticsearch_indices_stats",
		clusterIndicesTotalExpected,
		map[string]string{"index_name": "_all", "cluster_name": ""}, es.collectCache)
}

func TestMapHealthStatusToCode(t *testing.T) {
	assert.Equal(t, mapHealthStatusToCode("GREEN"), 1)
	assert.Equal(t, mapHealthStatusToCode("YELLOW"), 2)
	assert.Equal(t, mapHealthStatusToCode("RED"), 3)
	assert.Equal(t, mapHealthStatusToCode("NULL"), 0)
}

func TestInput(t *testing.T) {
	es := newElasticsearchWithClient()
	assert.Equal(t, es.Catalog(), "db")
	assert.Equal(t, es.SampleConfig(), sampleConfig)

	pipelineMap := es.PipelineConfig()
	assert.Equal(t, pipelineMap["elasticsearch"], pipelineCfg)

	assert.Equal(t, es.AvailableArchs(), datakit.AllOSWithElection)

	samples := es.SampleMeasurement()
	assert.Greater(t, len(samples), 0)
}

func TestCreateHTTPClient(t *testing.T) {
	es := newElasticsearchWithClient()
	_, err := es.createHTTPClient()
	if err != nil {
		t.Fail()
	}

	es.TLSOpen = true
	_, err = es.createHTTPClient()
	if err == nil {
		t.Fail()
	}
}

func TestTlsConfig(t *testing.T) {
	if _, err := TLSConfig("", "", ""); err == nil {
		t.Fail()
	}
}

func TestGatherClusterStatsMaster(t *testing.T) {
	es := newElasticsearchWithClient()
	es.ClusterStats = true
	es.Servers = []string{uu}
	es.serverInfo = make(map[string]serverInfo)
	info := serverInfo{nodeID: "SDFsfSDFsdfFSDSDfSFDSDF", masterID: ""}

	es.client.Transport = newTransportMock(IsMasterResult)
	masterID, err := es.getCatMaster("")
	if err != nil {
		t.Fatal(err)
	}
	info.masterID = masterID
	es.serverInfo[uu] = info

	IsMasterResultTokens := strings.Split(string(IsMasterResult), " ")

	if masterID != IsMasterResultTokens[0] {
		assert.Fail(t, "catmaster is incorrect")
	}

	es.Local = true
	es.client.Transport = newTransportMock(nodeStatsResponse)

	if _, err := es.gatherNodeStats(""); err != nil {
		t.Fatal(err)
	}

	checkIsMaster(t, es, es.Servers[0], true)
	tags := defaultTags()
	nodestatsExpected := nodestatsExpectedFunc()
	AssertContainsTaggedFields(t, "elasticsearch_node_stats", nodestatsExpected, tags, es.collectCache)

	es.client.Transport = newTransportMock(clusterStatsResponse)

	if err := es.gatherClusterStats(""); err != nil {
		t.Fatal(err)
	}

	tags = map[string]string{
		"cluster_name": "es-testcluster",
		"node_name":    "test.host.com",
		"status":       "red",
	}

	for field := range clusterstatsExpected {
		_, ok := clusterStatsFields[field]
		if !ok {
			delete(clusterstatsExpected, field)
		}
	}

	AssertContainsTaggedFields(t, "elasticsearch_cluster_stats", clusterstatsExpected, tags, es.collectCache)
}

func AssertContainsTaggedFields(t *testing.T,
	measurement string,
	fields map[string]interface{},
	tags map[string]string,
	collectCache []*point.Point,
) {
	t.Helper()
	for _, p := range collectCache {
		gotTags := p.MapTags()

		// for k, v := range gotTags {
		// 	fmt.Printf("gotTags k = %s, v = %s\n", k, v)
		// }

		// fmt.Println("============")

		// for k, v := range tags {
		// 	fmt.Printf("tags k = %s, v = %s\n", k, v)
		// }

		if !reflect.DeepEqual(tags, gotTags) {
			continue
		}

		if p.Name() != measurement {
			continue
		}

		gotFields := p.InfluxFields()

		// for k, v := range gotFields {
		// 	fmt.Printf("gotFields k = %s, v = %s\n", k, v)
		// }

		// fmt.Println("============")

		// for k, v := range fields {
		// 	fmt.Printf("fields k = %s, v = %s\n", k, v)
		// }

		if reflect.DeepEqual(fields, gotFields) {
			return
		}
	}

	for _, p := range collectCache {
		if p.Name() == measurement {
			t.Log("measurement", p.Name(), "tags", p.InfluxTags(), "fields", p.InfluxFields())
		}
	}

	assert.Fail(t, fmt.Sprintf("unknown measurement %q with tags %v", measurement, tags))
}

func AssertDoesNotContainsTaggedFields(t *testing.T,
	measurement string,
	fields map[string]interface{},
	tags map[string]string,
	collectCache []*point.Point,
) {
	t.Helper()
	for _, p := range collectCache {
		if !reflect.DeepEqual(tags, p.Tags()) {
			continue
		}
		if p.Name() == measurement && reflect.DeepEqual(fields, p.Fields()) {
			assert.Fail(t, fmt.Sprintf("found measurement %s with tagged fields (tags %v) which should not be there", measurement, tags))
		}
	}
}

func checkIsMaster(t *testing.T, es *Input, server string, expected bool) {
	t.Helper()
	if es.serverInfo[server].isMaster() != expected {
		assert.Fail(t, "IsMaster set incorrectly")
	}
}

func newElasticsearchWithClient() *Input {
	es := defaultInput()
	es.client = &http.Client{}
	return es
}

func TestGetVersion(t *testing.T) {
	es := newElasticsearchWithClient()
	es.Servers = []string{uu}
	es.client.Transport = newTransportMock(clusterInfo)
	version, err := es.getVersion("")

	assert.NoError(t, err)
	assert.Equal(t, "6.8.20", version)
}

func TestGetLifeCycleErrorCount(t *testing.T) {
	es := newElasticsearchWithClient()
	es.client.Transport = newTransportMock(lifeCycleStateResponse)
	es.serverInfo = make(map[string]serverInfo)
	es.serverInfo[uu] = serverInfo{version: "6.8"}

	assert.Equal(t, 1, es.getLifeCycleErrorCount(uu))

	es.serverInfo[uu] = serverInfo{version: "7.8"}
	assert.Equal(t, 2, es.getLifeCycleErrorCount(uu))
}

func TestMetric(t *testing.T) {
	m := elasticsearchMeasurement{}
	assert.Equal(t, m.Info().Name, inputName)
	m1 := nodeStatsMeasurement{}
	assert.Equal(t, m1.Info().Name, "elasticsearch_node_stats")
	m2 := clusterStatsMeasurement{}
	assert.Equal(t, m2.Info().Name, "elasticsearch_cluster_stats")
	m3 := clusterHealthMeasurement{}
	assert.Equal(t, m3.Info().Name, "elasticsearch_cluster_health")
	m4 := clusterHealthIndicesMeasurement{}
	assert.Equal(t, m4.Info().Name, "elasticsearch_cluster_health_indices")
	m5 := indicesStatsShardsTotalMeasurement{}
	assert.Equal(t, m5.Info().Name, "elasticsearch_indices_stats_shards_total")
	m6 := indicesStatsMeasurement{}
	assert.Equal(t, m6.Info().Name, "elasticsearch_indices_stats")
	m7 := indicesStatsShardsMeasurement{}
	assert.Equal(t, m7.Info().Name, "elasticsearch_indices_stats_shards")
}

func TestSetServerInfo(t *testing.T) {
	es := newElasticsearchWithClient()
	es.Servers = []string{uu}
	es.client.Transport = newTransportMock(clusterInfo)

	err := es.setServerInfo()
	assert.NoError(t, err)
	assert.Equal(t, clusterInfoExpected["version"], es.serverInfo[uu].version)
}

func TestGetUserPrivilege(t *testing.T) {
	es := newElasticsearchWithClient()
	es.client.Transport = newTransportMock(privilegeResponse)

	p := es.getUserPrivilege("")

	assert.True(t, p.Cluster.Monitor)
}
