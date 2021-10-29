//nolint:lll
package solr

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
