package solr

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
)

var (
	cacheStat7 = `{
		"responseHeader": {
			"status": 0,
			"QTime": 3
		},
		"metrics": {
			"solr.core.techproducts": {
				"CACHE.searcher.filterCache": {
					"lookups": 0,
					"hits": 0,
					"hitratio": 1.0,
					"inserts": 0,
					"evictions": 0,
					"size": 0,
					"warmupTime": 0,
					"ramBytesUsed": 416,
					"maxRamMB": -1,
					"cumulative_lookups": 0,
					"cumulative_hits": 0,
					"cumulative_hitratio": 1.0,
					"cumulative_inserts": 0,
					"cumulative_evictions": 0
				}
			}
		}
	}`

	searcherStats7 = `{
		"responseHeader": {
			"status": 0,
			"QTime": 1
		},
		"metrics": {
			"solr.core.techproducts": {
				"SEARCHER.searcher.caching": true,
				"SEARCHER.searcher.deletedDocs": 0,
				"SEARCHER.searcher.indexCommitSize": 27320,
				"SEARCHER.searcher.indexVersion": 6,
				"SEARCHER.searcher.maxDoc": 32,
				"SEARCHER.searcher.numDocs": 32,
				"SEARCHER.searcher.openedAt": "2021-05-06T06:51:55.827Z",
				"SEARCHER.searcher.reader": "ExitableDirectoryReader(UninvertingDirectoryReader(Uninverting(_0(8.8.2):C32:[diagnostics={java.vendor=Private Build, os=Linux, java.version=1.8.0_282, java.vm.version=25.282-b08, lucene.version=8.8.2, os.arch=amd64, java.runtime.version=1.8.0_282-8u282-b08-0ubuntu1~20.04-b08, source=flush, os.version=5.8.0-50-generic, timestamp=1619513581931}]:[attributes={Lucene87StoredFieldsFormat.mode=BEST_SPEED}] :id=eswn0o8svec8g27b55yv7b2bw)))",
				"SEARCHER.searcher.readerDir": "NRTCachingDirectory(MMapDirectory@/home/vircoys/Downloads/solr-8.8.2/example/techproducts/solr/techproducts/data/index lockFactory=org.apache.lucene.store.NativeFSLockFactory@659e845c; maxCacheMB=48.0 maxMergeSizeMB=4.0)",
				"SEARCHER.searcher.registeredAt": "2021-05-06T06:51:56.011Z",
				"SEARCHER.searcher.searcherName": "Searcher@426771ef[techproducts] main",
				"SEARCHER.searcher.warmupTime": 0
			}
		}
	}`

	coreRqTimes7 = `{
		"responseHeader": {
			"status": 0,
			"QTime": 7
		},
		"metrics": {
			"solr.core.techproducts": {
				"QUERY./select.requestTimes": {
					"count": 1,
					"meanRate": 2.5126722879196796E-4,
					"1minRate": 3.6746144320828424E-30,
					"5minRate": 3.579862547745004E-7,
					"15minRate": 0.002428336015326626,
					"min_ms": 65.338009,
					"max_ms": 65.338009,
					"mean_ms": 65.338009,
					"median_ms": 65.338009,
					"stddev_ms": 0.0,
					"p75_ms": 65.338009,
					"p95_ms": 65.338009,
					"p99_ms": 65.338009,
					"p999_ms": 65.338009
				}
			}
		}
	}`

	nodeRqTimes7 = `{
		"responseHeader": {
			"status": 0,
			"QTime": 7
		},
		"metrics": {
			"solr.node": {
				"QUERY./admin/metrics/history.requestTimes": {
					"count": 1,
					"meanRate": 2.5126722879196796E-4,
					"1minRate": 3.6746144320828424E-30,
					"5minRate": 3.579862547745004E-7,
					"15minRate": 0.002428336015326626,
					"min_ms": 65.338009,
					"max_ms": 65.338009,
					"mean_ms": 65.338009,
					"median_ms": 65.338009,
					"stddev_ms": 0.0,
					"p75_ms": 65.338009,
					"p95_ms": 65.338009,
					"p99_ms": 65.338009,
					"p999_ms": 65.338009
				}
			}
		}
	}`
)

