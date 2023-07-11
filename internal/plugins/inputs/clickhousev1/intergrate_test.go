// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package clickhousev1

import (
	"fmt"
	"net"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/BurntSushi/toml"
	"github.com/GuanceCloud/cliutils/point"
	dt "github.com/ory/dockertest/v3"
	"github.com/ory/dockertest/v3/docker"
	"github.com/stretchr/testify/require"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/testutils"
)

// ATTENTION: Docker version should use v20.10.18 in integrate tests. Other versions are not tested.

func TestIntegrate(t *testing.T) {
	if !testutils.CheckIntegrationTestingRunning() {
		t.Skip()
	}

	testutils.PurgeRemoteByName(inputName)       // purge at first.
	defer testutils.PurgeRemoteByName(inputName) // purge at last.

	start := time.Now()
	cases, err := buildCases(t)
	if err != nil {
		cr := &testutils.CaseResult{
			Name:          t.Name(),
			Status:        testutils.TestPassed,
			FailedMessage: err.Error(),
			Cost:          time.Since(start),
		}

		_ = testutils.Flush(cr)
		return
	}

	t.Logf("testing %d cases...", len(cases))

	for _, tc := range cases {
		func(tc *caseSpec) {
			t.Run(tc.name, func(t *testing.T) {
				t.Parallel()
				caseStart := time.Now()

				t.Logf("testing %s...", tc.name)

				if err := testutils.RetryTestRun(tc.run); err != nil {
					tc.cr.Status = testutils.TestFailed
					tc.cr.FailedMessage = err.Error()

					panic(err)
				} else {
					tc.cr.Status = testutils.TestPassed
				}

				tc.cr.Cost = time.Since(caseStart)

				require.NoError(t, testutils.Flush(tc.cr))

				t.Cleanup(func() {
					// clean remote docker resources
					if tc.resource == nil {
						return
					}

					tc.pool.Purge(tc.resource)
				})
			})
		}(tc)
	}
}

func getConfAccessPoint(host, port string) string {
	return fmt.Sprintf("http://%s/metrics", net.JoinHostPort(host, port))
}

