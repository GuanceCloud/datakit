// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package clickhousev1

import (
	"fmt"
	"net"
	"net/netip"
	"net/url"
	"os"
	"sync"
	T "testing"
	"time"

	"github.com/BurntSushi/toml"
	"github.com/GuanceCloud/cliutils/point"
	dt "github.com/ory/dockertest/v3"
	"github.com/ory/dockertest/v3/docker"
	"github.com/stretchr/testify/assert"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/testutils"
)

type caseSpec struct {
	t *T.T

	name        string
	repo        string // docker name
	repoTag     string // docker tag
	envs        []string
	servicePort string // port (rand)）

	optsProfileEvents []inputs.PointCheckOption
	optsMetrics       []inputs.PointCheckOption
	optsAsyncMetrics  []inputs.PointCheckOption
	optsStatusInfo    []inputs.PointCheckOption

	ipt    *Input // This is real prom
	feeder *io.MockedFeeder

	pool     *dt.Pool
	resource *dt.Resource

	cr *testutils.CaseResult // collect `go test -run` metric
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
		default: // TODO: check other measurement
			return nil
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

	hostname, err := os.Hostname()
	if err != nil {
		cs.t.Logf("get hostname failed: %s, ignored", err)
		hostname = "unknown-hostname"
	}

	containerName := fmt.Sprintf("%s.%s", hostname, cs.name)

	// remove the container if exist.
	if err := p.RemoveContainerByName(containerName); err != nil {
		return err
	}

	resource, err := p.RunWithOptions(&dt.RunOptions{
		// specify container image & tag
		Repository: cs.repo,
		Tag:        cs.repoTag,

		// port binding
		PortBindings: map[docker.Port][]docker.PortBinding{
			"9363/tcp": {{HostIP: "0.0.0.0", HostPort: cs.servicePort}},
		},

		Name: containerName,

		// container run-time envs
		Env: cs.envs,
	}, func(c *docker.HostConfig) {
		c.RestartPolicy = docker.RestartPolicy{Name: "no"}
	})
	if err != nil {
		return err
	}

	cs.pool = p
	cs.resource = resource

	cs.t.Logf("check service(%s:%s)...", r.Host, cs.servicePort)
	if !r.PortOK(cs.servicePort, time.Minute) {
		return fmt.Errorf("service checking failed")
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
	if err := cs.checkPoint(pts); err != nil {
		return err
	}

	cs.t.Logf("stop input...")
	cs.ipt.Terminate()

	cs.t.Logf("exit...")
	wg.Wait()

	return nil
}

func buildCases(t *T.T) ([]*caseSpec, error) {
	t.Helper()

	remote := testutils.GetRemote()

	bases := []struct {
		name              string
		repo              string // docker name
		repoTag           string // docker tag
		conf              string
		optsProfileEvents []inputs.PointCheckOption
		optsMetrics       []inputs.PointCheckOption
		optsAsyncMetrics  []inputs.PointCheckOption
		optsStatusInfo    []inputs.PointCheckOption
	}{
		{
			name:    "remote-clickhouse",
			repo:    "pubrepo.jiagouyun.com/image-repo-for-testing/clickhouse/clickhouse-server",
			repoTag: "22.8.15.23",
			conf: fmt.Sprintf(`
source = "clickhouse"
# metric_types = ["counter", "gauge"]

interval = "10s"
tls_open = false
urls = ["http://%s/metrics"]

[tags]
  tag1 = "some_value"
  tag2 = "some_other_value"`, net.JoinHostPort(remote.Host, fmt.Sprintf("%d", testutils.RandPort("tcp")))),
			optsProfileEvents: []inputs.PointCheckOption{
				inputs.WithTypeChecking(false),
				inputs.WithExtraTags(map[string]string{"instance": "", "tag1": "", "tag2": ""}),
				inputs.WithOptionalFields("Query", "SelectQuery", "InsertQuery", "AsyncInsertQuery", "AsyncInsertBytes", "AsyncInsertCacheHits", "FailedQuery", "FailedSelectQuery", "FailedInsertQuery", "FailedAsyncInsertQuery", "QueryTimeMicroseconds", "SelectQueryTimeMicroseconds", "InsertQueryTimeMicroseconds", "OtherQueryTimeMicroseconds", "FileOpen", "Seek", "ReadBufferFromFileDescriptorRead", "ReadBufferFromFileDescriptorReadFailed", "ReadBufferFromFileDescriptorReadBytes", "WriteBufferFromFileDescriptorWrite", "WriteBufferFromFileDescriptorWriteFailed", "WriteBufferFromFileDescriptorWriteBytes", "FileSync", "DirectorySync", "FileSyncElapsedMicroseconds", "DirectorySyncElapsedMicroseconds", "ReadCompressedBytes", "CompressedReadBufferBlocks", "CompressedReadBufferBytes", "UncompressedCacheHits", "UncompressedCacheMisses", "UncompressedCacheWeightLost", "MMappedFileCacheHits", "MMappedFileCacheMisses", "OpenedFileCacheHits", "OpenedFileCacheMisses", "AIOWrite", "AIOWriteBytes", "AIORead", "AIOReadBytes", "IOBufferAllocs", "IOBufferAllocBytes", "ArenaAllocChunks", "ArenaAllocBytes", "FunctionExecute", "TableFunctionExecute", "MarkCacheHits", "MarkCacheMisses", "QueryCacheHits", "QueryCacheMisses", "CreatedReadBufferOrdinary", "CreatedReadBufferDirectIO", "CreatedReadBufferDirectIOFailed", "CreatedReadBufferMMap", "CreatedReadBufferMMapFailed", "DiskReadElapsedMicroseconds", "DiskWriteElapsedMicroseconds", "NetworkReceiveElapsedMicroseconds", "NetworkSendElapsedMicroseconds", "NetworkReceiveBytes", "NetworkSendBytes", "DiskS3GetRequestThrottlerCount", "DiskS3GetRequestThrottlerSleepMicroseconds", "DiskS3PutRequestThrottlerCount", "DiskS3PutRequestThrottlerSleepMicroseconds", "S3GetRequestThrottlerCount", "S3GetRequestThrottlerSleepMicroseconds", "S3PutRequestThrottlerCount", "S3PutRequestThrottlerSleepMicroseconds", "RemoteReadThrottlerBytes", "RemoteReadThrottlerSleepMicroseconds", "RemoteWriteThrottlerBytes", "RemoteWriteThrottlerSleepMicroseconds", "LocalReadThrottlerBytes", "LocalReadThrottlerSleepMicroseconds", "LocalWriteThrottlerBytes", "LocalWriteThrottlerSleepMicroseconds", "ThrottlerSleepMicroseconds", "QueryMaskingRulesMatch", "ReplicatedPartFetches", "ReplicatedPartFailedFetches", "ObsoleteReplicatedParts", "ReplicatedPartMerges", "ReplicatedPartFetchesOfMerged", "ReplicatedPartMutations", "ReplicatedPartChecks", "ReplicatedPartChecksFailed", "ReplicatedDataLoss", "InsertedRows", "InsertedBytes", "DelayedInserts", "RejectedInserts", "DelayedInsertsMilliseconds", "DistributedDelayedInserts", "DistributedRejectedInserts", "DistributedDelayedInsertsMilliseconds", "DuplicatedInsertedBlocks", "ZooKeeperInit", "ZooKeeperTransactions", "ZooKeeperList", "ZooKeeperCreate", "ZooKeeperRemove", "ZooKeeperExists", "ZooKeeperGet", "ZooKeeperSet", "ZooKeeperMulti", "ZooKeeperCheck", "ZooKeeperSync", "ZooKeeperClose", "ZooKeeperWatchResponse", "ZooKeeperUserExceptions", "ZooKeeperHardwareExceptions", "ZooKeeperOtherExceptions", "ZooKeeperWaitMicroseconds", "ZooKeeperBytesSent", "ZooKeeperBytesReceived", "DistributedConnectionFailTry", "DistributedConnectionMissingTable", "DistributedConnectionStaleReplica", "DistributedConnectionFailAtAll", "HedgedRequestsChangeReplica", "SuspendSendingQueryToShard", "CompileFunction", "CompiledFunctionExecute", "CompileExpressionsMicroseconds", "CompileExpressionsBytes", "ExecuteShellCommand", "ExternalProcessingCompressedBytesTotal", "ExternalProcessingUncompressedBytesTotal", "ExternalProcessingFilesTotal", "ExternalSortWritePart", "ExternalSortMerge", "ExternalSortCompressedBytes", "ExternalSortUncompressedBytes", "ExternalAggregationWritePart", "ExternalAggregationMerge", "ExternalAggregationCompressedBytes", "ExternalAggregationUncompressedBytes", "ExternalJoinWritePart", "ExternalJoinMerge", "ExternalJoinCompressedBytes", "ExternalJoinUncompressedBytes", "SlowRead", "ReadBackoff", "ReplicaPartialShutdown", "SelectedParts", "SelectedRanges", "SelectedMarks", "SelectedRows", "SelectedBytes", "WaitMarksLoadMicroseconds", "BackgroundLoadingMarksTasks", "LoadedMarksCount", "LoadedMarksMemoryBytes", "Merge", "MergedRows", "MergedUncompressedBytes", "MergesTimeMilliseconds", "MergeTreeDataWriterRows", "MergeTreeDataWriterUncompressedBytes", "MergeTreeDataWriterCompressedBytes", "MergeTreeDataWriterBlocks", "MergeTreeDataWriterBlocksAlreadySorted", "InsertedWideParts", "InsertedCompactParts", "InsertedInMemoryParts", "MergedIntoWideParts", "MergedIntoCompactParts", "MergedIntoInMemoryParts", "MergeTreeDataProjectionWriterRows", "MergeTreeDataProjectionWriterUncompressedBytes", "MergeTreeDataProjectionWriterCompressedBytes", "MergeTreeDataProjectionWriterBlocks", "MergeTreeDataProjectionWriterBlocksAlreadySorted", "CannotRemoveEphemeralNode", "RegexpCreated", "ContextLock", "StorageBufferFlush", "StorageBufferErrorOnFlush", "StorageBufferPassedAllMinThresholds", "StorageBufferPassedTimeMaxThreshold", "StorageBufferPassedRowsMaxThreshold", "StorageBufferPassedBytesMaxThreshold", "StorageBufferPassedTimeFlushThreshold", "StorageBufferPassedRowsFlushThreshold", "StorageBufferPassedBytesFlushThreshold", "StorageBufferLayerLockReadersWaitMilliseconds", "StorageBufferLayerLockWritersWaitMilliseconds", "DictCacheKeysRequested", "DictCacheKeysRequestedMiss", "DictCacheKeysRequestedFound", "DictCacheKeysExpired", "DictCacheKeysNotFound", "DictCacheKeysHit", "DictCacheRequestTimeNs", "DictCacheRequests", "DictCacheLockWriteNs", "DictCacheLockReadNs", "DistributedSyncInsertionTimeoutExceeded", "DataAfterMergeDiffersFromReplica", "DataAfterMutationDiffersFromReplica", "PolygonsAddedToPool", "PolygonsInPoolAllocatedBytes", "RWLockAcquiredReadLocks", "RWLockAcquiredWriteLocks", "RWLockReadersWaitMilliseconds", "RWLockWritersWaitMilliseconds", "DNSError", "RealTimeMicroseconds", "UserTimeMicroseconds", "SystemTimeMicroseconds", "MemoryOvercommitWaitTimeMicroseconds", "MemoryAllocatorPurge", "MemoryAllocatorPurgeTimeMicroseconds", "SoftPageFaults", "HardPageFaults", "OSIOWaitMicroseconds", "OSCPUWaitMicroseconds", "OSCPUVirtualTimeMicroseconds", "OSReadBytes", "OSWriteBytes", "OSReadChars", "OSWriteChars", "PerfCpuCycles", "PerfInstructions", "PerfCacheReferences", "PerfCacheMisses", "PerfBranchInstructions", "PerfBranchMisses", "PerfBusCycles", "PerfStalledCyclesFrontend", "PerfStalledCyclesBackend", "PerfRefCpuCycles", "PerfCpuClock", "PerfTaskClock", "PerfContextSwitches", "PerfCpuMigrations", "PerfAlignmentFaults", "PerfEmulationFaults", "PerfMinEnabledTime", "PerfMinEnabledRunningTime", "PerfDataTLBReferences", "PerfDataTLBMisses", "PerfInstructionTLBReferences", "PerfInstructionTLBMisses", "PerfLocalMemoryReferences", "PerfLocalMemoryMisses", "CreatedHTTPConnections", "CannotWriteToWriteBufferDiscard", "QueryProfilerSignalOverruns", "QueryProfilerRuns", "CreatedLogEntryForMerge", "NotCreatedLogEntryForMerge", "CreatedLogEntryForMutation", "NotCreatedLogEntryForMutation", "S3ReadMicroseconds", "S3ReadRequestsCount", "S3ReadRequestsErrors", "S3ReadRequestsThrottling", "S3ReadRequestsRedirects", "S3WriteMicroseconds", "S3WriteRequestsCount", "S3WriteRequestsErrors", "S3WriteRequestsThrottling", "S3WriteRequestsRedirects", "DiskS3ReadMicroseconds", "DiskS3ReadRequestsCount", "DiskS3ReadRequestsErrors", "DiskS3ReadRequestsThrottling", "DiskS3ReadRequestsRedirects", "DiskS3WriteMicroseconds", "DiskS3WriteRequestsCount", "DiskS3WriteRequestsErrors", "DiskS3WriteRequestsThrottling", "DiskS3WriteRequestsRedirects", "S3DeleteObjects", "S3CopyObject", "S3ListObjects", "S3HeadObject", "S3GetObjectAttributes", "S3CreateMultipartUpload", "S3UploadPartCopy", "S3UploadPart", "S3AbortMultipartUpload", "S3CompleteMultipartUpload", "S3PutObject", "S3GetObject", "DiskS3DeleteObjects", "DiskS3CopyObject", "DiskS3ListObjects", "DiskS3HeadObject", "DiskS3GetObjectAttributes", "DiskS3CreateMultipartUpload", "DiskS3UploadPartCopy", "DiskS3UploadPart", "DiskS3AbortMultipartUpload", "DiskS3CompleteMultipartUpload", "DiskS3PutObject", "DiskS3GetObject", "ReadBufferFromS3Microseconds", "ReadBufferFromS3InitMicroseconds", "ReadBufferFromS3Bytes", "ReadBufferFromS3RequestsErrors", "WriteBufferFromS3Microseconds", "WriteBufferFromS3Bytes", "WriteBufferFromS3RequestsErrors", "QueryMemoryLimitExceeded", "CachedReadBufferReadFromSourceMicroseconds", "CachedReadBufferReadFromCacheMicroseconds", "CachedReadBufferReadFromSourceBytes", "CachedReadBufferReadFromCacheBytes", "CachedReadBufferCacheWriteBytes", "CachedReadBufferCacheWriteMicroseconds", "CachedWriteBufferCacheWriteBytes", "CachedWriteBufferCacheWriteMicroseconds", "RemoteFSSeeks", "RemoteFSPrefetches", "RemoteFSCancelledPrefetches", "RemoteFSUnusedPrefetches", "RemoteFSPrefetchedReads", "RemoteFSPrefetchedBytes", "RemoteFSUnprefetchedReads", "RemoteFSUnprefetchedBytes", "RemoteFSLazySeeks", "RemoteFSSeeksWithReset", "RemoteFSBuffers", "MergeTreePrefetchedReadPoolInit", "WaitPrefetchTaskMicroseconds", "ThreadpoolReaderTaskMicroseconds", "ThreadpoolReaderReadBytes", "ThreadpoolReaderSubmit", "FileSegmentWaitReadBufferMicroseconds", "FileSegmentReadMicroseconds", "FileSegmentWriteMicroseconds", "FileSegmentCacheWriteMicroseconds", "FileSegmentPredownloadMicroseconds", "FileSegmentUsedBytes", "ReadBufferSeekCancelConnection", "SleepFunctionCalls", "SleepFunctionMicroseconds", "ThreadPoolReaderPageCacheHit", "ThreadPoolReaderPageCacheHitBytes", "ThreadPoolReaderPageCacheHitElapsedMicroseconds", "ThreadPoolReaderPageCacheMiss", "ThreadPoolReaderPageCacheMissBytes", "ThreadPoolReaderPageCacheMissElapsedMicroseconds", "AsynchronousReadWaitMicroseconds", "AsynchronousRemoteReadWaitMicroseconds", "SynchronousRemoteReadWaitMicroseconds", "ExternalDataSourceLocalCacheReadBytes", "MainConfigLoads", "AggregationPreallocatedElementsInHashTables", "AggregationHashTablesInitializedAsTwoLevel", "MergeTreeMetadataCacheGet", "MergeTreeMetadataCachePut", "MergeTreeMetadataCacheDelete", "MergeTreeMetadataCacheSeek", "MergeTreeMetadataCacheHit", "MergeTreeMetadataCacheMiss", "KafkaRebalanceRevocations", "KafkaRebalanceAssignments", "KafkaRebalanceErrors", "KafkaMessagesPolled", "KafkaMessagesRead", "KafkaMessagesFailed", "KafkaRowsRead", "KafkaRowsRejected", "KafkaDirectReads", "KafkaBackgroundReads", "KafkaCommits", "KafkaCommitFailures", "KafkaConsumerErrors", "KafkaWrites", "KafkaRowsWritten", "KafkaProducerFlushes", "KafkaMessagesProduced", "KafkaProducerErrors", "ScalarSubqueriesGlobalCacheHit", "ScalarSubqueriesLocalCacheHit", "ScalarSubqueriesCacheMiss", "SchemaInferenceCacheHits", "SchemaInferenceCacheMisses", "SchemaInferenceCacheEvictions", "SchemaInferenceCacheInvalidations", "KeeperPacketsSent", "KeeperPacketsReceived", "KeeperRequestTotal", "KeeperLatency", "KeeperCommits", "KeeperCommitsFailed", "KeeperSnapshotCreations", "KeeperSnapshotCreationsFailed", "KeeperSnapshotApplys", "KeeperSnapshotApplysFailed", "KeeperReadSnapshot", "KeeperSaveSnapshot", "KeeperCreateRequest", "KeeperRemoveRequest", "KeeperSetRequest", "KeeperCheckRequest", "KeeperMultiRequest", "KeeperMultiReadRequest", "KeeperGetRequest", "KeeperListRequest", "KeeperExistsRequest", "OverflowBreak", "OverflowThrow", "OverflowAny", "ServerStartupMilliseconds", "IOUringSQEsSubmitted", "IOUringSQEsResubmits", "IOUringCQEsCompleted", "IOUringCQEsFailed", "ReadTaskRequestsReceived", "MergeTreeReadTaskRequestsReceived", "ReadTaskRequestsSent", "MergeTreeReadTaskRequestsSent", "MergeTreeAllRangesAnnouncementsSent", "ReadTaskRequestsSentElapsedMicroseconds", "MergeTreeReadTaskRequestsSentElapsedMicroseconds", "MergeTreeAllRangesAnnouncementsSentElapsedMicroseconds", "LogTest", "LogTrace", "LogDebug", "LogInfo", "LogWarning", "LogError", "LogFatal", "CreatedReadBufferAIOFailed", "ReadBufferAIORead", "ReadBufferAIOReadBytes", "VoluntaryContextSwitches", "WriteBufferAIOWrite", "WriteBufferAIOWriteBytes", "CreatedReadBufferAIO", "InvoluntaryContextSwitches", "S3ReadBytes", "S3WriteBytes"), // nolint:lll
			},
			optsMetrics: []inputs.PointCheckOption{
				inputs.WithTypeChecking(false),
				inputs.WithExtraTags(map[string]string{"instance": "", "tag1": "", "tag2": ""}),
				inputs.WithOptionalFields("Merge", "Move", "PartMutation", "ReplicatedFetch", "ReplicatedSend", "ReplicatedChecks", "BackgroundMergesAndMutationsPoolTask", "BackgroundMergesAndMutationsPoolSize", "BackgroundFetchesPoolTask", "BackgroundFetchesPoolSize", "BackgroundCommonPoolTask", "BackgroundCommonPoolSize", "BackgroundMovePoolTask", "BackgroundMovePoolSize", "BackgroundSchedulePoolTask", "BackgroundSchedulePoolSize", "BackgroundBufferFlushSchedulePoolTask", "BackgroundBufferFlushSchedulePoolSize", "BackgroundDistributedSchedulePoolTask", "BackgroundDistributedSchedulePoolSize", "BackgroundMessageBrokerSchedulePoolTask", "BackgroundMessageBrokerSchedulePoolSize", "CacheDictionaryUpdateQueueBatches", "CacheDictionaryUpdateQueueKeys", "DiskSpaceReservedForMerge", "DistributedSend", "QueryPreempted", "TCPConnection", "MySQLConnection", "HTTPConnection", "InterserverConnection", "PostgreSQLConnection", "OpenFileForRead", "OpenFileForWrite", "TotalTemporaryFiles", "TemporaryFilesForSort", "TemporaryFilesForAggregation", "TemporaryFilesForJoin", "TemporaryFilesUnknown", "Read", "RemoteRead", "Write", "NetworkReceive", "NetworkSend", "SendScalars", "SendExternalTables", "QueryThread", "ReadonlyReplica", "MemoryTracking", "EphemeralNode", "ZooKeeperSession", "ZooKeeperWatch", "ZooKeeperRequest", "DelayedInserts", "ContextLockWait", "StorageBufferRows", "StorageBufferBytes", "DictCacheRequests", "Revision", "VersionInteger", "RWLockWaitingReaders", "RWLockWaitingWriters", "RWLockActiveReaders", "RWLockActiveWriters", "GlobalThread", "GlobalThreadActive", "LocalThread", "LocalThreadActive", "MergeTreeDataSelectExecutorThreads", "MergeTreeDataSelectExecutorThreadsActive", "BackupsThreads", "BackupsThreadsActive", "RestoreThreads", "RestoreThreadsActive", "MarksLoaderThreads", "MarksLoaderThreadsActive", "IOPrefetchThreads", "IOPrefetchThreadsActive", "IOWriterThreads", "IOWriterThreadsActive", "IOThreads", "IOThreadsActive", "ThreadPoolRemoteFSReaderThreads", "ThreadPoolRemoteFSReaderThreadsActive", "ThreadPoolFSReaderThreads", "ThreadPoolFSReaderThreadsActive", "BackupsIOThreads", "BackupsIOThreadsActive", "DiskObjectStorageAsyncThreads", "DiskObjectStorageAsyncThreadsActive", "StorageHiveThreads", "StorageHiveThreadsActive", "TablesLoaderThreads", "TablesLoaderThreadsActive", "DatabaseOrdinaryThreads", "DatabaseOrdinaryThreadsActive", "DatabaseOnDiskThreads", "DatabaseOnDiskThreadsActive", "DatabaseCatalogThreads", "DatabaseCatalogThreadsActive", "DestroyAggregatesThreads", "DestroyAggregatesThreadsActive", "HashedDictionaryThreads", "HashedDictionaryThreadsActive", "CacheDictionaryThreads", "CacheDictionaryThreadsActive", "ParallelFormattingOutputFormatThreads", "ParallelFormattingOutputFormatThreadsActive", "ParallelParsingInputFormatThreads", "ParallelParsingInputFormatThreadsActive", "MergeTreeBackgroundExecutorThreads", "MergeTreeBackgroundExecutorThreadsActive", "AsynchronousInsertThreads", "AsynchronousInsertThreadsActive", "StartupSystemTablesThreads", "StartupSystemTablesThreadsActive", "AggregatorThreads", "AggregatorThreadsActive", "DDLWorkerThreads", "DDLWorkerThreadsActive", "StorageDistributedThreads", "StorageDistributedThreadsActive", "DistributedInsertThreads", "DistributedInsertThreadsActive", "StorageS3Threads", "StorageS3ThreadsActive", "MergeTreePartsLoaderThreads", "MergeTreePartsLoaderThreadsActive", "MergeTreePartsCleanerThreads", "MergeTreePartsCleanerThreadsActive", "SystemReplicasThreads", "SystemReplicasThreadsActive", "RestartReplicaThreads", "RestartReplicaThreadsActive", "QueryPipelineExecutorThreads", "QueryPipelineExecutorThreadsActive", "ParquetDecoderThreads", "ParquetDecoderThreadsActive", "DistributedFilesToInsert", "BrokenDistributedFilesToInsert", "TablesToDropQueueSize", "MaxDDLEntryID", "MaxPushedDDLEntryID", "PartsTemporary", "PartsPreCommitted", "PartsCommitted", "PartsPreActive", "PartsActive", "PartsOutdated", "PartsDeleting", "PartsDeleteOnDestroy", "PartsWide", "PartsCompact", "PartsInMemory", "MMappedFiles", "MMappedFileBytes", "MMappedAllocs", "MMappedAllocBytes", "AsynchronousReadWait", "PendingAsyncInsert", "KafkaConsumers", "KafkaConsumersWithAssignment", "KafkaProducers", "KafkaLibrdkafkaThreads", "KafkaBackgroundReads", "KafkaConsumersInUse", "KafkaWrites", "KafkaAssignedPartitions", "FilesystemCacheReadBuffers", "CacheFileSegments", "CacheDetachedFileSegments", "FilesystemCacheSize", "FilesystemCacheElements", "AsyncInsertCacheSize", "S3Requests", "KeeperAliveConnections", "KeeperOutstandingRequets", "ThreadsInOvercommitTracker", "IOUringPendingEvents", "IOUringInFlightEvents", "ReadTaskRequestsSent", "MergeTreeReadTaskRequestsSent", "MergeTreeAllRangesAnnouncementsSent", "SyncDrainedConnections", "Query", "ActiveAsyncDrainedConnections", "ActiveSyncDrainedConnections", "AsyncDrainedConnections", "BackgroundPoolTask"), // nolint:lll
			},
			optsAsyncMetrics: []inputs.PointCheckOption{
				inputs.WithTypeChecking(false),
				inputs.WithExtraTags(map[string]string{"instance": "", "tag1": "", "tag2": ""}),
				inputs.WithOptionalTags("cpu", "disk", "eth", "unit"),
				inputs.WithOptionalFields("AsynchronousHeavyMetricsCalculationTimeSpent", "AsynchronousHeavyMetricsUpdateInterval", "AsynchronousMetricsCalculationTimeSpent", "AsynchronousMetricsUpdateInterval", "BlockActiveTime", "BlockDiscardBytes", "BlockDiscardMerges", "BlockDiscardOps", "BlockDiscardTime", "BlockInFlightOps", "BlockQueueTime", "BlockReadBytes", "BlockReadMerges", "BlockReadOps", "BlockReadTime", "BlockWriteBytes", "BlockWriteMerges", "BlockWriteOps", "BlockWriteTime", "CPUFrequencyMHz", "CompiledExpressionCacheBytes", "CompiledExpressionCacheCount", "DiskAvailable", "DiskTotal", "DiskUnreserved", "DiskUsed", "FilesystemCacheBytes", "FilesystemCacheFiles", "FilesystemLogsPathAvailableBytes", "FilesystemLogsPathAvailableINodes", "FilesystemLogsPathTotalBytes", "FilesystemLogsPathTotalINodes", "FilesystemLogsPathUsedBytes", "FilesystemLogsPathUsedINodes", "FilesystemMainPathAvailableBytes", "FilesystemMainPathAvailableINodes", "FilesystemMainPathTotalBytes", "FilesystemMainPathTotalINodes", "FilesystemMainPathUsedBytes", "FilesystemMainPathUsedINodes", "HTTPThreads", "InterserverThreads", "Jitter", "LoadAverage", "MMapCacheCells", "MarkCacheBytes", "MarkCacheFiles", "MaxPartCountForPartition", "MemoryCode", "MemoryDataAndStack", "MemoryResident", "MemoryShared", "MemoryVirtual", "MySQLThreads", "NetworkReceiveBytes", "NetworkReceiveDrop", "NetworkReceiveErrors", "NetworkReceivePackets", "NetworkSendBytes", "NetworkSendDrop", "NetworkSendErrors", "NetworkSendPackets", "NumberOfDatabases", "NumberOfDetachedByUserParts", "NumberOfDetachedParts", "NumberOfTables", "OSContextSwitches", "OSGuestNiceTime", "OSGuestNiceTimeCPU", "OSGuestNiceTimeNormalized", "OSGuestTime", "OSGuestTimeCPU", "OSGuestTimeNormalized", "OSIOWaitTime", "OSIOWaitTimeCPU", "OSIOWaitTimeNormalized", "OSIdleTime", "OSIdleTimeCPU", "OSIdleTimeNormalized", "OSInterrupts", "OSIrqTime", "OSIrqTimeCPU", "OSIrqTimeNormalized", "OSMemoryAvailable", "OSMemoryBuffers", "OSMemoryCached", "OSMemoryFreePlusCached", "OSMemoryFreeWithoutCached", "OSMemoryTotal", "OSNiceTime", "OSNiceTimeCPU", "OSNiceTimeNormalized", "OSOpenFiles", "OSProcessesBlocked", "OSProcessesCreated", "OSProcessesRunning", "OSSoftIrqTime", "OSSoftIrqTimeCPU", "OSSoftIrqTimeNormalized", "OSStealTime", "OSStealTimeCPU", "OSStealTimeNormalized", "OSSystemTime", "OSSystemTimeCPU", "OSSystemTimeNormalized", "OSThreadsRunnable", "OSThreadsTotal", "OSUptime", "OSUserTime", "OSUserTimeCPU", "OSUserTimeNormalized", "PostgreSQLThreads", "ReplicasMaxAbsoluteDelay", "ReplicasMaxInsertsInQueue", "ReplicasMaxMergesInQueue", "ReplicasMaxQueueSize", "ReplicasMaxRelativeDelay", "ReplicasSumInsertsInQueue", "ReplicasSumMergesInQueue", "ReplicasSumQueueSize", "TCPThreads", "Temperature", "TotalBytesOfMergeTreeTables", "TotalPartsOfMergeTreeTables", "TotalRowsOfMergeTreeTables", "UncompressedCacheBytes", "UncompressedCacheCells", "Uptime", "jemalloc_active", "jemalloc_allocated", "jemalloc_arenas_all_dirty_purged", "jemalloc_arenas_all_muzzy_purged", "jemalloc_arenas_all_pactive", "jemalloc_arenas_all_pdirty", "jemalloc_arenas_all_pmuzzy", "jemalloc_background_thread_num_runs", "jemalloc_background_thread_num_threads", "jemalloc_background_thread_run_intervals", "jemalloc_epoch", "jemalloc_mapped", "jemalloc_metadata", "jemalloc_metadata_thp", "jemalloc_resident", "jemalloc_retained", "OSMemorySwapCached", "PrometheusThreads"), // nolint:lll
			},
			optsStatusInfo: []inputs.PointCheckOption{},
		},
		{
			name:    "remote-clickhouse",
			repo:    "pubrepo.jiagouyun.com/image-repo-for-testing/clickhouse/clickhouse-server",
			repoTag: "21.8.15.7",
			conf: fmt.Sprintf(`
source = "clickhouse"
# metric_types = ["counter", "gauge"]

interval = "10s"
tls_open = false
urls = ["http://%s/metrics"]

`, net.JoinHostPort(remote.Host, fmt.Sprintf("%d", testutils.RandPort("tcp")))),
			optsProfileEvents: []inputs.PointCheckOption{
				inputs.WithTypeChecking(false),
				inputs.WithExtraTags(map[string]string{"instance": ""}),
				inputs.WithOptionalFields("Query", "SelectQuery", "InsertQuery", "AsyncInsertQuery", "AsyncInsertBytes", "AsyncInsertCacheHits", "FailedQuery", "FailedSelectQuery", "FailedInsertQuery", "FailedAsyncInsertQuery", "QueryTimeMicroseconds", "SelectQueryTimeMicroseconds", "InsertQueryTimeMicroseconds", "OtherQueryTimeMicroseconds", "FileOpen", "Seek", "ReadBufferFromFileDescriptorRead", "ReadBufferFromFileDescriptorReadFailed", "ReadBufferFromFileDescriptorReadBytes", "WriteBufferFromFileDescriptorWrite", "WriteBufferFromFileDescriptorWriteFailed", "WriteBufferFromFileDescriptorWriteBytes", "FileSync", "DirectorySync", "FileSyncElapsedMicroseconds", "DirectorySyncElapsedMicroseconds", "ReadCompressedBytes", "CompressedReadBufferBlocks", "CompressedReadBufferBytes", "UncompressedCacheHits", "UncompressedCacheMisses", "UncompressedCacheWeightLost", "MMappedFileCacheHits", "MMappedFileCacheMisses", "OpenedFileCacheHits", "OpenedFileCacheMisses", "AIOWrite", "AIOWriteBytes", "AIORead", "AIOReadBytes", "IOBufferAllocs", "IOBufferAllocBytes", "ArenaAllocChunks", "ArenaAllocBytes", "FunctionExecute", "TableFunctionExecute", "MarkCacheHits", "MarkCacheMisses", "QueryCacheHits", "QueryCacheMisses", "CreatedReadBufferOrdinary", "CreatedReadBufferDirectIO", "CreatedReadBufferDirectIOFailed", "CreatedReadBufferMMap", "CreatedReadBufferMMapFailed", "DiskReadElapsedMicroseconds", "DiskWriteElapsedMicroseconds", "NetworkReceiveElapsedMicroseconds", "NetworkSendElapsedMicroseconds", "NetworkReceiveBytes", "NetworkSendBytes", "DiskS3GetRequestThrottlerCount", "DiskS3GetRequestThrottlerSleepMicroseconds", "DiskS3PutRequestThrottlerCount", "DiskS3PutRequestThrottlerSleepMicroseconds", "S3GetRequestThrottlerCount", "S3GetRequestThrottlerSleepMicroseconds", "S3PutRequestThrottlerCount", "S3PutRequestThrottlerSleepMicroseconds", "RemoteReadThrottlerBytes", "RemoteReadThrottlerSleepMicroseconds", "RemoteWriteThrottlerBytes", "RemoteWriteThrottlerSleepMicroseconds", "LocalReadThrottlerBytes", "LocalReadThrottlerSleepMicroseconds", "LocalWriteThrottlerBytes", "LocalWriteThrottlerSleepMicroseconds", "ThrottlerSleepMicroseconds", "QueryMaskingRulesMatch", "ReplicatedPartFetches", "ReplicatedPartFailedFetches", "ObsoleteReplicatedParts", "ReplicatedPartMerges", "ReplicatedPartFetchesOfMerged", "ReplicatedPartMutations", "ReplicatedPartChecks", "ReplicatedPartChecksFailed", "ReplicatedDataLoss", "InsertedRows", "InsertedBytes", "DelayedInserts", "RejectedInserts", "DelayedInsertsMilliseconds", "DistributedDelayedInserts", "DistributedRejectedInserts", "DistributedDelayedInsertsMilliseconds", "DuplicatedInsertedBlocks", "ZooKeeperInit", "ZooKeeperTransactions", "ZooKeeperList", "ZooKeeperCreate", "ZooKeeperRemove", "ZooKeeperExists", "ZooKeeperGet", "ZooKeeperSet", "ZooKeeperMulti", "ZooKeeperCheck", "ZooKeeperSync", "ZooKeeperClose", "ZooKeeperWatchResponse", "ZooKeeperUserExceptions", "ZooKeeperHardwareExceptions", "ZooKeeperOtherExceptions", "ZooKeeperWaitMicroseconds", "ZooKeeperBytesSent", "ZooKeeperBytesReceived", "DistributedConnectionFailTry", "DistributedConnectionMissingTable", "DistributedConnectionStaleReplica", "DistributedConnectionFailAtAll", "HedgedRequestsChangeReplica", "SuspendSendingQueryToShard", "CompileFunction", "CompiledFunctionExecute", "CompileExpressionsMicroseconds", "CompileExpressionsBytes", "ExecuteShellCommand", "ExternalProcessingCompressedBytesTotal", "ExternalProcessingUncompressedBytesTotal", "ExternalProcessingFilesTotal", "ExternalSortWritePart", "ExternalSortMerge", "ExternalSortCompressedBytes", "ExternalSortUncompressedBytes", "ExternalAggregationWritePart", "ExternalAggregationMerge", "ExternalAggregationCompressedBytes", "ExternalAggregationUncompressedBytes", "ExternalJoinWritePart", "ExternalJoinMerge", "ExternalJoinCompressedBytes", "ExternalJoinUncompressedBytes", "SlowRead", "ReadBackoff", "ReplicaPartialShutdown", "SelectedParts", "SelectedRanges", "SelectedMarks", "SelectedRows", "SelectedBytes", "WaitMarksLoadMicroseconds", "BackgroundLoadingMarksTasks", "LoadedMarksCount", "LoadedMarksMemoryBytes", "Merge", "MergedRows", "MergedUncompressedBytes", "MergesTimeMilliseconds", "MergeTreeDataWriterRows", "MergeTreeDataWriterUncompressedBytes", "MergeTreeDataWriterCompressedBytes", "MergeTreeDataWriterBlocks", "MergeTreeDataWriterBlocksAlreadySorted", "InsertedWideParts", "InsertedCompactParts", "InsertedInMemoryParts", "MergedIntoWideParts", "MergedIntoCompactParts", "MergedIntoInMemoryParts", "MergeTreeDataProjectionWriterRows", "MergeTreeDataProjectionWriterUncompressedBytes", "MergeTreeDataProjectionWriterCompressedBytes", "MergeTreeDataProjectionWriterBlocks", "MergeTreeDataProjectionWriterBlocksAlreadySorted", "CannotRemoveEphemeralNode", "RegexpCreated", "ContextLock", "StorageBufferFlush", "StorageBufferErrorOnFlush", "StorageBufferPassedAllMinThresholds", "StorageBufferPassedTimeMaxThreshold", "StorageBufferPassedRowsMaxThreshold", "StorageBufferPassedBytesMaxThreshold", "StorageBufferPassedTimeFlushThreshold", "StorageBufferPassedRowsFlushThreshold", "StorageBufferPassedBytesFlushThreshold", "StorageBufferLayerLockReadersWaitMilliseconds", "StorageBufferLayerLockWritersWaitMilliseconds", "DictCacheKeysRequested", "DictCacheKeysRequestedMiss", "DictCacheKeysRequestedFound", "DictCacheKeysExpired", "DictCacheKeysNotFound", "DictCacheKeysHit", "DictCacheRequestTimeNs", "DictCacheRequests", "DictCacheLockWriteNs", "DictCacheLockReadNs", "DistributedSyncInsertionTimeoutExceeded", "DataAfterMergeDiffersFromReplica", "DataAfterMutationDiffersFromReplica", "PolygonsAddedToPool", "PolygonsInPoolAllocatedBytes", "RWLockAcquiredReadLocks", "RWLockAcquiredWriteLocks", "RWLockReadersWaitMilliseconds", "RWLockWritersWaitMilliseconds", "DNSError", "RealTimeMicroseconds", "UserTimeMicroseconds", "SystemTimeMicroseconds", "MemoryOvercommitWaitTimeMicroseconds", "MemoryAllocatorPurge", "MemoryAllocatorPurgeTimeMicroseconds", "SoftPageFaults", "HardPageFaults", "OSIOWaitMicroseconds", "OSCPUWaitMicroseconds", "OSCPUVirtualTimeMicroseconds", "OSReadBytes", "OSWriteBytes", "OSReadChars", "OSWriteChars", "PerfCpuCycles", "PerfInstructions", "PerfCacheReferences", "PerfCacheMisses", "PerfBranchInstructions", "PerfBranchMisses", "PerfBusCycles", "PerfStalledCyclesFrontend", "PerfStalledCyclesBackend", "PerfRefCpuCycles", "PerfCpuClock", "PerfTaskClock", "PerfContextSwitches", "PerfCpuMigrations", "PerfAlignmentFaults", "PerfEmulationFaults", "PerfMinEnabledTime", "PerfMinEnabledRunningTime", "PerfDataTLBReferences", "PerfDataTLBMisses", "PerfInstructionTLBReferences", "PerfInstructionTLBMisses", "PerfLocalMemoryReferences", "PerfLocalMemoryMisses", "CreatedHTTPConnections", "CannotWriteToWriteBufferDiscard", "QueryProfilerSignalOverruns", "QueryProfilerRuns", "CreatedLogEntryForMerge", "NotCreatedLogEntryForMerge", "CreatedLogEntryForMutation", "NotCreatedLogEntryForMutation", "S3ReadMicroseconds", "S3ReadRequestsCount", "S3ReadRequestsErrors", "S3ReadRequestsThrottling", "S3ReadRequestsRedirects", "S3WriteMicroseconds", "S3WriteRequestsCount", "S3WriteRequestsErrors", "S3WriteRequestsThrottling", "S3WriteRequestsRedirects", "DiskS3ReadMicroseconds", "DiskS3ReadRequestsCount", "DiskS3ReadRequestsErrors", "DiskS3ReadRequestsThrottling", "DiskS3ReadRequestsRedirects", "DiskS3WriteMicroseconds", "DiskS3WriteRequestsCount", "DiskS3WriteRequestsErrors", "DiskS3WriteRequestsThrottling", "DiskS3WriteRequestsRedirects", "S3DeleteObjects", "S3CopyObject", "S3ListObjects", "S3HeadObject", "S3GetObjectAttributes", "S3CreateMultipartUpload", "S3UploadPartCopy", "S3UploadPart", "S3AbortMultipartUpload", "S3CompleteMultipartUpload", "S3PutObject", "S3GetObject", "DiskS3DeleteObjects", "DiskS3CopyObject", "DiskS3ListObjects", "DiskS3HeadObject", "DiskS3GetObjectAttributes", "DiskS3CreateMultipartUpload", "DiskS3UploadPartCopy", "DiskS3UploadPart", "DiskS3AbortMultipartUpload", "DiskS3CompleteMultipartUpload", "DiskS3PutObject", "DiskS3GetObject", "ReadBufferFromS3Microseconds", "ReadBufferFromS3InitMicroseconds", "ReadBufferFromS3Bytes", "ReadBufferFromS3RequestsErrors", "WriteBufferFromS3Microseconds", "WriteBufferFromS3Bytes", "WriteBufferFromS3RequestsErrors", "QueryMemoryLimitExceeded", "CachedReadBufferReadFromSourceMicroseconds", "CachedReadBufferReadFromCacheMicroseconds", "CachedReadBufferReadFromSourceBytes", "CachedReadBufferReadFromCacheBytes", "CachedReadBufferCacheWriteBytes", "CachedReadBufferCacheWriteMicroseconds", "CachedWriteBufferCacheWriteBytes", "CachedWriteBufferCacheWriteMicroseconds", "RemoteFSSeeks", "RemoteFSPrefetches", "RemoteFSCancelledPrefetches", "RemoteFSUnusedPrefetches", "RemoteFSPrefetchedReads", "RemoteFSPrefetchedBytes", "RemoteFSUnprefetchedReads", "RemoteFSUnprefetchedBytes", "RemoteFSLazySeeks", "RemoteFSSeeksWithReset", "RemoteFSBuffers", "MergeTreePrefetchedReadPoolInit", "WaitPrefetchTaskMicroseconds", "ThreadpoolReaderTaskMicroseconds", "ThreadpoolReaderReadBytes", "ThreadpoolReaderSubmit", "FileSegmentWaitReadBufferMicroseconds", "FileSegmentReadMicroseconds", "FileSegmentWriteMicroseconds", "FileSegmentCacheWriteMicroseconds", "FileSegmentPredownloadMicroseconds", "FileSegmentUsedBytes", "ReadBufferSeekCancelConnection", "SleepFunctionCalls", "SleepFunctionMicroseconds", "ThreadPoolReaderPageCacheHit", "ThreadPoolReaderPageCacheHitBytes", "ThreadPoolReaderPageCacheHitElapsedMicroseconds", "ThreadPoolReaderPageCacheMiss", "ThreadPoolReaderPageCacheMissBytes", "ThreadPoolReaderPageCacheMissElapsedMicroseconds", "AsynchronousReadWaitMicroseconds", "AsynchronousRemoteReadWaitMicroseconds", "SynchronousRemoteReadWaitMicroseconds", "ExternalDataSourceLocalCacheReadBytes", "MainConfigLoads", "AggregationPreallocatedElementsInHashTables", "AggregationHashTablesInitializedAsTwoLevel", "MergeTreeMetadataCacheGet", "MergeTreeMetadataCachePut", "MergeTreeMetadataCacheDelete", "MergeTreeMetadataCacheSeek", "MergeTreeMetadataCacheHit", "MergeTreeMetadataCacheMiss", "KafkaRebalanceRevocations", "KafkaRebalanceAssignments", "KafkaRebalanceErrors", "KafkaMessagesPolled", "KafkaMessagesRead", "KafkaMessagesFailed", "KafkaRowsRead", "KafkaRowsRejected", "KafkaDirectReads", "KafkaBackgroundReads", "KafkaCommits", "KafkaCommitFailures", "KafkaConsumerErrors", "KafkaWrites", "KafkaRowsWritten", "KafkaProducerFlushes", "KafkaMessagesProduced", "KafkaProducerErrors", "ScalarSubqueriesGlobalCacheHit", "ScalarSubqueriesLocalCacheHit", "ScalarSubqueriesCacheMiss", "SchemaInferenceCacheHits", "SchemaInferenceCacheMisses", "SchemaInferenceCacheEvictions", "SchemaInferenceCacheInvalidations", "KeeperPacketsSent", "KeeperPacketsReceived", "KeeperRequestTotal", "KeeperLatency", "KeeperCommits", "KeeperCommitsFailed", "KeeperSnapshotCreations", "KeeperSnapshotCreationsFailed", "KeeperSnapshotApplys", "KeeperSnapshotApplysFailed", "KeeperReadSnapshot", "KeeperSaveSnapshot", "KeeperCreateRequest", "KeeperRemoveRequest", "KeeperSetRequest", "KeeperCheckRequest", "KeeperMultiRequest", "KeeperMultiReadRequest", "KeeperGetRequest", "KeeperListRequest", "KeeperExistsRequest", "OverflowBreak", "OverflowThrow", "OverflowAny", "ServerStartupMilliseconds", "IOUringSQEsSubmitted", "IOUringSQEsResubmits", "IOUringCQEsCompleted", "IOUringCQEsFailed", "ReadTaskRequestsReceived", "MergeTreeReadTaskRequestsReceived", "ReadTaskRequestsSent", "MergeTreeReadTaskRequestsSent", "MergeTreeAllRangesAnnouncementsSent", "ReadTaskRequestsSentElapsedMicroseconds", "MergeTreeReadTaskRequestsSentElapsedMicroseconds", "MergeTreeAllRangesAnnouncementsSentElapsedMicroseconds", "LogTest", "LogTrace", "LogDebug", "LogInfo", "LogWarning", "LogError", "LogFatal", "CreatedReadBufferAIOFailed", "ReadBufferAIORead", "ReadBufferAIOReadBytes", "VoluntaryContextSwitches", "WriteBufferAIOWrite", "WriteBufferAIOWriteBytes", "CreatedReadBufferAIO", "InvoluntaryContextSwitches", "S3ReadBytes", "S3WriteBytes"), // nolint:lll
			},
			optsMetrics: []inputs.PointCheckOption{
				inputs.WithTypeChecking(false),
				inputs.WithExtraTags(map[string]string{"instance": ""}),
				inputs.WithOptionalFields("Merge", "Move", "PartMutation", "ReplicatedFetch", "ReplicatedSend", "ReplicatedChecks", "BackgroundMergesAndMutationsPoolTask", "BackgroundMergesAndMutationsPoolSize", "BackgroundFetchesPoolTask", "BackgroundFetchesPoolSize", "BackgroundCommonPoolTask", "BackgroundCommonPoolSize", "BackgroundMovePoolTask", "BackgroundMovePoolSize", "BackgroundSchedulePoolTask", "BackgroundSchedulePoolSize", "BackgroundBufferFlushSchedulePoolTask", "BackgroundBufferFlushSchedulePoolSize", "BackgroundDistributedSchedulePoolTask", "BackgroundDistributedSchedulePoolSize", "BackgroundMessageBrokerSchedulePoolTask", "BackgroundMessageBrokerSchedulePoolSize", "CacheDictionaryUpdateQueueBatches", "CacheDictionaryUpdateQueueKeys", "DiskSpaceReservedForMerge", "DistributedSend", "QueryPreempted", "TCPConnection", "MySQLConnection", "HTTPConnection", "InterserverConnection", "PostgreSQLConnection", "OpenFileForRead", "OpenFileForWrite", "TotalTemporaryFiles", "TemporaryFilesForSort", "TemporaryFilesForAggregation", "TemporaryFilesForJoin", "TemporaryFilesUnknown", "Read", "RemoteRead", "Write", "NetworkReceive", "NetworkSend", "SendScalars", "SendExternalTables", "QueryThread", "ReadonlyReplica", "MemoryTracking", "EphemeralNode", "ZooKeeperSession", "ZooKeeperWatch", "ZooKeeperRequest", "DelayedInserts", "ContextLockWait", "StorageBufferRows", "StorageBufferBytes", "DictCacheRequests", "Revision", "VersionInteger", "RWLockWaitingReaders", "RWLockWaitingWriters", "RWLockActiveReaders", "RWLockActiveWriters", "GlobalThread", "GlobalThreadActive", "LocalThread", "LocalThreadActive", "MergeTreeDataSelectExecutorThreads", "MergeTreeDataSelectExecutorThreadsActive", "BackupsThreads", "BackupsThreadsActive", "RestoreThreads", "RestoreThreadsActive", "MarksLoaderThreads", "MarksLoaderThreadsActive", "IOPrefetchThreads", "IOPrefetchThreadsActive", "IOWriterThreads", "IOWriterThreadsActive", "IOThreads", "IOThreadsActive", "ThreadPoolRemoteFSReaderThreads", "ThreadPoolRemoteFSReaderThreadsActive", "ThreadPoolFSReaderThreads", "ThreadPoolFSReaderThreadsActive", "BackupsIOThreads", "BackupsIOThreadsActive", "DiskObjectStorageAsyncThreads", "DiskObjectStorageAsyncThreadsActive", "StorageHiveThreads", "StorageHiveThreadsActive", "TablesLoaderThreads", "TablesLoaderThreadsActive", "DatabaseOrdinaryThreads", "DatabaseOrdinaryThreadsActive", "DatabaseOnDiskThreads", "DatabaseOnDiskThreadsActive", "DatabaseCatalogThreads", "DatabaseCatalogThreadsActive", "DestroyAggregatesThreads", "DestroyAggregatesThreadsActive", "HashedDictionaryThreads", "HashedDictionaryThreadsActive", "CacheDictionaryThreads", "CacheDictionaryThreadsActive", "ParallelFormattingOutputFormatThreads", "ParallelFormattingOutputFormatThreadsActive", "ParallelParsingInputFormatThreads", "ParallelParsingInputFormatThreadsActive", "MergeTreeBackgroundExecutorThreads", "MergeTreeBackgroundExecutorThreadsActive", "AsynchronousInsertThreads", "AsynchronousInsertThreadsActive", "StartupSystemTablesThreads", "StartupSystemTablesThreadsActive", "AggregatorThreads", "AggregatorThreadsActive", "DDLWorkerThreads", "DDLWorkerThreadsActive", "StorageDistributedThreads", "StorageDistributedThreadsActive", "DistributedInsertThreads", "DistributedInsertThreadsActive", "StorageS3Threads", "StorageS3ThreadsActive", "MergeTreePartsLoaderThreads", "MergeTreePartsLoaderThreadsActive", "MergeTreePartsCleanerThreads", "MergeTreePartsCleanerThreadsActive", "SystemReplicasThreads", "SystemReplicasThreadsActive", "RestartReplicaThreads", "RestartReplicaThreadsActive", "QueryPipelineExecutorThreads", "QueryPipelineExecutorThreadsActive", "ParquetDecoderThreads", "ParquetDecoderThreadsActive", "DistributedFilesToInsert", "BrokenDistributedFilesToInsert", "TablesToDropQueueSize", "MaxDDLEntryID", "MaxPushedDDLEntryID", "PartsTemporary", "PartsPreCommitted", "PartsCommitted", "PartsPreActive", "PartsActive", "PartsOutdated", "PartsDeleting", "PartsDeleteOnDestroy", "PartsWide", "PartsCompact", "PartsInMemory", "MMappedFiles", "MMappedFileBytes", "MMappedAllocs", "MMappedAllocBytes", "AsynchronousReadWait", "PendingAsyncInsert", "KafkaConsumers", "KafkaConsumersWithAssignment", "KafkaProducers", "KafkaLibrdkafkaThreads", "KafkaBackgroundReads", "KafkaConsumersInUse", "KafkaWrites", "KafkaAssignedPartitions", "FilesystemCacheReadBuffers", "CacheFileSegments", "CacheDetachedFileSegments", "FilesystemCacheSize", "FilesystemCacheElements", "AsyncInsertCacheSize", "S3Requests", "KeeperAliveConnections", "KeeperOutstandingRequets", "ThreadsInOvercommitTracker", "IOUringPendingEvents", "IOUringInFlightEvents", "ReadTaskRequestsSent", "MergeTreeReadTaskRequestsSent", "MergeTreeAllRangesAnnouncementsSent", "SyncDrainedConnections", "Query", "ActiveAsyncDrainedConnections", "ActiveSyncDrainedConnections", "AsyncDrainedConnections", "BackgroundPoolTask"), // nolint:lll
			},
			optsAsyncMetrics: []inputs.PointCheckOption{
				inputs.WithTypeChecking(false),
				inputs.WithExtraTags(map[string]string{"instance": ""}),
				inputs.WithOptionalTags("cpu", "disk", "eth", "unit"),
				inputs.WithOptionalFields("AsynchronousHeavyMetricsCalculationTimeSpent", "AsynchronousHeavyMetricsUpdateInterval", "AsynchronousMetricsCalculationTimeSpent", "AsynchronousMetricsUpdateInterval", "BlockActiveTime", "BlockDiscardBytes", "BlockDiscardMerges", "BlockDiscardOps", "BlockDiscardTime", "BlockInFlightOps", "BlockQueueTime", "BlockReadBytes", "BlockReadMerges", "BlockReadOps", "BlockReadTime", "BlockWriteBytes", "BlockWriteMerges", "BlockWriteOps", "BlockWriteTime", "CPUFrequencyMHz", "CompiledExpressionCacheBytes", "CompiledExpressionCacheCount", "DiskAvailable", "DiskTotal", "DiskUnreserved", "DiskUsed", "FilesystemCacheBytes", "FilesystemCacheFiles", "FilesystemLogsPathAvailableBytes", "FilesystemLogsPathAvailableINodes", "FilesystemLogsPathTotalBytes", "FilesystemLogsPathTotalINodes", "FilesystemLogsPathUsedBytes", "FilesystemLogsPathUsedINodes", "FilesystemMainPathAvailableBytes", "FilesystemMainPathAvailableINodes", "FilesystemMainPathTotalBytes", "FilesystemMainPathTotalINodes", "FilesystemMainPathUsedBytes", "FilesystemMainPathUsedINodes", "HTTPThreads", "InterserverThreads", "Jitter", "LoadAverage", "MMapCacheCells", "MarkCacheBytes", "MarkCacheFiles", "MaxPartCountForPartition", "MemoryCode", "MemoryDataAndStack", "MemoryResident", "MemoryShared", "MemoryVirtual", "MySQLThreads", "NetworkReceiveBytes", "NetworkReceiveDrop", "NetworkReceiveErrors", "NetworkReceivePackets", "NetworkSendBytes", "NetworkSendDrop", "NetworkSendErrors", "NetworkSendPackets", "NumberOfDatabases", "NumberOfDetachedByUserParts", "NumberOfDetachedParts", "NumberOfTables", "OSContextSwitches", "OSGuestNiceTime", "OSGuestNiceTimeCPU", "OSGuestNiceTimeNormalized", "OSGuestTime", "OSGuestTimeCPU", "OSGuestTimeNormalized", "OSIOWaitTime", "OSIOWaitTimeCPU", "OSIOWaitTimeNormalized", "OSIdleTime", "OSIdleTimeCPU", "OSIdleTimeNormalized", "OSInterrupts", "OSIrqTime", "OSIrqTimeCPU", "OSIrqTimeNormalized", "OSMemoryAvailable", "OSMemoryBuffers", "OSMemoryCached", "OSMemoryFreePlusCached", "OSMemoryFreeWithoutCached", "OSMemoryTotal", "OSNiceTime", "OSNiceTimeCPU", "OSNiceTimeNormalized", "OSOpenFiles", "OSProcessesBlocked", "OSProcessesCreated", "OSProcessesRunning", "OSSoftIrqTime", "OSSoftIrqTimeCPU", "OSSoftIrqTimeNormalized", "OSStealTime", "OSStealTimeCPU", "OSStealTimeNormalized", "OSSystemTime", "OSSystemTimeCPU", "OSSystemTimeNormalized", "OSThreadsRunnable", "OSThreadsTotal", "OSUptime", "OSUserTime", "OSUserTimeCPU", "OSUserTimeNormalized", "PostgreSQLThreads", "ReplicasMaxAbsoluteDelay", "ReplicasMaxInsertsInQueue", "ReplicasMaxMergesInQueue", "ReplicasMaxQueueSize", "ReplicasMaxRelativeDelay", "ReplicasSumInsertsInQueue", "ReplicasSumMergesInQueue", "ReplicasSumQueueSize", "TCPThreads", "Temperature", "TotalBytesOfMergeTreeTables", "TotalPartsOfMergeTreeTables", "TotalRowsOfMergeTreeTables", "UncompressedCacheBytes", "UncompressedCacheCells", "Uptime", "jemalloc_active", "jemalloc_allocated", "jemalloc_arenas_all_dirty_purged", "jemalloc_arenas_all_muzzy_purged", "jemalloc_arenas_all_pactive", "jemalloc_arenas_all_pdirty", "jemalloc_arenas_all_pmuzzy", "jemalloc_background_thread_num_runs", "jemalloc_background_thread_num_threads", "jemalloc_background_thread_run_intervals", "jemalloc_epoch", "jemalloc_mapped", "jemalloc_metadata", "jemalloc_metadata_thp", "jemalloc_resident", "jemalloc_retained", "OSMemorySwapCached", "PrometheusThreads"), // nolint:lll
			},
			optsStatusInfo: []inputs.PointCheckOption{},
		},
	}

	perImageCfgs := []interface{}{}
	_ = perImageCfgs

	var cases []*caseSpec

	// compose cases
	for _, base := range bases {
		feeder := io.NewMockedFeeder()

		ipt := NewProm()    // This is real prom
		ipt.Feeder = feeder // Flush metric data to testing_metrics

		// URL from ENV.
		_, err := toml.Decode(base.conf, ipt)
		assert.NoError(t, err)

		url, err := url.Parse(ipt.URLs[0])
		assert.NoError(t, err)

		ipport, err := netip.ParseAddrPort(url.Host)
		assert.NoError(t, err, "parse %s failed: %s", ipt.URLs[0], err)

		cases = append(cases, &caseSpec{
			t:      t,
			ipt:    ipt,
			name:   base.name,
			feeder: feeder,
			// envs:   envs,

			repo:    base.repo,    // docker name
			repoTag: base.repoTag, // docker tag

			servicePort: fmt.Sprintf("%d", ipport.Port()),

			optsProfileEvents: base.optsProfileEvents,
			optsMetrics:       base.optsMetrics,
			optsAsyncMetrics:  base.optsAsyncMetrics,
			optsStatusInfo:    base.optsStatusInfo,
			// Test case result.
			cr: &testutils.CaseResult{
				Name:        t.Name(),
				Case:        base.name,
				ExtraFields: map[string]any{},
				ExtraTags: map[string]string{
					"image":         base.repo,
					"image_tag":     base.repoTag,
					"remote_server": ipt.URLs[0],
				},
			},
		})
	}
	return cases, nil
}

func TestClickHouseInput(t *T.T) {
	if !testutils.CheckIntegrationTestingRunning() {
		t.Skip()
	}

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
		t.Run(tc.name, func(t *T.T) {
			caseStart := time.Now()

			t.Logf("testing %s...", tc.name)

			// Run a test case.
			if err := tc.run(); err != nil {
				tc.cr.Status = testutils.TestFailed
				tc.cr.FailedMessage = err.Error()

				assert.NoError(t, err)
			} else {
				tc.cr.Status = testutils.TestPassed
			}

			tc.cr.Cost = time.Since(caseStart)

			assert.NoError(t, testutils.Flush(tc.cr))

			t.Cleanup(func() {
				// clean remote docker resources
				if tc.resource == nil {
					return
				}

				assert.NoError(t, tc.pool.Purge(tc.resource))
			})
		})
	}
}