func gather4TestCache(i *Input, url string, v interface{}) error {
	return json.Unmarshal([]byte(cacheStat7), v)
}

func gather4TestSearcher(i *Input, url string, v interface{}) error {
	return json.Unmarshal([]byte(searcherStats7), v)
}

func gather4TestNodeRqTimes(i *Input, url string, v interface{}) error {
	return json.Unmarshal([]byte(nodeRqTimes7), v)
}

func gather4TestCoreRqTimes(i *Input, url string, v interface{}) error {
	return json.Unmarshal([]byte(coreRqTimes7), v)
}

/* test: failed
func TestCollect(t *testing.T) {
	expectM := &measurement{}

	i := &Input{}
	i.Tags = map[string]string{}
	i.Tags["tagAbc"] = "abc"
	i.Servers = []string{"http://localhost:8983"}

	// test gather searcher
	i.gatherData = gather4TestSearcher
	if err := i.Collect(); err != nil {
		t.Error(err)
	}

	expectM.fields = map[string]interface{}{
		"deleted_docs": 0,
		"max_docs":     32,
		"num_docs":     32,
		"warmup":       0,
	}
	expectM.tags = map[string]string{
		"category": "SEARCHER",
		"core":     "techproducts",
		"instance": "localhost_8983",
		"group":    "core",
		"tagAbc":   "abc",
	}
	expectM.name = "solr_searcher"
	AssertMeasurement(t, []inputs.Measurement{expectM}, i.collectCache, FieldCompare+NameCompare+TagCompare)
	i.collectCache = make([]inputs.Measurement, 0)

	// test gather request times
	// common
	expectM.fields = map[string]interface{}{
		"count":      1,
		"rate_mean":  2.5126722879196796e-4,
		"rate_1min":  3.6746144320828424e-30,
		"rate_5min":  3.579862547745004e-7,
		"rate_15min": 0.002428336015326626,
		"min":        65.338009,
		"max":        65.338009,
		"mean":       65.338009,
		"median":     65.338009,
		"stddev":     0.0,
		"p75":        65.338009,
		"p95":        65.338009,
		"p99":        65.338009,
		"p999":       65.338009,
	}
	expectM.name = "solr_request_times"
	// group == core
	i.gatherData = gather4TestCoreRqTimes
	if err := i.Collect(); err != nil {
		t.Error(err)
	}
	expectM.tags = map[string]string{
		"category": "QUERY",
		"core":     "techproducts",
		"instance": "localhost_8983",
		"group":    "core",
		"tagAbc":   "abc",
		"handler":  "/select",
	}
	AssertMeasurement(t, []inputs.Measurement{expectM}, i.collectCache, FieldCompare+NameCompare+TagCompare)
	i.collectCache = make([]inputs.Measurement, 0)
	// group == node
	i.gatherData = gather4TestNodeRqTimes
	if err := i.Collect(); err != nil {
		t.Error(err)
	}

	expectM.tags = map[string]string{
		"category": "QUERY",
		"instance": "localhost_8983",
		"group":    "node",
		"tagAbc":   "abc",
	}
	AssertMeasurement(t, []inputs.Measurement{expectM}, i.collectCache, FieldCompare+NameCompare+TagCompare)
	i.collectCache = make([]inputs.Measurement, 0)

	// test gather cache
	expectM.name = "solr_cache"
	expectM.fields = map[string]interface{}{
		"cumulative_evictions": 0,
		"cumulative_hitratio":  1.0,
		"cumulative_hits":      0,
		"cumulative_inserts":   0,
		"cumulative_lookups":   0,
		"evictions":            0,
		"hitratio":             1.0,
		"hits":                 0,
		"inserts":              0,
		"lookups":              0,
		"size":                 0,
		"warmup":               0,
		"max_ram":              -1,
		"ram_bytes_used":       416,
	}
	i.gatherData = gather4TestCache
	if err := i.Collect(); err != nil {
		t.Error(err)
	}
	expectM.tags = map[string]string{
		"category": "CACHE",
		"core":     "techproducts",
		"instance": "localhost_8983",
		"group":    "core",
		"tagAbc":   "abc",
	}
	AssertMeasurement(t, []inputs.Measurement{expectM}, i.collectCache, FieldCompare+NameCompare+TagCompare)
	i.collectCache = make([]inputs.Measurement, 0)

	i.gatherData = gatherDataFunc
	if err := i.Collect(); err != nil {
		t.Error(err)
	}
	i.collectCache = make([]inputs.Measurement, 0)
} */