func buildCases(t *testing.T) ([]*caseSpec, error) {
	t.Helper()

	remote := testutils.GetRemote()

	bases := []struct {
		name         string // Also used as build image name:tag.
		conf         string
		exposedPorts []string
		mPathCount   map[string]int

		optsProfileEvents []inputs.PointCheckOption
		optsMetrics       []inputs.PointCheckOption
		optsAsyncMetrics  []inputs.PointCheckOption
		optsStatusInfo    []inputs.PointCheckOption
	}{
		{
			name: "pubrepo.jiagouyun.com/image-repo-for-testing/clickhouse/clickhouse-server:22.8.15.23",
			// selfBuild: true,
			conf: `
source = "clickhouse"
interval = "10s"
tls_open = false
urls = [""]
[tags]
  tag1 = "some_value"
  tag2 = "some_other_value"`, // set conf URL later.
			exposedPorts: []string{"9363/tcp"},
			mPathCount:   map[string]int{"/": 10},
			optsProfileEvents: []inputs.PointCheckOption{
				inputs.WithTypeChecking(false),
				inputs.WithExtraTags(map[string]string{"tag1": "", "tag2": ""}),
				inputs.WithOptionalFields("Query", "SelectQuery", "InsertQuery", "AsyncInsertQuery", "AsyncInsertBytes", "AsyncInsertCacheHits", "FailedQuery", "FailedSelectQuery", "FailedInsertQuery", "FailedAsyncInsertQuery", "QueryTimeMicroseconds", "SelectQueryTimeMicroseconds", "InsertQueryTimeMicroseconds", "OtherQueryTimeMicroseconds", "FileOpen", "Seek", "ReadBufferFromFileDescriptorRead", "ReadBufferFromFileDescriptorReadFailed", "ReadBufferFromFileDescriptorReadBytes", "WriteBufferFromFileDescriptorWrite", "WriteBufferFromFileDescriptorWriteFailed", "WriteBufferFromFileDescriptorWriteBytes", "FileSync", "DirectorySync", "FileSyncElapsedMicroseconds", "DirectorySyncElapsedMicroseconds", "ReadCompressedBytes", "CompressedReadBufferBlocks", "CompressedReadBufferBytes", "UncompressedCacheHits", "UncompressedCacheMisses", "UncompressedCacheWeightLost", "MMappedFileCacheHits", "MMappedFileCacheMisses", "OpenedFileCacheHits", "OpenedFileCacheMisses", "AIOWrite", "AIOWriteBytes", "AIORead", "AIOReadBytes", "IOBufferAllocs", "IOBufferAllocBytes", "ArenaAllocChunks", "ArenaAllocBytes", "FunctionExecute", "TableFunctionExecute", "MarkCacheHits", "MarkCacheMisses", "QueryCacheHits", "QueryCacheMisses", "CreatedReadBufferOrdinary", "CreatedReadBufferDirectIO", "CreatedReadBufferDirectIOFailed", "CreatedReadBufferMMap", "CreatedReadBufferMMapFailed", "DiskReadElapsedMicroseconds", "DiskWriteElapsedMicroseconds", "NetworkReceiveElapsedMicroseconds", "NetworkSendElapsedMicroseconds", "NetworkReceiveBytes", "NetworkSendBytes", "DiskS3GetRequestThrottlerCount", "DiskS3GetRequestThrottlerSleepMicroseconds", "DiskS3PutRequestThrottlerCount", "DiskS3PutRequestThrottlerSleepMicroseconds", "S3GetRequestThrottlerCount", "S3GetRequestThrottlerSleepMicroseconds", "S3PutRequestThrottlerCount", "S3PutRequestThrottlerSleepMicroseconds", "RemoteReadThrottlerBytes", "RemoteReadThrottlerSleepMicroseconds", "RemoteWriteThrottlerBytes", "RemoteWriteThrottlerSleepMicroseconds", "LocalReadThrottlerBytes", "LocalReadThrottlerSleepMicroseconds", "LocalWriteThrottlerBytes", "LocalWriteThrottlerSleepMicroseconds", "ThrottlerSleepMicroseconds", "QueryMaskingRulesMatch", "ReplicatedPartFetches", "ReplicatedPartFailedFetches", "ObsoleteReplicatedParts", "ReplicatedPartMerges", "ReplicatedPartFetchesOfMerged", "ReplicatedPartMutations", "ReplicatedPartChecks", "ReplicatedPartChecksFailed", "ReplicatedDataLoss", "InsertedRows", "InsertedBytes", "DelayedInserts", "RejectedInserts", "DelayedInsertsMilliseconds", "DistributedDelayedInserts", "DistributedRejectedInserts", "DistributedDelayedInsertsMilliseconds", "DuplicatedInsertedBlocks", "ZooKeeperInit", "ZooKeeperTransactions", "ZooKeeperList", "ZooKeeperCreate", "ZooKeeperRemove", "ZooKeeperExists", "ZooKeeperGet", "ZooKeeperSet", "ZooKeeperMulti", "ZooKeeperCheck", "ZooKeeperSync", "ZooKeeperClose", "ZooKeeperWatchResponse", "ZooKeeperUserExceptions", "ZooKeeperHardwareExceptions", "ZooKeeperOtherExceptions", "ZooKeeperWaitMicroseconds", "ZooKeeperBytesSent", "ZooKeeperBytesReceived", "DistributedConnectionFailTry", "DistributedConnectionMissingTable", "DistributedConnectionStaleReplica", "DistributedConnectionFailAtAll", "HedgedRequestsChangeReplica", "SuspendSendingQueryToShard", "CompileFunction", "CompiledFunctionExecute", "CompileExpressionsMicroseconds", "CompileExpressionsBytes", "ExecuteShellCommand", "ExternalProcessingCompressedBytesTotal", "ExternalProcessingUncompressedBytesTotal", "ExternalProcessingFilesTotal", "ExternalSortWritePart", "ExternalSortMerge", "ExternalSortCompressedBytes", "ExternalSortUncompressedBytes", "ExternalAggregationWritePart", "ExternalAggregationMerge", "ExternalAggregationCompressedBytes", "ExternalAggregationUncompressedBytes", "ExternalJoinWritePart", "ExternalJoinMerge", "ExternalJoinCompressedBytes", "ExternalJoinUncompressedBytes", "SlowRead", "ReadBackoff", "ReplicaPartialShutdown", "SelectedParts", "SelectedRanges", "SelectedMarks", "SelectedRows", "SelectedBytes", "WaitMarksLoadMicroseconds", "BackgroundLoadingMarksTasks", "LoadedMarksCount", "LoadedMarksMemoryBytes", "Merge", "MergedRows", "MergedUncompressedBytes", "MergesTimeMilliseconds", "MergeTreeDataWriterRows", "MergeTreeDataWriterUncompressedBytes", "MergeTreeDataWriterCompressedBytes", "MergeTreeDataWriterBlocks", "MergeTreeDataWriterBlocksAlreadySorted", "InsertedWideParts", "InsertedCompactParts", "InsertedInMemoryParts", "MergedIntoWideParts", "MergedIntoCompactParts", "MergedIntoInMemoryParts", "MergeTreeDataProjectionWriterRows", "MergeTreeDataProjectionWriterUncompressedBytes", "MergeTreeDataProjectionWriterCompressedBytes", "MergeTreeDataProjectionWriterBlocks", "MergeTreeDataProjectionWriterBlocksAlreadySorted", "CannotRemoveEphemeralNode", "RegexpCreated", "ContextLock", "StorageBufferFlush", "StorageBufferErrorOnFlush", "StorageBufferPassedAllMinThresholds", "StorageBufferPassedTimeMaxThreshold", "StorageBufferPassedRowsMaxThreshold", "StorageBufferPassedBytesMaxThreshold", "StorageBufferPassedTimeFlushThreshold", "StorageBufferPassedRowsFlushThreshold", "StorageBufferPassedBytesFlushThreshold", "StorageBufferLayerLockReadersWaitMilliseconds", "StorageBufferLayerLockWritersWaitMilliseconds", "DictCacheKeysRequested", "DictCacheKeysRequestedMiss", "DictCacheKeysRequestedFound", "DictCacheKeysExpired", "DictCacheKeysNotFound", "DictCacheKeysHit", "DictCacheRequestTimeNs", "DictCacheRequests", "DictCacheLockWriteNs", "DictCacheLockReadNs", "DistributedSyncInsertionTimeoutExceeded", "DataAfterMergeDiffersFromReplica", "DataAfterMutationDiffersFromReplica", "PolygonsAddedToPool", "PolygonsInPoolAllocatedBytes", "RWLockAcquiredReadLocks", "RWLockAcquiredWriteLocks", "RWLockReadersWaitMilliseconds", "RWLockWritersWaitMilliseconds", "DNSError", "RealTimeMicroseconds", "UserTimeMicroseconds", "SystemTimeMicroseconds", "MemoryOvercommitWaitTimeMicroseconds", "MemoryAllocatorPurge", "MemoryAllocatorPurgeTimeMicroseconds", "SoftPageFaults", "HardPageFaults", "OSIOWaitMicroseconds", "OSCPUWaitMicroseconds", "OSCPUVirtualTimeMicroseconds", "OSReadBytes", "OSWriteBytes", "OSReadChars", "OSWriteChars", "PerfCpuCycles", "PerfInstructions", "PerfCacheReferences", "PerfCacheMisses", "PerfBranchInstructions", "PerfBranchMisses", "PerfBusCycles", "PerfStalledCyclesFrontend", "PerfStalledCyclesBackend", "PerfRefCpuCycles", "PerfCpuClock", "PerfTaskClock", "PerfContextSwitches", "PerfCpuMigrations", "PerfAlignmentFaults", "PerfEmulationFaults", "PerfMinEnabledTime", "PerfMinEnabledRunningTime", "PerfDataTLBReferences", "PerfDataTLBMisses", "PerfInstructionTLBReferences", "PerfInstructionTLBMisses", "PerfLocalMemoryReferences", "PerfLocalMemoryMisses", "CreatedHTTPConnections", "CannotWriteToWriteBufferDiscard", "QueryProfilerSignalOverruns", "QueryProfilerRuns", "CreatedLogEntryForMerge", "NotCreatedLogEntryForMerge", "CreatedLogEntryForMutation", "NotCreatedLogEntryForMutation", "S3ReadMicroseconds", "S3ReadRequestsCount", "S3ReadRequestsErrors", "S3ReadRequestsThrottling", "S3ReadRequestsRedirects", "S3WriteMicroseconds", "S3WriteRequestsCount", "S3WriteRequestsErrors", "S3WriteRequestsThrottling", "S3WriteRequestsRedirects", "DiskS3ReadMicroseconds", "DiskS3ReadRequestsCount", "DiskS3ReadRequestsErrors", "DiskS3ReadRequestsThrottling", "DiskS3ReadRequestsRedirects", "DiskS3WriteMicroseconds", "DiskS3WriteRequestsCount", "DiskS3WriteRequestsErrors", "DiskS3WriteRequestsThrottling", "DiskS3WriteRequestsRedirects", "S3DeleteObjects", "S3CopyObject", "S3ListObjects", "S3HeadObject", "S3GetObjectAttributes", "S3CreateMultipartUpload", "S3UploadPartCopy", "S3UploadPart", "S3AbortMultipartUpload", "S3CompleteMultipartUpload", "S3PutObject", "S3GetObject", "DiskS3DeleteObjects", "DiskS3CopyObject", "DiskS3ListObjects", "DiskS3HeadObject", "DiskS3GetObjectAttributes", "DiskS3CreateMultipartUpload", "DiskS3UploadPartCopy", "DiskS3UploadPart", "DiskS3AbortMultipartUpload", "DiskS3CompleteMultipartUpload", "DiskS3PutObject", "DiskS3GetObject", "ReadBufferFromS3Microseconds", "ReadBufferFromS3InitMicroseconds", "ReadBufferFromS3Bytes", "ReadBufferFromS3RequestsErrors", "WriteBufferFromS3Microseconds", "WriteBufferFromS3Bytes", "WriteBufferFromS3RequestsErrors", "QueryMemoryLimitExceeded", "CachedReadBufferReadFromSourceMicroseconds", "CachedReadBufferReadFromCacheMicroseconds", "CachedReadBufferReadFromSourceBytes", "CachedReadBufferReadFromCacheBytes", "CachedReadBufferCacheWriteBytes", "CachedReadBufferCacheWriteMicroseconds", "CachedWriteBufferCacheWriteBytes", "CachedWriteBufferCacheWriteMicroseconds", "RemoteFSSeeks", "RemoteFSPrefetches", "RemoteFSCancelledPrefetches", "RemoteFSUnusedPrefetches", "RemoteFSPrefetchedReads", "RemoteFSPrefetchedBytes", "RemoteFSUnprefetchedReads", "RemoteFSUnprefetchedBytes", "RemoteFSLazySeeks", "RemoteFSSeeksWithReset", "RemoteFSBuffers", "MergeTreePrefetchedReadPoolInit", "WaitPrefetchTaskMicroseconds", "ThreadpoolReaderTaskMicroseconds", "ThreadpoolReaderReadBytes", "ThreadpoolReaderSubmit", "FileSegmentWaitReadBufferMicroseconds", "FileSegmentReadMicroseconds", "FileSegmentWriteMicroseconds", "FileSegmentCacheWriteMicroseconds", "FileSegmentPredownloadMicroseconds", "FileSegmentUsedBytes", "ReadBufferSeekCancelConnection", "SleepFunctionCalls", "SleepFunctionMicroseconds", "ThreadPoolReaderPageCacheHit", "ThreadPoolReaderPageCacheHitBytes", "ThreadPoolReaderPageCacheHitElapsedMicroseconds", "ThreadPoolReaderPageCacheMiss", "ThreadPoolReaderPageCacheMissBytes", "ThreadPoolReaderPageCacheMissElapsedMicroseconds", "AsynchronousReadWaitMicroseconds", "AsynchronousRemoteReadWaitMicroseconds", "SynchronousRemoteReadWaitMicroseconds", "ExternalDataSourceLocalCacheReadBytes", "MainConfigLoads", "AggregationPreallocatedElementsInHashTables", "AggregationHashTablesInitializedAsTwoLevel", "MergeTreeMetadataCacheGet", "MergeTreeMetadataCachePut", "MergeTreeMetadataCacheDelete", "MergeTreeMetadataCacheSeek", "MergeTreeMetadataCacheHit", "MergeTreeMetadataCacheMiss", "KafkaRebalanceRevocations", "KafkaRebalanceAssignments", "KafkaRebalanceErrors", "KafkaMessagesPolled", "KafkaMessagesRead", "KafkaMessagesFailed", "KafkaRowsRead", "KafkaRowsRejected", "KafkaDirectReads", "KafkaBackgroundReads", "KafkaCommits", "KafkaCommitFailures", "KafkaConsumerErrors", "KafkaWrites", "KafkaRowsWritten", "KafkaProducerFlushes", "KafkaMessagesProduced", "KafkaProducerErrors", "ScalarSubqueriesGlobalCacheHit", "ScalarSubqueriesLocalCacheHit", "ScalarSubqueriesCacheMiss", "SchemaInferenceCacheHits", "SchemaInferenceCacheMisses", "SchemaInferenceCacheEvictions", "SchemaInferenceCacheInvalidations", "KeeperPacketsSent", "KeeperPacketsReceived", "KeeperRequestTotal", "KeeperLatency", "KeeperCommits", "KeeperCommitsFailed", "KeeperSnapshotCreations", "KeeperSnapshotCreationsFailed", "KeeperSnapshotApplys", "KeeperSnapshotApplysFailed", "KeeperReadSnapshot", "KeeperSaveSnapshot", "KeeperCreateRequest", "KeeperRemoveRequest", "KeeperSetRequest", "KeeperCheckRequest", "KeeperMultiRequest", "KeeperMultiReadRequest", "KeeperGetRequest", "KeeperListRequest", "KeeperExistsRequest", "OverflowBreak", "OverflowThrow", "OverflowAny", "ServerStartupMilliseconds", "IOUringSQEsSubmitted", "IOUringSQEsResubmits", "IOUringCQEsCompleted", "IOUringCQEsFailed", "ReadTaskRequestsReceived", "MergeTreeReadTaskRequestsReceived", "ReadTaskRequestsSent", "MergeTreeReadTaskRequestsSent", "MergeTreeAllRangesAnnouncementsSent", "ReadTaskRequestsSentElapsedMicroseconds", "MergeTreeReadTaskRequestsSentElapsedMicroseconds", "MergeTreeAllRangesAnnouncementsSentElapsedMicroseconds", "LogTest", "LogTrace", "LogDebug", "LogInfo", "LogWarning", "LogError", "LogFatal", "CreatedReadBufferAIOFailed", "ReadBufferAIORead", "ReadBufferAIOReadBytes", "VoluntaryContextSwitches", "WriteBufferAIOWrite", "WriteBufferAIOWriteBytes", "CreatedReadBufferAIO", "InvoluntaryContextSwitches", "S3ReadBytes", "S3WriteBytes"), // nolint:lll
			},
			optsMetrics: []inputs.PointCheckOption{
				inputs.WithTypeChecking(false),
				inputs.WithExtraTags(map[string]string{"tag1": "", "tag2": ""}),
				inputs.WithOptionalFields("Merge", "Move", "PartMutation", "ReplicatedFetch", "ReplicatedSend", "ReplicatedChecks", "BackgroundMergesAndMutationsPoolTask", "BackgroundMergesAndMutationsPoolSize", "BackgroundFetchesPoolTask", "BackgroundFetchesPoolSize", "BackgroundCommonPoolTask", "BackgroundCommonPoolSize", "BackgroundMovePoolTask", "BackgroundMovePoolSize", "BackgroundSchedulePoolTask", "BackgroundSchedulePoolSize", "BackgroundBufferFlushSchedulePoolTask", "BackgroundBufferFlushSchedulePoolSize", "BackgroundDistributedSchedulePoolTask", "BackgroundDistributedSchedulePoolSize", "BackgroundMessageBrokerSchedulePoolTask", "BackgroundMessageBrokerSchedulePoolSize", "CacheDictionaryUpdateQueueBatches", "CacheDictionaryUpdateQueueKeys", "DiskSpaceReservedForMerge", "DistributedSend", "QueryPreempted", "TCPConnection", "MySQLConnection", "HTTPConnection", "InterserverConnection", "PostgreSQLConnection", "OpenFileForRead", "OpenFileForWrite", "TotalTemporaryFiles", "TemporaryFilesForSort", "TemporaryFilesForAggregation", "TemporaryFilesForJoin", "TemporaryFilesUnknown", "Read", "RemoteRead", "Write", "NetworkReceive", "NetworkSend", "SendScalars", "SendExternalTables", "QueryThread", "ReadonlyReplica", "MemoryTracking", "EphemeralNode", "ZooKeeperSession", "ZooKeeperWatch", "ZooKeeperRequest", "DelayedInserts", "ContextLockWait", "StorageBufferRows", "StorageBufferBytes", "DictCacheRequests", "Revision", "VersionInteger", "RWLockWaitingReaders", "RWLockWaitingWriters", "RWLockActiveReaders", "RWLockActiveWriters", "GlobalThread", "GlobalThreadActive", "LocalThread", "LocalThreadActive", "MergeTreeDataSelectExecutorThreads", "MergeTreeDataSelectExecutorThreadsActive", "BackupsThreads", "BackupsThreadsActive", "RestoreThreads", "RestoreThreadsActive", "MarksLoaderThreads", "MarksLoaderThreadsActive", "IOPrefetchThreads", "IOPrefetchThreadsActive", "IOWriterThreads", "IOWriterThreadsActive", "IOThreads", "IOThreadsActive", "ThreadPoolRemoteFSReaderThreads", "ThreadPoolRemoteFSReaderThreadsActive", "ThreadPoolFSReaderThreads", "ThreadPoolFSReaderThreadsActive", "BackupsIOThreads", "BackupsIOThreadsActive", "DiskObjectStorageAsyncThreads", "DiskObjectStorageAsyncThreadsActive", "StorageHiveThreads", "StorageHiveThreadsActive", "TablesLoaderThreads", "TablesLoaderThreadsActive", "DatabaseOrdinaryThreads", "DatabaseOrdinaryThreadsActive", "DatabaseOnDiskThreads", "DatabaseOnDiskThreadsActive", "DatabaseCatalogThreads", "DatabaseCatalogThreadsActive", "DestroyAggregatesThreads", "DestroyAggregatesThreadsActive", "HashedDictionaryThreads", "HashedDictionaryThreadsActive", "CacheDictionaryThreads", "CacheDictionaryThreadsActive", "ParallelFormattingOutputFormatThreads", "ParallelFormattingOutputFormatThreadsActive", "ParallelParsingInputFormatThreads", "ParallelParsingInputFormatThreadsActive", "MergeTreeBackgroundExecutorThreads", "MergeTreeBackgroundExecutorThreadsActive", "AsynchronousInsertThreads", "AsynchronousInsertThreadsActive", "StartupSystemTablesThreads", "StartupSystemTablesThreadsActive", "AggregatorThreads", "AggregatorThreadsActive", "DDLWorkerThreads", "DDLWorkerThreadsActive", "StorageDistributedThreads", "StorageDistributedThreadsActive", "DistributedInsertThreads", "DistributedInsertThreadsActive", "StorageS3Threads", "StorageS3ThreadsActive", "MergeTreePartsLoaderThreads", "MergeTreePartsLoaderThreadsActive", "MergeTreePartsCleanerThreads", "MergeTreePartsCleanerThreadsActive", "SystemReplicasThreads", "SystemReplicasThreadsActive", "RestartReplicaThreads", "RestartReplicaThreadsActive", "QueryPipelineExecutorThreads", "QueryPipelineExecutorThreadsActive", "ParquetDecoderThreads", "ParquetDecoderThreadsActive", "DistributedFilesToInsert", "BrokenDistributedFilesToInsert", "TablesToDropQueueSize", "MaxDDLEntryID", "MaxPushedDDLEntryID", "PartsTemporary", "PartsPreCommitted", "PartsCommitted", "PartsPreActive", "PartsActive", "PartsOutdated", "PartsDeleting", "PartsDeleteOnDestroy", "PartsWide", "PartsCompact", "PartsInMemory", "MMappedFiles", "MMappedFileBytes", "MMappedAllocs", "MMappedAllocBytes", "AsynchronousReadWait", "PendingAsyncInsert", "KafkaConsumers", "KafkaConsumersWithAssignment", "KafkaProducers", "KafkaLibrdkafkaThreads", "KafkaBackgroundReads", "KafkaConsumersInUse", "KafkaWrites", "KafkaAssignedPartitions", "FilesystemCacheReadBuffers", "CacheFileSegments", "CacheDetachedFileSegments", "FilesystemCacheSize", "FilesystemCacheElements", "AsyncInsertCacheSize", "S3Requests", "KeeperAliveConnections", "KeeperOutstandingRequets", "ThreadsInOvercommitTracker", "IOUringPendingEvents", "IOUringInFlightEvents", "ReadTaskRequestsSent", "MergeTreeReadTaskRequestsSent", "MergeTreeAllRangesAnnouncementsSent", "SyncDrainedConnections", "Query", "ActiveAsyncDrainedConnections", "ActiveSyncDrainedConnections", "AsyncDrainedConnections", "BackgroundPoolTask"), // nolint:lll
			},
			optsAsyncMetrics: []inputs.PointCheckOption{
				inputs.WithTypeChecking(false),
				inputs.WithExtraTags(map[string]string{"tag1": "", "tag2": ""}),
				inputs.WithOptionalTags("cpu", "disk", "eth", "unit"),
				inputs.WithOptionalFields("AsynchronousHeavyMetricsCalculationTimeSpent", "AsynchronousHeavyMetricsUpdateInterval", "AsynchronousMetricsCalculationTimeSpent", "AsynchronousMetricsUpdateInterval", "BlockActiveTime", "BlockDiscardBytes", "BlockDiscardMerges", "BlockDiscardOps", "BlockDiscardTime", "BlockInFlightOps", "BlockQueueTime", "BlockReadBytes", "BlockReadMerges", "BlockReadOps", "BlockReadTime", "BlockWriteBytes", "BlockWriteMerges", "BlockWriteOps", "BlockWriteTime", "CPUFrequencyMHz", "CompiledExpressionCacheBytes", "CompiledExpressionCacheCount", "DiskAvailable", "DiskTotal", "DiskUnreserved", "DiskUsed", "FilesystemCacheBytes", "FilesystemCacheFiles", "FilesystemLogsPathAvailableBytes", "FilesystemLogsPathAvailableINodes", "FilesystemLogsPathTotalBytes", "FilesystemLogsPathTotalINodes", "FilesystemLogsPathUsedBytes", "FilesystemLogsPathUsedINodes", "FilesystemMainPathAvailableBytes", "FilesystemMainPathAvailableINodes", "FilesystemMainPathTotalBytes", "FilesystemMainPathTotalINodes", "FilesystemMainPathUsedBytes", "FilesystemMainPathUsedINodes", "HTTPThreads", "InterserverThreads", "Jitter", "LoadAverage", "MMapCacheCells", "MarkCacheBytes", "MarkCacheFiles", "MaxPartCountForPartition", "MemoryCode", "MemoryDataAndStack", "MemoryResident", "MemoryShared", "MemoryVirtual", "MySQLThreads", "NetworkReceiveBytes", "NetworkReceiveDrop", "NetworkReceiveErrors", "NetworkReceivePackets", "NetworkSendBytes", "NetworkSendDrop", "NetworkSendErrors", "NetworkSendPackets", "NumberOfDatabases", "NumberOfDetachedByUserParts", "NumberOfDetachedParts", "NumberOfTables", "OSContextSwitches", "OSGuestNiceTime", "OSGuestNiceTimeCPU", "OSGuestNiceTimeNormalized", "OSGuestTime", "OSGuestTimeCPU", "OSGuestTimeNormalized", "OSIOWaitTime", "OSIOWaitTimeCPU", "OSIOWaitTimeNormalized", "OSIdleTime", "OSIdleTimeCPU", "OSIdleTimeNormalized", "OSInterrupts", "OSIrqTime", "OSIrqTimeCPU", "OSIrqTimeNormalized", "OSMemoryAvailable", "OSMemoryBuffers", "OSMemoryCached", "OSMemoryFreePlusCached", "OSMemoryFreeWithoutCached", "OSMemoryTotal", "OSNiceTime", "OSNiceTimeCPU", "OSNiceTimeNormalized", "OSOpenFiles", "OSProcessesBlocked", "OSProcessesCreated", "OSProcessesRunning", "OSSoftIrqTime", "OSSoftIrqTimeCPU", "OSSoftIrqTimeNormalized", "OSStealTime", "OSStealTimeCPU", "OSStealTimeNormalized", "OSSystemTime", "OSSystemTimeCPU", "OSSystemTimeNormalized", "OSThreadsRunnable", "OSThreadsTotal", "OSUptime", "OSUserTime", "OSUserTimeCPU", "OSUserTimeNormalized", "PostgreSQLThreads", "ReplicasMaxAbsoluteDelay", "ReplicasMaxInsertsInQueue", "ReplicasMaxMergesInQueue", "ReplicasMaxQueueSize", "ReplicasMaxRelativeDelay", "ReplicasSumInsertsInQueue", "ReplicasSumMergesInQueue", "ReplicasSumQueueSize", "TCPThreads", "Temperature", "TotalBytesOfMergeTreeTables", "TotalPartsOfMergeTreeTables", "TotalRowsOfMergeTreeTables", "UncompressedCacheBytes", "UncompressedCacheCells", "Uptime", "jemalloc_active", "jemalloc_allocated", "jemalloc_arenas_all_dirty_purged", "jemalloc_arenas_all_muzzy_purged", "jemalloc_arenas_all_pactive", "jemalloc_arenas_all_pdirty", "jemalloc_arenas_all_pmuzzy", "jemalloc_background_thread_num_runs", "jemalloc_background_thread_num_threads", "jemalloc_background_thread_run_intervals", "jemalloc_epoch", "jemalloc_mapped", "jemalloc_metadata", "jemalloc_metadata_thp", "jemalloc_resident", "jemalloc_retained", "OSMemorySwapCached", "PrometheusThreads"), // nolint:lll
			},
			optsStatusInfo: []inputs.PointCheckOption{
				inputs.WithExtraTags(map[string]string{"tag1": "", "tag2": ""}),
			},
		},
		{
			name: "pubrepo.jiagouyun.com/image-repo-for-testing/clickhouse/clickhouse-server:21.8.15.7",
			// selfBuild: true,
			conf: `
source = "clickhouse"
interval = "10s"
tls_open = false
urls = [""]
[tags]
  tag1 = "some_value"
  tag2 = "some_other_value"`, // set conf URL later.
			exposedPorts: []string{"9363/tcp"},
			mPathCount:   map[string]int{"/": 10},
			optsProfileEvents: []inputs.PointCheckOption{
				inputs.WithTypeChecking(false),
				inputs.WithExtraTags(map[string]string{"tag1": "", "tag2": ""}),
				inputs.WithOptionalFields("Query", "SelectQuery", "InsertQuery", "AsyncInsertQuery", "AsyncInsertBytes", "AsyncInsertCacheHits", "FailedQuery", "FailedSelectQuery", "FailedInsertQuery", "FailedAsyncInsertQuery", "QueryTimeMicroseconds", "SelectQueryTimeMicroseconds", "InsertQueryTimeMicroseconds", "OtherQueryTimeMicroseconds", "FileOpen", "Seek", "ReadBufferFromFileDescriptorRead", "ReadBufferFromFileDescriptorReadFailed", "ReadBufferFromFileDescriptorReadBytes", "WriteBufferFromFileDescriptorWrite", "WriteBufferFromFileDescriptorWriteFailed", "WriteBufferFromFileDescriptorWriteBytes", "FileSync", "DirectorySync", "FileSyncElapsedMicroseconds", "DirectorySyncElapsedMicroseconds", "ReadCompressedBytes", "CompressedReadBufferBlocks", "CompressedReadBufferBytes", "UncompressedCacheHits", "UncompressedCacheMisses", "UncompressedCacheWeightLost", "MMappedFileCacheHits", "MMappedFileCacheMisses", "OpenedFileCacheHits", "OpenedFileCacheMisses", "AIOWrite", "AIOWriteBytes", "AIORead", "AIOReadBytes", "IOBufferAllocs", "IOBufferAllocBytes", "ArenaAllocChunks", "ArenaAllocBytes", "FunctionExecute", "TableFunctionExecute", "MarkCacheHits", "MarkCacheMisses", "QueryCacheHits", "QueryCacheMisses", "CreatedReadBufferOrdinary", "CreatedReadBufferDirectIO", "CreatedReadBufferDirectIOFailed", "CreatedReadBufferMMap", "CreatedReadBufferMMapFailed", "DiskReadElapsedMicroseconds", "DiskWriteElapsedMicroseconds", "NetworkReceiveElapsedMicroseconds", "NetworkSendElapsedMicroseconds", "NetworkReceiveBytes", "NetworkSendBytes", "DiskS3GetRequestThrottlerCount", "DiskS3GetRequestThrottlerSleepMicroseconds", "DiskS3PutRequestThrottlerCount", "DiskS3PutRequestThrottlerSleepMicroseconds", "S3GetRequestThrottlerCount", "S3GetRequestThrottlerSleepMicroseconds", "S3PutRequestThrottlerCount", "S3PutRequestThrottlerSleepMicroseconds", "RemoteReadThrottlerBytes", "RemoteReadThrottlerSleepMicroseconds", "RemoteWriteThrottlerBytes", "RemoteWriteThrottlerSleepMicroseconds", "LocalReadThrottlerBytes", "LocalReadThrottlerSleepMicroseconds", "LocalWriteThrottlerBytes", "LocalWriteThrottlerSleepMicroseconds", "ThrottlerSleepMicroseconds", "QueryMaskingRulesMatch", "ReplicatedPartFetches", "ReplicatedPartFailedFetches", "ObsoleteReplicatedParts", "ReplicatedPartMerges", "ReplicatedPartFetchesOfMerged", "ReplicatedPartMutations", "ReplicatedPartChecks", "ReplicatedPartChecksFailed", "ReplicatedDataLoss", "InsertedRows", "InsertedBytes", "DelayedInserts", "RejectedInserts", "DelayedInsertsMilliseconds", "DistributedDelayedInserts", "DistributedRejectedInserts", "DistributedDelayedInsertsMilliseconds", "DuplicatedInsertedBlocks", "ZooKeeperInit", "ZooKeeperTransactions", "ZooKeeperList", "ZooKeeperCreate", "ZooKeeperRemove", "ZooKeeperExists", "ZooKeeperGet", "ZooKeeperSet", "ZooKeeperMulti", "ZooKeeperCheck", "ZooKeeperSync", "ZooKeeperClose", "ZooKeeperWatchResponse", "ZooKeeperUserExceptions", "ZooKeeperHardwareExceptions", "ZooKeeperOtherExceptions", "ZooKeeperWaitMicroseconds", "ZooKeeperBytesSent", "ZooKeeperBytesReceived", "DistributedConnectionFailTry", "DistributedConnectionMissingTable", "DistributedConnectionStaleReplica", "DistributedConnectionFailAtAll", "HedgedRequestsChangeReplica", "SuspendSendingQueryToShard", "CompileFunction", "CompiledFunctionExecute", "CompileExpressionsMicroseconds", "CompileExpressionsBytes", "ExecuteShellCommand", "ExternalProcessingCompressedBytesTotal", "ExternalProcessingUncompressedBytesTotal", "ExternalProcessingFilesTotal", "ExternalSortWritePart", "ExternalSortMerge", "ExternalSortCompressedBytes", "ExternalSortUncompressedBytes", "ExternalAggregationWritePart", "ExternalAggregationMerge", "ExternalAggregationCompressedBytes", "ExternalAggregationUncompressedBytes", "ExternalJoinWritePart", "ExternalJoinMerge", "ExternalJoinCompressedBytes", "ExternalJoinUncompressedBytes", "SlowRead", "ReadBackoff", "ReplicaPartialShutdown", "SelectedParts", "SelectedRanges", "SelectedMarks", "SelectedRows", "SelectedBytes", "WaitMarksLoadMicroseconds", "BackgroundLoadingMarksTasks", "LoadedMarksCount", "LoadedMarksMemoryBytes", "Merge", "MergedRows", "MergedUncompressedBytes", "MergesTimeMilliseconds", "MergeTreeDataWriterRows", "MergeTreeDataWriterUncompressedBytes", "MergeTreeDataWriterCompressedBytes", "MergeTreeDataWriterBlocks", "MergeTreeDataWriterBlocksAlreadySorted", "InsertedWideParts", "InsertedCompactParts", "InsertedInMemoryParts", "MergedIntoWideParts", "MergedIntoCompactParts", "MergedIntoInMemoryParts", "MergeTreeDataProjectionWriterRows", "MergeTreeDataProjectionWriterUncompressedBytes", "MergeTreeDataProjectionWriterCompressedBytes", "MergeTreeDataProjectionWriterBlocks", "MergeTreeDataProjectionWriterBlocksAlreadySorted", "CannotRemoveEphemeralNode", "RegexpCreated", "ContextLock", "StorageBufferFlush", "StorageBufferErrorOnFlush", "StorageBufferPassedAllMinThresholds", "StorageBufferPassedTimeMaxThreshold", "StorageBufferPassedRowsMaxThreshold", "StorageBufferPassedBytesMaxThreshold", "StorageBufferPassedTimeFlushThreshold", "StorageBufferPassedRowsFlushThreshold", "StorageBufferPassedBytesFlushThreshold", "StorageBufferLayerLockReadersWaitMilliseconds", "StorageBufferLayerLockWritersWaitMilliseconds", "DictCacheKeysRequested", "DictCacheKeysRequestedMiss", "DictCacheKeysRequestedFound", "DictCacheKeysExpired", "DictCacheKeysNotFound", "DictCacheKeysHit", "DictCacheRequestTimeNs", "DictCacheRequests", "DictCacheLockWriteNs", "DictCacheLockReadNs", "DistributedSyncInsertionTimeoutExceeded", "DataAfterMergeDiffersFromReplica", "DataAfterMutationDiffersFromReplica", "PolygonsAddedToPool", "PolygonsInPoolAllocatedBytes", "RWLockAcquiredReadLocks", "RWLockAcquiredWriteLocks", "RWLockReadersWaitMilliseconds", "RWLockWritersWaitMilliseconds", "DNSError", "RealTimeMicroseconds", "UserTimeMicroseconds", "SystemTimeMicroseconds", "MemoryOvercommitWaitTimeMicroseconds", "MemoryAllocatorPurge", "MemoryAllocatorPurgeTimeMicroseconds", "SoftPageFaults", "HardPageFaults", "OSIOWaitMicroseconds", "OSCPUWaitMicroseconds", "OSCPUVirtualTimeMicroseconds", "OSReadBytes", "OSWriteBytes", "OSReadChars", "OSWriteChars", "PerfCpuCycles", "PerfInstructions", "PerfCacheReferences", "PerfCacheMisses", "PerfBranchInstructions", "PerfBranchMisses", "PerfBusCycles", "PerfStalledCyclesFrontend", "PerfStalledCyclesBackend", "PerfRefCpuCycles", "PerfCpuClock", "PerfTaskClock", "PerfContextSwitches", "PerfCpuMigrations", "PerfAlignmentFaults", "PerfEmulationFaults", "PerfMinEnabledTime", "PerfMinEnabledRunningTime", "PerfDataTLBReferences", "PerfDataTLBMisses", "PerfInstructionTLBReferences", "PerfInstructionTLBMisses", "PerfLocalMemoryReferences", "PerfLocalMemoryMisses", "CreatedHTTPConnections", "CannotWriteToWriteBufferDiscard", "QueryProfilerSignalOverruns", "QueryProfilerRuns", "CreatedLogEntryForMerge", "NotCreatedLogEntryForMerge", "CreatedLogEntryForMutation", "NotCreatedLogEntryForMutation", "S3ReadMicroseconds", "S3ReadRequestsCount", "S3ReadRequestsErrors", "S3ReadRequestsThrottling", "S3ReadRequestsRedirects", "S3WriteMicroseconds", "S3WriteRequestsCount", "S3WriteRequestsErrors", "S3WriteRequestsThrottling", "S3WriteRequestsRedirects", "DiskS3ReadMicroseconds", "DiskS3ReadRequestsCount", "DiskS3ReadRequestsErrors", "DiskS3ReadRequestsThrottling", "DiskS3ReadRequestsRedirects", "DiskS3WriteMicroseconds", "DiskS3WriteRequestsCount", "DiskS3WriteRequestsErrors", "DiskS3WriteRequestsThrottling", "DiskS3WriteRequestsRedirects", "S3DeleteObjects", "S3CopyObject", "S3ListObjects", "S3HeadObject", "S3GetObjectAttributes", "S3CreateMultipartUpload", "S3UploadPartCopy", "S3UploadPart", "S3AbortMultipartUpload", "S3CompleteMultipartUpload", "S3PutObject", "S3GetObject", "DiskS3DeleteObjects", "DiskS3CopyObject", "DiskS3ListObjects", "DiskS3HeadObject", "DiskS3GetObjectAttributes", "DiskS3CreateMultipartUpload", "DiskS3UploadPartCopy", "DiskS3UploadPart", "DiskS3AbortMultipartUpload", "DiskS3CompleteMultipartUpload", "DiskS3PutObject", "DiskS3GetObject", "ReadBufferFromS3Microseconds", "ReadBufferFromS3InitMicroseconds", "ReadBufferFromS3Bytes", "ReadBufferFromS3RequestsErrors", "WriteBufferFromS3Microseconds", "WriteBufferFromS3Bytes", "WriteBufferFromS3RequestsErrors", "QueryMemoryLimitExceeded", "CachedReadBufferReadFromSourceMicroseconds", "CachedReadBufferReadFromCacheMicroseconds", "CachedReadBufferReadFromSourceBytes", "CachedReadBufferReadFromCacheBytes", "CachedReadBufferCacheWriteBytes", "CachedReadBufferCacheWriteMicroseconds", "CachedWriteBufferCacheWriteBytes", "CachedWriteBufferCacheWriteMicroseconds", "RemoteFSSeeks", "RemoteFSPrefetches", "RemoteFSCancelledPrefetches", "RemoteFSUnusedPrefetches", "RemoteFSPrefetchedReads", "RemoteFSPrefetchedBytes", "RemoteFSUnprefetchedReads", "RemoteFSUnprefetchedBytes", "RemoteFSLazySeeks", "RemoteFSSeeksWithReset", "RemoteFSBuffers", "MergeTreePrefetchedReadPoolInit", "WaitPrefetchTaskMicroseconds", "ThreadpoolReaderTaskMicroseconds", "ThreadpoolReaderReadBytes", "ThreadpoolReaderSubmit", "FileSegmentWaitReadBufferMicroseconds", "FileSegmentReadMicroseconds", "FileSegmentWriteMicroseconds", "FileSegmentCacheWriteMicroseconds", "FileSegmentPredownloadMicroseconds", "FileSegmentUsedBytes", "ReadBufferSeekCancelConnection", "SleepFunctionCalls", "SleepFunctionMicroseconds", "ThreadPoolReaderPageCacheHit", "ThreadPoolReaderPageCacheHitBytes", "ThreadPoolReaderPageCacheHitElapsedMicroseconds", "ThreadPoolReaderPageCacheMiss", "ThreadPoolReaderPageCacheMissBytes", "ThreadPoolReaderPageCacheMissElapsedMicroseconds", "AsynchronousReadWaitMicroseconds", "AsynchronousRemoteReadWaitMicroseconds", "SynchronousRemoteReadWaitMicroseconds", "ExternalDataSourceLocalCacheReadBytes", "MainConfigLoads", "AggregationPreallocatedElementsInHashTables", "AggregationHashTablesInitializedAsTwoLevel", "MergeTreeMetadataCacheGet", "MergeTreeMetadataCachePut", "MergeTreeMetadataCacheDelete", "MergeTreeMetadataCacheSeek", "MergeTreeMetadataCacheHit", "MergeTreeMetadataCacheMiss", "KafkaRebalanceRevocations", "KafkaRebalanceAssignments", "KafkaRebalanceErrors", "KafkaMessagesPolled", "KafkaMessagesRead", "KafkaMessagesFailed", "KafkaRowsRead", "KafkaRowsRejected", "KafkaDirectReads", "KafkaBackgroundReads", "KafkaCommits", "KafkaCommitFailures", "KafkaConsumerErrors", "KafkaWrites", "KafkaRowsWritten", "KafkaProducerFlushes", "KafkaMessagesProduced", "KafkaProducerErrors", "ScalarSubqueriesGlobalCacheHit", "ScalarSubqueriesLocalCacheHit", "ScalarSubqueriesCacheMiss", "SchemaInferenceCacheHits", "SchemaInferenceCacheMisses", "SchemaInferenceCacheEvictions", "SchemaInferenceCacheInvalidations", "KeeperPacketsSent", "KeeperPacketsReceived", "KeeperRequestTotal", "KeeperLatency", "KeeperCommits", "KeeperCommitsFailed", "KeeperSnapshotCreations", "KeeperSnapshotCreationsFailed", "KeeperSnapshotApplys", "KeeperSnapshotApplysFailed", "KeeperReadSnapshot", "KeeperSaveSnapshot", "KeeperCreateRequest", "KeeperRemoveRequest", "KeeperSetRequest", "KeeperCheckRequest", "KeeperMultiRequest", "KeeperMultiReadRequest", "KeeperGetRequest", "KeeperListRequest", "KeeperExistsRequest", "OverflowBreak", "OverflowThrow", "OverflowAny", "ServerStartupMilliseconds", "IOUringSQEsSubmitted", "IOUringSQEsResubmits", "IOUringCQEsCompleted", "IOUringCQEsFailed", "ReadTaskRequestsReceived", "MergeTreeReadTaskRequestsReceived", "ReadTaskRequestsSent", "MergeTreeReadTaskRequestsSent", "MergeTreeAllRangesAnnouncementsSent", "ReadTaskRequestsSentElapsedMicroseconds", "MergeTreeReadTaskRequestsSentElapsedMicroseconds", "MergeTreeAllRangesAnnouncementsSentElapsedMicroseconds", "LogTest", "LogTrace", "LogDebug", "LogInfo", "LogWarning", "LogError", "LogFatal", "CreatedReadBufferAIOFailed", "ReadBufferAIORead", "ReadBufferAIOReadBytes", "VoluntaryContextSwitches", "WriteBufferAIOWrite", "WriteBufferAIOWriteBytes", "CreatedReadBufferAIO", "InvoluntaryContextSwitches", "S3ReadBytes", "S3WriteBytes"), // nolint:lll
			},
			optsMetrics: []inputs.PointCheckOption{
				inputs.WithTypeChecking(false),
				inputs.WithExtraTags(map[string]string{"tag1": "", "tag2": ""}),
				inputs.WithOptionalFields("Merge", "Move", "PartMutation", "ReplicatedFetch", "ReplicatedSend", "ReplicatedChecks", "BackgroundMergesAndMutationsPoolTask", "BackgroundMergesAndMutationsPoolSize", "BackgroundFetchesPoolTask", "BackgroundFetchesPoolSize", "BackgroundCommonPoolTask", "BackgroundCommonPoolSize", "BackgroundMovePoolTask", "BackgroundMovePoolSize", "BackgroundSchedulePoolTask", "BackgroundSchedulePoolSize", "BackgroundBufferFlushSchedulePoolTask", "BackgroundBufferFlushSchedulePoolSize", "BackgroundDistributedSchedulePoolTask", "BackgroundDistributedSchedulePoolSize", "BackgroundMessageBrokerSchedulePoolTask", "BackgroundMessageBrokerSchedulePoolSize", "CacheDictionaryUpdateQueueBatches", "CacheDictionaryUpdateQueueKeys", "DiskSpaceReservedForMerge", "DistributedSend", "QueryPreempted", "TCPConnection", "MySQLConnection", "HTTPConnection", "InterserverConnection", "PostgreSQLConnection", "OpenFileForRead", "OpenFileForWrite", "TotalTemporaryFiles", "TemporaryFilesForSort", "TemporaryFilesForAggregation", "TemporaryFilesForJoin", "TemporaryFilesUnknown", "Read", "RemoteRead", "Write", "NetworkReceive", "NetworkSend", "SendScalars", "SendExternalTables", "QueryThread", "ReadonlyReplica", "MemoryTracking", "EphemeralNode", "ZooKeeperSession", "ZooKeeperWatch", "ZooKeeperRequest", "DelayedInserts", "ContextLockWait", "StorageBufferRows", "StorageBufferBytes", "DictCacheRequests", "Revision", "VersionInteger", "RWLockWaitingReaders", "RWLockWaitingWriters", "RWLockActiveReaders", "RWLockActiveWriters", "GlobalThread", "GlobalThreadActive", "LocalThread", "LocalThreadActive", "MergeTreeDataSelectExecutorThreads", "MergeTreeDataSelectExecutorThreadsActive", "BackupsThreads", "BackupsThreadsActive", "RestoreThreads", "RestoreThreadsActive", "MarksLoaderThreads", "MarksLoaderThreadsActive", "IOPrefetchThreads", "IOPrefetchThreadsActive", "IOWriterThreads", "IOWriterThreadsActive", "IOThreads", "IOThreadsActive", "ThreadPoolRemoteFSReaderThreads", "ThreadPoolRemoteFSReaderThreadsActive", "ThreadPoolFSReaderThreads", "ThreadPoolFSReaderThreadsActive", "BackupsIOThreads", "BackupsIOThreadsActive", "DiskObjectStorageAsyncThreads", "DiskObjectStorageAsyncThreadsActive", "StorageHiveThreads", "StorageHiveThreadsActive", "TablesLoaderThreads", "TablesLoaderThreadsActive", "DatabaseOrdinaryThreads", "DatabaseOrdinaryThreadsActive", "DatabaseOnDiskThreads", "DatabaseOnDiskThreadsActive", "DatabaseCatalogThreads", "DatabaseCatalogThreadsActive", "DestroyAggregatesThreads", "DestroyAggregatesThreadsActive", "HashedDictionaryThreads", "HashedDictionaryThreadsActive", "CacheDictionaryThreads", "CacheDictionaryThreadsActive", "ParallelFormattingOutputFormatThreads", "ParallelFormattingOutputFormatThreadsActive", "ParallelParsingInputFormatThreads", "ParallelParsingInputFormatThreadsActive", "MergeTreeBackgroundExecutorThreads", "MergeTreeBackgroundExecutorThreadsActive", "AsynchronousInsertThreads", "AsynchronousInsertThreadsActive", "StartupSystemTablesThreads", "StartupSystemTablesThreadsActive", "AggregatorThreads", "AggregatorThreadsActive", "DDLWorkerThreads", "DDLWorkerThreadsActive", "StorageDistributedThreads", "StorageDistributedThreadsActive", "DistributedInsertThreads", "DistributedInsertThreadsActive", "StorageS3Threads", "StorageS3ThreadsActive", "MergeTreePartsLoaderThreads", "MergeTreePartsLoaderThreadsActive", "MergeTreePartsCleanerThreads", "MergeTreePartsCleanerThreadsActive", "SystemReplicasThreads", "SystemReplicasThreadsActive", "RestartReplicaThreads", "RestartReplicaThreadsActive", "QueryPipelineExecutorThreads", "QueryPipelineExecutorThreadsActive", "ParquetDecoderThreads", "ParquetDecoderThreadsActive", "DistributedFilesToInsert", "BrokenDistributedFilesToInsert", "TablesToDropQueueSize", "MaxDDLEntryID", "MaxPushedDDLEntryID", "PartsTemporary", "PartsPreCommitted", "PartsCommitted", "PartsPreActive", "PartsActive", "PartsOutdated", "PartsDeleting", "PartsDeleteOnDestroy", "PartsWide", "PartsCompact", "PartsInMemory", "MMappedFiles", "MMappedFileBytes", "MMappedAllocs", "MMappedAllocBytes", "AsynchronousReadWait", "PendingAsyncInsert", "KafkaConsumers", "KafkaConsumersWithAssignment", "KafkaProducers", "KafkaLibrdkafkaThreads", "KafkaBackgroundReads", "KafkaConsumersInUse", "KafkaWrites", "KafkaAssignedPartitions", "FilesystemCacheReadBuffers", "CacheFileSegments", "CacheDetachedFileSegments", "FilesystemCacheSize", "FilesystemCacheElements", "AsyncInsertCacheSize", "S3Requests", "KeeperAliveConnections", "KeeperOutstandingRequets", "ThreadsInOvercommitTracker", "IOUringPendingEvents", "IOUringInFlightEvents", "ReadTaskRequestsSent", "MergeTreeReadTaskRequestsSent", "MergeTreeAllRangesAnnouncementsSent", "SyncDrainedConnections", "Query", "ActiveAsyncDrainedConnections", "ActiveSyncDrainedConnections", "AsyncDrainedConnections", "BackgroundPoolTask"), // nolint:lll
			},
			optsAsyncMetrics: []inputs.PointCheckOption{
				inputs.WithTypeChecking(false),
				inputs.WithExtraTags(map[string]string{"tag1": "", "tag2": ""}),
				inputs.WithOptionalTags("cpu", "disk", "eth", "unit"),
				inputs.WithOptionalFields("AsynchronousHeavyMetricsCalculationTimeSpent", "AsynchronousHeavyMetricsUpdateInterval", "AsynchronousMetricsCalculationTimeSpent", "AsynchronousMetricsUpdateInterval", "BlockActiveTime", "BlockDiscardBytes", "BlockDiscardMerges", "BlockDiscardOps", "BlockDiscardTime", "BlockInFlightOps", "BlockQueueTime", "BlockReadBytes", "BlockReadMerges", "BlockReadOps", "BlockReadTime", "BlockWriteBytes", "BlockWriteMerges", "BlockWriteOps", "BlockWriteTime", "CPUFrequencyMHz", "CompiledExpressionCacheBytes", "CompiledExpressionCacheCount", "DiskAvailable", "DiskTotal", "DiskUnreserved", "DiskUsed", "FilesystemCacheBytes", "FilesystemCacheFiles", "FilesystemLogsPathAvailableBytes", "FilesystemLogsPathAvailableINodes", "FilesystemLogsPathTotalBytes", "FilesystemLogsPathTotalINodes", "FilesystemLogsPathUsedBytes", "FilesystemLogsPathUsedINodes", "FilesystemMainPathAvailableBytes", "FilesystemMainPathAvailableINodes", "FilesystemMainPathTotalBytes", "FilesystemMainPathTotalINodes", "FilesystemMainPathUsedBytes", "FilesystemMainPathUsedINodes", "HTTPThreads", "InterserverThreads", "Jitter", "LoadAverage", "MMapCacheCells", "MarkCacheBytes", "MarkCacheFiles", "MaxPartCountForPartition", "MemoryCode", "MemoryDataAndStack", "MemoryResident", "MemoryShared", "MemoryVirtual", "MySQLThreads", "NetworkReceiveBytes", "NetworkReceiveDrop", "NetworkReceiveErrors", "NetworkReceivePackets", "NetworkSendBytes", "NetworkSendDrop", "NetworkSendErrors", "NetworkSendPackets", "NumberOfDatabases", "NumberOfDetachedByUserParts", "NumberOfDetachedParts", "NumberOfTables", "OSContextSwitches", "OSGuestNiceTime", "OSGuestNiceTimeCPU", "OSGuestNiceTimeNormalized", "OSGuestTime", "OSGuestTimeCPU", "OSGuestTimeNormalized", "OSIOWaitTime", "OSIOWaitTimeCPU", "OSIOWaitTimeNormalized", "OSIdleTime", "OSIdleTimeCPU", "OSIdleTimeNormalized", "OSInterrupts", "OSIrqTime", "OSIrqTimeCPU", "OSIrqTimeNormalized", "OSMemoryAvailable", "OSMemoryBuffers", "OSMemoryCached", "OSMemoryFreePlusCached", "OSMemoryFreeWithoutCached", "OSMemoryTotal", "OSNiceTime", "OSNiceTimeCPU", "OSNiceTimeNormalized", "OSOpenFiles", "OSProcessesBlocked", "OSProcessesCreated", "OSProcessesRunning", "OSSoftIrqTime", "OSSoftIrqTimeCPU", "OSSoftIrqTimeNormalized", "OSStealTime", "OSStealTimeCPU", "OSStealTimeNormalized", "OSSystemTime", "OSSystemTimeCPU", "OSSystemTimeNormalized", "OSThreadsRunnable", "OSThreadsTotal", "OSUptime", "OSUserTime", "OSUserTimeCPU", "OSUserTimeNormalized", "PostgreSQLThreads", "ReplicasMaxAbsoluteDelay", "ReplicasMaxInsertsInQueue", "ReplicasMaxMergesInQueue", "ReplicasMaxQueueSize", "ReplicasMaxRelativeDelay", "ReplicasSumInsertsInQueue", "ReplicasSumMergesInQueue", "ReplicasSumQueueSize", "TCPThreads", "Temperature", "TotalBytesOfMergeTreeTables", "TotalPartsOfMergeTreeTables", "TotalRowsOfMergeTreeTables", "UncompressedCacheBytes", "UncompressedCacheCells", "Uptime", "jemalloc_active", "jemalloc_allocated", "jemalloc_arenas_all_dirty_purged", "jemalloc_arenas_all_muzzy_purged", "jemalloc_arenas_all_pactive", "jemalloc_arenas_all_pdirty", "jemalloc_arenas_all_pmuzzy", "jemalloc_background_thread_num_runs", "jemalloc_background_thread_num_threads", "jemalloc_background_thread_run_intervals", "jemalloc_epoch", "jemalloc_mapped", "jemalloc_metadata", "jemalloc_metadata_thp", "jemalloc_resident", "jemalloc_retained", "OSMemorySwapCached", "PrometheusThreads"), // nolint:lll
			},
			optsStatusInfo: []inputs.PointCheckOption{
				inputs.WithExtraTags(map[string]string{"tag1": "", "tag2": ""}),
			},
		},
	}

	var cases []*caseSpec

	// compose cases
	for _, base := range bases {
		feeder := io.NewMockedFeeder()

		ipt := NewProm()
		ipt.Feeder = feeder

		_, err := toml.Decode(base.conf, ipt)
		require.NoError(t, err)

		// URL from ENV.
		envs := []string{
			"ALLOW_NONE_AUTHENTICATION=yes",
		}

		repoTag := strings.Split(base.name, ":")
		cases = append(cases, &caseSpec{
			t:            t,
			ipt:          ipt,
			name:         base.name,
			feeder:       feeder,
			envs:         envs,
			repo:         repoTag[0],
			repoTag:      repoTag[1],
			exposedPorts: base.exposedPorts,

			optsProfileEvents: base.optsProfileEvents,
			optsMetrics:       base.optsMetrics,
			optsAsyncMetrics:  base.optsAsyncMetrics,
			optsStatusInfo:    base.optsStatusInfo,

			cr: &testutils.CaseResult{
				Name:        t.Name(),
				Case:        base.name,
				ExtraFields: map[string]any{},
				ExtraTags: map[string]string{
					"image":       repoTag[0],
					"image_tag":   repoTag[1],
					"docker_host": remote.Host,
					"docker_port": remote.Port,
				},
			},
		})
	}

	return cases, nil
}

