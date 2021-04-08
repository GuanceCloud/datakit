package elasticsearch

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"reflect"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
)

var (
	url         = "http://example.com:9200"
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
	es.Servers = []string{url}
	es.client.Transport = newTransportMock(nodeStatsResponse)
	es.serverInfo = make(map[string]serverInfo)
	es.serverInfo[url] = defaultServerInfo()

	if err := es.gatherNodeStats(""); err != nil {
		t.Fatal(err)
	}

	tags := defaultTags()

	checkIsMaster(es, es.Servers[0], false, t)

	AssertContainsTaggedFields(t, "elasticsearch_node_stats", nodestatsExpected, tags, es.collectCache)
}

func TestGatherClusterHealthEmptyClusterHealth(t *testing.T) {
	es := newElasticsearchWithClient()
	es.Servers = []string{url}
	es.ClusterHealth = true
	es.ClusterHealthLevel = ""
	es.client.Transport = newTransportMock(clusterHealthResponse)
	es.serverInfo = make(map[string]serverInfo)
	es.serverInfo[url] = defaultServerInfo()

	if err := es.gatherClusterHealth(""); err != nil {
		t.Fatal(err)
	}

	checkIsMaster(es, es.Servers[0], false, t)

	AssertContainsTaggedFields(t, "elasticsearch_cluster_health",
		clusterHealthExpected,
		map[string]string{"name": clusterName}, es.collectCache)

	AssertDoesNotContainsTaggedFields(t, "elasticsearch_cluster_health_indices",
		v1IndexExpected,
		map[string]string{"index": "v1"}, es.collectCache)

	AssertDoesNotContainsTaggedFields(t, "elasticsearch_cluster_health_indices",
		v2IndexExpected,
		map[string]string{"index": "v2"}, es.collectCache)
}

func TestGatherClusterHealthSpecificClusterHealth(t *testing.T) {
	es := newElasticsearchWithClient()
	es.Servers = []string{url}
	es.ClusterHealth = true
	es.ClusterHealthLevel = "cluster"
	es.client.Transport = newTransportMock(clusterHealthResponse)
	es.serverInfo = make(map[string]serverInfo)
	es.serverInfo[url] = defaultServerInfo()

	if err := es.gatherClusterHealth(""); err != nil {
		t.Fatal(err)
	}

	checkIsMaster(es, es.Servers[0], false, t)

	AssertContainsTaggedFields(t, "elasticsearch_cluster_health",
		clusterHealthExpected,
		map[string]string{"name": clusterName}, es.collectCache)

	AssertDoesNotContainsTaggedFields(t, "elasticsearch_cluster_health_indices",
		v1IndexExpected,
		map[string]string{"index": "v1"}, es.collectCache)

	AssertDoesNotContainsTaggedFields(t, "elasticsearch_cluster_health_indices",
		v2IndexExpected,
		map[string]string{"index": "v2"}, es.collectCache)
}

func TestGatherClusterHealthAlsoIndicesHealth(t *testing.T) {
	es := newElasticsearchWithClient()
	es.Servers = []string{url}
	es.ClusterHealth = true
	es.ClusterHealthLevel = "indices"
	es.client.Transport = newTransportMock(clusterHealthResponseWithIndices)
	es.serverInfo = make(map[string]serverInfo)
	es.serverInfo[url] = defaultServerInfo()

	if err := es.gatherClusterHealth(""); err != nil {
		t.Fatal(err)
	}

	checkIsMaster(es, es.Servers[0], false, t)

	AssertContainsTaggedFields(t, "elasticsearch_cluster_health",
		clusterHealthExpected,
		map[string]string{"name": clusterName}, es.collectCache)

	AssertContainsTaggedFields(t, "elasticsearch_cluster_health_indices",
		v1IndexExpected,
		map[string]string{"index": "v1", "name": clusterName}, es.collectCache)

	AssertContainsTaggedFields(t, "elasticsearch_cluster_health_indices",
		v2IndexExpected,
		map[string]string{"index": "v2", "name": clusterName}, es.collectCache)
}

func TestGatherClusterStatsMaster(t *testing.T) {
	es := newElasticsearchWithClient()
	es.ClusterStats = true
	es.Servers = []string{url}
	es.serverInfo = make(map[string]serverInfo)
	info := serverInfo{nodeID: "SDFsfSDFsdfFSDSDfSFDSDF", masterID: ""}

	es.client.Transport = newTransportMock(IsMasterResult)
	masterID, err := es.getCatMaster("")
	if err != nil {
		t.Fatal(err)
	}
	info.masterID = masterID
	es.serverInfo[url] = info

	IsMasterResultTokens := strings.Split(string(IsMasterResult), " ")

	if masterID != IsMasterResultTokens[0] {
		assert.Fail(t, "catmaster is incorrect")
	}

	es.Local = true
	es.client.Transport = newTransportMock(nodeStatsResponse)

	if err := es.gatherNodeStats(""); err != nil {
		t.Fatal(err)
	}

	checkIsMaster(es, es.Servers[0], true, t)
	tags := defaultTags()
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

	AssertContainsTaggedFields(t, "elasticsearch_cluster_stats", clusterstatsExpected, tags, es.collectCache)
}

func AssertContainsTaggedFields(t *testing.T, measurement string, fields map[string]interface{}, tags map[string]string, collectCache []inputs.Measurement) {
	for _, metric := range collectCache {
		m := metric.(*elasticsearchMeasurement)
		if !reflect.DeepEqual(tags, m.tags) {
			continue
		}
		if m.name == measurement && reflect.DeepEqual(fields, m.fields) {
			return
		}
	}

	for _, metric := range collectCache {
		m := metric.(*elasticsearchMeasurement)
		if m.name == measurement {
			t.Log("measurement", m.name, "tags", m.tags, "fields", m.fields)
		}
	}

	assert.Fail(t, fmt.Sprintf("unknown measurement %q with tags %v", measurement, tags))
}

func AssertDoesNotContainsTaggedFields(
	t *testing.T,
	measurement string,
	fields map[string]interface{},
	tags map[string]string,
	collectCache []inputs.Measurement,
) {
	for _, metric := range collectCache {
		m := metric.(*elasticsearchMeasurement)
		if !reflect.DeepEqual(tags, m.tags) {
			continue
		}
		fmt.Println(m.tags)
		if m.name == measurement && reflect.DeepEqual(fields, m.fields) {
			assert.Fail(t, fmt.Sprintf("found measurement %s with tagged fields (tags %v) which should not be there", measurement, tags))
		}
	}
}

func checkIsMaster(es *Input, server string, expected bool, t *testing.T) {
	if es.serverInfo[server].isMaster() != expected {
		assert.Fail(t, "IsMaster set incorrectly")
	}
}

func newElasticsearchWithClient() *Input {
	es := NewElasticsearch()
	es.client = &http.Client{}
	return es
}