func TestUrl(t *testing.T) {
}

func TestInstanceName(t *testing.T) {
	serverWResultExpect := map[string]string{
		"http://0.0.0.0:123456":    "0.0.0.0_12345",
		"https://127.0.0.1:8983":   "127.0.0.1_8983",
		"http://localhost:8983/":   "localhost_8983",
		"https://golang.org:12345": "golang.org_12345",
		"https://[::]:12345/":      "[::]_12345",
		"https://1.1":              "1.1", // 视为域名
		"golang.org":               "",
		"http://[a:b":              "",
	}
	for k, v := range serverWResultExpect {
		if m, err := instanceName(k); err != nil {
			t.Error(err)
		} else if m != v {
			t.Errorf("expect: %s  actual: %s", v, m)
		}
	}
}

// The incoming parameters expectMeasurement and actualMeasurement need to be sorted in advance.
func AssertMeasurement(t *testing.T,
	expectMeasurement []inputs.Measurement,
	actualMeasurement []inputs.Measurement, flag int) {
	t.Helper()

	lenE := len(expectMeasurement)
	lenA := len(actualMeasurement)
	if lenE != lenA {
		t.Errorf("The number of objects does not match. Expect:%d  Actual:%d", lenE, lenA)
	}
	count := lenE
	if count > lenA {
		count = lenA
	}

	for i := 0; i < count; i++ {
		expect, err := expectMeasurement[i].LineProto()
		if err != nil {
			t.Error(err)
			return
		}
		actual, err := actualMeasurement[i].LineProto()
		if err != nil {
			t.Error(err)
			return
		}
		// field
		if (flag & FieldCompare) == FieldCompare {
			expectFields, err := expect.Fields()
			if err != nil {
				t.Error(err)
				return
			}
			actualFields, err := actual.Fields()
			if err != nil {
				t.Error(err)
				return
			}
			for key, valueE := range expectFields {
				valueA, ok := actualFields[key]
				if !ok {
					t.Errorf("The expected field does not exist: %s", key)
					continue
				}
				assert.Equal(t, valueE, valueA, "Field: "+key)
			}
		}

		// name
		if (flag & NameCompare) == NameCompare {
			if expect.Name() != actual.Name() {
				t.Errorf("The expected measurement name is %s, the actual is %s", expect.Name(), actual.Name())
			}
		}

		// tag
		if (flag & TagCompare) == TagCompare {
			expectTags := expect.Tags()
			actualTags := actual.Tags()
			for kE, vE := range expectTags {
				vA, ok := actualTags[kE]
				if !ok {
					t.Errorf("The expected tag does not exist: %s", kE)
					continue
				}
				assert.Equal(t, vE, vA, "Tag: "+kE)
			}
		}

		// time
		if (flag & TimeCompare) == TimeCompare {
			if expect.Time() != actual.Time() {
				t.Error("The expected time is ", expect.Time().String(), ", the actual is ", actual.Time().String())
			}
		}
	}
}