////////////////////////////////////////////////////////////////////////////////

// caseSpec.
type caseSpec struct {
	t *testing.T

	name         string
	repo         string
	repoTag      string
	envs         []string
	exposedPorts []string
	serverPorts  []string
	mCount       map[string]struct{}

	optsProfileEvents []inputs.PointCheckOption
	optsMetrics       []inputs.PointCheckOption
	optsAsyncMetrics  []inputs.PointCheckOption
	optsStatusInfo    []inputs.PointCheckOption

	ipt    *Input
	feeder *io.MockedFeeder

	pool     *dt.Pool
	resource *dt.Resource

	cr *testutils.CaseResult
}

func (cs *caseSpec) checkPoint(pts []*point.Point) error {
	for _, pt := range pts {
		measurement := string(pt.Name())
		var opts []inputs.PointCheckOption
		opts = append(opts, inputs.WithExtraTags(cs.ipt.Tags))

		switch measurement {
		case "ClickHouseProfileEvents":
			opts = append(opts, cs.optsProfileEvents...)
			opts = append(opts, inputs.WithDoc(&ProfileEventsMeasurement{}))
			msgs := inputs.CheckPoint(pt, opts...)

			for _, msg := range msgs {
				cs.t.Logf("check measurement %s failed: %+#v", measurement, msg)
			}

			if len(msgs) > 0 {
				return fmt.Errorf("check measurement %s failed: %+#v", measurement, msgs)
			}

			cs.mCount[measurement] = struct{}{}
		case "ClickHouseMetrics":
			opts = append(opts, cs.optsMetrics...)
			opts = append(opts, inputs.WithDoc(&MetricsMeasurement{}))
			msgs := inputs.CheckPoint(pt, opts...)

			for _, msg := range msgs {
				cs.t.Logf("check measurement %s failed: %+#v", measurement, msg)
			}

			if len(msgs) > 0 {
				return fmt.Errorf("check measurement %s failed: %+#v", measurement, msgs)
			}

			cs.mCount[measurement] = struct{}{}
		case "ClickHouseAsyncMetrics":
			opts = append(opts, cs.optsAsyncMetrics...)
			opts = append(opts, inputs.WithDoc(&AsyncMetricsMeasurement{}))
			msgs := inputs.CheckPoint(pt, opts...)

			for _, msg := range msgs {
				cs.t.Logf("check measurement %s failed: %+#v", measurement, msg)
			}

			if len(msgs) > 0 {
				return fmt.Errorf("check measurement %s failed: %+#v", measurement, msgs)
			}

			cs.mCount[measurement] = struct{}{}
		case "ClickHouseStatusInfo":
			opts = append(opts, cs.optsStatusInfo...)
			opts = append(opts, inputs.WithDoc(&StatusInfoMeasurement{}))
			msgs := inputs.CheckPoint(pt, opts...)

			for _, msg := range msgs {
				cs.t.Logf("check measurement %s failed: %+#v", measurement, msg)
			}

			if len(msgs) > 0 {
				return fmt.Errorf("cheoptsDBck measurement %s failed: %+#v", measurement, msgs)
			}

			cs.mCount[measurement] = struct{}{}

		default: // TODO: check other measurement
			panic("unknown measurement: " + measurement)
		}

		// check if tag appended
		if len(cs.ipt.Tags) != 0 {
			cs.t.Logf("checking tags %+#v...", cs.ipt.Tags)

			tags := pt.Tags()
			for k, expect := range cs.ipt.Tags {
				if v := tags.Get([]byte(k)); v != nil {
					got := string(v.GetD())
					if got != expect {
						return fmt.Errorf("expect tag value %s, got %s", expect, got)
					}
				} else {
					return fmt.Errorf("tag %s not found, got %v", k, tags)
				}
			}
		}
	}

	// TODO: some other checking on @pts, such as `if some required measurements exist'...

	return nil
}

