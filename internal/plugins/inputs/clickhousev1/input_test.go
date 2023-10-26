// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package clickhousev1

import (
	"io/ioutil"
	"net/http"
	"sort"
	"strings"
	"testing"

	"github.com/BurntSushi/toml"
	"github.com/stretchr/testify/assert"
)

func TestCollect(t *testing.T) {
	testcases := []struct {
		name     string
		mockConf string
		mockData string
		want     []string
	}{
		{
			name: "with tag",
			mockConf: `urls = ["http://127.0.0.1:9363/metrics"]
` + mockCfg + `

[tags]
  tag1 = "some_value"
  tag2 = "some_other_value"`,
			mockData: mockBody_01,
			want:     want_01,
		},
		{
			name: "v22.8.15.23 and local",
			mockConf: `urls = ["http://localhost:9363/metrics"]
` + mockCfg,
			mockData: mockBody_v22_8_15_23,
			want:     want_v22_8_15_23,
		},
		{
			name: "v22.8.15.23 and local no election",
			mockConf: `urls = ["http://localhost:9363/metrics"]
election = false
` + mockCfg,
			mockData: mockBody_v22_8_15_23,
			want:     want_v22_8_15_23_no_election,
		},
		{
			name: "v21.8.15.7 and remote",
			mockConf: `urls = ["http://1.2.3.4:9363/metrics"]
` + mockCfg,
			mockData: mockBody_v21_8_15_7,
			want:     want_v21_8_15_7,
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			// init input
			ipt := NewProm()

			// ipt config
			_, err := toml.Decode(tc.mockConf, ipt)
			assert.NoError(t, err)
			// ipt.URLs = []string{mockURL}

			ipt.tagger = &mockTagger{}

			// init ipt.pm
			err = ipt.setup()
			assert.NoError(t, err)
			ipt.pm.SetClient(&http.Client{Transport: newTransportMock(tc.mockData)})

			// collect points
			pts, err := ipt.getPts()
			assert.NoError(t, err)
			_ = pts

			// check points
			left := make([]string, 0)
			for _, pt := range pts {
				str := pt.LineProto()
				// remove timestamp and float value
				str = str[:strings.LastIndex(str, "=")]
				left = append(left, str)
			}
			sort.Strings(left)
			sort.Strings(tc.want)

			if len(left) != len(tc.want) {
				t.Errorf("got len = %d want len = %d", len(left), len(tc.want))
			}
			for i := 0; i < len(left); i++ {
				if left[i] != tc.want[i] {
					t.Errorf("\ngot  = %v, \nwant = %v", left[i], tc.want[i])
				}
			}
		})
	}
}

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
	res.Body = ioutil.NopCloser(strings.NewReader(t.body))
	return res, nil
}

func (t *transportMock) CancelRequest(_ *http.Request) {}

func newTransportMock(body string) http.RoundTripper {
	return &transportMock{statusCode: http.StatusOK, body: body}
}

var mockBody_01 string = `
ClickHouseAsyncMetrics_BlockQueueTime_nvme0n1 0
ClickHouseAsyncMetrics_BlockQueueTime_sda 0.000018

ClickHouseAsyncMetrics_BlockReadBytes_nvme0n1 0
ClickHouseAsyncMetrics_BlockReadBytes_sda 151552

ClickHouseAsyncMetrics_CPUFrequencyMHz_0 1295.227
ClickHouseAsyncMetrics_CPUFrequencyMHz_1 2400
ClickHouseAsyncMetrics_CPUFrequencyMHz_2 2400
ClickHouseAsyncMetrics_CPUFrequencyMHz_3 2400
ClickHouseAsyncMetrics_CPUFrequencyMHz_4 2400
ClickHouseAsyncMetrics_CPUFrequencyMHz_5 2400
ClickHouseAsyncMetrics_CPUFrequencyMHz_6 2400
ClickHouseAsyncMetrics_CPUFrequencyMHz_7 2400

ClickHouseAsyncMetrics_CompiledExpressionCacheBytes 0
ClickHouseAsyncMetrics_CompiledExpressionCacheCount 0
ClickHouseAsyncMetrics_DiskAvailable_default 764679319552
ClickHouseAsyncMetrics_DiskTotal_default 982806224896
ClickHouseAsyncMetrics_DiskUnreserved_default 764679319552
ClickHouseAsyncMetrics_DiskUsed_default 218126905344
ClickHouseAsyncMetrics_FilesystemLogsPathAvailableBytes 764679319552
ClickHouseAsyncMetrics_FilesystemLogsPathAvailableINodes 59323832

ClickHouseAsyncMetrics_OSIOWaitTimeCPU0 0
ClickHouseAsyncMetrics_OSIOWaitTimeCPU1 0
ClickHouseAsyncMetrics_OSIOWaitTimeCPU2 0.009999060088351695
ClickHouseAsyncMetrics_OSIOWaitTimeCPU3 0
ClickHouseAsyncMetrics_OSIOWaitTimeCPU4 0.009999060088351695
ClickHouseAsyncMetrics_OSIOWaitTimeCPU5 0
ClickHouseAsyncMetrics_OSIOWaitTimeCPU6 0
ClickHouseAsyncMetrics_OSIOWaitTimeCPU7 0

ClickHouseAsyncMetrics_OSIdleTimeCPU0 0.9199135281283559
ClickHouseAsyncMetrics_OSIdleTimeCPU1 0.8099238671564872
ClickHouseAsyncMetrics_OSIdleTimeCPU2 0.8999154079516525
ClickHouseAsyncMetrics_OSIdleTimeCPU3 0.8999154079516525
ClickHouseAsyncMetrics_OSIdleTimeCPU4 0.8699182276865974
ClickHouseAsyncMetrics_OSIdleTimeCPU5 0.8599191675982457
ClickHouseAsyncMetrics_OSIdleTimeCPU6 0.8599191675982457
ClickHouseAsyncMetrics_OSIdleTimeCPU7 0.8699182276865974
`

var want_01 = []string{
	"ClickHouseAsyncMetrics,instance=127.0.0.1:9363,tag1=some_value,tag2=some_other_value,unit=nvme0n1 BlockQueueTime",
	"ClickHouseAsyncMetrics,cpu=1,instance=127.0.0.1:9363,tag1=some_value,tag2=some_other_value CPUFrequencyMHz",
	"ClickHouseAsyncMetrics,disk=default,instance=127.0.0.1:9363,tag1=some_value,tag2=some_other_value DiskAvailable",
	"ClickHouseAsyncMetrics,disk=default,instance=127.0.0.1:9363,tag1=some_value,tag2=some_other_value DiskTotal",
	"ClickHouseAsyncMetrics,cpu=0,instance=127.0.0.1:9363,tag1=some_value,tag2=some_other_value OSIOWaitTimeCPU",
	"ClickHouseAsyncMetrics,cpu=6,instance=127.0.0.1:9363,tag1=some_value,tag2=some_other_value OSIOWaitTimeCPU",
	"ClickHouseAsyncMetrics,cpu=1,instance=127.0.0.1:9363,tag1=some_value,tag2=some_other_value OSIdleTimeCPU",
	"ClickHouseAsyncMetrics,cpu=2,instance=127.0.0.1:9363,tag1=some_value,tag2=some_other_value CPUFrequencyMHz",
	"ClickHouseAsyncMetrics,cpu=7,instance=127.0.0.1:9363,tag1=some_value,tag2=some_other_value CPUFrequencyMHz",
	"ClickHouseAsyncMetrics,instance=127.0.0.1:9363,tag1=some_value,tag2=some_other_value FilesystemLogsPathAvailableBytes",
	"ClickHouseAsyncMetrics,instance=127.0.0.1:9363,tag1=some_value,tag2=some_other_value FilesystemLogsPathAvailableINodes",
	"ClickHouseAsyncMetrics,cpu=4,instance=127.0.0.1:9363,tag1=some_value,tag2=some_other_value OSIdleTimeCPU",
	"ClickHouseAsyncMetrics,cpu=6,instance=127.0.0.1:9363,tag1=some_value,tag2=some_other_value OSIdleTimeCPU",
	"ClickHouseAsyncMetrics,disk=default,instance=127.0.0.1:9363,tag1=some_value,tag2=some_other_value DiskUnreserved",
	"ClickHouseAsyncMetrics,cpu=2,instance=127.0.0.1:9363,tag1=some_value,tag2=some_other_value OSIOWaitTimeCPU",
	"ClickHouseAsyncMetrics,cpu=0,instance=127.0.0.1:9363,tag1=some_value,tag2=some_other_value OSIdleTimeCPU",
	"ClickHouseAsyncMetrics,instance=127.0.0.1:9363,tag1=some_value,tag2=some_other_value,unit=sda BlockQueueTime",
	"ClickHouseAsyncMetrics,instance=127.0.0.1:9363,tag1=some_value,tag2=some_other_value,unit=sda BlockReadBytes",
	"ClickHouseAsyncMetrics,cpu=1,instance=127.0.0.1:9363,tag1=some_value,tag2=some_other_value OSIOWaitTimeCPU",
	"ClickHouseAsyncMetrics,cpu=5,instance=127.0.0.1:9363,tag1=some_value,tag2=some_other_value OSIOWaitTimeCPU",
	"ClickHouseAsyncMetrics,cpu=6,instance=127.0.0.1:9363,tag1=some_value,tag2=some_other_value CPUFrequencyMHz",
	"ClickHouseAsyncMetrics,instance=127.0.0.1:9363,tag1=some_value,tag2=some_other_value CompiledExpressionCacheBytes",
	"ClickHouseAsyncMetrics,disk=default,instance=127.0.0.1:9363,tag1=some_value,tag2=some_other_value DiskUsed",
	"ClickHouseAsyncMetrics,cpu=3,instance=127.0.0.1:9363,tag1=some_value,tag2=some_other_value OSIOWaitTimeCPU",
	"ClickHouseAsyncMetrics,cpu=2,instance=127.0.0.1:9363,tag1=some_value,tag2=some_other_value OSIdleTimeCPU",
	"ClickHouseAsyncMetrics,cpu=3,instance=127.0.0.1:9363,tag1=some_value,tag2=some_other_value OSIdleTimeCPU",
	"ClickHouseAsyncMetrics,cpu=5,instance=127.0.0.1:9363,tag1=some_value,tag2=some_other_value CPUFrequencyMHz",
	"ClickHouseAsyncMetrics,instance=127.0.0.1:9363,tag1=some_value,tag2=some_other_value CompiledExpressionCacheCount",
	"ClickHouseAsyncMetrics,cpu=0,instance=127.0.0.1:9363,tag1=some_value,tag2=some_other_value CPUFrequencyMHz",
	"ClickHouseAsyncMetrics,cpu=3,instance=127.0.0.1:9363,tag1=some_value,tag2=some_other_value CPUFrequencyMHz",
	"ClickHouseAsyncMetrics,cpu=4,instance=127.0.0.1:9363,tag1=some_value,tag2=some_other_value CPUFrequencyMHz",
	"ClickHouseAsyncMetrics,cpu=7,instance=127.0.0.1:9363,tag1=some_value,tag2=some_other_value OSIOWaitTimeCPU",
	"ClickHouseAsyncMetrics,instance=127.0.0.1:9363,tag1=some_value,tag2=some_other_value,unit=nvme0n1 BlockReadBytes",
	"ClickHouseAsyncMetrics,cpu=4,instance=127.0.0.1:9363,tag1=some_value,tag2=some_other_value OSIOWaitTimeCPU",
	"ClickHouseAsyncMetrics,cpu=5,instance=127.0.0.1:9363,tag1=some_value,tag2=some_other_value OSIdleTimeCPU",
	"ClickHouseAsyncMetrics,cpu=7,instance=127.0.0.1:9363,tag1=some_value,tag2=some_other_value OSIdleTimeCPU",
	"ClickHouseAsyncMetrics,instance=127.0.0.1:9363,tag1=some_value,tag2=some_other_value,unit=average BlockQueueTime",
	"ClickHouseAsyncMetrics,instance=127.0.0.1:9363,tag1=some_value,tag2=some_other_value,unit=total BlockReadBytes",
	"ClickHouseAsyncMetrics,cpu=average,instance=127.0.0.1:9363,tag1=some_value,tag2=some_other_value CPUFrequencyMHz",
	"ClickHouseAsyncMetrics,disk=total,instance=127.0.0.1:9363,tag1=some_value,tag2=some_other_value DiskAvailable",
	"ClickHouseAsyncMetrics,disk=total,instance=127.0.0.1:9363,tag1=some_value,tag2=some_other_value DiskTotal",
	"ClickHouseAsyncMetrics,disk=total,instance=127.0.0.1:9363,tag1=some_value,tag2=some_other_value DiskUnreserved",
	"ClickHouseAsyncMetrics,disk=total,instance=127.0.0.1:9363,tag1=some_value,tag2=some_other_value DiskUsed",
	"ClickHouseAsyncMetrics,cpu=average,instance=127.0.0.1:9363,tag1=some_value,tag2=some_other_value OSIOWaitTimeCPU",
	"ClickHouseAsyncMetrics,cpu=average,instance=127.0.0.1:9363,tag1=some_value,tag2=some_other_value OSIdleTimeCPU",
}

var mockBody_v22_8_15_23 string = `
# HELP ClickHouseProfileEvents_Query Number of queries to be interpreted and potentially executed. Does not include queries that failed to parse or were rejected due to AST size limits, quota limits or limits on the number of simultaneously running queries. May include internal queries initiated by ClickHouse itself. Does not count subqueries.
# TYPE ClickHouseProfileEvents_Query counter
ClickHouseProfileEvents_Query 0
# HELP ClickHouseProfileEvents_SelectQuery Same as Query, but only for SELECT queries.
# TYPE ClickHouseProfileEvents_SelectQuery counter
ClickHouseProfileEvents_SelectQuery 0
# HELP ClickHouseProfileEvents_InsertQuery Same as Query, but only for INSERT queries.
# TYPE ClickHouseProfileEvents_InsertQuery counter
ClickHouseProfileEvents_InsertQuery 0
# HELP ClickHouseProfileEvents_AsyncInsertQuery Same as InsertQuery, but only for asynchronous INSERT queries.
# TYPE ClickHouseProfileEvents_AsyncInsertQuery counter
ClickHouseProfileEvents_AsyncInsertQuery 0
# HELP ClickHouseProfileEvents_AsyncInsertBytes Data size in bytes of asynchronous INSERT queries.
# TYPE ClickHouseProfileEvents_AsyncInsertBytes counter
ClickHouseProfileEvents_AsyncInsertBytes 0
# HELP ClickHouseProfileEvents_FailedQuery Number of failed queries.
# TYPE ClickHouseProfileEvents_FailedQuery counter
ClickHouseProfileEvents_FailedQuery 0
# HELP ClickHouseProfileEvents_FailedSelectQuery Same as FailedQuery, but only for SELECT queries.
# TYPE ClickHouseProfileEvents_FailedSelectQuery counter
ClickHouseProfileEvents_FailedSelectQuery 0
# HELP ClickHouseProfileEvents_FailedInsertQuery Same as FailedQuery, but only for INSERT queries.
# TYPE ClickHouseProfileEvents_FailedInsertQuery counter
ClickHouseProfileEvents_FailedInsertQuery 0
# HELP ClickHouseProfileEvents_QueryTimeMicroseconds Total time of all queries.
# TYPE ClickHouseProfileEvents_QueryTimeMicroseconds counter
ClickHouseProfileEvents_QueryTimeMicroseconds 0
# HELP ClickHouseProfileEvents_SelectQueryTimeMicroseconds Total time of SELECT queries.
# TYPE ClickHouseProfileEvents_SelectQueryTimeMicroseconds counter
ClickHouseProfileEvents_SelectQueryTimeMicroseconds 0
# HELP ClickHouseProfileEvents_InsertQueryTimeMicroseconds Total time of INSERT queries.
# TYPE ClickHouseProfileEvents_InsertQueryTimeMicroseconds counter
ClickHouseProfileEvents_InsertQueryTimeMicroseconds 0
# HELP ClickHouseProfileEvents_OtherQueryTimeMicroseconds Total time of queries that are not SELECT or INSERT.
# TYPE ClickHouseProfileEvents_OtherQueryTimeMicroseconds counter
ClickHouseProfileEvents_OtherQueryTimeMicroseconds 0
# HELP ClickHouseProfileEvents_FileOpen Number of files opened.
# TYPE ClickHouseProfileEvents_FileOpen counter
ClickHouseProfileEvents_FileOpen 2129
# HELP ClickHouseProfileEvents_Seek Number of times the 'lseek' function was called.
# TYPE ClickHouseProfileEvents_Seek counter
ClickHouseProfileEvents_Seek 0
# HELP ClickHouseProfileEvents_ReadBufferFromFileDescriptorRead Number of reads (read/pread) from a file descriptor. Does not include sockets.
# TYPE ClickHouseProfileEvents_ReadBufferFromFileDescriptorRead counter
ClickHouseProfileEvents_ReadBufferFromFileDescriptorRead 3669
# HELP ClickHouseProfileEvents_ReadBufferFromFileDescriptorReadFailed Number of times the read (read/pread) from a file descriptor have failed.
# TYPE ClickHouseProfileEvents_ReadBufferFromFileDescriptorReadFailed counter
ClickHouseProfileEvents_ReadBufferFromFileDescriptorReadFailed 0
# HELP ClickHouseProfileEvents_ReadBufferFromFileDescriptorReadBytes Number of bytes read from file descriptors. If the file is compressed, this will show the compressed data size.
# TYPE ClickHouseProfileEvents_ReadBufferFromFileDescriptorReadBytes counter
ClickHouseProfileEvents_ReadBufferFromFileDescriptorReadBytes 5279674
# HELP ClickHouseProfileEvents_WriteBufferFromFileDescriptorWrite Number of writes (write/pwrite) to a file descriptor. Does not include sockets.
# TYPE ClickHouseProfileEvents_WriteBufferFromFileDescriptorWrite counter
ClickHouseProfileEvents_WriteBufferFromFileDescriptorWrite 1006
# HELP ClickHouseProfileEvents_WriteBufferFromFileDescriptorWriteFailed Number of times the write (write/pwrite) to a file descriptor have failed.
# TYPE ClickHouseProfileEvents_WriteBufferFromFileDescriptorWriteFailed counter
ClickHouseProfileEvents_WriteBufferFromFileDescriptorWriteFailed 0
# HELP ClickHouseProfileEvents_WriteBufferFromFileDescriptorWriteBytes Number of bytes written to file descriptors. If the file is compressed, this will show compressed data size.
# TYPE ClickHouseProfileEvents_WriteBufferFromFileDescriptorWriteBytes counter
ClickHouseProfileEvents_WriteBufferFromFileDescriptorWriteBytes 4327830
# HELP ClickHouseProfileEvents_FileSync Number of times the F_FULLFSYNC/fsync/fdatasync function was called for files.
# TYPE ClickHouseProfileEvents_FileSync counter
ClickHouseProfileEvents_FileSync 0
# HELP ClickHouseProfileEvents_DirectorySync Number of times the F_FULLFSYNC/fsync/fdatasync function was called for directories.
# TYPE ClickHouseProfileEvents_DirectorySync counter
ClickHouseProfileEvents_DirectorySync 0
# HELP ClickHouseProfileEvents_FileSyncElapsedMicroseconds Total time spent waiting for F_FULLFSYNC/fsync/fdatasync syscall for files.
# TYPE ClickHouseProfileEvents_FileSyncElapsedMicroseconds counter
ClickHouseProfileEvents_FileSyncElapsedMicroseconds 0
# HELP ClickHouseProfileEvents_DirectorySyncElapsedMicroseconds Total time spent waiting for F_FULLFSYNC/fsync/fdatasync syscall for directories.
# TYPE ClickHouseProfileEvents_DirectorySyncElapsedMicroseconds counter
ClickHouseProfileEvents_DirectorySyncElapsedMicroseconds 0
# HELP ClickHouseProfileEvents_ReadCompressedBytes Number of bytes (the number of bytes before decompression) read from compressed sources (files, network).
# TYPE ClickHouseProfileEvents_ReadCompressedBytes counter
ClickHouseProfileEvents_ReadCompressedBytes 4110878
# HELP ClickHouseProfileEvents_CompressedReadBufferBlocks Number of compressed blocks (the blocks of data that are compressed independent of each other) read from compressed sources (files, network).
# TYPE ClickHouseProfileEvents_CompressedReadBufferBlocks counter
ClickHouseProfileEvents_CompressedReadBufferBlocks 2979
# HELP ClickHouseProfileEvents_CompressedReadBufferBytes Number of uncompressed bytes (the number of bytes after decompression) read from compressed sources (files, network).
# TYPE ClickHouseProfileEvents_CompressedReadBufferBytes counter
ClickHouseProfileEvents_CompressedReadBufferBytes 115512558
# HELP ClickHouseProfileEvents_UncompressedCacheHits 
# TYPE ClickHouseProfileEvents_UncompressedCacheHits counter
ClickHouseProfileEvents_UncompressedCacheHits 0
# HELP ClickHouseProfileEvents_UncompressedCacheMisses 
# TYPE ClickHouseProfileEvents_UncompressedCacheMisses counter
ClickHouseProfileEvents_UncompressedCacheMisses 0
# HELP ClickHouseProfileEvents_UncompressedCacheWeightLost 
# TYPE ClickHouseProfileEvents_UncompressedCacheWeightLost counter
ClickHouseProfileEvents_UncompressedCacheWeightLost 0
# HELP ClickHouseProfileEvents_MMappedFileCacheHits 
# TYPE ClickHouseProfileEvents_MMappedFileCacheHits counter
ClickHouseProfileEvents_MMappedFileCacheHits 0
# HELP ClickHouseProfileEvents_MMappedFileCacheMisses 
# TYPE ClickHouseProfileEvents_MMappedFileCacheMisses counter
ClickHouseProfileEvents_MMappedFileCacheMisses 0
# HELP ClickHouseProfileEvents_OpenedFileCacheHits 
# TYPE ClickHouseProfileEvents_OpenedFileCacheHits counter
ClickHouseProfileEvents_OpenedFileCacheHits 0
# HELP ClickHouseProfileEvents_OpenedFileCacheMisses 
# TYPE ClickHouseProfileEvents_OpenedFileCacheMisses counter
ClickHouseProfileEvents_OpenedFileCacheMisses 1097
# HELP ClickHouseProfileEvents_AIOWrite Number of writes with Linux or FreeBSD AIO interface
# TYPE ClickHouseProfileEvents_AIOWrite counter
ClickHouseProfileEvents_AIOWrite 0
# HELP ClickHouseProfileEvents_AIOWriteBytes Number of bytes written with Linux or FreeBSD AIO interface
# TYPE ClickHouseProfileEvents_AIOWriteBytes counter
ClickHouseProfileEvents_AIOWriteBytes 0
# HELP ClickHouseProfileEvents_AIORead Number of reads with Linux or FreeBSD AIO interface
# TYPE ClickHouseProfileEvents_AIORead counter
ClickHouseProfileEvents_AIORead 0
# HELP ClickHouseProfileEvents_AIOReadBytes Number of bytes read with Linux or FreeBSD AIO interface
# TYPE ClickHouseProfileEvents_AIOReadBytes counter
ClickHouseProfileEvents_AIOReadBytes 0
# HELP ClickHouseProfileEvents_IOBufferAllocs 
# TYPE ClickHouseProfileEvents_IOBufferAllocs counter
ClickHouseProfileEvents_IOBufferAllocs 4452
# HELP ClickHouseProfileEvents_IOBufferAllocBytes 
# TYPE ClickHouseProfileEvents_IOBufferAllocBytes counter
ClickHouseProfileEvents_IOBufferAllocBytes 1038979410
# HELP ClickHouseProfileEvents_ArenaAllocChunks 
# TYPE ClickHouseProfileEvents_ArenaAllocChunks counter
ClickHouseProfileEvents_ArenaAllocChunks 0
# HELP ClickHouseProfileEvents_ArenaAllocBytes 
# TYPE ClickHouseProfileEvents_ArenaAllocBytes counter
ClickHouseProfileEvents_ArenaAllocBytes 0
# HELP ClickHouseProfileEvents_FunctionExecute 
# TYPE ClickHouseProfileEvents_FunctionExecute counter
ClickHouseProfileEvents_FunctionExecute 1277
# HELP ClickHouseProfileEvents_TableFunctionExecute 
# TYPE ClickHouseProfileEvents_TableFunctionExecute counter
ClickHouseProfileEvents_TableFunctionExecute 0
# HELP ClickHouseProfileEvents_MarkCacheHits 
# TYPE ClickHouseProfileEvents_MarkCacheHits counter
ClickHouseProfileEvents_MarkCacheHits 0
# HELP ClickHouseProfileEvents_MarkCacheMisses 
# TYPE ClickHouseProfileEvents_MarkCacheMisses counter
ClickHouseProfileEvents_MarkCacheMisses 0
# HELP ClickHouseProfileEvents_CreatedReadBufferOrdinary 
# TYPE ClickHouseProfileEvents_CreatedReadBufferOrdinary counter
ClickHouseProfileEvents_CreatedReadBufferOrdinary 1097
# HELP ClickHouseProfileEvents_CreatedReadBufferDirectIO 
# TYPE ClickHouseProfileEvents_CreatedReadBufferDirectIO counter
ClickHouseProfileEvents_CreatedReadBufferDirectIO 0
# HELP ClickHouseProfileEvents_CreatedReadBufferDirectIOFailed 
# TYPE ClickHouseProfileEvents_CreatedReadBufferDirectIOFailed counter
ClickHouseProfileEvents_CreatedReadBufferDirectIOFailed 0
# HELP ClickHouseProfileEvents_CreatedReadBufferMMap 
# TYPE ClickHouseProfileEvents_CreatedReadBufferMMap counter
ClickHouseProfileEvents_CreatedReadBufferMMap 0
# HELP ClickHouseProfileEvents_CreatedReadBufferMMapFailed 
# TYPE ClickHouseProfileEvents_CreatedReadBufferMMapFailed counter
ClickHouseProfileEvents_CreatedReadBufferMMapFailed 0
# HELP ClickHouseProfileEvents_DiskReadElapsedMicroseconds Total time spent waiting for read syscall. This include reads from page cache.
# TYPE ClickHouseProfileEvents_DiskReadElapsedMicroseconds counter
ClickHouseProfileEvents_DiskReadElapsedMicroseconds 1392783
# HELP ClickHouseProfileEvents_DiskWriteElapsedMicroseconds Total time spent waiting for write syscall. This include writes to page cache.
# TYPE ClickHouseProfileEvents_DiskWriteElapsedMicroseconds counter
ClickHouseProfileEvents_DiskWriteElapsedMicroseconds 5993
# HELP ClickHouseProfileEvents_NetworkReceiveElapsedMicroseconds Total time spent waiting for data to receive or receiving data from network. Only ClickHouse-related network interaction is included, not by 3rd party libraries.
# TYPE ClickHouseProfileEvents_NetworkReceiveElapsedMicroseconds counter
ClickHouseProfileEvents_NetworkReceiveElapsedMicroseconds 8
# HELP ClickHouseProfileEvents_NetworkSendElapsedMicroseconds Total time spent waiting for data to send to network or sending data to network. Only ClickHouse-related network interaction is included, not by 3rd party libraries..
# TYPE ClickHouseProfileEvents_NetworkSendElapsedMicroseconds counter
ClickHouseProfileEvents_NetworkSendElapsedMicroseconds 0
# HELP ClickHouseProfileEvents_NetworkReceiveBytes Total number of bytes received from network. Only ClickHouse-related network interaction is included, not by 3rd party libraries.
# TYPE ClickHouseProfileEvents_NetworkReceiveBytes counter
ClickHouseProfileEvents_NetworkReceiveBytes 1356
# HELP ClickHouseProfileEvents_NetworkSendBytes Total number of bytes send to network. Only ClickHouse-related network interaction is included, not by 3rd party libraries.
# TYPE ClickHouseProfileEvents_NetworkSendBytes counter
ClickHouseProfileEvents_NetworkSendBytes 0
# HELP ClickHouseProfileEvents_ThrottlerSleepMicroseconds Total time a query was sleeping to conform the 'max_network_bandwidth' setting.
# TYPE ClickHouseProfileEvents_ThrottlerSleepMicroseconds counter
ClickHouseProfileEvents_ThrottlerSleepMicroseconds 0
# HELP ClickHouseProfileEvents_QueryMaskingRulesMatch Number of times query masking rules was successfully matched.
# TYPE ClickHouseProfileEvents_QueryMaskingRulesMatch counter
ClickHouseProfileEvents_QueryMaskingRulesMatch 0
# HELP ClickHouseProfileEvents_ReplicatedPartFetches Number of times a data part was downloaded from replica of a ReplicatedMergeTree table.
# TYPE ClickHouseProfileEvents_ReplicatedPartFetches counter
ClickHouseProfileEvents_ReplicatedPartFetches 0
# HELP ClickHouseProfileEvents_ReplicatedPartFailedFetches Number of times a data part was failed to download from replica of a ReplicatedMergeTree table.
# TYPE ClickHouseProfileEvents_ReplicatedPartFailedFetches counter
ClickHouseProfileEvents_ReplicatedPartFailedFetches 0
# HELP ClickHouseProfileEvents_ObsoleteReplicatedParts 
# TYPE ClickHouseProfileEvents_ObsoleteReplicatedParts counter
ClickHouseProfileEvents_ObsoleteReplicatedParts 0
# HELP ClickHouseProfileEvents_ReplicatedPartMerges Number of times data parts of ReplicatedMergeTree tables were successfully merged.
# TYPE ClickHouseProfileEvents_ReplicatedPartMerges counter
ClickHouseProfileEvents_ReplicatedPartMerges 0
# HELP ClickHouseProfileEvents_ReplicatedPartFetchesOfMerged Number of times we prefer to download already merged part from replica of ReplicatedMergeTree table instead of performing a merge ourself (usually we prefer doing a merge ourself to save network traffic). This happens when we have not all source parts to perform a merge or when the data part is old enough.
# TYPE ClickHouseProfileEvents_ReplicatedPartFetchesOfMerged counter
ClickHouseProfileEvents_ReplicatedPartFetchesOfMerged 0
# HELP ClickHouseProfileEvents_ReplicatedPartMutations 
# TYPE ClickHouseProfileEvents_ReplicatedPartMutations counter
ClickHouseProfileEvents_ReplicatedPartMutations 0
# HELP ClickHouseProfileEvents_ReplicatedPartChecks 
# TYPE ClickHouseProfileEvents_ReplicatedPartChecks counter
ClickHouseProfileEvents_ReplicatedPartChecks 0
# HELP ClickHouseProfileEvents_ReplicatedPartChecksFailed 
# TYPE ClickHouseProfileEvents_ReplicatedPartChecksFailed counter
ClickHouseProfileEvents_ReplicatedPartChecksFailed 0
# HELP ClickHouseProfileEvents_ReplicatedDataLoss Number of times a data part that we wanted doesn't exist on any replica (even on replicas that are offline right now). That data parts are definitely lost. This is normal due to asynchronous replication (if quorum inserts were not enabled), when the replica on which the data part was written was failed and when it became online after fail it doesn't contain that data part.
# TYPE ClickHouseProfileEvents_ReplicatedDataLoss counter
ClickHouseProfileEvents_ReplicatedDataLoss 0
# HELP ClickHouseProfileEvents_InsertedRows Number of rows INSERTed to all tables.
# TYPE ClickHouseProfileEvents_InsertedRows counter
ClickHouseProfileEvents_InsertedRows 6263
# HELP ClickHouseProfileEvents_InsertedBytes Number of bytes (uncompressed; for columns as they stored in memory) INSERTed to all tables.
# TYPE ClickHouseProfileEvents_InsertedBytes counter
ClickHouseProfileEvents_InsertedBytes 455089
# HELP ClickHouseProfileEvents_DelayedInserts Number of times the INSERT of a block to a MergeTree table was throttled due to high number of active data parts for partition.
# TYPE ClickHouseProfileEvents_DelayedInserts counter
ClickHouseProfileEvents_DelayedInserts 0
# HELP ClickHouseProfileEvents_RejectedInserts Number of times the INSERT of a block to a MergeTree table was rejected with 'Too many parts' exception due to high number of active data parts for partition.
# TYPE ClickHouseProfileEvents_RejectedInserts counter
ClickHouseProfileEvents_RejectedInserts 0
# HELP ClickHouseProfileEvents_DelayedInsertsMilliseconds Total number of milliseconds spent while the INSERT of a block to a MergeTree table was throttled due to high number of active data parts for partition.
# TYPE ClickHouseProfileEvents_DelayedInsertsMilliseconds counter
ClickHouseProfileEvents_DelayedInsertsMilliseconds 0
# HELP ClickHouseProfileEvents_DistributedDelayedInserts Number of times the INSERT of a block to a Distributed table was throttled due to high number of pending bytes.
# TYPE ClickHouseProfileEvents_DistributedDelayedInserts counter
ClickHouseProfileEvents_DistributedDelayedInserts 0
# HELP ClickHouseProfileEvents_DistributedRejectedInserts Number of times the INSERT of a block to a Distributed table was rejected with 'Too many bytes' exception due to high number of pending bytes.
# TYPE ClickHouseProfileEvents_DistributedRejectedInserts counter
ClickHouseProfileEvents_DistributedRejectedInserts 0
# HELP ClickHouseProfileEvents_DistributedDelayedInsertsMilliseconds Total number of milliseconds spent while the INSERT of a block to a Distributed table was throttled due to high number of pending bytes.
# TYPE ClickHouseProfileEvents_DistributedDelayedInsertsMilliseconds counter
ClickHouseProfileEvents_DistributedDelayedInsertsMilliseconds 0
# HELP ClickHouseProfileEvents_DuplicatedInsertedBlocks Number of times the INSERTed block to a ReplicatedMergeTree table was deduplicated.
# TYPE ClickHouseProfileEvents_DuplicatedInsertedBlocks counter
ClickHouseProfileEvents_DuplicatedInsertedBlocks 0
# HELP ClickHouseProfileEvents_ZooKeeperInit 
# TYPE ClickHouseProfileEvents_ZooKeeperInit counter
ClickHouseProfileEvents_ZooKeeperInit 0
# HELP ClickHouseProfileEvents_ZooKeeperTransactions 
# TYPE ClickHouseProfileEvents_ZooKeeperTransactions counter
ClickHouseProfileEvents_ZooKeeperTransactions 0
# HELP ClickHouseProfileEvents_ZooKeeperList 
# TYPE ClickHouseProfileEvents_ZooKeeperList counter
ClickHouseProfileEvents_ZooKeeperList 0
# HELP ClickHouseProfileEvents_ZooKeeperCreate 
# TYPE ClickHouseProfileEvents_ZooKeeperCreate counter
ClickHouseProfileEvents_ZooKeeperCreate 0
# HELP ClickHouseProfileEvents_ZooKeeperRemove 
# TYPE ClickHouseProfileEvents_ZooKeeperRemove counter
ClickHouseProfileEvents_ZooKeeperRemove 0
# HELP ClickHouseProfileEvents_ZooKeeperExists 
# TYPE ClickHouseProfileEvents_ZooKeeperExists counter
ClickHouseProfileEvents_ZooKeeperExists 0
# HELP ClickHouseProfileEvents_ZooKeeperGet 
# TYPE ClickHouseProfileEvents_ZooKeeperGet counter
ClickHouseProfileEvents_ZooKeeperGet 0
# HELP ClickHouseProfileEvents_ZooKeeperSet 
# TYPE ClickHouseProfileEvents_ZooKeeperSet counter
ClickHouseProfileEvents_ZooKeeperSet 0
# HELP ClickHouseProfileEvents_ZooKeeperMulti 
# TYPE ClickHouseProfileEvents_ZooKeeperMulti counter
ClickHouseProfileEvents_ZooKeeperMulti 0
# HELP ClickHouseProfileEvents_ZooKeeperCheck 
# TYPE ClickHouseProfileEvents_ZooKeeperCheck counter
ClickHouseProfileEvents_ZooKeeperCheck 0
# HELP ClickHouseProfileEvents_ZooKeeperSync 
# TYPE ClickHouseProfileEvents_ZooKeeperSync counter
ClickHouseProfileEvents_ZooKeeperSync 0
# HELP ClickHouseProfileEvents_ZooKeeperClose 
# TYPE ClickHouseProfileEvents_ZooKeeperClose counter
ClickHouseProfileEvents_ZooKeeperClose 0
# HELP ClickHouseProfileEvents_ZooKeeperWatchResponse 
# TYPE ClickHouseProfileEvents_ZooKeeperWatchResponse counter
ClickHouseProfileEvents_ZooKeeperWatchResponse 0
# HELP ClickHouseProfileEvents_ZooKeeperUserExceptions 
# TYPE ClickHouseProfileEvents_ZooKeeperUserExceptions counter
ClickHouseProfileEvents_ZooKeeperUserExceptions 0
# HELP ClickHouseProfileEvents_ZooKeeperHardwareExceptions 
# TYPE ClickHouseProfileEvents_ZooKeeperHardwareExceptions counter
ClickHouseProfileEvents_ZooKeeperHardwareExceptions 0
# HELP ClickHouseProfileEvents_ZooKeeperOtherExceptions 
# TYPE ClickHouseProfileEvents_ZooKeeperOtherExceptions counter
ClickHouseProfileEvents_ZooKeeperOtherExceptions 0
# HELP ClickHouseProfileEvents_ZooKeeperWaitMicroseconds 
# TYPE ClickHouseProfileEvents_ZooKeeperWaitMicroseconds counter
ClickHouseProfileEvents_ZooKeeperWaitMicroseconds 0
# HELP ClickHouseProfileEvents_ZooKeeperBytesSent 
# TYPE ClickHouseProfileEvents_ZooKeeperBytesSent counter
ClickHouseProfileEvents_ZooKeeperBytesSent 0
# HELP ClickHouseProfileEvents_ZooKeeperBytesReceived 
# TYPE ClickHouseProfileEvents_ZooKeeperBytesReceived counter
ClickHouseProfileEvents_ZooKeeperBytesReceived 0
# HELP ClickHouseProfileEvents_DistributedConnectionFailTry Total count when distributed connection fails with retry
# TYPE ClickHouseProfileEvents_DistributedConnectionFailTry counter
ClickHouseProfileEvents_DistributedConnectionFailTry 0
# HELP ClickHouseProfileEvents_DistributedConnectionMissingTable 
# TYPE ClickHouseProfileEvents_DistributedConnectionMissingTable counter
ClickHouseProfileEvents_DistributedConnectionMissingTable 0
# HELP ClickHouseProfileEvents_DistributedConnectionStaleReplica 
# TYPE ClickHouseProfileEvents_DistributedConnectionStaleReplica counter
ClickHouseProfileEvents_DistributedConnectionStaleReplica 0
# HELP ClickHouseProfileEvents_DistributedConnectionFailAtAll Total count when distributed connection fails after all retries finished
# TYPE ClickHouseProfileEvents_DistributedConnectionFailAtAll counter
ClickHouseProfileEvents_DistributedConnectionFailAtAll 0
# HELP ClickHouseProfileEvents_HedgedRequestsChangeReplica Total count when timeout for changing replica expired in hedged requests.
# TYPE ClickHouseProfileEvents_HedgedRequestsChangeReplica counter
ClickHouseProfileEvents_HedgedRequestsChangeReplica 0
# HELP ClickHouseProfileEvents_CompileFunction Number of times a compilation of generated LLVM code (to create fused function for complex expressions) was initiated.
# TYPE ClickHouseProfileEvents_CompileFunction counter
ClickHouseProfileEvents_CompileFunction 0
# HELP ClickHouseProfileEvents_CompiledFunctionExecute Number of times a compiled function was executed.
# TYPE ClickHouseProfileEvents_CompiledFunctionExecute counter
ClickHouseProfileEvents_CompiledFunctionExecute 0
# HELP ClickHouseProfileEvents_CompileExpressionsMicroseconds Total time spent for compilation of expressions to LLVM code.
# TYPE ClickHouseProfileEvents_CompileExpressionsMicroseconds counter
ClickHouseProfileEvents_CompileExpressionsMicroseconds 0
# HELP ClickHouseProfileEvents_CompileExpressionsBytes Number of bytes used for expressions compilation.
# TYPE ClickHouseProfileEvents_CompileExpressionsBytes counter
ClickHouseProfileEvents_CompileExpressionsBytes 0
# HELP ClickHouseProfileEvents_ExecuteShellCommand Number of shell command executions.
# TYPE ClickHouseProfileEvents_ExecuteShellCommand counter
ClickHouseProfileEvents_ExecuteShellCommand 0
# HELP ClickHouseProfileEvents_ExternalSortWritePart 
# TYPE ClickHouseProfileEvents_ExternalSortWritePart counter
ClickHouseProfileEvents_ExternalSortWritePart 0
# HELP ClickHouseProfileEvents_ExternalSortMerge 
# TYPE ClickHouseProfileEvents_ExternalSortMerge counter
ClickHouseProfileEvents_ExternalSortMerge 0
# HELP ClickHouseProfileEvents_ExternalAggregationWritePart 
# TYPE ClickHouseProfileEvents_ExternalAggregationWritePart counter
ClickHouseProfileEvents_ExternalAggregationWritePart 0
# HELP ClickHouseProfileEvents_ExternalAggregationMerge 
# TYPE ClickHouseProfileEvents_ExternalAggregationMerge counter
ClickHouseProfileEvents_ExternalAggregationMerge 0
# HELP ClickHouseProfileEvents_ExternalAggregationCompressedBytes 
# TYPE ClickHouseProfileEvents_ExternalAggregationCompressedBytes counter
ClickHouseProfileEvents_ExternalAggregationCompressedBytes 0
# HELP ClickHouseProfileEvents_ExternalAggregationUncompressedBytes 
# TYPE ClickHouseProfileEvents_ExternalAggregationUncompressedBytes counter
ClickHouseProfileEvents_ExternalAggregationUncompressedBytes 0
# HELP ClickHouseProfileEvents_SlowRead Number of reads from a file that were slow. This indicate system overload. Thresholds are controlled by read_backoff_* settings.
# TYPE ClickHouseProfileEvents_SlowRead counter
ClickHouseProfileEvents_SlowRead 0
# HELP ClickHouseProfileEvents_ReadBackoff Number of times the number of query processing threads was lowered due to slow reads.
# TYPE ClickHouseProfileEvents_ReadBackoff counter
ClickHouseProfileEvents_ReadBackoff 0
# HELP ClickHouseProfileEvents_ReplicaPartialShutdown How many times Replicated table has to deinitialize its state due to session expiration in ZooKeeper. The state is reinitialized every time when ZooKeeper is available again.
# TYPE ClickHouseProfileEvents_ReplicaPartialShutdown counter
ClickHouseProfileEvents_ReplicaPartialShutdown 0
# HELP ClickHouseProfileEvents_SelectedParts Number of data parts selected to read from a MergeTree table.
# TYPE ClickHouseProfileEvents_SelectedParts counter
ClickHouseProfileEvents_SelectedParts 0
# HELP ClickHouseProfileEvents_SelectedRanges Number of (non-adjacent) ranges in all data parts selected to read from a MergeTree table.
# TYPE ClickHouseProfileEvents_SelectedRanges counter
ClickHouseProfileEvents_SelectedRanges 0
# HELP ClickHouseProfileEvents_SelectedMarks Number of marks (index granules) selected to read from a MergeTree table.
# TYPE ClickHouseProfileEvents_SelectedMarks counter
ClickHouseProfileEvents_SelectedMarks 0
# HELP ClickHouseProfileEvents_SelectedRows Number of rows SELECTed from all tables.
# TYPE ClickHouseProfileEvents_SelectedRows counter
ClickHouseProfileEvents_SelectedRows 6263
# HELP ClickHouseProfileEvents_SelectedBytes Number of bytes (uncompressed; for columns as they stored in memory) SELECTed from all tables.
# TYPE ClickHouseProfileEvents_SelectedBytes counter
ClickHouseProfileEvents_SelectedBytes 455089
# HELP ClickHouseProfileEvents_Merge Number of launched background merges.
# TYPE ClickHouseProfileEvents_Merge counter
ClickHouseProfileEvents_Merge 401
# HELP ClickHouseProfileEvents_MergedRows Rows read for background merges. This is the number of rows before merge.
# TYPE ClickHouseProfileEvents_MergedRows counter
ClickHouseProfileEvents_MergedRows 2996048
# HELP ClickHouseProfileEvents_MergedUncompressedBytes Uncompressed bytes (for columns as they stored in memory) that was read for background merges. This is the number before merge.
# TYPE ClickHouseProfileEvents_MergedUncompressedBytes counter
ClickHouseProfileEvents_MergedUncompressedBytes 118902299
# HELP ClickHouseProfileEvents_MergesTimeMilliseconds Total time spent for background merges.
# TYPE ClickHouseProfileEvents_MergesTimeMilliseconds counter
ClickHouseProfileEvents_MergesTimeMilliseconds 717
# HELP ClickHouseProfileEvents_MergeTreeDataWriterRows Number of rows INSERTed to MergeTree tables.
# TYPE ClickHouseProfileEvents_MergeTreeDataWriterRows counter
ClickHouseProfileEvents_MergeTreeDataWriterRows 6263
# HELP ClickHouseProfileEvents_MergeTreeDataWriterUncompressedBytes Uncompressed bytes (for columns as they stored in memory) INSERTed to MergeTree tables.
# TYPE ClickHouseProfileEvents_MergeTreeDataWriterUncompressedBytes counter
ClickHouseProfileEvents_MergeTreeDataWriterUncompressedBytes 455089
# HELP ClickHouseProfileEvents_MergeTreeDataWriterCompressedBytes Bytes written to filesystem for data INSERTed to MergeTree tables.
# TYPE ClickHouseProfileEvents_MergeTreeDataWriterCompressedBytes counter
ClickHouseProfileEvents_MergeTreeDataWriterCompressedBytes 115473
# HELP ClickHouseProfileEvents_MergeTreeDataWriterBlocks Number of blocks INSERTed to MergeTree tables. Each block forms a data part of level zero.
# TYPE ClickHouseProfileEvents_MergeTreeDataWriterBlocks counter
ClickHouseProfileEvents_MergeTreeDataWriterBlocks 7
# HELP ClickHouseProfileEvents_MergeTreeDataWriterBlocksAlreadySorted Number of blocks INSERTed to MergeTree tables that appeared to be already sorted.
# TYPE ClickHouseProfileEvents_MergeTreeDataWriterBlocksAlreadySorted counter
ClickHouseProfileEvents_MergeTreeDataWriterBlocksAlreadySorted 4
# HELP ClickHouseProfileEvents_InsertedWideParts Number of parts inserted in Wide format.
# TYPE ClickHouseProfileEvents_InsertedWideParts counter
ClickHouseProfileEvents_InsertedWideParts 0
# HELP ClickHouseProfileEvents_InsertedCompactParts Number of parts inserted in Compact format.
# TYPE ClickHouseProfileEvents_InsertedCompactParts counter
ClickHouseProfileEvents_InsertedCompactParts 7
# HELP ClickHouseProfileEvents_InsertedInMemoryParts Number of parts inserted in InMemory format.
# TYPE ClickHouseProfileEvents_InsertedInMemoryParts counter
ClickHouseProfileEvents_InsertedInMemoryParts 0
# HELP ClickHouseProfileEvents_MergedIntoWideParts Number of parts merged into Wide format.
# TYPE ClickHouseProfileEvents_MergedIntoWideParts counter
ClickHouseProfileEvents_MergedIntoWideParts 3
# HELP ClickHouseProfileEvents_MergedIntoCompactParts Number of parts merged into Compact format.
# TYPE ClickHouseProfileEvents_MergedIntoCompactParts counter
ClickHouseProfileEvents_MergedIntoCompactParts 2
# HELP ClickHouseProfileEvents_MergedIntoInMemoryParts Number of parts in merged into InMemory format.
# TYPE ClickHouseProfileEvents_MergedIntoInMemoryParts counter
ClickHouseProfileEvents_MergedIntoInMemoryParts 0
# HELP ClickHouseProfileEvents_MergeTreeDataProjectionWriterRows Number of rows INSERTed to MergeTree tables projection.
# TYPE ClickHouseProfileEvents_MergeTreeDataProjectionWriterRows counter
ClickHouseProfileEvents_MergeTreeDataProjectionWriterRows 0
# HELP ClickHouseProfileEvents_MergeTreeDataProjectionWriterUncompressedBytes Uncompressed bytes (for columns as they stored in memory) INSERTed to MergeTree tables projection.
# TYPE ClickHouseProfileEvents_MergeTreeDataProjectionWriterUncompressedBytes counter
ClickHouseProfileEvents_MergeTreeDataProjectionWriterUncompressedBytes 0
# HELP ClickHouseProfileEvents_MergeTreeDataProjectionWriterCompressedBytes Bytes written to filesystem for data INSERTed to MergeTree tables projection.
# TYPE ClickHouseProfileEvents_MergeTreeDataProjectionWriterCompressedBytes counter
ClickHouseProfileEvents_MergeTreeDataProjectionWriterCompressedBytes 0
# HELP ClickHouseProfileEvents_MergeTreeDataProjectionWriterBlocks Number of blocks INSERTed to MergeTree tables projection. Each block forms a data part of level zero.
# TYPE ClickHouseProfileEvents_MergeTreeDataProjectionWriterBlocks counter
ClickHouseProfileEvents_MergeTreeDataProjectionWriterBlocks 0
# HELP ClickHouseProfileEvents_MergeTreeDataProjectionWriterBlocksAlreadySorted Number of blocks INSERTed to MergeTree tables projection that appeared to be already sorted.
# TYPE ClickHouseProfileEvents_MergeTreeDataProjectionWriterBlocksAlreadySorted counter
ClickHouseProfileEvents_MergeTreeDataProjectionWriterBlocksAlreadySorted 0
# HELP ClickHouseProfileEvents_CannotRemoveEphemeralNode Number of times an error happened while trying to remove ephemeral node. This is not an issue, because our implementation of ZooKeeper library guarantee that the session will expire and the node will be removed.
# TYPE ClickHouseProfileEvents_CannotRemoveEphemeralNode counter
ClickHouseProfileEvents_CannotRemoveEphemeralNode 0
# HELP ClickHouseProfileEvents_RegexpCreated Compiled regular expressions. Identical regular expressions compiled just once and cached forever.
# TYPE ClickHouseProfileEvents_RegexpCreated counter
ClickHouseProfileEvents_RegexpCreated 0
# HELP ClickHouseProfileEvents_ContextLock Number of times the lock of Context was acquired or tried to acquire. This is global lock.
# TYPE ClickHouseProfileEvents_ContextLock counter
ClickHouseProfileEvents_ContextLock 665
# HELP ClickHouseProfileEvents_StorageBufferFlush 
# TYPE ClickHouseProfileEvents_StorageBufferFlush counter
ClickHouseProfileEvents_StorageBufferFlush 0
# HELP ClickHouseProfileEvents_StorageBufferErrorOnFlush 
# TYPE ClickHouseProfileEvents_StorageBufferErrorOnFlush counter
ClickHouseProfileEvents_StorageBufferErrorOnFlush 0
# HELP ClickHouseProfileEvents_StorageBufferPassedAllMinThresholds 
# TYPE ClickHouseProfileEvents_StorageBufferPassedAllMinThresholds counter
ClickHouseProfileEvents_StorageBufferPassedAllMinThresholds 0
# HELP ClickHouseProfileEvents_StorageBufferPassedTimeMaxThreshold 
# TYPE ClickHouseProfileEvents_StorageBufferPassedTimeMaxThreshold counter
ClickHouseProfileEvents_StorageBufferPassedTimeMaxThreshold 0
# HELP ClickHouseProfileEvents_StorageBufferPassedRowsMaxThreshold 
# TYPE ClickHouseProfileEvents_StorageBufferPassedRowsMaxThreshold counter
ClickHouseProfileEvents_StorageBufferPassedRowsMaxThreshold 0
# HELP ClickHouseProfileEvents_StorageBufferPassedBytesMaxThreshold 
# TYPE ClickHouseProfileEvents_StorageBufferPassedBytesMaxThreshold counter
ClickHouseProfileEvents_StorageBufferPassedBytesMaxThreshold 0
# HELP ClickHouseProfileEvents_StorageBufferPassedTimeFlushThreshold 
# TYPE ClickHouseProfileEvents_StorageBufferPassedTimeFlushThreshold counter
ClickHouseProfileEvents_StorageBufferPassedTimeFlushThreshold 0
# HELP ClickHouseProfileEvents_StorageBufferPassedRowsFlushThreshold 
# TYPE ClickHouseProfileEvents_StorageBufferPassedRowsFlushThreshold counter
ClickHouseProfileEvents_StorageBufferPassedRowsFlushThreshold 0
# HELP ClickHouseProfileEvents_StorageBufferPassedBytesFlushThreshold 
# TYPE ClickHouseProfileEvents_StorageBufferPassedBytesFlushThreshold counter
ClickHouseProfileEvents_StorageBufferPassedBytesFlushThreshold 0
# HELP ClickHouseProfileEvents_StorageBufferLayerLockReadersWaitMilliseconds Time for waiting for Buffer layer during reading
# TYPE ClickHouseProfileEvents_StorageBufferLayerLockReadersWaitMilliseconds counter
ClickHouseProfileEvents_StorageBufferLayerLockReadersWaitMilliseconds 0
# HELP ClickHouseProfileEvents_StorageBufferLayerLockWritersWaitMilliseconds Time for waiting free Buffer layer to write to (can be used to tune Buffer layers)
# TYPE ClickHouseProfileEvents_StorageBufferLayerLockWritersWaitMilliseconds counter
ClickHouseProfileEvents_StorageBufferLayerLockWritersWaitMilliseconds 0
# HELP ClickHouseProfileEvents_DictCacheKeysRequested 
# TYPE ClickHouseProfileEvents_DictCacheKeysRequested counter
ClickHouseProfileEvents_DictCacheKeysRequested 0
# HELP ClickHouseProfileEvents_DictCacheKeysRequestedMiss 
# TYPE ClickHouseProfileEvents_DictCacheKeysRequestedMiss counter
ClickHouseProfileEvents_DictCacheKeysRequestedMiss 0
# HELP ClickHouseProfileEvents_DictCacheKeysRequestedFound 
# TYPE ClickHouseProfileEvents_DictCacheKeysRequestedFound counter
ClickHouseProfileEvents_DictCacheKeysRequestedFound 0
# HELP ClickHouseProfileEvents_DictCacheKeysExpired 
# TYPE ClickHouseProfileEvents_DictCacheKeysExpired counter
ClickHouseProfileEvents_DictCacheKeysExpired 0
# HELP ClickHouseProfileEvents_DictCacheKeysNotFound 
# TYPE ClickHouseProfileEvents_DictCacheKeysNotFound counter
ClickHouseProfileEvents_DictCacheKeysNotFound 0
# HELP ClickHouseProfileEvents_DictCacheKeysHit 
# TYPE ClickHouseProfileEvents_DictCacheKeysHit counter
ClickHouseProfileEvents_DictCacheKeysHit 0
# HELP ClickHouseProfileEvents_DictCacheRequestTimeNs 
# TYPE ClickHouseProfileEvents_DictCacheRequestTimeNs counter
ClickHouseProfileEvents_DictCacheRequestTimeNs 0
# HELP ClickHouseProfileEvents_DictCacheRequests 
# TYPE ClickHouseProfileEvents_DictCacheRequests counter
ClickHouseProfileEvents_DictCacheRequests 0
# HELP ClickHouseProfileEvents_DictCacheLockWriteNs 
# TYPE ClickHouseProfileEvents_DictCacheLockWriteNs counter
ClickHouseProfileEvents_DictCacheLockWriteNs 0
# HELP ClickHouseProfileEvents_DictCacheLockReadNs 
# TYPE ClickHouseProfileEvents_DictCacheLockReadNs counter
ClickHouseProfileEvents_DictCacheLockReadNs 0
# HELP ClickHouseProfileEvents_DistributedSyncInsertionTimeoutExceeded 
# TYPE ClickHouseProfileEvents_DistributedSyncInsertionTimeoutExceeded counter
ClickHouseProfileEvents_DistributedSyncInsertionTimeoutExceeded 0
# HELP ClickHouseProfileEvents_DataAfterMergeDiffersFromReplica 
# TYPE ClickHouseProfileEvents_DataAfterMergeDiffersFromReplica counter
ClickHouseProfileEvents_DataAfterMergeDiffersFromReplica 0
# HELP ClickHouseProfileEvents_DataAfterMutationDiffersFromReplica 
# TYPE ClickHouseProfileEvents_DataAfterMutationDiffersFromReplica counter
ClickHouseProfileEvents_DataAfterMutationDiffersFromReplica 0
# HELP ClickHouseProfileEvents_PolygonsAddedToPool 
# TYPE ClickHouseProfileEvents_PolygonsAddedToPool counter
ClickHouseProfileEvents_PolygonsAddedToPool 0
# HELP ClickHouseProfileEvents_PolygonsInPoolAllocatedBytes 
# TYPE ClickHouseProfileEvents_PolygonsInPoolAllocatedBytes counter
ClickHouseProfileEvents_PolygonsInPoolAllocatedBytes 0
# HELP ClickHouseProfileEvents_RWLockAcquiredReadLocks 
# TYPE ClickHouseProfileEvents_RWLockAcquiredReadLocks counter
ClickHouseProfileEvents_RWLockAcquiredReadLocks 117
# HELP ClickHouseProfileEvents_RWLockAcquiredWriteLocks 
# TYPE ClickHouseProfileEvents_RWLockAcquiredWriteLocks counter
ClickHouseProfileEvents_RWLockAcquiredWriteLocks 0
# HELP ClickHouseProfileEvents_RWLockReadersWaitMilliseconds 
# TYPE ClickHouseProfileEvents_RWLockReadersWaitMilliseconds counter
ClickHouseProfileEvents_RWLockReadersWaitMilliseconds 3
# HELP ClickHouseProfileEvents_RWLockWritersWaitMilliseconds 
# TYPE ClickHouseProfileEvents_RWLockWritersWaitMilliseconds counter
ClickHouseProfileEvents_RWLockWritersWaitMilliseconds 0
# HELP ClickHouseProfileEvents_DNSError Total count of errors in DNS resolution
# TYPE ClickHouseProfileEvents_DNSError counter
ClickHouseProfileEvents_DNSError 0
# HELP ClickHouseProfileEvents_RealTimeMicroseconds Total (wall clock) time spent in processing (queries and other tasks) threads (not that this is a sum).
# TYPE ClickHouseProfileEvents_RealTimeMicroseconds counter
ClickHouseProfileEvents_RealTimeMicroseconds 0
# HELP ClickHouseProfileEvents_UserTimeMicroseconds Total time spent in processing (queries and other tasks) threads executing CPU instructions in user space. This include time CPU pipeline was stalled due to cache misses, branch mispredictions, hyper-threading, etc.
# TYPE ClickHouseProfileEvents_UserTimeMicroseconds counter
ClickHouseProfileEvents_UserTimeMicroseconds 0
# HELP ClickHouseProfileEvents_SystemTimeMicroseconds Total time spent in processing (queries and other tasks) threads executing CPU instructions in OS kernel space. This include time CPU pipeline was stalled due to cache misses, branch mispredictions, hyper-threading, etc.
# TYPE ClickHouseProfileEvents_SystemTimeMicroseconds counter
ClickHouseProfileEvents_SystemTimeMicroseconds 0
# HELP ClickHouseProfileEvents_MemoryOvercommitWaitTimeMicroseconds Total time spent in waiting for memory to be freed in OvercommitTracker.
# TYPE ClickHouseProfileEvents_MemoryOvercommitWaitTimeMicroseconds counter
ClickHouseProfileEvents_MemoryOvercommitWaitTimeMicroseconds 0
# HELP ClickHouseProfileEvents_SoftPageFaults 
# TYPE ClickHouseProfileEvents_SoftPageFaults counter
ClickHouseProfileEvents_SoftPageFaults 0
# HELP ClickHouseProfileEvents_HardPageFaults 
# TYPE ClickHouseProfileEvents_HardPageFaults counter
ClickHouseProfileEvents_HardPageFaults 0
# HELP ClickHouseProfileEvents_OSIOWaitMicroseconds Total time a thread spent waiting for a result of IO operation, from the OS point of view. This is real IO that doesn't include page cache.
# TYPE ClickHouseProfileEvents_OSIOWaitMicroseconds counter
ClickHouseProfileEvents_OSIOWaitMicroseconds 0
# HELP ClickHouseProfileEvents_OSCPUWaitMicroseconds Total time a thread was ready for execution but waiting to be scheduled by OS, from the OS point of view.
# TYPE ClickHouseProfileEvents_OSCPUWaitMicroseconds counter
ClickHouseProfileEvents_OSCPUWaitMicroseconds 0
# HELP ClickHouseProfileEvents_OSCPUVirtualTimeMicroseconds CPU time spent seen by OS. Does not include involuntary waits due to virtualization.
# TYPE ClickHouseProfileEvents_OSCPUVirtualTimeMicroseconds counter
ClickHouseProfileEvents_OSCPUVirtualTimeMicroseconds 0
# HELP ClickHouseProfileEvents_OSReadBytes Number of bytes read from disks or block devices. Doesn't include bytes read from page cache. May include excessive data due to block size, readahead, etc.
# TYPE ClickHouseProfileEvents_OSReadBytes counter
ClickHouseProfileEvents_OSReadBytes 0
# HELP ClickHouseProfileEvents_OSWriteBytes Number of bytes written to disks or block devices. Doesn't include bytes that are in page cache dirty pages. May not include data that was written by OS asynchronously.
# TYPE ClickHouseProfileEvents_OSWriteBytes counter
ClickHouseProfileEvents_OSWriteBytes 0
# HELP ClickHouseProfileEvents_OSReadChars Number of bytes read from filesystem, including page cache.
# TYPE ClickHouseProfileEvents_OSReadChars counter
ClickHouseProfileEvents_OSReadChars 0
# HELP ClickHouseProfileEvents_OSWriteChars Number of bytes written to filesystem, including page cache.
# TYPE ClickHouseProfileEvents_OSWriteChars counter
ClickHouseProfileEvents_OSWriteChars 0
# HELP ClickHouseProfileEvents_PerfCpuCycles Total cycles. Be wary of what happens during CPU frequency scaling.
# TYPE ClickHouseProfileEvents_PerfCpuCycles counter
ClickHouseProfileEvents_PerfCpuCycles 0
# HELP ClickHouseProfileEvents_PerfInstructions Retired instructions. Be careful, these can be affected by various issues, most notably hardware interrupt counts.
# TYPE ClickHouseProfileEvents_PerfInstructions counter
ClickHouseProfileEvents_PerfInstructions 0
# HELP ClickHouseProfileEvents_PerfCacheReferences Cache accesses. Usually this indicates Last Level Cache accesses but this may vary depending on your CPU. This may include prefetches and coherency messages; again this depends on the design of your CPU.
# TYPE ClickHouseProfileEvents_PerfCacheReferences counter
ClickHouseProfileEvents_PerfCacheReferences 0
# HELP ClickHouseProfileEvents_PerfCacheMisses Cache misses. Usually this indicates Last Level Cache misses; this is intended to be used in con‐junction with the PERFCOUNTHWCACHEREFERENCES event to calculate cache miss rates.
# TYPE ClickHouseProfileEvents_PerfCacheMisses counter
ClickHouseProfileEvents_PerfCacheMisses 0
# HELP ClickHouseProfileEvents_PerfBranchInstructions Retired branch instructions. Prior to Linux 2.6.35, this used the wrong event on AMD processors.
# TYPE ClickHouseProfileEvents_PerfBranchInstructions counter
ClickHouseProfileEvents_PerfBranchInstructions 0
# HELP ClickHouseProfileEvents_PerfBranchMisses Mispredicted branch instructions.
# TYPE ClickHouseProfileEvents_PerfBranchMisses counter
ClickHouseProfileEvents_PerfBranchMisses 0
# HELP ClickHouseProfileEvents_PerfBusCycles Bus cycles, which can be different from total cycles.
# TYPE ClickHouseProfileEvents_PerfBusCycles counter
ClickHouseProfileEvents_PerfBusCycles 0
# HELP ClickHouseProfileEvents_PerfStalledCyclesFrontend Stalled cycles during issue.
# TYPE ClickHouseProfileEvents_PerfStalledCyclesFrontend counter
ClickHouseProfileEvents_PerfStalledCyclesFrontend 0
# HELP ClickHouseProfileEvents_PerfStalledCyclesBackend Stalled cycles during retirement.
# TYPE ClickHouseProfileEvents_PerfStalledCyclesBackend counter
ClickHouseProfileEvents_PerfStalledCyclesBackend 0
# HELP ClickHouseProfileEvents_PerfRefCpuCycles Total cycles; not affected by CPU frequency scaling.
# TYPE ClickHouseProfileEvents_PerfRefCpuCycles counter
ClickHouseProfileEvents_PerfRefCpuCycles 0
# HELP ClickHouseProfileEvents_PerfCpuClock The CPU clock, a high-resolution per-CPU timer
# TYPE ClickHouseProfileEvents_PerfCpuClock counter
ClickHouseProfileEvents_PerfCpuClock 0
# HELP ClickHouseProfileEvents_PerfTaskClock A clock count specific to the task that is running
# TYPE ClickHouseProfileEvents_PerfTaskClock counter
ClickHouseProfileEvents_PerfTaskClock 0
# HELP ClickHouseProfileEvents_PerfContextSwitches Number of context switches
# TYPE ClickHouseProfileEvents_PerfContextSwitches counter
ClickHouseProfileEvents_PerfContextSwitches 0
# HELP ClickHouseProfileEvents_PerfCpuMigrations Number of times the process has migrated to a new CPU
# TYPE ClickHouseProfileEvents_PerfCpuMigrations counter
ClickHouseProfileEvents_PerfCpuMigrations 0
# HELP ClickHouseProfileEvents_PerfAlignmentFaults Number of alignment faults. These happen when unaligned memory accesses happen; the kernel can handle these but it reduces performance. This happens only on some architectures (never on x86).
# TYPE ClickHouseProfileEvents_PerfAlignmentFaults counter
ClickHouseProfileEvents_PerfAlignmentFaults 0
# HELP ClickHouseProfileEvents_PerfEmulationFaults Number of emulation faults. The kernel sometimes traps on unimplemented instructions and emulates them for user space. This can negatively impact performance.
# TYPE ClickHouseProfileEvents_PerfEmulationFaults counter
ClickHouseProfileEvents_PerfEmulationFaults 0
# HELP ClickHouseProfileEvents_PerfMinEnabledTime For all events, minimum time that an event was enabled. Used to track event multiplexing influence
# TYPE ClickHouseProfileEvents_PerfMinEnabledTime counter
ClickHouseProfileEvents_PerfMinEnabledTime 0
# HELP ClickHouseProfileEvents_PerfMinEnabledRunningTime Running time for event with minimum enabled time. Used to track the amount of event multiplexing
# TYPE ClickHouseProfileEvents_PerfMinEnabledRunningTime counter
ClickHouseProfileEvents_PerfMinEnabledRunningTime 0
# HELP ClickHouseProfileEvents_PerfDataTLBReferences Data TLB references
# TYPE ClickHouseProfileEvents_PerfDataTLBReferences counter
ClickHouseProfileEvents_PerfDataTLBReferences 0
# HELP ClickHouseProfileEvents_PerfDataTLBMisses Data TLB misses
# TYPE ClickHouseProfileEvents_PerfDataTLBMisses counter
ClickHouseProfileEvents_PerfDataTLBMisses 0
# HELP ClickHouseProfileEvents_PerfInstructionTLBReferences Instruction TLB references
# TYPE ClickHouseProfileEvents_PerfInstructionTLBReferences counter
ClickHouseProfileEvents_PerfInstructionTLBReferences 0
# HELP ClickHouseProfileEvents_PerfInstructionTLBMisses Instruction TLB misses
# TYPE ClickHouseProfileEvents_PerfInstructionTLBMisses counter
ClickHouseProfileEvents_PerfInstructionTLBMisses 0
# HELP ClickHouseProfileEvents_PerfLocalMemoryReferences Local NUMA node memory reads
# TYPE ClickHouseProfileEvents_PerfLocalMemoryReferences counter
ClickHouseProfileEvents_PerfLocalMemoryReferences 0
# HELP ClickHouseProfileEvents_PerfLocalMemoryMisses Local NUMA node memory read misses
# TYPE ClickHouseProfileEvents_PerfLocalMemoryMisses counter
ClickHouseProfileEvents_PerfLocalMemoryMisses 0
# HELP ClickHouseProfileEvents_CreatedHTTPConnections Total amount of created HTTP connections (counter increase every time connection is created).
# TYPE ClickHouseProfileEvents_CreatedHTTPConnections counter
ClickHouseProfileEvents_CreatedHTTPConnections 0
# HELP ClickHouseProfileEvents_CannotWriteToWriteBufferDiscard Number of stack traces dropped by query profiler or signal handler because pipe is full or cannot write to pipe.
# TYPE ClickHouseProfileEvents_CannotWriteToWriteBufferDiscard counter
ClickHouseProfileEvents_CannotWriteToWriteBufferDiscard 18
# HELP ClickHouseProfileEvents_QueryProfilerSignalOverruns Number of times we drop processing of a query profiler signal due to overrun plus the number of signals that OS has not delivered due to overrun.
# TYPE ClickHouseProfileEvents_QueryProfilerSignalOverruns counter
ClickHouseProfileEvents_QueryProfilerSignalOverruns 0
# HELP ClickHouseProfileEvents_QueryProfilerRuns Number of times QueryProfiler had been run.
# TYPE ClickHouseProfileEvents_QueryProfilerRuns counter
ClickHouseProfileEvents_QueryProfilerRuns 0
# HELP ClickHouseProfileEvents_CreatedLogEntryForMerge Successfully created log entry to merge parts in ReplicatedMergeTree.
# TYPE ClickHouseProfileEvents_CreatedLogEntryForMerge counter
ClickHouseProfileEvents_CreatedLogEntryForMerge 0
# HELP ClickHouseProfileEvents_NotCreatedLogEntryForMerge Log entry to merge parts in ReplicatedMergeTree is not created due to concurrent log update by another replica.
# TYPE ClickHouseProfileEvents_NotCreatedLogEntryForMerge counter
ClickHouseProfileEvents_NotCreatedLogEntryForMerge 0
# HELP ClickHouseProfileEvents_CreatedLogEntryForMutation Successfully created log entry to mutate parts in ReplicatedMergeTree.
# TYPE ClickHouseProfileEvents_CreatedLogEntryForMutation counter
ClickHouseProfileEvents_CreatedLogEntryForMutation 0
# HELP ClickHouseProfileEvents_NotCreatedLogEntryForMutation Log entry to mutate parts in ReplicatedMergeTree is not created due to concurrent log update by another replica.
# TYPE ClickHouseProfileEvents_NotCreatedLogEntryForMutation counter
ClickHouseProfileEvents_NotCreatedLogEntryForMutation 0
# HELP ClickHouseProfileEvents_S3ReadMicroseconds Time of GET and HEAD requests to S3 storage.
# TYPE ClickHouseProfileEvents_S3ReadMicroseconds counter
ClickHouseProfileEvents_S3ReadMicroseconds 0
# HELP ClickHouseProfileEvents_S3ReadRequestsCount Number of GET and HEAD requests to S3 storage.
# TYPE ClickHouseProfileEvents_S3ReadRequestsCount counter
ClickHouseProfileEvents_S3ReadRequestsCount 0
# HELP ClickHouseProfileEvents_S3ReadRequestsErrors Number of non-throttling errors in GET and HEAD requests to S3 storage.
# TYPE ClickHouseProfileEvents_S3ReadRequestsErrors counter
ClickHouseProfileEvents_S3ReadRequestsErrors 0
# HELP ClickHouseProfileEvents_S3ReadRequestsThrottling Number of 429 and 503 errors in GET and HEAD requests to S3 storage.
# TYPE ClickHouseProfileEvents_S3ReadRequestsThrottling counter
ClickHouseProfileEvents_S3ReadRequestsThrottling 0
# HELP ClickHouseProfileEvents_S3ReadRequestsRedirects Number of redirects in GET and HEAD requests to S3 storage.
# TYPE ClickHouseProfileEvents_S3ReadRequestsRedirects counter
ClickHouseProfileEvents_S3ReadRequestsRedirects 0
# HELP ClickHouseProfileEvents_S3WriteMicroseconds Time of POST, DELETE, PUT and PATCH requests to S3 storage.
# TYPE ClickHouseProfileEvents_S3WriteMicroseconds counter
ClickHouseProfileEvents_S3WriteMicroseconds 0
# HELP ClickHouseProfileEvents_S3WriteRequestsCount Number of POST, DELETE, PUT and PATCH requests to S3 storage.
# TYPE ClickHouseProfileEvents_S3WriteRequestsCount counter
ClickHouseProfileEvents_S3WriteRequestsCount 0
# HELP ClickHouseProfileEvents_S3WriteRequestsErrors Number of non-throttling errors in POST, DELETE, PUT and PATCH requests to S3 storage.
# TYPE ClickHouseProfileEvents_S3WriteRequestsErrors counter
ClickHouseProfileEvents_S3WriteRequestsErrors 0
# HELP ClickHouseProfileEvents_S3WriteRequestsThrottling Number of 429 and 503 errors in POST, DELETE, PUT and PATCH requests to S3 storage.
# TYPE ClickHouseProfileEvents_S3WriteRequestsThrottling counter
ClickHouseProfileEvents_S3WriteRequestsThrottling 0
# HELP ClickHouseProfileEvents_S3WriteRequestsRedirects Number of redirects in POST, DELETE, PUT and PATCH requests to S3 storage.
# TYPE ClickHouseProfileEvents_S3WriteRequestsRedirects counter
ClickHouseProfileEvents_S3WriteRequestsRedirects 0
# HELP ClickHouseProfileEvents_ReadBufferFromS3Microseconds Time spend in reading from S3.
# TYPE ClickHouseProfileEvents_ReadBufferFromS3Microseconds counter
ClickHouseProfileEvents_ReadBufferFromS3Microseconds 0
# HELP ClickHouseProfileEvents_ReadBufferFromS3Bytes Bytes read from S3.
# TYPE ClickHouseProfileEvents_ReadBufferFromS3Bytes counter
ClickHouseProfileEvents_ReadBufferFromS3Bytes 0
# HELP ClickHouseProfileEvents_ReadBufferFromS3RequestsErrors Number of exceptions while reading from S3.
# TYPE ClickHouseProfileEvents_ReadBufferFromS3RequestsErrors counter
ClickHouseProfileEvents_ReadBufferFromS3RequestsErrors 0
# HELP ClickHouseProfileEvents_WriteBufferFromS3Bytes Bytes written to S3.
# TYPE ClickHouseProfileEvents_WriteBufferFromS3Bytes counter
ClickHouseProfileEvents_WriteBufferFromS3Bytes 0
# HELP ClickHouseProfileEvents_QueryMemoryLimitExceeded Number of times when memory limit exceeded for query.
# TYPE ClickHouseProfileEvents_QueryMemoryLimitExceeded counter
ClickHouseProfileEvents_QueryMemoryLimitExceeded 0
# HELP ClickHouseProfileEvents_CachedReadBufferReadFromSourceMicroseconds Time reading from filesystem cache source (from remote filesystem, etc)
# TYPE ClickHouseProfileEvents_CachedReadBufferReadFromSourceMicroseconds counter
ClickHouseProfileEvents_CachedReadBufferReadFromSourceMicroseconds 0
# HELP ClickHouseProfileEvents_CachedReadBufferReadFromCacheMicroseconds Time reading from filesystem cache
# TYPE ClickHouseProfileEvents_CachedReadBufferReadFromCacheMicroseconds counter
ClickHouseProfileEvents_CachedReadBufferReadFromCacheMicroseconds 0
# HELP ClickHouseProfileEvents_CachedReadBufferReadFromSourceBytes Bytes read from filesystem cache source (from remote fs, etc)
# TYPE ClickHouseProfileEvents_CachedReadBufferReadFromSourceBytes counter
ClickHouseProfileEvents_CachedReadBufferReadFromSourceBytes 0
# HELP ClickHouseProfileEvents_CachedReadBufferReadFromCacheBytes Bytes read from filesystem cache
# TYPE ClickHouseProfileEvents_CachedReadBufferReadFromCacheBytes counter
ClickHouseProfileEvents_CachedReadBufferReadFromCacheBytes 0
# HELP ClickHouseProfileEvents_CachedReadBufferCacheWriteBytes Bytes written from source (remote fs, etc) to filesystem cache
# TYPE ClickHouseProfileEvents_CachedReadBufferCacheWriteBytes counter
ClickHouseProfileEvents_CachedReadBufferCacheWriteBytes 0
# HELP ClickHouseProfileEvents_CachedReadBufferCacheWriteMicroseconds Time spent writing data into filesystem cache
# TYPE ClickHouseProfileEvents_CachedReadBufferCacheWriteMicroseconds counter
ClickHouseProfileEvents_CachedReadBufferCacheWriteMicroseconds 0
# HELP ClickHouseProfileEvents_CachedWriteBufferCacheWriteBytes Bytes written from source (remote fs, etc) to filesystem cache
# TYPE ClickHouseProfileEvents_CachedWriteBufferCacheWriteBytes counter
ClickHouseProfileEvents_CachedWriteBufferCacheWriteBytes 0
# HELP ClickHouseProfileEvents_CachedWriteBufferCacheWriteMicroseconds Time spent writing data into filesystem cache
# TYPE ClickHouseProfileEvents_CachedWriteBufferCacheWriteMicroseconds counter
ClickHouseProfileEvents_CachedWriteBufferCacheWriteMicroseconds 0
# HELP ClickHouseProfileEvents_RemoteFSSeeks Total number of seeks for async buffer
# TYPE ClickHouseProfileEvents_RemoteFSSeeks counter
ClickHouseProfileEvents_RemoteFSSeeks 0
# HELP ClickHouseProfileEvents_RemoteFSPrefetches Number of prefetches made with asynchronous reading from remote filesystem
# TYPE ClickHouseProfileEvents_RemoteFSPrefetches counter
ClickHouseProfileEvents_RemoteFSPrefetches 0
# HELP ClickHouseProfileEvents_RemoteFSCancelledPrefetches Number of canceled prefecthes (because of seek)
# TYPE ClickHouseProfileEvents_RemoteFSCancelledPrefetches counter
ClickHouseProfileEvents_RemoteFSCancelledPrefetches 0
# HELP ClickHouseProfileEvents_RemoteFSUnusedPrefetches Number of prefetches pending at buffer destruction
# TYPE ClickHouseProfileEvents_RemoteFSUnusedPrefetches counter
ClickHouseProfileEvents_RemoteFSUnusedPrefetches 0
# HELP ClickHouseProfileEvents_RemoteFSPrefetchedReads Number of reads from prefecthed buffer
# TYPE ClickHouseProfileEvents_RemoteFSPrefetchedReads counter
ClickHouseProfileEvents_RemoteFSPrefetchedReads 0
# HELP ClickHouseProfileEvents_RemoteFSUnprefetchedReads Number of reads from unprefetched buffer
# TYPE ClickHouseProfileEvents_RemoteFSUnprefetchedReads counter
ClickHouseProfileEvents_RemoteFSUnprefetchedReads 0
# HELP ClickHouseProfileEvents_RemoteFSLazySeeks Number of lazy seeks
# TYPE ClickHouseProfileEvents_RemoteFSLazySeeks counter
ClickHouseProfileEvents_RemoteFSLazySeeks 0
# HELP ClickHouseProfileEvents_RemoteFSSeeksWithReset Number of seeks which lead to a new connection
# TYPE ClickHouseProfileEvents_RemoteFSSeeksWithReset counter
ClickHouseProfileEvents_RemoteFSSeeksWithReset 0
# HELP ClickHouseProfileEvents_RemoteFSBuffers Number of buffers created for asynchronous reading from remote filesystem
# TYPE ClickHouseProfileEvents_RemoteFSBuffers counter
ClickHouseProfileEvents_RemoteFSBuffers 0
# HELP ClickHouseProfileEvents_ThreadpoolReaderTaskMicroseconds Time spent getting the data in asynchronous reading
# TYPE ClickHouseProfileEvents_ThreadpoolReaderTaskMicroseconds counter
ClickHouseProfileEvents_ThreadpoolReaderTaskMicroseconds 0
# HELP ClickHouseProfileEvents_ThreadpoolReaderReadBytes Bytes read from a threadpool task in asynchronous reading
# TYPE ClickHouseProfileEvents_ThreadpoolReaderReadBytes counter
ClickHouseProfileEvents_ThreadpoolReaderReadBytes 0
# HELP ClickHouseProfileEvents_FileSegmentWaitReadBufferMicroseconds Metric per file segment. Time spend waiting for internal read buffer (includes cache waiting)
# TYPE ClickHouseProfileEvents_FileSegmentWaitReadBufferMicroseconds counter
ClickHouseProfileEvents_FileSegmentWaitReadBufferMicroseconds 0
# HELP ClickHouseProfileEvents_FileSegmentReadMicroseconds Metric per file segment. Time spend reading from file
# TYPE ClickHouseProfileEvents_FileSegmentReadMicroseconds counter
ClickHouseProfileEvents_FileSegmentReadMicroseconds 0
# HELP ClickHouseProfileEvents_FileSegmentCacheWriteMicroseconds Metric per file segment. Time spend writing data to cache
# TYPE ClickHouseProfileEvents_FileSegmentCacheWriteMicroseconds counter
ClickHouseProfileEvents_FileSegmentCacheWriteMicroseconds 0
# HELP ClickHouseProfileEvents_FileSegmentPredownloadMicroseconds Metric per file segment. Time spent predownloading data to cache (predownloading - finishing file segment download (after someone who failed to do that) up to the point current thread was requested to do)
# TYPE ClickHouseProfileEvents_FileSegmentPredownloadMicroseconds counter
ClickHouseProfileEvents_FileSegmentPredownloadMicroseconds 0
# HELP ClickHouseProfileEvents_FileSegmentUsedBytes Metric per file segment. How many bytes were actually used from current file segment
# TYPE ClickHouseProfileEvents_FileSegmentUsedBytes counter
ClickHouseProfileEvents_FileSegmentUsedBytes 0
# HELP ClickHouseProfileEvents_ReadBufferSeekCancelConnection Number of seeks which lead to new connection (s3, http)
# TYPE ClickHouseProfileEvents_ReadBufferSeekCancelConnection counter
ClickHouseProfileEvents_ReadBufferSeekCancelConnection 0
# HELP ClickHouseProfileEvents_SleepFunctionCalls Number of times a sleep function (sleep, sleepEachRow) has been called.
# TYPE ClickHouseProfileEvents_SleepFunctionCalls counter
ClickHouseProfileEvents_SleepFunctionCalls 0
# HELP ClickHouseProfileEvents_SleepFunctionMicroseconds Time spent sleeping due to a sleep function call.
# TYPE ClickHouseProfileEvents_SleepFunctionMicroseconds counter
ClickHouseProfileEvents_SleepFunctionMicroseconds 0
# HELP ClickHouseProfileEvents_ThreadPoolReaderPageCacheHit Number of times the read inside ThreadPoolReader was done from page cache.
# TYPE ClickHouseProfileEvents_ThreadPoolReaderPageCacheHit counter
ClickHouseProfileEvents_ThreadPoolReaderPageCacheHit 0
# HELP ClickHouseProfileEvents_ThreadPoolReaderPageCacheHitBytes Number of bytes read inside ThreadPoolReader when it was done from page cache.
# TYPE ClickHouseProfileEvents_ThreadPoolReaderPageCacheHitBytes counter
ClickHouseProfileEvents_ThreadPoolReaderPageCacheHitBytes 0
# HELP ClickHouseProfileEvents_ThreadPoolReaderPageCacheHitElapsedMicroseconds Time spent reading data from page cache in ThreadPoolReader.
# TYPE ClickHouseProfileEvents_ThreadPoolReaderPageCacheHitElapsedMicroseconds counter
ClickHouseProfileEvents_ThreadPoolReaderPageCacheHitElapsedMicroseconds 0
# HELP ClickHouseProfileEvents_ThreadPoolReaderPageCacheMiss Number of times the read inside ThreadPoolReader was not done from page cache and was hand off to thread pool.
# TYPE ClickHouseProfileEvents_ThreadPoolReaderPageCacheMiss counter
ClickHouseProfileEvents_ThreadPoolReaderPageCacheMiss 0
# HELP ClickHouseProfileEvents_ThreadPoolReaderPageCacheMissBytes Number of bytes read inside ThreadPoolReader when read was not done from page cache and was hand off to thread pool.
# TYPE ClickHouseProfileEvents_ThreadPoolReaderPageCacheMissBytes counter
ClickHouseProfileEvents_ThreadPoolReaderPageCacheMissBytes 0
# HELP ClickHouseProfileEvents_ThreadPoolReaderPageCacheMissElapsedMicroseconds Time spent reading data inside the asynchronous job in ThreadPoolReader - when read was not done from page cache.
# TYPE ClickHouseProfileEvents_ThreadPoolReaderPageCacheMissElapsedMicroseconds counter
ClickHouseProfileEvents_ThreadPoolReaderPageCacheMissElapsedMicroseconds 0
# HELP ClickHouseProfileEvents_AsynchronousReadWaitMicroseconds Time spent in waiting for asynchronous reads.
# TYPE ClickHouseProfileEvents_AsynchronousReadWaitMicroseconds counter
ClickHouseProfileEvents_AsynchronousReadWaitMicroseconds 0
# HELP ClickHouseProfileEvents_ExternalDataSourceLocalCacheReadBytes Bytes read from local cache buffer in RemoteReadBufferCache
# TYPE ClickHouseProfileEvents_ExternalDataSourceLocalCacheReadBytes counter
ClickHouseProfileEvents_ExternalDataSourceLocalCacheReadBytes 0
# HELP ClickHouseProfileEvents_MainConfigLoads Number of times the main configuration was reloaded.
# TYPE ClickHouseProfileEvents_MainConfigLoads counter
ClickHouseProfileEvents_MainConfigLoads 1
# HELP ClickHouseProfileEvents_AggregationPreallocatedElementsInHashTables How many elements were preallocated in hash tables for aggregation.
# TYPE ClickHouseProfileEvents_AggregationPreallocatedElementsInHashTables counter
ClickHouseProfileEvents_AggregationPreallocatedElementsInHashTables 0
# HELP ClickHouseProfileEvents_AggregationHashTablesInitializedAsTwoLevel How many hash tables were inited as two-level for aggregation.
# TYPE ClickHouseProfileEvents_AggregationHashTablesInitializedAsTwoLevel counter
ClickHouseProfileEvents_AggregationHashTablesInitializedAsTwoLevel 0
# HELP ClickHouseProfileEvents_MergeTreeMetadataCacheGet Number of rocksdb reads(used for merge tree metadata cache)
# TYPE ClickHouseProfileEvents_MergeTreeMetadataCacheGet counter
ClickHouseProfileEvents_MergeTreeMetadataCacheGet 0
# HELP ClickHouseProfileEvents_MergeTreeMetadataCachePut Number of rocksdb puts(used for merge tree metadata cache)
# TYPE ClickHouseProfileEvents_MergeTreeMetadataCachePut counter
ClickHouseProfileEvents_MergeTreeMetadataCachePut 0
# HELP ClickHouseProfileEvents_MergeTreeMetadataCacheDelete Number of rocksdb deletes(used for merge tree metadata cache)
# TYPE ClickHouseProfileEvents_MergeTreeMetadataCacheDelete counter
ClickHouseProfileEvents_MergeTreeMetadataCacheDelete 0
# HELP ClickHouseProfileEvents_MergeTreeMetadataCacheSeek Number of rocksdb seeks(used for merge tree metadata cache)
# TYPE ClickHouseProfileEvents_MergeTreeMetadataCacheSeek counter
ClickHouseProfileEvents_MergeTreeMetadataCacheSeek 0
# HELP ClickHouseProfileEvents_MergeTreeMetadataCacheHit Number of times the read of meta file was done from MergeTree metadata cache
# TYPE ClickHouseProfileEvents_MergeTreeMetadataCacheHit counter
ClickHouseProfileEvents_MergeTreeMetadataCacheHit 0
# HELP ClickHouseProfileEvents_MergeTreeMetadataCacheMiss Number of times the read of meta file was not done from MergeTree metadata cache
# TYPE ClickHouseProfileEvents_MergeTreeMetadataCacheMiss counter
ClickHouseProfileEvents_MergeTreeMetadataCacheMiss 0
# HELP ClickHouseProfileEvents_KafkaRebalanceRevocations Number of partition revocations (the first stage of consumer group rebalance)
# TYPE ClickHouseProfileEvents_KafkaRebalanceRevocations counter
ClickHouseProfileEvents_KafkaRebalanceRevocations 0
# HELP ClickHouseProfileEvents_KafkaRebalanceAssignments Number of partition assignments (the final stage of consumer group rebalance)
# TYPE ClickHouseProfileEvents_KafkaRebalanceAssignments counter
ClickHouseProfileEvents_KafkaRebalanceAssignments 0
# HELP ClickHouseProfileEvents_KafkaRebalanceErrors Number of failed consumer group rebalances
# TYPE ClickHouseProfileEvents_KafkaRebalanceErrors counter
ClickHouseProfileEvents_KafkaRebalanceErrors 0
# HELP ClickHouseProfileEvents_KafkaMessagesPolled Number of Kafka messages polled from librdkafka to ClickHouse
# TYPE ClickHouseProfileEvents_KafkaMessagesPolled counter
ClickHouseProfileEvents_KafkaMessagesPolled 0
# HELP ClickHouseProfileEvents_KafkaMessagesRead Number of Kafka messages already processed by ClickHouse
# TYPE ClickHouseProfileEvents_KafkaMessagesRead counter
ClickHouseProfileEvents_KafkaMessagesRead 0
# HELP ClickHouseProfileEvents_KafkaMessagesFailed Number of Kafka messages ClickHouse failed to parse
# TYPE ClickHouseProfileEvents_KafkaMessagesFailed counter
ClickHouseProfileEvents_KafkaMessagesFailed 0
# HELP ClickHouseProfileEvents_KafkaRowsRead Number of rows parsed from Kafka messages
# TYPE ClickHouseProfileEvents_KafkaRowsRead counter
ClickHouseProfileEvents_KafkaRowsRead 0
# HELP ClickHouseProfileEvents_KafkaRowsRejected Number of parsed rows which were later rejected (due to rebalances / errors or similar reasons). Those rows will be consumed again after the rebalance.
# TYPE ClickHouseProfileEvents_KafkaRowsRejected counter
ClickHouseProfileEvents_KafkaRowsRejected 0
# HELP ClickHouseProfileEvents_KafkaDirectReads Number of direct selects from Kafka tables since server start
# TYPE ClickHouseProfileEvents_KafkaDirectReads counter
ClickHouseProfileEvents_KafkaDirectReads 0
# HELP ClickHouseProfileEvents_KafkaBackgroundReads Number of background reads populating materialized views from Kafka since server start
# TYPE ClickHouseProfileEvents_KafkaBackgroundReads counter
ClickHouseProfileEvents_KafkaBackgroundReads 0
# HELP ClickHouseProfileEvents_KafkaCommits Number of successful commits of consumed offsets to Kafka (normally should be the same as KafkaBackgroundReads)
# TYPE ClickHouseProfileEvents_KafkaCommits counter
ClickHouseProfileEvents_KafkaCommits 0
# HELP ClickHouseProfileEvents_KafkaCommitFailures Number of failed commits of consumed offsets to Kafka (usually is a sign of some data duplication)
# TYPE ClickHouseProfileEvents_KafkaCommitFailures counter
ClickHouseProfileEvents_KafkaCommitFailures 0
# HELP ClickHouseProfileEvents_KafkaConsumerErrors Number of errors reported by librdkafka during polls
# TYPE ClickHouseProfileEvents_KafkaConsumerErrors counter
ClickHouseProfileEvents_KafkaConsumerErrors 0
# HELP ClickHouseProfileEvents_KafkaWrites Number of writes (inserts) to Kafka tables 
# TYPE ClickHouseProfileEvents_KafkaWrites counter
ClickHouseProfileEvents_KafkaWrites 0
# HELP ClickHouseProfileEvents_KafkaRowsWritten Number of rows inserted into Kafka tables
# TYPE ClickHouseProfileEvents_KafkaRowsWritten counter
ClickHouseProfileEvents_KafkaRowsWritten 0
# HELP ClickHouseProfileEvents_KafkaProducerFlushes Number of explicit flushes to Kafka producer
# TYPE ClickHouseProfileEvents_KafkaProducerFlushes counter
ClickHouseProfileEvents_KafkaProducerFlushes 0
# HELP ClickHouseProfileEvents_KafkaMessagesProduced Number of messages produced to Kafka
# TYPE ClickHouseProfileEvents_KafkaMessagesProduced counter
ClickHouseProfileEvents_KafkaMessagesProduced 0
# HELP ClickHouseProfileEvents_KafkaProducerErrors Number of errors during producing the messages to Kafka
# TYPE ClickHouseProfileEvents_KafkaProducerErrors counter
ClickHouseProfileEvents_KafkaProducerErrors 0
# HELP ClickHouseProfileEvents_ScalarSubqueriesGlobalCacheHit Number of times a read from a scalar subquery was done using the global cache
# TYPE ClickHouseProfileEvents_ScalarSubqueriesGlobalCacheHit counter
ClickHouseProfileEvents_ScalarSubqueriesGlobalCacheHit 0
# HELP ClickHouseProfileEvents_ScalarSubqueriesLocalCacheHit Number of times a read from a scalar subquery was done using the local cache
# TYPE ClickHouseProfileEvents_ScalarSubqueriesLocalCacheHit counter
ClickHouseProfileEvents_ScalarSubqueriesLocalCacheHit 0
# HELP ClickHouseProfileEvents_ScalarSubqueriesCacheMiss Number of times a read from a scalar subquery was not cached and had to be calculated completely
# TYPE ClickHouseProfileEvents_ScalarSubqueriesCacheMiss counter
ClickHouseProfileEvents_ScalarSubqueriesCacheMiss 0
# HELP ClickHouseProfileEvents_SchemaInferenceCacheHits Number of times a schema from cache was used for schema inference
# TYPE ClickHouseProfileEvents_SchemaInferenceCacheHits counter
ClickHouseProfileEvents_SchemaInferenceCacheHits 0
# HELP ClickHouseProfileEvents_SchemaInferenceCacheMisses Number of times a schema is not in cache while schema inference
# TYPE ClickHouseProfileEvents_SchemaInferenceCacheMisses counter
ClickHouseProfileEvents_SchemaInferenceCacheMisses 0
# HELP ClickHouseProfileEvents_SchemaInferenceCacheEvictions Number of times a schema from cache was evicted due to overflow
# TYPE ClickHouseProfileEvents_SchemaInferenceCacheEvictions counter
ClickHouseProfileEvents_SchemaInferenceCacheEvictions 0
# HELP ClickHouseProfileEvents_SchemaInferenceCacheInvalidations Number of times a schema in cache became invalid due to changes in data
# TYPE ClickHouseProfileEvents_SchemaInferenceCacheInvalidations counter
ClickHouseProfileEvents_SchemaInferenceCacheInvalidations 0
# HELP ClickHouseProfileEvents_KeeperPacketsSent Packets sent by keeper server
# TYPE ClickHouseProfileEvents_KeeperPacketsSent counter
ClickHouseProfileEvents_KeeperPacketsSent 0
# HELP ClickHouseProfileEvents_KeeperPacketsReceived Packets received by keeper server
# TYPE ClickHouseProfileEvents_KeeperPacketsReceived counter
ClickHouseProfileEvents_KeeperPacketsReceived 0
# HELP ClickHouseProfileEvents_KeeperRequestTotal Total requests number on keeper server
# TYPE ClickHouseProfileEvents_KeeperRequestTotal counter
ClickHouseProfileEvents_KeeperRequestTotal 0
# HELP ClickHouseProfileEvents_KeeperLatency Keeper latency
# TYPE ClickHouseProfileEvents_KeeperLatency counter
ClickHouseProfileEvents_KeeperLatency 0
# HELP ClickHouseProfileEvents_KeeperCommits Number of successful commits
# TYPE ClickHouseProfileEvents_KeeperCommits counter
ClickHouseProfileEvents_KeeperCommits 0
# HELP ClickHouseProfileEvents_KeeperCommitsFailed Number of failed commits
# TYPE ClickHouseProfileEvents_KeeperCommitsFailed counter
ClickHouseProfileEvents_KeeperCommitsFailed 0
# HELP ClickHouseProfileEvents_KeeperSnapshotCreations Number of snapshots creations
# TYPE ClickHouseProfileEvents_KeeperSnapshotCreations counter
ClickHouseProfileEvents_KeeperSnapshotCreations 0
# HELP ClickHouseProfileEvents_KeeperSnapshotCreationsFailed Number of failed snapshot creations
# TYPE ClickHouseProfileEvents_KeeperSnapshotCreationsFailed counter
ClickHouseProfileEvents_KeeperSnapshotCreationsFailed 0
# HELP ClickHouseProfileEvents_KeeperSnapshotApplys Number of snapshot applying
# TYPE ClickHouseProfileEvents_KeeperSnapshotApplys counter
ClickHouseProfileEvents_KeeperSnapshotApplys 0
# HELP ClickHouseProfileEvents_KeeperSnapshotApplysFailed Number of failed snapshot applying
# TYPE ClickHouseProfileEvents_KeeperSnapshotApplysFailed counter
ClickHouseProfileEvents_KeeperSnapshotApplysFailed 0
# HELP ClickHouseProfileEvents_KeeperReadSnapshot Number of snapshot read(serialization)
# TYPE ClickHouseProfileEvents_KeeperReadSnapshot counter
ClickHouseProfileEvents_KeeperReadSnapshot 0
# HELP ClickHouseProfileEvents_KeeperSaveSnapshot Number of snapshot save
# TYPE ClickHouseProfileEvents_KeeperSaveSnapshot counter
ClickHouseProfileEvents_KeeperSaveSnapshot 0
# HELP ClickHouseProfileEvents_OverflowBreak Number of times, data processing was canceled by query complexity limitation with setting '*_overflow_mode' = 'break' and the result is incomplete.
# TYPE ClickHouseProfileEvents_OverflowBreak counter
ClickHouseProfileEvents_OverflowBreak 0
# HELP ClickHouseProfileEvents_OverflowThrow Number of times, data processing was canceled by query complexity limitation with setting '*_overflow_mode' = 'throw' and exception was thrown.
# TYPE ClickHouseProfileEvents_OverflowThrow counter
ClickHouseProfileEvents_OverflowThrow 0
# HELP ClickHouseProfileEvents_OverflowAny Number of times approximate GROUP BY was in effect: when aggregation was performed only on top of first 'max_rows_to_group_by' unique keys and other keys were ignored due to 'group_by_overflow_mode' = 'any'.
# TYPE ClickHouseProfileEvents_OverflowAny counter
ClickHouseProfileEvents_OverflowAny 0
# HELP ClickHouseMetrics_Query Number of executing queries
# TYPE ClickHouseMetrics_Query gauge
ClickHouseMetrics_Query 0
# HELP ClickHouseMetrics_Merge Number of executing background merges
# TYPE ClickHouseMetrics_Merge gauge
ClickHouseMetrics_Merge 0
# HELP ClickHouseMetrics_PartMutation Number of mutations (ALTER DELETE/UPDATE)
# TYPE ClickHouseMetrics_PartMutation gauge
ClickHouseMetrics_PartMutation 0
# HELP ClickHouseMetrics_ReplicatedFetch Number of data parts being fetched from replica
# TYPE ClickHouseMetrics_ReplicatedFetch gauge
ClickHouseMetrics_ReplicatedFetch 0
# HELP ClickHouseMetrics_ReplicatedSend Number of data parts being sent to replicas
# TYPE ClickHouseMetrics_ReplicatedSend gauge
ClickHouseMetrics_ReplicatedSend 0
# HELP ClickHouseMetrics_ReplicatedChecks Number of data parts checking for consistency
# TYPE ClickHouseMetrics_ReplicatedChecks gauge
ClickHouseMetrics_ReplicatedChecks 0
# HELP ClickHouseMetrics_BackgroundMergesAndMutationsPoolTask Number of active merges and mutations in an associated background pool
# TYPE ClickHouseMetrics_BackgroundMergesAndMutationsPoolTask gauge
ClickHouseMetrics_BackgroundMergesAndMutationsPoolTask 0
# HELP ClickHouseMetrics_BackgroundFetchesPoolTask Number of active fetches in an associated background pool
# TYPE ClickHouseMetrics_BackgroundFetchesPoolTask gauge
ClickHouseMetrics_BackgroundFetchesPoolTask 0
# HELP ClickHouseMetrics_BackgroundCommonPoolTask Number of active tasks in an associated background pool
# TYPE ClickHouseMetrics_BackgroundCommonPoolTask gauge
ClickHouseMetrics_BackgroundCommonPoolTask 0
# HELP ClickHouseMetrics_BackgroundMovePoolTask Number of active tasks in BackgroundProcessingPool for moves
# TYPE ClickHouseMetrics_BackgroundMovePoolTask gauge
ClickHouseMetrics_BackgroundMovePoolTask 0
# HELP ClickHouseMetrics_BackgroundSchedulePoolTask Number of active tasks in BackgroundSchedulePool. This pool is used for periodic ReplicatedMergeTree tasks, like cleaning old data parts, altering data parts, replica re-initialization, etc.
# TYPE ClickHouseMetrics_BackgroundSchedulePoolTask gauge
ClickHouseMetrics_BackgroundSchedulePoolTask 0
# HELP ClickHouseMetrics_BackgroundBufferFlushSchedulePoolTask Number of active tasks in BackgroundBufferFlushSchedulePool. This pool is used for periodic Buffer flushes
# TYPE ClickHouseMetrics_BackgroundBufferFlushSchedulePoolTask gauge
ClickHouseMetrics_BackgroundBufferFlushSchedulePoolTask 0
# HELP ClickHouseMetrics_BackgroundDistributedSchedulePoolTask Number of active tasks in BackgroundDistributedSchedulePool. This pool is used for distributed sends that is done in background.
# TYPE ClickHouseMetrics_BackgroundDistributedSchedulePoolTask gauge
ClickHouseMetrics_BackgroundDistributedSchedulePoolTask 0
# HELP ClickHouseMetrics_BackgroundMessageBrokerSchedulePoolTask Number of active tasks in BackgroundProcessingPool for message streaming
# TYPE ClickHouseMetrics_BackgroundMessageBrokerSchedulePoolTask gauge
ClickHouseMetrics_BackgroundMessageBrokerSchedulePoolTask 0
# HELP ClickHouseMetrics_CacheDictionaryUpdateQueueBatches Number of 'batches' (a set of keys) in update queue in CacheDictionaries.
# TYPE ClickHouseMetrics_CacheDictionaryUpdateQueueBatches gauge
ClickHouseMetrics_CacheDictionaryUpdateQueueBatches 0
# HELP ClickHouseMetrics_CacheDictionaryUpdateQueueKeys Exact number of keys in update queue in CacheDictionaries.
# TYPE ClickHouseMetrics_CacheDictionaryUpdateQueueKeys gauge
ClickHouseMetrics_CacheDictionaryUpdateQueueKeys 0
# HELP ClickHouseMetrics_DiskSpaceReservedForMerge Disk space reserved for currently running background merges. It is slightly more than the total size of currently merging parts.
# TYPE ClickHouseMetrics_DiskSpaceReservedForMerge gauge
ClickHouseMetrics_DiskSpaceReservedForMerge 0
# HELP ClickHouseMetrics_DistributedSend Number of connections to remote servers sending data that was INSERTed into Distributed tables. Both synchronous and asynchronous mode.
# TYPE ClickHouseMetrics_DistributedSend gauge
ClickHouseMetrics_DistributedSend 0
# HELP ClickHouseMetrics_QueryPreempted Number of queries that are stopped and waiting due to 'priority' setting.
# TYPE ClickHouseMetrics_QueryPreempted gauge
ClickHouseMetrics_QueryPreempted 0
# HELP ClickHouseMetrics_TCPConnection Number of connections to TCP server (clients with native interface), also included server-server distributed query connections
# TYPE ClickHouseMetrics_TCPConnection gauge
ClickHouseMetrics_TCPConnection 0
# HELP ClickHouseMetrics_MySQLConnection Number of client connections using MySQL protocol
# TYPE ClickHouseMetrics_MySQLConnection gauge
ClickHouseMetrics_MySQLConnection 0
# HELP ClickHouseMetrics_HTTPConnection Number of connections to HTTP server
# TYPE ClickHouseMetrics_HTTPConnection gauge
ClickHouseMetrics_HTTPConnection 0
# HELP ClickHouseMetrics_InterserverConnection Number of connections from other replicas to fetch parts
# TYPE ClickHouseMetrics_InterserverConnection gauge
ClickHouseMetrics_InterserverConnection 0
# HELP ClickHouseMetrics_PostgreSQLConnection Number of client connections using PostgreSQL protocol
# TYPE ClickHouseMetrics_PostgreSQLConnection gauge
ClickHouseMetrics_PostgreSQLConnection 0
# HELP ClickHouseMetrics_OpenFileForRead Number of files open for reading
# TYPE ClickHouseMetrics_OpenFileForRead gauge
ClickHouseMetrics_OpenFileForRead 29
# HELP ClickHouseMetrics_OpenFileForWrite Number of files open for writing
# TYPE ClickHouseMetrics_OpenFileForWrite gauge
ClickHouseMetrics_OpenFileForWrite 0
# HELP ClickHouseMetrics_Read Number of read (read, pread, io_getevents, etc.) syscalls in fly
# TYPE ClickHouseMetrics_Read gauge
ClickHouseMetrics_Read 2
# HELP ClickHouseMetrics_Write Number of write (write, pwrite, io_getevents, etc.) syscalls in fly
# TYPE ClickHouseMetrics_Write gauge
ClickHouseMetrics_Write 0
# HELP ClickHouseMetrics_NetworkReceive Number of threads receiving data from network. Only ClickHouse-related network interaction is included, not by 3rd party libraries.
# TYPE ClickHouseMetrics_NetworkReceive gauge
ClickHouseMetrics_NetworkReceive 0
# HELP ClickHouseMetrics_NetworkSend Number of threads sending data to network. Only ClickHouse-related network interaction is included, not by 3rd party libraries.
# TYPE ClickHouseMetrics_NetworkSend gauge
ClickHouseMetrics_NetworkSend 0
# HELP ClickHouseMetrics_SendScalars Number of connections that are sending data for scalars to remote servers.
# TYPE ClickHouseMetrics_SendScalars gauge
ClickHouseMetrics_SendScalars 0
# HELP ClickHouseMetrics_SendExternalTables Number of connections that are sending data for external tables to remote servers. External tables are used to implement GLOBAL IN and GLOBAL JOIN operators with distributed subqueries.
# TYPE ClickHouseMetrics_SendExternalTables gauge
ClickHouseMetrics_SendExternalTables 0
# HELP ClickHouseMetrics_QueryThread Number of query processing threads
# TYPE ClickHouseMetrics_QueryThread gauge
ClickHouseMetrics_QueryThread 0
# HELP ClickHouseMetrics_ReadonlyReplica Number of Replicated tables that are currently in readonly state due to re-initialization after ZooKeeper session loss or due to startup without ZooKeeper configured.
# TYPE ClickHouseMetrics_ReadonlyReplica gauge
ClickHouseMetrics_ReadonlyReplica 0
# HELP ClickHouseMetrics_MemoryTracking Total amount of memory (bytes) allocated by the server.
# TYPE ClickHouseMetrics_MemoryTracking gauge
ClickHouseMetrics_MemoryTracking 578779615
# HELP ClickHouseMetrics_EphemeralNode Number of ephemeral nodes hold in ZooKeeper.
# TYPE ClickHouseMetrics_EphemeralNode gauge
ClickHouseMetrics_EphemeralNode 0
# HELP ClickHouseMetrics_ZooKeeperSession Number of sessions (connections) to ZooKeeper. Should be no more than one, because using more than one connection to ZooKeeper may lead to bugs due to lack of linearizability (stale reads) that ZooKeeper consistency model allows.
# TYPE ClickHouseMetrics_ZooKeeperSession gauge
ClickHouseMetrics_ZooKeeperSession 0
# HELP ClickHouseMetrics_ZooKeeperWatch Number of watches (event subscriptions) in ZooKeeper.
# TYPE ClickHouseMetrics_ZooKeeperWatch gauge
ClickHouseMetrics_ZooKeeperWatch 0
# HELP ClickHouseMetrics_ZooKeeperRequest Number of requests to ZooKeeper in fly.
# TYPE ClickHouseMetrics_ZooKeeperRequest gauge
ClickHouseMetrics_ZooKeeperRequest 0
# HELP ClickHouseMetrics_DelayedInserts Number of INSERT queries that are throttled due to high number of active data parts for partition in a MergeTree table.
# TYPE ClickHouseMetrics_DelayedInserts gauge
ClickHouseMetrics_DelayedInserts 0
# HELP ClickHouseMetrics_ContextLockWait Number of threads waiting for lock in Context. This is global lock.
# TYPE ClickHouseMetrics_ContextLockWait gauge
ClickHouseMetrics_ContextLockWait 0
# HELP ClickHouseMetrics_StorageBufferRows Number of rows in buffers of Buffer tables
# TYPE ClickHouseMetrics_StorageBufferRows gauge
ClickHouseMetrics_StorageBufferRows 0
# HELP ClickHouseMetrics_StorageBufferBytes Number of bytes in buffers of Buffer tables
# TYPE ClickHouseMetrics_StorageBufferBytes gauge
ClickHouseMetrics_StorageBufferBytes 0
# HELP ClickHouseMetrics_DictCacheRequests Number of requests in fly to data sources of dictionaries of cache type.
# TYPE ClickHouseMetrics_DictCacheRequests gauge
ClickHouseMetrics_DictCacheRequests 0
# HELP ClickHouseMetrics_Revision Revision of the server. It is a number incremented for every release or release candidate except patch releases.
# TYPE ClickHouseMetrics_Revision gauge
ClickHouseMetrics_Revision 54465
# HELP ClickHouseMetrics_VersionInteger Version of the server in a single integer number in base-1000. For example, version 11.22.33 is translated to 11022033.
# TYPE ClickHouseMetrics_VersionInteger gauge
ClickHouseMetrics_VersionInteger 22008015
# HELP ClickHouseMetrics_RWLockWaitingReaders Number of threads waiting for read on a table RWLock.
# TYPE ClickHouseMetrics_RWLockWaitingReaders gauge
ClickHouseMetrics_RWLockWaitingReaders 0
# HELP ClickHouseMetrics_RWLockWaitingWriters Number of threads waiting for write on a table RWLock.
# TYPE ClickHouseMetrics_RWLockWaitingWriters gauge
ClickHouseMetrics_RWLockWaitingWriters 0
# HELP ClickHouseMetrics_RWLockActiveReaders Number of threads holding read lock in a table RWLock.
# TYPE ClickHouseMetrics_RWLockActiveReaders gauge
ClickHouseMetrics_RWLockActiveReaders 0
# HELP ClickHouseMetrics_RWLockActiveWriters Number of threads holding write lock in a table RWLock.
# TYPE ClickHouseMetrics_RWLockActiveWriters gauge
ClickHouseMetrics_RWLockActiveWriters 0
# HELP ClickHouseMetrics_GlobalThread Number of threads in global thread pool.
# TYPE ClickHouseMetrics_GlobalThread gauge
ClickHouseMetrics_GlobalThread 189
# HELP ClickHouseMetrics_GlobalThreadActive Number of threads in global thread pool running a task.
# TYPE ClickHouseMetrics_GlobalThreadActive gauge
ClickHouseMetrics_GlobalThreadActive 189
# HELP ClickHouseMetrics_LocalThread Number of threads in local thread pools. The threads in local thread pools are taken from the global thread pool.
# TYPE ClickHouseMetrics_LocalThread gauge
ClickHouseMetrics_LocalThread 40
# HELP ClickHouseMetrics_LocalThreadActive Number of threads in local thread pools running a task.
# TYPE ClickHouseMetrics_LocalThreadActive gauge
ClickHouseMetrics_LocalThreadActive 40
# HELP ClickHouseMetrics_DistributedFilesToInsert Number of pending files to process for asynchronous insertion into Distributed tables. Number of files for every shard is summed.
# TYPE ClickHouseMetrics_DistributedFilesToInsert gauge
ClickHouseMetrics_DistributedFilesToInsert 0
# HELP ClickHouseMetrics_BrokenDistributedFilesToInsert Number of files for asynchronous insertion into Distributed tables that has been marked as broken. This metric will starts from 0 on start. Number of files for every shard is summed.
# TYPE ClickHouseMetrics_BrokenDistributedFilesToInsert gauge
ClickHouseMetrics_BrokenDistributedFilesToInsert 0
# HELP ClickHouseMetrics_TablesToDropQueueSize Number of dropped tables, that are waiting for background data removal.
# TYPE ClickHouseMetrics_TablesToDropQueueSize gauge
ClickHouseMetrics_TablesToDropQueueSize 0
# HELP ClickHouseMetrics_MaxDDLEntryID Max processed DDL entry of DDLWorker.
# TYPE ClickHouseMetrics_MaxDDLEntryID gauge
ClickHouseMetrics_MaxDDLEntryID 0
# HELP ClickHouseMetrics_MaxPushedDDLEntryID Max DDL entry of DDLWorker that pushed to zookeeper.
# TYPE ClickHouseMetrics_MaxPushedDDLEntryID gauge
ClickHouseMetrics_MaxPushedDDLEntryID 0
# HELP ClickHouseMetrics_PartsTemporary The part is generating now, it is not in data_parts list.
# TYPE ClickHouseMetrics_PartsTemporary gauge
ClickHouseMetrics_PartsTemporary 0
# HELP ClickHouseMetrics_PartsPreCommitted Deprecated. See PartsPreActive.
# TYPE ClickHouseMetrics_PartsPreCommitted gauge
ClickHouseMetrics_PartsPreCommitted 0
# HELP ClickHouseMetrics_PartsCommitted Deprecated. See PartsActive.
# TYPE ClickHouseMetrics_PartsCommitted gauge
ClickHouseMetrics_PartsCommitted 19
# HELP ClickHouseMetrics_PartsPreActive The part is in data_parts, but not used for SELECTs.
# TYPE ClickHouseMetrics_PartsPreActive gauge
ClickHouseMetrics_PartsPreActive 0
# HELP ClickHouseMetrics_PartsActive Active data part, used by current and upcoming SELECTs.
# TYPE ClickHouseMetrics_PartsActive gauge
ClickHouseMetrics_PartsActive 19
# HELP ClickHouseMetrics_PartsOutdated Not active data part, but could be used by only current SELECTs, could be deleted after SELECTs finishes.
# TYPE ClickHouseMetrics_PartsOutdated gauge
ClickHouseMetrics_PartsOutdated 18
# HELP ClickHouseMetrics_PartsDeleting Not active data part with identity refcounter, it is deleting right now by a cleaner.
# TYPE ClickHouseMetrics_PartsDeleting gauge
ClickHouseMetrics_PartsDeleting 0
# HELP ClickHouseMetrics_PartsDeleteOnDestroy Part was moved to another disk and should be deleted in own destructor.
# TYPE ClickHouseMetrics_PartsDeleteOnDestroy gauge
ClickHouseMetrics_PartsDeleteOnDestroy 0
# HELP ClickHouseMetrics_PartsWide Wide parts.
# TYPE ClickHouseMetrics_PartsWide gauge
ClickHouseMetrics_PartsWide 10
# HELP ClickHouseMetrics_PartsCompact Compact parts.
# TYPE ClickHouseMetrics_PartsCompact gauge
ClickHouseMetrics_PartsCompact 27
# HELP ClickHouseMetrics_PartsInMemory In-memory parts.
# TYPE ClickHouseMetrics_PartsInMemory gauge
ClickHouseMetrics_PartsInMemory 0
# HELP ClickHouseMetrics_MMappedFiles Total number of mmapped files.
# TYPE ClickHouseMetrics_MMappedFiles gauge
ClickHouseMetrics_MMappedFiles 1
# HELP ClickHouseMetrics_MMappedFileBytes Sum size of mmapped file regions.
# TYPE ClickHouseMetrics_MMappedFileBytes gauge
ClickHouseMetrics_MMappedFileBytes 461388904
# HELP ClickHouseMetrics_AsyncDrainedConnections Number of connections drained asynchronously.
# TYPE ClickHouseMetrics_AsyncDrainedConnections gauge
ClickHouseMetrics_AsyncDrainedConnections 0
# HELP ClickHouseMetrics_ActiveAsyncDrainedConnections Number of active connections drained asynchronously.
# TYPE ClickHouseMetrics_ActiveAsyncDrainedConnections gauge
ClickHouseMetrics_ActiveAsyncDrainedConnections 0
# HELP ClickHouseMetrics_SyncDrainedConnections Number of connections drained synchronously.
# TYPE ClickHouseMetrics_SyncDrainedConnections gauge
ClickHouseMetrics_SyncDrainedConnections 0
# HELP ClickHouseMetrics_ActiveSyncDrainedConnections Number of active connections drained synchronously.
# TYPE ClickHouseMetrics_ActiveSyncDrainedConnections gauge
ClickHouseMetrics_ActiveSyncDrainedConnections 0
# HELP ClickHouseMetrics_AsynchronousReadWait Number of threads waiting for asynchronous read.
# TYPE ClickHouseMetrics_AsynchronousReadWait gauge
ClickHouseMetrics_AsynchronousReadWait 0
# HELP ClickHouseMetrics_PendingAsyncInsert Number of asynchronous inserts that are waiting for flush.
# TYPE ClickHouseMetrics_PendingAsyncInsert gauge
ClickHouseMetrics_PendingAsyncInsert 0
# HELP ClickHouseMetrics_KafkaConsumers Number of active Kafka consumers
# TYPE ClickHouseMetrics_KafkaConsumers gauge
ClickHouseMetrics_KafkaConsumers 0
# HELP ClickHouseMetrics_KafkaConsumersWithAssignment Number of active Kafka consumers which have some partitions assigned.
# TYPE ClickHouseMetrics_KafkaConsumersWithAssignment gauge
ClickHouseMetrics_KafkaConsumersWithAssignment 0
# HELP ClickHouseMetrics_KafkaProducers Number of active Kafka producer created
# TYPE ClickHouseMetrics_KafkaProducers gauge
ClickHouseMetrics_KafkaProducers 0
# HELP ClickHouseMetrics_KafkaLibrdkafkaThreads Number of active librdkafka threads
# TYPE ClickHouseMetrics_KafkaLibrdkafkaThreads gauge
ClickHouseMetrics_KafkaLibrdkafkaThreads 0
# HELP ClickHouseMetrics_KafkaBackgroundReads Number of background reads currently working (populating materialized views from Kafka)
# TYPE ClickHouseMetrics_KafkaBackgroundReads gauge
ClickHouseMetrics_KafkaBackgroundReads 0
# HELP ClickHouseMetrics_KafkaConsumersInUse Number of consumers which are currently used by direct or background reads
# TYPE ClickHouseMetrics_KafkaConsumersInUse gauge
ClickHouseMetrics_KafkaConsumersInUse 0
# HELP ClickHouseMetrics_KafkaWrites Number of currently running inserts to Kafka
# TYPE ClickHouseMetrics_KafkaWrites gauge
ClickHouseMetrics_KafkaWrites 0
# HELP ClickHouseMetrics_KafkaAssignedPartitions Number of partitions Kafka tables currently assigned to
# TYPE ClickHouseMetrics_KafkaAssignedPartitions gauge
ClickHouseMetrics_KafkaAssignedPartitions 0
# HELP ClickHouseMetrics_FilesystemCacheReadBuffers Number of active cache buffers
# TYPE ClickHouseMetrics_FilesystemCacheReadBuffers gauge
ClickHouseMetrics_FilesystemCacheReadBuffers 0
# HELP ClickHouseMetrics_CacheFileSegments Number of existing cache file segments
# TYPE ClickHouseMetrics_CacheFileSegments gauge
ClickHouseMetrics_CacheFileSegments 0
# HELP ClickHouseMetrics_CacheDetachedFileSegments Number of existing detached cache file segments
# TYPE ClickHouseMetrics_CacheDetachedFileSegments gauge
ClickHouseMetrics_CacheDetachedFileSegments 0
# HELP ClickHouseMetrics_FilesystemCacheSize Filesystem cache size in bytes
# TYPE ClickHouseMetrics_FilesystemCacheSize gauge
ClickHouseMetrics_FilesystemCacheSize 0
# HELP ClickHouseMetrics_FilesystemCacheElements Filesystem cache elements (file segments)
# TYPE ClickHouseMetrics_FilesystemCacheElements gauge
ClickHouseMetrics_FilesystemCacheElements 0
# HELP ClickHouseMetrics_S3Requests S3 requests
# TYPE ClickHouseMetrics_S3Requests gauge
ClickHouseMetrics_S3Requests 0
# HELP ClickHouseMetrics_KeeperAliveConnections Number of alive connections
# TYPE ClickHouseMetrics_KeeperAliveConnections gauge
ClickHouseMetrics_KeeperAliveConnections 0
# HELP ClickHouseMetrics_KeeperOutstandingRequets Number of outstanding requests
# TYPE ClickHouseMetrics_KeeperOutstandingRequets gauge
ClickHouseMetrics_KeeperOutstandingRequets 0
# TYPE ClickHouseAsyncMetrics_AsynchronousMetricsCalculationTimeSpent gauge
ClickHouseAsyncMetrics_AsynchronousMetricsCalculationTimeSpent 0.02115035
# TYPE ClickHouseAsyncMetrics_jemalloc_arenas_all_muzzy_purged gauge
ClickHouseAsyncMetrics_jemalloc_arenas_all_muzzy_purged 78981
# TYPE ClickHouseAsyncMetrics_jemalloc_arenas_all_dirty_purged gauge
ClickHouseAsyncMetrics_jemalloc_arenas_all_dirty_purged 299380
# TYPE ClickHouseAsyncMetrics_jemalloc_arenas_all_pmuzzy gauge
ClickHouseAsyncMetrics_jemalloc_arenas_all_pmuzzy 217673
# TYPE ClickHouseAsyncMetrics_jemalloc_background_thread_run_intervals gauge
ClickHouseAsyncMetrics_jemalloc_background_thread_run_intervals 0
# TYPE ClickHouseAsyncMetrics_jemalloc_metadata_thp gauge
ClickHouseAsyncMetrics_jemalloc_metadata_thp 0
# TYPE ClickHouseAsyncMetrics_jemalloc_metadata gauge
ClickHouseAsyncMetrics_jemalloc_metadata 16048160
# TYPE ClickHouseAsyncMetrics_jemalloc_allocated gauge
ClickHouseAsyncMetrics_jemalloc_allocated 68771112
# TYPE ClickHouseAsyncMetrics_PostgreSQLThreads gauge
ClickHouseAsyncMetrics_PostgreSQLThreads 0
# TYPE ClickHouseAsyncMetrics_TCPThreads gauge
ClickHouseAsyncMetrics_TCPThreads 0
# TYPE ClickHouseAsyncMetrics_HTTPThreads gauge
ClickHouseAsyncMetrics_HTTPThreads 0
# TYPE ClickHouseAsyncMetrics_TotalPartsOfMergeTreeTables gauge
ClickHouseAsyncMetrics_TotalPartsOfMergeTreeTables 19
# TYPE ClickHouseAsyncMetrics_NumberOfTables gauge
ClickHouseAsyncMetrics_NumberOfTables 85
# TYPE ClickHouseAsyncMetrics_NumberOfDatabases gauge
ClickHouseAsyncMetrics_NumberOfDatabases 5
# TYPE ClickHouseAsyncMetrics_ReplicasMaxRelativeDelay gauge
ClickHouseAsyncMetrics_ReplicasMaxRelativeDelay 0
# TYPE ClickHouseAsyncMetrics_ReplicasMaxAbsoluteDelay gauge
ClickHouseAsyncMetrics_ReplicasMaxAbsoluteDelay 0
# TYPE ClickHouseAsyncMetrics_ReplicasSumMergesInQueue gauge
ClickHouseAsyncMetrics_ReplicasSumMergesInQueue 0
# TYPE ClickHouseAsyncMetrics_ReplicasSumInsertsInQueue gauge
ClickHouseAsyncMetrics_ReplicasSumInsertsInQueue 0
# TYPE ClickHouseAsyncMetrics_ReplicasSumQueueSize gauge
ClickHouseAsyncMetrics_ReplicasSumQueueSize 0
# TYPE ClickHouseAsyncMetrics_ReplicasMaxMergesInQueue gauge
ClickHouseAsyncMetrics_ReplicasMaxMergesInQueue 0
# TYPE ClickHouseAsyncMetrics_ReplicasMaxInsertsInQueue gauge
ClickHouseAsyncMetrics_ReplicasMaxInsertsInQueue 0
# TYPE ClickHouseAsyncMetrics_ReplicasMaxQueueSize gauge
ClickHouseAsyncMetrics_ReplicasMaxQueueSize 0
# TYPE ClickHouseAsyncMetrics_DiskUnreserved_default gauge
ClickHouseAsyncMetrics_DiskUnreserved_default 764679319552
# TYPE ClickHouseAsyncMetrics_DiskUsed_default gauge
ClickHouseAsyncMetrics_DiskUsed_default 218126905344
# TYPE ClickHouseAsyncMetrics_FilesystemLogsPathUsedINodes gauge
ClickHouseAsyncMetrics_FilesystemLogsPathUsedINodes 1690184
# TYPE ClickHouseAsyncMetrics_jemalloc_arenas_all_pactive gauge
ClickHouseAsyncMetrics_jemalloc_arenas_all_pactive 18190
# TYPE ClickHouseAsyncMetrics_FilesystemLogsPathTotalINodes gauge
ClickHouseAsyncMetrics_FilesystemLogsPathTotalINodes 61014016
# TYPE ClickHouseAsyncMetrics_FilesystemLogsPathUsedBytes gauge
ClickHouseAsyncMetrics_FilesystemLogsPathUsedBytes 218126905344
# TYPE ClickHouseAsyncMetrics_DiskTotal_default gauge
ClickHouseAsyncMetrics_DiskTotal_default 982806224896
# TYPE ClickHouseAsyncMetrics_FilesystemLogsPathAvailableBytes gauge
ClickHouseAsyncMetrics_FilesystemLogsPathAvailableBytes 764679319552
# TYPE ClickHouseAsyncMetrics_FilesystemLogsPathTotalBytes gauge
ClickHouseAsyncMetrics_FilesystemLogsPathTotalBytes 982806224896
# TYPE ClickHouseAsyncMetrics_FilesystemMainPathAvailableINodes gauge
ClickHouseAsyncMetrics_FilesystemMainPathAvailableINodes 59323832
# TYPE ClickHouseAsyncMetrics_FilesystemMainPathUsedBytes gauge
ClickHouseAsyncMetrics_FilesystemMainPathUsedBytes 218126905344
# TYPE ClickHouseAsyncMetrics_FilesystemMainPathAvailableBytes gauge
ClickHouseAsyncMetrics_FilesystemMainPathAvailableBytes 764679319552
# TYPE ClickHouseAsyncMetrics_FilesystemMainPathTotalBytes gauge
ClickHouseAsyncMetrics_FilesystemMainPathTotalBytes 982806224896
# TYPE ClickHouseAsyncMetrics_jemalloc_resident gauge
ClickHouseAsyncMetrics_jemalloc_resident 187969536
# TYPE ClickHouseAsyncMetrics_Temperature_nvme_Sensor_1 gauge
ClickHouseAsyncMetrics_Temperature_nvme_Sensor_1 32.85
# TYPE ClickHouseAsyncMetrics_OSNiceTimeCPU5 gauge
ClickHouseAsyncMetrics_OSNiceTimeCPU5 0
# TYPE ClickHouseAsyncMetrics_Temperature_coretemp_Package_id_0 gauge
ClickHouseAsyncMetrics_Temperature_coretemp_Package_id_0 57
# TYPE ClickHouseAsyncMetrics_Temperature_coretemp_Core_1 gauge
ClickHouseAsyncMetrics_Temperature_coretemp_Core_1 57
# TYPE ClickHouseAsyncMetrics_Temperature7 gauge
ClickHouseAsyncMetrics_Temperature7 55
# TYPE ClickHouseAsyncMetrics_jemalloc_active gauge
ClickHouseAsyncMetrics_jemalloc_active 74506240
# TYPE ClickHouseAsyncMetrics_Temperature3 gauge
ClickHouseAsyncMetrics_Temperature3 40.050000000000004
# TYPE ClickHouseAsyncMetrics_OSIrqTimeNormalized gauge
ClickHouseAsyncMetrics_OSIrqTimeNormalized 0
# TYPE ClickHouseAsyncMetrics_Temperature1 gauge
ClickHouseAsyncMetrics_Temperature1 48.050000000000004
# TYPE ClickHouseAsyncMetrics_Temperature0 gauge
ClickHouseAsyncMetrics_Temperature0 53
# TYPE ClickHouseAsyncMetrics_CPUFrequencyMHz_1 gauge
ClickHouseAsyncMetrics_CPUFrequencyMHz_1 2400
# TYPE ClickHouseAsyncMetrics_NetworkSendDrop_eth0 gauge
ClickHouseAsyncMetrics_NetworkSendDrop_eth0 0
# TYPE ClickHouseAsyncMetrics_NetworkSendPackets_eth0 gauge
ClickHouseAsyncMetrics_NetworkSendPackets_eth0 0
# TYPE ClickHouseAsyncMetrics_NetworkReceiveDrop_eth0 gauge
ClickHouseAsyncMetrics_NetworkReceiveDrop_eth0 0
# TYPE ClickHouseAsyncMetrics_NetworkSendErrors_eth0 gauge
ClickHouseAsyncMetrics_NetworkSendErrors_eth0 0
# TYPE ClickHouseAsyncMetrics_NetworkReceiveErrors_eth0 gauge
ClickHouseAsyncMetrics_NetworkReceiveErrors_eth0 0
# TYPE ClickHouseAsyncMetrics_NetworkReceivePackets_eth0 gauge
ClickHouseAsyncMetrics_NetworkReceivePackets_eth0 0
# TYPE ClickHouseAsyncMetrics_BlockActiveTime_nvme0n1 gauge
ClickHouseAsyncMetrics_BlockActiveTime_nvme0n1 0
# TYPE ClickHouseAsyncMetrics_BlockDiscardTime_nvme0n1 gauge
ClickHouseAsyncMetrics_BlockDiscardTime_nvme0n1 0
# TYPE ClickHouseAsyncMetrics_BlockWriteTime_nvme0n1 gauge
ClickHouseAsyncMetrics_BlockWriteTime_nvme0n1 0
# TYPE ClickHouseAsyncMetrics_Temperature_pch_cannonlake gauge
ClickHouseAsyncMetrics_Temperature_pch_cannonlake 52
# TYPE ClickHouseAsyncMetrics_BlockReadBytes_nvme0n1 gauge
ClickHouseAsyncMetrics_BlockReadBytes_nvme0n1 0
# TYPE ClickHouseAsyncMetrics_jemalloc_mapped gauge
ClickHouseAsyncMetrics_jemalloc_mapped 1104584704
# TYPE ClickHouseAsyncMetrics_BlockReadMerges_nvme0n1 gauge
ClickHouseAsyncMetrics_BlockReadMerges_nvme0n1 0
# TYPE ClickHouseAsyncMetrics_BlockDiscardOps_nvme0n1 gauge
ClickHouseAsyncMetrics_BlockDiscardOps_nvme0n1 0
# TYPE ClickHouseAsyncMetrics_BlockQueueTime_sda gauge
ClickHouseAsyncMetrics_BlockQueueTime_sda 0.000018
# TYPE ClickHouseAsyncMetrics_BlockInFlightOps_sda gauge
ClickHouseAsyncMetrics_BlockInFlightOps_sda 0
# TYPE ClickHouseAsyncMetrics_BlockWriteTime_sda gauge
ClickHouseAsyncMetrics_BlockWriteTime_sda 0.000012
# TYPE ClickHouseAsyncMetrics_Temperature_nvme_Composite gauge
ClickHouseAsyncMetrics_Temperature_nvme_Composite 32.85
# TYPE ClickHouseAsyncMetrics_BlockReadTime_sda gauge
ClickHouseAsyncMetrics_BlockReadTime_sda 0.0000049999999999999996
# TYPE ClickHouseAsyncMetrics_MemoryResident gauge
ClickHouseAsyncMetrics_MemoryResident 577699840
# TYPE ClickHouseAsyncMetrics_OSSoftIrqTimeCPU5 gauge
ClickHouseAsyncMetrics_OSSoftIrqTimeCPU5 0
# TYPE ClickHouseAsyncMetrics_BlockDiscardMerges_sda gauge
ClickHouseAsyncMetrics_BlockDiscardMerges_sda 0
# TYPE ClickHouseAsyncMetrics_BlockDiscardOps_sda gauge
ClickHouseAsyncMetrics_BlockDiscardOps_sda 0
# TYPE ClickHouseAsyncMetrics_BlockWriteOps_sda gauge
ClickHouseAsyncMetrics_BlockWriteOps_sda 7
# TYPE ClickHouseAsyncMetrics_BlockReadOps_sda gauge
ClickHouseAsyncMetrics_BlockReadOps_sda 7
# TYPE ClickHouseAsyncMetrics_jemalloc_epoch gauge
ClickHouseAsyncMetrics_jemalloc_epoch 29
# TYPE ClickHouseAsyncMetrics_OSOpenFiles gauge
ClickHouseAsyncMetrics_OSOpenFiles 18176
# TYPE ClickHouseAsyncMetrics_OSIrqTimeCPU5 gauge
ClickHouseAsyncMetrics_OSIrqTimeCPU5 0
# TYPE ClickHouseAsyncMetrics_BlockQueueTime_nvme0n1 gauge
ClickHouseAsyncMetrics_BlockQueueTime_nvme0n1 0
# TYPE ClickHouseAsyncMetrics_Temperature5 gauge
ClickHouseAsyncMetrics_Temperature5 53.050000000000004
# TYPE ClickHouseAsyncMetrics_DiskAvailable_default gauge
ClickHouseAsyncMetrics_DiskAvailable_default 764679319552
# TYPE ClickHouseAsyncMetrics_CPUFrequencyMHz_7 gauge
ClickHouseAsyncMetrics_CPUFrequencyMHz_7 2400
# TYPE ClickHouseAsyncMetrics_BlockActiveTime_sda gauge
ClickHouseAsyncMetrics_BlockActiveTime_sda 0.000032
# TYPE ClickHouseAsyncMetrics_CPUFrequencyMHz_6 gauge
ClickHouseAsyncMetrics_CPUFrequencyMHz_6 2400
# TYPE ClickHouseAsyncMetrics_CPUFrequencyMHz_4 gauge
ClickHouseAsyncMetrics_CPUFrequencyMHz_4 2400
# TYPE ClickHouseAsyncMetrics_BlockWriteMerges_sda gauge
ClickHouseAsyncMetrics_BlockWriteMerges_sda 22
# TYPE ClickHouseAsyncMetrics_OSNiceTime gauge
ClickHouseAsyncMetrics_OSNiceTime 0
# TYPE ClickHouseAsyncMetrics_CPUFrequencyMHz_3 gauge
ClickHouseAsyncMetrics_CPUFrequencyMHz_3 2400
# TYPE ClickHouseAsyncMetrics_CPUFrequencyMHz_2 gauge
ClickHouseAsyncMetrics_CPUFrequencyMHz_2 2400
# TYPE ClickHouseAsyncMetrics_OSMemoryBuffers gauge
ClickHouseAsyncMetrics_OSMemoryBuffers 1117601792
# TYPE ClickHouseAsyncMetrics_OSMemoryFreeWithoutCached gauge
ClickHouseAsyncMetrics_OSMemoryFreeWithoutCached 11124215808
# TYPE ClickHouseAsyncMetrics_OSGuestNiceTimeCPU3 gauge
ClickHouseAsyncMetrics_OSGuestNiceTimeCPU3 0
# TYPE ClickHouseAsyncMetrics_Temperature_nvme_Sensor_2 gauge
ClickHouseAsyncMetrics_Temperature_nvme_Sensor_2 32.85
# TYPE ClickHouseAsyncMetrics_OSGuestNiceTimeNormalized gauge
ClickHouseAsyncMetrics_OSGuestNiceTimeNormalized 0
# TYPE ClickHouseAsyncMetrics_OSSoftIrqTimeNormalized gauge
ClickHouseAsyncMetrics_OSSoftIrqTimeNormalized 0
# TYPE ClickHouseAsyncMetrics_Temperature_acpitz gauge
ClickHouseAsyncMetrics_Temperature_acpitz 53
# TYPE ClickHouseAsyncMetrics_OSSystemTimeNormalized gauge
ClickHouseAsyncMetrics_OSSystemTimeNormalized 0.0262475327319232
# TYPE ClickHouseAsyncMetrics_FilesystemMainPathUsedINodes gauge
ClickHouseAsyncMetrics_FilesystemMainPathUsedINodes 1690184
# TYPE ClickHouseAsyncMetrics_OSProcessesCreated gauge
ClickHouseAsyncMetrics_OSProcessesCreated 24
# TYPE ClickHouseAsyncMetrics_OSContextSwitches gauge
ClickHouseAsyncMetrics_OSContextSwitches 21433
# TYPE ClickHouseAsyncMetrics_OSProcessesBlocked gauge
ClickHouseAsyncMetrics_OSProcessesBlocked 0
# TYPE ClickHouseAsyncMetrics_OSSystemTime gauge
ClickHouseAsyncMetrics_OSSystemTime 0.2099802618553856
# TYPE ClickHouseAsyncMetrics_BlockDiscardMerges_nvme0n1 gauge
ClickHouseAsyncMetrics_BlockDiscardMerges_nvme0n1 0
# TYPE ClickHouseAsyncMetrics_OSGuestNiceTimeCPU7 gauge
ClickHouseAsyncMetrics_OSGuestNiceTimeCPU7 0
# TYPE ClickHouseAsyncMetrics_OSMemoryTotal gauge
ClickHouseAsyncMetrics_OSMemoryTotal 25006321664
# TYPE ClickHouseAsyncMetrics_PrometheusThreads gauge
ClickHouseAsyncMetrics_PrometheusThreads 0
# TYPE ClickHouseAsyncMetrics_OSStealTimeCPU7 gauge
ClickHouseAsyncMetrics_OSStealTimeCPU7 0
# TYPE ClickHouseAsyncMetrics_OSSoftIrqTimeCPU7 gauge
ClickHouseAsyncMetrics_OSSoftIrqTimeCPU7 0
# TYPE ClickHouseAsyncMetrics_OSIOWaitTimeCPU7 gauge
ClickHouseAsyncMetrics_OSIOWaitTimeCPU7 0
# TYPE ClickHouseAsyncMetrics_OSSystemTimeCPU2 gauge
ClickHouseAsyncMetrics_OSSystemTimeCPU2 0.009999060088351695
# TYPE ClickHouseAsyncMetrics_OSIdleTimeCPU7 gauge
ClickHouseAsyncMetrics_OSIdleTimeCPU7 0.8699182276865974
# TYPE ClickHouseAsyncMetrics_OSNiceTimeCPU7 gauge
ClickHouseAsyncMetrics_OSNiceTimeCPU7 0
# TYPE ClickHouseAsyncMetrics_OSInterrupts gauge
ClickHouseAsyncMetrics_OSInterrupts 10126
# TYPE ClickHouseAsyncMetrics_OSUserTimeCPU7 gauge
ClickHouseAsyncMetrics_OSUserTimeCPU7 0.07999248070681356
# TYPE ClickHouseAsyncMetrics_BlockReadTime_nvme0n1 gauge
ClickHouseAsyncMetrics_BlockReadTime_nvme0n1 0
# TYPE ClickHouseAsyncMetrics_OSGuestTimeCPU6 gauge
ClickHouseAsyncMetrics_OSGuestTimeCPU6 0
# TYPE ClickHouseAsyncMetrics_BlockDiscardBytes_sda gauge
ClickHouseAsyncMetrics_BlockDiscardBytes_sda 0
# TYPE ClickHouseAsyncMetrics_OSIdleTimeCPU6 gauge
ClickHouseAsyncMetrics_OSIdleTimeCPU6 0.8599191675982457
# TYPE ClickHouseAsyncMetrics_jemalloc_arenas_all_pdirty gauge
ClickHouseAsyncMetrics_jemalloc_arenas_all_pdirty 25107
# TYPE ClickHouseAsyncMetrics_OSIdleTimeNormalized gauge
ClickHouseAsyncMetrics_OSIdleTimeNormalized 0.8724179927086854
# TYPE ClickHouseAsyncMetrics_OSStealTimeCPU6 gauge
ClickHouseAsyncMetrics_OSStealTimeCPU6 0
# TYPE ClickHouseAsyncMetrics_OSIrqTimeCPU7 gauge
ClickHouseAsyncMetrics_OSIrqTimeCPU7 0
# TYPE ClickHouseAsyncMetrics_OSIOWaitTimeCPU6 gauge
ClickHouseAsyncMetrics_OSIOWaitTimeCPU6 0
# TYPE ClickHouseAsyncMetrics_OSGuestTimeCPU7 gauge
ClickHouseAsyncMetrics_OSGuestTimeCPU7 0
# TYPE ClickHouseAsyncMetrics_OSSystemTimeCPU7 gauge
ClickHouseAsyncMetrics_OSSystemTimeCPU7 0.029997180265055084
# TYPE ClickHouseAsyncMetrics_OSGuestNiceTimeCPU4 gauge
ClickHouseAsyncMetrics_OSGuestNiceTimeCPU4 0
# TYPE ClickHouseAsyncMetrics_OSMemoryCached gauge
ClickHouseAsyncMetrics_OSMemoryCached 6021181440
# TYPE ClickHouseAsyncMetrics_OSUserTimeCPU6 gauge
ClickHouseAsyncMetrics_OSUserTimeCPU6 0.06999342061846187
# TYPE ClickHouseAsyncMetrics_OSGuestNiceTimeCPU5 gauge
ClickHouseAsyncMetrics_OSGuestNiceTimeCPU5 0
# TYPE ClickHouseAsyncMetrics_OSGuestNiceTimeCPU6 gauge
ClickHouseAsyncMetrics_OSGuestNiceTimeCPU6 0
# TYPE ClickHouseAsyncMetrics_OSGuestNiceTime gauge
ClickHouseAsyncMetrics_OSGuestNiceTime 0
# TYPE ClickHouseAsyncMetrics_OSSystemTimeCPU6 gauge
ClickHouseAsyncMetrics_OSSystemTimeCPU6 0.029997180265055084
# TYPE ClickHouseAsyncMetrics_MemoryDataAndStack gauge
ClickHouseAsyncMetrics_MemoryDataAndStack 3152109568
# TYPE ClickHouseAsyncMetrics_FilesystemMainPathTotalINodes gauge
ClickHouseAsyncMetrics_FilesystemMainPathTotalINodes 61014016
# TYPE ClickHouseAsyncMetrics_OSGuestTimeCPU3 gauge
ClickHouseAsyncMetrics_OSGuestTimeCPU3 0
# TYPE ClickHouseAsyncMetrics_OSUserTimeCPU5 gauge
ClickHouseAsyncMetrics_OSUserTimeCPU5 0.10998966097186864
# TYPE ClickHouseAsyncMetrics_OSStealTimeNormalized gauge
ClickHouseAsyncMetrics_OSStealTimeNormalized 0
# TYPE ClickHouseAsyncMetrics_OSGuestNiceTimeCPU2 gauge
ClickHouseAsyncMetrics_OSGuestNiceTimeCPU2 0
# TYPE ClickHouseAsyncMetrics_OSSoftIrqTimeCPU2 gauge
ClickHouseAsyncMetrics_OSSoftIrqTimeCPU2 0
# TYPE ClickHouseAsyncMetrics_OSSystemTimeCPU5 gauge
ClickHouseAsyncMetrics_OSSystemTimeCPU5 0.01999812017670339
# TYPE ClickHouseAsyncMetrics_OSSoftIrqTimeCPU4 gauge
ClickHouseAsyncMetrics_OSSoftIrqTimeCPU4 0
# TYPE ClickHouseAsyncMetrics_OSGuestNiceTimeCPU0 gauge
ClickHouseAsyncMetrics_OSGuestNiceTimeCPU0 0
# TYPE ClickHouseAsyncMetrics_OSIrqTime gauge
ClickHouseAsyncMetrics_OSIrqTime 0
# TYPE ClickHouseAsyncMetrics_OSStealTimeCPU1 gauge
ClickHouseAsyncMetrics_OSStealTimeCPU1 0
# TYPE ClickHouseAsyncMetrics_OSSoftIrqTimeCPU1 gauge
ClickHouseAsyncMetrics_OSSoftIrqTimeCPU1 0
# TYPE ClickHouseAsyncMetrics_MemoryShared gauge
ClickHouseAsyncMetrics_MemoryShared 376786944
# TYPE ClickHouseAsyncMetrics_OSIdleTimeCPU3 gauge
ClickHouseAsyncMetrics_OSIdleTimeCPU3 0.8999154079516525
# TYPE ClickHouseAsyncMetrics_TotalRowsOfMergeTreeTables gauge
ClickHouseAsyncMetrics_TotalRowsOfMergeTreeTables 10878715
# TYPE ClickHouseAsy5_23ncMetrics_OSGuestTimeCPU4 gauge
ClickHouseAsyncMetrics_OSGuestTimeCPU4 0
# TYPE ClickHouseAsyncMetrics_BlockWriteMerges_nvme0n1 gauge
ClickHouseAsyncMetrics_BlockWriteMerges_nvme0n1 0
# TYPE ClickHouseAsyncMetrics_OSIrqTimeCPU4 gauge
ClickHouseAsyncMetrics_OSIrqTimeCPU4 0
# TYPE ClickHouseAsyncMetrics_UncompressedCacheBytes gauge
ClickHouseAsyncMetrics_UncompressedCacheBytes 0
# TYPE ClickHouseAsyncMetrics_Uptime gauge
ClickHouseAsyncMetrics_Uptime 27
# TYPE ClickHouseAsyncMetrics_BlockReadOps_nvme0n1 gauge
ClickHouseAsyncMetrics_BlockReadOps_nvme0n1 0
# TYPE ClickHouseAsyncMetrics_OSIdleTimeCPU4 gauge
ClickHouseAsyncMetrics_OSIdleTimeCPU4 0.8699182276865974
# TYPE ClickHouseAsyncMetrics_NetworkReceiveBytes_eth0 gauge
ClickHouseAsyncMetrics_NetworkReceiveBytes_eth0 0
# TYPE ClickHouseAsyncMetrics_MaxPartCountForPartition gauge
ClickHouseAsyncMetrics_MaxPartCountForPartition 5
# TYPE ClickHouseAsyncMetrics_OSUserTimeCPU4 gauge
ClickHouseAsyncMetrics_OSUserTimeCPU4 0.05999436053011017
# TYPE ClickHouseAsyncMetrics_OSNiceTimeCPU4 gauge
ClickHouseAsyncMetrics_OSNiceTimeCPU4 0
# TYPE ClickHouseAsyncMetrics_OSStealTimeCPU3 gauge
ClickHouseAsyncMetrics_OSStealTimeCPU3 0
# TYPE ClickHouseAsyncMetrics_OSIrqTimeCPU3 gauge
ClickHouseAsyncMetrics_OSIrqTimeCPU3 0
# TYPE ClickHouseAsyncMetrics_OSIOWaitTimeCPU3 gauge
ClickHouseAsyncMetrics_OSIOWaitTimeCPU3 0
# TYPE ClickHouseAsyncMetrics_NetworkSendBytes_eth0 gauge
ClickHouseAsyncMetrics_NetworkSendBytes_eth0 0
# TYPE ClickHouseAsyncMetrics_OSIdleTimeCPU5 gauge
ClickHouseAsyncMetrics_OSIdleTimeCPU5 0.8599191675982457
# TYPE ClickHouseAsyncMetrics_OSUserTime gauge
ClickHouseAsyncMetrics_OSUserTime 0.5799454851243983
# TYPE ClickHouseAsyncMetrics_OSGuestTime gauge
ClickHouseAsyncMetrics_OSGuestTime 0
# TYPE ClickHouseAsyncMetrics_OSUserTimeCPU3 gauge
ClickHouseAsyncMetrics_OSUserTimeCPU3 0.05999436053011017
# TYPE ClickHouseAsyncMetrics_OSGuestTimeCPU2 gauge
ClickHouseAsyncMetrics_OSGuestTimeCPU2 0
# TYPE ClickHouseAsyncMetrics_OSUserTimeNormalized gauge
ClickHouseAsyncMetrics_OSUserTimeNormalized 0.07249318564054978
# TYPE ClickHouseAsyncMetrics_OSIrqTimeCPU2 gauge
ClickHouseAsyncMetrics_OSIrqTimeCPU2 0
# TYPE ClickHouseAsyncMetrics_Temperature4 gauge
ClickHouseAsyncMetrics_Temperature4 52
# TYPE ClickHouseAsyncMetrics_OSIOWaitTimeCPU2 gauge
ClickHouseAsyncMetrics_OSIOWaitTimeCPU2 0.009999060088351695
# TYPE ClickHouseAsyncMetrics_MemoryCode gauge
ClickHouseAsyncMetrics_MemoryCode 292159488
# TYPE ClickHouseAsyncMetrics_OSStealTime gauge
ClickHouseAsyncMetrics_OSStealTime 0
# TYPE ClickHouseAsyncMetrics_OSIOWaitTimeNormalized gauge
ClickHouseAsyncMetrics_OSIOWaitTimeNormalized 0.0024997650220879237
# TYPE ClickHouseAsyncMetrics_OSIdleTimeCPU2 gauge
ClickHouseAsyncMetrics_OSIdleTimeCPU2 0.8999154079516525
# TYPE ClickHouseAsyncMetrics_OSProcessesRunning gauge
ClickHouseAsyncMetrics_OSProcessesRunning 1
# TYPE ClickHouseAsyncMetrics_OSNiceTimeCPU2 gauge
ClickHouseAsyncMetrics_OSNiceTimeCPU2 0
# TYPE ClickHouseAsyncMetrics_LoadAverage1 gauge
ClickHouseAsyncMetrics_LoadAverage1 1.02
# TYPE ClickHouseAsyncMetrics_OSStealTimeCPU5 gauge
ClickHouseAsyncMetrics_OSStealTimeCPU5 0
# TYPE ClickHouseAsyncMetrics_OSGuestNiceTimeCPU1 gauge
ClickHouseAsyncMetrics_OSGuestNiceTimeCPU1 0
# TYPE ClickHouseAsyncMetrics_jemalloc_retained gauge
ClickHouseAsyncMetrics_jemalloc_retained 372334592
# TYPE ClickHouseAsyncMetrics_OSGuestTimeCPU1 gauge
ClickHouseAsyncMetrics_OSGuestTimeCPU1 0
# TYPE ClickHouseAsyncMetrics_BlockWriteBytes_sda gauge
ClickHouseAsyncMetrics_BlockWriteBytes_sda 118784
# TYPE ClickHouseAsyncMetrics_OSIdleTimeCPU0 gauge
ClickHouseAsyncMetrics_OSIdleTimeCPU0 0.9199135281283559
# TYPE ClickHouseAsyncMetrics_LoadAverage15 gauge
ClickHouseAsyncMetrics_LoadAverage15 0.81
# TYPE ClickHouseAsyncMetrics_OSIOWaitTime gauge
ClickHouseAsyncMetrics_OSIOWaitTime 0.01999812017670339
# TYPE ClickHouseAsyncMetrics_OSIOWaitTimeCPU1 gauge
ClickHouseAsyncMetrics_OSIOWaitTimeCPU1 0
# TYPE ClickHouseAsyncMetrics_OSIrqTimeCPU1 gauge
ClickHouseAsyncMetrics_OSIrqTimeCPU1 0
# TYPE ClickHouseAsyncMetrics_MMapCacheCells gauge
ClickHouseAsyncMetrics_MMapCacheCells 0
# TYPE ClickHouseAsyncMetrics_OSStealTimeCPU2 gauge
ClickHouseAsyncMetrics_OSStealTimeCPU2 0
# TYPE ClickHouseAsyncMetrics_OSSystemTimeCPU1 gauge
ClickHouseAsyncMetrics_OSSystemTimeCPU1 0.05999436053011017
# TYPE ClickHouseAsyncMetrics_Temperature_coretemp_Core_2 gauge
ClickHouseAsyncMetrics_Temperature_coretemp_Core_2 56
# TYPE ClickHouseAsyncMetrics_OSUserTimeCPU1 gauge
ClickHouseAsyncMetrics_OSUserTimeCPU1 0.08999154079516525
# TYPE ClickHouseAsyncMetrics_Temperature2 gauge
ClickHouseAsyncMetrics_Temperature2 20
# TYPE ClickHouseAsyncMetrics_OSSoftIrqTimeCPU3 gauge
ClickHouseAsyncMetrics_OSSoftIrqTimeCPU3 0
# TYPE ClickHouseAsyncMetrics_OSNiceTimeCPU3 gauge
ClickHouseAsyncMetrics_OSNiceTimeCPU3 0
# TYPE ClickHouseAsyncMetrics_OSSystemTimeCPU3 gauge
ClickHouseAsyncMetrics_OSSystemTimeCPU3 0
# TYPE ClickHouseAsyncMetrics_MemoryVirtual gauge
ClickHouseAsyncMetrics_MemoryVirtual 4081573888
# TYPE ClickHouseAsyncMetrics_OSStealTimeCPU0 gauge
ClickHouseAsyncMetrics_OSStealTimeCPU0 0
# TYPE ClickHouseAsyncMetrics_FilesystemLogsPathAvailableINodes gauge
ClickHouseAsyncMetrics_FilesystemLogsPathAvailableINodes 59323832
# TYPE ClickHouseAsyncMetrics_OSGuestTimeNormalized gauge
ClickHouseAsyncMetrics_OSGuestTimeNormalized 0
# TYPE ClickHouseAsyncMetrics_OSIOWaitTimeCPU5 gauge
ClickHouseAsyncMetrics_OSIOWaitTimeCPU5 0
# TYPE ClickHouseAsyncMetrics_BlockReadBytes_sda gauge
ClickHouseAsyncMetrics_BlockReadBytes_sda 151552
# TYPE ClickHouseAsyncMetrics_Temperature_coretemp_Core_3 gauge
ClickHouseAsyncMetrics_Temperature_coretemp_Core_3 56
# TYPE ClickHouseAsyncMetrics_OSNiceTimeCPU1 gauge
ClickHouseAsyncMetrics_OSNiceTimeCPU1 0
# TYPE ClickHouseAsyncMetrics_OSIOWaitTimeCPU0 gauge
ClickHouseAsyncMetrics_OSIOWaitTimeCPU0 0
# TYPE ClickHouseAsyncMetrics_OSSoftIrqTimeCPU6 gauge
ClickHouseAsyncMetrics_OSSoftIrqTimeCPU6 0
# TYPE ClickHouseAsyncMetrics_jemalloc_background_thread_num_threads gauge
ClickHouseAsyncMetrics_jemalloc_background_thread_num_threads 0
# TYPE ClickHouseAsyncMetrics_OSSystemTimeCPU0 gauge
ClickHouseAsyncMetrics_OSSystemTimeCPU0 0.01999812017670339
# TYPE ClickHouseAsyncMetrics_Temperature_coretemp_Core_0 gauge
ClickHouseAsyncMetrics_Temperature_coretemp_Core_0 57
# TYPE ClickHouseAsyncMetrics_InterserverThreads gauge
ClickHouseAsyncMetrics_InterserverThreads 0
# TYPE ClickHouseAsyncMetrics_OSNiceTimeCPU0 gauge
ClickHouseAsyncMetrics_OSNiceTimeCPU0 0
# TYPE ClickHouseAsyncMetrics_OSUserTimeCPU0 gauge
ClickHouseAsyncMetrics_OSUserTimeCPU0 0.029997180265055084
# TYPE ClickHouseAsyncMetrics_OSSystemTimeCPU4 gauge
ClickHouseAsyncMetrics_OSSystemTimeCPU4 0.01999812017670339
# TYPE ClickHouseAsyncMetrics_MarkCacheFiles gauge
ClickHouseAsyncMetrics_MarkCacheFiles 0
# TYPE ClickHouseAsyncMetrics_OSIOWaitTimeCPU4 gauge
ClickHouseAsyncMetrics_OSIOWaitTimeCPU4 0.009999060088351695
# TYPE ClickHouseAsyncMetrics_BlockReadMerges_sda gauge
ClickHouseAsyncMetrics_BlockReadMerges_sda 0
# TYPE ClickHouseAsyncMetrics_OSSoftIrqTime gauge
ClickHouseAsyncMetrics_OSSoftIrqTime 0
# TYPE ClickHouseAsyncMetrics_CPUFrequencyMHz_0 gauge
ClickHouseAsyncMetrics_CPUFrequencyMHz_0 1295.227
# TYPE ClickHouseAsyncMetrics_CompiledExpressionCacheBytes gauge
ClickHouseAsyncMetrics_CompiledExpressionCacheBytes 0
# TYPE ClickHouseAsyncMetrics_OSIdleTime gauge
ClickHouseAsyncMetrics_OSIdleTime 6.979343941669483
# TYPE ClickHouseAsyncMetrics_BlockWriteOps_nvme0n1 gauge
ClickHouseAsyncMetrics_BlockWriteOps_nvme0n1 0
# TYPE ClickHouseAsyncMetrics_BlockWriteBytes_nvme0n1 gauge
ClickHouseAsyncMetrics_BlockWriteBytes_nvme0n1 0
# TYPE ClickHouseAsyncMetrics_OSStealTimeCPU4 gauge
ClickHouseAsyncMetrics_OSStealTimeCPU4 0
# TYPE ClickHouseAsyncMetrics_OSMemoryAvailable gauge
ClickHouseAsyncMetrics_OSMemoryAvailable 17156100096
# TYPE ClickHouseAsyncMetrics_CompiledExpressionCacheCount gauge
ClickHouseAsyncMetrics_CompiledExpressionCacheCount 0
# TYPE ClickHouseAsyncMetrics_TotalBytesOfMergeTreeTables gauge
ClickHouseAsyncMetrics_TotalBytesOfMergeTreeTables 15884392
# TYPE ClickHouseAsyncMetrics_OSNiceTimeCPU6 gauge
ClickHouseAsyncMetrics_OSNiceTimeCPU6 0
# TYPE ClickHouseAsyncMetrics_MarkCacheBytes gauge
ClickHouseAsyncMetrics_MarkCacheBytes 0
# TYPE ClickHouseAsyncMetrics_BlockDiscardTime_sda gauge
ClickHouseAsyncMetrics_BlockDiscardTime_sda 0
# TYPE ClickHouseAsyncMetrics_Temperature6 gauge
ClickHouseAsyncMetrics_Temperature6 54
# TYPE ClickHouseAsyncMetrics_OSIrqTimeCPU0 gauge
ClickHouseAsyncMetrics_OSIrqTimeCPU0 0
# TYPE ClickHouseAsyncMetrics_OSUptime gauge
ClickHouseAsyncMetrics_OSUptime 522.1
# TYPE ClickHouseAsyncMetrics_BlockInFlightOps_nvme0n1 gauge
ClickHouseAsyncMetrics_BlockInFlightOps_nvme0n1 0
# TYPE ClickHouseAsyncMetrics_OSSoftIrqTimeCPU0 gauge
ClickHouseAsyncMetrics_OSSoftIrqTimeCPU0 0
# TYPE ClickHouseAsyncMetrics_OSGuestTimeCPU5 gauge
ClickHouseAsyncMetrics_OSGuestTimeCPU5 0
# TYPE ClickHouseAsyncMetrics_OSMemoryFreePlusCached gauge
ClickHouseAsyncMetrics_OSMemoryFreePlusCached 17145397248
# TYPE ClickHouseAsyncMetrics_OSThreadsRunnable gauge
ClickHouseAsyncMetrics_OSThreadsRunnable 1
# TYPE ClickHouseAsyncMetrics_LoadAverage5 gauge
ClickHouseAsyncMetrics_LoadAverage5 1.21
# TYPE ClickHouseAsyncMetrics_OSUserTimeCPU2 gauge
ClickHouseAsyncMetrics_OSUserTimeCPU2 0.06999342061846187
# TYPE ClickHouseAsyncMetrics_OSIdleTimeCPU1 gauge
ClickHouseAsyncMetrics_OSIdleTimeCPU1 0.8099238671564872
# TYPE ClickHouseAsyncMetrics_Temperature_iwlwifi_1 gauge
ClickHouseAsyncMetrics_Temperature_iwlwifi_1 54
# TYPE ClickHouseAsyncMetrics_OSNiceTimeNormalized gauge
ClickHouseAsyncMetrics_OSNiceTimeNormalized 0
# TYPE ClickHouseAsyncMetrics_OSThreadsTotal gauge
ClickHouseAsyncMetrics_OSThreadsTotal 1768
# TYPE ClickHouseAsyncMetrics_OSGuestTimeCPU0 gauge
ClickHouseAsyncMetrics_OSGuestTimeCPU0 0
# TYPE ClickHouseAsyncMetrics_OSIrqTimeCPU6 gauge
ClickHouseAsyncMetrics_OSIrqTimeCPU6 0
# TYPE ClickHouseAsyncMetrics_jemalloc_background_thread_num_runs gauge
ClickHouseAsyncMetrics_jemalloc_background_thread_num_runs 0
# TYPE ClickHouseAsyncMetrics_BlockDiscardBytes_nvme0n1 gauge
ClickHouseAsyncMetrics_BlockDiscardBytes_nvme0n1 0
# TYPE ClickHouseAsyncMetrics_CPUFrequencyMHz_5 gauge
ClickHouseAsyncMetrics_CPUFrequencyMHz_5 2400
# TYPE ClickHouseAsyncMetrics_UncompressedCacheCells gauge
ClickHouseAsyncMetrics_UncompressedCacheCells 0
# TYPE ClickHouseAsyncMetrics_Jitter gauge
ClickHouseAsyncMetrics_Jitter 0.000094
# TYPE ClickHouseAsyncMetrics_MySQLThreads gauge
ClickHouseAsyncMetrics_MySQLThreads 0
# HELP ClickHouseStatusInfo_DictionaryStatus "Dictionary Status."
# TYPE ClickHouseStatusInfo_DictionaryStatus gauge
`

var want_v22_8_15_23 = []string{
	"ClickHouseProfileEvents,instance=localhost:9363 CachedReadBufferReadFromSourceMicroseconds",
	"ClickHouseAsyncMetrics,instance=localhost:9363,unit=sda BlockWriteOps",
	"ClickHouseAsyncMetrics,instance=localhost:9363,unit=coretemp_Core_3 Temperature",
	"ClickHouseAsyncMetrics,cpu=6,instance=localhost:9363 CPUFrequencyMHz",
	"ClickHouseAsyncMetrics,instance=localhost:9363 Uptime",
	"ClickHouseAsyncMetrics,cpu=3,instance=localhost:9363 OSStealTimeCPU",
	"ClickHouseAsyncMetrics,instance=localhost:9363 MemoryCode",
	"ClickHouseAsyncMetrics,cpu=3,instance=localhost:9363 OSSystemTimeCPU",
	"ClickHouseAsyncMetrics,cpu=1,instance=localhost:9363 OSIdleTimeCPU",
	"ClickHouseMetrics,instance=localhost:9363 DelayedInserts",
	"ClickHouseAsyncMetrics,cpu=7,instance=localhost:9363 CPUFrequencyMHz",
	"ClickHouseProfileEvents,instance=localhost:9363 Query",
	"ClickHouseProfileEvents,instance=localhost:9363 CreatedReadBufferDirectIO",
	"ClickHouseProfileEvents,instance=localhost:9363 DiskWriteElapsedMicroseconds",
	"ClickHouseProfileEvents,instance=localhost:9363 ZooKeeperTransactions",
	"ClickHouseProfileEvents,instance=localhost:9363 ExternalAggregationUncompressedBytes",
	"ClickHouseMetrics,instance=localhost:9363 ReplicatedFetch",
	"ClickHouseAsyncMetrics,cpu=2,instance=localhost:9363 OSGuestNiceTimeCPU",
	"ClickHouseAsyncMetrics,cpu=6,instance=localhost:9363 OSSoftIrqTimeCPU",
	"ClickHouseAsyncMetrics,cpu=0,instance=localhost:9363 OSGuestTimeCPU",
	"ClickHouseProfileEvents,instance=localhost:9363 MergeTreeDataWriterCompressedBytes",
	"ClickHouseProfileEvents,instance=localhost:9363 MergeTreeDataProjectionWriterUncompressedBytes",
	"ClickHouseProfileEvents,instance=localhost:9363 OSReadBytes",
	"ClickHouseProfileEvents,instance=localhost:9363 PerfInstructions",
	"ClickHouseProfileEvents,instance=localhost:9363 PerfTaskClock",
	"ClickHouseProfileEvents,instance=localhost:9363 PerfDataTLBMisses",
	"ClickHouseProfileEvents,instance=localhost:9363 DictCacheKeysHit",
	"ClickHouseProfileEvents,instance=localhost:9363 RemoteFSLazySeeks",
	"ClickHouseMetrics,instance=localhost:9363 GlobalThreadActive",
	"ClickHouseAsyncMetrics,instance=localhost:9363 ReplicasMaxInsertsInQueue",
	"ClickHouseAsyncMetrics,instance=localhost:9363 OSContextSwitches",
	"ClickHouseAsyncMetrics,instance=localhost:9363 jemalloc_background_thread_num_runs",
	"ClickHouseAsyncMetrics,instance=localhost:9363,unit=coretemp_Core_2 Temperature",
	"ClickHouseProfileEvents,instance=localhost:9363 S3ReadRequestsThrottling",
	"ClickHouseProfileEvents,instance=localhost:9363 KafkaCommits",
	"ClickHouseMetrics,instance=localhost:9363 BackgroundFetchesPoolTask",
	"ClickHouseMetrics,instance=localhost:9363 MemoryTracking",
	"ClickHouseAsyncMetrics,instance=localhost:9363,unit=1 Temperature",
	"ClickHouseAsyncMetrics,cpu=6,instance=localhost:9363 OSUserTimeCPU",
	"ClickHouseProfileEvents,instance=localhost:9363 MergeTreeDataProjectionWriterCompressedBytes",
	"ClickHouseProfileEvents,instance=localhost:9363 QueryProfilerRuns",
	"ClickHouseProfileEvents,instance=localhost:9363 FileSegmentWaitReadBufferMicroseconds",
	"ClickHouseProfileEvents,instance=localhost:9363 ExternalDataSourceLocalCacheReadBytes",
	"ClickHouseProfileEvents,instance=localhost:9363 MergeTreeMetadataCacheSeek",
	"ClickHouseAsyncMetrics,instance=localhost:9363 FilesystemMainPathUsedINodes",
	"ClickHouseMetrics,instance=localhost:9363 Read",
	"ClickHouseAsyncMetrics,instance=localhost:9363 MemoryDataAndStack",
	"ClickHouseAsyncMetrics,cpu=4,instance=localhost:9363 OSIdleTimeCPU",
	"ClickHouseAsyncMetrics,instance=localhost:9363,unit=sda BlockWriteMerges",
	"ClickHouseAsyncMetrics,eth=eth0,instance=localhost:9363 NetworkSendBytes",
	"ClickHouseProfileEvents,instance=localhost:9363 ZooKeeperList",
	"ClickHouseProfileEvents,instance=localhost:9363 ZooKeeperBytesReceived",
	"ClickHouseProfileEvents,instance=localhost:9363 DistributedConnectionFailTry",
	"ClickHouseProfileEvents,instance=localhost:9363 ScalarSubqueriesLocalCacheHit",
	"ClickHouseMetrics,instance=localhost:9363 PartMutation",
	"ClickHouseMetrics,instance=localhost:9363 SyncDrainedConnections",
	"ClickHouseAsyncMetrics,instance=localhost:9363 UncompressedCacheCells",
	"ClickHouseProfileEvents,instance=localhost:9363 StorageBufferPassedTimeFlushThreshold",
	"ClickHouseMetrics,instance=localhost:9363 FilesystemCacheElements",
	"ClickHouseAsyncMetrics,disk=default,instance=localhost:9363 DiskUnreserved",
	"ClickHouseProfileEvents,instance=localhost:9363 HedgedRequestsChangeReplica",
	"ClickHouseProfileEvents,instance=localhost:9363 KeeperPacketsSent",
	"ClickHouseAsyncMetrics,instance=localhost:9363 OSIrqTime",
	"ClickHouseProfileEvents,instance=localhost:9363 TableFunctionExecute",
	"ClickHouseProfileEvents,instance=localhost:9363 RemoteFSSeeksWithReset",
	"ClickHouseAsyncMetrics,instance=localhost:9363 ReplicasSumInsertsInQueue",
	"ClickHouseAsyncMetrics,cpu=7,instance=localhost:9363 OSStealTimeCPU",
	"ClickHouseAsyncMetrics,instance=localhost:9363 OSGuestNiceTime",
	"ClickHouseProfileEvents,instance=localhost:9363 CompressedReadBufferBytes",
	"ClickHouseProfileEvents,instance=localhost:9363 StorageBufferFlush",
	"ClickHouseProfileEvents,instance=localhost:9363 S3WriteRequestsCount",
	"ClickHouseAsyncMetrics,instance=localhost:9363 ReplicasSumMergesInQueue",
	"ClickHouseAsyncMetrics,eth=eth0,instance=localhost:9363 NetworkSendPackets",
	"ClickHouseAsyncMetrics,instance=localhost:9363 OSOpenFiles",
	"ClickHouseAsyncMetrics,instance=localhost:9363,unit=sda BlockWriteBytes",
	"ClickHouseAsyncMetrics,instance=localhost:9363,unit=15 LoadAverage",
	"ClickHouseProfileEvents,instance=localhost:9363 DiskReadElapsedMicroseconds",
	"ClickHouseMetrics,instance=localhost:9363 OpenFileForWrite",
	"ClickHouseMetrics,instance=localhost:9363 PendingAsyncInsert",
	"ClickHouseAsyncMetrics,instance=localhost:9363 jemalloc_arenas_all_pdirty",
	"ClickHouseAsyncMetrics,cpu=2,instance=localhost:9363 OSGuestTimeCPU",
	"ClickHouseProfileEvents,instance=localhost:9363 ArenaAllocChunks",
	"ClickHouseProfileEvents,instance=localhost:9363 ReplicatedPartFetchesOfMerged",
	"ClickHouseProfileEvents,instance=localhost:9363 KeeperSaveSnapshot",
	"ClickHouseMetrics,instance=localhost:9363 Query",
	"ClickHouseAsyncMetrics,instance=localhost:9363 MySQLThreads",
	"ClickHouseAsyncMetrics,instance=localhost:9363,unit=sda BlockActiveTime",
	"ClickHouseAsyncMetrics,cpu=3,instance=localhost:9363 CPUFrequencyMHz",
	"ClickHouseProfileEvents,instance=localhost:9363 MMappedFileCacheMisses",
	"ClickHouseProfileEvents,instance=localhost:9363 PerfCacheMisses",
	"ClickHouseProfileEvents,instance=localhost:9363 WriteBufferFromS3Bytes",
	"ClickHouseProfileEvents,instance=localhost:9363 FileSegmentPredownloadMicroseconds",
	"ClickHouseAsyncMetrics,instance=localhost:9363 jemalloc_background_thread_run_intervals",
	"ClickHouseAsyncMetrics,instance=localhost:9363,unit=nvme0n1 BlockActiveTime",
	"ClickHouseAsyncMetrics,cpu=6,instance=localhost:9363 OSGuestNiceTimeCPU",
	"ClickHouseProfileEvents,instance=localhost:9363 ZooKeeperHardwareExceptions",
	"ClickHouseProfileEvents,instance=localhost:9363 DictCacheKeysRequestedFound",
	"ClickHouseProfileEvents,instance=localhost:9363 PerfCpuMigrations",
	"ClickHouseProfileEvents,instance=localhost:9363 S3ReadRequestsCount",
	"ClickHouseProfileEvents,instance=localhost:9363 FileSegmentCacheWriteMicroseconds",
	"ClickHouseAsyncMetrics,instance=localhost:9363 OSMemoryFreeWithoutCached",
	"ClickHouseProfileEvents,instance=localhost:9363 ThreadPoolReaderPageCacheHit",
	"ClickHouseMetrics,instance=localhost:9363 PartsPreActive",
	"ClickHouseAsyncMetrics,instance=localhost:9363 FilesystemLogsPathTotalINodes",
	"ClickHouseMetrics,instance=localhost:9363 ReadonlyReplica",
	"ClickHouseMetrics,instance=localhost:9363 MaxPushedDDLEntryID",
	"ClickHouseProfileEvents,instance=localhost:9363 ObsoleteReplicatedParts",
	"ClickHouseProfileEvents,instance=localhost:9363 InsertedRows",
	"ClickHouseProfileEvents,instance=localhost:9363 DistributedDelayedInserts",
	"ClickHouseProfileEvents,instance=localhost:9363 SelectedBytes",
	"ClickHouseMetrics,instance=localhost:9363 HTTPConnection",
	"ClickHouseMetrics,instance=localhost:9363 SendScalars",
	"ClickHouseAsyncMetrics,instance=localhost:9363 OSGuestTimeNormalized",
	"ClickHouseProfileEvents,instance=localhost:9363 ZooKeeperWatchResponse",
	"ClickHouseProfileEvents,instance=localhost:9363 CompileFunction",
	"ClickHouseProfileEvents,instance=localhost:9363 PerfCpuClock",
	"ClickHouseMetrics,instance=localhost:9363 NetworkSend",
	"ClickHouseMetrics,instance=localhost:9363 Revision",
	"ClickHouseProfileEvents,instance=localhost:9363 FileSyncElapsedMicroseconds",
	"ClickHouseProfileEvents,instance=localhost:9363 ZooKeeperMulti",
	"ClickHouseProfileEvents,instance=localhost:9363 S3WriteRequestsErrors",
	"ClickHouseProfileEvents,instance=localhost:9363 CachedReadBufferReadFromSourceBytes",
	"ClickHouseAsyncMetrics,instance=localhost:9363,unit=sda BlockQueueTime",
	"ClickHouseAsyncMetrics,instance=localhost:9363,unit=acpitz Temperature",
	"ClickHouseProfileEvents,instance=localhost:9363 MMappedFileCacheHits",
	"ClickHouseProfileEvents,instance=localhost:9363 RemoteFSPrefetchedReads",
	"ClickHouseAsyncMetrics,instance=localhost:9363 OSProcessesBlocked",
	"ClickHouseAsyncMetrics,instance=localhost:9363,unit=nvme0n1 BlockDiscardMerges",
	"ClickHouseProfileEvents,instance=localhost:9363 SelectedRanges",
	"ClickHouseProfileEvents,instance=localhost:9363 PerfMinEnabledRunningTime",
	"ClickHouseProfileEvents,instance=localhost:9363 ReadBufferFromS3Bytes",
	"ClickHouseProfileEvents,instance=localhost:9363 KafkaRebalanceAssignments",
	"ClickHouseProfileEvents,instance=localhost:9363 ThreadPoolReaderPageCacheMiss",
	"ClickHouseMetrics,instance=localhost:9363 ContextLockWait",
	"ClickHouseAsyncMetrics,instance=localhost:9363 jemalloc_metadata",
	"ClickHouseAsyncMetrics,instance=localhost:9363 FilesystemLogsPathUsedINodes",
	"ClickHouseAsyncMetrics,cpu=3,instance=localhost:9363 OSSoftIrqTimeCPU",
	"ClickHouseAsyncMetrics,cpu=5,instance=localhost:9363 OSIOWaitTimeCPU",
	"ClickHouseAsyncMetrics,cpu=0,instance=localhost:9363 OSIrqTimeCPU",
	"ClickHouseProfileEvents,instance=localhost:9363 ReadBufferFromFileDescriptorReadFailed",
	"ClickHouseProfileEvents,instance=localhost:9363 IOBufferAllocs",
	"ClickHouseProfileEvents,instance=localhost:9363 PerfEmulationFaults",
	"ClickHouseProfileEvents,instance=localhost:9363 ReadBufferFromS3RequestsErrors",
	"ClickHouseMetrics,instance=localhost:9363 ZooKeeperWatch",
	"ClickHouseAsyncMetrics,cpu=0,instance=localhost:9363 OSNiceTimeCPU",
	"ClickHouseProfileEvents,instance=localhost:9363 QueryProfilerSignalOverruns",
	"ClickHouseProfileEvents,instance=localhost:9363 ScalarSubqueriesGlobalCacheHit",
	"ClickHouseMetrics,instance=localhost:9363 DistributedSend",
	"ClickHouseMetrics,instance=localhost:9363 ZooKeeperSession",
	"ClickHouseAsyncMetrics,instance=localhost:9363,unit=sda BlockReadOps",
	"ClickHouseProfileEvents,instance=localhost:9363 FailedSelectQuery",
	"ClickHouseProfileEvents,instance=localhost:9363 AsynchronousReadWaitMicroseconds",
	"ClickHouseAsyncMetrics,instance=localhost:9363 OSSoftIrqTime",
	"ClickHouseProfileEvents,instance=localhost:9363 PerfCpuCycles",
	"ClickHouseMetrics,instance=localhost:9363 StorageBufferRows",
	"ClickHouseAsyncMetrics,instance=localhost:9363 ReplicasMaxRelativeDelay",
	"ClickHouseAsyncMetrics,instance=localhost:9363 OSInterrupts",
	"ClickHouseAsyncMetrics,instance=localhost:9363 MemoryVirtual",
	"ClickHouseProfileEvents,instance=localhost:9363 ReplicatedDataLoss",
	"ClickHouseProfileEvents,instance=localhost:9363 UserTimeMicroseconds",
	"ClickHouseProfileEvents,instance=localhost:9363 PerfRefCpuCycles",
	"ClickHouseMetrics,instance=localhost:9363 KafkaConsumersInUse",
	"ClickHouseAsyncMetrics,instance=localhost:9363,unit=sda BlockReadTime",
	"ClickHouseProfileEvents,instance=localhost:9363 MergedIntoInMemoryParts",
	"ClickHouseAsyncMetrics,instance=localhost:9363 ReplicasMaxAbsoluteDelay",
	"ClickHouseAsyncMetrics,disk=default,instance=localhost:9363 DiskAvailable",
	"ClickHouseProfileEvents,instance=localhost:9363 RemoteFSUnprefetchedReads",
	"ClickHouseProfileEvents,instance=localhost:9363 ThreadpoolReaderReadBytes",
	"ClickHouseMetrics,instance=localhost:9363 ReplicatedSend",
	"ClickHouseAsyncMetrics,cpu=7,instance=localhost:9363 OSNiceTimeCPU",
	"ClickHouseAsyncMetrics,instance=localhost:9363,unit=sda BlockReadBytes",
	"ClickHouseAsyncMetrics,instance=localhost:9363,unit=5 Temperature",
	"ClickHouseAsyncMetrics,cpu=2,instance=localhost:9363 OSUserTimeCPU",
	"ClickHouseProfileEvents,instance=localhost:9363 DictCacheRequests",
	"ClickHouseProfileEvents,instance=localhost:9363 KeeperReadSnapshot",
	"ClickHouseMetrics,instance=localhost:9363 PartsOutdated",
	"ClickHouseMetrics,instance=localhost:9363 KafkaBackgroundReads",
	"ClickHouseAsyncMetrics,instance=localhost:9363 FilesystemLogsPathAvailableBytes",
	"ClickHouseAsyncMetrics,cpu=5,instance=localhost:9363 OSNiceTimeCPU",
	"ClickHouseProfileEvents,instance=localhost:9363 DataAfterMutationDiffersFromReplica",
	"ClickHouseProfileEvents,instance=localhost:9363 PerfDataTLBReferences",
	"ClickHouseMetrics,instance=localhost:9363 KafkaLibrdkafkaThreads",
	"ClickHouseAsyncMetrics,cpu=5,instance=localhost:9363 OSGuestTimeCPU",
	"ClickHouseAsyncMetrics,instance=localhost:9363,unit=nvme0n1 BlockDiscardBytes",
	"ClickHouseProfileEvents,instance=localhost:9363 DirectorySync",
	"ClickHouseProfileEvents,instance=localhost:9363 DelayedInserts",
	"ClickHouseProfileEvents,instance=localhost:9363 RejectedInserts",
	"ClickHouseProfileEvents,instance=localhost:9363 SelectedMarks",
	"ClickHouseAsyncMetrics,cpu=5,instance=localhost:9363 OSGuestNiceTimeCPU",
	"ClickHouseAsyncMetrics,cpu=3,instance=localhost:9363 OSUserTimeCPU",
	"ClickHouseProfileEvents,instance=localhost:9363 RemoteFSPrefetches",
	"ClickHouseMetrics,instance=localhost:9363 AsyncDrainedConnections",
	"ClickHouseAsyncMetrics,cpu=0,instance=localhost:9363 OSGuestNiceTimeCPU",
	"ClickHouseProfileEvents,instance=localhost:9363 MergeTreeDataWriterRows",
	"ClickHouseProfileEvents,instance=localhost:9363 CannotRemoveEphemeralNode",
	"ClickHouseProfileEvents,instance=localhost:9363 DataAfterMergeDiffersFromReplica",
	"ClickHouseMetrics,instance=localhost:9363 RWLockActiveReaders",
	"ClickHouseAsyncMetrics,instance=localhost:9363 HTTPThreads",
	"ClickHouseAsyncMetrics,instance=localhost:9363,unit=iwlwifi_1 Temperature",
	"ClickHouseProfileEvents,instance=localhost:9363 DictCacheRequestTimeNs",
	"ClickHouseProfileEvents,instance=localhost:9363 CachedReadBufferReadFromCacheMicroseconds",
	"ClickHouseProfileEvents,instance=localhost:9363 KafkaMessagesFailed",
	"ClickHouseAsyncMetrics,instance=localhost:9363,unit=0 Temperature",
	"ClickHouseAsyncMetrics,cpu=3,instance=localhost:9363 OSGuestTimeCPU",
	"ClickHouseAsyncMetrics,cpu=4,instance=localhost:9363 OSGuestTimeCPU",
	"ClickHouseAsyncMetrics,instance=localhost:9363,unit=nvme0n1 BlockReadBytes",
	"ClickHouseAsyncMetrics,instance=localhost:9363,unit=sda BlockWriteTime",
	"ClickHouseProfileEvents,instance=localhost:9363 AsyncInsertQuery",
	"ClickHouseProfileEvents,instance=localhost:9363 ZooKeeperInit",
	"ClickHouseProfileEvents,instance=localhost:9363 ExternalAggregationMerge",
	"ClickHouseProfileEvents,instance=localhost:9363 DictCacheKeysExpired",
	"ClickHouseMetrics,instance=localhost:9363 ReplicatedChecks",
	"ClickHouseAsyncMetrics,instance=localhost:9363 jemalloc_metadata_thp",
	"ClickHouseAsyncMetrics,instance=localhost:9363 OSStealTimeNormalized",
	"ClickHouseAsyncMetrics,instance=localhost:9363 MemoryShared",
	"ClickHouseAsyncMetrics,instance=localhost:9363 OSUserTimeNormalized",
	"ClickHouseAsyncMetrics,instance=localhost:9363,unit=6 Temperature",
	"ClickHouseProfileEvents,instance=localhost:9363 ReadBackoff",
	"ClickHouseProfileEvents,instance=localhost:9363 OSWriteBytes",
	"ClickHouseProfileEvents,instance=localhost:9363 CachedReadBufferCacheWriteMicroseconds",
	"ClickHouseAsyncMetrics,instance=localhost:9363 TotalBytesOfMergeTreeTables",
	"ClickHouseProfileEvents,instance=localhost:9363 Merge",
	"ClickHouseProfileEvents,instance=localhost:9363 ThreadPoolReaderPageCacheHitElapsedMicroseconds",
	"ClickHouseAsyncMetrics,instance=localhost:9363 ReplicasSumQueueSize",
	"ClickHouseProfileEvents,instance=localhost:9363 OpenedFileCacheHits",
	"ClickHouseProfileEvents,instance=localhost:9363 ContextLock",
	"ClickHouseProfileEvents,instance=localhost:9363 RWLockReadersWaitMilliseconds",
	"ClickHouseProfileEvents,instance=localhost:9363 RemoteFSUnusedPrefetches",
	"ClickHouseProfileEvents,instance=localhost:9363 ThreadPoolReaderPageCacheMissElapsedMicroseconds",
	"ClickHouseAsyncMetrics,instance=localhost:9363 CompiledExpressionCacheBytes",
	"ClickHouseProfileEvents,instance=localhost:9363 ReadBufferFromFileDescriptorRead",
	"ClickHouseProfileEvents,instance=localhost:9363 MergeTreeDataWriterBlocksAlreadySorted",
	"ClickHouseProfileEvents,instance=localhost:9363 S3WriteRequestsThrottling",
	"ClickHouseAsyncMetrics,instance=localhost:9363 OSGuestNiceTimeNormalized",
	"ClickHouseAsyncMetrics,cpu=7,instance=localhost:9363 OSIOWaitTimeCPU",
	"ClickHouseAsyncMetrics,instance=localhost:9363 OSIdleTimeNormalized",
	"ClickHouseProfileEvents,instance=localhost:9363 FailedInsertQuery",
	"ClickHouseProfileEvents,instance=localhost:9363 WriteBufferFromFileDescriptorWrite",
	"ClickHouseProfileEvents,instance=localhost:9363 CreatedHTTPConnections",
	"ClickHouseMetrics,instance=localhost:9363 BackgroundBufferFlushSchedulePoolTask",
	"ClickHouseMetrics,instance=localhost:9363 InterserverConnection",
	"ClickHouseMetrics,instance=localhost:9363 PartsDeleting",
	"ClickHouseProfileEvents,instance=localhost:9363 DistributedDelayedInsertsMilliseconds",
	"ClickHouseProfileEvents,instance=localhost:9363 RemoteFSBuffers",
	"ClickHouseAsyncMetrics,instance=localhost:9363,unit=7 Temperature",
	"ClickHouseAsyncMetrics,instance=localhost:9363,unit=nvme0n1 BlockDiscardOps",
	"ClickHouseAsyncMetrics,cpu=3,instance=localhost:9363 OSIrqTimeCPU",
	"ClickHouseAsyncMetrics,instance=localhost:9363,unit=nvme0n1 BlockWriteBytes",
	"ClickHouseMetrics,instance=localhost:9363 Write",
	"ClickHouseMetrics,instance=localhost:9363 NetworkReceive",
	"ClickHouseAsyncMetrics,cpu=7,instance=localhost:9363 OSSystemTimeCPU",
	"ClickHouseAsyncMetrics,instance=localhost:9363,unit=3 Temperature",
	"ClickHouseAsyncMetrics,instance=localhost:9363 jemalloc_mapped",
	"ClickHouseProfileEvents,instance=localhost:9363 QueryMemoryLimitExceeded",
	"ClickHouseProfileEvents,instance=localhost:9363 MainConfigLoads",
	"ClickHouseProfileEvents,instance=localhost:9363 MergeTreeMetadataCacheMiss",
	"ClickHouseProfileEvents,instance=localhost:9363 KeeperSnapshotCreations",
	"ClickHouseMetrics,instance=localhost:9363 PartsTemporary",
	"ClickHouseMetrics,instance=localhost:9363 PartsInMemory",
	"ClickHouseProfileEvents,instance=localhost:9363 SelectQueryTimeMicroseconds",
	"ClickHouseProfileEvents,instance=localhost:9363 RealTimeMicroseconds",
	"ClickHouseAsyncMetrics,instance=localhost:9363 jemalloc_arenas_all_muzzy_purged",
	"ClickHouseAsyncMetrics,instance=localhost:9363,unit=sda BlockDiscardMerges",
	"ClickHouseAsyncMetrics,cpu=2,instance=localhost:9363 CPUFrequencyMHz",
	"ClickHouseAsyncMetrics,instance=localhost:9363 MarkCacheBytes",
	"ClickHouseProfileEvents,instance=localhost:9363 ReplicatedPartChecksFailed",
	"ClickHouseProfileEvents,instance=localhost:9363 MemoryOvercommitWaitTimeMicroseconds",
	"ClickHouseProfileEvents,instance=localhost:9363 PerfCacheReferences",
	"ClickHouseProfileEvents,instance=localhost:9363 ThreadpoolReaderTaskMicroseconds",
	"ClickHouseProfileEvents,instance=localhost:9363 KeeperPacketsReceived",
	"ClickHouseProfileEvents,instance=localhost:9363 DirectorySyncElapsedMicroseconds",
	"ClickHouseProfileEvents,instance=localhost:9363 PerfBranchMisses",
	"ClickHouseProfileEvents,instance=localhost:9363 PerfAlignmentFaults",
	"ClickHouseAsyncMetrics,instance=localhost:9363 TCPThreads",
	"ClickHouseProfileEvents,instance=localhost:9363 PolygonsInPoolAllocatedBytes",
	"ClickHouseProfileEvents,instance=localhost:9363 RWLockAcquiredReadLocks",
	"ClickHouseProfileEvents,instance=localhost:9363 SchemaInferenceCacheHits",
	"ClickHouseAsyncMetrics,disk=default,instance=localhost:9363 DiskTotal",
	"ClickHouseProfileEvents,instance=localhost:9363 MarkCacheHits",
	"ClickHouseProfileEvents,instance=localhost:9363 CachedWriteBufferCacheWriteMicroseconds",
	"ClickHouseProfileEvents,instance=localhost:9363 KafkaConsumerErrors",
	"ClickHouseAsyncMetrics,instance=localhost:9363 TotalPartsOfMergeTreeTables",
	"ClickHouseAsyncMetrics,cpu=0,instance=localhost:9363 OSStealTimeCPU",
	"ClickHouseProfileEvents,instance=localhost:9363 ZooKeeperSet",
	"ClickHouseMetrics,instance=localhost:9363 CacheDictionaryUpdateQueueKeys",
	"ClickHouseAsyncMetrics,instance=localhost:9363 ReplicasMaxQueueSize",
	"ClickHouseAsyncMetrics,instance=localhost:9363 CompiledExpressionCacheCount",
	"ClickHouseProfileEvents,instance=localhost:9363 FailedQuery",
	"ClickHouseProfileEvents,instance=localhost:9363 StorageBufferPassedBytesMaxThreshold",
	"ClickHouseMetrics,instance=localhost:9363 MaxDDLEntryID",
	"ClickHouseAsyncMetrics,instance=localhost:9363 NumberOfTables",
	"ClickHouseProfileEvents,instance=localhost:9363 StorageBufferLayerLockReadersWaitMilliseconds",
	"ClickHouseAsyncMetrics,instance=localhost:9363,unit=sda BlockDiscardBytes",
	"ClickHouseAsyncMetrics,instance=localhost:9363 OSThreadsTotal",
	"ClickHouseProfileEvents,instance=localhost:9363 ZooKeeperGet",
	"ClickHouseProfileEvents,instance=localhost:9363 ExternalSortMerge",
	"ClickHouseProfileEvents,instance=localhost:9363 MergeTreeDataWriterBlocks",
	"ClickHouseProfileEvents,instance=localhost:9363 InsertedWideParts",
	"ClickHouseProfileEvents,instance=localhost:9363 KafkaRowsWritten",
	"ClickHouseProfileEvents,instance=localhost:9363 DistributedConnectionFailAtAll",
	"ClickHouseAsyncMetrics,disk=default,instance=localhost:9363 DiskUsed",
	"ClickHouseProfileEvents,instance=localhost:9363 DictCacheLockWriteNs",
	"ClickHouseProfileEvents,instance=localhost:9363 PerfLocalMemoryReferences",
	"ClickHouseMetrics,instance=localhost:9363 BackgroundDistributedSchedulePoolTask",
	"ClickHouseAsyncMetrics,cpu=6,instance=localhost:9363 OSGuestTimeCPU",
	"ClickHouseAsyncMetrics,cpu=4,instance=localhost:9363 OSNiceTimeCPU",
	"ClickHouseProfileEvents,instance=localhost:9363 ReadBufferFromFileDescriptorReadBytes",
	"ClickHouseProfileEvents,instance=localhost:9363 DictCacheLockReadNs",
	"ClickHouseProfileEvents,instance=localhost:9363 RWLockAcquiredWriteLocks",
	"ClickHouseProfileEvents,instance=localhost:9363 MergeTreeMetadataCacheDelete",
	"ClickHouseMetrics,instance=localhost:9363 PartsDeleteOnDestroy",
	"ClickHouseAsyncMetrics,instance=localhost:9363,unit=sda BlockDiscardOps",
	"ClickHouseProfileEvents,instance=localhost:9363 CompiledFunctionExecute",
	"ClickHouseProfileEvents,instance=localhost:9363 MergeTreeMetadataCacheHit",
	"ClickHouseMetrics,instance=localhost:9363 DiskSpaceReservedForMerge",
	"ClickHouseMetrics,instance=localhost:9363 RWLockActiveWriters",
	"ClickHouseAsyncMetrics,instance=localhost:9363 OSMemoryCached",
	"ClickHouseAsyncMetrics,instance=localhost:9363,unit=nvme0n1 BlockInFlightOps",
	"ClickHouseMetrics,instance=localhost:9363 BackgroundMessageBrokerSchedulePoolTask",
	"ClickHouseAsyncMetrics,instance=localhost:9363 TotalRowsOfMergeTreeTables",
	"ClickHouseAsyncMetrics,instance=localhost:9363 OSNiceTimeNormalized",
	"ClickHouseProfileEvents,instance=localhost:9363 CreatedLogEntryForMutation",
	"ClickHouseProfileEvents,instance=localhost:9363 FileSegmentReadMicroseconds",
	"ClickHouseAsyncMetrics,instance=localhost:9363 NumberOfDatabases",
	"ClickHouseAsyncMetrics,instance=localhost:9363,unit=nvme_Sensor_1 Temperature",
	"ClickHouseAsyncMetrics,instance=localhost:9363 UncompressedCacheBytes",
	"ClickHouseAsyncMetrics,cpu=4,instance=localhost:9363 OSSystemTimeCPU",
	"ClickHouseProfileEvents,instance=localhost:9363 ZooKeeperCheck",
	"ClickHouseProfileEvents,instance=localhost:9363 StorageBufferPassedTimeMaxThreshold",
	"ClickHouseAsyncMetrics,cpu=5,instance=localhost:9363 OSIrqTimeCPU",
	"ClickHouseProfileEvents,instance=localhost:9363 AggregationHashTablesInitializedAsTwoLevel",
	"ClickHouseProfileEvents,instance=localhost:9363 OverflowAny",
	"ClickHouseAsyncMetrics,cpu=3,instance=localhost:9363 OSIOWaitTimeCPU",
	"ClickHouseAsyncMetrics,cpu=2,instance=localhost:9363 OSIOWaitTimeCPU",
	"ClickHouseProfileEvents,instance=localhost:9363 ReplicatedPartFailedFetches",
	"ClickHouseProfileEvents,instance=localhost:9363 MergeTreeDataProjectionWriterBlocks",
	"ClickHouseProfileEvents,instance=localhost:9363 ThreadPoolReaderPageCacheMissBytes",
	"ClickHouseMetrics,instance=localhost:9363 AsynchronousReadWait",
	"ClickHouseAsyncMetrics,instance=localhost:9363 OSSystemTime",
	"ClickHouseAsyncMetrics,cpu=6,instance=localhost:9363 OSSystemTimeCPU",
	"ClickHouseMetrics,instance=localhost:9363 KafkaWrites",
	"ClickHouseAsyncMetrics,instance=localhost:9363 FilesystemMainPathTotalINodes",
	"ClickHouseAsyncMetrics,cpu=1,instance=localhost:9363 OSIOWaitTimeCPU",
	"ClickHouseProfileEvents,instance=localhost:9363 QueryTimeMicroseconds",
	"ClickHouseProfileEvents,instance=localhost:9363 SystemTimeMicroseconds",
	"ClickHouseProfileEvents,instance=localhost:9363 SleepFunctionMicroseconds",
	"ClickHouseProfileEvents,instance=localhost:9363 KafkaRebalanceErrors",
	"ClickHouseMetrics,instance=localhost:9363 DictCacheRequests",
	"ClickHouseAsyncMetrics,instance=localhost:9363 OSStealTime",
	"ClickHouseProfileEvents,instance=localhost:9363 QueryMaskingRulesMatch",
	"ClickHouseMetrics,instance=localhost:9363 LocalThreadActive",
	"ClickHouseMetrics,instance=localhost:9363 CacheFileSegments",
	"ClickHouseAsyncMetrics,instance=localhost:9363 InterserverThreads",
	"ClickHouseProfileEvents,instance=localhost:9363 UncompressedCacheWeightLost",
	"ClickHouseProfileEvents,instance=localhost:9363 CreatedReadBufferOrdinary",
	"ClickHouseAsyncMetrics,instance=localhost:9363,unit=nvme_Composite Temperature",
	"ClickHouseAsyncMetrics,instance=localhost:9363 OSUptime",
	"ClickHouseProfileEvents,instance=localhost:9363 UncompressedCacheMisses",
	"ClickHouseProfileEvents,instance=localhost:9363 StorageBufferPassedRowsFlushThreshold",
	"ClickHouseProfileEvents,instance=localhost:9363 KafkaProducerErrors",
	"ClickHouseAsyncMetrics,instance=localhost:9363 FilesystemLogsPathTotalBytes",
	"ClickHouseAsyncMetrics,cpu=0,instance=localhost:9363 OSUserTimeCPU",
	"ClickHouseAsyncMetrics,instance=localhost:9363 OSIdleTime",
	"ClickHouseMetrics,instance=localhost:9363 BackgroundMovePoolTask",
	"ClickHouseAsyncMetrics,instance=localhost:9363,unit=pch_cannonlake Temperature",
	"ClickHouseProfileEvents,instance=localhost:9363 FileSync",
	"ClickHouseProfileEvents,instance=localhost:9363 ExternalSortWritePart",
	"ClickHouseProfileEvents,instance=localhost:9363 StorageBufferPassedRowsMaxThreshold",
	"ClickHouseProfileEvents,instance=localhost:9363 StorageBufferLayerLockWritersWaitMilliseconds",
	"ClickHouseProfileEvents,instance=localhost:9363 PerfContextSwitches",
	"ClickHouseProfileEvents,instance=localhost:9363 ThreadPoolReaderPageCacheHitBytes",
	"ClickHouseProfileEvents,instance=localhost:9363 ArenaAllocBytes",
	"ClickHouseProfileEvents,instance=localhost:9363 CachedWriteBufferCacheWriteBytes",
	"ClickHouseMetrics,instance=localhost:9363 QueryPreempted",
	"ClickHouseAsyncMetrics,instance=localhost:9363 FilesystemMainPathUsedBytes",
	"ClickHouseAsyncMetrics,instance=localhost:9363 jemalloc_arenas_all_dirty_purged",
	"ClickHouseAsyncMetrics,cpu=5,instance=localhost:9363 OSSystemTimeCPU",
	"ClickHouseProfileEvents,instance=localhost:9363 IOBufferAllocBytes",
	"ClickHouseProfileEvents,instance=localhost:9363 ZooKeeperSync",
	"ClickHouseProfileEvents,instance=localhost:9363 PerfBusCycles",
	"ClickHouseProfileEvents,instance=localhost:9363 KafkaCommitFailures",
	"ClickHouseProfileEvents,instance=localhost:9363 KeeperCommits",
	"ClickHouseMetrics,instance=localhost:9363 KeeperAliveConnections",
	"ClickHouseProfileEvents,instance=localhost:9363 Seek",
	"ClickHouseProfileEvents,instance=localhost:9363 PerfInstructionTLBReferences",
	"ClickHouseProfileEvents,instance=localhost:9363 FileSegmentUsedBytes",
	"ClickHouseAsyncMetrics,instance=localhost:9363 AsynchronousMetricsCalculationTimeSpent",
	"ClickHouseAsyncMetrics,instance=localhost:9363 MemoryResident",
	"ClickHouseProfileEvents,instance=localhost:9363 MergedUncompressedBytes",
	"ClickHouseProfileEvents,instance=localhost:9363 KafkaRebalanceRevocations",
	"ClickHouseMetrics,instance=localhost:9363 PartsCompact",
	"ClickHouseAsyncMetrics,instance=localhost:9363 OSProcessesCreated",
	"ClickHouseAsyncMetrics,instance=localhost:9363,unit=nvme0n1 BlockWriteOps",
	"ClickHouseProfileEvents,instance=localhost:9363 SchemaInferenceCacheMisses",
	"ClickHouseMetrics,instance=localhost:9363 FilesystemCacheSize",
	"ClickHouseProfileEvents,instance=localhost:9363 ThrottlerSleepMicroseconds",
	"ClickHouseProfileEvents,instance=localhost:9363 DelayedInsertsMilliseconds",
	"ClickHouseProfileEvents,instance=localhost:9363 ReadBufferFromS3Microseconds",
	"ClickHouseProfileEvents,instance=localhost:9363 KeeperSnapshotApplysFailed",
	"ClickHouseMetrics,instance=localhost:9363 TablesToDropQueueSize",
	"ClickHouseAsyncMetrics,instance=localhost:9363 OSMemoryTotal",
	"ClickHouseProfileEvents,instance=localhost:9363 KeeperCommitsFailed",
	"ClickHouseMetrics,instance=localhost:9363 GlobalThread",
	"ClickHouseMetrics,instance=localhost:9363 PartsWide",
	"ClickHouseAsyncMetrics,eth=eth0,instance=localhost:9363 NetworkReceiveErrors",
	"ClickHouseAsyncMetrics,cpu=4,instance=localhost:9363 OSIrqTimeCPU",
	"ClickHouseAsyncMetrics,cpu=5,instance=localhost:9363 OSIdleTimeCPU",
	"ClickHouseProfileEvents,instance=localhost:9363 NetworkReceiveBytes",
	"ClickHouseProfileEvents,instance=localhost:9363 SelectedRows",
	"ClickHouseMetrics,instance=localhost:9363 BackgroundMergesAndMutationsPoolTask",
	"ClickHouseAsyncMetrics,cpu=7,instance=localhost:9363 OSIdleTimeCPU",
	"ClickHouseAsyncMetrics,cpu=4,instance=localhost:9363 OSSoftIrqTimeCPU",
	"ClickHouseAsyncMetrics,instance=localhost:9363 OSGuestTime",
	"ClickHouseProfileEvents,instance=localhost:9363 StorageBufferErrorOnFlush",
	"ClickHouseProfileEvents,instance=localhost:9363 SchemaInferenceCacheInvalidations",
	"ClickHouseAsyncMetrics,instance=localhost:9363,unit=nvme0n1 BlockReadMerges",
	"ClickHouseAsyncMetrics,cpu=6,instance=localhost:9363 OSStealTimeCPU",
	"ClickHouseAsyncMetrics,cpu=5,instance=localhost:9363 OSUserTimeCPU",
	"ClickHouseProfileEvents,instance=localhost:9363 SchemaInferenceCacheEvictions",
	"ClickHouseAsyncMetrics,instance=localhost:9363 OSMemoryAvailable",
	"ClickHouseAsyncMetrics,instance=localhost:9363,unit=5 LoadAverage",
	"ClickHouseProfileEvents,instance=localhost:9363 OSCPUWaitMicroseconds",
	"ClickHouseAsyncMetrics,instance=localhost:9363 PrometheusThreads",
	"ClickHouseAsyncMetrics,instance=localhost:9363 MaxPartCountForPartition",
	"ClickHouseAsyncMetrics,cpu=2,instance=localhost:9363 OSNiceTimeCPU",
	"ClickHouseAsyncMetrics,instance=localhost:9363 MMapCacheCells",
	"ClickHouseMetrics,instance=localhost:9363 QueryThread",
	"ClickHouseAsyncMetrics,instance=localhost:9363 OSSystemTimeNormalized",
	"ClickHouseAsyncMetrics,cpu=7,instance=localhost:9363 OSSoftIrqTimeCPU",
	"ClickHouseAsyncMetrics,instance=localhost:9363 jemalloc_background_thread_num_threads",
	"ClickHouseProfileEvents,instance=localhost:9363 OSCPUVirtualTimeMicroseconds",
	"ClickHouseProfileEvents,instance=localhost:9363 CachedReadBufferCacheWriteBytes",
	"ClickHouseMetrics,instance=localhost:9363 EphemeralNode",
	"ClickHouseAsyncMetrics,instance=localhost:9363 jemalloc_active",
	"ClickHouseAsyncMetrics,instance=localhost:9363,unit=4 Temperature",
	"ClickHouseAsyncMetrics,cpu=6,instance=localhost:9363 OSNiceTimeCPU",
	"ClickHouseProfileEvents,instance=localhost:9363 RemoteFSCancelledPrefetches",
	"ClickHouseProfileEvents,instance=localhost:9363 OverflowThrow",
	"ClickHouseMetrics,instance=localhost:9363 MySQLConnection",
	"ClickHouseMetrics,instance=localhost:9363 DistributedFilesToInsert",
	"ClickHouseAsyncMetrics,cpu=2,instance=localhost:9363 OSSoftIrqTimeCPU",
	"ClickHouseAsyncMetrics,instance=localhost:9363,unit=1 LoadAverage",
	"ClickHouseAsyncMetrics,cpu=2,instance=localhost:9363 OSIdleTimeCPU",
	"ClickHouseAsyncMetrics,cpu=1,instance=localhost:9363 OSGuestNiceTimeCPU",
	"ClickHouseProfileEvents,instance=localhost:9363 AsyncInsertBytes",
	"ClickHouseProfileEvents,instance=localhost:9363 FileOpen",
	"ClickHouseProfileEvents,instance=localhost:9363 ZooKeeperWaitMicroseconds",
	"ClickHouseMetrics,instance=localhost:9363 CacheDetachedFileSegments",
	"ClickHouseAsyncMetrics,instance=localhost:9363 FilesystemLogsPathUsedBytes",
	"ClickHouseAsyncMetrics,instance=localhost:9363,unit=nvme0n1 BlockWriteMerges",
	"ClickHouseProfileEvents,instance=localhost:9363 RWLockWritersWaitMilliseconds",
	"ClickHouseProfileEvents,instance=localhost:9363 KeeperLatency",
	"ClickHouseMetrics,instance=localhost:9363 LocalThread",
	"ClickHouseAsyncMetrics,eth=eth0,instance=localhost:9363 NetworkReceiveBytes",
	"ClickHouseAsyncMetrics,cpu=6,instance=localhost:9363 OSIrqTimeCPU",
	"ClickHouseProfileEvents,instance=localhost:9363 KafkaProducerFlushes",
	"ClickHouseAsyncMetrics,eth=eth0,instance=localhost:9363 NetworkReceivePackets",
	"ClickHouseProfileEvents,instance=localhost:9363 AIOWrite",
	"ClickHouseProfileEvents,instance=localhost:9363 CompileExpressionsBytes",
	"ClickHouseProfileEvents,instance=localhost:9363 MergeTreeDataProjectionWriterBlocksAlreadySorted",
	"ClickHouseProfileEvents,instance=localhost:9363 CompressedReadBufferBlocks",
	"ClickHouseProfileEvents,instance=localhost:9363 AIOWriteBytes",
	"ClickHouseProfileEvents,instance=localhost:9363 ReplicatedPartMerges",
	"ClickHouseMetrics,instance=localhost:9363 StorageBufferBytes",
	"ClickHouseAsyncMetrics,instance=localhost:9363,unit=nvme_Sensor_2 Temperature",
	"ClickHouseProfileEvents,instance=localhost:9363 ExecuteShellCommand",
	"ClickHouseProfileEvents,instance=localhost:9363 MergedIntoWideParts",
	"ClickHouseProfileEvents,instance=localhost:9363 DictCacheKeysRequestedMiss",
	"ClickHouseProfileEvents,instance=localhost:9363 KafkaMessagesPolled",
	"ClickHouseMetrics,instance=localhost:9363 TCPConnection",
	"ClickHouseAsyncMetrics,instance=localhost:9363,unit=nvme0n1 BlockDiscardTime",
	"ClickHouseProfileEvents,instance=localhost:9363 CompileExpressionsMicroseconds",
	"ClickHouseProfileEvents,instance=localhost:9363 CannotWriteToWriteBufferDiscard",
	"ClickHouseMetrics,instance=localhost:9363 SendExternalTables",
	"ClickHouseProfileEvents,instance=localhost:9363 CreatedReadBufferMMapFailed",
	"ClickHouseProfileEvents,instance=localhost:9363 NetworkSendElapsedMicroseconds",
	"ClickHouseProfileEvents,instance=localhost:9363 AggregationPreallocatedElementsInHashTables",
	"ClickHouseProfileEvents,instance=localhost:9363 KeeperRequestTotal",
	"ClickHouseProfileEvents,instance=localhost:9363 KeeperSnapshotApplys",
	"ClickHouseAsyncMetrics,cpu=0,instance=localhost:9363 OSIOWaitTimeCPU",
	"ClickHouseAsyncMetrics,instance=localhost:9363 jemalloc_retained",
	"ClickHouseAsyncMetrics,cpu=1,instance=localhost:9363 OSGuestTimeCPU",
	"ClickHouseProfileEvents,instance=localhost:9363 CreatedReadBufferDirectIOFailed",
	"ClickHouseProfileEvents,instance=localhost:9363 InsertedBytes",
	"ClickHouseProfileEvents,instance=localhost:9363 ZooKeeperRemove",
	"ClickHouseProfileEvents,instance=localhost:9363 ReplicaPartialShutdown",
	"ClickHouseProfileEvents,instance=localhost:9363 NotCreatedLogEntryForMerge",
	"ClickHouseAsyncMetrics,instance=localhost:9363,unit=nvme0n1 BlockReadTime",
	"ClickHouseProfileEvents,instance=localhost:9363 NetworkSendBytes",
	"ClickHouseProfileEvents,instance=localhost:9363 DistributedConnectionStaleReplica",
	"ClickHouseMetrics,instance=localhost:9363 S3Requests",
	"ClickHouseAsyncMetrics,instance=localhost:9363 FilesystemMainPathTotalBytes",
	"ClickHouseAsyncMetrics,cpu=1,instance=localhost:9363 OSIrqTimeCPU",
	"ClickHouseAsyncMetrics,cpu=5,instance=localhost:9363 OSSoftIrqTimeCPU",
	"ClickHouseProfileEvents,instance=localhost:9363 ZooKeeperOtherExceptions",
	"ClickHouseProfileEvents,instance=localhost:9363 DistributedConnectionMissingTable",
	"ClickHouseProfileEvents,instance=localhost:9363 MergesTimeMilliseconds",
	"ClickHouseProfileEvents,instance=localhost:9363 MergedIntoCompactParts",
	"ClickHouseProfileEvents,instance=localhost:9363 StorageBufferPassedAllMinThresholds",
	"ClickHouseProfileEvents,instance=localhost:9363 S3ReadMicroseconds",
	"ClickHouseProfileEvents,instance=localhost:9363 DuplicatedInsertedBlocks",
	"ClickHouseProfileEvents,instance=localhost:9363 MergeTreeMetadataCachePut",
	"ClickHouseProfileEvents,instance=localhost:9363 KeeperSnapshotCreationsFailed",
	"ClickHouseMetrics,instance=localhost:9363 BackgroundSchedulePoolTask",
	"ClickHouseMetrics,instance=localhost:9363 CacheDictionaryUpdateQueueBatches",
	"ClickHouseMetrics,instance=localhost:9363 OpenFileForRead",
	"ClickHouseMetrics,instance=localhost:9363 KafkaConsumers",
	"ClickHouseAsyncMetrics,cpu=3,instance=localhost:9363 OSIdleTimeCPU",
	"ClickHouseAsyncMetrics,instance=localhost:9363,unit=sda BlockReadMerges",
	"ClickHouseAsyncMetrics,instance=localhost:9363,unit=2 Temperature",
	"ClickHouseAsyncMetrics,cpu=0,instance=localhost:9363 OSSystemTimeCPU",
	"ClickHouseProfileEvents,instance=localhost:9363 SelectedParts",
	"ClickHouseProfileEvents,instance=localhost:9363 PerfMinEnabledTime",
	"ClickHouseAsyncMetrics,instance=localhost:9363 jemalloc_allocated",
	"ClickHouseAsyncMetrics,instance=localhost:9363 jemalloc_epoch",
	"ClickHouseAsyncMetrics,cpu=7,instance=localhost:9363 OSUserTimeCPU",
	"ClickHouseProfileEvents,instance=localhost:9363 HardPageFaults",
	"ClickHouseProfileEvents,instance=localhost:9363 PerfBranchInstructions",
	"ClickHouseProfileEvents,instance=localhost:9363 ReadBufferSeekCancelConnection",
	"ClickHouseProfileEvents,instance=localhost:9363 KafkaMessagesRead",
	"ClickHouseAsyncMetrics,cpu=6,instance=localhost:9363 OSIOWaitTimeCPU",
	"ClickHouseMetrics,instance=localhost:9363 ZooKeeperRequest",
	"ClickHouseMetrics,instance=localhost:9363 KeeperOutstandingRequets",
	"ClickHouseAsyncMetrics,instance=localhost:9363 jemalloc_arenas_all_pactive",
	"ClickHouseProfileEvents,instance=localhost:9363 SoftPageFaults",
	"ClickHouseMetrics,instance=localhost:9363 PartsActive",
	"ClickHouseMetrics,instance=localhost:9363 ActiveAsyncDrainedConnections",
	"ClickHouseAsyncMetrics,cpu=1,instance=localhost:9363 OSSystemTimeCPU",
	"ClickHouseAsyncMetrics,instance=localhost:9363 Jitter",
	"ClickHouseAsyncMetrics,instance=localhost:9363 FilesystemLogsPathAvailableINodes",
	"ClickHouseAsyncMetrics,cpu=1,instance=localhost:9363 OSNiceTimeCPU",
	"ClickHouseProfileEvents,instance=localhost:9363 AIORead",
	"ClickHouseProfileEvents,instance=localhost:9363 ReplicatedPartMutations",
	"ClickHouseProfileEvents,instance=localhost:9363 OSIOWaitMicroseconds",
	"ClickHouseProfileEvents,instance=localhost:9363 S3WriteRequestsRedirects",
	"ClickHouseAsyncMetrics,cpu=2,instance=localhost:9363 OSSystemTimeCPU",
	"ClickHouseAsyncMetrics,cpu=1,instance=localhost:9363 OSSoftIrqTimeCPU",
	"ClickHouseAsyncMetrics,instance=localhost:9363 OSMemoryFreePlusCached",
	"ClickHouseAsyncMetrics,cpu=5,instance=localhost:9363 CPUFrequencyMHz",
	"ClickHouseProfileEvents,instance=localhost:9363 OSReadChars",
	"ClickHouseProfileEvents,instance=localhost:9363 InsertQuery",
	"ClickHouseProfileEvents,instance=localhost:9363 MarkCacheMisses",
	"ClickHouseProfileEvents,instance=localhost:9363 SlowRead",
	"ClickHouseProfileEvents,instance=localhost:9363 MergedRows",
	"ClickHouseAsyncMetrics,cpu=0,instance=localhost:9363 OSSoftIrqTimeCPU",
	"ClickHouseProfileEvents,instance=localhost:9363 KafkaWrites",
	"ClickHouseMetrics,instance=localhost:9363 KafkaAssignedPartitions",
	"ClickHouseProfileEvents,instance=localhost:9363 CreatedReadBufferMMap",
	"ClickHouseProfileEvents,instance=localhost:9363 ExternalAggregationCompressedBytes",
	"ClickHouseAsyncMetrics,instance=localhost:9363 FilesystemMainPathAvailableBytes",
	"ClickHouseAsyncMetrics,instance=localhost:9363,unit=coretemp_Package_id_0 Temperature",
	"ClickHouseAsyncMetrics,instance=localhost:9363,unit=nvme0n1 BlockReadOps",
	"ClickHouseAsyncMetrics,instance=localhost:9363 OSThreadsRunnable",
	"ClickHouseAsyncMetrics,eth=eth0,instance=localhost:9363 NetworkSendErrors",
	"ClickHouseAsyncMetrics,cpu=0,instance=localhost:9363 OSIdleTimeCPU",
	"ClickHouseProfileEvents,instance=localhost:9363 MergeTreeMetadataCacheGet",
	"ClickHouseProfileEvents,instance=localhost:9363 KafkaRowsRead",
	"ClickHouseProfileEvents,instance=localhost:9363 ScalarSubqueriesCacheMiss",
	"ClickHouseMetrics,instance=localhost:9363 BrokenDistributedFilesToInsert",
	"ClickHouseAsyncMetrics,instance=localhost:9363 jemalloc_resident",
	"ClickHouseAsyncMetrics,eth=eth0,instance=localhost:9363 NetworkSendDrop",
	"ClickHouseAsyncMetrics,instance=localhost:9363 MarkCacheFiles",
	"ClickHouseAsyncMetrics,cpu=4,instance=localhost:9363 CPUFrequencyMHz",
	"ClickHouseAsyncMetrics,instance=localhost:9363 OSIOWaitTimeNormalized",
	"ClickHouseProfileEvents,instance=localhost:9363 PolygonsAddedToPool",
	"ClickHouseProfileEvents,instance=localhost:9363 KafkaMessagesProduced",
	"ClickHouseProfileEvents,instance=localhost:9363 OverflowBreak",
	"ClickHouseMetrics,instance=localhost:9363 RWLockWaitingReaders",
	"ClickHouseAsyncMetrics,instance=localhost:9363 PostgreSQLThreads",
	"ClickHouseAsyncMetrics,eth=eth0,instance=localhost:9363 NetworkReceiveDrop",
	"ClickHouseAsyncMetrics,cpu=1,instance=localhost:9363 OSUserTimeCPU",
	"ClickHouseProfileEvents,instance=localhost:9363 DictCacheKeysNotFound",
	"ClickHouseProfileEvents,instance=localhost:9363 DistributedSyncInsertionTimeoutExceeded",
	"ClickHouseProfileEvents,instance=localhost:9363 NotCreatedLogEntryForMutation",
	"ClickHouseMetrics,instance=localhost:9363 RWLockWaitingWriters",
	"ClickHouseAsyncMetrics,instance=localhost:9363 ReplicasMaxMergesInQueue",
	"ClickHouseAsyncMetrics,instance=localhost:9363,unit=nvme0n1 BlockQueueTime",
	"ClickHouseProfileEvents,instance=localhost:9363 OpenedFileCacheMisses",
	"ClickHouseProfileEvents,instance=localhost:9363 AIOReadBytes",
	"ClickHouseAsyncMetrics,instance=localhost:9363 jemalloc_arenas_all_pmuzzy",
	"ClickHouseAsyncMetrics,instance=localhost:9363,unit=nvme0n1 BlockWriteTime",
	"ClickHouseAsyncMetrics,cpu=4,instance=localhost:9363 OSIOWaitTimeCPU",
	"ClickHouseAsyncMetrics,instance=localhost:9363,unit=coretemp_Core_1 Temperature",
	"ClickHouseAsyncMetrics,instance=localhost:9363 OSIOWaitTime",
	"ClickHouseProfileEvents,instance=localhost:9363 NetworkReceiveElapsedMicroseconds",
	"ClickHouseProfileEvents,instance=localhost:9363 ReplicatedPartChecks",
	"ClickHouseProfileEvents,instance=localhost:9363 DistributedRejectedInserts",
	"ClickHouseProfileEvents,instance=localhost:9363 MergeTreeDataProjectionWriterRows",
	"ClickHouseProfileEvents,instance=localhost:9363 S3ReadRequestsRedirects",
	"ClickHouseMetrics,instance=localhost:9363 KafkaConsumersWithAssignment",
	"ClickHouseProfileEvents,instance=localhost:9363 OSWriteChars",
	"ClickHouseAsyncMetrics,cpu=1,instance=localhost:9363 OSStealTimeCPU",
	"ClickHouseAsyncMetrics,cpu=2,instance=localhost:9363 OSIrqTimeCPU",
	"ClickHouseAsyncMetrics,cpu=5,instance=localhost:9363 OSStealTimeCPU",
	"ClickHouseProfileEvents,instance=localhost:9363 OtherQueryTimeMicroseconds",
	"ClickHouseProfileEvents,instance=localhost:9363 ZooKeeperCreate",
	"ClickHouseProfileEvents,instance=localhost:9363 InsertedCompactParts",
	"ClickHouseProfileEvents,instance=localhost:9363 StorageBufferPassedBytesFlushThreshold",
	"ClickHouseProfileEvents,instance=localhost:9363 S3ReadRequestsErrors",
	"ClickHouseAsyncMetrics,instance=localhost:9363 OSNiceTime",
	"ClickHouseProfileEvents,instance=localhost:9363 ExternalAggregationWritePart",
	"ClickHouseProfileEvents,instance=localhost:9363 DNSError",
	"ClickHouseProfileEvents,instance=localhost:9363 S3WriteMicroseconds",
	"ClickHouseProfileEvents,instance=localhost:9363 RegexpCreated",
	"ClickHouseMetrics,instance=localhost:9363 MMappedFileBytes",
	"ClickHouseAsyncMetrics,cpu=4,instance=localhost:9363 OSUserTimeCPU",
	"ClickHouseAsyncMetrics,cpu=2,instance=localhost:9363 OSStealTimeCPU",
	"ClickHouseProfileEvents,instance=localhost:9363 WriteBufferFromFileDescriptorWriteFailed",
	"ClickHouseMetrics,instance=localhost:9363 KafkaProducers",
	"ClickHouseAsyncMetrics,cpu=1,instance=localhost:9363 CPUFrequencyMHz",
	"ClickHouseAsyncMetrics,instance=localhost:9363 OSProcessesRunning",
	"ClickHouseProfileEvents,instance=localhost:9363 InsertQueryTimeMicroseconds",
	"ClickHouseProfileEvents,instance=localhost:9363 WriteBufferFromFileDescriptorWriteBytes",
	"ClickHouseProfileEvents,instance=localhost:9363 KafkaRowsRejected",
	"ClickHouseAsyncMetrics,cpu=6,instance=localhost:9363 OSIdleTimeCPU",
	"ClickHouseAsyncMetrics,cpu=0,instance=localhost:9363 CPUFrequencyMHz",
	"ClickHouseProfileEvents,instance=localhost:9363 UncompressedCacheHits",
	"ClickHouseMetrics,instance=localhost:9363 VersionInteger",
	"ClickHouseAsyncMetrics,instance=localhost:9363 OSIrqTimeNormalized",
	"ClickHouseProfileEvents,instance=localhost:9363 ZooKeeperExists",
	"ClickHouseProfileEvents,instance=localhost:9363 InsertedInMemoryParts",
	"ClickHouseAsyncMetrics,cpu=7,instance=localhost:9363 OSGuestNiceTimeCPU",
	"ClickHouseAsyncMetrics,instance=localhost:9363,unit=sda BlockDiscardTime",
	"ClickHouseProfileEvents,instance=localhost:9363 FunctionExecute",
	"ClickHouseMetrics,instance=localhost:9363 Merge",
	"ClickHouseMetrics,instance=localhost:9363 ActiveSyncDrainedConnections",
	"ClickHouseAsyncMetrics,cpu=3,instance=localhost:9363 OSGuestNiceTimeCPU",
	"ClickHouseAsyncMetrics,cpu=7,instance=localhost:9363 OSGuestTimeCPU",
	"ClickHouseAsyncMetrics,cpu=4,instance=localhost:9363 OSGuestNiceTimeCPU",
	"ClickHouseProfileEvents,instance=localhost:9363 ZooKeeperUserExceptions",
	"ClickHouseProfileEvents,instance=localhost:9363 PerfStalledCyclesFrontend",
	"ClickHouseProfileEvents,instance=localhost:9363 RemoteFSSeeks",
	"ClickHouseMetrics,instance=localhost:9363 PostgreSQLConnection",
	"ClickHouseAsyncMetrics,instance=localhost:9363,unit=sda BlockInFlightOps",
	"ClickHouseAsyncMetrics,instance=localhost:9363 OSMemoryBuffers",
	"ClickHouseMetrics,instance=localhost:9363 MMappedFiles",
	"ClickHouseAsyncMetrics,instance=localhost:9363 FilesystemMainPathAvailableINodes",
	"ClickHouseAsyncMetrics,instance=localhost:9363 OSSoftIrqTimeNormalized",
	"ClickHouseAsyncMetrics,cpu=3,instance=localhost:9363 OSNiceTimeCPU",
	"ClickHouseMetrics,instance=localhost:9363 BackgroundCommonPoolTask",
	"ClickHouseMetrics,instance=localhost:9363 PartsPreCommitted",
	"ClickHouseProfileEvents,instance=localhost:9363 SelectQuery",
	"ClickHouseProfileEvents,instance=localhost:9363 MergeTreeDataWriterUncompressedBytes",
	"ClickHouseProfileEvents,instance=localhost:9363 DictCacheKeysRequested",
	"ClickHouseProfileEvents,instance=localhost:9363 PerfStalledCyclesBackend",
	"ClickHouseProfileEvents,instance=localhost:9363 SleepFunctionCalls",
	"ClickHouseProfileEvents,instance=localhost:9363 KafkaDirectReads",
	"ClickHouseAsyncMetrics,cpu=4,instance=localhost:9363 OSStealTimeCPU",
	"ClickHouseProfileEvents,instance=localhost:9363 ReplicatedPartFetches",
	"ClickHouseProfileEvents,instance=localhost:9363 ZooKeeperClose",
	"ClickHouseProfileEvents,instance=localhost:9363 KafkaBackgroundReads",
	"ClickHouseAsyncMetrics,cpu=7,instance=localhost:9363 OSIrqTimeCPU",
	"ClickHouseMetrics,instance=localhost:9363 PartsCommitted",
	"ClickHouseMetrics,instance=localhost:9363 FilesystemCacheReadBuffers",
	"ClickHouseProfileEvents,instance=localhost:9363 ReadCompressedBytes",
	"ClickHouseProfileEvents,instance=localhost:9363 ZooKeeperBytesSent",
	"ClickHouseProfileEvents,instance=localhost:9363 PerfInstructionTLBMisses",
	"ClickHouseProfileEvents,instance=localhost:9363 PerfLocalMemoryMisses",
	"ClickHouseProfileEvents,instance=localhost:9363 CreatedLogEntryForMerge",
	"ClickHouseProfileEvents,instance=localhost:9363 CachedReadBufferReadFromCacheBytes",
	"ClickHouseAsyncMetrics,instance=localhost:9363 OSUserTime",
	"ClickHouseAsyncMetrics,instance=localhost:9363,unit=coretemp_Core_0 Temperature",
	"ClickHouseAsyncMetrics,instance=localhost:9363,unit=average BlockActiveTime",
	"ClickHouseAsyncMetrics,instance=localhost:9363,unit=total BlockDiscardBytes",
	"ClickHouseAsyncMetrics,instance=localhost:9363,unit=total BlockDiscardMerges",
	"ClickHouseAsyncMetrics,instance=localhost:9363,unit=total BlockDiscardOps",
	"ClickHouseAsyncMetrics,instance=localhost:9363,unit=average BlockDiscardTime",
	"ClickHouseAsyncMetrics,instance=localhost:9363,unit=total BlockInFlightOps",
	"ClickHouseAsyncMetrics,instance=localhost:9363,unit=average BlockQueueTime",
	"ClickHouseAsyncMetrics,instance=localhost:9363,unit=total BlockReadBytes",
	"ClickHouseAsyncMetrics,instance=localhost:9363,unit=total BlockReadMerges",
	"ClickHouseAsyncMetrics,instance=localhost:9363,unit=total BlockReadOps",
	"ClickHouseAsyncMetrics,instance=localhost:9363,unit=average BlockReadTime",
	"ClickHouseAsyncMetrics,instance=localhost:9363,unit=total BlockWriteBytes",
	"ClickHouseAsyncMetrics,instance=localhost:9363,unit=total BlockWriteMerges",
	"ClickHouseAsyncMetrics,instance=localhost:9363,unit=total BlockWriteOps",
	"ClickHouseAsyncMetrics,instance=localhost:9363,unit=average BlockWriteTime",
	"ClickHouseAsyncMetrics,cpu=average,instance=localhost:9363 CPUFrequencyMHz",
	"ClickHouseAsyncMetrics,disk=total,instance=localhost:9363 DiskAvailable",
	"ClickHouseAsyncMetrics,disk=total,instance=localhost:9363 DiskTotal",
	"ClickHouseAsyncMetrics,disk=total,instance=localhost:9363 DiskUnreserved",
	"ClickHouseAsyncMetrics,disk=total,instance=localhost:9363 DiskUsed",
	"ClickHouseAsyncMetrics,instance=localhost:9363,unit=average LoadAverage",
	"ClickHouseAsyncMetrics,eth=total,instance=localhost:9363 NetworkReceiveBytes",
	"ClickHouseAsyncMetrics,eth=total,instance=localhost:9363 NetworkReceiveDrop",
	"ClickHouseAsyncMetrics,eth=total,instance=localhost:9363 NetworkReceiveErrors",
	"ClickHouseAsyncMetrics,eth=total,instance=localhost:9363 NetworkReceivePackets",
	"ClickHouseAsyncMetrics,eth=total,instance=localhost:9363 NetworkSendBytes",
	"ClickHouseAsyncMetrics,eth=total,instance=localhost:9363 NetworkSendDrop",
	"ClickHouseAsyncMetrics,eth=total,instance=localhost:9363 NetworkSendErrors",
	"ClickHouseAsyncMetrics,eth=total,instance=localhost:9363 NetworkSendPackets",
	"ClickHouseAsyncMetrics,cpu=average,instance=localhost:9363 OSGuestNiceTimeCPU",
	"ClickHouseAsyncMetrics,cpu=average,instance=localhost:9363 OSGuestTimeCPU",
	"ClickHouseAsyncMetrics,cpu=average,instance=localhost:9363 OSIOWaitTimeCPU",
	"ClickHouseAsyncMetrics,cpu=average,instance=localhost:9363 OSIdleTimeCPU",
	"ClickHouseAsyncMetrics,cpu=average,instance=localhost:9363 OSIrqTimeCPU",
	"ClickHouseAsyncMetrics,cpu=average,instance=localhost:9363 OSNiceTimeCPU",
	"ClickHouseAsyncMetrics,cpu=average,instance=localhost:9363 OSSoftIrqTimeCPU",
	"ClickHouseAsyncMetrics,cpu=average,instance=localhost:9363 OSStealTimeCPU",
	"ClickHouseAsyncMetrics,cpu=average,instance=localhost:9363 OSSystemTimeCPU",
	"ClickHouseAsyncMetrics,cpu=average,instance=localhost:9363 OSUserTimeCPU",
	"ClickHouseAsyncMetrics,instance=localhost:9363,unit=average Temperature",
}

var want_v22_8_15_23_no_election = []string{
	"ClickHouseAsyncMetrics,cpu=0,host=me,instance=localhost:9363 CPUFrequencyMHz",
	"ClickHouseAsyncMetrics,cpu=0,host=me,instance=localhost:9363 OSGuestNiceTimeCPU",
	"ClickHouseAsyncMetrics,cpu=0,host=me,instance=localhost:9363 OSGuestTimeCPU",
	"ClickHouseAsyncMetrics,cpu=0,host=me,instance=localhost:9363 OSIOWaitTimeCPU",
	"ClickHouseAsyncMetrics,cpu=0,host=me,instance=localhost:9363 OSIdleTimeCPU",
	"ClickHouseAsyncMetrics,cpu=0,host=me,instance=localhost:9363 OSIrqTimeCPU",
	"ClickHouseAsyncMetrics,cpu=0,host=me,instance=localhost:9363 OSNiceTimeCPU",
	"ClickHouseAsyncMetrics,cpu=0,host=me,instance=localhost:9363 OSSoftIrqTimeCPU",
	"ClickHouseAsyncMetrics,cpu=0,host=me,instance=localhost:9363 OSStealTimeCPU",
	"ClickHouseAsyncMetrics,cpu=0,host=me,instance=localhost:9363 OSSystemTimeCPU",
	"ClickHouseAsyncMetrics,cpu=0,host=me,instance=localhost:9363 OSUserTimeCPU",
	"ClickHouseAsyncMetrics,cpu=1,host=me,instance=localhost:9363 CPUFrequencyMHz",
	"ClickHouseAsyncMetrics,cpu=1,host=me,instance=localhost:9363 OSGuestNiceTimeCPU",
	"ClickHouseAsyncMetrics,cpu=1,host=me,instance=localhost:9363 OSGuestTimeCPU",
	"ClickHouseAsyncMetrics,cpu=1,host=me,instance=localhost:9363 OSIOWaitTimeCPU",
	"ClickHouseAsyncMetrics,cpu=1,host=me,instance=localhost:9363 OSIdleTimeCPU",
	"ClickHouseAsyncMetrics,cpu=1,host=me,instance=localhost:9363 OSIrqTimeCPU",
	"ClickHouseAsyncMetrics,cpu=1,host=me,instance=localhost:9363 OSNiceTimeCPU",
	"ClickHouseAsyncMetrics,cpu=1,host=me,instance=localhost:9363 OSSoftIrqTimeCPU",
	"ClickHouseAsyncMetrics,cpu=1,host=me,instance=localhost:9363 OSStealTimeCPU",
	"ClickHouseAsyncMetrics,cpu=1,host=me,instance=localhost:9363 OSSystemTimeCPU",
	"ClickHouseAsyncMetrics,cpu=1,host=me,instance=localhost:9363 OSUserTimeCPU",
	"ClickHouseAsyncMetrics,cpu=2,host=me,instance=localhost:9363 CPUFrequencyMHz",
	"ClickHouseAsyncMetrics,cpu=2,host=me,instance=localhost:9363 OSGuestNiceTimeCPU",
	"ClickHouseAsyncMetrics,cpu=2,host=me,instance=localhost:9363 OSGuestTimeCPU",
	"ClickHouseAsyncMetrics,cpu=2,host=me,instance=localhost:9363 OSIOWaitTimeCPU",
	"ClickHouseAsyncMetrics,cpu=2,host=me,instance=localhost:9363 OSIdleTimeCPU",
	"ClickHouseAsyncMetrics,cpu=2,host=me,instance=localhost:9363 OSIrqTimeCPU",
	"ClickHouseAsyncMetrics,cpu=2,host=me,instance=localhost:9363 OSNiceTimeCPU",
	"ClickHouseAsyncMetrics,cpu=2,host=me,instance=localhost:9363 OSSoftIrqTimeCPU",
	"ClickHouseAsyncMetrics,cpu=2,host=me,instance=localhost:9363 OSStealTimeCPU",
	"ClickHouseAsyncMetrics,cpu=2,host=me,instance=localhost:9363 OSSystemTimeCPU",
	"ClickHouseAsyncMetrics,cpu=2,host=me,instance=localhost:9363 OSUserTimeCPU",
	"ClickHouseAsyncMetrics,cpu=3,host=me,instance=localhost:9363 CPUFrequencyMHz",
	"ClickHouseAsyncMetrics,cpu=3,host=me,instance=localhost:9363 OSGuestNiceTimeCPU",
	"ClickHouseAsyncMetrics,cpu=3,host=me,instance=localhost:9363 OSGuestTimeCPU",
	"ClickHouseAsyncMetrics,cpu=3,host=me,instance=localhost:9363 OSIOWaitTimeCPU",
	"ClickHouseAsyncMetrics,cpu=3,host=me,instance=localhost:9363 OSIdleTimeCPU",
	"ClickHouseAsyncMetrics,cpu=3,host=me,instance=localhost:9363 OSIrqTimeCPU",
	"ClickHouseAsyncMetrics,cpu=3,host=me,instance=localhost:9363 OSNiceTimeCPU",
	"ClickHouseAsyncMetrics,cpu=3,host=me,instance=localhost:9363 OSSoftIrqTimeCPU",
	"ClickHouseAsyncMetrics,cpu=3,host=me,instance=localhost:9363 OSStealTimeCPU",
	"ClickHouseAsyncMetrics,cpu=3,host=me,instance=localhost:9363 OSSystemTimeCPU",
	"ClickHouseAsyncMetrics,cpu=3,host=me,instance=localhost:9363 OSUserTimeCPU",
	"ClickHouseAsyncMetrics,cpu=4,host=me,instance=localhost:9363 CPUFrequencyMHz",
	"ClickHouseAsyncMetrics,cpu=4,host=me,instance=localhost:9363 OSGuestNiceTimeCPU",
	"ClickHouseAsyncMetrics,cpu=4,host=me,instance=localhost:9363 OSGuestTimeCPU",
	"ClickHouseAsyncMetrics,cpu=4,host=me,instance=localhost:9363 OSIOWaitTimeCPU",
	"ClickHouseAsyncMetrics,cpu=4,host=me,instance=localhost:9363 OSIdleTimeCPU",
	"ClickHouseAsyncMetrics,cpu=4,host=me,instance=localhost:9363 OSIrqTimeCPU",
	"ClickHouseAsyncMetrics,cpu=4,host=me,instance=localhost:9363 OSNiceTimeCPU",
	"ClickHouseAsyncMetrics,cpu=4,host=me,instance=localhost:9363 OSSoftIrqTimeCPU",
	"ClickHouseAsyncMetrics,cpu=4,host=me,instance=localhost:9363 OSStealTimeCPU",
	"ClickHouseAsyncMetrics,cpu=4,host=me,instance=localhost:9363 OSSystemTimeCPU",
	"ClickHouseAsyncMetrics,cpu=4,host=me,instance=localhost:9363 OSUserTimeCPU",
	"ClickHouseAsyncMetrics,cpu=5,host=me,instance=localhost:9363 CPUFrequencyMHz",
	"ClickHouseAsyncMetrics,cpu=5,host=me,instance=localhost:9363 OSGuestNiceTimeCPU",
	"ClickHouseAsyncMetrics,cpu=5,host=me,instance=localhost:9363 OSGuestTimeCPU",
	"ClickHouseAsyncMetrics,cpu=5,host=me,instance=localhost:9363 OSIOWaitTimeCPU",
	"ClickHouseAsyncMetrics,cpu=5,host=me,instance=localhost:9363 OSIdleTimeCPU",
	"ClickHouseAsyncMetrics,cpu=5,host=me,instance=localhost:9363 OSIrqTimeCPU",
	"ClickHouseAsyncMetrics,cpu=5,host=me,instance=localhost:9363 OSNiceTimeCPU",
	"ClickHouseAsyncMetrics,cpu=5,host=me,instance=localhost:9363 OSSoftIrqTimeCPU",
	"ClickHouseAsyncMetrics,cpu=5,host=me,instance=localhost:9363 OSStealTimeCPU",
	"ClickHouseAsyncMetrics,cpu=5,host=me,instance=localhost:9363 OSSystemTimeCPU",
	"ClickHouseAsyncMetrics,cpu=5,host=me,instance=localhost:9363 OSUserTimeCPU",
	"ClickHouseAsyncMetrics,cpu=6,host=me,instance=localhost:9363 CPUFrequencyMHz",
	"ClickHouseAsyncMetrics,cpu=6,host=me,instance=localhost:9363 OSGuestNiceTimeCPU",
	"ClickHouseAsyncMetrics,cpu=6,host=me,instance=localhost:9363 OSGuestTimeCPU",
	"ClickHouseAsyncMetrics,cpu=6,host=me,instance=localhost:9363 OSIOWaitTimeCPU",
	"ClickHouseAsyncMetrics,cpu=6,host=me,instance=localhost:9363 OSIdleTimeCPU",
	"ClickHouseAsyncMetrics,cpu=6,host=me,instance=localhost:9363 OSIrqTimeCPU",
	"ClickHouseAsyncMetrics,cpu=6,host=me,instance=localhost:9363 OSNiceTimeCPU",
	"ClickHouseAsyncMetrics,cpu=6,host=me,instance=localhost:9363 OSSoftIrqTimeCPU",
	"ClickHouseAsyncMetrics,cpu=6,host=me,instance=localhost:9363 OSStealTimeCPU",
	"ClickHouseAsyncMetrics,cpu=6,host=me,instance=localhost:9363 OSSystemTimeCPU",
	"ClickHouseAsyncMetrics,cpu=6,host=me,instance=localhost:9363 OSUserTimeCPU",
	"ClickHouseAsyncMetrics,cpu=7,host=me,instance=localhost:9363 CPUFrequencyMHz",
	"ClickHouseAsyncMetrics,cpu=7,host=me,instance=localhost:9363 OSGuestNiceTimeCPU",
	"ClickHouseAsyncMetrics,cpu=7,host=me,instance=localhost:9363 OSGuestTimeCPU",
	"ClickHouseAsyncMetrics,cpu=7,host=me,instance=localhost:9363 OSIOWaitTimeCPU",
	"ClickHouseAsyncMetrics,cpu=7,host=me,instance=localhost:9363 OSIdleTimeCPU",
	"ClickHouseAsyncMetrics,cpu=7,host=me,instance=localhost:9363 OSIrqTimeCPU",
	"ClickHouseAsyncMetrics,cpu=7,host=me,instance=localhost:9363 OSNiceTimeCPU",
	"ClickHouseAsyncMetrics,cpu=7,host=me,instance=localhost:9363 OSSoftIrqTimeCPU",
	"ClickHouseAsyncMetrics,cpu=7,host=me,instance=localhost:9363 OSStealTimeCPU",
	"ClickHouseAsyncMetrics,cpu=7,host=me,instance=localhost:9363 OSSystemTimeCPU",
	"ClickHouseAsyncMetrics,cpu=7,host=me,instance=localhost:9363 OSUserTimeCPU",
	"ClickHouseAsyncMetrics,cpu=average,host=me,instance=localhost:9363 CPUFrequencyMHz",
	"ClickHouseAsyncMetrics,cpu=average,host=me,instance=localhost:9363 OSGuestNiceTimeCPU",
	"ClickHouseAsyncMetrics,cpu=average,host=me,instance=localhost:9363 OSGuestTimeCPU",
	"ClickHouseAsyncMetrics,cpu=average,host=me,instance=localhost:9363 OSIOWaitTimeCPU",
	"ClickHouseAsyncMetrics,cpu=average,host=me,instance=localhost:9363 OSIdleTimeCPU",
	"ClickHouseAsyncMetrics,cpu=average,host=me,instance=localhost:9363 OSIrqTimeCPU",
	"ClickHouseAsyncMetrics,cpu=average,host=me,instance=localhost:9363 OSNiceTimeCPU",
	"ClickHouseAsyncMetrics,cpu=average,host=me,instance=localhost:9363 OSSoftIrqTimeCPU",
	"ClickHouseAsyncMetrics,cpu=average,host=me,instance=localhost:9363 OSStealTimeCPU",
	"ClickHouseAsyncMetrics,cpu=average,host=me,instance=localhost:9363 OSSystemTimeCPU",
	"ClickHouseAsyncMetrics,cpu=average,host=me,instance=localhost:9363 OSUserTimeCPU",
	"ClickHouseAsyncMetrics,disk=default,host=me,instance=localhost:9363 DiskAvailable",
	"ClickHouseAsyncMetrics,disk=default,host=me,instance=localhost:9363 DiskTotal",
	"ClickHouseAsyncMetrics,disk=default,host=me,instance=localhost:9363 DiskUnreserved",
	"ClickHouseAsyncMetrics,disk=default,host=me,instance=localhost:9363 DiskUsed",
	"ClickHouseAsyncMetrics,disk=total,host=me,instance=localhost:9363 DiskAvailable",
	"ClickHouseAsyncMetrics,disk=total,host=me,instance=localhost:9363 DiskTotal",
	"ClickHouseAsyncMetrics,disk=total,host=me,instance=localhost:9363 DiskUnreserved",
	"ClickHouseAsyncMetrics,disk=total,host=me,instance=localhost:9363 DiskUsed",
	"ClickHouseAsyncMetrics,eth=eth0,host=me,instance=localhost:9363 NetworkReceiveBytes",
	"ClickHouseAsyncMetrics,eth=eth0,host=me,instance=localhost:9363 NetworkReceiveDrop",
	"ClickHouseAsyncMetrics,eth=eth0,host=me,instance=localhost:9363 NetworkReceiveErrors",
	"ClickHouseAsyncMetrics,eth=eth0,host=me,instance=localhost:9363 NetworkReceivePackets",
	"ClickHouseAsyncMetrics,eth=eth0,host=me,instance=localhost:9363 NetworkSendBytes",
	"ClickHouseAsyncMetrics,eth=eth0,host=me,instance=localhost:9363 NetworkSendDrop",
	"ClickHouseAsyncMetrics,eth=eth0,host=me,instance=localhost:9363 NetworkSendErrors",
	"ClickHouseAsyncMetrics,eth=eth0,host=me,instance=localhost:9363 NetworkSendPackets",
	"ClickHouseAsyncMetrics,eth=total,host=me,instance=localhost:9363 NetworkReceiveBytes",
	"ClickHouseAsyncMetrics,eth=total,host=me,instance=localhost:9363 NetworkReceiveDrop",
	"ClickHouseAsyncMetrics,eth=total,host=me,instance=localhost:9363 NetworkReceiveErrors",
	"ClickHouseAsyncMetrics,eth=total,host=me,instance=localhost:9363 NetworkReceivePackets",
	"ClickHouseAsyncMetrics,eth=total,host=me,instance=localhost:9363 NetworkSendBytes",
	"ClickHouseAsyncMetrics,eth=total,host=me,instance=localhost:9363 NetworkSendDrop",
	"ClickHouseAsyncMetrics,eth=total,host=me,instance=localhost:9363 NetworkSendErrors",
	"ClickHouseAsyncMetrics,eth=total,host=me,instance=localhost:9363 NetworkSendPackets",
	"ClickHouseAsyncMetrics,host=me,instance=localhost:9363 AsynchronousMetricsCalculationTimeSpent",
	"ClickHouseAsyncMetrics,host=me,instance=localhost:9363 CompiledExpressionCacheBytes",
	"ClickHouseAsyncMetrics,host=me,instance=localhost:9363 CompiledExpressionCacheCount",
	"ClickHouseAsyncMetrics,host=me,instance=localhost:9363 FilesystemLogsPathAvailableBytes",
	"ClickHouseAsyncMetrics,host=me,instance=localhost:9363 FilesystemLogsPathAvailableINodes",
	"ClickHouseAsyncMetrics,host=me,instance=localhost:9363 FilesystemLogsPathTotalBytes",
	"ClickHouseAsyncMetrics,host=me,instance=localhost:9363 FilesystemLogsPathTotalINodes",
	"ClickHouseAsyncMetrics,host=me,instance=localhost:9363 FilesystemLogsPathUsedBytes",
	"ClickHouseAsyncMetrics,host=me,instance=localhost:9363 FilesystemLogsPathUsedINodes",
	"ClickHouseAsyncMetrics,host=me,instance=localhost:9363 FilesystemMainPathAvailableBytes",
	"ClickHouseAsyncMetrics,host=me,instance=localhost:9363 FilesystemMainPathAvailableINodes",
	"ClickHouseAsyncMetrics,host=me,instance=localhost:9363 FilesystemMainPathTotalBytes",
	"ClickHouseAsyncMetrics,host=me,instance=localhost:9363 FilesystemMainPathTotalINodes",
	"ClickHouseAsyncMetrics,host=me,instance=localhost:9363 FilesystemMainPathUsedBytes",
	"ClickHouseAsyncMetrics,host=me,instance=localhost:9363 FilesystemMainPathUsedINodes",
	"ClickHouseAsyncMetrics,host=me,instance=localhost:9363 HTTPThreads",
	"ClickHouseAsyncMetrics,host=me,instance=localhost:9363 InterserverThreads",
	"ClickHouseAsyncMetrics,host=me,instance=localhost:9363 Jitter",
	"ClickHouseAsyncMetrics,host=me,instance=localhost:9363 MMapCacheCells",
	"ClickHouseAsyncMetrics,host=me,instance=localhost:9363 MarkCacheBytes",
	"ClickHouseAsyncMetrics,host=me,instance=localhost:9363 MarkCacheFiles",
	"ClickHouseAsyncMetrics,host=me,instance=localhost:9363 MaxPartCountForPartition",
	"ClickHouseAsyncMetrics,host=me,instance=localhost:9363 MemoryCode",
	"ClickHouseAsyncMetrics,host=me,instance=localhost:9363 MemoryDataAndStack",
	"ClickHouseAsyncMetrics,host=me,instance=localhost:9363 MemoryResident",
	"ClickHouseAsyncMetrics,host=me,instance=localhost:9363 MemoryShared",
	"ClickHouseAsyncMetrics,host=me,instance=localhost:9363 MemoryVirtual",
	"ClickHouseAsyncMetrics,host=me,instance=localhost:9363 MySQLThreads",
	"ClickHouseAsyncMetrics,host=me,instance=localhost:9363 NumberOfDatabases",
	"ClickHouseAsyncMetrics,host=me,instance=localhost:9363 NumberOfTables",
	"ClickHouseAsyncMetrics,host=me,instance=localhost:9363 OSContextSwitches",
	"ClickHouseAsyncMetrics,host=me,instance=localhost:9363 OSGuestNiceTime",
	"ClickHouseAsyncMetrics,host=me,instance=localhost:9363 OSGuestNiceTimeNormalized",
	"ClickHouseAsyncMetrics,host=me,instance=localhost:9363 OSGuestTime",
	"ClickHouseAsyncMetrics,host=me,instance=localhost:9363 OSGuestTimeNormalized",
	"ClickHouseAsyncMetrics,host=me,instance=localhost:9363 OSIOWaitTime",
	"ClickHouseAsyncMetrics,host=me,instance=localhost:9363 OSIOWaitTimeNormalized",
	"ClickHouseAsyncMetrics,host=me,instance=localhost:9363 OSIdleTime",
	"ClickHouseAsyncMetrics,host=me,instance=localhost:9363 OSIdleTimeNormalized",
	"ClickHouseAsyncMetrics,host=me,instance=localhost:9363 OSInterrupts",
	"ClickHouseAsyncMetrics,host=me,instance=localhost:9363 OSIrqTime",
	"ClickHouseAsyncMetrics,host=me,instance=localhost:9363 OSIrqTimeNormalized",
	"ClickHouseAsyncMetrics,host=me,instance=localhost:9363 OSMemoryAvailable",
	"ClickHouseAsyncMetrics,host=me,instance=localhost:9363 OSMemoryBuffers",
	"ClickHouseAsyncMetrics,host=me,instance=localhost:9363 OSMemoryCached",
	"ClickHouseAsyncMetrics,host=me,instance=localhost:9363 OSMemoryFreePlusCached",
	"ClickHouseAsyncMetrics,host=me,instance=localhost:9363 OSMemoryFreeWithoutCached",
	"ClickHouseAsyncMetrics,host=me,instance=localhost:9363 OSMemoryTotal",
	"ClickHouseAsyncMetrics,host=me,instance=localhost:9363 OSNiceTime",
	"ClickHouseAsyncMetrics,host=me,instance=localhost:9363 OSNiceTimeNormalized",
	"ClickHouseAsyncMetrics,host=me,instance=localhost:9363 OSOpenFiles",
	"ClickHouseAsyncMetrics,host=me,instance=localhost:9363 OSProcessesBlocked",
	"ClickHouseAsyncMetrics,host=me,instance=localhost:9363 OSProcessesCreated",
	"ClickHouseAsyncMetrics,host=me,instance=localhost:9363 OSProcessesRunning",
	"ClickHouseAsyncMetrics,host=me,instance=localhost:9363 OSSoftIrqTime",
	"ClickHouseAsyncMetrics,host=me,instance=localhost:9363 OSSoftIrqTimeNormalized",
	"ClickHouseAsyncMetrics,host=me,instance=localhost:9363 OSStealTime",
	"ClickHouseAsyncMetrics,host=me,instance=localhost:9363 OSStealTimeNormalized",
	"ClickHouseAsyncMetrics,host=me,instance=localhost:9363 OSSystemTime",
	"ClickHouseAsyncMetrics,host=me,instance=localhost:9363 OSSystemTimeNormalized",
	"ClickHouseAsyncMetrics,host=me,instance=localhost:9363 OSThreadsRunnable",
	"ClickHouseAsyncMetrics,host=me,instance=localhost:9363 OSThreadsTotal",
	"ClickHouseAsyncMetrics,host=me,instance=localhost:9363 OSUptime",
	"ClickHouseAsyncMetrics,host=me,instance=localhost:9363 OSUserTime",
	"ClickHouseAsyncMetrics,host=me,instance=localhost:9363 OSUserTimeNormalized",
	"ClickHouseAsyncMetrics,host=me,instance=localhost:9363 PostgreSQLThreads",
	"ClickHouseAsyncMetrics,host=me,instance=localhost:9363 PrometheusThreads",
	"ClickHouseAsyncMetrics,host=me,instance=localhost:9363 ReplicasMaxAbsoluteDelay",
	"ClickHouseAsyncMetrics,host=me,instance=localhost:9363 ReplicasMaxInsertsInQueue",
	"ClickHouseAsyncMetrics,host=me,instance=localhost:9363 ReplicasMaxMergesInQueue",
	"ClickHouseAsyncMetrics,host=me,instance=localhost:9363 ReplicasMaxQueueSize",
	"ClickHouseAsyncMetrics,host=me,instance=localhost:9363 ReplicasMaxRelativeDelay",
	"ClickHouseAsyncMetrics,host=me,instance=localhost:9363 ReplicasSumInsertsInQueue",
	"ClickHouseAsyncMetrics,host=me,instance=localhost:9363 ReplicasSumMergesInQueue",
	"ClickHouseAsyncMetrics,host=me,instance=localhost:9363 ReplicasSumQueueSize",
	"ClickHouseAsyncMetrics,host=me,instance=localhost:9363 TCPThreads",
	"ClickHouseAsyncMetrics,host=me,instance=localhost:9363 TotalBytesOfMergeTreeTables",
	"ClickHouseAsyncMetrics,host=me,instance=localhost:9363 TotalPartsOfMergeTreeTables",
	"ClickHouseAsyncMetrics,host=me,instance=localhost:9363 TotalRowsOfMergeTreeTables",
	"ClickHouseAsyncMetrics,host=me,instance=localhost:9363 UncompressedCacheBytes",
	"ClickHouseAsyncMetrics,host=me,instance=localhost:9363 UncompressedCacheCells",
	"ClickHouseAsyncMetrics,host=me,instance=localhost:9363 Uptime",
	"ClickHouseAsyncMetrics,host=me,instance=localhost:9363 jemalloc_active",
	"ClickHouseAsyncMetrics,host=me,instance=localhost:9363 jemalloc_allocated",
	"ClickHouseAsyncMetrics,host=me,instance=localhost:9363 jemalloc_arenas_all_dirty_purged",
	"ClickHouseAsyncMetrics,host=me,instance=localhost:9363 jemalloc_arenas_all_muzzy_purged",
	"ClickHouseAsyncMetrics,host=me,instance=localhost:9363 jemalloc_arenas_all_pactive",
	"ClickHouseAsyncMetrics,host=me,instance=localhost:9363 jemalloc_arenas_all_pdirty",
	"ClickHouseAsyncMetrics,host=me,instance=localhost:9363 jemalloc_arenas_all_pmuzzy",
	"ClickHouseAsyncMetrics,host=me,instance=localhost:9363 jemalloc_background_thread_num_runs",
	"ClickHouseAsyncMetrics,host=me,instance=localhost:9363 jemalloc_background_thread_num_threads",
	"ClickHouseAsyncMetrics,host=me,instance=localhost:9363 jemalloc_background_thread_run_intervals",
	"ClickHouseAsyncMetrics,host=me,instance=localhost:9363 jemalloc_epoch",
	"ClickHouseAsyncMetrics,host=me,instance=localhost:9363 jemalloc_mapped",
	"ClickHouseAsyncMetrics,host=me,instance=localhost:9363 jemalloc_metadata",
	"ClickHouseAsyncMetrics,host=me,instance=localhost:9363 jemalloc_metadata_thp",
	"ClickHouseAsyncMetrics,host=me,instance=localhost:9363 jemalloc_resident",
	"ClickHouseAsyncMetrics,host=me,instance=localhost:9363 jemalloc_retained",
	"ClickHouseAsyncMetrics,host=me,instance=localhost:9363,unit=0 Temperature",
	"ClickHouseAsyncMetrics,host=me,instance=localhost:9363,unit=1 LoadAverage",
	"ClickHouseAsyncMetrics,host=me,instance=localhost:9363,unit=1 Temperature",
	"ClickHouseAsyncMetrics,host=me,instance=localhost:9363,unit=15 LoadAverage",
	"ClickHouseAsyncMetrics,host=me,instance=localhost:9363,unit=2 Temperature",
	"ClickHouseAsyncMetrics,host=me,instance=localhost:9363,unit=3 Temperature",
	"ClickHouseAsyncMetrics,host=me,instance=localhost:9363,unit=4 Temperature",
	"ClickHouseAsyncMetrics,host=me,instance=localhost:9363,unit=5 LoadAverage",
	"ClickHouseAsyncMetrics,host=me,instance=localhost:9363,unit=5 Temperature",
	"ClickHouseAsyncMetrics,host=me,instance=localhost:9363,unit=6 Temperature",
	"ClickHouseAsyncMetrics,host=me,instance=localhost:9363,unit=7 Temperature",
	"ClickHouseAsyncMetrics,host=me,instance=localhost:9363,unit=acpitz Temperature",
	"ClickHouseAsyncMetrics,host=me,instance=localhost:9363,unit=average BlockActiveTime",
	"ClickHouseAsyncMetrics,host=me,instance=localhost:9363,unit=average BlockDiscardTime",
	"ClickHouseAsyncMetrics,host=me,instance=localhost:9363,unit=average BlockQueueTime",
	"ClickHouseAsyncMetrics,host=me,instance=localhost:9363,unit=average BlockReadTime",
	"ClickHouseAsyncMetrics,host=me,instance=localhost:9363,unit=average BlockWriteTime",
	"ClickHouseAsyncMetrics,host=me,instance=localhost:9363,unit=average LoadAverage",
	"ClickHouseAsyncMetrics,host=me,instance=localhost:9363,unit=average Temperature",
	"ClickHouseAsyncMetrics,host=me,instance=localhost:9363,unit=coretemp_Core_0 Temperature",
	"ClickHouseAsyncMetrics,host=me,instance=localhost:9363,unit=coretemp_Core_1 Temperature",
	"ClickHouseAsyncMetrics,host=me,instance=localhost:9363,unit=coretemp_Core_2 Temperature",
	"ClickHouseAsyncMetrics,host=me,instance=localhost:9363,unit=coretemp_Core_3 Temperature",
	"ClickHouseAsyncMetrics,host=me,instance=localhost:9363,unit=coretemp_Package_id_0 Temperature",
	"ClickHouseAsyncMetrics,host=me,instance=localhost:9363,unit=iwlwifi_1 Temperature",
	"ClickHouseAsyncMetrics,host=me,instance=localhost:9363,unit=nvme0n1 BlockActiveTime",
	"ClickHouseAsyncMetrics,host=me,instance=localhost:9363,unit=nvme0n1 BlockDiscardBytes",
	"ClickHouseAsyncMetrics,host=me,instance=localhost:9363,unit=nvme0n1 BlockDiscardMerges",
	"ClickHouseAsyncMetrics,host=me,instance=localhost:9363,unit=nvme0n1 BlockDiscardOps",
	"ClickHouseAsyncMetrics,host=me,instance=localhost:9363,unit=nvme0n1 BlockDiscardTime",
	"ClickHouseAsyncMetrics,host=me,instance=localhost:9363,unit=nvme0n1 BlockInFlightOps",
	"ClickHouseAsyncMetrics,host=me,instance=localhost:9363,unit=nvme0n1 BlockQueueTime",
	"ClickHouseAsyncMetrics,host=me,instance=localhost:9363,unit=nvme0n1 BlockReadBytes",
	"ClickHouseAsyncMetrics,host=me,instance=localhost:9363,unit=nvme0n1 BlockReadMerges",
	"ClickHouseAsyncMetrics,host=me,instance=localhost:9363,unit=nvme0n1 BlockReadOps",
	"ClickHouseAsyncMetrics,host=me,instance=localhost:9363,unit=nvme0n1 BlockReadTime",
	"ClickHouseAsyncMetrics,host=me,instance=localhost:9363,unit=nvme0n1 BlockWriteBytes",
	"ClickHouseAsyncMetrics,host=me,instance=localhost:9363,unit=nvme0n1 BlockWriteMerges",
	"ClickHouseAsyncMetrics,host=me,instance=localhost:9363,unit=nvme0n1 BlockWriteOps",
	"ClickHouseAsyncMetrics,host=me,instance=localhost:9363,unit=nvme0n1 BlockWriteTime",
	"ClickHouseAsyncMetrics,host=me,instance=localhost:9363,unit=nvme_Composite Temperature",
	"ClickHouseAsyncMetrics,host=me,instance=localhost:9363,unit=nvme_Sensor_1 Temperature",
	"ClickHouseAsyncMetrics,host=me,instance=localhost:9363,unit=nvme_Sensor_2 Temperature",
	"ClickHouseAsyncMetrics,host=me,instance=localhost:9363,unit=pch_cannonlake Temperature",
	"ClickHouseAsyncMetrics,host=me,instance=localhost:9363,unit=sda BlockActiveTime",
	"ClickHouseAsyncMetrics,host=me,instance=localhost:9363,unit=sda BlockDiscardBytes",
	"ClickHouseAsyncMetrics,host=me,instance=localhost:9363,unit=sda BlockDiscardMerges",
	"ClickHouseAsyncMetrics,host=me,instance=localhost:9363,unit=sda BlockDiscardOps",
	"ClickHouseAsyncMetrics,host=me,instance=localhost:9363,unit=sda BlockDiscardTime",
	"ClickHouseAsyncMetrics,host=me,instance=localhost:9363,unit=sda BlockInFlightOps",
	"ClickHouseAsyncMetrics,host=me,instance=localhost:9363,unit=sda BlockQueueTime",
	"ClickHouseAsyncMetrics,host=me,instance=localhost:9363,unit=sda BlockReadBytes",
	"ClickHouseAsyncMetrics,host=me,instance=localhost:9363,unit=sda BlockReadMerges",
	"ClickHouseAsyncMetrics,host=me,instance=localhost:9363,unit=sda BlockReadOps",
	"ClickHouseAsyncMetrics,host=me,instance=localhost:9363,unit=sda BlockReadTime",
	"ClickHouseAsyncMetrics,host=me,instance=localhost:9363,unit=sda BlockWriteBytes",
	"ClickHouseAsyncMetrics,host=me,instance=localhost:9363,unit=sda BlockWriteMerges",
	"ClickHouseAsyncMetrics,host=me,instance=localhost:9363,unit=sda BlockWriteOps",
	"ClickHouseAsyncMetrics,host=me,instance=localhost:9363,unit=sda BlockWriteTime",
	"ClickHouseAsyncMetrics,host=me,instance=localhost:9363,unit=total BlockDiscardBytes",
	"ClickHouseAsyncMetrics,host=me,instance=localhost:9363,unit=total BlockDiscardMerges",
	"ClickHouseAsyncMetrics,host=me,instance=localhost:9363,unit=total BlockDiscardOps",
	"ClickHouseAsyncMetrics,host=me,instance=localhost:9363,unit=total BlockInFlightOps",
	"ClickHouseAsyncMetrics,host=me,instance=localhost:9363,unit=total BlockReadBytes",
	"ClickHouseAsyncMetrics,host=me,instance=localhost:9363,unit=total BlockReadMerges",
	"ClickHouseAsyncMetrics,host=me,instance=localhost:9363,unit=total BlockReadOps",
	"ClickHouseAsyncMetrics,host=me,instance=localhost:9363,unit=total BlockWriteBytes",
	"ClickHouseAsyncMetrics,host=me,instance=localhost:9363,unit=total BlockWriteMerges",
	"ClickHouseAsyncMetrics,host=me,instance=localhost:9363,unit=total BlockWriteOps",
	"ClickHouseMetrics,host=me,instance=localhost:9363 ActiveAsyncDrainedConnections",
	"ClickHouseMetrics,host=me,instance=localhost:9363 ActiveSyncDrainedConnections",
	"ClickHouseMetrics,host=me,instance=localhost:9363 AsyncDrainedConnections",
	"ClickHouseMetrics,host=me,instance=localhost:9363 AsynchronousReadWait",
	"ClickHouseMetrics,host=me,instance=localhost:9363 BackgroundBufferFlushSchedulePoolTask",
	"ClickHouseMetrics,host=me,instance=localhost:9363 BackgroundCommonPoolTask",
	"ClickHouseMetrics,host=me,instance=localhost:9363 BackgroundDistributedSchedulePoolTask",
	"ClickHouseMetrics,host=me,instance=localhost:9363 BackgroundFetchesPoolTask",
	"ClickHouseMetrics,host=me,instance=localhost:9363 BackgroundMergesAndMutationsPoolTask",
	"ClickHouseMetrics,host=me,instance=localhost:9363 BackgroundMessageBrokerSchedulePoolTask",
	"ClickHouseMetrics,host=me,instance=localhost:9363 BackgroundMovePoolTask",
	"ClickHouseMetrics,host=me,instance=localhost:9363 BackgroundSchedulePoolTask",
	"ClickHouseMetrics,host=me,instance=localhost:9363 BrokenDistributedFilesToInsert",
	"ClickHouseMetrics,host=me,instance=localhost:9363 CacheDetachedFileSegments",
	"ClickHouseMetrics,host=me,instance=localhost:9363 CacheDictionaryUpdateQueueBatches",
	"ClickHouseMetrics,host=me,instance=localhost:9363 CacheDictionaryUpdateQueueKeys",
	"ClickHouseMetrics,host=me,instance=localhost:9363 CacheFileSegments",
	"ClickHouseMetrics,host=me,instance=localhost:9363 ContextLockWait",
	"ClickHouseMetrics,host=me,instance=localhost:9363 DelayedInserts",
	"ClickHouseMetrics,host=me,instance=localhost:9363 DictCacheRequests",
	"ClickHouseMetrics,host=me,instance=localhost:9363 DiskSpaceReservedForMerge",
	"ClickHouseMetrics,host=me,instance=localhost:9363 DistributedFilesToInsert",
	"ClickHouseMetrics,host=me,instance=localhost:9363 DistributedSend",
	"ClickHouseMetrics,host=me,instance=localhost:9363 EphemeralNode",
	"ClickHouseMetrics,host=me,instance=localhost:9363 FilesystemCacheElements",
	"ClickHouseMetrics,host=me,instance=localhost:9363 FilesystemCacheReadBuffers",
	"ClickHouseMetrics,host=me,instance=localhost:9363 FilesystemCacheSize",
	"ClickHouseMetrics,host=me,instance=localhost:9363 GlobalThread",
	"ClickHouseMetrics,host=me,instance=localhost:9363 GlobalThreadActive",
	"ClickHouseMetrics,host=me,instance=localhost:9363 HTTPConnection",
	"ClickHouseMetrics,host=me,instance=localhost:9363 InterserverConnection",
	"ClickHouseMetrics,host=me,instance=localhost:9363 KafkaAssignedPartitions",
	"ClickHouseMetrics,host=me,instance=localhost:9363 KafkaBackgroundReads",
	"ClickHouseMetrics,host=me,instance=localhost:9363 KafkaConsumers",
	"ClickHouseMetrics,host=me,instance=localhost:9363 KafkaConsumersInUse",
	"ClickHouseMetrics,host=me,instance=localhost:9363 KafkaConsumersWithAssignment",
	"ClickHouseMetrics,host=me,instance=localhost:9363 KafkaLibrdkafkaThreads",
	"ClickHouseMetrics,host=me,instance=localhost:9363 KafkaProducers",
	"ClickHouseMetrics,host=me,instance=localhost:9363 KafkaWrites",
	"ClickHouseMetrics,host=me,instance=localhost:9363 KeeperAliveConnections",
	"ClickHouseMetrics,host=me,instance=localhost:9363 KeeperOutstandingRequets",
	"ClickHouseMetrics,host=me,instance=localhost:9363 LocalThread",
	"ClickHouseMetrics,host=me,instance=localhost:9363 LocalThreadActive",
	"ClickHouseMetrics,host=me,instance=localhost:9363 MMappedFileBytes",
	"ClickHouseMetrics,host=me,instance=localhost:9363 MMappedFiles",
	"ClickHouseMetrics,host=me,instance=localhost:9363 MaxDDLEntryID",
	"ClickHouseMetrics,host=me,instance=localhost:9363 MaxPushedDDLEntryID",
	"ClickHouseMetrics,host=me,instance=localhost:9363 MemoryTracking",
	"ClickHouseMetrics,host=me,instance=localhost:9363 Merge",
	"ClickHouseMetrics,host=me,instance=localhost:9363 MySQLConnection",
	"ClickHouseMetrics,host=me,instance=localhost:9363 NetworkReceive",
	"ClickHouseMetrics,host=me,instance=localhost:9363 NetworkSend",
	"ClickHouseMetrics,host=me,instance=localhost:9363 OpenFileForRead",
	"ClickHouseMetrics,host=me,instance=localhost:9363 OpenFileForWrite",
	"ClickHouseMetrics,host=me,instance=localhost:9363 PartMutation",
	"ClickHouseMetrics,host=me,instance=localhost:9363 PartsActive",
	"ClickHouseMetrics,host=me,instance=localhost:9363 PartsCommitted",
	"ClickHouseMetrics,host=me,instance=localhost:9363 PartsCompact",
	"ClickHouseMetrics,host=me,instance=localhost:9363 PartsDeleteOnDestroy",
	"ClickHouseMetrics,host=me,instance=localhost:9363 PartsDeleting",
	"ClickHouseMetrics,host=me,instance=localhost:9363 PartsInMemory",
	"ClickHouseMetrics,host=me,instance=localhost:9363 PartsOutdated",
	"ClickHouseMetrics,host=me,instance=localhost:9363 PartsPreActive",
	"ClickHouseMetrics,host=me,instance=localhost:9363 PartsPreCommitted",
	"ClickHouseMetrics,host=me,instance=localhost:9363 PartsTemporary",
	"ClickHouseMetrics,host=me,instance=localhost:9363 PartsWide",
	"ClickHouseMetrics,host=me,instance=localhost:9363 PendingAsyncInsert",
	"ClickHouseMetrics,host=me,instance=localhost:9363 PostgreSQLConnection",
	"ClickHouseMetrics,host=me,instance=localhost:9363 Query",
	"ClickHouseMetrics,host=me,instance=localhost:9363 QueryPreempted",
	"ClickHouseMetrics,host=me,instance=localhost:9363 QueryThread",
	"ClickHouseMetrics,host=me,instance=localhost:9363 RWLockActiveReaders",
	"ClickHouseMetrics,host=me,instance=localhost:9363 RWLockActiveWriters",
	"ClickHouseMetrics,host=me,instance=localhost:9363 RWLockWaitingReaders",
	"ClickHouseMetrics,host=me,instance=localhost:9363 RWLockWaitingWriters",
	"ClickHouseMetrics,host=me,instance=localhost:9363 Read",
	"ClickHouseMetrics,host=me,instance=localhost:9363 ReadonlyReplica",
	"ClickHouseMetrics,host=me,instance=localhost:9363 ReplicatedChecks",
	"ClickHouseMetrics,host=me,instance=localhost:9363 ReplicatedFetch",
	"ClickHouseMetrics,host=me,instance=localhost:9363 ReplicatedSend",
	"ClickHouseMetrics,host=me,instance=localhost:9363 Revision",
	"ClickHouseMetrics,host=me,instance=localhost:9363 S3Requests",
	"ClickHouseMetrics,host=me,instance=localhost:9363 SendExternalTables",
	"ClickHouseMetrics,host=me,instance=localhost:9363 SendScalars",
	"ClickHouseMetrics,host=me,instance=localhost:9363 StorageBufferBytes",
	"ClickHouseMetrics,host=me,instance=localhost:9363 StorageBufferRows",
	"ClickHouseMetrics,host=me,instance=localhost:9363 SyncDrainedConnections",
	"ClickHouseMetrics,host=me,instance=localhost:9363 TCPConnection",
	"ClickHouseMetrics,host=me,instance=localhost:9363 TablesToDropQueueSize",
	"ClickHouseMetrics,host=me,instance=localhost:9363 VersionInteger",
	"ClickHouseMetrics,host=me,instance=localhost:9363 Write",
	"ClickHouseMetrics,host=me,instance=localhost:9363 ZooKeeperRequest",
	"ClickHouseMetrics,host=me,instance=localhost:9363 ZooKeeperSession",
	"ClickHouseMetrics,host=me,instance=localhost:9363 ZooKeeperWatch",
	"ClickHouseProfileEvents,host=me,instance=localhost:9363 AIORead",
	"ClickHouseProfileEvents,host=me,instance=localhost:9363 AIOReadBytes",
	"ClickHouseProfileEvents,host=me,instance=localhost:9363 AIOWrite",
	"ClickHouseProfileEvents,host=me,instance=localhost:9363 AIOWriteBytes",
	"ClickHouseProfileEvents,host=me,instance=localhost:9363 AggregationHashTablesInitializedAsTwoLevel",
	"ClickHouseProfileEvents,host=me,instance=localhost:9363 AggregationPreallocatedElementsInHashTables",
	"ClickHouseProfileEvents,host=me,instance=localhost:9363 ArenaAllocBytes",
	"ClickHouseProfileEvents,host=me,instance=localhost:9363 ArenaAllocChunks",
	"ClickHouseProfileEvents,host=me,instance=localhost:9363 AsyncInsertBytes",
	"ClickHouseProfileEvents,host=me,instance=localhost:9363 AsyncInsertQuery",
	"ClickHouseProfileEvents,host=me,instance=localhost:9363 AsynchronousReadWaitMicroseconds",
	"ClickHouseProfileEvents,host=me,instance=localhost:9363 CachedReadBufferCacheWriteBytes",
	"ClickHouseProfileEvents,host=me,instance=localhost:9363 CachedReadBufferCacheWriteMicroseconds",
	"ClickHouseProfileEvents,host=me,instance=localhost:9363 CachedReadBufferReadFromCacheBytes",
	"ClickHouseProfileEvents,host=me,instance=localhost:9363 CachedReadBufferReadFromCacheMicroseconds",
	"ClickHouseProfileEvents,host=me,instance=localhost:9363 CachedReadBufferReadFromSourceBytes",
	"ClickHouseProfileEvents,host=me,instance=localhost:9363 CachedReadBufferReadFromSourceMicroseconds",
	"ClickHouseProfileEvents,host=me,instance=localhost:9363 CachedWriteBufferCacheWriteBytes",
	"ClickHouseProfileEvents,host=me,instance=localhost:9363 CachedWriteBufferCacheWriteMicroseconds",
	"ClickHouseProfileEvents,host=me,instance=localhost:9363 CannotRemoveEphemeralNode",
	"ClickHouseProfileEvents,host=me,instance=localhost:9363 CannotWriteToWriteBufferDiscard",
	"ClickHouseProfileEvents,host=me,instance=localhost:9363 CompileExpressionsBytes",
	"ClickHouseProfileEvents,host=me,instance=localhost:9363 CompileExpressionsMicroseconds",
	"ClickHouseProfileEvents,host=me,instance=localhost:9363 CompileFunction",
	"ClickHouseProfileEvents,host=me,instance=localhost:9363 CompiledFunctionExecute",
	"ClickHouseProfileEvents,host=me,instance=localhost:9363 CompressedReadBufferBlocks",
	"ClickHouseProfileEvents,host=me,instance=localhost:9363 CompressedReadBufferBytes",
	"ClickHouseProfileEvents,host=me,instance=localhost:9363 ContextLock",
	"ClickHouseProfileEvents,host=me,instance=localhost:9363 CreatedHTTPConnections",
	"ClickHouseProfileEvents,host=me,instance=localhost:9363 CreatedLogEntryForMerge",
	"ClickHouseProfileEvents,host=me,instance=localhost:9363 CreatedLogEntryForMutation",
	"ClickHouseProfileEvents,host=me,instance=localhost:9363 CreatedReadBufferDirectIO",
	"ClickHouseProfileEvents,host=me,instance=localhost:9363 CreatedReadBufferDirectIOFailed",
	"ClickHouseProfileEvents,host=me,instance=localhost:9363 CreatedReadBufferMMap",
	"ClickHouseProfileEvents,host=me,instance=localhost:9363 CreatedReadBufferMMapFailed",
	"ClickHouseProfileEvents,host=me,instance=localhost:9363 CreatedReadBufferOrdinary",
	"ClickHouseProfileEvents,host=me,instance=localhost:9363 DNSError",
	"ClickHouseProfileEvents,host=me,instance=localhost:9363 DataAfterMergeDiffersFromReplica",
	"ClickHouseProfileEvents,host=me,instance=localhost:9363 DataAfterMutationDiffersFromReplica",
	"ClickHouseProfileEvents,host=me,instance=localhost:9363 DelayedInserts",
	"ClickHouseProfileEvents,host=me,instance=localhost:9363 DelayedInsertsMilliseconds",
	"ClickHouseProfileEvents,host=me,instance=localhost:9363 DictCacheKeysExpired",
	"ClickHouseProfileEvents,host=me,instance=localhost:9363 DictCacheKeysHit",
	"ClickHouseProfileEvents,host=me,instance=localhost:9363 DictCacheKeysNotFound",
	"ClickHouseProfileEvents,host=me,instance=localhost:9363 DictCacheKeysRequested",
	"ClickHouseProfileEvents,host=me,instance=localhost:9363 DictCacheKeysRequestedFound",
	"ClickHouseProfileEvents,host=me,instance=localhost:9363 DictCacheKeysRequestedMiss",
	"ClickHouseProfileEvents,host=me,instance=localhost:9363 DictCacheLockReadNs",
	"ClickHouseProfileEvents,host=me,instance=localhost:9363 DictCacheLockWriteNs",
	"ClickHouseProfileEvents,host=me,instance=localhost:9363 DictCacheRequestTimeNs",
	"ClickHouseProfileEvents,host=me,instance=localhost:9363 DictCacheRequests",
	"ClickHouseProfileEvents,host=me,instance=localhost:9363 DirectorySync",
	"ClickHouseProfileEvents,host=me,instance=localhost:9363 DirectorySyncElapsedMicroseconds",
	"ClickHouseProfileEvents,host=me,instance=localhost:9363 DiskReadElapsedMicroseconds",
	"ClickHouseProfileEvents,host=me,instance=localhost:9363 DiskWriteElapsedMicroseconds",
	"ClickHouseProfileEvents,host=me,instance=localhost:9363 DistributedConnectionFailAtAll",
	"ClickHouseProfileEvents,host=me,instance=localhost:9363 DistributedConnectionFailTry",
	"ClickHouseProfileEvents,host=me,instance=localhost:9363 DistributedConnectionMissingTable",
	"ClickHouseProfileEvents,host=me,instance=localhost:9363 DistributedConnectionStaleReplica",
	"ClickHouseProfileEvents,host=me,instance=localhost:9363 DistributedDelayedInserts",
	"ClickHouseProfileEvents,host=me,instance=localhost:9363 DistributedDelayedInsertsMilliseconds",
	"ClickHouseProfileEvents,host=me,instance=localhost:9363 DistributedRejectedInserts",
	"ClickHouseProfileEvents,host=me,instance=localhost:9363 DistributedSyncInsertionTimeoutExceeded",
	"ClickHouseProfileEvents,host=me,instance=localhost:9363 DuplicatedInsertedBlocks",
	"ClickHouseProfileEvents,host=me,instance=localhost:9363 ExecuteShellCommand",
	"ClickHouseProfileEvents,host=me,instance=localhost:9363 ExternalAggregationCompressedBytes",
	"ClickHouseProfileEvents,host=me,instance=localhost:9363 ExternalAggregationMerge",
	"ClickHouseProfileEvents,host=me,instance=localhost:9363 ExternalAggregationUncompressedBytes",
	"ClickHouseProfileEvents,host=me,instance=localhost:9363 ExternalAggregationWritePart",
	"ClickHouseProfileEvents,host=me,instance=localhost:9363 ExternalDataSourceLocalCacheReadBytes",
	"ClickHouseProfileEvents,host=me,instance=localhost:9363 ExternalSortMerge",
	"ClickHouseProfileEvents,host=me,instance=localhost:9363 ExternalSortWritePart",
	"ClickHouseProfileEvents,host=me,instance=localhost:9363 FailedInsertQuery",
	"ClickHouseProfileEvents,host=me,instance=localhost:9363 FailedQuery",
	"ClickHouseProfileEvents,host=me,instance=localhost:9363 FailedSelectQuery",
	"ClickHouseProfileEvents,host=me,instance=localhost:9363 FileOpen",
	"ClickHouseProfileEvents,host=me,instance=localhost:9363 FileSegmentCacheWriteMicroseconds",
	"ClickHouseProfileEvents,host=me,instance=localhost:9363 FileSegmentPredownloadMicroseconds",
	"ClickHouseProfileEvents,host=me,instance=localhost:9363 FileSegmentReadMicroseconds",
	"ClickHouseProfileEvents,host=me,instance=localhost:9363 FileSegmentUsedBytes",
	"ClickHouseProfileEvents,host=me,instance=localhost:9363 FileSegmentWaitReadBufferMicroseconds",
	"ClickHouseProfileEvents,host=me,instance=localhost:9363 FileSync",
	"ClickHouseProfileEvents,host=me,instance=localhost:9363 FileSyncElapsedMicroseconds",
	"ClickHouseProfileEvents,host=me,instance=localhost:9363 FunctionExecute",
	"ClickHouseProfileEvents,host=me,instance=localhost:9363 HardPageFaults",
	"ClickHouseProfileEvents,host=me,instance=localhost:9363 HedgedRequestsChangeReplica",
	"ClickHouseProfileEvents,host=me,instance=localhost:9363 IOBufferAllocBytes",
	"ClickHouseProfileEvents,host=me,instance=localhost:9363 IOBufferAllocs",
	"ClickHouseProfileEvents,host=me,instance=localhost:9363 InsertQuery",
	"ClickHouseProfileEvents,host=me,instance=localhost:9363 InsertQueryTimeMicroseconds",
	"ClickHouseProfileEvents,host=me,instance=localhost:9363 InsertedBytes",
	"ClickHouseProfileEvents,host=me,instance=localhost:9363 InsertedCompactParts",
	"ClickHouseProfileEvents,host=me,instance=localhost:9363 InsertedInMemoryParts",
	"ClickHouseProfileEvents,host=me,instance=localhost:9363 InsertedRows",
	"ClickHouseProfileEvents,host=me,instance=localhost:9363 InsertedWideParts",
	"ClickHouseProfileEvents,host=me,instance=localhost:9363 KafkaBackgroundReads",
	"ClickHouseProfileEvents,host=me,instance=localhost:9363 KafkaCommitFailures",
	"ClickHouseProfileEvents,host=me,instance=localhost:9363 KafkaCommits",
	"ClickHouseProfileEvents,host=me,instance=localhost:9363 KafkaConsumerErrors",
	"ClickHouseProfileEvents,host=me,instance=localhost:9363 KafkaDirectReads",
	"ClickHouseProfileEvents,host=me,instance=localhost:9363 KafkaMessagesFailed",
	"ClickHouseProfileEvents,host=me,instance=localhost:9363 KafkaMessagesPolled",
	"ClickHouseProfileEvents,host=me,instance=localhost:9363 KafkaMessagesProduced",
	"ClickHouseProfileEvents,host=me,instance=localhost:9363 KafkaMessagesRead",
	"ClickHouseProfileEvents,host=me,instance=localhost:9363 KafkaProducerErrors",
	"ClickHouseProfileEvents,host=me,instance=localhost:9363 KafkaProducerFlushes",
	"ClickHouseProfileEvents,host=me,instance=localhost:9363 KafkaRebalanceAssignments",
	"ClickHouseProfileEvents,host=me,instance=localhost:9363 KafkaRebalanceErrors",
	"ClickHouseProfileEvents,host=me,instance=localhost:9363 KafkaRebalanceRevocations",
	"ClickHouseProfileEvents,host=me,instance=localhost:9363 KafkaRowsRead",
	"ClickHouseProfileEvents,host=me,instance=localhost:9363 KafkaRowsRejected",
	"ClickHouseProfileEvents,host=me,instance=localhost:9363 KafkaRowsWritten",
	"ClickHouseProfileEvents,host=me,instance=localhost:9363 KafkaWrites",
	"ClickHouseProfileEvents,host=me,instance=localhost:9363 KeeperCommits",
	"ClickHouseProfileEvents,host=me,instance=localhost:9363 KeeperCommitsFailed",
	"ClickHouseProfileEvents,host=me,instance=localhost:9363 KeeperLatency",
	"ClickHouseProfileEvents,host=me,instance=localhost:9363 KeeperPacketsReceived",
	"ClickHouseProfileEvents,host=me,instance=localhost:9363 KeeperPacketsSent",
	"ClickHouseProfileEvents,host=me,instance=localhost:9363 KeeperReadSnapshot",
	"ClickHouseProfileEvents,host=me,instance=localhost:9363 KeeperRequestTotal",
	"ClickHouseProfileEvents,host=me,instance=localhost:9363 KeeperSaveSnapshot",
	"ClickHouseProfileEvents,host=me,instance=localhost:9363 KeeperSnapshotApplys",
	"ClickHouseProfileEvents,host=me,instance=localhost:9363 KeeperSnapshotApplysFailed",
	"ClickHouseProfileEvents,host=me,instance=localhost:9363 KeeperSnapshotCreations",
	"ClickHouseProfileEvents,host=me,instance=localhost:9363 KeeperSnapshotCreationsFailed",
	"ClickHouseProfileEvents,host=me,instance=localhost:9363 MMappedFileCacheHits",
	"ClickHouseProfileEvents,host=me,instance=localhost:9363 MMappedFileCacheMisses",
	"ClickHouseProfileEvents,host=me,instance=localhost:9363 MainConfigLoads",
	"ClickHouseProfileEvents,host=me,instance=localhost:9363 MarkCacheHits",
	"ClickHouseProfileEvents,host=me,instance=localhost:9363 MarkCacheMisses",
	"ClickHouseProfileEvents,host=me,instance=localhost:9363 MemoryOvercommitWaitTimeMicroseconds",
	"ClickHouseProfileEvents,host=me,instance=localhost:9363 Merge",
	"ClickHouseProfileEvents,host=me,instance=localhost:9363 MergeTreeDataProjectionWriterBlocks",
	"ClickHouseProfileEvents,host=me,instance=localhost:9363 MergeTreeDataProjectionWriterBlocksAlreadySorted",
	"ClickHouseProfileEvents,host=me,instance=localhost:9363 MergeTreeDataProjectionWriterCompressedBytes",
	"ClickHouseProfileEvents,host=me,instance=localhost:9363 MergeTreeDataProjectionWriterRows",
	"ClickHouseProfileEvents,host=me,instance=localhost:9363 MergeTreeDataProjectionWriterUncompressedBytes",
	"ClickHouseProfileEvents,host=me,instance=localhost:9363 MergeTreeDataWriterBlocks",
	"ClickHouseProfileEvents,host=me,instance=localhost:9363 MergeTreeDataWriterBlocksAlreadySorted",
	"ClickHouseProfileEvents,host=me,instance=localhost:9363 MergeTreeDataWriterCompressedBytes",
	"ClickHouseProfileEvents,host=me,instance=localhost:9363 MergeTreeDataWriterRows",
	"ClickHouseProfileEvents,host=me,instance=localhost:9363 MergeTreeDataWriterUncompressedBytes",
	"ClickHouseProfileEvents,host=me,instance=localhost:9363 MergeTreeMetadataCacheDelete",
	"ClickHouseProfileEvents,host=me,instance=localhost:9363 MergeTreeMetadataCacheGet",
	"ClickHouseProfileEvents,host=me,instance=localhost:9363 MergeTreeMetadataCacheHit",
	"ClickHouseProfileEvents,host=me,instance=localhost:9363 MergeTreeMetadataCacheMiss",
	"ClickHouseProfileEvents,host=me,instance=localhost:9363 MergeTreeMetadataCachePut",
	"ClickHouseProfileEvents,host=me,instance=localhost:9363 MergeTreeMetadataCacheSeek",
	"ClickHouseProfileEvents,host=me,instance=localhost:9363 MergedIntoCompactParts",
	"ClickHouseProfileEvents,host=me,instance=localhost:9363 MergedIntoInMemoryParts",
	"ClickHouseProfileEvents,host=me,instance=localhost:9363 MergedIntoWideParts",
	"ClickHouseProfileEvents,host=me,instance=localhost:9363 MergedRows",
	"ClickHouseProfileEvents,host=me,instance=localhost:9363 MergedUncompressedBytes",
	"ClickHouseProfileEvents,host=me,instance=localhost:9363 MergesTimeMilliseconds",
	"ClickHouseProfileEvents,host=me,instance=localhost:9363 NetworkReceiveBytes",
	"ClickHouseProfileEvents,host=me,instance=localhost:9363 NetworkReceiveElapsedMicroseconds",
	"ClickHouseProfileEvents,host=me,instance=localhost:9363 NetworkSendBytes",
	"ClickHouseProfileEvents,host=me,instance=localhost:9363 NetworkSendElapsedMicroseconds",
	"ClickHouseProfileEvents,host=me,instance=localhost:9363 NotCreatedLogEntryForMerge",
	"ClickHouseProfileEvents,host=me,instance=localhost:9363 NotCreatedLogEntryForMutation",
	"ClickHouseProfileEvents,host=me,instance=localhost:9363 OSCPUVirtualTimeMicroseconds",
	"ClickHouseProfileEvents,host=me,instance=localhost:9363 OSCPUWaitMicroseconds",
	"ClickHouseProfileEvents,host=me,instance=localhost:9363 OSIOWaitMicroseconds",
	"ClickHouseProfileEvents,host=me,instance=localhost:9363 OSReadBytes",
	"ClickHouseProfileEvents,host=me,instance=localhost:9363 OSReadChars",
	"ClickHouseProfileEvents,host=me,instance=localhost:9363 OSWriteBytes",
	"ClickHouseProfileEvents,host=me,instance=localhost:9363 OSWriteChars",
	"ClickHouseProfileEvents,host=me,instance=localhost:9363 ObsoleteReplicatedParts",
	"ClickHouseProfileEvents,host=me,instance=localhost:9363 OpenedFileCacheHits",
	"ClickHouseProfileEvents,host=me,instance=localhost:9363 OpenedFileCacheMisses",
	"ClickHouseProfileEvents,host=me,instance=localhost:9363 OtherQueryTimeMicroseconds",
	"ClickHouseProfileEvents,host=me,instance=localhost:9363 OverflowAny",
	"ClickHouseProfileEvents,host=me,instance=localhost:9363 OverflowBreak",
	"ClickHouseProfileEvents,host=me,instance=localhost:9363 OverflowThrow",
	"ClickHouseProfileEvents,host=me,instance=localhost:9363 PerfAlignmentFaults",
	"ClickHouseProfileEvents,host=me,instance=localhost:9363 PerfBranchInstructions",
	"ClickHouseProfileEvents,host=me,instance=localhost:9363 PerfBranchMisses",
	"ClickHouseProfileEvents,host=me,instance=localhost:9363 PerfBusCycles",
	"ClickHouseProfileEvents,host=me,instance=localhost:9363 PerfCacheMisses",
	"ClickHouseProfileEvents,host=me,instance=localhost:9363 PerfCacheReferences",
	"ClickHouseProfileEvents,host=me,instance=localhost:9363 PerfContextSwitches",
	"ClickHouseProfileEvents,host=me,instance=localhost:9363 PerfCpuClock",
	"ClickHouseProfileEvents,host=me,instance=localhost:9363 PerfCpuCycles",
	"ClickHouseProfileEvents,host=me,instance=localhost:9363 PerfCpuMigrations",
	"ClickHouseProfileEvents,host=me,instance=localhost:9363 PerfDataTLBMisses",
	"ClickHouseProfileEvents,host=me,instance=localhost:9363 PerfDataTLBReferences",
	"ClickHouseProfileEvents,host=me,instance=localhost:9363 PerfEmulationFaults",
	"ClickHouseProfileEvents,host=me,instance=localhost:9363 PerfInstructionTLBMisses",
	"ClickHouseProfileEvents,host=me,instance=localhost:9363 PerfInstructionTLBReferences",
	"ClickHouseProfileEvents,host=me,instance=localhost:9363 PerfInstructions",
	"ClickHouseProfileEvents,host=me,instance=localhost:9363 PerfLocalMemoryMisses",
	"ClickHouseProfileEvents,host=me,instance=localhost:9363 PerfLocalMemoryReferences",
	"ClickHouseProfileEvents,host=me,instance=localhost:9363 PerfMinEnabledRunningTime",
	"ClickHouseProfileEvents,host=me,instance=localhost:9363 PerfMinEnabledTime",
	"ClickHouseProfileEvents,host=me,instance=localhost:9363 PerfRefCpuCycles",
	"ClickHouseProfileEvents,host=me,instance=localhost:9363 PerfStalledCyclesBackend",
	"ClickHouseProfileEvents,host=me,instance=localhost:9363 PerfStalledCyclesFrontend",
	"ClickHouseProfileEvents,host=me,instance=localhost:9363 PerfTaskClock",
	"ClickHouseProfileEvents,host=me,instance=localhost:9363 PolygonsAddedToPool",
	"ClickHouseProfileEvents,host=me,instance=localhost:9363 PolygonsInPoolAllocatedBytes",
	"ClickHouseProfileEvents,host=me,instance=localhost:9363 Query",
	"ClickHouseProfileEvents,host=me,instance=localhost:9363 QueryMaskingRulesMatch",
	"ClickHouseProfileEvents,host=me,instance=localhost:9363 QueryMemoryLimitExceeded",
	"ClickHouseProfileEvents,host=me,instance=localhost:9363 QueryProfilerRuns",
	"ClickHouseProfileEvents,host=me,instance=localhost:9363 QueryProfilerSignalOverruns",
	"ClickHouseProfileEvents,host=me,instance=localhost:9363 QueryTimeMicroseconds",
	"ClickHouseProfileEvents,host=me,instance=localhost:9363 RWLockAcquiredReadLocks",
	"ClickHouseProfileEvents,host=me,instance=localhost:9363 RWLockAcquiredWriteLocks",
	"ClickHouseProfileEvents,host=me,instance=localhost:9363 RWLockReadersWaitMilliseconds",
	"ClickHouseProfileEvents,host=me,instance=localhost:9363 RWLockWritersWaitMilliseconds",
	"ClickHouseProfileEvents,host=me,instance=localhost:9363 ReadBackoff",
	"ClickHouseProfileEvents,host=me,instance=localhost:9363 ReadBufferFromFileDescriptorRead",
	"ClickHouseProfileEvents,host=me,instance=localhost:9363 ReadBufferFromFileDescriptorReadBytes",
	"ClickHouseProfileEvents,host=me,instance=localhost:9363 ReadBufferFromFileDescriptorReadFailed",
	"ClickHouseProfileEvents,host=me,instance=localhost:9363 ReadBufferFromS3Bytes",
	"ClickHouseProfileEvents,host=me,instance=localhost:9363 ReadBufferFromS3Microseconds",
	"ClickHouseProfileEvents,host=me,instance=localhost:9363 ReadBufferFromS3RequestsErrors",
	"ClickHouseProfileEvents,host=me,instance=localhost:9363 ReadBufferSeekCancelConnection",
	"ClickHouseProfileEvents,host=me,instance=localhost:9363 ReadCompressedBytes",
	"ClickHouseProfileEvents,host=me,instance=localhost:9363 RealTimeMicroseconds",
	"ClickHouseProfileEvents,host=me,instance=localhost:9363 RegexpCreated",
	"ClickHouseProfileEvents,host=me,instance=localhost:9363 RejectedInserts",
	"ClickHouseProfileEvents,host=me,instance=localhost:9363 RemoteFSBuffers",
	"ClickHouseProfileEvents,host=me,instance=localhost:9363 RemoteFSCancelledPrefetches",
	"ClickHouseProfileEvents,host=me,instance=localhost:9363 RemoteFSLazySeeks",
	"ClickHouseProfileEvents,host=me,instance=localhost:9363 RemoteFSPrefetchedReads",
	"ClickHouseProfileEvents,host=me,instance=localhost:9363 RemoteFSPrefetches",
	"ClickHouseProfileEvents,host=me,instance=localhost:9363 RemoteFSSeeks",
	"ClickHouseProfileEvents,host=me,instance=localhost:9363 RemoteFSSeeksWithReset",
	"ClickHouseProfileEvents,host=me,instance=localhost:9363 RemoteFSUnprefetchedReads",
	"ClickHouseProfileEvents,host=me,instance=localhost:9363 RemoteFSUnusedPrefetches",
	"ClickHouseProfileEvents,host=me,instance=localhost:9363 ReplicaPartialShutdown",
	"ClickHouseProfileEvents,host=me,instance=localhost:9363 ReplicatedDataLoss",
	"ClickHouseProfileEvents,host=me,instance=localhost:9363 ReplicatedPartChecks",
	"ClickHouseProfileEvents,host=me,instance=localhost:9363 ReplicatedPartChecksFailed",
	"ClickHouseProfileEvents,host=me,instance=localhost:9363 ReplicatedPartFailedFetches",
	"ClickHouseProfileEvents,host=me,instance=localhost:9363 ReplicatedPartFetches",
	"ClickHouseProfileEvents,host=me,instance=localhost:9363 ReplicatedPartFetchesOfMerged",
	"ClickHouseProfileEvents,host=me,instance=localhost:9363 ReplicatedPartMerges",
	"ClickHouseProfileEvents,host=me,instance=localhost:9363 ReplicatedPartMutations",
	"ClickHouseProfileEvents,host=me,instance=localhost:9363 S3ReadMicroseconds",
	"ClickHouseProfileEvents,host=me,instance=localhost:9363 S3ReadRequestsCount",
	"ClickHouseProfileEvents,host=me,instance=localhost:9363 S3ReadRequestsErrors",
	"ClickHouseProfileEvents,host=me,instance=localhost:9363 S3ReadRequestsRedirects",
	"ClickHouseProfileEvents,host=me,instance=localhost:9363 S3ReadRequestsThrottling",
	"ClickHouseProfileEvents,host=me,instance=localhost:9363 S3WriteMicroseconds",
	"ClickHouseProfileEvents,host=me,instance=localhost:9363 S3WriteRequestsCount",
	"ClickHouseProfileEvents,host=me,instance=localhost:9363 S3WriteRequestsErrors",
	"ClickHouseProfileEvents,host=me,instance=localhost:9363 S3WriteRequestsRedirects",
	"ClickHouseProfileEvents,host=me,instance=localhost:9363 S3WriteRequestsThrottling",
	"ClickHouseProfileEvents,host=me,instance=localhost:9363 ScalarSubqueriesCacheMiss",
	"ClickHouseProfileEvents,host=me,instance=localhost:9363 ScalarSubqueriesGlobalCacheHit",
	"ClickHouseProfileEvents,host=me,instance=localhost:9363 ScalarSubqueriesLocalCacheHit",
	"ClickHouseProfileEvents,host=me,instance=localhost:9363 SchemaInferenceCacheEvictions",
	"ClickHouseProfileEvents,host=me,instance=localhost:9363 SchemaInferenceCacheHits",
	"ClickHouseProfileEvents,host=me,instance=localhost:9363 SchemaInferenceCacheInvalidations",
	"ClickHouseProfileEvents,host=me,instance=localhost:9363 SchemaInferenceCacheMisses",
	"ClickHouseProfileEvents,host=me,instance=localhost:9363 Seek",
	"ClickHouseProfileEvents,host=me,instance=localhost:9363 SelectQuery",
	"ClickHouseProfileEvents,host=me,instance=localhost:9363 SelectQueryTimeMicroseconds",
	"ClickHouseProfileEvents,host=me,instance=localhost:9363 SelectedBytes",
	"ClickHouseProfileEvents,host=me,instance=localhost:9363 SelectedMarks",
	"ClickHouseProfileEvents,host=me,instance=localhost:9363 SelectedParts",
	"ClickHouseProfileEvents,host=me,instance=localhost:9363 SelectedRanges",
	"ClickHouseProfileEvents,host=me,instance=localhost:9363 SelectedRows",
	"ClickHouseProfileEvents,host=me,instance=localhost:9363 SleepFunctionCalls",
	"ClickHouseProfileEvents,host=me,instance=localhost:9363 SleepFunctionMicroseconds",
	"ClickHouseProfileEvents,host=me,instance=localhost:9363 SlowRead",
	"ClickHouseProfileEvents,host=me,instance=localhost:9363 SoftPageFaults",
	"ClickHouseProfileEvents,host=me,instance=localhost:9363 StorageBufferErrorOnFlush",
	"ClickHouseProfileEvents,host=me,instance=localhost:9363 StorageBufferFlush",
	"ClickHouseProfileEvents,host=me,instance=localhost:9363 StorageBufferLayerLockReadersWaitMilliseconds",
	"ClickHouseProfileEvents,host=me,instance=localhost:9363 StorageBufferLayerLockWritersWaitMilliseconds",
	"ClickHouseProfileEvents,host=me,instance=localhost:9363 StorageBufferPassedAllMinThresholds",
	"ClickHouseProfileEvents,host=me,instance=localhost:9363 StorageBufferPassedBytesFlushThreshold",
	"ClickHouseProfileEvents,host=me,instance=localhost:9363 StorageBufferPassedBytesMaxThreshold",
	"ClickHouseProfileEvents,host=me,instance=localhost:9363 StorageBufferPassedRowsFlushThreshold",
	"ClickHouseProfileEvents,host=me,instance=localhost:9363 StorageBufferPassedRowsMaxThreshold",
	"ClickHouseProfileEvents,host=me,instance=localhost:9363 StorageBufferPassedTimeFlushThreshold",
	"ClickHouseProfileEvents,host=me,instance=localhost:9363 StorageBufferPassedTimeMaxThreshold",
	"ClickHouseProfileEvents,host=me,instance=localhost:9363 SystemTimeMicroseconds",
	"ClickHouseProfileEvents,host=me,instance=localhost:9363 TableFunctionExecute",
	"ClickHouseProfileEvents,host=me,instance=localhost:9363 ThreadPoolReaderPageCacheHit",
	"ClickHouseProfileEvents,host=me,instance=localhost:9363 ThreadPoolReaderPageCacheHitBytes",
	"ClickHouseProfileEvents,host=me,instance=localhost:9363 ThreadPoolReaderPageCacheHitElapsedMicroseconds",
	"ClickHouseProfileEvents,host=me,instance=localhost:9363 ThreadPoolReaderPageCacheMiss",
	"ClickHouseProfileEvents,host=me,instance=localhost:9363 ThreadPoolReaderPageCacheMissBytes",
	"ClickHouseProfileEvents,host=me,instance=localhost:9363 ThreadPoolReaderPageCacheMissElapsedMicroseconds",
	"ClickHouseProfileEvents,host=me,instance=localhost:9363 ThreadpoolReaderReadBytes",
	"ClickHouseProfileEvents,host=me,instance=localhost:9363 ThreadpoolReaderTaskMicroseconds",
	"ClickHouseProfileEvents,host=me,instance=localhost:9363 ThrottlerSleepMicroseconds",
	"ClickHouseProfileEvents,host=me,instance=localhost:9363 UncompressedCacheHits",
	"ClickHouseProfileEvents,host=me,instance=localhost:9363 UncompressedCacheMisses",
	"ClickHouseProfileEvents,host=me,instance=localhost:9363 UncompressedCacheWeightLost",
	"ClickHouseProfileEvents,host=me,instance=localhost:9363 UserTimeMicroseconds",
	"ClickHouseProfileEvents,host=me,instance=localhost:9363 WriteBufferFromFileDescriptorWrite",
	"ClickHouseProfileEvents,host=me,instance=localhost:9363 WriteBufferFromFileDescriptorWriteBytes",
	"ClickHouseProfileEvents,host=me,instance=localhost:9363 WriteBufferFromFileDescriptorWriteFailed",
	"ClickHouseProfileEvents,host=me,instance=localhost:9363 WriteBufferFromS3Bytes",
	"ClickHouseProfileEvents,host=me,instance=localhost:9363 ZooKeeperBytesReceived",
	"ClickHouseProfileEvents,host=me,instance=localhost:9363 ZooKeeperBytesSent",
	"ClickHouseProfileEvents,host=me,instance=localhost:9363 ZooKeeperCheck",
	"ClickHouseProfileEvents,host=me,instance=localhost:9363 ZooKeeperClose",
	"ClickHouseProfileEvents,host=me,instance=localhost:9363 ZooKeeperCreate",
	"ClickHouseProfileEvents,host=me,instance=localhost:9363 ZooKeeperExists",
	"ClickHouseProfileEvents,host=me,instance=localhost:9363 ZooKeeperGet",
	"ClickHouseProfileEvents,host=me,instance=localhost:9363 ZooKeeperHardwareExceptions",
	"ClickHouseProfileEvents,host=me,instance=localhost:9363 ZooKeeperInit",
	"ClickHouseProfileEvents,host=me,instance=localhost:9363 ZooKeeperList",
	"ClickHouseProfileEvents,host=me,instance=localhost:9363 ZooKeeperMulti",
	"ClickHouseProfileEvents,host=me,instance=localhost:9363 ZooKeeperOtherExceptions",
	"ClickHouseProfileEvents,host=me,instance=localhost:9363 ZooKeeperRemove",
	"ClickHouseProfileEvents,host=me,instance=localhost:9363 ZooKeeperSet",
	"ClickHouseProfileEvents,host=me,instance=localhost:9363 ZooKeeperSync",
	"ClickHouseProfileEvents,host=me,instance=localhost:9363 ZooKeeperTransactions",
	"ClickHouseProfileEvents,host=me,instance=localhost:9363 ZooKeeperUserExceptions",
	"ClickHouseProfileEvents,host=me,instance=localhost:9363 ZooKeeperWaitMicroseconds",
	"ClickHouseProfileEvents,host=me,instance=localhost:9363 ZooKeeperWatchResponse",
}

var mockBody_v21_8_15_7 string = `
# HELP ClickHouseProfileEvents_Query Number of queries to be interpreted and potentially executed. Does not include queries that failed to parse or were rejected due to AST size limits, quota limits or limits on the number of simultaneously running queries. May include internal queries initiated by ClickHouse itself. Does not count subqueries.
# TYPE ClickHouseProfileEvents_Query counter
ClickHouseProfileEvents_Query 0
# HELP ClickHouseProfileEvents_SelectQuery Same as Query, but only for SELECT queries.
# TYPE ClickHouseProfileEvents_SelectQuery counter
ClickHouseProfileEvents_SelectQuery 0
# HELP ClickHouseProfileEvents_InsertQuery Same as Query, but only for INSERT queries.
# TYPE ClickHouseProfileEvents_InsertQuery counter
ClickHouseProfileEvents_InsertQuery 0
# HELP ClickHouseProfileEvents_FailedQuery Number of failed queries.
# TYPE ClickHouseProfileEvents_FailedQuery counter
ClickHouseProfileEvents_FailedQuery 0
# HELP ClickHouseProfileEvents_FailedSelectQuery Same as FailedQuery, but only for SELECT queries.
# TYPE ClickHouseProfileEvents_FailedSelectQuery counter
ClickHouseProfileEvents_FailedSelectQuery 0
# HELP ClickHouseProfileEvents_FailedInsertQuery Same as FailedQuery, but only for INSERT queries.
# TYPE ClickHouseProfileEvents_FailedInsertQuery counter
ClickHouseProfileEvents_FailedInsertQuery 0
# HELP ClickHouseProfileEvents_QueryTimeMicroseconds Total time of all queries.
# TYPE ClickHouseProfileEvents_QueryTimeMicroseconds counter
ClickHouseProfileEvents_QueryTimeMicroseconds 0
# HELP ClickHouseProfileEvents_SelectQueryTimeMicroseconds Total time of SELECT queries.
# TYPE ClickHouseProfileEvents_SelectQueryTimeMicroseconds counter
ClickHouseProfileEvents_SelectQueryTimeMicroseconds 0
# HELP ClickHouseProfileEvents_InsertQueryTimeMicroseconds Total time of INSERT queries.
# TYPE ClickHouseProfileEvents_InsertQueryTimeMicroseconds counter
ClickHouseProfileEvents_InsertQueryTimeMicroseconds 0
# HELP ClickHouseProfileEvents_FileOpen Number of files opened.
# TYPE ClickHouseProfileEvents_FileOpen counter
ClickHouseProfileEvents_FileOpen 155
# HELP ClickHouseProfileEvents_Seek Number of times the 'lseek' function was called.
# TYPE ClickHouseProfileEvents_Seek counter
ClickHouseProfileEvents_Seek 896
# HELP ClickHouseProfileEvents_ReadBufferFromFileDescriptorRead Number of reads (read/pread) from a file descriptor. Does not include sockets.
# TYPE ClickHouseProfileEvents_ReadBufferFromFileDescriptorRead counter
ClickHouseProfileEvents_ReadBufferFromFileDescriptorRead 1176
# HELP ClickHouseProfileEvents_ReadBufferFromFileDescriptorReadFailed Number of times the read (read/pread) from a file descriptor have failed.
# TYPE ClickHouseProfileEvents_ReadBufferFromFileDescriptorReadFailed counter
ClickHouseProfileEvents_ReadBufferFromFileDescriptorReadFailed 0
# HELP ClickHouseProfileEvents_ReadBufferFromFileDescriptorReadBytes Number of bytes read from file descriptors. If the file is compressed, this will show the compressed data size.
# TYPE ClickHouseProfileEvents_ReadBufferFromFileDescriptorReadBytes counter
ClickHouseProfileEvents_ReadBufferFromFileDescriptorReadBytes 624277
# HELP ClickHouseProfileEvents_WriteBufferFromFileDescriptorWrite Number of writes (write/pwrite) to a file descriptor. Does not include sockets.
# TYPE ClickHouseProfileEvents_WriteBufferFromFileDescriptorWrite counter
ClickHouseProfileEvents_WriteBufferFromFileDescriptorWrite 124
# HELP ClickHouseProfileEvents_WriteBufferFromFileDescriptorWriteFailed Number of times the write (write/pwrite) to a file descriptor have failed.
# TYPE ClickHouseProfileEvents_WriteBufferFromFileDescriptorWriteFailed counter
ClickHouseProfileEvents_WriteBufferFromFileDescriptorWriteFailed 0
# HELP ClickHouseProfileEvents_WriteBufferFromFileDescriptorWriteBytes Number of bytes written to file descriptors. If the file is compressed, this will show compressed data size.
# TYPE ClickHouseProfileEvents_WriteBufferFromFileDescriptorWriteBytes counter
ClickHouseProfileEvents_WriteBufferFromFileDescriptorWriteBytes 180762
# HELP ClickHouseProfileEvents_ReadBufferAIORead 
# TYPE ClickHouseProfileEvents_ReadBufferAIORead counter
ClickHouseProfileEvents_ReadBufferAIORead 0
# HELP ClickHouseProfileEvents_ReadBufferAIOReadBytes 
# TYPE ClickHouseProfileEvents_ReadBufferAIOReadBytes counter
ClickHouseProfileEvents_ReadBufferAIOReadBytes 0
# HELP ClickHouseProfileEvents_WriteBufferAIOWrite 
# TYPE ClickHouseProfileEvents_WriteBufferAIOWrite counter
ClickHouseProfileEvents_WriteBufferAIOWrite 0
# HELP ClickHouseProfileEvents_WriteBufferAIOWriteBytes 
# TYPE ClickHouseProfileEvents_WriteBufferAIOWriteBytes counter
ClickHouseProfileEvents_WriteBufferAIOWriteBytes 0
# HELP ClickHouseProfileEvents_ReadCompressedBytes Number of bytes (the number of bytes before decompression) read from compressed sources (files, network).
# TYPE ClickHouseProfileEvents_ReadCompressedBytes counter
ClickHouseProfileEvents_ReadCompressedBytes 0
# HELP ClickHouseProfileEvents_CompressedReadBufferBlocks Number of compressed blocks (the blocks of data that are compressed independent of each other) read from compressed sources (files, network).
# TYPE ClickHouseProfileEvents_CompressedReadBufferBlocks counter
ClickHouseProfileEvents_CompressedReadBufferBlocks 0
# HELP ClickHouseProfileEvents_CompressedReadBufferBytes Number of uncompressed bytes (the number of bytes after decompression) read from compressed sources (files, network).
# TYPE ClickHouseProfileEvents_CompressedReadBufferBytes counter
ClickHouseProfileEvents_CompressedReadBufferBytes 0
# HELP ClickHouseProfileEvents_UncompressedCacheHits 
# TYPE ClickHouseProfileEvents_UncompressedCacheHits counter
ClickHouseProfileEvents_UncompressedCacheHits 0
# HELP ClickHouseProfileEvents_UncompressedCacheMisses 
# TYPE ClickHouseProfileEvents_UncompressedCacheMisses counter
ClickHouseProfileEvents_UncompressedCacheMisses 0
# HELP ClickHouseProfileEvents_UncompressedCacheWeightLost 
# TYPE ClickHouseProfileEvents_UncompressedCacheWeightLost counter
ClickHouseProfileEvents_UncompressedCacheWeightLost 0
# HELP ClickHouseProfileEvents_MMappedFileCacheHits 
# TYPE ClickHouseProfileEvents_MMappedFileCacheHits counter
ClickHouseProfileEvents_MMappedFileCacheHits 0
# HELP ClickHouseProfileEvents_MMappedFileCacheMisses 
# TYPE ClickHouseProfileEvents_MMappedFileCacheMisses counter
ClickHouseProfileEvents_MMappedFileCacheMisses 0
# HELP ClickHouseProfileEvents_IOBufferAllocs 
# TYPE ClickHouseProfileEvents_IOBufferAllocs counter
ClickHouseProfileEvents_IOBufferAllocs 247
# HELP ClickHouseProfileEvents_IOBufferAllocBytes 
# TYPE ClickHouseProfileEvents_IOBufferAllocBytes counter
ClickHouseProfileEvents_IOBufferAllocBytes 72502716
# HELP ClickHouseProfileEvents_ArenaAllocChunks 
# TYPE ClickHouseProfileEvents_ArenaAllocChunks counter
ClickHouseProfileEvents_ArenaAllocChunks 0
# HELP ClickHouseProfileEvents_ArenaAllocBytes 
# TYPE ClickHouseProfileEvents_ArenaAllocBytes counter
ClickHouseProfileEvents_ArenaAllocBytes 0
# HELP ClickHouseProfileEvents_FunctionExecute 
# TYPE ClickHouseProfileEvents_FunctionExecute counter
ClickHouseProfileEvents_FunctionExecute 10
# HELP ClickHouseProfileEvents_TableFunctionExecute 
# TYPE ClickHouseProfileEvents_TableFunctionExecute counter
ClickHouseProfileEvents_TableFunctionExecute 0
# HELP ClickHouseProfileEvents_MarkCacheHits 
# TYPE ClickHouseProfileEvents_MarkCacheHits counter
ClickHouseProfileEvents_MarkCacheHits 0
# HELP ClickHouseProfileEvents_MarkCacheMisses 
# TYPE ClickHouseProfileEvents_MarkCacheMisses counter
ClickHouseProfileEvents_MarkCacheMisses 0
# HELP ClickHouseProfileEvents_CreatedReadBufferOrdinary 
# TYPE ClickHouseProfileEvents_CreatedReadBufferOrdinary counter
ClickHouseProfileEvents_CreatedReadBufferOrdinary 0
# HELP ClickHouseProfileEvents_CreatedReadBufferAIO 
# TYPE ClickHouseProfileEvents_CreatedReadBufferAIO counter
ClickHouseProfileEvents_CreatedReadBufferAIO 0
# HELP ClickHouseProfileEvents_CreatedReadBufferAIOFailed 
# TYPE ClickHouseProfileEvents_CreatedReadBufferAIOFailed counter
ClickHouseProfileEvents_CreatedReadBufferAIOFailed 0
# HELP ClickHouseProfileEvents_CreatedReadBufferMMap 
# TYPE ClickHouseProfileEvents_CreatedReadBufferMMap counter
ClickHouseProfileEvents_CreatedReadBufferMMap 0
# HELP ClickHouseProfileEvents_CreatedReadBufferMMapFailed 
# TYPE ClickHouseProfileEvents_CreatedReadBufferMMapFailed counter
ClickHouseProfileEvents_CreatedReadBufferMMapFailed 0
# HELP ClickHouseProfileEvents_DiskReadElapsedMicroseconds Total time spent waiting for read syscall. This include reads from page cache.
# TYPE ClickHouseProfileEvents_DiskReadElapsedMicroseconds counter
ClickHouseProfileEvents_DiskReadElapsedMicroseconds 14765612
# HELP ClickHouseProfileEvents_DiskWriteElapsedMicroseconds Total time spent waiting for write syscall. This include writes to page cache.
# TYPE ClickHouseProfileEvents_DiskWriteElapsedMicroseconds counter
ClickHouseProfileEvents_DiskWriteElapsedMicroseconds 2340
# HELP ClickHouseProfileEvents_NetworkReceiveElapsedMicroseconds Total time spent waiting for data to receive or receiving data from network. Only ClickHouse-related network interaction is included, not by 3rd party libraries.
# TYPE ClickHouseProfileEvents_NetworkReceiveElapsedMicroseconds counter
ClickHouseProfileEvents_NetworkReceiveElapsedMicroseconds 123
# HELP ClickHouseProfileEvents_NetworkSendElapsedMicroseconds Total time spent waiting for data to send to network or sending data to network. Only ClickHouse-related network interaction is included, not by 3rd party libraries..
# TYPE ClickHouseProfileEvents_NetworkSendElapsedMicroseconds counter
ClickHouseProfileEvents_NetworkSendElapsedMicroseconds 0
# HELP ClickHouseProfileEvents_NetworkReceiveBytes Total number of bytes received from network. Only ClickHouse-related network interaction is included, not by 3rd party libraries.
# TYPE ClickHouseProfileEvents_NetworkReceiveBytes counter
ClickHouseProfileEvents_NetworkReceiveBytes 3264
# HELP ClickHouseProfileEvents_NetworkSendBytes Total number of bytes send to network. Only ClickHouse-related network interaction is included, not by 3rd party libraries.
# TYPE ClickHouseProfileEvents_NetworkSendBytes counter
ClickHouseProfileEvents_NetworkSendBytes 0
# HELP ClickHouseProfileEvents_ThrottlerSleepMicroseconds Total time a query was sleeping to conform the 'max_network_bandwidth' setting.
# TYPE ClickHouseProfileEvents_ThrottlerSleepMicroseconds counter
ClickHouseProfileEvents_ThrottlerSleepMicroseconds 0
# HELP ClickHouseProfileEvents_QueryMaskingRulesMatch Number of times query masking rules was successfully matched.
# TYPE ClickHouseProfileEvents_QueryMaskingRulesMatch counter
ClickHouseProfileEvents_QueryMaskingRulesMatch 0
# HELP ClickHouseProfileEvents_ReplicatedPartFetches Number of times a data part was downloaded from replica of a ReplicatedMergeTree table.
# TYPE ClickHouseProfileEvents_ReplicatedPartFetches counter
ClickHouseProfileEvents_ReplicatedPartFetches 0
# HELP ClickHouseProfileEvents_ReplicatedPartFailedFetches Number of times a data part was failed to download from replica of a ReplicatedMergeTree table.
# TYPE ClickHouseProfileEvents_ReplicatedPartFailedFetches counter
ClickHouseProfileEvents_ReplicatedPartFailedFetches 0
# HELP ClickHouseProfileEvents_ObsoleteReplicatedParts 
# TYPE ClickHouseProfileEvents_ObsoleteReplicatedParts counter
ClickHouseProfileEvents_ObsoleteReplicatedParts 0
# HELP ClickHouseProfileEvents_ReplicatedPartMerges Number of times data parts of ReplicatedMergeTree tables were successfully merged.
# TYPE ClickHouseProfileEvents_ReplicatedPartMerges counter
ClickHouseProfileEvents_ReplicatedPartMerges 0
# HELP ClickHouseProfileEvents_ReplicatedPartFetchesOfMerged Number of times we prefer to download already merged part from replica of ReplicatedMergeTree table instead of performing a merge ourself (usually we prefer doing a merge ourself to save network traffic). This happens when we have not all source parts to perform a merge or when the data part is old enough.
# TYPE ClickHouseProfileEvents_ReplicatedPartFetchesOfMerged counter
ClickHouseProfileEvents_ReplicatedPartFetchesOfMerged 0
# HELP ClickHouseProfileEvents_ReplicatedPartMutations 
# TYPE ClickHouseProfileEvents_ReplicatedPartMutations counter
ClickHouseProfileEvents_ReplicatedPartMutations 0
# HELP ClickHouseProfileEvents_ReplicatedPartChecks 
# TYPE ClickHouseProfileEvents_ReplicatedPartChecks counter
ClickHouseProfileEvents_ReplicatedPartChecks 0
# HELP ClickHouseProfileEvents_ReplicatedPartChecksFailed 
# TYPE ClickHouseProfileEvents_ReplicatedPartChecksFailed counter
ClickHouseProfileEvents_ReplicatedPartChecksFailed 0
# HELP ClickHouseProfileEvents_ReplicatedDataLoss Number of times a data part that we wanted doesn't exist on any replica (even on replicas that are offline right now). That data parts are definitely lost. This is normal due to asynchronous replication (if quorum inserts were not enabled), when the replica on which the data part was written was failed and when it became online after fail it doesn't contain that data part.
# TYPE ClickHouseProfileEvents_ReplicatedDataLoss counter
ClickHouseProfileEvents_ReplicatedDataLoss 0
# HELP ClickHouseProfileEvents_InsertedRows Number of rows INSERTed to all tables.
# TYPE ClickHouseProfileEvents_InsertedRows counter
ClickHouseProfileEvents_InsertedRows 7258
# HELP ClickHouseProfileEvents_InsertedBytes Number of bytes (uncompressed; for columns as they stored in memory) INSERTed to all tables.
# TYPE ClickHouseProfileEvents_InsertedBytes counter
ClickHouseProfileEvents_InsertedBytes 275881
# HELP ClickHouseProfileEvents_DelayedInserts Number of times the INSERT of a block to a MergeTree table was throttled due to high number of active data parts for partition.
# TYPE ClickHouseProfileEvents_DelayedInserts counter
ClickHouseProfileEvents_DelayedInserts 0
# HELP ClickHouseProfileEvents_RejectedInserts Number of times the INSERT of a block to a MergeTree table was rejected with 'Too many parts' exception due to high number of active data parts for partition.
# TYPE ClickHouseProfileEvents_RejectedInserts counter
ClickHouseProfileEvents_RejectedInserts 0
# HELP ClickHouseProfileEvents_DelayedInsertsMilliseconds Total number of milliseconds spent while the INSERT of a block to a MergeTree table was throttled due to high number of active data parts for partition.
# TYPE ClickHouseProfileEvents_DelayedInsertsMilliseconds counter
ClickHouseProfileEvents_DelayedInsertsMilliseconds 0
# HELP ClickHouseProfileEvents_DistributedDelayedInserts Number of times the INSERT of a block to a Distributed table was throttled due to high number of pending bytes.
# TYPE ClickHouseProfileEvents_DistributedDelayedInserts counter
ClickHouseProfileEvents_DistributedDelayedInserts 0
# HELP ClickHouseProfileEvents_DistributedRejectedInserts Number of times the INSERT of a block to a Distributed table was rejected with 'Too many bytes' exception due to high number of pending bytes.
# TYPE ClickHouseProfileEvents_DistributedRejectedInserts counter
ClickHouseProfileEvents_DistributedRejectedInserts 0
# HELP ClickHouseProfileEvents_DistributedDelayedInsertsMilliseconds Total number of milliseconds spent while the INSERT of a block to a Distributed table was throttled due to high number of pending bytes.
# TYPE ClickHouseProfileEvents_DistributedDelayedInsertsMilliseconds counter
ClickHouseProfileEvents_DistributedDelayedInsertsMilliseconds 0
# HELP ClickHouseProfileEvents_DuplicatedInsertedBlocks Number of times the INSERTed block to a ReplicatedMergeTree table was deduplicated.
# TYPE ClickHouseProfileEvents_DuplicatedInsertedBlocks counter
ClickHouseProfileEvents_DuplicatedInsertedBlocks 0
# HELP ClickHouseProfileEvents_ZooKeeperInit 
# TYPE ClickHouseProfileEvents_ZooKeeperInit counter
ClickHouseProfileEvents_ZooKeeperInit 0
# HELP ClickHouseProfileEvents_ZooKeeperTransactions 
# TYPE ClickHouseProfileEvents_ZooKeeperTransactions counter
ClickHouseProfileEvents_ZooKeeperTransactions 0
# HELP ClickHouseProfileEvents_ZooKeeperList 
# TYPE ClickHouseProfileEvents_ZooKeeperList counter
ClickHouseProfileEvents_ZooKeeperList 0
# HELP ClickHouseProfileEvents_ZooKeeperCreate 
# TYPE ClickHouseProfileEvents_ZooKeeperCreate counter
ClickHouseProfileEvents_ZooKeeperCreate 0
# HELP ClickHouseProfileEvents_ZooKeeperRemove 
# TYPE ClickHouseProfileEvents_ZooKeeperRemove counter
ClickHouseProfileEvents_ZooKeeperRemove 0
# HELP ClickHouseProfileEvents_ZooKeeperExists 
# TYPE ClickHouseProfileEvents_ZooKeeperExists counter
ClickHouseProfileEvents_ZooKeeperExists 0
# HELP ClickHouseProfileEvents_ZooKeeperGet 
# TYPE ClickHouseProfileEvents_ZooKeeperGet counter
ClickHouseProfileEvents_ZooKeeperGet 0
# HELP ClickHouseProfileEvents_ZooKeeperSet 
# TYPE ClickHouseProfileEvents_ZooKeeperSet counter
ClickHouseProfileEvents_ZooKeeperSet 0
# HELP ClickHouseProfileEvents_ZooKeeperMulti 
# TYPE ClickHouseProfileEvents_ZooKeeperMulti counter
ClickHouseProfileEvents_ZooKeeperMulti 0
# HELP ClickHouseProfileEvents_ZooKeeperCheck 
# TYPE ClickHouseProfileEvents_ZooKeeperCheck counter
ClickHouseProfileEvents_ZooKeeperCheck 0
# HELP ClickHouseProfileEvents_ZooKeeperClose 
# TYPE ClickHouseProfileEvents_ZooKeeperClose counter
ClickHouseProfileEvents_ZooKeeperClose 0
# HELP ClickHouseProfileEvents_ZooKeeperWatchResponse 
# TYPE ClickHouseProfileEvents_ZooKeeperWatchResponse counter
ClickHouseProfileEvents_ZooKeeperWatchResponse 0
# HELP ClickHouseProfileEvents_ZooKeeperUserExceptions 
# TYPE ClickHouseProfileEvents_ZooKeeperUserExceptions counter
ClickHouseProfileEvents_ZooKeeperUserExceptions 0
# HELP ClickHouseProfileEvents_ZooKeeperHardwareExceptions 
# TYPE ClickHouseProfileEvents_ZooKeeperHardwareExceptions counter
ClickHouseProfileEvents_ZooKeeperHardwareExceptions 0
# HELP ClickHouseProfileEvents_ZooKeeperOtherExceptions 
# TYPE ClickHouseProfileEvents_ZooKeeperOtherExceptions counter
ClickHouseProfileEvents_ZooKeeperOtherExceptions 0
# HELP ClickHouseProfileEvents_ZooKeeperWaitMicroseconds 
# TYPE ClickHouseProfileEvents_ZooKeeperWaitMicroseconds counter
ClickHouseProfileEvents_ZooKeeperWaitMicroseconds 0
# HELP ClickHouseProfileEvents_ZooKeeperBytesSent 
# TYPE ClickHouseProfileEvents_ZooKeeperBytesSent counter
ClickHouseProfileEvents_ZooKeeperBytesSent 0
# HELP ClickHouseProfileEvents_ZooKeeperBytesReceived 
# TYPE ClickHouseProfileEvents_ZooKeeperBytesReceived counter
ClickHouseProfileEvents_ZooKeeperBytesReceived 0
# HELP ClickHouseProfileEvents_DistributedConnectionFailTry Total count when distributed connection fails with retry
# TYPE ClickHouseProfileEvents_DistributedConnectionFailTry counter
ClickHouseProfileEvents_DistributedConnectionFailTry 0
# HELP ClickHouseProfileEvents_DistributedConnectionMissingTable 
# TYPE ClickHouseProfileEvents_DistributedConnectionMissingTable counter
ClickHouseProfileEvents_DistributedConnectionMissingTable 0
# HELP ClickHouseProfileEvents_DistributedConnectionStaleReplica 
# TYPE ClickHouseProfileEvents_DistributedConnectionStaleReplica counter
ClickHouseProfileEvents_DistributedConnectionStaleReplica 0
# HELP ClickHouseProfileEvents_DistributedConnectionFailAtAll Total count when distributed connection fails after all retries finished
# TYPE ClickHouseProfileEvents_DistributedConnectionFailAtAll counter
ClickHouseProfileEvents_DistributedConnectionFailAtAll 0
# HELP ClickHouseProfileEvents_HedgedRequestsChangeReplica Total count when timeout for changing replica expired in hedged requests.
# TYPE ClickHouseProfileEvents_HedgedRequestsChangeReplica counter
ClickHouseProfileEvents_HedgedRequestsChangeReplica 0
# HELP ClickHouseProfileEvents_CompileFunction Number of times a compilation of generated LLVM code (to create fused function for complex expressions) was initiated.
# TYPE ClickHouseProfileEvents_CompileFunction counter
ClickHouseProfileEvents_CompileFunction 0
# HELP ClickHouseProfileEvents_CompiledFunctionExecute Number of times a compiled function was executed.
# TYPE ClickHouseProfileEvents_CompiledFunctionExecute counter
ClickHouseProfileEvents_CompiledFunctionExecute 0
# HELP ClickHouseProfileEvents_CompileExpressionsMicroseconds Total time spent for compilation of expressions to LLVM code.
# TYPE ClickHouseProfileEvents_CompileExpressionsMicroseconds counter
ClickHouseProfileEvents_CompileExpressionsMicroseconds 0
# HELP ClickHouseProfileEvents_CompileExpressionsBytes Number of bytes used for expressions compilation.
# TYPE ClickHouseProfileEvents_CompileExpressionsBytes counter
ClickHouseProfileEvents_CompileExpressionsBytes 0
# HELP ClickHouseProfileEvents_ExternalSortWritePart 
# TYPE ClickHouseProfileEvents_ExternalSortWritePart counter
ClickHouseProfileEvents_ExternalSortWritePart 0
# HELP ClickHouseProfileEvents_ExternalSortMerge 
# TYPE ClickHouseProfileEvents_ExternalSortMerge counter
ClickHouseProfileEvents_ExternalSortMerge 0
# HELP ClickHouseProfileEvents_ExternalAggregationWritePart 
# TYPE ClickHouseProfileEvents_ExternalAggregationWritePart counter
ClickHouseProfileEvents_ExternalAggregationWritePart 0
# HELP ClickHouseProfileEvents_ExternalAggregationMerge 
# TYPE ClickHouseProfileEvents_ExternalAggregationMerge counter
ClickHouseProfileEvents_ExternalAggregationMerge 0
# HELP ClickHouseProfileEvents_ExternalAggregationCompressedBytes 
# TYPE ClickHouseProfileEvents_ExternalAggregationCompressedBytes counter
ClickHouseProfileEvents_ExternalAggregationCompressedBytes 0
# HELP ClickHouseProfileEvents_ExternalAggregationUncompressedBytes 
# TYPE ClickHouseProfileEvents_ExternalAggregationUncompressedBytes counter
ClickHouseProfileEvents_ExternalAggregationUncompressedBytes 0
# HELP ClickHouseProfileEvents_SlowRead Number of reads from a file that were slow. This indicate system overload. Thresholds are controlled by read_backoff_* settings.
# TYPE ClickHouseProfileEvents_SlowRead counter
ClickHouseProfileEvents_SlowRead 0
# HELP ClickHouseProfileEvents_ReadBackoff Number of times the number of query processing threads was lowered due to slow reads.
# TYPE ClickHouseProfileEvents_ReadBackoff counter
ClickHouseProfileEvents_ReadBackoff 0
# HELP ClickHouseProfileEvents_ReplicaPartialShutdown How many times Replicated table has to deinitialize its state due to session expiration in ZooKeeper. The state is reinitialized every time when ZooKeeper is available again.
# TYPE ClickHouseProfileEvents_ReplicaPartialShutdown counter
ClickHouseProfileEvents_ReplicaPartialShutdown 0
# HELP ClickHouseProfileEvents_SelectedParts Number of data parts selected to read from a MergeTree table.
# TYPE ClickHouseProfileEvents_SelectedParts counter
ClickHouseProfileEvents_SelectedParts 0
# HELP ClickHouseProfileEvents_SelectedRanges Number of (non-adjacent) ranges in all data parts selected to read from a MergeTree table.
# TYPE ClickHouseProfileEvents_SelectedRanges counter
ClickHouseProfileEvents_SelectedRanges 0
# HELP ClickHouseProfileEvents_SelectedMarks Number of marks (index granules) selected to read from a MergeTree table.
# TYPE ClickHouseProfileEvents_SelectedMarks counter
ClickHouseProfileEvents_SelectedMarks 0
# HELP ClickHouseProfileEvents_SelectedRows Number of rows SELECTed from all tables.
# TYPE ClickHouseProfileEvents_SelectedRows counter
ClickHouseProfileEvents_SelectedRows 0
# HELP ClickHouseProfileEvents_SelectedBytes Number of bytes (uncompressed; for columns as they stored in memory) SELECTed from all tables.
# TYPE ClickHouseProfileEvents_SelectedBytes counter
ClickHouseProfileEvents_SelectedBytes 0
# HELP ClickHouseProfileEvents_Merge Number of launched background merges.
# TYPE ClickHouseProfileEvents_Merge counter
ClickHouseProfileEvents_Merge 0
# HELP ClickHouseProfileEvents_MergedRows Rows read for background merges. This is the number of rows before merge.
# TYPE ClickHouseProfileEvents_MergedRows counter
ClickHouseProfileEvents_MergedRows 0
# HELP ClickHouseProfileEvents_MergedUncompressedBytes Uncompressed bytes (for columns as they stored in memory) that was read for background merges. This is the number before merge.
# TYPE ClickHouseProfileEvents_MergedUncompressedBytes counter
ClickHouseProfileEvents_MergedUncompressedBytes 0
# HELP ClickHouseProfileEvents_MergesTimeMilliseconds Total time spent for background merges.
# TYPE ClickHouseProfileEvents_MergesTimeMilliseconds counter
ClickHouseProfileEvents_MergesTimeMilliseconds 0
# HELP ClickHouseProfileEvents_MergeTreeDataWriterRows Number of rows INSERTed to MergeTree tables.
# TYPE ClickHouseProfileEvents_MergeTreeDataWriterRows counter
ClickHouseProfileEvents_MergeTreeDataWriterRows 7258
# HELP ClickHouseProfileEvents_MergeTreeDataWriterUncompressedBytes Uncompressed bytes (for columns as they stored in memory) INSERTed to MergeTree tables.
# TYPE ClickHouseProfileEvents_MergeTreeDataWriterUncompressedBytes counter
ClickHouseProfileEvents_MergeTreeDataWriterUncompressedBytes 275881
# HELP ClickHouseProfileEvents_MergeTreeDataWriterCompressedBytes Bytes written to filesystem for data INSERTed to MergeTree tables.
# TYPE ClickHouseProfileEvents_MergeTreeDataWriterCompressedBytes counter
ClickHouseProfileEvents_MergeTreeDataWriterCompressedBytes 111153
# HELP ClickHouseProfileEvents_MergeTreeDataWriterBlocks Number of blocks INSERTed to MergeTree tables. Each block forms a data part of level zero.
# TYPE ClickHouseProfileEvents_MergeTreeDataWriterBlocks counter
ClickHouseProfileEvents_MergeTreeDataWriterBlocks 10
# HELP ClickHouseProfileEvents_MergeTreeDataWriterBlocksAlreadySorted Number of blocks INSERTed to MergeTree tables that appeared to be already sorted.
# TYPE ClickHouseProfileEvents_MergeTreeDataWriterBlocksAlreadySorted counter
ClickHouseProfileEvents_MergeTreeDataWriterBlocksAlreadySorted 10
# HELP ClickHouseProfileEvents_MergeTreeDataProjectionWriterRows Number of rows INSERTed to MergeTree tables projection.
# TYPE ClickHouseProfileEvents_MergeTreeDataProjectionWriterRows counter
ClickHouseProfileEvents_MergeTreeDataProjectionWriterRows 0
# HELP ClickHouseProfileEvents_MergeTreeDataProjectionWriterUncompressedBytes Uncompressed bytes (for columns as they stored in memory) INSERTed to MergeTree tables projection.
# TYPE ClickHouseProfileEvents_MergeTreeDataProjectionWriterUncompressedBytes counter
ClickHouseProfileEvents_MergeTreeDataProjectionWriterUncompressedBytes 0
# HELP ClickHouseProfileEvents_MergeTreeDataProjectionWriterCompressedBytes Bytes written to filesystem for data INSERTed to MergeTree tables projection.
# TYPE ClickHouseProfileEvents_MergeTreeDataProjectionWriterCompressedBytes counter
ClickHouseProfileEvents_MergeTreeDataProjectionWriterCompressedBytes 0
# HELP ClickHouseProfileEvents_MergeTreeDataProjectionWriterBlocks Number of blocks INSERTed to MergeTree tables projection. Each block forms a data part of level zero.
# TYPE ClickHouseProfileEvents_MergeTreeDataProjectionWriterBlocks counter
ClickHouseProfileEvents_MergeTreeDataProjectionWriterBlocks 0
# HELP ClickHouseProfileEvents_MergeTreeDataProjectionWriterBlocksAlreadySorted Number of blocks INSERTed to MergeTree tables projection that appeared to be already sorted.
# TYPE ClickHouseProfileEvents_MergeTreeDataProjectionWriterBlocksAlreadySorted counter
ClickHouseProfileEvents_MergeTreeDataProjectionWriterBlocksAlreadySorted 0
# HELP ClickHouseProfileEvents_CannotRemoveEphemeralNode Number of times an error happened while trying to remove ephemeral node. This is not an issue, because our implementation of ZooKeeper library guarantee that the session will expire and the node will be removed.
# TYPE ClickHouseProfileEvents_CannotRemoveEphemeralNode counter
ClickHouseProfileEvents_CannotRemoveEphemeralNode 0
# HELP ClickHouseProfileEvents_RegexpCreated Compiled regular expressions. Identical regular expressions compiled just once and cached forever.
# TYPE ClickHouseProfileEvents_RegexpCreated counter
ClickHouseProfileEvents_RegexpCreated 0
# HELP ClickHouseProfileEvents_ContextLock Number of times the lock of Context was acquired or tried to acquire. This is global lock.
# TYPE ClickHouseProfileEvents_ContextLock counter
ClickHouseProfileEvents_ContextLock 392
# HELP ClickHouseProfileEvents_StorageBufferFlush 
# TYPE ClickHouseProfileEvents_StorageBufferFlush counter
ClickHouseProfileEvents_StorageBufferFlush 0
# HELP ClickHouseProfileEvents_StorageBufferErrorOnFlush 
# TYPE ClickHouseProfileEvents_StorageBufferErrorOnFlush counter
ClickHouseProfileEvents_StorageBufferErrorOnFlush 0
# HELP ClickHouseProfileEvents_StorageBufferPassedAllMinThresholds 
# TYPE ClickHouseProfileEvents_StorageBufferPassedAllMinThresholds counter
ClickHouseProfileEvents_StorageBufferPassedAllMinThresholds 0
# HELP ClickHouseProfileEvents_StorageBufferPassedTimeMaxThreshold 
# TYPE ClickHouseProfileEvents_StorageBufferPassedTimeMaxThreshold counter
ClickHouseProfileEvents_StorageBufferPassedTimeMaxThreshold 0
# HELP ClickHouseProfileEvents_StorageBufferPassedRowsMaxThreshold 
# TYPE ClickHouseProfileEvents_StorageBufferPassedRowsMaxThreshold counter
ClickHouseProfileEvents_StorageBufferPassedRowsMaxThreshold 0
# HELP ClickHouseProfileEvents_StorageBufferPassedBytesMaxThreshold 
# TYPE ClickHouseProfileEvents_StorageBufferPassedBytesMaxThreshold counter
ClickHouseProfileEvents_StorageBufferPassedBytesMaxThreshold 0
# HELP ClickHouseProfileEvents_StorageBufferPassedTimeFlushThreshold 
# TYPE ClickHouseProfileEvents_StorageBufferPassedTimeFlushThreshold counter
ClickHouseProfileEvents_StorageBufferPassedTimeFlushThreshold 0
# HELP ClickHouseProfileEvents_StorageBufferPassedRowsFlushThreshold 
# TYPE ClickHouseProfileEvents_StorageBufferPassedRowsFlushThreshold counter
ClickHouseProfileEvents_StorageBufferPassedRowsFlushThreshold 0
# HELP ClickHouseProfileEvents_StorageBufferPassedBytesFlushThreshold 
# TYPE ClickHouseProfileEvents_StorageBufferPassedBytesFlushThreshold counter
ClickHouseProfileEvents_StorageBufferPassedBytesFlushThreshold 0
# HELP ClickHouseProfileEvents_StorageBufferLayerLockReadersWaitMilliseconds Time for waiting for Buffer layer during reading
# TYPE ClickHouseProfileEvents_StorageBufferLayerLockReadersWaitMilliseconds counter
ClickHouseProfileEvents_StorageBufferLayerLockReadersWaitMilliseconds 0
# HELP ClickHouseProfileEvents_StorageBufferLayerLockWritersWaitMilliseconds Time for waiting free Buffer layer to write to (can be used to tune Buffer layers)
# TYPE ClickHouseProfileEvents_StorageBufferLayerLockWritersWaitMilliseconds counter
ClickHouseProfileEvents_StorageBufferLayerLockWritersWaitMilliseconds 0
# HELP ClickHouseProfileEvents_DictCacheKeysRequested 
# TYPE ClickHouseProfileEvents_DictCacheKeysRequested counter
ClickHouseProfileEvents_DictCacheKeysRequested 0
# HELP ClickHouseProfileEvents_DictCacheKeysRequestedMiss 
# TYPE ClickHouseProfileEvents_DictCacheKeysRequestedMiss counter
ClickHouseProfileEvents_DictCacheKeysRequestedMiss 0
# HELP ClickHouseProfileEvents_DictCacheKeysRequestedFound 
# TYPE ClickHouseProfileEvents_DictCacheKeysRequestedFound counter
ClickHouseProfileEvents_DictCacheKeysRequestedFound 0
# HELP ClickHouseProfileEvents_DictCacheKeysExpired 
# TYPE ClickHouseProfileEvents_DictCacheKeysExpired counter
ClickHouseProfileEvents_DictCacheKeysExpired 0
# HELP ClickHouseProfileEvents_DictCacheKeysNotFound 
# TYPE ClickHouseProfileEvents_DictCacheKeysNotFound counter
ClickHouseProfileEvents_DictCacheKeysNotFound 0
# HELP ClickHouseProfileEvents_DictCacheKeysHit 
# TYPE ClickHouseProfileEvents_DictCacheKeysHit counter
ClickHouseProfileEvents_DictCacheKeysHit 0
# HELP ClickHouseProfileEvents_DictCacheRequestTimeNs 
# TYPE ClickHouseProfileEvents_DictCacheRequestTimeNs counter
ClickHouseProfileEvents_DictCacheRequestTimeNs 0
# HELP ClickHouseProfileEvents_DictCacheRequests 
# TYPE ClickHouseProfileEvents_DictCacheRequests counter
ClickHouseProfileEvents_DictCacheRequests 0
# HELP ClickHouseProfileEvents_DictCacheLockWriteNs 
# TYPE ClickHouseProfileEvents_DictCacheLockWriteNs counter
ClickHouseProfileEvents_DictCacheLockWriteNs 0
# HELP ClickHouseProfileEvents_DictCacheLockReadNs 
# TYPE ClickHouseProfileEvents_DictCacheLockReadNs counter
ClickHouseProfileEvents_DictCacheLockReadNs 0
# HELP ClickHouseProfileEvents_DistributedSyncInsertionTimeoutExceeded 
# TYPE ClickHouseProfileEvents_DistributedSyncInsertionTimeoutExceeded counter
ClickHouseProfileEvents_DistributedSyncInsertionTimeoutExceeded 0
# HELP ClickHouseProfileEvents_DataAfterMergeDiffersFromReplica 
# TYPE ClickHouseProfileEvents_DataAfterMergeDiffersFromReplica counter
ClickHouseProfileEvents_DataAfterMergeDiffersFromReplica 0
# HELP ClickHouseProfileEvents_DataAfterMutationDiffersFromReplica 
# TYPE ClickHouseProfileEvents_DataAfterMutationDiffersFromReplica counter
ClickHouseProfileEvents_DataAfterMutationDiffersFromReplica 0
# HELP ClickHouseProfileEvents_PolygonsAddedToPool 
# TYPE ClickHouseProfileEvents_PolygonsAddedToPool counter
ClickHouseProfileEvents_PolygonsAddedToPool 0
# HELP ClickHouseProfileEvents_PolygonsInPoolAllocatedBytes 
# TYPE ClickHouseProfileEvents_PolygonsInPoolAllocatedBytes counter
ClickHouseProfileEvents_PolygonsInPoolAllocatedBytes 0
# HELP ClickHouseProfileEvents_RWLockAcquiredReadLocks 
# TYPE ClickHouseProfileEvents_RWLockAcquiredReadLocks counter
ClickHouseProfileEvents_RWLockAcquiredReadLocks 233
# HELP ClickHouseProfileEvents_RWLockAcquiredWriteLocks 
# TYPE ClickHouseProfileEvents_RWLockAcquiredWriteLocks counter
ClickHouseProfileEvents_RWLockAcquiredWriteLocks 0
# HELP ClickHouseProfileEvents_RWLockReadersWaitMilliseconds 
# TYPE ClickHouseProfileEvents_RWLockReadersWaitMilliseconds counter
ClickHouseProfileEvents_RWLockReadersWaitMilliseconds 0
# HELP ClickHouseProfileEvents_RWLockWritersWaitMilliseconds 
# TYPE ClickHouseProfileEvents_RWLockWritersWaitMilliseconds counter
ClickHouseProfileEvents_RWLockWritersWaitMilliseconds 0
# HELP ClickHouseProfileEvents_DNSError Total count of errors in DNS resolution
# TYPE ClickHouseProfileEvents_DNSError counter
ClickHouseProfileEvents_DNSError 0
# HELP ClickHouseProfileEvents_RealTimeMicroseconds Total (wall clock) time spent in processing (queries and other tasks) threads (not that this is a sum).
# TYPE ClickHouseProfileEvents_RealTimeMicroseconds counter
ClickHouseProfileEvents_RealTimeMicroseconds 0
# HELP ClickHouseProfileEvents_UserTimeMicroseconds Total time spent in processing (queries and other tasks) threads executing CPU instructions in user space. This include time CPU pipeline was stalled due to cache misses, branch mispredictions, hyper-threading, etc.
# TYPE ClickHouseProfileEvents_UserTimeMicroseconds counter
ClickHouseProfileEvents_UserTimeMicroseconds 0
# HELP ClickHouseProfileEvents_SystemTimeMicroseconds Total time spent in processing (queries and other tasks) threads executing CPU instructions in OS kernel space. This include time CPU pipeline was stalled due to cache misses, branch mispredictions, hyper-threading, etc.
# TYPE ClickHouseProfileEvents_SystemTimeMicroseconds counter
ClickHouseProfileEvents_SystemTimeMicroseconds 0
# HELP ClickHouseProfileEvents_SoftPageFaults 
# TYPE ClickHouseProfileEvents_SoftPageFaults counter
ClickHouseProfileEvents_SoftPageFaults 0
# HELP ClickHouseProfileEvents_HardPageFaults 
# TYPE ClickHouseProfileEvents_HardPageFaults counter
ClickHouseProfileEvents_HardPageFaults 0
# HELP ClickHouseProfileEvents_VoluntaryContextSwitches 
# TYPE ClickHouseProfileEvents_VoluntaryContextSwitches counter
ClickHouseProfileEvents_VoluntaryContextSwitches 0
# HELP ClickHouseProfileEvents_InvoluntaryContextSwitches 
# TYPE ClickHouseProfileEvents_InvoluntaryContextSwitches counter
ClickHouseProfileEvents_InvoluntaryContextSwitches 0
# HELP ClickHouseProfileEvents_OSIOWaitMicroseconds Total time a thread spent waiting for a result of IO operation, from the OS point of view. This is real IO that doesn't include page cache.
# TYPE ClickHouseProfileEvents_OSIOWaitMicroseconds counter
ClickHouseProfileEvents_OSIOWaitMicroseconds 0
# HELP ClickHouseProfileEvents_OSCPUWaitMicroseconds Total time a thread was ready for execution but waiting to be scheduled by OS, from the OS point of view.
# TYPE ClickHouseProfileEvents_OSCPUWaitMicroseconds counter
ClickHouseProfileEvents_OSCPUWaitMicroseconds 0
# HELP ClickHouseProfileEvents_OSCPUVirtualTimeMicroseconds CPU time spent seen by OS. Does not include involuntary waits due to virtualization.
# TYPE ClickHouseProfileEvents_OSCPUVirtualTimeMicroseconds counter
ClickHouseProfileEvents_OSCPUVirtualTimeMicroseconds 0
# HELP ClickHouseProfileEvents_OSReadBytes Number of bytes read from disks or block devices. Doesn't include bytes read from page cache. May include excessive data due to block size, readahead, etc.
# TYPE ClickHouseProfileEvents_OSReadBytes counter
ClickHouseProfileEvents_OSReadBytes 0
# HELP ClickHouseProfileEvents_OSWriteBytes Number of bytes written to disks or block devices. Doesn't include bytes that are in page cache dirty pages. May not include data that was written by OS asynchronously.
# TYPE ClickHouseProfileEvents_OSWriteBytes counter
ClickHouseProfileEvents_OSWriteBytes 0
# HELP ClickHouseProfileEvents_OSReadChars Number of bytes read from filesystem, including page cache.
# TYPE ClickHouseProfileEvents_OSReadChars counter
ClickHouseProfileEvents_OSReadChars 0
# HELP ClickHouseProfileEvents_OSWriteChars Number of bytes written to filesystem, including page cache.
# TYPE ClickHouseProfileEvents_OSWriteChars counter
ClickHouseProfileEvents_OSWriteChars 0
# HELP ClickHouseProfileEvents_PerfCpuCycles Total cycles. Be wary of what happens during CPU frequency scaling.
# TYPE ClickHouseProfileEvents_PerfCpuCycles counter
ClickHouseProfileEvents_PerfCpuCycles 0
# HELP ClickHouseProfileEvents_PerfInstructions Retired instructions. Be careful, these can be affected by various issues, most notably hardware interrupt counts.
# TYPE ClickHouseProfileEvents_PerfInstructions counter
ClickHouseProfileEvents_PerfInstructions 0
# HELP ClickHouseProfileEvents_PerfCacheReferences Cache accesses. Usually this indicates Last Level Cache accesses but this may vary depending on your CPU. This may include prefetches and coherency messages; again this depends on the design of your CPU.
# TYPE ClickHouseProfileEvents_PerfCacheReferences counter
ClickHouseProfileEvents_PerfCacheReferences 0
# HELP ClickHouseProfileEvents_PerfCacheMisses Cache misses. Usually this indicates Last Level Cache misses; this is intended to be used in con‐junction with the PERFCOUNTHWCACHEREFERENCES event to calculate cache miss rates.
# TYPE ClickHouseProfileEvents_PerfCacheMisses counter
ClickHouseProfileEvents_PerfCacheMisses 0
# HELP ClickHouseProfileEvents_PerfBranchInstructions Retired branch instructions. Prior to Linux 2.6.35, this used the wrong event on AMD processors.
# TYPE ClickHouseProfileEvents_PerfBranchInstructions counter
ClickHouseProfileEvents_PerfBranchInstructions 0
# HELP ClickHouseProfileEvents_PerfBranchMisses Mispredicted branch instructions.
# TYPE ClickHouseProfileEvents_PerfBranchMisses counter
ClickHouseProfileEvents_PerfBranchMisses 0
# HELP ClickHouseProfileEvents_PerfBusCycles Bus cycles, which can be different from total cycles.
# TYPE ClickHouseProfileEvents_PerfBusCycles counter
ClickHouseProfileEvents_PerfBusCycles 0
# HELP ClickHouseProfileEvents_PerfStalledCyclesFrontend Stalled cycles during issue.
# TYPE ClickHouseProfileEvents_PerfStalledCyclesFrontend counter
ClickHouseProfileEvents_PerfStalledCyclesFrontend 0
# HELP ClickHouseProfileEvents_PerfStalledCyclesBackend Stalled cycles during retirement.
# TYPE ClickHouseProfileEvents_PerfStalledCyclesBackend counter
ClickHouseProfileEvents_PerfStalledCyclesBackend 0
# HELP ClickHouseProfileEvents_PerfRefCpuCycles Total cycles; not affected by CPU frequency scaling.
# TYPE ClickHouseProfileEvents_PerfRefCpuCycles counter
ClickHouseProfileEvents_PerfRefCpuCycles 0
# HELP ClickHouseProfileEvents_PerfCpuClock The CPU clock, a high-resolution per-CPU timer
# TYPE ClickHouseProfileEvents_PerfCpuClock counter
ClickHouseProfileEvents_PerfCpuClock 0
# HELP ClickHouseProfileEvents_PerfTaskClock A clock count specific to the task that is running
# TYPE ClickHouseProfileEvents_PerfTaskClock counter
ClickHouseProfileEvents_PerfTaskClock 0
# HELP ClickHouseProfileEvents_PerfContextSwitches Number of context switches
# TYPE ClickHouseProfileEvents_PerfContextSwitches counter
ClickHouseProfileEvents_PerfContextSwitches 0
# HELP ClickHouseProfileEvents_PerfCpuMigrations Number of times the process has migrated to a new CPU
# TYPE ClickHouseProfileEvents_PerfCpuMigrations counter
ClickHouseProfileEvents_PerfCpuMigrations 0
# HELP ClickHouseProfileEvents_PerfAlignmentFaults Number of alignment faults. These happen when unaligned memory accesses happen; the kernel can handle these but it reduces performance. This happens only on some architectures (never on x86).
# TYPE ClickHouseProfileEvents_PerfAlignmentFaults counter
ClickHouseProfileEvents_PerfAlignmentFaults 0
# HELP ClickHouseProfileEvents_PerfEmulationFaults Number of emulation faults. The kernel sometimes traps on unimplemented instructions and emulates them for user space. This can negatively impact performance.
# TYPE ClickHouseProfileEvents_PerfEmulationFaults counter
ClickHouseProfileEvents_PerfEmulationFaults 0
# HELP ClickHouseProfileEvents_PerfMinEnabledTime For all events, minimum time that an event was enabled. Used to track event multiplexing influence
# TYPE ClickHouseProfileEvents_PerfMinEnabledTime counter
ClickHouseProfileEvents_PerfMinEnabledTime 0
# HELP ClickHouseProfileEvents_PerfMinEnabledRunningTime Running time for event with minimum enabled time. Used to track the amount of event multiplexing
# TYPE ClickHouseProfileEvents_PerfMinEnabledRunningTime counter
ClickHouseProfileEvents_PerfMinEnabledRunningTime 0
# HELP ClickHouseProfileEvents_PerfDataTLBReferences Data TLB references
# TYPE ClickHouseProfileEvents_PerfDataTLBReferences counter
ClickHouseProfileEvents_PerfDataTLBReferences 0
# HELP ClickHouseProfileEvents_PerfDataTLBMisses Data TLB misses
# TYPE ClickHouseProfileEvents_PerfDataTLBMisses counter
ClickHouseProfileEvents_PerfDataTLBMisses 0
# HELP ClickHouseProfileEvents_PerfInstructionTLBReferences Instruction TLB references
# TYPE ClickHouseProfileEvents_PerfInstructionTLBReferences counter
ClickHouseProfileEvents_PerfInstructionTLBReferences 0
# HELP ClickHouseProfileEvents_PerfInstructionTLBMisses Instruction TLB misses
# TYPE ClickHouseProfileEvents_PerfInstructionTLBMisses counter
ClickHouseProfileEvents_PerfInstructionTLBMisses 0
# HELP ClickHouseProfileEvents_PerfLocalMemoryReferences Local NUMA node memory reads
# TYPE ClickHouseProfileEvents_PerfLocalMemoryReferences counter
ClickHouseProfileEvents_PerfLocalMemoryReferences 0
# HELP ClickHouseProfileEvents_PerfLocalMemoryMisses Local NUMA node memory read misses
# TYPE ClickHouseProfileEvents_PerfLocalMemoryMisses counter
ClickHouseProfileEvents_PerfLocalMemoryMisses 0
# HELP ClickHouseProfileEvents_CreatedHTTPConnections Total amount of created HTTP connections (closed or opened).
# TYPE ClickHouseProfileEvents_CreatedHTTPConnections counter
ClickHouseProfileEvents_CreatedHTTPConnections 0
# HELP ClickHouseProfileEvents_CannotWriteToWriteBufferDiscard Number of stack traces dropped by query profiler or signal handler because pipe is full or cannot write to pipe.
# TYPE ClickHouseProfileEvents_CannotWriteToWriteBufferDiscard counter
ClickHouseProfileEvents_CannotWriteToWriteBufferDiscard 0
# HELP ClickHouseProfileEvents_QueryProfilerSignalOverruns Number of times we drop processing of a signal due to overrun plus the number of signals that OS has not delivered due to overrun.
# TYPE ClickHouseProfileEvents_QueryProfilerSignalOverruns counter
ClickHouseProfileEvents_QueryProfilerSignalOverruns 0
# HELP ClickHouseProfileEvents_CreatedLogEntryForMerge Successfully created log entry to merge parts in ReplicatedMergeTree.
# TYPE ClickHouseProfileEvents_CreatedLogEntryForMerge counter
ClickHouseProfileEvents_CreatedLogEntryForMerge 0
# HELP ClickHouseProfileEvents_NotCreatedLogEntryForMerge Log entry to merge parts in ReplicatedMergeTree is not created due to concurrent log update by another replica.
# TYPE ClickHouseProfileEvents_NotCreatedLogEntryForMerge counter
ClickHouseProfileEvents_NotCreatedLogEntryForMerge 0
# HELP ClickHouseProfileEvents_CreatedLogEntryForMutation Successfully created log entry to mutate parts in ReplicatedMergeTree.
# TYPE ClickHouseProfileEvents_CreatedLogEntryForMutation counter
ClickHouseProfileEvents_CreatedLogEntryForMutation 0
# HELP ClickHouseProfileEvents_NotCreatedLogEntryForMutation Log entry to mutate parts in ReplicatedMergeTree is not created due to concurrent log update by another replica.
# TYPE ClickHouseProfileEvents_NotCreatedLogEntryForMutation counter
ClickHouseProfileEvents_NotCreatedLogEntryForMutation 0
# HELP ClickHouseProfileEvents_S3ReadMicroseconds Time of GET and HEAD requests to S3 storage.
# TYPE ClickHouseProfileEvents_S3ReadMicroseconds counter
ClickHouseProfileEvents_S3ReadMicroseconds 0
# HELP ClickHouseProfileEvents_S3ReadBytes Read bytes (incoming) in GET and HEAD requests to S3 storage.
# TYPE ClickHouseProfileEvents_S3ReadBytes counter
ClickHouseProfileEvents_S3ReadBytes 0
# HELP ClickHouseProfileEvents_S3ReadRequestsCount Number of GET and HEAD requests to S3 storage.
# TYPE ClickHouseProfileEvents_S3ReadRequestsCount counter
ClickHouseProfileEvents_S3ReadRequestsCount 0
# HELP ClickHouseProfileEvents_S3ReadRequestsErrors Number of non-throttling errors in GET and HEAD requests to S3 storage.
# TYPE ClickHouseProfileEvents_S3ReadRequestsErrors counter
ClickHouseProfileEvents_S3ReadRequestsErrors 0
# HELP ClickHouseProfileEvents_S3ReadRequestsThrottling Number of 429 and 503 errors in GET and HEAD requests to S3 storage.
# TYPE ClickHouseProfileEvents_S3ReadRequestsThrottling counter
ClickHouseProfileEvents_S3ReadRequestsThrottling 0
# HELP ClickHouseProfileEvents_S3ReadRequestsRedirects Number of redirects in GET and HEAD requests to S3 storage.
# TYPE ClickHouseProfileEvents_S3ReadRequestsRedirects counter
ClickHouseProfileEvents_S3ReadRequestsRedirects 0
# HELP ClickHouseProfileEvents_S3WriteMicroseconds Time of POST, DELETE, PUT and PATCH requests to S3 storage.
# TYPE ClickHouseProfileEvents_S3WriteMicroseconds counter
ClickHouseProfileEvents_S3WriteMicroseconds 0
# HELP ClickHouseProfileEvents_S3WriteBytes Write bytes (outgoing) in POST, DELETE, PUT and PATCH requests to S3 storage.
# TYPE ClickHouseProfileEvents_S3WriteBytes counter
ClickHouseProfileEvents_S3WriteBytes 0
# HELP ClickHouseProfileEvents_S3WriteRequestsCount Number of POST, DELETE, PUT and PATCH requests to S3 storage.
# TYPE ClickHouseProfileEvents_S3WriteRequestsCount counter
ClickHouseProfileEvents_S3WriteRequestsCount 0
# HELP ClickHouseProfileEvents_S3WriteRequestsErrors Number of non-throttling errors in POST, DELETE, PUT and PATCH requests to S3 storage.
# TYPE ClickHouseProfileEvents_S3WriteRequestsErrors counter
ClickHouseProfileEvents_S3WriteRequestsErrors 0
# HELP ClickHouseProfileEvents_S3WriteRequestsThrottling Number of 429 and 503 errors in POST, DELETE, PUT and PATCH requests to S3 storage.
# TYPE ClickHouseProfileEvents_S3WriteRequestsThrottling counter
ClickHouseProfileEvents_S3WriteRequestsThrottling 0
# HELP ClickHouseProfileEvents_S3WriteRequestsRedirects Number of redirects in POST, DELETE, PUT and PATCH requests to S3 storage.
# TYPE ClickHouseProfileEvents_S3WriteRequestsRedirects counter
ClickHouseProfileEvents_S3WriteRequestsRedirects 0
# HELP ClickHouseProfileEvents_QueryMemoryLimitExceeded Number of times when memory limit exceeded for query.
# TYPE ClickHouseProfileEvents_QueryMemoryLimitExceeded counter
ClickHouseProfileEvents_QueryMemoryLimitExceeded 0
# HELP ClickHouseMetrics_Query Number of executing queries
# TYPE ClickHouseMetrics_Query gauge
ClickHouseMetrics_Query 0
# HELP ClickHouseMetrics_Merge Number of executing background merges
# TYPE ClickHouseMetrics_Merge gauge
ClickHouseMetrics_Merge 0
# HELP ClickHouseMetrics_PartMutation Number of mutations (ALTER DELETE/UPDATE)
# TYPE ClickHouseMetrics_PartMutation gauge
ClickHouseMetrics_PartMutation 0
# HELP ClickHouseMetrics_ReplicatedFetch Number of data parts being fetched from replica
# TYPE ClickHouseMetrics_ReplicatedFetch gauge
ClickHouseMetrics_ReplicatedFetch 0
# HELP ClickHouseMetrics_ReplicatedSend Number of data parts being sent to replicas
# TYPE ClickHouseMetrics_ReplicatedSend gauge
ClickHouseMetrics_ReplicatedSend 0
# HELP ClickHouseMetrics_ReplicatedChecks Number of data parts checking for consistency
# TYPE ClickHouseMetrics_ReplicatedChecks gauge
ClickHouseMetrics_ReplicatedChecks 0
# HELP ClickHouseMetrics_BackgroundPoolTask Number of active tasks in BackgroundProcessingPool (merges, mutations, or replication queue bookkeeping)
# TYPE ClickHouseMetrics_BackgroundPoolTask gauge
ClickHouseMetrics_BackgroundPoolTask 0
# HELP ClickHouseMetrics_BackgroundFetchesPoolTask Number of active tasks in BackgroundFetchesPool
# TYPE ClickHouseMetrics_BackgroundFetchesPoolTask gauge
ClickHouseMetrics_BackgroundFetchesPoolTask 0
# HELP ClickHouseMetrics_BackgroundMovePoolTask Number of active tasks in BackgroundProcessingPool for moves
# TYPE ClickHouseMetrics_BackgroundMovePoolTask gauge
ClickHouseMetrics_BackgroundMovePoolTask 0
# HELP ClickHouseMetrics_BackgroundSchedulePoolTask Number of active tasks in BackgroundSchedulePool. This pool is used for periodic ReplicatedMergeTree tasks, like cleaning old data parts, altering data parts, replica re-initialization, etc.
# TYPE ClickHouseMetrics_BackgroundSchedulePoolTask gauge
ClickHouseMetrics_BackgroundSchedulePoolTask 0
# HELP ClickHouseMetrics_BackgroundBufferFlushSchedulePoolTask Number of active tasks in BackgroundBufferFlushSchedulePool. This pool is used for periodic Buffer flushes
# TYPE ClickHouseMetrics_BackgroundBufferFlushSchedulePoolTask gauge
ClickHouseMetrics_BackgroundBufferFlushSchedulePoolTask 0
# HELP ClickHouseMetrics_BackgroundDistributedSchedulePoolTask Number of active tasks in BackgroundDistributedSchedulePool. This pool is used for distributed sends that is done in background.
# TYPE ClickHouseMetrics_BackgroundDistributedSchedulePoolTask gauge
ClickHouseMetrics_BackgroundDistributedSchedulePoolTask 0
# HELP ClickHouseMetrics_BackgroundMessageBrokerSchedulePoolTask Number of active tasks in BackgroundProcessingPool for message streaming
# TYPE ClickHouseMetrics_BackgroundMessageBrokerSchedulePoolTask gauge
ClickHouseMetrics_BackgroundMessageBrokerSchedulePoolTask 0
# HELP ClickHouseMetrics_CacheDictionaryUpdateQueueBatches Number of 'batches' (a set of keys) in update queue in CacheDictionaries.
# TYPE ClickHouseMetrics_CacheDictionaryUpdateQueueBatches gauge
ClickHouseMetrics_CacheDictionaryUpdateQueueBatches 0
# HELP ClickHouseMetrics_CacheDictionaryUpdateQueueKeys Exact number of keys in update queue in CacheDictionaries.
# TYPE ClickHouseMetrics_CacheDictionaryUpdateQueueKeys gauge
ClickHouseMetrics_CacheDictionaryUpdateQueueKeys 0
# HELP ClickHouseMetrics_DiskSpaceReservedForMerge Disk space reserved for currently running background merges. It is slightly more than the total size of currently merging parts.
# TYPE ClickHouseMetrics_DiskSpaceReservedForMerge gauge
ClickHouseMetrics_DiskSpaceReservedForMerge 0
# HELP ClickHouseMetrics_DistributedSend Number of connections to remote servers sending data that was INSERTed into Distributed tables. Both synchronous and asynchronous mode.
# TYPE ClickHouseMetrics_DistributedSend gauge
ClickHouseMetrics_DistributedSend 0
# HELP ClickHouseMetrics_QueryPreempted Number of queries that are stopped and waiting due to 'priority' setting.
# TYPE ClickHouseMetrics_QueryPreempted gauge
ClickHouseMetrics_QueryPreempted 0
# HELP ClickHouseMetrics_TCPConnection Number of connections to TCP server (clients with native interface), also included server-server distributed query connections
# TYPE ClickHouseMetrics_TCPConnection gauge
ClickHouseMetrics_TCPConnection 0
# HELP ClickHouseMetrics_MySQLConnection Number of client connections using MySQL protocol
# TYPE ClickHouseMetrics_MySQLConnection gauge
ClickHouseMetrics_MySQLConnection 0
# HELP ClickHouseMetrics_HTTPConnection Number of connections to HTTP server
# TYPE ClickHouseMetrics_HTTPConnection gauge
ClickHouseMetrics_HTTPConnection 0
# HELP ClickHouseMetrics_InterserverConnection Number of connections from other replicas to fetch parts
# TYPE ClickHouseMetrics_InterserverConnection gauge
ClickHouseMetrics_InterserverConnection 0
# HELP ClickHouseMetrics_PostgreSQLConnection Number of client connections using PostgreSQL protocol
# TYPE ClickHouseMetrics_PostgreSQLConnection gauge
ClickHouseMetrics_PostgreSQLConnection 0
# HELP ClickHouseMetrics_OpenFileForRead Number of files open for reading
# TYPE ClickHouseMetrics_OpenFileForRead gauge
ClickHouseMetrics_OpenFileForRead 36
# HELP ClickHouseMetrics_OpenFileForWrite Number of files open for writing
# TYPE ClickHouseMetrics_OpenFileForWrite gauge
ClickHouseMetrics_OpenFileForWrite 0
# HELP ClickHouseMetrics_Read Number of read (read, pread, io_getevents, etc.) syscalls in fly
# TYPE ClickHouseMetrics_Read gauge
ClickHouseMetrics_Read 2
# HELP ClickHouseMetrics_Write Number of write (write, pwrite, io_getevents, etc.) syscalls in fly
# TYPE ClickHouseMetrics_Write gauge
ClickHouseMetrics_Write 0
# HELP ClickHouseMetrics_NetworkReceive Number of threads receiving data from network. Only ClickHouse-related network interaction is included, not by 3rd party libraries.
# TYPE ClickHouseMetrics_NetworkReceive gauge
ClickHouseMetrics_NetworkReceive 0
# HELP ClickHouseMetrics_NetworkSend Number of threads sending data to network. Only ClickHouse-related network interaction is included, not by 3rd party libraries.
# TYPE ClickHouseMetrics_NetworkSend gauge
ClickHouseMetrics_NetworkSend 0
# HELP ClickHouseMetrics_SendScalars Number of connections that are sending data for scalars to remote servers.
# TYPE ClickHouseMetrics_SendScalars gauge
ClickHouseMetrics_SendScalars 0
# HELP ClickHouseMetrics_SendExternalTables Number of connections that are sending data for external tables to remote servers. External tables are used to implement GLOBAL IN and GLOBAL JOIN operators with distributed subqueries.
# TYPE ClickHouseMetrics_SendExternalTables gauge
ClickHouseMetrics_SendExternalTables 0
# HELP ClickHouseMetrics_QueryThread Number of query processing threads
# TYPE ClickHouseMetrics_QueryThread gauge
ClickHouseMetrics_QueryThread 0
# HELP ClickHouseMetrics_ReadonlyReplica Number of Replicated tables that are currently in readonly state due to re-initialization after ZooKeeper session loss or due to startup without ZooKeeper configured.
# TYPE ClickHouseMetrics_ReadonlyReplica gauge
ClickHouseMetrics_ReadonlyReplica 0
# HELP ClickHouseMetrics_MemoryTracking Total amount of memory (bytes) allocated by the server.
# TYPE ClickHouseMetrics_MemoryTracking gauge
ClickHouseMetrics_MemoryTracking 294608911
# HELP ClickHouseMetrics_EphemeralNode Number of ephemeral nodes hold in ZooKeeper.
# TYPE ClickHouseMetrics_EphemeralNode gauge
ClickHouseMetrics_EphemeralNode 0
# HELP ClickHouseMetrics_ZooKeeperSession Number of sessions (connections) to ZooKeeper. Should be no more than one, because using more than one connection to ZooKeeper may lead to bugs due to lack of linearizability (stale reads) that ZooKeeper consistency model allows.
# TYPE ClickHouseMetrics_ZooKeeperSession gauge
ClickHouseMetrics_ZooKeeperSession 0
# HELP ClickHouseMetrics_ZooKeeperWatch Number of watches (event subscriptions) in ZooKeeper.
# TYPE ClickHouseMetrics_ZooKeeperWatch gauge
ClickHouseMetrics_ZooKeeperWatch 0
# HELP ClickHouseMetrics_ZooKeeperRequest Number of requests to ZooKeeper in fly.
# TYPE ClickHouseMetrics_ZooKeeperRequest gauge
ClickHouseMetrics_ZooKeeperRequest 0
# HELP ClickHouseMetrics_DelayedInserts Number of INSERT queries that are throttled due to high number of active data parts for partition in a MergeTree table.
# TYPE ClickHouseMetrics_DelayedInserts gauge
ClickHouseMetrics_DelayedInserts 0
# HELP ClickHouseMetrics_ContextLockWait Number of threads waiting for lock in Context. This is global lock.
# TYPE ClickHouseMetrics_ContextLockWait gauge
ClickHouseMetrics_ContextLockWait 0
# HELP ClickHouseMetrics_StorageBufferRows Number of rows in buffers of Buffer tables
# TYPE ClickHouseMetrics_StorageBufferRows gauge
ClickHouseMetrics_StorageBufferRows 0
# HELP ClickHouseMetrics_StorageBufferBytes Number of bytes in buffers of Buffer tables
# TYPE ClickHouseMetrics_StorageBufferBytes gauge
ClickHouseMetrics_StorageBufferBytes 0
# HELP ClickHouseMetrics_DictCacheRequests Number of requests in fly to data sources of dictionaries of cache type.
# TYPE ClickHouseMetrics_DictCacheRequests gauge
ClickHouseMetrics_DictCacheRequests 0
# HELP ClickHouseMetrics_Revision Revision of the server. It is a number incremented for every release or release candidate except patch releases.
# TYPE ClickHouseMetrics_Revision gauge
ClickHouseMetrics_Revision 54453
# HELP ClickHouseMetrics_VersionInteger Version of the server in a single integer number in base-1000. For example, version 11.22.33 is translated to 11022033.
# TYPE ClickHouseMetrics_VersionInteger gauge
ClickHouseMetrics_VersionInteger 21008015
# HELP ClickHouseMetrics_RWLockWaitingReaders Number of threads waiting for read on a table RWLock.
# TYPE ClickHouseMetrics_RWLockWaitingReaders gauge
ClickHouseMetrics_RWLockWaitingReaders 0
# HELP ClickHouseMetrics_RWLockWaitingWriters Number of threads waiting for write on a table RWLock.
# TYPE ClickHouseMetrics_RWLockWaitingWriters gauge
ClickHouseMetrics_RWLockWaitingWriters 0
# HELP ClickHouseMetrics_RWLockActiveReaders Number of threads holding read lock in a table RWLock.
# TYPE ClickHouseMetrics_RWLockActiveReaders gauge
ClickHouseMetrics_RWLockActiveReaders 0
# HELP ClickHouseMetrics_RWLockActiveWriters Number of threads holding write lock in a table RWLock.
# TYPE ClickHouseMetrics_RWLockActiveWriters gauge
ClickHouseMetrics_RWLockActiveWriters 0
# HELP ClickHouseMetrics_GlobalThread Number of threads in global thread pool.
# TYPE ClickHouseMetrics_GlobalThread gauge
ClickHouseMetrics_GlobalThread 145
# HELP ClickHouseMetrics_GlobalThreadActive Number of threads in global thread pool running a task.
# TYPE ClickHouseMetrics_GlobalThreadActive gauge
ClickHouseMetrics_GlobalThreadActive 144
# HELP ClickHouseMetrics_LocalThread Number of threads in local thread pools. The threads in local thread pools are taken from the global thread pool.
# TYPE ClickHouseMetrics_LocalThread gauge
ClickHouseMetrics_LocalThread 0
# HELP ClickHouseMetrics_LocalThreadActive Number of threads in local thread pools running a task.
# TYPE ClickHouseMetrics_LocalThreadActive gauge
ClickHouseMetrics_LocalThreadActive 0
# HELP ClickHouseMetrics_DistributedFilesToInsert Number of pending files to process for asynchronous insertion into Distributed tables. Number of files for every shard is summed.
# TYPE ClickHouseMetrics_DistributedFilesToInsert gauge
ClickHouseMetrics_DistributedFilesToInsert 0
# HELP ClickHouseMetrics_BrokenDistributedFilesToInsert Number of files for asynchronous insertion into Distributed tables that has been marked as broken. This metric will starts from 0 on start. Number of files for every shard is summed.
# TYPE ClickHouseMetrics_BrokenDistributedFilesToInsert gauge
ClickHouseMetrics_BrokenDistributedFilesToInsert 0
# HELP ClickHouseMetrics_TablesToDropQueueSize Number of dropped tables, that are waiting for background data removal.
# TYPE ClickHouseMetrics_TablesToDropQueueSize gauge
ClickHouseMetrics_TablesToDropQueueSize 0
# HELP ClickHouseMetrics_MaxDDLEntryID Max processed DDL entry of DDLWorker.
# TYPE ClickHouseMetrics_MaxDDLEntryID gauge
ClickHouseMetrics_MaxDDLEntryID 0
# HELP ClickHouseMetrics_PartsTemporary The part is generating now, it is not in data_parts list.
# TYPE ClickHouseMetrics_PartsTemporary gauge
ClickHouseMetrics_PartsTemporary 0
# HELP ClickHouseMetrics_PartsPreCommitted The part is in data_parts, but not used for SELECTs.
# TYPE ClickHouseMetrics_PartsPreCommitted gauge
ClickHouseMetrics_PartsPreCommitted 0
# HELP ClickHouseMetrics_PartsCommitted Active data part, used by current and upcoming SELECTs.
# TYPE ClickHouseMetrics_PartsCommitted gauge
ClickHouseMetrics_PartsCommitted 10
# HELP ClickHouseMetrics_PartsOutdated Not active data part, but could be used by only current SELECTs, could be deleted after SELECTs finishes.
# TYPE ClickHouseMetrics_PartsOutdated gauge
ClickHouseMetrics_PartsOutdated 0
# HELP ClickHouseMetrics_PartsDeleting Not active data part with identity refcounter, it is deleting right now by a cleaner.
# TYPE ClickHouseMetrics_PartsDeleting gauge
ClickHouseMetrics_PartsDeleting 0
# HELP ClickHouseMetrics_PartsDeleteOnDestroy Part was moved to another disk and should be deleted in own destructor.
# TYPE ClickHouseMetrics_PartsDeleteOnDestroy gauge
ClickHouseMetrics_PartsDeleteOnDestroy 0
# HELP ClickHouseMetrics_PartsWide Wide parts.
# TYPE ClickHouseMetrics_PartsWide gauge
ClickHouseMetrics_PartsWide 0
# HELP ClickHouseMetrics_PartsCompact Compact parts.
# TYPE ClickHouseMetrics_PartsCompact gauge
ClickHouseMetrics_PartsCompact 10
# HELP ClickHouseMetrics_PartsInMemory In-memory parts.
# TYPE ClickHouseMetrics_PartsInMemory gauge
ClickHouseMetrics_PartsInMemory 0
# HELP ClickHouseMetrics_MMappedFiles Total number of mmapped files.
# TYPE ClickHouseMetrics_MMappedFiles gauge
ClickHouseMetrics_MMappedFiles 8
# HELP ClickHouseMetrics_MMappedFileBytes Sum size of mmapped file regions.
# TYPE ClickHouseMetrics_MMappedFileBytes gauge
ClickHouseMetrics_MMappedFileBytes 365545040
# TYPE ClickHouseAsyncMetrics_AsynchronousMetricsCalculationTimeSpent gauge
ClickHouseAsyncMetrics_AsynchronousMetricsCalculationTimeSpent 0.038335538
# TYPE ClickHouseAsyncMetrics_jemalloc_arenas_all_muzzy_purged gauge
ClickHouseAsyncMetrics_jemalloc_arenas_all_muzzy_purged 3491
# TYPE ClickHouseAsyncMetrics_jemalloc_arenas_all_dirty_purged gauge
ClickHouseAsyncMetrics_jemalloc_arenas_all_dirty_purged 5697
# TYPE ClickHouseAsyncMetrics_jemalloc_arenas_all_pmuzzy gauge
ClickHouseAsyncMetrics_jemalloc_arenas_all_pmuzzy 541
# TYPE ClickHouseAsyncMetrics_jemalloc_background_thread_run_intervals gauge
ClickHouseAsyncMetrics_jemalloc_background_thread_run_intervals 0
# TYPE ClickHouseAsyncMetrics_jemalloc_metadata_thp gauge
ClickHouseAsyncMetrics_jemalloc_metadata_thp 0
# TYPE ClickHouseAsyncMetrics_jemalloc_metadata gauge
ClickHouseAsyncMetrics_jemalloc_metadata 11612080
# TYPE ClickHouseAsyncMetrics_jemalloc_allocated gauge
ClickHouseAsyncMetrics_jemalloc_allocated 49307512
# TYPE ClickHouseAsyncMetrics_PostgreSQLThreads gauge
ClickHouseAsyncMetrics_PostgreSQLThreads 0
# TYPE ClickHouseAsyncMetrics_TCPThreads gauge
ClickHouseAsyncMetrics_TCPThreads 0
# TYPE ClickHouseAsyncMetrics_HTTPThreads gauge
ClickHouseAsyncMetrics_HTTPThreads 0
# TYPE ClickHouseAsyncMetrics_TotalPartsOfMergeTreeTables gauge
ClickHouseAsyncMetrics_TotalPartsOfMergeTreeTables 10
# TYPE ClickHouseAsyncMetrics_NumberOfTables gauge
ClickHouseAsyncMetrics_NumberOfTables 67
# TYPE ClickHouseAsyncMetrics_NumberOfDatabases gauge
ClickHouseAsyncMetrics_NumberOfDatabases 3
# TYPE ClickHouseAsyncMetrics_ReplicasMaxRelativeDelay gauge
ClickHouseAsyncMetrics_ReplicasMaxRelativeDelay 0
# TYPE ClickHouseAsyncMetrics_ReplicasMaxAbsoluteDelay gauge
ClickHouseAsyncMetrics_ReplicasMaxAbsoluteDelay 0
# TYPE ClickHouseAsyncMetrics_ReplicasSumMergesInQueue gauge
ClickHouseAsyncMetrics_ReplicasSumMergesInQueue 0
# TYPE ClickHouseAsyncMetrics_ReplicasSumInsertsInQueue gauge
ClickHouseAsyncMetrics_ReplicasSumInsertsInQueue 0
# TYPE ClickHouseAsyncMetrics_ReplicasSumQueueSize gauge
ClickHouseAsyncMetrics_ReplicasSumQueueSize 0
# TYPE ClickHouseAsyncMetrics_ReplicasMaxMergesInQueue gauge
ClickHouseAsyncMetrics_ReplicasMaxMergesInQueue 0
# TYPE ClickHouseAsyncMetrics_ReplicasMaxInsertsInQueue gauge
ClickHouseAsyncMetrics_ReplicasMaxInsertsInQueue 0
# TYPE ClickHouseAsyncMetrics_ReplicasMaxQueueSize gauge
ClickHouseAsyncMetrics_ReplicasMaxQueueSize 0
# TYPE ClickHouseAsyncMetrics_DiskUnreserved_default gauge
ClickHouseAsyncMetrics_DiskUnreserved_default 761144999936
# TYPE ClickHouseAsyncMetrics_DiskUsed_default gauge
ClickHouseAsyncMetrics_DiskUsed_default 221661224960
# TYPE ClickHouseAsyncMetrics_FilesystemLogsPathUsedINodes gauge
ClickHouseAsyncMetrics_FilesystemLogsPathUsedINodes 1732551
# TYPE ClickHouseAsyncMetrics_jemalloc_arenas_all_pactive gauge
ClickHouseAsyncMetrics_jemalloc_arenas_all_pactive 12923
# TYPE ClickHouseAsyncMetrics_FilesystemLogsPathTotalINodes gauge
ClickHouseAsyncMetrics_FilesystemLogsPathTotalINodes 61014016
# TYPE ClickHouseAsyncMetrics_FilesystemLogsPathUsedBytes gauge
ClickHouseAsyncMetrics_FilesystemLogsPathUsedBytes 221661224960
# TYPE ClickHouseAsyncMetrics_DiskTotal_default gauge
ClickHouseAsyncMetrics_DiskTotal_default 982806224896
# TYPE ClickHouseAsyncMetrics_FilesystemLogsPathAvailableBytes gauge
ClickHouseAsyncMetrics_FilesystemLogsPathAvailableBytes 761144999936
# TYPE ClickHouseAsyncMetrics_FilesystemLogsPathTotalBytes gauge
ClickHouseAsyncMetrics_FilesystemLogsPathTotalBytes 982806224896
# TYPE ClickHouseAsyncMetrics_FilesystemMainPathAvailableINodes gauge
ClickHouseAsyncMetrics_FilesystemMainPathAvailableINodes 59281465
# TYPE ClickHouseAsyncMetrics_FilesystemMainPathUsedBytes gauge
ClickHouseAsyncMetrics_FilesystemMainPathUsedBytes 221661224960
# TYPE ClickHouseAsyncMetrics_FilesystemMainPathAvailableBytes gauge
ClickHouseAsyncMetrics_FilesystemMainPathAvailableBytes 761144999936
# TYPE ClickHouseAsyncMetrics_FilesystemMainPathTotalBytes gauge
ClickHouseAsyncMetrics_FilesystemMainPathTotalBytes 982806224896
# TYPE ClickHouseAsyncMetrics_jemalloc_resident gauge
ClickHouseAsyncMetrics_jemalloc_resident 94285824
# TYPE ClickHouseAsyncMetrics_Temperature_nvme_Sensor_1 gauge
ClickHouseAsyncMetrics_Temperature_nvme_Sensor_1 25.85
# TYPE ClickHouseAsyncMetrics_OSNiceTimeCPU5 gauge
ClickHouseAsyncMetrics_OSNiceTimeCPU5 0
# TYPE ClickHouseAsyncMetrics_Temperature_coretemp_Package_id_0 gauge
ClickHouseAsyncMetrics_Temperature_coretemp_Package_id_0 51
# TYPE ClickHouseAsyncMetrics_Temperature_coretemp_Core_1 gauge
ClickHouseAsyncMetrics_Temperature_coretemp_Core_1 50
# TYPE ClickHouseAsyncMetrics_Temperature7 gauge
ClickHouseAsyncMetrics_Temperature7 51
# TYPE ClickHouseAsyncMetrics_jemalloc_active gauge
ClickHouseAsyncMetrics_jemalloc_active 52932608
# TYPE ClickHouseAsyncMetrics_Temperature3 gauge
ClickHouseAsyncMetrics_Temperature3 41.050000000000004
# TYPE ClickHouseAsyncMetrics_OSIrqTimeNormalized gauge
ClickHouseAsyncMetrics_OSIrqTimeNormalized 0
# TYPE ClickHouseAsyncMetrics_Temperature1 gauge
ClickHouseAsyncMetrics_Temperature1 45.050000000000004
# TYPE ClickHouseAsyncMetrics_Temperature0 gauge
ClickHouseAsyncMetrics_Temperature0 52
# TYPE ClickHouseAsyncMetrics_CPUFrequencyMHz_1 gauge
ClickHouseAsyncMetrics_CPUFrequencyMHz_1 1949.692
# TYPE ClickHouseAsyncMetrics_NetworkSendDrop_eth0 gauge
ClickHouseAsyncMetrics_NetworkSendDrop_eth0 0
# TYPE ClickHouseAsyncMetrics_NetworkSendPackets_eth0 gauge
ClickHouseAsyncMetrics_NetworkSendPackets_eth0 0
# TYPE ClickHouseAsyncMetrics_NetworkReceiveDrop_eth0 gauge
ClickHouseAsyncMetrics_NetworkReceiveDrop_eth0 0
# TYPE ClickHouseAsyncMetrics_NetworkSendErrors_eth0 gauge
ClickHouseAsyncMetrics_NetworkSendErrors_eth0 0
# TYPE ClickHouseAsyncMetrics_NetworkReceiveErrors_eth0 gauge
ClickHouseAsyncMetrics_NetworkReceiveErrors_eth0 0
# TYPE ClickHouseAsyncMetrics_NetworkReceivePackets_eth0 gauge
ClickHouseAsyncMetrics_NetworkReceivePackets_eth0 0
# TYPE ClickHouseAsyncMetrics_BlockActiveTime_nvme0n1 gauge
ClickHouseAsyncMetrics_BlockActiveTime_nvme0n1 0
# TYPE ClickHouseAsyncMetrics_BlockDiscardTime_nvme0n1 gauge
ClickHouseAsyncMetrics_BlockDiscardTime_nvme0n1 0
# TYPE ClickHouseAsyncMetrics_BlockWriteTime_nvme0n1 gauge
ClickHouseAsyncMetrics_BlockWriteTime_nvme0n1 0
# TYPE ClickHouseAsyncMetrics_Temperature_pch_cannonlake gauge
ClickHouseAsyncMetrics_Temperature_pch_cannonlake 53
# TYPE ClickHouseAsyncMetrics_BlockReadBytes_nvme0n1 gauge
ClickHouseAsyncMetrics_BlockReadBytes_nvme0n1 0
# TYPE ClickHouseAsyncMetrics_jemalloc_mapped gauge
ClickHouseAsyncMetrics_jemalloc_mapped 114192384
# TYPE ClickHouseAsyncMetrics_BlockReadMerges_nvme0n1 gauge
ClickHouseAsyncMetrics_BlockReadMerges_nvme0n1 0
# TYPE ClickHouseAsyncMetrics_BlockDiscardOps_nvme0n1 gauge
ClickHouseAsyncMetrics_BlockDiscardOps_nvme0n1 0
# TYPE ClickHouseAsyncMetrics_BlockQueueTime_sda gauge
ClickHouseAsyncMetrics_BlockQueueTime_sda 0
# TYPE ClickHouseAsyncMetrics_BlockInFlightOps_sda gauge
ClickHouseAsyncMetrics_BlockInFlightOps_sda 0
# TYPE ClickHouseAsyncMetrics_BlockWriteTime_sda gauge
ClickHouseAsyncMetrics_BlockWriteTime_sda 0
# TYPE ClickHouseAsyncMetrics_Temperature_nvme_Composite gauge
ClickHouseAsyncMetrics_Temperature_nvme_Composite 25.85
# TYPE ClickHouseAsyncMetrics_BlockReadTime_sda gauge
ClickHouseAsyncMetrics_BlockReadTime_sda 0
# TYPE ClickHouseAsyncMetrics_MemoryResident gauge
ClickHouseAsyncMetrics_MemoryResident 293552128
# TYPE ClickHouseAsyncMetrics_OSSoftIrqTimeCPU5 gauge
ClickHouseAsyncMetrics_OSSoftIrqTimeCPU5 0
# TYPE ClickHouseAsyncMetrics_BlockDiscardMerges_sda gauge
ClickHouseAsyncMetrics_BlockDiscardMerges_sda 0
# TYPE ClickHouseAsyncMetrics_BlockDiscardOps_sda gauge
ClickHouseAsyncMetrics_BlockDiscardOps_sda 0
# TYPE ClickHouseAsyncMetrics_BlockWriteOps_sda gauge
ClickHouseAsyncMetrics_BlockWriteOps_sda 0
# TYPE ClickHouseAsyncMetrics_BlockReadOps_sda gauge
ClickHouseAsyncMetrics_BlockReadOps_sda 0
# TYPE ClickHouseAsyncMetrics_jemalloc_epoch gauge
ClickHouseAsyncMetrics_jemalloc_epoch 33
# TYPE ClickHouseAsyncMetrics_OSOpenFiles gauge
ClickHouseAsyncMetrics_OSOpenFiles 18118
# TYPE ClickHouseAsyncMetrics_OSIrqTimeCPU5 gauge
ClickHouseAsyncMetrics_OSIrqTimeCPU5 0
# TYPE ClickHouseAsyncMetrics_BlockQueueTime_nvme0n1 gauge
ClickHouseAsyncMetrics_BlockQueueTime_nvme0n1 0
# TYPE ClickHouseAsyncMetrics_Temperature5 gauge
ClickHouseAsyncMetrics_Temperature5 52.050000000000004
# TYPE ClickHouseAsyncMetrics_DiskAvailable_default gauge
ClickHouseAsyncMetrics_DiskAvailable_default 761144999936
# TYPE ClickHouseAsyncMetrics_CPUFrequencyMHz_7 gauge
ClickHouseAsyncMetrics_CPUFrequencyMHz_7 2400
# TYPE ClickHouseAsyncMetrics_BlockActiveTime_sda gauge
ClickHouseAsyncMetrics_BlockActiveTime_sda 0
# TYPE ClickHouseAsyncMetrics_CPUFrequencyMHz_6 gauge
ClickHouseAsyncMetrics_CPUFrequencyMHz_6 2400
# TYPE ClickHouseAsyncMetrics_CPUFrequencyMHz_4 gauge
ClickHouseAsyncMetrics_CPUFrequencyMHz_4 2400
# TYPE ClickHouseAsyncMetrics_BlockWriteMerges_sda gauge
ClickHouseAsyncMetrics_BlockWriteMerges_sda 0
# TYPE ClickHouseAsyncMetrics_OSNiceTime gauge
ClickHouseAsyncMetrics_OSNiceTime 0
# TYPE ClickHouseAsyncMetrics_CPUFrequencyMHz_3 gauge
ClickHouseAsyncMetrics_CPUFrequencyMHz_3 2400
# TYPE ClickHouseAsyncMetrics_CPUFrequencyMHz_2 gauge
ClickHouseAsyncMetrics_CPUFrequencyMHz_2 2400
# TYPE ClickHouseAsyncMetrics_OSMemoryBuffers gauge
ClickHouseAsyncMetrics_OSMemoryBuffers 1472196608
# TYPE ClickHouseAsyncMetrics_OSMemoryFreeWithoutCached gauge
ClickHouseAsyncMetrics_OSMemoryFreeWithoutCached 4645593088
# TYPE ClickHouseAsyncMetrics_OSGuestNiceTimeCPU3 gauge
ClickHouseAsyncMetrics_OSGuestNiceTimeCPU3 0
# TYPE ClickHouseAsyncMetrics_Temperature_nvme_Sensor_2 gauge
ClickHouseAsyncMetrics_Temperature_nvme_Sensor_2 34.85
# TYPE ClickHouseAsyncMetrics_OSGuestNiceTimeNormalized gauge
ClickHouseAsyncMetrics_OSGuestNiceTimeNormalized 0
# TYPE ClickHouseAsyncMetrics_OSSoftIrqTimeNormalized gauge
ClickHouseAsyncMetrics_OSSoftIrqTimeNormalized 0.0012499137559508393
# TYPE ClickHouseAsyncMetrics_Temperature_acpitz gauge
ClickHouseAsyncMetrics_Temperature_acpitz 52
# TYPE ClickHouseAsyncMetrics_OSSystemTimeNormalized gauge
ClickHouseAsyncMetrics_OSSystemTimeNormalized 0.01874870633926259
# TYPE ClickHouseAsyncMetrics_FilesystemMainPathUsedINodes gauge
ClickHouseAsyncMetrics_FilesystemMainPathUsedINodes 1732551
# TYPE ClickHouseAsyncMetrics_OSProcessesCreated gauge
ClickHouseAsyncMetrics_OSProcessesCreated 7
# TYPE ClickHouseAsyncMetrics_OSContextSwitches gauge
ClickHouseAsyncMetrics_OSContextSwitches 32090
# TYPE ClickHouseAsyncMetrics_OSProcessesBlocked gauge
ClickHouseAsyncMetrics_OSProcessesBlocked 0
# TYPE ClickHouseAsyncMetrics_OSSystemTime gauge
ClickHouseAsyncMetrics_OSSystemTime 0.14998965071410073
# TYPE ClickHouseAsyncMetrics_BlockDiscardMerges_nvme0n1 gauge
ClickHouseAsyncMetrics_BlockDiscardMerges_nvme0n1 0
# TYPE ClickHouseAsyncMetrics_OSGuestNiceTimeCPU7 gauge
ClickHouseAsyncMetrics_OSGuestNiceTimeCPU7 0
# TYPE ClickHouseAsyncMetrics_OSMemoryTotal gauge
ClickHouseAsyncMetrics_OSMemoryTotal 25006321664
# TYPE ClickHouseAsyncMetrics_PrometheusThreads gauge
ClickHouseAsyncMetrics_PrometheusThreads 1
# TYPE ClickHouseAsyncMetrics_OSStealTimeCPU7 gauge
ClickHouseAsyncMetrics_OSStealTimeCPU7 0
# TYPE ClickHouseAsyncMetrics_OSSoftIrqTimeCPU7 gauge
ClickHouseAsyncMetrics_OSSoftIrqTimeCPU7 0
# TYPE ClickHouseAsyncMetrics_OSIOWaitTimeCPU7 gauge
ClickHouseAsyncMetrics_OSIOWaitTimeCPU7 0
# TYPE ClickHouseAsyncMetrics_OSSystemTimeCPU2 gauge
ClickHouseAsyncMetrics_OSSystemTimeCPU2 0.01999862009521343
# TYPE ClickHouseAsyncMetrics_OSIdleTimeCPU7 gauge
ClickHouseAsyncMetrics_OSIdleTimeCPU7 0.9399351444750311
# TYPE ClickHouseAsyncMetrics_OSNiceTimeCPU7 gauge
ClickHouseAsyncMetrics_OSNiceTimeCPU7 0
# TYPE ClickHouseAsyncMetrics_OSInterrupts gauge
ClickHouseAsyncMetrics_OSInterrupts 17158
# TYPE ClickHouseAsyncMetrics_OSUserTimeCPU7 gauge
ClickHouseAsyncMetrics_OSUserTimeCPU7 0.029997930142820144
# TYPE ClickHouseAsyncMetrics_BlockReadTime_nvme0n1 gauge
ClickHouseAsyncMetrics_BlockReadTime_nvme0n1 0
# TYPE ClickHouseAsyncMetrics_OSGuestTimeCPU6 gauge
ClickHouseAsyncMetrics_OSGuestTimeCPU6 0
# TYPE ClickHouseAsyncMetrics_BlockDiscardBytes_sda gauge
ClickHouseAsyncMetrics_BlockDiscardBytes_sda 0
# TYPE ClickHouseAsyncMetrics_OSIdleTimeCPU6 gauge
ClickHouseAsyncMetrics_OSIdleTimeCPU6 0.8799392841893909
# TYPE ClickHouseAsyncMetrics_jemalloc_arenas_all_pdirty gauge
ClickHouseAsyncMetrics_jemalloc_arenas_all_pdirty 7759
# TYPE ClickHouseAsyncMetrics_OSIdleTimeNormalized gauge
ClickHouseAsyncMetrics_OSIdleTimeNormalized 0.8924384217488993
# TYPE ClickHouseAsyncMetrics_OSStealTimeCPU6 gauge
ClickHouseAsyncMetrics_OSStealTimeCPU6 0
# TYPE ClickHouseAsyncMetrics_OSIrqTimeCPU7 gauge
ClickHouseAsyncMetrics_OSIrqTimeCPU7 0
# TYPE ClickHouseAsyncMetrics_OSIOWaitTimeCPU6 gauge
ClickHouseAsyncMetrics_OSIOWaitTimeCPU6 0
# TYPE ClickHouseAsyncMetrics_OSGuestTimeCPU7 gauge
ClickHouseAsyncMetrics_OSGuestTimeCPU7 0
# TYPE ClickHouseAsyncMetrics_OSSystemTimeCPU7 gauge
ClickHouseAsyncMetrics_OSSystemTimeCPU7 0.01999862009521343
# TYPE ClickHouseAsyncMetrics_OSGuestNiceTimeCPU4 gauge
ClickHouseAsyncMetrics_OSGuestNiceTimeCPU4 0
# TYPE ClickHouseAsyncMetrics_OSMemoryCached gauge
ClickHouseAsyncMetrics_OSMemoryCached 11901247488
# TYPE ClickHouseAsyncMetrics_OSUserTimeCPU6 gauge
ClickHouseAsyncMetrics_OSUserTimeCPU6 0.09999310047606715
# TYPE ClickHouseAsyncMetrics_OSGuestNiceTimeCPU5 gauge
ClickHouseAsyncMetrics_OSGuestNiceTimeCPU5 0
# TYPE ClickHouseAsyncMetrics_OSGuestNiceTimeCPU6 gauge
ClickHouseAsyncMetrics_OSGuestNiceTimeCPU6 0
# TYPE ClickHouseAsyncMetrics_OSGuestNiceTime gauge
ClickHouseAsyncMetrics_OSGuestNiceTime 0
# TYPE ClickHouseAsyncMetrics_OSSystemTimeCPU6 gauge
ClickHouseAsyncMetrics_OSSystemTimeCPU6 0.009999310047606715
# TYPE ClickHouseAsyncMetrics_MemoryDataAndStack gauge
ClickHouseAsyncMetrics_MemoryDataAndStack 1456848896
# TYPE ClickHouseAsyncMetrics_FilesystemMainPathTotalINodes gauge
ClickHouseAsyncMetrics_FilesystemMainPathTotalINodes 61014016
# TYPE ClickHouseAsyncMetrics_OSGuestTimeCPU3 gauge
ClickHouseAsyncMetrics_OSGuestTimeCPU3 0
# TYPE ClickHouseAsyncMetrics_OSUserTimeCPU5 gauge
ClickHouseAsyncMetrics_OSUserTimeCPU5 0.18998689090452758
# TYPE ClickHouseAsyncMetrics_OSStealTimeNormalized gauge
ClickHouseAsyncMetrics_OSStealTimeNormalized 0
# TYPE ClickHouseAsyncMetrics_OSGuestNiceTimeCPU2 gauge
ClickHouseAsyncMetrics_OSGuestNiceTimeCPU2 0
# TYPE ClickHouseAsyncMetrics_OSSoftIrqTimeCPU2 gauge
ClickHouseAsyncMetrics_OSSoftIrqTimeCPU2 0
# TYPE ClickHouseAsyncMetrics_OSSystemTimeCPU5 gauge
ClickHouseAsyncMetrics_OSSystemTimeCPU5 0.009999310047606715
# TYPE ClickHouseAsyncMetrics_OSSoftIrqTimeCPU4 gauge
ClickHouseAsyncMetrics_OSSoftIrqTimeCPU4 0
# TYPE ClickHouseAsyncMetrics_OSGuestNiceTimeCPU0 gauge
ClickHouseAsyncMetrics_OSGuestNiceTimeCPU0 0
# TYPE ClickHouseAsyncMetrics_OSIrqTime gauge
ClickHouseAsyncMetrics_OSIrqTime 0
# TYPE ClickHouseAsyncMetrics_OSStealTimeCPU1 gauge
ClickHouseAsyncMetrics_OSStealTimeCPU1 0
# TYPE ClickHouseAsyncMetrics_OSSoftIrqTimeCPU1 gauge
ClickHouseAsyncMetrics_OSSoftIrqTimeCPU1 0
# TYPE ClickHouseAsyncMetrics_MemoryShared gauge
ClickHouseAsyncMetrics_MemoryShared 240275456
# TYPE ClickHouseAsyncMetrics_OSIdleTimeCPU3 gauge
ClickHouseAsyncMetrics_OSIdleTimeCPU3 0.8899385942369976
# TYPE ClickHouseAsyncMetrics_TotalRowsOfMergeTreeTables gauge
ClickHouseAsyncMetrics_TotalRowsOfMergeTreeTables 7258
# TYPE ClickHouseAsyncMetrics_OSGuestTimeCPU4 gauge
ClickHouseAsyncMetrics_OSGuestTimeCPU4 0
# TYPE ClickHouseAsyncMetrics_BlockWriteMerges_nvme0n1 gauge
ClickHouseAsyncMetrics_BlockWriteMerges_nvme0n1 0
# TYPE ClickHouseAsyncMetrics_OSIrqTimeCPU4 gauge
ClickHouseAsyncMetrics_OSIrqTimeCPU4 0
# TYPE ClickHouseAsyncMetrics_UncompressedCacheBytes gauge
ClickHouseAsyncMetrics_UncompressedCacheBytes 0
# TYPE ClickHouseAsyncMetrics_Uptime gauge
ClickHouseAsyncMetrics_Uptime 31
# TYPE ClickHouseAsyncMetrics_BlockReadOps_nvme0n1 gauge
ClickHouseAsyncMetrics_BlockReadOps_nvme0n1 0
# TYPE ClickHouseAsyncMetrics_OSIdleTimeCPU4 gauge
ClickHouseAsyncMetrics_OSIdleTimeCPU4 0.9199365243798178
# TYPE ClickHouseAsyncMetrics_NetworkReceiveBytes_eth0 gauge
ClickHouseAsyncMetrics_NetworkReceiveBytes_eth0 0
# TYPE ClickHouseAsyncMetrics_MaxPartCountForPartition gauge
ClickHouseAsyncMetrics_MaxPartCountForPartition 4
# TYPE ClickHouseAsyncMetrics_OSUserTimeCPU4 gauge
ClickHouseAsyncMetrics_OSUserTimeCPU4 0.03999724019042686
# TYPE ClickHouseAsyncMetrics_OSNiceTimeCPU4 gauge
ClickHouseAsyncMetrics_OSNiceTimeCPU4 0
# TYPE ClickHouseAsyncMetrics_OSStealTimeCPU3 gauge
ClickHouseAsyncMetrics_OSStealTimeCPU3 0
# TYPE ClickHouseAsyncMetrics_OSIrqTimeCPU3 gauge
ClickHouseAsyncMetrics_OSIrqTimeCPU3 0
# TYPE ClickHouseAsyncMetrics_OSIOWaitTimeCPU3 gauge
ClickHouseAsyncMetrics_OSIOWaitTimeCPU3 0
# TYPE ClickHouseAsyncMetrics_NetworkSendBytes_eth0 gauge
ClickHouseAsyncMetrics_NetworkSendBytes_eth0 0
# TYPE ClickHouseAsyncMetrics_OSIdleTimeCPU5 gauge
ClickHouseAsyncMetrics_OSIdleTimeCPU5 0.769946873665717
# TYPE ClickHouseAsyncMetrics_OSUserTime gauge
ClickHouseAsyncMetrics_OSUserTime 0.6199572229516163
# TYPE ClickHouseAsyncMetrics_OSGuestTime gauge
ClickHouseAsyncMetrics_OSGuestTime 0
# TYPE ClickHouseAsyncMetrics_OSUserTimeCPU3 gauge
ClickHouseAsyncMetrics_OSUserTimeCPU3 0.07999448038085372
# TYPE ClickHouseAsyncMetrics_OSGuestTimeCPU2 gauge
ClickHouseAsyncMetrics_OSGuestTimeCPU2 0
# TYPE ClickHouseAsyncMetrics_OSUserTimeNormalized gauge
ClickHouseAsyncMetrics_OSUserTimeNormalized 0.07749465286895203
# TYPE ClickHouseAsyncMetrics_OSIrqTimeCPU2 gauge
ClickHouseAsyncMetrics_OSIrqTimeCPU2 0
# TYPE ClickHouseAsyncMetrics_Temperature4 gauge
ClickHouseAsyncMetrics_Temperature4 52
# TYPE ClickHouseAsyncMetrics_OSIOWaitTimeCPU2 gauge
ClickHouseAsyncMetrics_OSIOWaitTimeCPU2 0
# TYPE ClickHouseAsyncMetrics_MemoryCode gauge
ClickHouseAsyncMetrics_MemoryCode 212545536
# TYPE ClickHouseAsyncMetrics_OSStealTime gauge
ClickHouseAsyncMetrics_OSStealTime 0
# TYPE ClickHouseAsyncMetrics_OSIOWaitTimeNormalized gauge
ClickHouseAsyncMetrics_OSIOWaitTimeNormalized 0.0012499137559508393
# TYPE ClickHouseAsyncMetrics_OSIdleTimeCPU2 gauge
ClickHouseAsyncMetrics_OSIdleTimeCPU2 0.9399351444750311
# TYPE ClickHouseAsyncMetrics_OSProcessesRunning gauge
ClickHouseAsyncMetrics_OSProcessesRunning 2
# TYPE ClickHouseAsyncMetrics_OSNiceTimeCPU2 gauge
ClickHouseAsyncMetrics_OSNiceTimeCPU2 0
# TYPE ClickHouseAsyncMetrics_LoadAverage1 gauge
ClickHouseAsyncMetrics_LoadAverage1 0.62
# TYPE ClickHouseAsyncMetrics_OSStealTimeCPU5 gauge
ClickHouseAsyncMetrics_OSStealTimeCPU5 0
# TYPE ClickHouseAsyncMetrics_OSGuestNiceTimeCPU1 gauge
ClickHouseAsyncMetrics_OSGuestNiceTimeCPU1 0
# TYPE ClickHouseAsyncMetrics_jemalloc_retained gauge
ClickHouseAsyncMetrics_jemalloc_retained 37851136
# TYPE ClickHouseAsyncMetrics_OSGuestTimeCPU1 gauge
ClickHouseAsyncMetrics_OSGuestTimeCPU1 0
# TYPE ClickHouseAsyncMetrics_BlockWriteBytes_sda gauge
ClickHouseAsyncMetrics_BlockWriteBytes_sda 0
# TYPE ClickHouseAsyncMetrics_OSIdleTimeCPU0 gauge
ClickHouseAsyncMetrics_OSIdleTimeCPU0 0.9299358344274244
# TYPE ClickHouseAsyncMetrics_LoadAverage15 gauge
ClickHouseAsyncMetrics_LoadAverage15 1.03
# TYPE ClickHouseAsyncMetrics_OSIOWaitTime gauge
ClickHouseAsyncMetrics_OSIOWaitTime 0.009999310047606715
# TYPE ClickHouseAsyncMetrics_OSIOWaitTimeCPU1 gauge
ClickHouseAsyncMetrics_OSIOWaitTimeCPU1 0
# TYPE ClickHouseAsyncMetrics_OSIrqTimeCPU1 gauge
ClickHouseAsyncMetrics_OSIrqTimeCPU1 0
# TYPE ClickHouseAsyncMetrics_MMapCacheCells gauge
ClickHouseAsyncMetrics_MMapCacheCells 0
# TYPE ClickHouseAsyncMetrics_OSStealTimeCPU2 gauge
ClickHouseAsyncMetrics_OSStealTimeCPU2 0
# TYPE ClickHouseAsyncMetrics_OSSystemTimeCPU1 gauge
ClickHouseAsyncMetrics_OSSystemTimeCPU1 0.029997930142820144
# TYPE ClickHouseAsyncMetrics_Temperature_coretemp_Core_2 gauge
ClickHouseAsyncMetrics_Temperature_coretemp_Core_2 56
# TYPE ClickHouseAsyncMetrics_OSUserTimeCPU1 gauge
ClickHouseAsyncMetrics_OSUserTimeCPU1 0.08999379042846042
# TYPE ClickHouseAsyncMetrics_Temperature2 gauge
ClickHouseAsyncMetrics_Temperature2 20
# TYPE ClickHouseAsyncMetrics_OSSoftIrqTimeCPU3 gauge
ClickHouseAsyncMetrics_OSSoftIrqTimeCPU3 0
# TYPE ClickHouseAsyncMetrics_OSNiceTimeCPU3 gauge
ClickHouseAsyncMetrics_OSNiceTimeCPU3 0
# TYPE ClickHouseAsyncMetrics_OSSystemTimeCPU3 gauge
ClickHouseAsyncMetrics_OSSystemTimeCPU3 0.01999862009521343
# TYPE ClickHouseAsyncMetrics_MemoryVirtual gauge
ClickHouseAsyncMetrics_MemoryVirtual 2191757312
# TYPE ClickHouseAsyncMetrics_OSStealTimeCPU0 gauge
ClickHouseAsyncMetrics_OSStealTimeCPU0 0
# TYPE ClickHouseAsyncMetrics_FilesystemLogsPathAvailableINodes gauge
ClickHouseAsyncMetrics_FilesystemLogsPathAvailableINodes 59281465
# TYPE ClickHouseAsyncMetrics_OSGuestTimeNormalized gauge
ClickHouseAsyncMetrics_OSGuestTimeNormalized 0
# TYPE ClickHouseAsyncMetrics_OSIOWaitTimeCPU5 gauge
ClickHouseAsyncMetrics_OSIOWaitTimeCPU5 0
# TYPE ClickHouseAsyncMetrics_BlockReadBytes_sda gauge
ClickHouseAsyncMetrics_BlockReadBytes_sda 0
# TYPE ClickHouseAsyncMetrics_Temperature_coretemp_Core_3 gauge
ClickHouseAsyncMetrics_Temperature_coretemp_Core_3 49
# TYPE ClickHouseAsyncMetrics_OSNiceTimeCPU1 gauge
ClickHouseAsyncMetrics_OSNiceTimeCPU1 0
# TYPE ClickHouseAsyncMetrics_OSIOWaitTimeCPU0 gauge
ClickHouseAsyncMetrics_OSIOWaitTimeCPU0 0.009999310047606715
# TYPE ClickHouseAsyncMetrics_OSSoftIrqTimeCPU6 gauge
ClickHouseAsyncMetrics_OSSoftIrqTimeCPU6 0
# TYPE ClickHouseAsyncMetrics_jemalloc_background_thread_num_threads gauge
ClickHouseAsyncMetrics_jemalloc_background_thread_num_threads 0
# TYPE ClickHouseAsyncMetrics_OSSystemTimeCPU0 gauge
ClickHouseAsyncMetrics_OSSystemTimeCPU0 0.029997930142820144
# TYPE ClickHouseAsyncMetrics_Temperature_coretemp_Core_0 gauge
ClickHouseAsyncMetrics_Temperature_coretemp_Core_0 51
# TYPE ClickHouseAsyncMetrics_InterserverThreads gauge
ClickHouseAsyncMetrics_InterserverThreads 0
# TYPE ClickHouseAsyncMetrics_OSNiceTimeCPU0 gauge
ClickHouseAsyncMetrics_OSNiceTimeCPU0 0
# TYPE ClickHouseAsyncMetrics_OSUserTimeCPU0 gauge
ClickHouseAsyncMetrics_OSUserTimeCPU0 0.01999862009521343
# TYPE ClickHouseAsyncMetrics_OSSystemTimeCPU4 gauge
ClickHouseAsyncMetrics_OSSystemTimeCPU4 0.009999310047606715
# TYPE ClickHouseAsyncMetrics_MarkCacheFiles gauge
ClickHouseAsyncMetrics_MarkCacheFiles 0
# TYPE ClickHouseAsyncMetrics_OSIOWaitTimeCPU4 gauge
ClickHouseAsyncMetrics_OSIOWaitTimeCPU4 0
# TYPE ClickHouseAsyncMetrics_BlockReadMerges_sda gauge
ClickHouseAsyncMetrics_BlockReadMerges_sda 0
# TYPE ClickHouseAsyncMetrics_OSSoftIrqTime gauge
ClickHouseAsyncMetrics_OSSoftIrqTime 0.009999310047606715
# TYPE ClickHouseAsyncMetrics_CPUFrequencyMHz_0 gauge
ClickHouseAsyncMetrics_CPUFrequencyMHz_0 2075.311
# TYPE ClickHouseAsyncMetrics_CompiledExpressionCacheBytes gauge
ClickHouseAsyncMetrics_CompiledExpressionCacheBytes 0
# TYPE ClickHouseAsyncMetrics_OSIdleTime gauge
ClickHouseAsyncMetrics_OSIdleTime 7.1395073739911945
# TYPE ClickHouseAsyncMetrics_BlockWriteOps_nvme0n1 gauge
ClickHouseAsyncMetrics_BlockWriteOps_nvme0n1 0
# TYPE ClickHouseAsyncMetrics_BlockWriteBytes_nvme0n1 gauge
ClickHouseAsyncMetrics_BlockWriteBytes_nvme0n1 0
# TYPE ClickHouseAsyncMetrics_OSStealTimeCPU4 gauge
ClickHouseAsyncMetrics_OSStealTimeCPU4 0
# TYPE ClickHouseAsyncMetrics_OSMemoryAvailable gauge
ClickHouseAsyncMetrics_OSMemoryAvailable 17002708992
# TYPE ClickHouseAsyncMetrics_CompiledExpressionCacheCount gauge
ClickHouseAsyncMetrics_CompiledExpressionCacheCount 0
# TYPE ClickHouseAsyncMetrics_TotalBytesOfMergeTreeTables gauge
ClickHouseAsyncMetrics_TotalBytesOfMergeTreeTables 111153
# TYPE ClickHouseAsyncMetrics_OSNiceTimeCPU6 gauge
ClickHouseAsyncMetrics_OSNiceTimeCPU6 0
# TYPE ClickHouseAsyncMetrics_MarkCacheBytes gauge
ClickHouseAsyncMetrics_MarkCacheBytes 0
# TYPE ClickHouseAsyncMetrics_BlockDiscardTime_sda gauge
ClickHouseAsyncMetrics_BlockDiscardTime_sda 0
# TYPE ClickHouseAsyncMetrics_Temperature6 gauge
ClickHouseAsyncMetrics_Temperature6 54
# TYPE ClickHouseAsyncMetrics_OSIrqTimeCPU0 gauge
ClickHouseAsyncMetrics_OSIrqTimeCPU0 0
# TYPE ClickHouseAsyncMetrics_OSUptime gauge
ClickHouseAsyncMetrics_OSUptime 43414.7
# TYPE ClickHouseAsyncMetrics_BlockInFlightOps_nvme0n1 gauge
ClickHouseAsyncMetrics_BlockInFlightOps_nvme0n1 0
# TYPE ClickHouseAsyncMetrics_OSSoftIrqTimeCPU0 gauge
ClickHouseAsyncMetrics_OSSoftIrqTimeCPU0 0.009999310047606715
# TYPE ClickHouseAsyncMetrics_OSGuestTimeCPU5 gauge
ClickHouseAsyncMetrics_OSGuestTimeCPU5 0
# TYPE ClickHouseAsyncMetrics_OSMemoryFreePlusCached gauge
ClickHouseAsyncMetrics_OSMemoryFreePlusCached 16546840576
# TYPE ClickHouseAsyncMetrics_OSThreadsRunnable gauge
ClickHouseAsyncMetrics_OSThreadsRunnable 2
# TYPE ClickHouseAsyncMetrics_LoadAverage5 gauge
ClickHouseAsyncMetrics_LoadAverage5 0.94
# TYPE ClickHouseAsyncMetrics_OSUserTimeCPU2 gauge
ClickHouseAsyncMetrics_OSUserTimeCPU2 0.06999517033324701
# TYPE ClickHouseAsyncMetrics_OSIdleTimeCPU1 gauge
ClickHouseAsyncMetrics_OSIdleTimeCPU1 0.8799392841893909
# TYPE ClickHouseAsyncMetrics_Temperature_iwlwifi_1 gauge
ClickHouseAsyncMetrics_Temperature_iwlwifi_1 54
# TYPE ClickHouseAsyncMetrics_OSNiceTimeNormalized gauge
ClickHouseAsyncMetrics_OSNiceTimeNormalized 0
# TYPE ClickHouseAsyncMetrics_OSThreadsTotal gauge
ClickHouseAsyncMetrics_OSThreadsTotal 1718
# TYPE ClickHouseAsyncMetrics_OSGuestTimeCPU0 gauge
ClickHouseAsyncMetrics_OSGuestTimeCPU0 0
# TYPE ClickHouseAsyncMetrics_OSIrqTimeCPU6 gauge
ClickHouseAsyncMetrics_OSIrqTimeCPU6 0
# TYPE ClickHouseAsyncMetrics_jemalloc_background_thread_num_runs gauge
ClickHouseAsyncMetrics_jemalloc_background_thread_num_runs 0
# TYPE ClickHouseAsyncMetrics_BlockDiscardBytes_nvme0n1 gauge
ClickHouseAsyncMetrics_BlockDiscardBytes_nvme0n1 0
# TYPE ClickHouseAsyncMetrics_CPUFrequencyMHz_5 gauge
ClickHouseAsyncMetrics_CPUFrequencyMHz_5 2400
# TYPE ClickHouseAsyncMetrics_UncompressedCacheCells gauge
ClickHouseAsyncMetrics_UncompressedCacheCells 0
# TYPE ClickHouseAsyncMetrics_Jitter gauge
ClickHouseAsyncMetrics_Jitter 0.000069
# TYPE ClickHouseAsyncMetrics_MySQLThreads gauge
ClickHouseAsyncMetrics_MySQLThreads 0
# HELP ClickHouseStatusInfo_DictionaryStatus "Dictionary Status."
# TYPE ClickHouseStatusInfo_DictionaryStatus gauge
`

var want_v21_8_15_7 = []string{
	"ClickHouseAsyncMetrics,disk=default,host=1.2.3.4,instance=1.2.3.4:9363 DiskTotal",
	"ClickHouseAsyncMetrics,host=1.2.3.4,instance=1.2.3.4:9363 CompiledExpressionCacheBytes",
	"ClickHouseProfileEvents,host=1.2.3.4,instance=1.2.3.4:9363 PerfCpuCycles",
	"ClickHouseAsyncMetrics,host=1.2.3.4,instance=1.2.3.4:9363 OSContextSwitches",
	"ClickHouseAsyncMetrics,cpu=1,host=1.2.3.4,instance=1.2.3.4:9363 OSSoftIrqTimeCPU",
	"ClickHouseAsyncMetrics,host=1.2.3.4,instance=1.2.3.4:9363,unit=nvme0n1 BlockInFlightOps",
	"ClickHouseProfileEvents,host=1.2.3.4,instance=1.2.3.4:9363 S3WriteRequestsCount",
	"ClickHouseAsyncMetrics,eth=eth0,host=1.2.3.4,instance=1.2.3.4:9363 NetworkSendPackets",
	"ClickHouseAsyncMetrics,host=1.2.3.4,instance=1.2.3.4:9363 OSMemoryFreePlusCached",
	"ClickHouseProfileEvents,host=1.2.3.4,instance=1.2.3.4:9363 ZooKeeperRemove",
	"ClickHouseMetrics,host=1.2.3.4,instance=1.2.3.4:9363 ZooKeeperRequest",
	"ClickHouseAsyncMetrics,host=1.2.3.4,instance=1.2.3.4:9363,unit=sda BlockActiveTime",
	"ClickHouseAsyncMetrics,cpu=3,host=1.2.3.4,instance=1.2.3.4:9363 OSUserTimeCPU",
	"ClickHouseAsyncMetrics,host=1.2.3.4,instance=1.2.3.4:9363 MemoryVirtual",
	"ClickHouseProfileEvents,host=1.2.3.4,instance=1.2.3.4:9363 QueryMemoryLimitExceeded",
	"ClickHouseMetrics,host=1.2.3.4,instance=1.2.3.4:9363 Write",
	"ClickHouseAsyncMetrics,host=1.2.3.4,instance=1.2.3.4:9363,unit=nvme0n1 BlockDiscardOps",
	"ClickHouseAsyncMetrics,host=1.2.3.4,instance=1.2.3.4:9363,unit=sda BlockInFlightOps",
	"ClickHouseProfileEvents,host=1.2.3.4,instance=1.2.3.4:9363 ReadBufferFromFileDescriptorReadBytes",
	"ClickHouseProfileEvents,host=1.2.3.4,instance=1.2.3.4:9363 CreatedReadBufferAIO",
	"ClickHouseAsyncMetrics,host=1.2.3.4,instance=1.2.3.4:9363 jemalloc_arenas_all_dirty_purged",
	"ClickHouseAsyncMetrics,host=1.2.3.4,instance=1.2.3.4:9363,unit=pch_cannonlake Temperature",
	"ClickHouseAsyncMetrics,host=1.2.3.4,instance=1.2.3.4:9363 MaxPartCountForPartition",
	"ClickHouseAsyncMetrics,host=1.2.3.4,instance=1.2.3.4:9363 jemalloc_background_thread_num_threads",
	"ClickHouseAsyncMetrics,cpu=5,host=1.2.3.4,instance=1.2.3.4:9363 CPUFrequencyMHz",
	"ClickHouseAsyncMetrics,host=1.2.3.4,instance=1.2.3.4:9363 ReplicasMaxAbsoluteDelay",
	"ClickHouseAsyncMetrics,eth=eth0,host=1.2.3.4,instance=1.2.3.4:9363 NetworkSendBytes",
	"ClickHouseAsyncMetrics,cpu=1,host=1.2.3.4,instance=1.2.3.4:9363 OSNiceTimeCPU",
	"ClickHouseProfileEvents,host=1.2.3.4,instance=1.2.3.4:9363 WriteBufferAIOWrite",
	"ClickHouseMetrics,host=1.2.3.4,instance=1.2.3.4:9363 StorageBufferBytes",
	"ClickHouseProfileEvents,host=1.2.3.4,instance=1.2.3.4:9363 HedgedRequestsChangeReplica",
	"ClickHouseMetrics,host=1.2.3.4,instance=1.2.3.4:9363 PartsOutdated",
	"ClickHouseAsyncMetrics,host=1.2.3.4,instance=1.2.3.4:9363,unit=sda BlockReadTime",
	"ClickHouseAsyncMetrics,cpu=6,host=1.2.3.4,instance=1.2.3.4:9363 OSUserTimeCPU",
	"ClickHouseAsyncMetrics,host=1.2.3.4,instance=1.2.3.4:9363 MemoryCode",
	"ClickHouseProfileEvents,host=1.2.3.4,instance=1.2.3.4:9363 S3WriteRequestsThrottling",
	"ClickHouseAsyncMetrics,host=1.2.3.4,instance=1.2.3.4:9363 jemalloc_metadata",
	"ClickHouseAsyncMetrics,host=1.2.3.4,instance=1.2.3.4:9363 OSProcessesCreated",
	"ClickHouseAsyncMetrics,host=1.2.3.4,instance=1.2.3.4:9363 MMapCacheCells",
	"ClickHouseProfileEvents,host=1.2.3.4,instance=1.2.3.4:9363 ZooKeeperMulti",
	"ClickHouseProfileEvents,host=1.2.3.4,instance=1.2.3.4:9363 StorageBufferPassedRowsMaxThreshold",
	"ClickHouseMetrics,host=1.2.3.4,instance=1.2.3.4:9363 RWLockWaitingReaders",
	"ClickHouseAsyncMetrics,host=1.2.3.4,instance=1.2.3.4:9363 jemalloc_arenas_all_pmuzzy",
	"ClickHouseAsyncMetrics,host=1.2.3.4,instance=1.2.3.4:9363,unit=nvme0n1 BlockWriteBytes",
	"ClickHouseAsyncMetrics,host=1.2.3.4,instance=1.2.3.4:9363 OSThreadsRunnable",
	"ClickHouseProfileEvents,host=1.2.3.4,instance=1.2.3.4:9363 DictCacheKeysHit",
	"ClickHouseMetrics,host=1.2.3.4,instance=1.2.3.4:9363 NetworkSend",
	"ClickHouseAsyncMetrics,cpu=6,host=1.2.3.4,instance=1.2.3.4:9363 OSGuestNiceTimeCPU",
	"ClickHouseAsyncMetrics,host=1.2.3.4,instance=1.2.3.4:9363 OSGuestTime",
	"ClickHouseProfileEvents,host=1.2.3.4,instance=1.2.3.4:9363 WriteBufferAIOWriteBytes",
	"ClickHouseMetrics,host=1.2.3.4,instance=1.2.3.4:9363 DelayedInserts",
	"ClickHouseProfileEvents,host=1.2.3.4,instance=1.2.3.4:9363 ZooKeeperWatchResponse",
	"ClickHouseProfileEvents,host=1.2.3.4,instance=1.2.3.4:9363 ReadBackoff",
	"ClickHouseProfileEvents,host=1.2.3.4,instance=1.2.3.4:9363 StorageBufferFlush",
	"ClickHouseProfileEvents,host=1.2.3.4,instance=1.2.3.4:9363 RWLockReadersWaitMilliseconds",
	"ClickHouseAsyncMetrics,cpu=6,host=1.2.3.4,instance=1.2.3.4:9363 OSIdleTimeCPU",
	"ClickHouseAsyncMetrics,host=1.2.3.4,instance=1.2.3.4:9363,unit=nvme0n1 BlockWriteOps",
	"ClickHouseProfileEvents,host=1.2.3.4,instance=1.2.3.4:9363 ZooKeeperSet",
	"ClickHouseProfileEvents,host=1.2.3.4,instance=1.2.3.4:9363 ExternalAggregationUncompressedBytes",
	"ClickHouseProfileEvents,host=1.2.3.4,instance=1.2.3.4:9363 PerfCpuClock",
	"ClickHouseMetrics,host=1.2.3.4,instance=1.2.3.4:9363 CacheDictionaryUpdateQueueKeys",
	"ClickHouseAsyncMetrics,host=1.2.3.4,instance=1.2.3.4:9363,unit=nvme0n1 BlockDiscardTime",
	"ClickHouseAsyncMetrics,host=1.2.3.4,instance=1.2.3.4:9363,unit=sda BlockDiscardBytes",
	"ClickHouseAsyncMetrics,cpu=6,host=1.2.3.4,instance=1.2.3.4:9363 OSIrqTimeCPU",
	"ClickHouseProfileEvents,host=1.2.3.4,instance=1.2.3.4:9363 InsertedBytes",
	"ClickHouseProfileEvents,host=1.2.3.4,instance=1.2.3.4:9363 S3ReadBytes",
	"ClickHouseMetrics,host=1.2.3.4,instance=1.2.3.4:9363 PartsPreCommitted",
	"ClickHouseAsyncMetrics,cpu=7,host=1.2.3.4,instance=1.2.3.4:9363 CPUFrequencyMHz",
	"ClickHouseAsyncMetrics,cpu=5,host=1.2.3.4,instance=1.2.3.4:9363 OSIOWaitTimeCPU",
	"ClickHouseAsyncMetrics,host=1.2.3.4,instance=1.2.3.4:9363,unit=iwlwifi_1 Temperature",
	"ClickHouseProfileEvents,host=1.2.3.4,instance=1.2.3.4:9363 TableFunctionExecute",
	"ClickHouseProfileEvents,host=1.2.3.4,instance=1.2.3.4:9363 QueryMaskingRulesMatch",
	"ClickHouseMetrics,host=1.2.3.4,instance=1.2.3.4:9363 ReplicatedFetch",
	"ClickHouseAsyncMetrics,host=1.2.3.4,instance=1.2.3.4:9363 NumberOfTables",
	"ClickHouseAsyncMetrics,host=1.2.3.4,instance=1.2.3.4:9363 FilesystemLogsPathAvailableBytes",
	"ClickHouseAsyncMetrics,host=1.2.3.4,instance=1.2.3.4:9363 OSNiceTimeNormalized",
	"ClickHouseProfileEvents,host=1.2.3.4,instance=1.2.3.4:9363 ExternalAggregationWritePart",
	"ClickHouseProfileEvents,host=1.2.3.4,instance=1.2.3.4:9363 StorageBufferPassedRowsFlushThreshold",
	"ClickHouseAsyncMetrics,cpu=7,host=1.2.3.4,instance=1.2.3.4:9363 OSUserTimeCPU",
	"ClickHouseAsyncMetrics,host=1.2.3.4,instance=1.2.3.4:9363 FilesystemMainPathTotalINodes",
	"ClickHouseAsyncMetrics,cpu=1,host=1.2.3.4,instance=1.2.3.4:9363 OSUserTimeCPU",
	"ClickHouseAsyncMetrics,cpu=0,host=1.2.3.4,instance=1.2.3.4:9363 OSSystemTimeCPU",
	"ClickHouseProfileEvents,host=1.2.3.4,instance=1.2.3.4:9363 DistributedConnectionFailTry",
	"ClickHouseProfileEvents,host=1.2.3.4,instance=1.2.3.4:9363 OSWriteChars",
	"ClickHouseMetrics,host=1.2.3.4,instance=1.2.3.4:9363 DiskSpaceReservedForMerge",
	"ClickHouseAsyncMetrics,host=1.2.3.4,instance=1.2.3.4:9363 OSIOWaitTimeNormalized",
	"ClickHouseProfileEvents,host=1.2.3.4,instance=1.2.3.4:9363 ObsoleteReplicatedParts",
	"ClickHouseProfileEvents,host=1.2.3.4,instance=1.2.3.4:9363 DistributedRejectedInserts",
	"ClickHouseProfileEvents,host=1.2.3.4,instance=1.2.3.4:9363 S3ReadRequestsThrottling",
	"ClickHouseMetrics,host=1.2.3.4,instance=1.2.3.4:9363 QueryThread",
	"ClickHouseAsyncMetrics,cpu=4,host=1.2.3.4,instance=1.2.3.4:9363 OSIrqTimeCPU",
	"ClickHouseProfileEvents,host=1.2.3.4,instance=1.2.3.4:9363 DictCacheLockWriteNs",
	"ClickHouseProfileEvents,host=1.2.3.4,instance=1.2.3.4:9363 OSWriteBytes",
	"ClickHouseProfileEvents,host=1.2.3.4,instance=1.2.3.4:9363 PerfMinEnabledRunningTime",
	"ClickHouseProfileEvents,host=1.2.3.4,instance=1.2.3.4:9363 NotCreatedLogEntryForMutation",
	"ClickHouseMetrics,host=1.2.3.4,instance=1.2.3.4:9363 MMappedFiles",
	"ClickHouseProfileEvents,host=1.2.3.4,instance=1.2.3.4:9363 ReadBufferFromFileDescriptorRead",
	"ClickHouseAsyncMetrics,cpu=6,host=1.2.3.4,instance=1.2.3.4:9363 CPUFrequencyMHz",
	"ClickHouseAsyncMetrics,cpu=7,host=1.2.3.4,instance=1.2.3.4:9363 OSSoftIrqTimeCPU",
	"ClickHouseProfileEvents,host=1.2.3.4,instance=1.2.3.4:9363 SelectQueryTimeMicroseconds",
	"ClickHouseProfileEvents,host=1.2.3.4,instance=1.2.3.4:9363 DictCacheKeysRequested",
	"ClickHouseProfileEvents,host=1.2.3.4,instance=1.2.3.4:9363 PerfAlignmentFaults",
	"ClickHouseMetrics,host=1.2.3.4,instance=1.2.3.4:9363 MemoryTracking",
	"ClickHouseMetrics,host=1.2.3.4,instance=1.2.3.4:9363 PartsCommitted",
	"ClickHouseProfileEvents,host=1.2.3.4,instance=1.2.3.4:9363 FailedInsertQuery",
	"ClickHouseProfileEvents,host=1.2.3.4,instance=1.2.3.4:9363 UncompressedCacheHits",
	"ClickHouseProfileEvents,host=1.2.3.4,instance=1.2.3.4:9363 ExternalAggregationMerge",
	"ClickHouseMetrics,host=1.2.3.4,instance=1.2.3.4:9363 MySQLConnection",
	"ClickHouseAsyncMetrics,host=1.2.3.4,instance=1.2.3.4:9363,unit=5 Temperature",
	"ClickHouseAsyncMetrics,host=1.2.3.4,instance=1.2.3.4:9363 OSUserTimeNormalized",
	"ClickHouseProfileEvents,host=1.2.3.4,instance=1.2.3.4:9363 DataAfterMergeDiffersFromReplica",
	"ClickHouseMetrics,host=1.2.3.4,instance=1.2.3.4:9363 BrokenDistributedFilesToInsert",
	"ClickHouseAsyncMetrics,host=1.2.3.4,instance=1.2.3.4:9363 OSUptime",
	"ClickHouseMetrics,host=1.2.3.4,instance=1.2.3.4:9363 NetworkReceive",
	"ClickHouseAsyncMetrics,host=1.2.3.4,instance=1.2.3.4:9363,unit=sda BlockWriteBytes",
	"ClickHouseProfileEvents,host=1.2.3.4,instance=1.2.3.4:9363 MergeTreeDataWriterUncompressedBytes",
	"ClickHouseProfileEvents,host=1.2.3.4,instance=1.2.3.4:9363 MergeTreeDataProjectionWriterUncompressedBytes",
	"ClickHouseProfileEvents,host=1.2.3.4,instance=1.2.3.4:9363 PerfBranchInstructions",
	"ClickHouseAsyncMetrics,cpu=3,host=1.2.3.4,instance=1.2.3.4:9363 OSNiceTimeCPU",
	"ClickHouseProfileEvents,host=1.2.3.4,instance=1.2.3.4:9363 NetworkSendBytes",
	"ClickHouseProfileEvents,host=1.2.3.4,instance=1.2.3.4:9363 ReplicatedPartFetchesOfMerged",
	"ClickHouseProfileEvents,host=1.2.3.4,instance=1.2.3.4:9363 ReplicatedPartChecks",
	"ClickHouseAsyncMetrics,eth=eth0,host=1.2.3.4,instance=1.2.3.4:9363 NetworkSendErrors",
	"ClickHouseAsyncMetrics,cpu=2,host=1.2.3.4,instance=1.2.3.4:9363 OSGuestTimeCPU",
	"ClickHouseAsyncMetrics,host=1.2.3.4,instance=1.2.3.4:9363,unit=6 Temperature",
	"ClickHouseAsyncMetrics,host=1.2.3.4,instance=1.2.3.4:9363 MemoryDataAndStack",
	"ClickHouseProfileEvents,host=1.2.3.4,instance=1.2.3.4:9363 FileOpen",
	"ClickHouseProfileEvents,host=1.2.3.4,instance=1.2.3.4:9363 ReadBufferAIORead",
	"ClickHouseProfileEvents,host=1.2.3.4,instance=1.2.3.4:9363 UncompressedCacheWeightLost",
	"ClickHouseProfileEvents,host=1.2.3.4,instance=1.2.3.4:9363 OSIOWaitMicroseconds",
	"ClickHouseAsyncMetrics,cpu=2,host=1.2.3.4,instance=1.2.3.4:9363 CPUFrequencyMHz",
	"ClickHouseAsyncMetrics,host=1.2.3.4,instance=1.2.3.4:9363 PrometheusThreads",
	"ClickHouseMetrics,host=1.2.3.4,instance=1.2.3.4:9363 DistributedSend",
	"ClickHouseAsyncMetrics,cpu=3,host=1.2.3.4,instance=1.2.3.4:9363 OSIdleTimeCPU",
	"ClickHouseAsyncMetrics,cpu=2,host=1.2.3.4,instance=1.2.3.4:9363 OSStealTimeCPU",
	"ClickHouseProfileEvents,host=1.2.3.4,instance=1.2.3.4:9363 PerfContextSwitches",
	"ClickHouseProfileEvents,host=1.2.3.4,instance=1.2.3.4:9363 NotCreatedLogEntryForMerge",
	"ClickHouseAsyncMetrics,host=1.2.3.4,instance=1.2.3.4:9363 MemoryShared",
	"ClickHouseProfileEvents,host=1.2.3.4,instance=1.2.3.4:9363 InsertQuery",
	"ClickHouseProfileEvents,host=1.2.3.4,instance=1.2.3.4:9363 PerfBranchMisses",
	"ClickHouseProfileEvents,host=1.2.3.4,instance=1.2.3.4:9363 S3WriteRequestsErrors",
	"ClickHouseProfileEvents,host=1.2.3.4,instance=1.2.3.4:9363 CannotWriteToWriteBufferDiscard",
	"ClickHouseAsyncMetrics,host=1.2.3.4,instance=1.2.3.4:9363,unit=1 Temperature",
	"ClickHouseAsyncMetrics,eth=eth0,host=1.2.3.4,instance=1.2.3.4:9363 NetworkReceiveBytes",
	"ClickHouseAsyncMetrics,cpu=1,host=1.2.3.4,instance=1.2.3.4:9363 OSIrqTimeCPU",
	"ClickHouseProfileEvents,host=1.2.3.4,instance=1.2.3.4:9363 MarkCacheMisses",
	"ClickHouseAsyncMetrics,cpu=3,host=1.2.3.4,instance=1.2.3.4:9363 OSIrqTimeCPU",
	"ClickHouseAsyncMetrics,host=1.2.3.4,instance=1.2.3.4:9363 MarkCacheFiles",
	"ClickHouseAsyncMetrics,host=1.2.3.4,instance=1.2.3.4:9363 TotalBytesOfMergeTreeTables",
	"ClickHouseProfileEvents,host=1.2.3.4,instance=1.2.3.4:9363 SlowRead",
	"ClickHouseProfileEvents,host=1.2.3.4,instance=1.2.3.4:9363 PerfLocalMemoryMisses",
	"ClickHouseMetrics,host=1.2.3.4,instance=1.2.3.4:9363 BackgroundSchedulePoolTask",
	"ClickHouseAsyncMetrics,host=1.2.3.4,instance=1.2.3.4:9363 jemalloc_arenas_all_muzzy_purged",
	"ClickHouseProfileEvents,host=1.2.3.4,instance=1.2.3.4:9363 RegexpCreated",
	"ClickHouseAsyncMetrics,eth=eth0,host=1.2.3.4,instance=1.2.3.4:9363 NetworkReceiveDrop",
	"ClickHouseAsyncMetrics,cpu=7,host=1.2.3.4,instance=1.2.3.4:9363 OSIrqTimeCPU",
	"ClickHouseAsyncMetrics,cpu=6,host=1.2.3.4,instance=1.2.3.4:9363 OSSystemTimeCPU",
	"ClickHouseProfileEvents,host=1.2.3.4,instance=1.2.3.4:9363 Seek",
	"ClickHouseMetrics,host=1.2.3.4,instance=1.2.3.4:9363 InterserverConnection",
	"ClickHouseAsyncMetrics,host=1.2.3.4,instance=1.2.3.4:9363 OSMemoryBuffers",
	"ClickHouseAsyncMetrics,host=1.2.3.4,instance=1.2.3.4:9363 jemalloc_retained",
	"ClickHouseProfileEvents,host=1.2.3.4,instance=1.2.3.4:9363 ReadBufferAIOReadBytes",
	"ClickHouseAsyncMetrics,host=1.2.3.4,instance=1.2.3.4:9363,unit=nvme0n1 BlockReadOps",
	"ClickHouseProfileEvents,host=1.2.3.4,instance=1.2.3.4:9363 PolygonsAddedToPool",
	"ClickHouseProfileEvents,host=1.2.3.4,instance=1.2.3.4:9363 S3WriteRequestsRedirects",
	"ClickHouseAsyncMetrics,host=1.2.3.4,instance=1.2.3.4:9363 OSMemoryTotal",
	"ClickHouseAsyncMetrics,host=1.2.3.4,instance=1.2.3.4:9363 OSIdleTime",
	"ClickHouseProfileEvents,host=1.2.3.4,instance=1.2.3.4:9363 DataAfterMutationDiffersFromReplica",
	"ClickHouseMetrics,host=1.2.3.4,instance=1.2.3.4:9363 HTTPConnection",
	"ClickHouseAsyncMetrics,cpu=7,host=1.2.3.4,instance=1.2.3.4:9363 OSGuestTimeCPU",
	"ClickHouseProfileEvents,host=1.2.3.4,instance=1.2.3.4:9363 DistributedConnectionMissingTable",
	"ClickHouseAsyncMetrics,host=1.2.3.4,instance=1.2.3.4:9363 ReplicasMaxMergesInQueue",
	"ClickHouseAsyncMetrics,host=1.2.3.4,instance=1.2.3.4:9363 FilesystemLogsPathUsedBytes",
	"ClickHouseAsyncMetrics,host=1.2.3.4,instance=1.2.3.4:9363,unit=nvme0n1 BlockWriteTime",
	"ClickHouseAsyncMetrics,cpu=5,host=1.2.3.4,instance=1.2.3.4:9363 OSSoftIrqTimeCPU",
	"ClickHouseProfileEvents,host=1.2.3.4,instance=1.2.3.4:9363 CreatedReadBufferAIOFailed",
	"ClickHouseProfileEvents,host=1.2.3.4,instance=1.2.3.4:9363 MergedRows",
	"ClickHouseProfileEvents,host=1.2.3.4,instance=1.2.3.4:9363 StorageBufferPassedBytesMaxThreshold",
	"ClickHouseProfileEvents,host=1.2.3.4,instance=1.2.3.4:9363 SoftPageFaults",
	"ClickHouseMetrics,host=1.2.3.4,instance=1.2.3.4:9363 PartMutation",
	"ClickHouseProfileEvents,host=1.2.3.4,instance=1.2.3.4:9363 MergeTreeDataProjectionWriterRows",
	"ClickHouseProfileEvents,host=1.2.3.4,instance=1.2.3.4:9363 S3WriteBytes",
	"ClickHouseAsyncMetrics,host=1.2.3.4,instance=1.2.3.4:9363,unit=7 Temperature",
	"ClickHouseAsyncMetrics,host=1.2.3.4,instance=1.2.3.4:9363,unit=sda BlockReadMerges",
	"ClickHouseProfileEvents,host=1.2.3.4,instance=1.2.3.4:9363 DictCacheKeysExpired",
	"ClickHouseAsyncMetrics,host=1.2.3.4,instance=1.2.3.4:9363,unit=nvme_Composite Temperature",
	"ClickHouseAsyncMetrics,host=1.2.3.4,instance=1.2.3.4:9363 jemalloc_epoch",
	"ClickHouseProfileEvents,host=1.2.3.4,instance=1.2.3.4:9363 FailedQuery",
	"ClickHouseProfileEvents,host=1.2.3.4,instance=1.2.3.4:9363 ZooKeeperCheck",
	"ClickHouseProfileEvents,host=1.2.3.4,instance=1.2.3.4:9363 UserTimeMicroseconds",
	"ClickHouseMetrics,host=1.2.3.4,instance=1.2.3.4:9363 ReplicatedChecks",
	"ClickHouseAsyncMetrics,host=1.2.3.4,instance=1.2.3.4:9363 TotalPartsOfMergeTreeTables",
	"ClickHouseAsyncMetrics,cpu=1,host=1.2.3.4,instance=1.2.3.4:9363 CPUFrequencyMHz",
	"ClickHouseAsyncMetrics,cpu=0,host=1.2.3.4,instance=1.2.3.4:9363 OSIOWaitTimeCPU",
	"ClickHouseMetrics,host=1.2.3.4,instance=1.2.3.4:9363 ZooKeeperSession",
	"ClickHouseAsyncMetrics,cpu=4,host=1.2.3.4,instance=1.2.3.4:9363 OSGuestTimeCPU",
	"ClickHouseAsyncMetrics,cpu=5,host=1.2.3.4,instance=1.2.3.4:9363 OSIdleTimeCPU",
	"ClickHouseProfileEvents,host=1.2.3.4,instance=1.2.3.4:9363 StorageBufferPassedBytesFlushThreshold",
	"ClickHouseAsyncMetrics,host=1.2.3.4,instance=1.2.3.4:9363 AsynchronousMetricsCalculationTimeSpent",
	"ClickHouseAsyncMetrics,host=1.2.3.4,instance=1.2.3.4:9363 Uptime",
	"ClickHouseAsyncMetrics,cpu=4,host=1.2.3.4,instance=1.2.3.4:9363 OSUserTimeCPU",
	"ClickHouseProfileEvents,host=1.2.3.4,instance=1.2.3.4:9363 SelectedRanges",
	"ClickHouseProfileEvents,host=1.2.3.4,instance=1.2.3.4:9363 DictCacheRequests",
	"ClickHouseProfileEvents,host=1.2.3.4,instance=1.2.3.4:9363 RWLockAcquiredReadLocks",
	"ClickHouseAsyncMetrics,host=1.2.3.4,instance=1.2.3.4:9363 FilesystemMainPathTotalBytes",
	"ClickHouseAsyncMetrics,cpu=3,host=1.2.3.4,instance=1.2.3.4:9363 OSSoftIrqTimeCPU",
	"ClickHouseAsyncMetrics,cpu=6,host=1.2.3.4,instance=1.2.3.4:9363 OSSoftIrqTimeCPU",
	"ClickHouseMetrics,host=1.2.3.4,instance=1.2.3.4:9363 Merge",
	"ClickHouseAsyncMetrics,host=1.2.3.4,instance=1.2.3.4:9363 OSInterrupts",
	"ClickHouseProfileEvents,host=1.2.3.4,instance=1.2.3.4:9363 DiskWriteElapsedMicroseconds",
	"ClickHouseProfileEvents,host=1.2.3.4,instance=1.2.3.4:9363 DistributedConnectionStaleReplica",
	"ClickHouseProfileEvents,host=1.2.3.4,instance=1.2.3.4:9363 SelectedRows",
	"ClickHouseProfileEvents,host=1.2.3.4,instance=1.2.3.4:9363 MergeTreeDataProjectionWriterBlocksAlreadySorted",
	"ClickHouseProfileEvents,host=1.2.3.4,instance=1.2.3.4:9363 StorageBufferPassedAllMinThresholds",
	"ClickHouseProfileEvents,host=1.2.3.4,instance=1.2.3.4:9363 PerfTaskClock",
	"ClickHouseProfileEvents,host=1.2.3.4,instance=1.2.3.4:9363 ZooKeeperBytesReceived",
	"ClickHouseProfileEvents,host=1.2.3.4,instance=1.2.3.4:9363 PerfDataTLBReferences",
	"ClickHouseAsyncMetrics,host=1.2.3.4,instance=1.2.3.4:9363 FilesystemLogsPathTotalBytes",
	"ClickHouseAsyncMetrics,host=1.2.3.4,instance=1.2.3.4:9363 OSMemoryCached",
	"ClickHouseProfileEvents,host=1.2.3.4,instance=1.2.3.4:9363 ZooKeeperTransactions",
	"ClickHouseAsyncMetrics,cpu=3,host=1.2.3.4,instance=1.2.3.4:9363 OSSystemTimeCPU",
	"ClickHouseAsyncMetrics,host=1.2.3.4,instance=1.2.3.4:9363 OSGuestTimeNormalized",
	"ClickHouseProfileEvents,host=1.2.3.4,instance=1.2.3.4:9363 ReplicatedPartFetches",
	"ClickHouseProfileEvents,host=1.2.3.4,instance=1.2.3.4:9363 DelayedInsertsMilliseconds",
	"ClickHouseProfileEvents,host=1.2.3.4,instance=1.2.3.4:9363 MergedUncompressedBytes",
	"ClickHouseProfileEvents,host=1.2.3.4,instance=1.2.3.4:9363 DictCacheRequestTimeNs",
	"ClickHouseAsyncMetrics,host=1.2.3.4,instance=1.2.3.4:9363 OSMemoryAvailable",
	"ClickHouseProfileEvents,host=1.2.3.4,instance=1.2.3.4:9363 CreatedReadBufferOrdinary",
	"ClickHouseAsyncMetrics,host=1.2.3.4,instance=1.2.3.4:9363,unit=sda BlockReadOps",
	"ClickHouseProfileEvents,host=1.2.3.4,instance=1.2.3.4:9363 CannotRemoveEphemeralNode",
	"ClickHouseProfileEvents,host=1.2.3.4,instance=1.2.3.4:9363 QueryProfilerSignalOverruns",
	"ClickHouseAsyncMetrics,host=1.2.3.4,instance=1.2.3.4:9363 FilesystemLogsPathTotalINodes",
	"ClickHouseAsyncMetrics,cpu=0,host=1.2.3.4,instance=1.2.3.4:9363 CPUFrequencyMHz",
	"ClickHouseProfileEvents,host=1.2.3.4,instance=1.2.3.4:9363 ReplicatedPartMerges",
	"ClickHouseAsyncMetrics,cpu=3,host=1.2.3.4,instance=1.2.3.4:9363 OSGuestNiceTimeCPU",
	"ClickHouseAsyncMetrics,cpu=0,host=1.2.3.4,instance=1.2.3.4:9363 OSSoftIrqTimeCPU",
	"ClickHouseProfileEvents,host=1.2.3.4,instance=1.2.3.4:9363 NetworkReceiveBytes",
	"ClickHouseProfileEvents,host=1.2.3.4,instance=1.2.3.4:9363 MergeTreeDataWriterBlocksAlreadySorted",
	"ClickHouseAsyncMetrics,cpu=4,host=1.2.3.4,instance=1.2.3.4:9363 OSSoftIrqTimeCPU",
	"ClickHouseAsyncMetrics,host=1.2.3.4,instance=1.2.3.4:9363 CompiledExpressionCacheCount",
	"ClickHouseProfileEvents,host=1.2.3.4,instance=1.2.3.4:9363 DiskReadElapsedMicroseconds",
	"ClickHouseAsyncMetrics,host=1.2.3.4,instance=1.2.3.4:9363,unit=nvme_Sensor_2 Temperature",
	"ClickHouseAsyncMetrics,host=1.2.3.4,instance=1.2.3.4:9363,unit=nvme0n1 BlockDiscardMerges",
	"ClickHouseProfileEvents,host=1.2.3.4,instance=1.2.3.4:9363 CompileExpressionsMicroseconds",
	"ClickHouseMetrics,host=1.2.3.4,instance=1.2.3.4:9363 Read",
	"ClickHouseAsyncMetrics,host=1.2.3.4,instance=1.2.3.4:9363,unit=0 Temperature",
	"ClickHouseAsyncMetrics,host=1.2.3.4,instance=1.2.3.4:9363,unit=sda BlockDiscardMerges",
	"ClickHouseAsyncMetrics,cpu=4,host=1.2.3.4,instance=1.2.3.4:9363 OSSystemTimeCPU",
	"ClickHouseProfileEvents,host=1.2.3.4,instance=1.2.3.4:9363 ZooKeeperHardwareExceptions",
	"ClickHouseAsyncMetrics,cpu=6,host=1.2.3.4,instance=1.2.3.4:9363 OSStealTimeCPU",
	"ClickHouseProfileEvents,host=1.2.3.4,instance=1.2.3.4:9363 MergesTimeMilliseconds",
	"ClickHouseMetrics,host=1.2.3.4,instance=1.2.3.4:9363 ZooKeeperWatch",
	"ClickHouseAsyncMetrics,host=1.2.3.4,instance=1.2.3.4:9363 TCPThreads",
	"ClickHouseAsyncMetrics,host=1.2.3.4,instance=1.2.3.4:9363,unit=acpitz Temperature",
	"ClickHouseAsyncMetrics,cpu=1,host=1.2.3.4,instance=1.2.3.4:9363 OSIOWaitTimeCPU",
	"ClickHouseAsyncMetrics,host=1.2.3.4,instance=1.2.3.4:9363,unit=2 Temperature",
	"ClickHouseProfileEvents,host=1.2.3.4,instance=1.2.3.4:9363 PerfCacheReferences",
	"ClickHouseMetrics,host=1.2.3.4,instance=1.2.3.4:9363 MaxDDLEntryID",
	"ClickHouseAsyncMetrics,cpu=5,host=1.2.3.4,instance=1.2.3.4:9363 OSNiceTimeCPU",
	"ClickHouseProfileEvents,host=1.2.3.4,instance=1.2.3.4:9363 IOBufferAllocs",
	"ClickHouseProfileEvents,host=1.2.3.4,instance=1.2.3.4:9363 CompiledFunctionExecute",
	"ClickHouseProfileEvents,host=1.2.3.4,instance=1.2.3.4:9363 OSCPUVirtualTimeMicroseconds",
	"ClickHouseAsyncMetrics,cpu=2,host=1.2.3.4,instance=1.2.3.4:9363 OSSystemTimeCPU",
	"ClickHouseAsyncMetrics,cpu=2,host=1.2.3.4,instance=1.2.3.4:9363 OSIOWaitTimeCPU",
	"ClickHouseAsyncMetrics,host=1.2.3.4,instance=1.2.3.4:9363 MySQLThreads",
	"ClickHouseProfileEvents,host=1.2.3.4,instance=1.2.3.4:9363 DuplicatedInsertedBlocks",
	"ClickHouseProfileEvents,host=1.2.3.4,instance=1.2.3.4:9363 DictCacheKeysRequestedFound",
	"ClickHouseAsyncMetrics,host=1.2.3.4,instance=1.2.3.4:9363 HTTPThreads",
	"ClickHouseAsyncMetrics,host=1.2.3.4,instance=1.2.3.4:9363 ReplicasMaxRelativeDelay",
	"ClickHouseAsyncMetrics,host=1.2.3.4,instance=1.2.3.4:9363 FilesystemMainPathUsedBytes",
	"ClickHouseAsyncMetrics,host=1.2.3.4,instance=1.2.3.4:9363 Jitter",
	"ClickHouseProfileEvents,host=1.2.3.4,instance=1.2.3.4:9363 ReplicatedPartChecksFailed",
	"ClickHouseProfileEvents,host=1.2.3.4,instance=1.2.3.4:9363 PolygonsInPoolAllocatedBytes",
	"ClickHouseAsyncMetrics,host=1.2.3.4,instance=1.2.3.4:9363 ReplicasSumQueueSize",
	"ClickHouseAsyncMetrics,cpu=3,host=1.2.3.4,instance=1.2.3.4:9363 OSStealTimeCPU",
	"ClickHouseAsyncMetrics,host=1.2.3.4,instance=1.2.3.4:9363 OSMemoryFreeWithoutCached",
	"ClickHouseAsyncMetrics,cpu=4,host=1.2.3.4,instance=1.2.3.4:9363 OSIOWaitTimeCPU",
	"ClickHouseProfileEvents,host=1.2.3.4,instance=1.2.3.4:9363 StorageBufferErrorOnFlush",
	"ClickHouseAsyncMetrics,cpu=0,host=1.2.3.4,instance=1.2.3.4:9363 OSNiceTimeCPU",
	"ClickHouseProfileEvents,host=1.2.3.4,instance=1.2.3.4:9363 DistributedConnectionFailAtAll",
	"ClickHouseProfileEvents,host=1.2.3.4,instance=1.2.3.4:9363 MergeTreeDataWriterRows",
	"ClickHouseAsyncMetrics,host=1.2.3.4,instance=1.2.3.4:9363 ReplicasMaxInsertsInQueue",
	"ClickHouseAsyncMetrics,host=1.2.3.4,instance=1.2.3.4:9363,unit=nvme0n1 BlockWriteMerges",
	"ClickHouseProfileEvents,host=1.2.3.4,instance=1.2.3.4:9363 DictCacheKeysNotFound",
	"ClickHouseMetrics,host=1.2.3.4,instance=1.2.3.4:9363 GlobalThreadActive",
	"ClickHouseMetrics,host=1.2.3.4,instance=1.2.3.4:9363 DistributedFilesToInsert",
	"ClickHouseAsyncMetrics,host=1.2.3.4,instance=1.2.3.4:9363 OSIdleTimeNormalized",
	"ClickHouseAsyncMetrics,cpu=5,host=1.2.3.4,instance=1.2.3.4:9363 OSGuestTimeCPU",
	"ClickHouseProfileEvents,host=1.2.3.4,instance=1.2.3.4:9363 DistributedDelayedInsertsMilliseconds",
	"ClickHouseProfileEvents,host=1.2.3.4,instance=1.2.3.4:9363 CompileFunction",
	"ClickHouseAsyncMetrics,host=1.2.3.4,instance=1.2.3.4:9363 jemalloc_mapped",
	"ClickHouseAsyncMetrics,cpu=5,host=1.2.3.4,instance=1.2.3.4:9363 OSUserTimeCPU",
	"ClickHouseAsyncMetrics,cpu=1,host=1.2.3.4,instance=1.2.3.4:9363 OSStealTimeCPU",
	"ClickHouseAsyncMetrics,host=1.2.3.4,instance=1.2.3.4:9363,unit=sda BlockReadBytes",
	"ClickHouseProfileEvents,host=1.2.3.4,instance=1.2.3.4:9363 InsertQueryTimeMicroseconds",
	"ClickHouseProfileEvents,host=1.2.3.4,instance=1.2.3.4:9363 ContextLock",
	"ClickHouseProfileEvents,host=1.2.3.4,instance=1.2.3.4:9363 PerfCacheMisses",
	"ClickHouseAsyncMetrics,cpu=7,host=1.2.3.4,instance=1.2.3.4:9363 OSNiceTimeCPU",
	"ClickHouseAsyncMetrics,host=1.2.3.4,instance=1.2.3.4:9363 OSThreadsTotal",
	"ClickHouseProfileEvents,host=1.2.3.4,instance=1.2.3.4:9363 ExternalSortMerge",
	"ClickHouseProfileEvents,host=1.2.3.4,instance=1.2.3.4:9363 CreatedLogEntryForMerge",
	"ClickHouseAsyncMetrics,cpu=4,host=1.2.3.4,instance=1.2.3.4:9363 OSStealTimeCPU",
	"ClickHouseAsyncMetrics,host=1.2.3.4,instance=1.2.3.4:9363 MarkCacheBytes",
	"ClickHouseProfileEvents,host=1.2.3.4,instance=1.2.3.4:9363 SelectQuery",
	"ClickHouseProfileEvents,host=1.2.3.4,instance=1.2.3.4:9363 WriteBufferFromFileDescriptorWrite",
	"ClickHouseMetrics,host=1.2.3.4,instance=1.2.3.4:9363 Revision",
	"ClickHouseProfileEvents,host=1.2.3.4,instance=1.2.3.4:9363 QueryTimeMicroseconds",
	"ClickHouseAsyncMetrics,host=1.2.3.4,instance=1.2.3.4:9363 OSSoftIrqTime",
	"ClickHouseAsyncMetrics,host=1.2.3.4,instance=1.2.3.4:9363,unit=nvme0n1 BlockDiscardBytes",
	"ClickHouseProfileEvents,host=1.2.3.4,instance=1.2.3.4:9363 FunctionExecute",
	"ClickHouseProfileEvents,host=1.2.3.4,instance=1.2.3.4:9363 ZooKeeperCreate",
	"ClickHouseProfileEvents,host=1.2.3.4,instance=1.2.3.4:9363 VoluntaryContextSwitches",
	"ClickHouseAsyncMetrics,host=1.2.3.4,instance=1.2.3.4:9363,unit=nvme0n1 BlockReadMerges",
	"ClickHouseAsyncMetrics,host=1.2.3.4,instance=1.2.3.4:9363 OSOpenFiles",
	"ClickHouseAsyncMetrics,cpu=5,host=1.2.3.4,instance=1.2.3.4:9363 OSStealTimeCPU",
	"ClickHouseProfileEvents,host=1.2.3.4,instance=1.2.3.4:9363 ReadBufferFromFileDescriptorReadFailed",
	"ClickHouseProfileEvents,host=1.2.3.4,instance=1.2.3.4:9363 MergeTreeDataWriterBlocks",
	"ClickHouseMetrics,host=1.2.3.4,instance=1.2.3.4:9363 PostgreSQLConnection",
	"ClickHouseAsyncMetrics,host=1.2.3.4,instance=1.2.3.4:9363 FilesystemMainPathUsedINodes",
	"ClickHouseProfileEvents,host=1.2.3.4,instance=1.2.3.4:9363 Query",
	"ClickHouseProfileEvents,host=1.2.3.4,instance=1.2.3.4:9363 ReadCompressedBytes",
	"ClickHouseProfileEvents,host=1.2.3.4,instance=1.2.3.4:9363 IOBufferAllocBytes",
	"ClickHouseProfileEvents,host=1.2.3.4,instance=1.2.3.4:9363 ZooKeeperOtherExceptions",
	"ClickHouseProfileEvents,host=1.2.3.4,instance=1.2.3.4:9363 ZooKeeperWaitMicroseconds",
	"ClickHouseProfileEvents,host=1.2.3.4,instance=1.2.3.4:9363 WriteBufferFromFileDescriptorWriteFailed",
	"ClickHouseMetrics,host=1.2.3.4,instance=1.2.3.4:9363 DictCacheRequests",
	"ClickHouseProfileEvents,host=1.2.3.4,instance=1.2.3.4:9363 PerfInstructionTLBReferences",
	"ClickHouseMetrics,host=1.2.3.4,instance=1.2.3.4:9363 QueryPreempted",
	"ClickHouseAsyncMetrics,host=1.2.3.4,instance=1.2.3.4:9363 OSNiceTime",
	"ClickHouseAsyncMetrics,cpu=5,host=1.2.3.4,instance=1.2.3.4:9363 OSSystemTimeCPU",
	"ClickHouseAsyncMetrics,host=1.2.3.4,instance=1.2.3.4:9363 InterserverThreads",
	"ClickHouseAsyncMetrics,host=1.2.3.4,instance=1.2.3.4:9363,unit=nvme0n1 BlockQueueTime",
	"ClickHouseAsyncMetrics,cpu=3,host=1.2.3.4,instance=1.2.3.4:9363 CPUFrequencyMHz",
	"ClickHouseProfileEvents,host=1.2.3.4,instance=1.2.3.4:9363 SelectedBytes",
	"ClickHouseProfileEvents,host=1.2.3.4,instance=1.2.3.4:9363 StorageBufferPassedTimeMaxThreshold",
	"ClickHouseMetrics,host=1.2.3.4,instance=1.2.3.4:9363 ContextLockWait",
	"ClickHouseMetrics,host=1.2.3.4,instance=1.2.3.4:9363 StorageBufferRows",
	"ClickHouseAsyncMetrics,host=1.2.3.4,instance=1.2.3.4:9363,unit=sda BlockWriteTime",
	"ClickHouseAsyncMetrics,host=1.2.3.4,instance=1.2.3.4:9363,unit=sda BlockDiscardOps",
	"ClickHouseAsyncMetrics,cpu=7,host=1.2.3.4,instance=1.2.3.4:9363 OSGuestNiceTimeCPU",
	"ClickHouseAsyncMetrics,host=1.2.3.4,instance=1.2.3.4:9363 jemalloc_background_thread_num_runs",
	"ClickHouseProfileEvents,host=1.2.3.4,instance=1.2.3.4:9363 WriteBufferFromFileDescriptorWriteBytes",
	"ClickHouseProfileEvents,host=1.2.3.4,instance=1.2.3.4:9363 ZooKeeperGet",
	"ClickHouseProfileEvents,host=1.2.3.4,instance=1.2.3.4:9363 StorageBufferPassedTimeFlushThreshold",
	"ClickHouseMetrics,host=1.2.3.4,instance=1.2.3.4:9363 BackgroundPoolTask",
	"ClickHouseMetrics,host=1.2.3.4,instance=1.2.3.4:9363 BackgroundBufferFlushSchedulePoolTask",
	"ClickHouseAsyncMetrics,cpu=6,host=1.2.3.4,instance=1.2.3.4:9363 OSNiceTimeCPU",
	"ClickHouseAsyncMetrics,disk=default,host=1.2.3.4,instance=1.2.3.4:9363 DiskUsed",
	"ClickHouseAsyncMetrics,host=1.2.3.4,instance=1.2.3.4:9363 OSStealTimeNormalized",
	"ClickHouseAsyncMetrics,host=1.2.3.4,instance=1.2.3.4:9363 OSUserTime",
	"ClickHouseAsyncMetrics,host=1.2.3.4,instance=1.2.3.4:9363,unit=4 Temperature",
	"ClickHouseProfileEvents,host=1.2.3.4,instance=1.2.3.4:9363 ZooKeeperInit",
	"ClickHouseProfileEvents,host=1.2.3.4,instance=1.2.3.4:9363 DNSError",
	"ClickHouseAsyncMetrics,cpu=0,host=1.2.3.4,instance=1.2.3.4:9363 OSUserTimeCPU",
	"ClickHouseProfileEvents,host=1.2.3.4,instance=1.2.3.4:9363 RejectedInserts",
	"ClickHouseProfileEvents,host=1.2.3.4,instance=1.2.3.4:9363 OSReadChars",
	"ClickHouseProfileEvents,host=1.2.3.4,instance=1.2.3.4:9363 S3WriteMicroseconds",
	"ClickHouseMetrics,host=1.2.3.4,instance=1.2.3.4:9363 OpenFileForWrite",
	"ClickHouseAsyncMetrics,host=1.2.3.4,instance=1.2.3.4:9363 FilesystemMainPathAvailableINodes",
	"ClickHouseAsyncMetrics,host=1.2.3.4,instance=1.2.3.4:9363 jemalloc_active",
	"ClickHouseProfileEvents,host=1.2.3.4,instance=1.2.3.4:9363 PerfMinEnabledTime",
	"ClickHouseProfileEvents,host=1.2.3.4,instance=1.2.3.4:9363 PerfLocalMemoryReferences",
	"ClickHouseProfileEvents,host=1.2.3.4,instance=1.2.3.4:9363 DelayedInserts",
	"ClickHouseAsyncMetrics,host=1.2.3.4,instance=1.2.3.4:9363,unit=sda BlockWriteOps",
	"ClickHouseAsyncMetrics,host=1.2.3.4,instance=1.2.3.4:9363,unit=sda BlockDiscardTime",
	"ClickHouseAsyncMetrics,cpu=2,host=1.2.3.4,instance=1.2.3.4:9363 OSGuestNiceTimeCPU",
	"ClickHouseAsyncMetrics,cpu=1,host=1.2.3.4,instance=1.2.3.4:9363 OSGuestNiceTimeCPU",
	"ClickHouseProfileEvents,host=1.2.3.4,instance=1.2.3.4:9363 ReplicatedPartFailedFetches",
	"ClickHouseProfileEvents,host=1.2.3.4,instance=1.2.3.4:9363 RealTimeMicroseconds",
	"ClickHouseProfileEvents,host=1.2.3.4,instance=1.2.3.4:9363 CreatedLogEntryForMutation",
	"ClickHouseMetrics,host=1.2.3.4,instance=1.2.3.4:9363 PartsTemporary",
	"ClickHouseAsyncMetrics,eth=eth0,host=1.2.3.4,instance=1.2.3.4:9363 NetworkReceiveErrors",
	"ClickHouseAsyncMetrics,cpu=4,host=1.2.3.4,instance=1.2.3.4:9363 OSGuestNiceTimeCPU",
	"ClickHouseAsyncMetrics,host=1.2.3.4,instance=1.2.3.4:9363,unit=coretemp_Core_0 Temperature",
	"ClickHouseProfileEvents,host=1.2.3.4,instance=1.2.3.4:9363 MMappedFileCacheMisses",
	"ClickHouseProfileEvents,host=1.2.3.4,instance=1.2.3.4:9363 S3ReadRequestsRedirects",
	"ClickHouseMetrics,host=1.2.3.4,instance=1.2.3.4:9363 VersionInteger",
	"ClickHouseMetrics,host=1.2.3.4,instance=1.2.3.4:9363 LocalThread",
	"ClickHouseAsyncMetrics,cpu=7,host=1.2.3.4,instance=1.2.3.4:9363 OSIdleTimeCPU",
	"ClickHouseProfileEvents,host=1.2.3.4,instance=1.2.3.4:9363 ReplicatedPartMutations",
	"ClickHouseProfileEvents,host=1.2.3.4,instance=1.2.3.4:9363 DistributedSyncInsertionTimeoutExceeded",
	"ClickHouseAsyncMetrics,host=1.2.3.4,instance=1.2.3.4:9363 jemalloc_metadata_thp",
	"ClickHouseAsyncMetrics,host=1.2.3.4,instance=1.2.3.4:9363 MemoryResident",
	"ClickHouseAsyncMetrics,host=1.2.3.4,instance=1.2.3.4:9363 OSSoftIrqTimeNormalized",
	"ClickHouseProfileEvents,host=1.2.3.4,instance=1.2.3.4:9363 ArenaAllocChunks",
	"ClickHouseProfileEvents,host=1.2.3.4,instance=1.2.3.4:9363 OSReadBytes",
	"ClickHouseAsyncMetrics,host=1.2.3.4,instance=1.2.3.4:9363 FilesystemMainPathAvailableBytes",
	"ClickHouseAsyncMetrics,cpu=0,host=1.2.3.4,instance=1.2.3.4:9363 OSIrqTimeCPU",
	"ClickHouseProfileEvents,host=1.2.3.4,instance=1.2.3.4:9363 PerfInstructionTLBMisses",
	"ClickHouseAsyncMetrics,host=1.2.3.4,instance=1.2.3.4:9363 PostgreSQLThreads",
	"ClickHouseAsyncMetrics,host=1.2.3.4,instance=1.2.3.4:9363,unit=15 LoadAverage",
	"ClickHouseProfileEvents,host=1.2.3.4,instance=1.2.3.4:9363 PerfBusCycles",
	"ClickHouseMetrics,host=1.2.3.4,instance=1.2.3.4:9363 RWLockActiveWriters",
	"ClickHouseMetrics,host=1.2.3.4,instance=1.2.3.4:9363 BackgroundFetchesPoolTask",
	"ClickHouseAsyncMetrics,host=1.2.3.4,instance=1.2.3.4:9363 ReplicasSumMergesInQueue",
	"ClickHouseAsyncMetrics,eth=eth0,host=1.2.3.4,instance=1.2.3.4:9363 NetworkReceivePackets",
	"ClickHouseProfileEvents,host=1.2.3.4,instance=1.2.3.4:9363 MMappedFileCacheHits",
	"ClickHouseMetrics,host=1.2.3.4,instance=1.2.3.4:9363 PartsInMemory",
	"ClickHouseAsyncMetrics,host=1.2.3.4,instance=1.2.3.4:9363,unit=nvme0n1 BlockReadTime",
	"ClickHouseProfileEvents,host=1.2.3.4,instance=1.2.3.4:9363 S3ReadMicroseconds",
	"ClickHouseAsyncMetrics,cpu=5,host=1.2.3.4,instance=1.2.3.4:9363 OSIrqTimeCPU",
	"ClickHouseAsyncMetrics,cpu=3,host=1.2.3.4,instance=1.2.3.4:9363 OSIOWaitTimeCPU",
	"ClickHouseProfileEvents,host=1.2.3.4,instance=1.2.3.4:9363 NetworkSendElapsedMicroseconds",
	"ClickHouseProfileEvents,host=1.2.3.4,instance=1.2.3.4:9363 CompileExpressionsBytes",
	"ClickHouseMetrics,host=1.2.3.4,instance=1.2.3.4:9363 Query",
	"ClickHouseAsyncMetrics,host=1.2.3.4,instance=1.2.3.4:9363,unit=coretemp_Core_1 Temperature",
	"ClickHouseProfileEvents,host=1.2.3.4,instance=1.2.3.4:9363 ArenaAllocBytes",
	"ClickHouseProfileEvents,host=1.2.3.4,instance=1.2.3.4:9363 ReplicatedDataLoss",
	"ClickHouseProfileEvents,host=1.2.3.4,instance=1.2.3.4:9363 PerfStalledCyclesBackend",
	"ClickHouseAsyncMetrics,cpu=4,host=1.2.3.4,instance=1.2.3.4:9363 CPUFrequencyMHz",
	"ClickHouseAsyncMetrics,cpu=2,host=1.2.3.4,instance=1.2.3.4:9363 OSIdleTimeCPU",
	"ClickHouseAsyncMetrics,host=1.2.3.4,instance=1.2.3.4:9363 ReplicasMaxQueueSize",
	"ClickHouseAsyncMetrics,host=1.2.3.4,instance=1.2.3.4:9363 jemalloc_arenas_all_pactive",
	"ClickHouseProfileEvents,host=1.2.3.4,instance=1.2.3.4:9363 SelectedMarks",
	"ClickHouseProfileEvents,host=1.2.3.4,instance=1.2.3.4:9363 PerfCpuMigrations",
	"ClickHouseMetrics,host=1.2.3.4,instance=1.2.3.4:9363 BackgroundMovePoolTask",
	"ClickHouseMetrics,host=1.2.3.4,instance=1.2.3.4:9363 BackgroundDistributedSchedulePoolTask",
	"ClickHouseMetrics,host=1.2.3.4,instance=1.2.3.4:9363 GlobalThread",
	"ClickHouseAsyncMetrics,host=1.2.3.4,instance=1.2.3.4:9363 NumberOfDatabases",
	"ClickHouseAsyncMetrics,host=1.2.3.4,instance=1.2.3.4:9363 jemalloc_resident",
	"ClickHouseAsyncMetrics,host=1.2.3.4,instance=1.2.3.4:9363,unit=3 Temperature",
	"ClickHouseAsyncMetrics,cpu=1,host=1.2.3.4,instance=1.2.3.4:9363 OSGuestTimeCPU",
	"ClickHouseAsyncMetrics,host=1.2.3.4,instance=1.2.3.4:9363,unit=sda BlockWriteMerges",
	"ClickHouseAsyncMetrics,host=1.2.3.4,instance=1.2.3.4:9363 OSSystemTime",
	"ClickHouseProfileEvents,host=1.2.3.4,instance=1.2.3.4:9363 UncompressedCacheMisses",
	"ClickHouseProfileEvents,host=1.2.3.4,instance=1.2.3.4:9363 MarkCacheHits",
	"ClickHouseMetrics,host=1.2.3.4,instance=1.2.3.4:9363 CacheDictionaryUpdateQueueBatches",
	"ClickHouseMetrics,host=1.2.3.4,instance=1.2.3.4:9363 ReadonlyReplica",
	"ClickHouseMetrics,host=1.2.3.4,instance=1.2.3.4:9363 PartsCompact",
	"ClickHouseAsyncMetrics,host=1.2.3.4,instance=1.2.3.4:9363,unit=sda BlockQueueTime",
	"ClickHouseProfileEvents,host=1.2.3.4,instance=1.2.3.4:9363 ExternalAggregationCompressedBytes",
	"ClickHouseProfileEvents,host=1.2.3.4,instance=1.2.3.4:9363 PerfDataTLBMisses",
	"ClickHouseProfileEvents,host=1.2.3.4,instance=1.2.3.4:9363 S3ReadRequestsCount",
	"ClickHouseAsyncMetrics,cpu=0,host=1.2.3.4,instance=1.2.3.4:9363 OSGuestTimeCPU",
	"ClickHouseProfileEvents,host=1.2.3.4,instance=1.2.3.4:9363 CompressedReadBufferBlocks",
	"ClickHouseProfileEvents,host=1.2.3.4,instance=1.2.3.4:9363 HardPageFaults",
	"ClickHouseMetrics,host=1.2.3.4,instance=1.2.3.4:9363 PartsDeleting",
	"ClickHouseMetrics,host=1.2.3.4,instance=1.2.3.4:9363 PartsDeleteOnDestroy",
	"ClickHouseAsyncMetrics,cpu=2,host=1.2.3.4,instance=1.2.3.4:9363 OSIrqTimeCPU",
	"ClickHouseMetrics,host=1.2.3.4,instance=1.2.3.4:9363 SendScalars",
	"ClickHouseMetrics,host=1.2.3.4,instance=1.2.3.4:9363 MMappedFileBytes",
	"ClickHouseAsyncMetrics,host=1.2.3.4,instance=1.2.3.4:9363,unit=nvme_Sensor_1 Temperature",
	"ClickHouseAsyncMetrics,host=1.2.3.4,instance=1.2.3.4:9363 OSIrqTimeNormalized",
	"ClickHouseAsyncMetrics,cpu=5,host=1.2.3.4,instance=1.2.3.4:9363 OSGuestNiceTimeCPU",
	"ClickHouseProfileEvents,host=1.2.3.4,instance=1.2.3.4:9363 ReplicaPartialShutdown",
	"ClickHouseProfileEvents,host=1.2.3.4,instance=1.2.3.4:9363 PerfStalledCyclesFrontend",
	"ClickHouseMetrics,host=1.2.3.4,instance=1.2.3.4:9363 TablesToDropQueueSize",
	"ClickHouseProfileEvents,host=1.2.3.4,instance=1.2.3.4:9363 ExternalSortWritePart",
	"ClickHouseMetrics,host=1.2.3.4,instance=1.2.3.4:9363 EphemeralNode",
	"ClickHouseMetrics,host=1.2.3.4,instance=1.2.3.4:9363 RWLockActiveReaders",
	"ClickHouseAsyncMetrics,host=1.2.3.4,instance=1.2.3.4:9363 OSGuestNiceTime",
	"ClickHouseAsyncMetrics,cpu=3,host=1.2.3.4,instance=1.2.3.4:9363 OSGuestTimeCPU",
	"ClickHouseAsyncMetrics,host=1.2.3.4,instance=1.2.3.4:9363 FilesystemLogsPathAvailableINodes",
	"ClickHouseProfileEvents,host=1.2.3.4,instance=1.2.3.4:9363 NetworkReceiveElapsedMicroseconds",
	"ClickHouseAsyncMetrics,host=1.2.3.4,instance=1.2.3.4:9363 ReplicasSumInsertsInQueue",
	"ClickHouseAsyncMetrics,host=1.2.3.4,instance=1.2.3.4:9363 OSIrqTime",
	"ClickHouseAsyncMetrics,host=1.2.3.4,instance=1.2.3.4:9363 OSIOWaitTime",
	"ClickHouseProfileEvents,host=1.2.3.4,instance=1.2.3.4:9363 ZooKeeperClose",
	"ClickHouseMetrics,host=1.2.3.4,instance=1.2.3.4:9363 ReplicatedSend",
	"ClickHouseMetrics,host=1.2.3.4,instance=1.2.3.4:9363 SendExternalTables",
	"ClickHouseAsyncMetrics,disk=default,host=1.2.3.4,instance=1.2.3.4:9363 DiskAvailable",
	"ClickHouseAsyncMetrics,host=1.2.3.4,instance=1.2.3.4:9363 OSGuestNiceTimeNormalized",
	"ClickHouseAsyncMetrics,cpu=2,host=1.2.3.4,instance=1.2.3.4:9363 OSSoftIrqTimeCPU",
	"ClickHouseProfileEvents,host=1.2.3.4,instance=1.2.3.4:9363 ZooKeeperExists",
	"ClickHouseAsyncMetrics,cpu=0,host=1.2.3.4,instance=1.2.3.4:9363 OSGuestNiceTimeCPU",
	"ClickHouseAsyncMetrics,host=1.2.3.4,instance=1.2.3.4:9363 UncompressedCacheBytes",
	"ClickHouseAsyncMetrics,cpu=4,host=1.2.3.4,instance=1.2.3.4:9363 OSIdleTimeCPU",
	"ClickHouseProfileEvents,host=1.2.3.4,instance=1.2.3.4:9363 SelectedParts",
	"ClickHouseProfileEvents,host=1.2.3.4,instance=1.2.3.4:9363 PerfInstructions",
	"ClickHouseMetrics,host=1.2.3.4,instance=1.2.3.4:9363 RWLockWaitingWriters",
	"ClickHouseAsyncMetrics,cpu=7,host=1.2.3.4,instance=1.2.3.4:9363 OSIOWaitTimeCPU",
	"ClickHouseAsyncMetrics,cpu=2,host=1.2.3.4,instance=1.2.3.4:9363 OSNiceTimeCPU",
	"ClickHouseMetrics,host=1.2.3.4,instance=1.2.3.4:9363 PartsWide",
	"ClickHouseAsyncMetrics,host=1.2.3.4,instance=1.2.3.4:9363,unit=nvme0n1 BlockReadBytes",
	"ClickHouseAsyncMetrics,host=1.2.3.4,instance=1.2.3.4:9363 OSProcessesBlocked",
	"ClickHouseAsyncMetrics,cpu=7,host=1.2.3.4,instance=1.2.3.4:9363 OSStealTimeCPU",
	"ClickHouseAsyncMetrics,host=1.2.3.4,instance=1.2.3.4:9363 jemalloc_arenas_all_pdirty",
	"ClickHouseAsyncMetrics,cpu=1,host=1.2.3.4,instance=1.2.3.4:9363 OSSystemTimeCPU",
	"ClickHouseAsyncMetrics,cpu=0,host=1.2.3.4,instance=1.2.3.4:9363 OSStealTimeCPU",
	"ClickHouseProfileEvents,host=1.2.3.4,instance=1.2.3.4:9363 CompressedReadBufferBytes",
	"ClickHouseProfileEvents,host=1.2.3.4,instance=1.2.3.4:9363 DictCacheKeysRequestedMiss",
	"ClickHouseProfileEvents,host=1.2.3.4,instance=1.2.3.4:9363 RWLockWritersWaitMilliseconds",
	"ClickHouseProfileEvents,host=1.2.3.4,instance=1.2.3.4:9363 OSCPUWaitMicroseconds",
	"ClickHouseAsyncMetrics,cpu=6,host=1.2.3.4,instance=1.2.3.4:9363 OSGuestTimeCPU",
	"ClickHouseAsyncMetrics,host=1.2.3.4,instance=1.2.3.4:9363 OSStealTime",
	"ClickHouseProfileEvents,host=1.2.3.4,instance=1.2.3.4:9363 CreatedReadBufferMMapFailed",
	"ClickHouseProfileEvents,host=1.2.3.4,instance=1.2.3.4:9363 ZooKeeperBytesSent",
	"ClickHouseAsyncMetrics,host=1.2.3.4,instance=1.2.3.4:9363 OSSystemTimeNormalized",
	"ClickHouseAsyncMetrics,cpu=6,host=1.2.3.4,instance=1.2.3.4:9363 OSIOWaitTimeCPU",
	"ClickHouseAsyncMetrics,disk=default,host=1.2.3.4,instance=1.2.3.4:9363 DiskUnreserved",
	"ClickHouseAsyncMetrics,cpu=4,host=1.2.3.4,instance=1.2.3.4:9363 OSNiceTimeCPU",
	"ClickHouseAsyncMetrics,host=1.2.3.4,instance=1.2.3.4:9363 OSProcessesRunning",
	"ClickHouseAsyncMetrics,host=1.2.3.4,instance=1.2.3.4:9363,unit=1 LoadAverage",
	"ClickHouseAsyncMetrics,host=1.2.3.4,instance=1.2.3.4:9363,unit=coretemp_Core_2 Temperature",
	"ClickHouseProfileEvents,host=1.2.3.4,instance=1.2.3.4:9363 MergeTreeDataWriterCompressedBytes",
	"ClickHouseProfileEvents,host=1.2.3.4,instance=1.2.3.4:9363 ThrottlerSleepMicroseconds",
	"ClickHouseProfileEvents,host=1.2.3.4,instance=1.2.3.4:9363 StorageBufferLayerLockWritersWaitMilliseconds",
	"ClickHouseProfileEvents,host=1.2.3.4,instance=1.2.3.4:9363 S3ReadRequestsErrors",
	"ClickHouseAsyncMetrics,cpu=0,host=1.2.3.4,instance=1.2.3.4:9363 OSIdleTimeCPU",
	"ClickHouseAsyncMetrics,cpu=2,host=1.2.3.4,instance=1.2.3.4:9363 OSUserTimeCPU",
	"ClickHouseProfileEvents,host=1.2.3.4,instance=1.2.3.4:9363 MergeTreeDataProjectionWriterBlocks",
	"ClickHouseProfileEvents,host=1.2.3.4,instance=1.2.3.4:9363 StorageBufferLayerLockReadersWaitMilliseconds",
	"ClickHouseMetrics,host=1.2.3.4,instance=1.2.3.4:9363 LocalThreadActive",
	"ClickHouseAsyncMetrics,host=1.2.3.4,instance=1.2.3.4:9363,unit=coretemp_Core_3 Temperature",
	"ClickHouseProfileEvents,host=1.2.3.4,instance=1.2.3.4:9363 FailedSelectQuery",
	"ClickHouseProfileEvents,host=1.2.3.4,instance=1.2.3.4:9363 ZooKeeperList",
	"ClickHouseProfileEvents,host=1.2.3.4,instance=1.2.3.4:9363 RWLockAcquiredWriteLocks",
	"ClickHouseProfileEvents,host=1.2.3.4,instance=1.2.3.4:9363 InvoluntaryContextSwitches",
	"ClickHouseMetrics,host=1.2.3.4,instance=1.2.3.4:9363 BackgroundMessageBrokerSchedulePoolTask",
	"ClickHouseProfileEvents,host=1.2.3.4,instance=1.2.3.4:9363 MergeTreeDataProjectionWriterCompressedBytes",
	"ClickHouseAsyncMetrics,cpu=1,host=1.2.3.4,instance=1.2.3.4:9363 OSIdleTimeCPU",
	"ClickHouseAsyncMetrics,host=1.2.3.4,instance=1.2.3.4:9363 UncompressedCacheCells",
	"ClickHouseProfileEvents,host=1.2.3.4,instance=1.2.3.4:9363 SystemTimeMicroseconds",
	"ClickHouseProfileEvents,host=1.2.3.4,instance=1.2.3.4:9363 PerfRefCpuCycles",
	"ClickHouseAsyncMetrics,eth=eth0,host=1.2.3.4,instance=1.2.3.4:9363 NetworkSendDrop",
	"ClickHouseAsyncMetrics,host=1.2.3.4,instance=1.2.3.4:9363,unit=5 LoadAverage",
	"ClickHouseProfileEvents,host=1.2.3.4,instance=1.2.3.4:9363 CreatedReadBufferMMap",
	"ClickHouseProfileEvents,host=1.2.3.4,instance=1.2.3.4:9363 CreatedHTTPConnections",
	"ClickHouseAsyncMetrics,host=1.2.3.4,instance=1.2.3.4:9363 jemalloc_allocated",
	"ClickHouseAsyncMetrics,host=1.2.3.4,instance=1.2.3.4:9363 FilesystemLogsPathUsedINodes",
	"ClickHouseAsyncMetrics,host=1.2.3.4,instance=1.2.3.4:9363,unit=nvme0n1 BlockActiveTime",
	"ClickHouseProfileEvents,host=1.2.3.4,instance=1.2.3.4:9363 Merge",
	"ClickHouseAsyncMetrics,host=1.2.3.4,instance=1.2.3.4:9363,unit=coretemp_Package_id_0 Temperature",
	"ClickHouseAsyncMetrics,cpu=7,host=1.2.3.4,instance=1.2.3.4:9363 OSSystemTimeCPU",
	"ClickHouseAsyncMetrics,host=1.2.3.4,instance=1.2.3.4:9363 TotalRowsOfMergeTreeTables",
	"ClickHouseProfileEvents,host=1.2.3.4,instance=1.2.3.4:9363 DistributedDelayedInserts",
	"ClickHouseProfileEvents,host=1.2.3.4,instance=1.2.3.4:9363 DictCacheLockReadNs",
	"ClickHouseProfileEvents,host=1.2.3.4,instance=1.2.3.4:9363 PerfEmulationFaults",
	"ClickHouseProfileEvents,host=1.2.3.4,instance=1.2.3.4:9363 InsertedRows",
	"ClickHouseProfileEvents,host=1.2.3.4,instance=1.2.3.4:9363 ZooKeeperUserExceptions",
	"ClickHouseMetrics,host=1.2.3.4,instance=1.2.3.4:9363 TCPConnection",
	"ClickHouseMetrics,host=1.2.3.4,instance=1.2.3.4:9363 OpenFileForRead",
	"ClickHouseAsyncMetrics,host=1.2.3.4,instance=1.2.3.4:9363 jemalloc_background_thread_run_intervals",
	"ClickHouseAsyncMetrics,host=1.2.3.4,instance=1.2.3.4:9363,unit=average BlockActiveTime",
	"ClickHouseAsyncMetrics,host=1.2.3.4,instance=1.2.3.4:9363,unit=total BlockDiscardBytes",
	"ClickHouseAsyncMetrics,host=1.2.3.4,instance=1.2.3.4:9363,unit=total BlockDiscardMerges",
	"ClickHouseAsyncMetrics,host=1.2.3.4,instance=1.2.3.4:9363,unit=total BlockDiscardOps",
	"ClickHouseAsyncMetrics,host=1.2.3.4,instance=1.2.3.4:9363,unit=average BlockDiscardTime",
	"ClickHouseAsyncMetrics,host=1.2.3.4,instance=1.2.3.4:9363,unit=total BlockInFlightOps",
	"ClickHouseAsyncMetrics,host=1.2.3.4,instance=1.2.3.4:9363,unit=average BlockQueueTime",
	"ClickHouseAsyncMetrics,host=1.2.3.4,instance=1.2.3.4:9363,unit=total BlockReadBytes",
	"ClickHouseAsyncMetrics,host=1.2.3.4,instance=1.2.3.4:9363,unit=total BlockReadMerges",
	"ClickHouseAsyncMetrics,host=1.2.3.4,instance=1.2.3.4:9363,unit=total BlockReadOps",
	"ClickHouseAsyncMetrics,host=1.2.3.4,instance=1.2.3.4:9363,unit=average BlockReadTime",
	"ClickHouseAsyncMetrics,host=1.2.3.4,instance=1.2.3.4:9363,unit=total BlockWriteBytes",
	"ClickHouseAsyncMetrics,host=1.2.3.4,instance=1.2.3.4:9363,unit=total BlockWriteMerges",
	"ClickHouseAsyncMetrics,host=1.2.3.4,instance=1.2.3.4:9363,unit=total BlockWriteOps",
	"ClickHouseAsyncMetrics,host=1.2.3.4,instance=1.2.3.4:9363,unit=average BlockWriteTime",
	"ClickHouseAsyncMetrics,cpu=average,host=1.2.3.4,instance=1.2.3.4:9363 CPUFrequencyMHz",
	"ClickHouseAsyncMetrics,disk=total,host=1.2.3.4,instance=1.2.3.4:9363 DiskAvailable",
	"ClickHouseAsyncMetrics,disk=total,host=1.2.3.4,instance=1.2.3.4:9363 DiskTotal",
	"ClickHouseAsyncMetrics,disk=total,host=1.2.3.4,instance=1.2.3.4:9363 DiskUnreserved",
	"ClickHouseAsyncMetrics,disk=total,host=1.2.3.4,instance=1.2.3.4:9363 DiskUsed",
	"ClickHouseAsyncMetrics,host=1.2.3.4,instance=1.2.3.4:9363,unit=average LoadAverage",
	"ClickHouseAsyncMetrics,eth=total,host=1.2.3.4,instance=1.2.3.4:9363 NetworkReceiveBytes",
	"ClickHouseAsyncMetrics,eth=total,host=1.2.3.4,instance=1.2.3.4:9363 NetworkReceiveDrop",
	"ClickHouseAsyncMetrics,eth=total,host=1.2.3.4,instance=1.2.3.4:9363 NetworkReceiveErrors",
	"ClickHouseAsyncMetrics,eth=total,host=1.2.3.4,instance=1.2.3.4:9363 NetworkReceivePackets",
	"ClickHouseAsyncMetrics,eth=total,host=1.2.3.4,instance=1.2.3.4:9363 NetworkSendBytes",
	"ClickHouseAsyncMetrics,eth=total,host=1.2.3.4,instance=1.2.3.4:9363 NetworkSendDrop",
	"ClickHouseAsyncMetrics,eth=total,host=1.2.3.4,instance=1.2.3.4:9363 NetworkSendErrors",
	"ClickHouseAsyncMetrics,eth=total,host=1.2.3.4,instance=1.2.3.4:9363 NetworkSendPackets",
	"ClickHouseAsyncMetrics,cpu=average,host=1.2.3.4,instance=1.2.3.4:9363 OSGuestNiceTimeCPU",
	"ClickHouseAsyncMetrics,cpu=average,host=1.2.3.4,instance=1.2.3.4:9363 OSGuestTimeCPU",
	"ClickHouseAsyncMetrics,cpu=average,host=1.2.3.4,instance=1.2.3.4:9363 OSIOWaitTimeCPU",
	"ClickHouseAsyncMetrics,cpu=average,host=1.2.3.4,instance=1.2.3.4:9363 OSIdleTimeCPU",
	"ClickHouseAsyncMetrics,cpu=average,host=1.2.3.4,instance=1.2.3.4:9363 OSIrqTimeCPU",
	"ClickHouseAsyncMetrics,cpu=average,host=1.2.3.4,instance=1.2.3.4:9363 OSNiceTimeCPU",
	"ClickHouseAsyncMetrics,cpu=average,host=1.2.3.4,instance=1.2.3.4:9363 OSSoftIrqTimeCPU",
	"ClickHouseAsyncMetrics,cpu=average,host=1.2.3.4,instance=1.2.3.4:9363 OSStealTimeCPU",
	"ClickHouseAsyncMetrics,cpu=average,host=1.2.3.4,instance=1.2.3.4:9363 OSSystemTimeCPU",
	"ClickHouseAsyncMetrics,cpu=average,host=1.2.3.4,instance=1.2.3.4:9363 OSUserTimeCPU",
	"ClickHouseAsyncMetrics,host=1.2.3.4,instance=1.2.3.4:9363,unit=average Temperature",
}

var mockCfg = `
source = "clickhouse"
uds_path = ""
ignore_req_err = false
measurement_prefix = ""
tls_open = false
disable_host_tag = false
disable_instance_tag = false
disable_info_tag = false

[[measurements]]
prefix = "ClickHouseProfileEvents_"
name = "ClickHouseProfileEvents"

[[measurements]]
prefix = "ClickHouseMetrics_"
name = "ClickHouseMetrics"

[[measurements]]
prefix = "ClickHouseAsyncMetrics_"
name = "ClickHouseAsyncMetrics"

[[measurements]]
prefix = "ClickHouseStatusInfo_"
name = "ClickHouseStatusInfo"`

type mockTagger struct{}

func (t *mockTagger) HostTags() map[string]string {
	return map[string]string{
		"host": "me",
	}
}

func (t *mockTagger) ElectionTags() map[string]string {
	return nil
}