func (cs *caseSpec) run() error {
	// start remote image server
	r := testutils.GetRemote()
	dockerTCP := r.TCPURL() // got "tcp://" + net.JoinHostPort(i.Host, i.Port) 2375

	cs.t.Logf("get remote: %+#v, TCP: %s", r, dockerTCP)

	start := time.Now()

	p, err := dt.NewPool(dockerTCP)
	if err != nil {
		return err
	}

	uniqueContainerName := testutils.GetUniqueContainerName(inputName)

	resource, err := p.RunWithOptions(&dt.RunOptions{
		Name: uniqueContainerName, // ATTENTION: not cs.name.

		// specify container image & tag
		Repository: cs.repo,
		Tag:        cs.repoTag,

		ExposedPorts: cs.exposedPorts,

		// container run-time envs
		Env: cs.envs,
	}, func(c *docker.HostConfig) {
		c.RestartPolicy = docker.RestartPolicy{Name: "no"}
		c.AutoRemove = true
	})
	if err != nil {
		return err
	}

	cs.pool = p
	cs.resource = resource

	if err := cs.getMappingPorts(); err != nil {
		return err
	}
	cs.ipt.URLs[0] = getConfAccessPoint(r.Host, cs.serverPorts[0]) // set conf URL here.

	cs.t.Logf("check service(%s:%v)...", r.Host, cs.serverPorts)

	if err := cs.portsOK(r); err != nil {
		return err
	}

	cs.cr.AddField("container_ready_cost", int64(time.Since(start)))

	var wg sync.WaitGroup

	// start input
	cs.t.Logf("start input...")
	wg.Add(1)
	go func() {
		defer wg.Done()
		cs.ipt.Run()
	}()

	// wait data
	start = time.Now()
	cs.t.Logf("wait points...")
	pts, err := cs.feeder.AnyPoints()
	if err != nil {
		return err
	}

	cs.cr.AddField("point_latency", int64(time.Since(start)))
	cs.cr.AddField("point_count", len(pts))

	cs.t.Logf("get %d points", len(pts))
	cs.mCount = make(map[string]struct{})
	if err := cs.checkPoint(pts); err != nil {
		return err
	}

	cs.t.Logf("stop input...")
	cs.ipt.Terminate()

	require.GreaterOrEqual(cs.t, len(cs.mCount), 3) // At lest 3 Metric out.

	cs.t.Logf("exit...")
	wg.Wait()

	return nil
}

func (cs *caseSpec) getMappingPorts() error {
	cs.serverPorts = make([]string, len(cs.exposedPorts))
	for k, v := range cs.exposedPorts {
		mapStr := cs.resource.GetHostPort(v)
		_, port, err := net.SplitHostPort(mapStr)
		if err != nil {
			return err
		}
		cs.serverPorts[k] = port
	}
	return nil
}

func (cs *caseSpec) portsOK(r *testutils.RemoteInfo) error {
	for _, v := range cs.serverPorts {
		if !r.PortOK(docker.Port(v).Port(), time.Minute) {
			return fmt.Errorf("service checking failed")
		}
	}
	return nil
}
