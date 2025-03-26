// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package sqlserver

import (
	_ "github.com/microsoft/go-mssqldb"
)

// The SQL scripts assemble the correct query based the version of SQL Server
// see https://sqlserverbuilds.blogspot.com/ for all the details about the version number of SQL Server
// Variable @MajorMinorVersion:
//   - 1000 --> SQL Server 2008
//   - 1050 --> SQL Server 2008 R2
//   - 1100 --> SQL Server 2012
//   - 1200 --> SQL Server 2014
//   - 1300 --> SQL Server 2016
//   - 1400 --> SQL Server 2017
//   - 1500 --> SQL Server 2019

//nolint:lll
// Thanks Bob Ward (http://aka.ms/bobwardms)
// and the folks at Stack Overflow (https://github.com/opserver/Opserver/blob/9c89c7e9936b58ad237b30e6f4cc6cd59c406889/Opserver.Core/Data/SQL/SQLInstance.Memory.cs)
// for putting most of the memory clerk definitions online!

//nolint:lll
const (
	sqlServerDatabaseIO = `
SET DEADLOCK_PRIORITY -10;
IF SERVERPROPERTY('EngineEdition') NOT IN (2,3,4) BEGIN /*NOT IN Standard,Enterpris,Express*/
	DECLARE @ErrorMessage AS nvarchar(500) = 'DataKit - Connection string Server:'+ @@ServerName + ',Database:' + DB_NAME() +' is not a SQL Server Standard,Enterprise or Express. Check the database_type parameter in the datakit configuration.';
	RAISERROR (@ErrorMessage,11,1)
	RETURN
END

DECLARE
	 @SqlStatement AS nvarchar(max)
	,@MajorMinorVersion AS int = CAST(PARSENAME(CAST(SERVERPROPERTY('ProductVersion') AS nvarchar),4) AS int) * 100 + CAST(PARSENAME(CAST(SERVERPROPERTY('ProductVersion') AS nvarchar),3) AS int)
	,@Columns AS nvarchar(max) = ''
	,@Tables AS nvarchar(max) = ''

IF @MajorMinorVersion > 1050 BEGIN
	/*in [volume_mount_point] any trailing "\" char will be automatically removed by datakit */
	SET @Columns += N'
	,[volume_mount_point]'
	SET @Tables += N'
	CROSS APPLY sys.dm_os_volume_stats(vfs.[database_id], vfs.[file_id]) AS vs'
END
IF @MajorMinorVersion > 1100 BEGIN
	SET @Columns += N'
	,vfs.[io_stall_queued_read_ms] AS [rg_read_stall_ms]
	,vfs.[io_stall_queued_write_ms] AS [rg_write_stall_ms]'
END

SET @SqlStatement = N'
SELECT
	''sqlserver_database_io'' AS [measurement]
	,REPLACE(@@SERVERNAME,''\'','':'') AS [sqlserver_host]
	,DB_NAME(vfs.[database_id]) AS [database_name]
	,COALESCE(mf.[physical_name],''RBPEX'') AS [physical_filename]	--RPBEX = Resilient Buffer Pool Extension
	,COALESCE(mf.[name],''RBPEX'') AS [logical_filename]	--RPBEX = Resilient Buffer Pool Extension
	,mf.[type_desc] AS [file_type]
	,vfs.[io_stall_read_ms] AS [read_latency_ms]
	,vfs.[num_of_reads] AS [reads]
	,vfs.[num_of_bytes_read] AS [read_bytes]
	,vfs.[io_stall_write_ms] AS [write_latency_ms]
	,vfs.[num_of_writes] AS [writes]
	,vfs.[num_of_bytes_written] AS [write_bytes]'
	+ @Columns + N'
FROM sys.dm_io_virtual_file_stats(NULL, NULL) AS vfs
INNER JOIN sys.master_files AS mf WITH (NOLOCK)
	ON vfs.[database_id] = mf.[database_id] AND vfs.[file_id] = mf.[file_id]'
+ @Tables;

EXEC sp_executesql @SqlStatement
`

	sqlServerProperties = `
IF SERVERPROPERTY('EngineEdition') NOT IN (2,3,4) BEGIN /*NOT IN Standard,Enterpris,Express*/
	DECLARE @ErrorMessage AS nvarchar(500) = 'DataKit - Connection string Server:'+ @@ServerName + ',Database:' + DB_NAME() +' is not a SQL Server Standard,Enterprise or Express. Check the database_type parameter in the datakit configuration.';
	RAISERROR (@ErrorMessage,11,1)
	RETURN
END

DECLARE
	 @SqlStatement AS nvarchar(max) = ''
	,@MajorMinorVersion AS int = CAST(PARSENAME(CAST(SERVERPROPERTY('ProductVersion') AS nvarchar),4) AS int)*100 + CAST(PARSENAME(CAST(SERVERPROPERTY('ProductVersion') AS nvarchar),3) AS int)
	,@Columns AS nvarchar(MAX) = ''


IF @MajorMinorVersion > 1050 BEGIN
	SET @Columns += N'
	,(si.physical_memory_kb*1024) AS [physical_memory]
	,si.virtual_memory_kb AS [virtual_memory]
	,(si.committed_kb*1024) AS [committed_memory]
	,(si.committed_target_kb*1024) AS [target_memory]'
END

SET @SqlStatement = '
SELECT
	 ''sqlserver'' AS [measurement]
	,REPLACE(@@SERVERNAME, ''\'', '':'') AS [sqlserver_host]
	,(si.ms_ticks) AS [uptime]
	,si.[cpu_count]
	,(SELECT [total_physical_memory_kb] FROM sys.[dm_os_sys_memory]) AS [server_memory]
	,dbs.[db_online]
	,dbs.[db_restoring]
	,dbs.[db_recovering]
	,dbs.[db_recovery_pending]
	,dbs.[db_suspect]
	,dbs.[db_offline]'
	+ @Columns + N'
	FROM sys.[dm_os_sys_info] AS si
	CROSS APPLY (
		SELECT
			 SUM(CASE WHEN state = 0 THEN 1 ELSE 0 END) AS [db_online]
			,SUM(CASE WHEN state = 1 THEN 1 ELSE 0 END) AS [db_restoring]
			,SUM(CASE WHEN state = 2 THEN 1 ELSE 0 END) AS [db_recovering]
			,SUM(CASE WHEN state = 3 THEN 1 ELSE 0 END) AS [db_recovery_pending]
			,SUM(CASE WHEN state = 4 THEN 1 ELSE 0 END) AS [db_suspect]
			,SUM(CASE WHEN state IN(6, 10) THEN 1 ELSE 0 END) AS [db_offline]
		FROM sys.databases
	) AS dbs
'

EXEC sp_executesql @SqlStatement
`

	sqlServerSchedulers = `
IF SERVERPROPERTY('EngineEdition') NOT IN (2,3,4) BEGIN /*NOT IN Standard,Enterpris,Express*/
	DECLARE @ErrorMessage AS nvarchar(500) = 'DataKit - Connection string Server:'+ @@ServerName + ',Database:' + DB_NAME() +' is not a SQL Server Standard,Enterprise or Express. Check the database_type parameter in the datakit configuration.';
	RAISERROR (@ErrorMessage,11,1)
	RETURN
END

DECLARE
	 @SqlStatement AS nvarchar(max)
	,@MajorMinorVersion AS int = CAST(PARSENAME(CAST(SERVERPROPERTY('ProductVersion') AS nvarchar),4) AS int)*100 + CAST(PARSENAME(CAST(SERVERPROPERTY('ProductVersion') AS nvarchar),3) AS int)
	,@Columns AS nvarchar(MAX) = ''

IF @MajorMinorVersion >= 1300 BEGIN
	SET @Columns += N'
	,s.[total_cpu_usage_ms]
	,s.[total_scheduler_delay_ms]'
END

SET @SqlStatement = N'
SELECT
	 ''sqlserver_schedulers'' AS [measurement]
	,REPLACE(@@SERVERNAME, ''\'', '':'') AS [sqlserver_host]
	,CAST(s.[scheduler_id] AS VARCHAR(4)) AS [scheduler_id]
	,CAST(s.[cpu_id] AS VARCHAR(4)) AS [cpu_id]
	,s.[is_online]
	,s.[is_idle]
	,s.[preemptive_switches_count]
	,s.[context_switches_count]
	,s.[current_tasks_count]
	,s.[runnable_tasks_count]
	,s.[current_workers_count]
	,s.[active_workers_count]
	,s.[work_queue_count]
	,s.[pending_disk_io_count]
	,s.[load_factor]
	,s.[yield_count]'
	+ @Columns + N'
FROM sys.dm_os_schedulers AS s'

EXEC sp_executesql @SqlStatement
`

	/*
	   This string defines a SQL statements to retrieve Performance Counters as documented here -
	   	SQL Server Performance Objects - https://docs.microsoft.com/en-us/sql/relational-databases/performance-monitor/use-sql-server-objects?view=sql-server-ver15#SQLServerPOs
	   Some of the specific objects used are -
	   	MSSQL$*:Access Methods - https://docs.microsoft.com/en-us/sql/relational-databases/performance-monitor/sql-server-access-methods-object?view=sql-server-ver15
	   	MSSQL$*:Buffer Manager - https://docs.microsoft.com/en-us/sql/relational-databases/performance-monitor/sql-server-buffer-manager-object?view=sql-server-ver15
	   	MSSQL$*:Databases - https://docs.microsoft.com/en-us/sql/relational-databases/performance-monitor/sql-server-databases-object?view=sql-server-ver15
	   	MSSQL$*:General Statistics - https://docs.microsoft.com/en-us/sql/relational-databases/performance-monitor/sql-server-general-statistics-object?view=sql-server-ver15
	   	MSSQL$*:Exec Statistics - https://docs.microsoft.com/en-us/sql/relational-databases/performance-monitor/sql-server-execstatistics-object?view=sql-server-ver15
	   	SQLServer:Query Store - https://docs.microsoft.com/en-us/sql/relational-databases/performance-monitor/sql-server-query-store-object?view=sql-server-ver15
	*/
	sqlServerPerformanceCounters = `
SET DEADLOCK_PRIORITY -10;
IF SERVERPROPERTY('EngineEdition') NOT IN (2,3,4) BEGIN /*NOT IN Standard,Enterpris,Express*/
	DECLARE @ErrorMessage AS nvarchar(500) = 'DataKit - Connection string Server:'+ @@ServerName + ',Database:' + DB_NAME() +' is not a SQL Server Standard,Enterprise or Express. Check the database_type parameter in the datakit configuration.';
	RAISERROR (@ErrorMessage,11,1)
	RETURN
END

DECLARE
	 @SqlStatement AS nvarchar(max)
	,@MajorMinorVersion AS int = CAST(PARSENAME(CAST(SERVERPROPERTY('ProductVersion') AS nvarchar),4) AS int)*100 + CAST(PARSENAME(CAST(SERVERPROPERTY('ProductVersion') AS nvarchar),3) AS int)

DECLARE @PCounters TABLE
(
	 [object_name] nvarchar(128)
	,[counter_name] nvarchar(128)
	,[instance_name] nvarchar(128)
	,[cntr_value] bigint
	,[cntr_type] int
	PRIMARY KEY([object_name], [counter_name], [instance_name])
);

WITH PerfCounters AS (
SELECT DISTINCT
	 RTRIM(spi.[object_name]) [object_name]
	,RTRIM(spi.[counter_name]) [counter_name]
	,RTRIM(spi.[instance_name]) AS [instance_name]
	,CAST(spi.[cntr_value] AS bigint) AS [cntr_value]
	,spi.[cntr_type]
	FROM sys.dm_os_performance_counters AS spi
	WHERE
		counter_name IN (
			 'SQL Compilations/sec'
			,'SQL Re-Compilations/sec'
			,'User Connections'
			,'Batch Requests/sec'
			,'Logouts/sec'
			,'Logins/sec'
			,'Processes blocked'
			,'Latch Waits/sec'
			,'Full Scans/sec'
			,'Index Searches/sec'
			,'Page Splits/sec'
			,'Page lookups/sec'
			,'Page reads/sec'
			,'Page writes/sec'
			,'Readahead pages/sec'
			,'Lazy writes/sec'
			,'Checkpoint pages/sec'
			,'Table Lock Escalations/sec'
			,'Page life expectancy'
			,'Log File(s) Size (KB)'
			,'Log File(s) Used Size (KB)'
			,'Data File(s) Size (KB)'
			,'Transactions/sec'
			,'Write Transactions/sec'
			,'Active Transactions'
			,'Log Growths'
			,'Active Temp Tables'
			,'Logical Connections'
			,'Temp Tables Creation Rate'
			,'Temp Tables For Destruction'
			,'Free Space in tempdb (KB)'
			,'Version Store Size (KB)'
			,'Version Cleanup rate (KB/s)'
			,'Version Generation rate (KB/s)'
			,'Memory Grants Pending'
			,'Memory Grants Outstanding'
			,'Free list stalls/sec'
			,'Buffer cache hit ratio'
			,'Buffer cache hit ratio base'
			,'Backup/Restore Throughput/sec'
			,'Total Server Memory (KB)'
			,'Target Server Memory (KB)'
			,'Log Flushes/sec'
			,'Log Flush Wait Time'
			,'Memory broker clerk size'
			,'Log Bytes Flushed/sec'
			,'Bytes Sent to Replica/sec'
			,'Log Send Queue'
			,'Bytes Sent to Transport/sec'
			,'Sends to Replica/sec'
			,'Bytes Sent to Transport/sec'
			,'Sends to Transport/sec'
			,'Bytes Received from Replica/sec'
			,'Receives from Replica/sec'
			,'Flow Control Time (ms/sec)'
			,'Flow Control/sec'
			,'Resent Messages/sec'
			,'Redone Bytes/sec'
			,'XTP Memory Used (KB)'
			,'Transaction Delay'
			,'Log Bytes Received/sec'
			,'Log Apply Pending Queue'
			,'Redone Bytes/sec'
			,'Recovery Queue'
			,'Log Apply Ready Queue'
			,'CPU usage %'
			,'CPU usage % base'
			,'Queued requests'
			,'Requests completed/sec'
			,'Blocked tasks'
			,'Active memory grant amount (KB)'
			,'Disk Read Bytes/sec'
			,'Disk Read IO Throttled/sec'
			,'Disk Read IO/sec'
			,'Disk Write Bytes/sec'
			,'Disk Write IO Throttled/sec'
			,'Disk Write IO/sec'
			,'Used memory (KB)'
			,'Forwarded Records/sec'
			,'Background Writer pages/sec'
			,'Percent Log Used'
			,'Log Send Queue KB'
			,'Redo Queue KB'
			,'Redo Bytes Remaining'
			,'Mirrored Write Transactions/sec'
			,'Group Commit Time'
			,'Group Commits/Sec'
			,'Distributed Query'
			,'DTC calls'
			,'Query Store CPU usage'
			,'Range Scans/sec'
			,'Probe Scans/sec'
			,'SQL Cache Memory (KB)'
			,'Database Cache Memory (KB)'
			,'Connection Memory (KB)'
			,'Optimizer Memory (KB)'
			,'Granted Workspace Memory (KB)'
			,'Lock Memory (KB)'
			,'Stolen Server Memory (KB)'
			,'Log Pool Memory (KB)'
			,'Auto-Param Attempts/sec'
			,'Failed Auto-Params/sec'
			,'Safe Auto-Params/sec'
			,'Cache Object Counts'
			,'Cache Pages'
			,'Longest Transaction Running Time'
		) OR (
			spi.[object_name] LIKE '%User Settable%'
			OR spi.[object_name] LIKE '%SQL Errors%'
			OR spi.[object_name] LIKE '%Batch Resp Statistics%'
		) OR (
			spi.[instance_name] IN ('_Total')
			AND spi.[counter_name] IN (
				 'Lock Timeouts/sec'
				,'Lock Timeouts (timeout > 0)/sec'
				,'Number of Deadlocks/sec'
				,'Lock Waits/sec'
				,'Latch Waits/sec'
			)
		)
)

INSERT INTO @PCounters SELECT * FROM PerfCounters;

SELECT
	 'sqlserver_performance' AS [measurement]
	,REPLACE(@@SERVERNAME,'\',':') AS [sqlserver_host]
	,pc.[object_name]
	,pc.[counter_name]
	,CASE pc.[instance_name] WHEN '_Total' THEN 'Total' ELSE ISNULL(pc.[instance_name],'') END AS [instance]
	,CAST(CASE WHEN pc.[cntr_type] = 537003264 AND pc1.[cntr_value] > 0 THEN (pc.[cntr_value] * 1.0) / (pc1.[cntr_value] * 1.0) * 100 ELSE pc.[cntr_value] END AS float(10)) AS [cntr_value]
	,CAST(pc.[cntr_type] AS varchar(25)) AS [counter_type]
FROM @PCounters AS pc
LEFT OUTER JOIN @PCounters AS pc1
	ON (
		pc.[counter_name] = REPLACE(pc1.[counter_name],' base','')
		OR pc.[counter_name] = REPLACE(pc1.[counter_name],' base',' (ms)')
	)
	AND pc.[object_name] = pc1.[object_name]
	AND pc.[instance_name] = pc1.[instance_name]
	AND pc1.[counter_name] LIKE '%base'
WHERE
	pc.[counter_name] NOT LIKE '% base'
OPTION(RECOMPILE)
`

	sqlServerWaitStatsCategorized = `
IF SERVERPROPERTY('EngineEdition') NOT IN (2,3,4) BEGIN /*NOT IN Standard,Enterpris,Express*/
	DECLARE @ErrorMessage AS nvarchar(500) = 'DataKit - Connection string Server:'+ @@ServerName + ',Database:' + DB_NAME() +' is not a SQL Server Standard,Enterprise or Express. Check the database_type parameter in the datakit configuration.';
	RAISERROR (@ErrorMessage,11,1)
	RETURN
END

SELECT
	 'sqlserver_waitstats' AS [measurement]
	,REPLACE(@@SERVERNAME,'\',':') AS [sqlserver_host]
	,ws.[wait_type]
	,ws.[wait_time_ms]
	,ws.[wait_time_ms] - ws.[signal_wait_time_ms] AS [resource_wait_ms]
	,ws.[signal_wait_time_ms]
	,ws.[max_wait_time_ms]
	,ws.[waiting_tasks_count]
	,ISNULL(wc.[wait_category],'OTHER') AS [wait_category]
FROM sys.dm_os_wait_stats AS ws WITH (NOLOCK)
LEFT OUTER JOIN ( VALUES
	('ASYNC_IO_COMPLETION','Other Disk IO'),
	('ASYNC_NETWORK_IO','Network IO'),
	('BACKUPIO','Other Disk IO'),
	('BROKER_CONNECTION_RECEIVE_TASK','Service Broker'),
	('BROKER_DISPATCHER','Service Broker'),
	('BROKER_ENDPOINT_STATE_MUTEX','Service Broker'),
	('BROKER_EVENTHANDLER','Service Broker'),
	('BROKER_FORWARDER','Service Broker'),
	('BROKER_INIT','Service Broker'),
	('BROKER_MASTERSTART','Service Broker'),
	('BROKER_RECEIVE_WAITFOR','User Wait'),
	('BROKER_REGISTERALLENDPOINTS','Service Broker'),
	('BROKER_SERVICE','Service Broker'),
	('BROKER_SHUTDOWN','Service Broker'),
	('BROKER_START','Service Broker'),
	('BROKER_TASK_SHUTDOWN','Service Broker'),
	('BROKER_TASK_STOP','Service Broker'),
	('BROKER_TASK_SUBMIT','Service Broker'),
	('BROKER_TO_FLUSH','Service Broker'),
	('BROKER_TRANSMISSION_OBJECT','Service Broker'),
	('BROKER_TRANSMISSION_TABLE','Service Broker'),
	('BROKER_TRANSMISSION_WORK','Service Broker'),
	('BROKER_TRANSMITTER','Service Broker'),
	('CHECKPOINT_QUEUE','Idle'),
	('CHKPT','Tran Log IO'),
	('CLR_AUTO_EVENT','SQL CLR'),
	('CLR_CRST','SQL CLR'),
	('CLR_JOIN','SQL CLR'),
	('CLR_MANUAL_EVENT','SQL CLR'),
	('CLR_MEMORY_SPY','SQL CLR'),
	('CLR_MONITOR','SQL CLR'),
	('CLR_RWLOCK_READER','SQL CLR'),
	('CLR_RWLOCK_WRITER','SQL CLR'),
	('CLR_SEMAPHORE','SQL CLR'),
	('CLR_TASK_START','SQL CLR'),
	('CLRHOST_STATE_ACCESS','SQL CLR'),
	('CMEMPARTITIONED','Memory'),
	('CMEMTHREAD','Memory'),
	('CXPACKET','Parallelism'),
	('CXCONSUMER','Parallelism'),
	('DBMIRROR_DBM_EVENT','Mirroring'),
	('DBMIRROR_DBM_MUTEX','Mirroring'),
	('DBMIRROR_EVENTS_QUEUE','Mirroring'),
	('DBMIRROR_SEND','Mirroring'),
	('DBMIRROR_WORKER_QUEUE','Mirroring'),
	('DBMIRRORING_CMD','Mirroring'),
	('DTC','Transaction'),
	('DTC_ABORT_REQUEST','Transaction'),
	('DTC_RESOLVE','Transaction'),
	('DTC_STATE','Transaction'),
	('DTC_TMDOWN_REQUEST','Transaction'),
	('DTC_WAITFOR_OUTCOME','Transaction'),
	('DTCNEW_ENLIST','Transaction'),
	('DTCNEW_PREPARE','Transaction'),
	('DTCNEW_RECOVERY','Transaction'),
	('DTCNEW_TM','Transaction'),
	('DTCNEW_TRANSACTION_ENLISTMENT','Transaction'),
	('DTCPNTSYNC','Transaction'),
	('EE_PMOLOCK','Memory'),
	('EXCHANGE','Parallelism'),
	('EXTERNAL_SCRIPT_NETWORK_IOF','Network IO'),
	('FCB_REPLICA_READ','Replication'),
	('FCB_REPLICA_WRITE','Replication'),
	('FT_COMPROWSET_RWLOCK','Full Text Search'),
	('FT_IFTS_RWLOCK','Full Text Search'),
	('FT_IFTS_SCHEDULER_IDLE_WAIT','Idle'),
	('FT_IFTSHC_MUTEX','Full Text Search'),
	('FT_IFTSISM_MUTEX','Full Text Search'),
	('FT_MASTER_MERGE','Full Text Search'),
	('FT_MASTER_MERGE_COORDINATOR','Full Text Search'),
	('FT_METADATA_MUTEX','Full Text Search'),
	('FT_PROPERTYLIST_CACHE','Full Text Search'),
	('FT_RESTART_CRAWL','Full Text Search'),
	('FULLTEXT GATHERER','Full Text Search'),
	('HADR_AG_MUTEX','Replication'),
	('HADR_AR_CRITICAL_SECTION_ENTRY','Replication'),
	('HADR_AR_MANAGER_MUTEX','Replication'),
	('HADR_AR_UNLOAD_COMPLETED','Replication'),
	('HADR_ARCONTROLLER_NOTIFICATIONS_SUBSCRIBER_LIST','Replication'),
	('HADR_BACKUP_BULK_LOCK','Replication'),
	('HADR_BACKUP_QUEUE','Replication'),
	('HADR_CLUSAPI_CALL','Replication'),
	('HADR_COMPRESSED_CACHE_SYNC','Replication'),
	('HADR_CONNECTIVITY_INFO','Replication'),
	('HADR_DATABASE_FLOW_CONTROL','Replication'),
	('HADR_DATABASE_VERSIONING_STATE','Replication'),
	('HADR_DATABASE_WAIT_FOR_RECOVERY','Replication'),
	('HADR_DATABASE_WAIT_FOR_RESTART','Replication'),
	('HADR_DATABASE_WAIT_FOR_TRANSITION_TO_VERSIONING','Replication'),
	('HADR_DB_COMMAND','Replication'),
	('HADR_DB_OP_COMPLETION_SYNC','Replication'),
	('HADR_DB_OP_START_SYNC','Replication'),
	('HADR_DBR_SUBSCRIBER','Replication'),
	('HADR_DBR_SUBSCRIBER_FILTER_LIST','Replication'),
	('HADR_DBSEEDING','Replication'),
	('HADR_DBSEEDING_LIST','Replication'),
	('HADR_DBSTATECHANGE_SYNC','Replication'),
	('HADR_FABRIC_CALLBACK','Replication'),
	('HADR_FILESTREAM_BLOCK_FLUSH','Replication'),
	('HADR_FILESTREAM_FILE_CLOSE','Replication'),
	('HADR_FILESTREAM_FILE_REQUEST','Replication'),
	('HADR_FILESTREAM_IOMGR','Replication'),
	('HADR_FILESTREAM_IOMGR_IOCOMPLETION','Replication'),
	('HADR_FILESTREAM_MANAGER','Replication'),
	('HADR_FILESTREAM_PREPROC','Replication'),
	('HADR_GROUP_COMMIT','Replication'),
	('HADR_LOGCAPTURE_SYNC','Replication'),
	('HADR_LOGCAPTURE_WAIT','Replication'),
	('HADR_LOGPROGRESS_SYNC','Replication'),
	('HADR_NOTIFICATION_DEQUEUE','Replication'),
	('HADR_NOTIFICATION_WORKER_EXCLUSIVE_ACCESS','Replication'),
	('HADR_NOTIFICATION_WORKER_STARTUP_SYNC','Replication'),
	('HADR_NOTIFICATION_WORKER_TERMINATION_SYNC','Replication'),
	('HADR_PARTNER_SYNC','Replication'),
	('HADR_READ_ALL_NETWORKS','Replication'),
	('HADR_RECOVERY_WAIT_FOR_CONNECTION','Replication'),
	('HADR_RECOVERY_WAIT_FOR_UNDO','Replication'),
	('HADR_REPLICAINFO_SYNC','Replication'),
	('HADR_SEEDING_CANCELLATION','Replication'),
	('HADR_SEEDING_FILE_LIST','Replication'),
	('HADR_SEEDING_LIMIT_BACKUPS','Replication'),
	('HADR_SEEDING_SYNC_COMPLETION','Replication'),
	('HADR_SEEDING_TIMEOUT_TASK','Replication'),
	('HADR_SEEDING_WAIT_FOR_COMPLETION','Replication'),
	('HADR_SYNC_COMMIT','Replication'),
	('HADR_SYNCHRONIZING_THROTTLE','Replication'),
	('HADR_TDS_LISTENER_SYNC','Replication'),
	('HADR_TDS_LISTENER_SYNC_PROCESSING','Replication'),
	('HADR_THROTTLE_LOG_RATE_GOVERNOR','Log Rate Governor'),
	('HADR_TIMER_TASK','Replication'),
	('HADR_TRANSPORT_DBRLIST','Replication'),
	('HADR_TRANSPORT_FLOW_CONTROL','Replication'),
	('HADR_TRANSPORT_SESSION','Replication'),
	('HADR_WORK_POOL','Replication'),
	('HADR_WORK_QUEUE','Replication'),
	('HADR_XRF_STACK_ACCESS','Replication'),
	('INSTANCE_LOG_RATE_GOVERNOR','Log Rate Governor'),
	('IO_COMPLETION','Other Disk IO'),
	('IO_QUEUE_LIMIT','Other Disk IO'),
	('IO_RETRY','Other Disk IO'),
	('LATCH_DT','Latch'),
	('LATCH_EX','Latch'),
	('LATCH_KP','Latch'),
	('LATCH_NL','Latch'),
	('LATCH_SH','Latch'),
	('LATCH_UP','Latch'),
	('LAZYWRITER_SLEEP','Idle'),
	('LCK_M_BU','Lock'),
	('LCK_M_BU_ABORT_BLOCKERS','Lock'),
	('LCK_M_BU_LOW_PRIORITY','Lock'),
	('LCK_M_IS','Lock'),
	('LCK_M_IS_ABORT_BLOCKERS','Lock'),
	('LCK_M_IS_LOW_PRIORITY','Lock'),
	('LCK_M_IU','Lock'),
	('LCK_M_IU_ABORT_BLOCKERS','Lock'),
	('LCK_M_IU_LOW_PRIORITY','Lock'),
	('LCK_M_IX','Lock'),
	('LCK_M_IX_ABORT_BLOCKERS','Lock'),
	('LCK_M_IX_LOW_PRIORITY','Lock'),
	('LCK_M_RIn_NL','Lock'),
	('LCK_M_RIn_NL_ABORT_BLOCKERS','Lock'),
	('LCK_M_RIn_NL_LOW_PRIORITY','Lock'),
	('LCK_M_RIn_S','Lock'),
	('LCK_M_RIn_S_ABORT_BLOCKERS','Lock'),
	('LCK_M_RIn_S_LOW_PRIORITY','Lock'),
	('LCK_M_RIn_U','Lock'),
	('LCK_M_RIn_U_ABORT_BLOCKERS','Lock'),
	('LCK_M_RIn_U_LOW_PRIORITY','Lock'),
	('LCK_M_RIn_X','Lock'),
	('LCK_M_RIn_X_ABORT_BLOCKERS','Lock'),
	('LCK_M_RIn_X_LOW_PRIORITY','Lock'),
	('LCK_M_RS_S','Lock'),
	('LCK_M_RS_S_ABORT_BLOCKERS','Lock'),
	('LCK_M_RS_S_LOW_PRIORITY','Lock'),
	('LCK_M_RS_U','Lock'),
	('LCK_M_RS_U_ABORT_BLOCKERS','Lock'),
	('LCK_M_RS_U_LOW_PRIORITY','Lock'),
	('LCK_M_RX_S','Lock'),
	('LCK_M_RX_S_ABORT_BLOCKERS','Lock'),
	('LCK_M_RX_S_LOW_PRIORITY','Lock'),
	('LCK_M_RX_U','Lock'),
	('LCK_M_RX_U_ABORT_BLOCKERS','Lock'),
	('LCK_M_RX_U_LOW_PRIORITY','Lock'),
	('LCK_M_RX_X','Lock'),
	('LCK_M_RX_X_ABORT_BLOCKERS','Lock'),
	('LCK_M_RX_X_LOW_PRIORITY','Lock'),
	('LCK_M_S','Lock'),
	('LCK_M_S_ABORT_BLOCKERS','Lock'),
	('LCK_M_S_LOW_PRIORITY','Lock'),
	('LCK_M_SCH_M','Lock'),
	('LCK_M_SCH_M_ABORT_BLOCKERS','Lock'),
	('LCK_M_SCH_M_LOW_PRIORITY','Lock'),
	('LCK_M_SCH_S','Lock'),
	('LCK_M_SCH_S_ABORT_BLOCKERS','Lock'),
	('LCK_M_SCH_S_LOW_PRIORITY','Lock'),
	('LCK_M_SIU','Lock'),
	('LCK_M_SIU_ABORT_BLOCKERS','Lock'),
	('LCK_M_SIU_LOW_PRIORITY','Lock'),
	('LCK_M_SIX','Lock'),
	('LCK_M_SIX_ABORT_BLOCKERS','Lock'),
	('LCK_M_SIX_LOW_PRIORITY','Lock'),
	('LCK_M_U','Lock'),
	('LCK_M_U_ABORT_BLOCKERS','Lock'),
	('LCK_M_U_LOW_PRIORITY','Lock'),
	('LCK_M_UIX','Lock'),
	('LCK_M_UIX_ABORT_BLOCKERS','Lock'),
	('LCK_M_UIX_LOW_PRIORITY','Lock'),
	('LCK_M_X','Lock'),
	('LCK_M_X_ABORT_BLOCKERS','Lock'),
	('LCK_M_X_LOW_PRIORITY','Lock'),
	('LOGBUFFER','Tran Log IO'),
	('LOGMGR','Tran Log IO'),
	('LOGMGR_FLUSH','Tran Log IO'),
	('LOGMGR_PMM_LOG','Tran Log IO'),
	('LOGMGR_QUEUE','Idle'),
	('LOGMGR_RESERVE_APPEND','Tran Log IO'),
	('MEMORY_ALLOCATION_EXT','Memory'),
	('MEMORY_GRANT_UPDATE','Memory'),
	('MSQL_XACT_MGR_MUTEX','Transaction'),
	('MSQL_XACT_MUTEX','Transaction'),
	('MSSEARCH','Full Text Search'),
	('NET_WAITFOR_PACKET','Network IO'),
	('ONDEMAND_TASK_QUEUE','Idle'),
	('PAGEIOLATCH_DT','Buffer IO'),
	('PAGEIOLATCH_EX','Buffer IO'),
	('PAGEIOLATCH_KP','Buffer IO'),
	('PAGEIOLATCH_NL','Buffer IO'),
	('PAGEIOLATCH_SH','Buffer IO'),
	('PAGEIOLATCH_UP','Buffer IO'),
	('PAGELATCH_DT','Buffer Latch'),
	('PAGELATCH_EX','Buffer Latch'),
	('PAGELATCH_KP','Buffer Latch'),
	('PAGELATCH_NL','Buffer Latch'),
	('PAGELATCH_SH','Buffer Latch'),
	('PAGELATCH_UP','Buffer Latch'),
	('POOL_LOG_RATE_GOVERNOR','Log Rate Governor'),
	('PREEMPTIVE_ABR','Preemptive'),
	('PREEMPTIVE_CLOSEBACKUPMEDIA','Preemptive'),
	('PREEMPTIVE_CLOSEBACKUPTAPE','Preemptive'),
	('PREEMPTIVE_CLOSEBACKUPVDIDEVICE','Preemptive'),
	('PREEMPTIVE_CLUSAPI_CLUSTERRESOURCECONTROL','Preemptive'),
	('PREEMPTIVE_COM_COCREATEINSTANCE','Preemptive'),
	('PREEMPTIVE_COM_COGETCLASSOBJECT','Preemptive'),
	('PREEMPTIVE_COM_CREATEACCESSOR','Preemptive'),
	('PREEMPTIVE_COM_DELETEROWS','Preemptive'),
	('PREEMPTIVE_COM_GETCOMMANDTEXT','Preemptive'),
	('PREEMPTIVE_COM_GETDATA','Preemptive'),
	('PREEMPTIVE_COM_GETNEXTROWS','Preemptive'),
	('PREEMPTIVE_COM_GETRESULT','Preemptive'),
	('PREEMPTIVE_COM_GETROWSBYBOOKMARK','Preemptive'),
	('PREEMPTIVE_COM_LBFLUSH','Preemptive'),
	('PREEMPTIVE_COM_LBLOCKREGION','Preemptive'),
	('PREEMPTIVE_COM_LBREADAT','Preemptive'),
	('PREEMPTIVE_COM_LBSETSIZE','Preemptive'),
	('PREEMPTIVE_COM_LBSTAT','Preemptive'),
	('PREEMPTIVE_COM_LBUNLOCKREGION','Preemptive'),
	('PREEMPTIVE_COM_LBWRITEAT','Preemptive'),
	('PREEMPTIVE_COM_QUERYINTERFACE','Preemptive'),
	('PREEMPTIVE_COM_RELEASE','Preemptive'),
	('PREEMPTIVE_COM_RELEASEACCESSOR','Preemptive'),
	('PREEMPTIVE_COM_RELEASEROWS','Preemptive'),
	('PREEMPTIVE_COM_RELEASESESSION','Preemptive'),
	('PREEMPTIVE_COM_RESTARTPOSITION','Preemptive'),
	('PREEMPTIVE_COM_SEQSTRMREAD','Preemptive'),
	('PREEMPTIVE_COM_SEQSTRMREADANDWRITE','Preemptive'),
	('PREEMPTIVE_COM_SETDATAFAILURE','Preemptive'),
	('PREEMPTIVE_COM_SETPARAMETERINFO','Preemptive'),
	('PREEMPTIVE_COM_SETPARAMETERPROPERTIES','Preemptive'),
	('PREEMPTIVE_COM_STRMLOCKREGION','Preemptive'),
	('PREEMPTIVE_COM_STRMSEEKANDREAD','Preemptive'),
	('PREEMPTIVE_COM_STRMSEEKANDWRITE','Preemptive'),
	('PREEMPTIVE_COM_STRMSETSIZE','Preemptive'),
	('PREEMPTIVE_COM_STRMSTAT','Preemptive'),
	('PREEMPTIVE_COM_STRMUNLOCKREGION','Preemptive'),
	('PREEMPTIVE_CONSOLEWRITE','Preemptive'),
	('PREEMPTIVE_CREATEPARAM','Preemptive'),
	('PREEMPTIVE_DEBUG','Preemptive'),
	('PREEMPTIVE_DFSADDLINK','Preemptive'),
	('PREEMPTIVE_DFSLINKEXISTCHECK','Preemptive'),
	('PREEMPTIVE_DFSLINKHEALTHCHECK','Preemptive'),
	('PREEMPTIVE_DFSREMOVELINK','Preemptive'),
	('PREEMPTIVE_DFSREMOVEROOT','Preemptive'),
	('PREEMPTIVE_DFSROOTFOLDERCHECK','Preemptive'),
	('PREEMPTIVE_DFSROOTINIT','Preemptive'),
	('PREEMPTIVE_DFSROOTSHARECHECK','Preemptive'),
	('PREEMPTIVE_DTC_ABORT','Preemptive'),
	('PREEMPTIVE_DTC_ABORTREQUESTDONE','Preemptive'),
	('PREEMPTIVE_DTC_BEGINTRANSACTION','Preemptive'),
	('PREEMPTIVE_DTC_COMMITREQUESTDONE','Preemptive'),
	('PREEMPTIVE_DTC_ENLIST','Preemptive'),
	('PREEMPTIVE_DTC_PREPAREREQUESTDONE','Preemptive'),
	('PREEMPTIVE_FILESIZEGET','Preemptive'),
	('PREEMPTIVE_FSAOLEDB_ABORTTRANSACTION','Preemptive'),
	('PREEMPTIVE_FSAOLEDB_COMMITTRANSACTION','Preemptive'),
	('PREEMPTIVE_FSAOLEDB_STARTTRANSACTION','Preemptive'),
	('PREEMPTIVE_FSRECOVER_UNCONDITIONALUNDO','Preemptive'),
	('PREEMPTIVE_GETRMINFO','Preemptive'),
	('PREEMPTIVE_HADR_LEASE_MECHANISM','Preemptive'),
	('PREEMPTIVE_HTTP_EVENT_WAIT','Preemptive'),
	('PREEMPTIVE_HTTP_REQUEST','Preemptive'),
	('PREEMPTIVE_LOCKMONITOR','Preemptive'),
	('PREEMPTIVE_MSS_RELEASE','Preemptive'),
	('PREEMPTIVE_ODBCOPS','Preemptive'),
	('PREEMPTIVE_OLE_UNINIT','Preemptive'),
	('PREEMPTIVE_OLEDB_ABORTORCOMMITTRAN','Preemptive'),
	('PREEMPTIVE_OLEDB_ABORTTRAN','Preemptive'),
	('PREEMPTIVE_OLEDB_GETDATASOURCE','Preemptive'),
	('PREEMPTIVE_OLEDB_GETLITERALINFO','Preemptive'),
	('PREEMPTIVE_OLEDB_GETPROPERTIES','Preemptive'),
	('PREEMPTIVE_OLEDB_GETPROPERTYINFO','Preemptive'),
	('PREEMPTIVE_OLEDB_GETSCHEMALOCK','Preemptive'),
	('PREEMPTIVE_OLEDB_JOINTRANSACTION','Preemptive'),
	('PREEMPTIVE_OLEDB_RELEASE','Preemptive'),
	('PREEMPTIVE_OLEDB_SETPROPERTIES','Preemptive'),
	('PREEMPTIVE_OLEDBOPS','Preemptive'),
	('PREEMPTIVE_OS_ACCEPTSECURITYCONTEXT','Preemptive'),
	('PREEMPTIVE_OS_ACQUIRECREDENTIALSHANDLE','Preemptive'),
	('PREEMPTIVE_OS_AUTHENTICATIONOPS','Preemptive'),
	('PREEMPTIVE_OS_AUTHORIZATIONOPS','Preemptive'),
	('PREEMPTIVE_OS_AUTHZGETINFORMATIONFROMCONTEXT','Preemptive'),
	('PREEMPTIVE_OS_AUTHZINITIALIZECONTEXTFROMSID','Preemptive'),
	('PREEMPTIVE_OS_AUTHZINITIALIZERESOURCEMANAGER','Preemptive'),
	('PREEMPTIVE_OS_BACKUPREAD','Preemptive'),
	('PREEMPTIVE_OS_CLOSEHANDLE','Preemptive'),
	('PREEMPTIVE_OS_CLUSTEROPS','Preemptive'),
	('PREEMPTIVE_OS_COMOPS','Preemptive'),
	('PREEMPTIVE_OS_COMPLETEAUTHTOKEN','Preemptive'),
	('PREEMPTIVE_OS_COPYFILE','Preemptive'),
	('PREEMPTIVE_OS_CREATEDIRECTORY','Preemptive'),
	('PREEMPTIVE_OS_CREATEFILE','Preemptive'),
	('PREEMPTIVE_OS_CRYPTACQUIRECONTEXT','Preemptive'),
	('PREEMPTIVE_OS_CRYPTIMPORTKEY','Preemptive'),
	('PREEMPTIVE_OS_CRYPTOPS','Preemptive'),
	('PREEMPTIVE_OS_DECRYPTMESSAGE','Preemptive'),
	('PREEMPTIVE_OS_DELETEFILE','Preemptive'),
	('PREEMPTIVE_OS_DELETESECURITYCONTEXT','Preemptive'),
	('PREEMPTIVE_OS_DEVICEIOCONTROL','Preemptive'),
	('PREEMPTIVE_OS_DEVICEOPS','Preemptive'),
	('PREEMPTIVE_OS_DIRSVC_NETWORKOPS','Preemptive'),
	('PREEMPTIVE_OS_DISCONNECTNAMEDPIPE','Preemptive'),
	('PREEMPTIVE_OS_DOMAINSERVICESOPS','Preemptive'),
	('PREEMPTIVE_OS_DSGETDCNAME','Preemptive'),
	('PREEMPTIVE_OS_DTCOPS','Preemptive'),
	('PREEMPTIVE_OS_ENCRYPTMESSAGE','Preemptive'),
	('PREEMPTIVE_OS_FILEOPS','Preemptive'),
	('PREEMPTIVE_OS_FINDFILE','Preemptive'),
	('PREEMPTIVE_OS_FLUSHFILEBUFFERS','Preemptive'),
	('PREEMPTIVE_OS_FORMATMESSAGE','Preemptive'),
	('PREEMPTIVE_OS_FREECREDENTIALSHANDLE','Preemptive'),
	('PREEMPTIVE_OS_FREELIBRARY','Preemptive'),
	('PREEMPTIVE_OS_GENERICOPS','Preemptive'),
	('PREEMPTIVE_OS_GETADDRINFO','Preemptive'),
	('PREEMPTIVE_OS_GETCOMPRESSEDFILESIZE','Preemptive'),
	('PREEMPTIVE_OS_GETDISKFREESPACE','Preemptive'),
	('PREEMPTIVE_OS_GETFILEATTRIBUTES','Preemptive'),
	('PREEMPTIVE_OS_GETFILESIZE','Preemptive'),
	('PREEMPTIVE_OS_GETFINALFILEPATHBYHANDLE','Preemptive'),
	('PREEMPTIVE_OS_GETLONGPATHNAME','Preemptive'),
	('PREEMPTIVE_OS_GETPROCADDRESS','Preemptive'),
	('PREEMPTIVE_OS_GETVOLUMENAMEFORVOLUMEMOUNTPOINT','Preemptive'),
	('PREEMPTIVE_OS_GETVOLUMEPATHNAME','Preemptive'),
	('PREEMPTIVE_OS_INITIALIZESECURITYCONTEXT','Preemptive'),
	('PREEMPTIVE_OS_LIBRARYOPS','Preemptive'),
	('PREEMPTIVE_OS_LOADLIBRARY','Preemptive'),
	('PREEMPTIVE_OS_LOGONUSER','Preemptive'),
	('PREEMPTIVE_OS_LOOKUPACCOUNTSID','Preemptive'),
	('PREEMPTIVE_OS_MESSAGEQUEUEOPS','Preemptive'),
	('PREEMPTIVE_OS_MOVEFILE','Preemptive'),
	('PREEMPTIVE_OS_NETGROUPGETUSERS','Preemptive'),
	('PREEMPTIVE_OS_NETLOCALGROUPGETMEMBERS','Preemptive'),
	('PREEMPTIVE_OS_NETUSERGETGROUPS','Preemptive'),
	('PREEMPTIVE_OS_NETUSERGETLOCALGROUPS','Preemptive'),
	('PREEMPTIVE_OS_NETUSERMODALSGET','Preemptive'),
	('PREEMPTIVE_OS_NETVALIDATEPASSWORDPOLICY','Preemptive'),
	('PREEMPTIVE_OS_NETVALIDATEPASSWORDPOLICYFREE','Preemptive'),
	('PREEMPTIVE_OS_OPENDIRECTORY','Preemptive'),
	('PREEMPTIVE_OS_PDH_WMI_INIT','Preemptive'),
	('PREEMPTIVE_OS_PIPEOPS','Preemptive'),
	('PREEMPTIVE_OS_PROCESSOPS','Preemptive'),
	('PREEMPTIVE_OS_QUERYCONTEXTATTRIBUTES','Preemptive'),
	('PREEMPTIVE_OS_QUERYREGISTRY','Preemptive'),
	('PREEMPTIVE_OS_QUERYSECURITYCONTEXTTOKEN','Preemptive'),
	('PREEMPTIVE_OS_REMOVEDIRECTORY','Preemptive'),
	('PREEMPTIVE_OS_REPORTEVENT','Preemptive'),
	('PREEMPTIVE_OS_REVERTTOSELF','Preemptive'),
	('PREEMPTIVE_OS_RSFXDEVICEOPS','Preemptive'),
	('PREEMPTIVE_OS_SECURITYOPS','Preemptive'),
	('PREEMPTIVE_OS_SERVICEOPS','Preemptive'),
	('PREEMPTIVE_OS_SETENDOFFILE','Preemptive'),
	('PREEMPTIVE_OS_SETFILEPOINTER','Preemptive'),
	('PREEMPTIVE_OS_SETFILEVALIDDATA','Preemptive'),
	('PREEMPTIVE_OS_SETNAMEDSECURITYINFO','Preemptive'),
	('PREEMPTIVE_OS_SQLCLROPS','Preemptive'),
	('PREEMPTIVE_OS_SQMLAUNCH','Preemptive'),
	('PREEMPTIVE_OS_VERIFYSIGNATURE','Preemptive'),
	('PREEMPTIVE_OS_VERIFYTRUST','Preemptive'),
	('PREEMPTIVE_OS_VSSOPS','Preemptive'),
	('PREEMPTIVE_OS_WAITFORSINGLEOBJECT','Preemptive'),
	('PREEMPTIVE_OS_WINSOCKOPS','Preemptive'),
	('PREEMPTIVE_OS_WRITEFILE','Preemptive'),
	('PREEMPTIVE_OS_WRITEFILEGATHER','Preemptive'),
	('PREEMPTIVE_OS_WSASETLASTERROR','Preemptive'),
	('PREEMPTIVE_REENLIST','Preemptive'),
	('PREEMPTIVE_RESIZELOG','Preemptive'),
	('PREEMPTIVE_ROLLFORWARDREDO','Preemptive'),
	('PREEMPTIVE_ROLLFORWARDUNDO','Preemptive'),
	('PREEMPTIVE_SB_STOPENDPOINT','Preemptive'),
	('PREEMPTIVE_SERVER_STARTUP','Preemptive'),
	('PREEMPTIVE_SETRMINFO','Preemptive'),
	('PREEMPTIVE_SHAREDMEM_GETDATA','Preemptive'),
	('PREEMPTIVE_SNIOPEN','Preemptive'),
	('PREEMPTIVE_SOSHOST','Preemptive'),
	('PREEMPTIVE_SOSTESTING','Preemptive'),
	('PREEMPTIVE_SP_SERVER_DIAGNOSTICS','Preemptive'),
	('PREEMPTIVE_STARTRM','Preemptive'),
	('PREEMPTIVE_STREAMFCB_CHECKPOINT','Preemptive'),
	('PREEMPTIVE_STREAMFCB_RECOVER','Preemptive'),
	('PREEMPTIVE_STRESSDRIVER','Preemptive'),
	('PREEMPTIVE_TESTING','Preemptive'),
	('PREEMPTIVE_TRANSIMPORT','Preemptive'),
	('PREEMPTIVE_UNMARSHALPROPAGATIONTOKEN','Preemptive'),
	('PREEMPTIVE_VSS_CREATESNAPSHOT','Preemptive'),
	('PREEMPTIVE_VSS_CREATEVOLUMESNAPSHOT','Preemptive'),
	('PREEMPTIVE_XE_CALLBACKEXECUTE','Preemptive'),
	('PREEMPTIVE_XE_CX_FILE_OPEN','Preemptive'),
	('PREEMPTIVE_XE_CX_HTTP_CALL','Preemptive'),
	('PREEMPTIVE_XE_DISPATCHER','Preemptive'),
	('PREEMPTIVE_XE_ENGINEINIT','Preemptive'),
	('PREEMPTIVE_XE_GETTARGETSTATE','Preemptive'),
	('PREEMPTIVE_XE_SESSIONCOMMIT','Preemptive'),
	('PREEMPTIVE_XE_TARGETFINALIZE','Preemptive'),
	('PREEMPTIVE_XE_TARGETINIT','Preemptive'),
	('PREEMPTIVE_XE_TIMERRUN','Preemptive'),
	('PREEMPTIVE_XETESTING','Preemptive'),
	('PWAIT_HADR_ACTION_COMPLETED','Replication'),
	('PWAIT_HADR_CHANGE_NOTIFIER_TERMINATION_SYNC','Replication'),
	('PWAIT_HADR_CLUSTER_INTEGRATION','Replication'),
	('PWAIT_HADR_FAILOVER_COMPLETED','Replication'),
	('PWAIT_HADR_JOIN','Replication'),
	('PWAIT_HADR_OFFLINE_COMPLETED','Replication'),
	('PWAIT_HADR_ONLINE_COMPLETED','Replication'),
	('PWAIT_HADR_POST_ONLINE_COMPLETED','Replication'),
	('PWAIT_HADR_SERVER_READY_CONNECTIONS','Replication'),
	('PWAIT_HADR_WORKITEM_COMPLETED','Replication'),
	('PWAIT_HADRSIM','Replication'),
	('PWAIT_RESOURCE_SEMAPHORE_FT_PARALLEL_QUERY_SYNC','Full Text Search'),
	('QUERY_TRACEOUT','Tracing'),
	('REPL_CACHE_ACCESS','Replication'),
	('REPL_HISTORYCACHE_ACCESS','Replication'),
	('REPL_SCHEMA_ACCESS','Replication'),
	('REPL_TRANFSINFO_ACCESS','Replication'),
	('REPL_TRANHASHTABLE_ACCESS','Replication'),
	('REPL_TRANTEXTINFO_ACCESS','Replication'),
	('REPLICA_WRITES','Replication'),
	('REQUEST_FOR_DEADLOCK_SEARCH','Idle'),
	('RESERVED_MEMORY_ALLOCATION_EXT','Memory'),
	('RESOURCE_SEMAPHORE','Memory'),
	('RESOURCE_SEMAPHORE_QUERY_COMPILE','Compilation'),
	('SLEEP_BPOOL_FLUSH','Idle'),
	('SLEEP_BUFFERPOOL_HELPLW','Idle'),
	('SLEEP_DBSTARTUP','Idle'),
	('SLEEP_DCOMSTARTUP','Idle'),
	('SLEEP_MASTERDBREADY','Idle'),
	('SLEEP_MASTERMDREADY','Idle'),
	('SLEEP_MASTERUPGRADED','Idle'),
	('SLEEP_MEMORYPOOL_ALLOCATEPAGES','Idle'),
	('SLEEP_MSDBSTARTUP','Idle'),
	('SLEEP_RETRY_VIRTUALALLOC','Idle'),
	('SLEEP_SYSTEMTASK','Idle'),
	('SLEEP_TASK','Idle'),
	('SLEEP_TEMPDBSTARTUP','Idle'),
	('SLEEP_WORKSPACE_ALLOCATEPAGE','Idle'),
	('SOS_SCHEDULER_YIELD','CPU'),
	('SQLCLR_APPDOMAIN','SQL CLR'),
	('SQLCLR_ASSEMBLY','SQL CLR'),
	('SQLCLR_DEADLOCK_DETECTION','SQL CLR'),
	('SQLCLR_QUANTUM_PUNISHMENT','SQL CLR'),
	('SQLTRACE_BUFFER_FLUSH','Idle'),
	('SQLTRACE_FILE_BUFFER','Tracing'),
	('SQLTRACE_FILE_READ_IO_COMPLETION','Tracing'),
	('SQLTRACE_FILE_WRITE_IO_COMPLETION','Tracing'),
	('SQLTRACE_INCREMENTAL_FLUSH_SLEEP','Idle'),
	('SQLTRACE_PENDING_BUFFER_WRITERS','Tracing'),
	('SQLTRACE_SHUTDOWN','Tracing'),
	('SQLTRACE_WAIT_ENTRIES','Idle'),
	('THREADPOOL','Worker Thread'),
	('TRACE_EVTNOTIF','Tracing'),
	('TRACEWRITE','Tracing'),
	('TRAN_MARKLATCH_DT','Transaction'),
	('TRAN_MARKLATCH_EX','Transaction'),
	('TRAN_MARKLATCH_KP','Transaction'),
	('TRAN_MARKLATCH_NL','Transaction'),
	('TRAN_MARKLATCH_SH','Transaction'),
	('TRAN_MARKLATCH_UP','Transaction'),
	('TRANSACTION_MUTEX','Transaction'),
	('WAIT_FOR_RESULTS','User Wait'),
	('WAITFOR','User Wait'),
	('WRITE_COMPLETION','Other Disk IO'),
	('WRITELOG','Tran Log IO'),
	('XACT_OWN_TRANSACTION','Transaction'),
	('XACT_RECLAIM_SESSION','Transaction'),
	('XACTLOCKINFO','Transaction'),
	('XACTWORKSPACE_MUTEX','Transaction'),
	('XE_DISPATCHER_WAIT','Idle'),
	('XE_TIMER_EVENT','Idle')
) AS wc([wait_type], [wait_category])
	ON ws.[wait_type] = wc.[wait_type]
WHERE
	ws.[wait_type] NOT IN (
		N'BROKER_EVENTHANDLER', N'BROKER_RECEIVE_WAITFOR', N'BROKER_TASK_STOP',
		N'BROKER_TO_FLUSH', N'BROKER_TRANSMITTER', N'CHECKPOINT_QUEUE',
		N'CHKPT', N'CLR_AUTO_EVENT', N'CLR_MANUAL_EVENT', N'CLR_SEMAPHORE',
		N'DBMIRROR_DBM_EVENT', N'DBMIRROR_EVENTS_QUEUE', N'DBMIRROR_WORKER_QUEUE',
		N'DBMIRRORING_CMD', N'DIRTY_PAGE_POLL', N'DISPATCHER_QUEUE_SEMAPHORE',
		N'EXECSYNC', N'FSAGENT', N'FT_IFTS_SCHEDULER_IDLE_WAIT', N'FT_IFTSHC_MUTEX',
		N'HADR_CLUSAPI_CALL', N'HADR_FILESTREAM_IOMGR_IOCOMPLETION', N'HADR_LOGCAPTURE_WAIT',
		N'HADR_NOTIFICATION_DEQUEUE', N'HADR_TIMER_TASK', N'HADR_WORK_QUEUE',
		N'KSOURCE_WAKEUP', N'LAZYWRITER_SLEEP', N'LOGMGR_QUEUE',
		N'MEMORY_ALLOCATION_EXT', N'ONDEMAND_TASK_QUEUE',
		N'PARALLEL_REDO_WORKER_WAIT_WORK',
		N'PREEMPTIVE_HADR_LEASE_MECHANISM', N'PREEMPTIVE_SP_SERVER_DIAGNOSTICS',
		N'PREEMPTIVE_OS_LIBRARYOPS', N'PREEMPTIVE_OS_COMOPS', N'PREEMPTIVE_OS_CRYPTOPS',
		N'PREEMPTIVE_OS_PIPEOPS','PREEMPTIVE_OS_GENERICOPS', N'PREEMPTIVE_OS_VERIFYTRUST',
		N'PREEMPTIVE_OS_DEVICEOPS',
		N'PREEMPTIVE_XE_CALLBACKEXECUTE', N'PREEMPTIVE_XE_DISPATCHER',
		N'PREEMPTIVE_XE_GETTARGETSTATE', N'PREEMPTIVE_XE_SESSIONCOMMIT',
		N'PREEMPTIVE_XE_TARGETINIT', N'PREEMPTIVE_XE_TARGETFINALIZE',
		N'PWAIT_ALL_COMPONENTS_INITIALIZED', N'PWAIT_DIRECTLOGCONSUMER_GETNEXT',
		N'QDS_PERSIST_TASK_MAIN_LOOP_SLEEP',
		N'QDS_ASYNC_QUEUE',
		N'QDS_CLEANUP_STALE_QUERIES_TASK_MAIN_LOOP_SLEEP', N'REQUEST_FOR_DEADLOCK_SEARCH',
		N'RESOURCE_QUEUE', N'SERVER_IDLE_CHECK', N'SLEEP_BPOOL_FLUSH', N'SLEEP_DBSTARTUP',
		N'SLEEP_DCOMSTARTUP', N'SLEEP_MASTERDBREADY', N'SLEEP_MASTERMDREADY',
		N'SLEEP_MASTERUPGRADED', N'SLEEP_MSDBSTARTUP', N'SLEEP_SYSTEMTASK', N'SLEEP_TASK',
		N'SLEEP_TEMPDBSTARTUP', N'SNI_HTTP_ACCEPT', N'SP_SERVER_DIAGNOSTICS_SLEEP',
		N'SQLTRACE_BUFFER_FLUSH', N'SQLTRACE_INCREMENTAL_FLUSH_SLEEP',
		N'SQLTRACE_WAIT_ENTRIES',
		N'WAIT_FOR_RESULTS', N'WAITFOR', N'WAITFOR_TASKSHUTDOWN', N'WAIT_XTP_HOST_WAIT',
		N'WAIT_XTP_OFFLINE_CKPT_NEW_LOG', N'WAIT_XTP_CKPT_CLOSE',
		N'XE_BUFFERMGR_ALLPROCESSED_EVENT', N'XE_DISPATCHER_JOIN',
		N'XE_DISPATCHER_WAIT', N'XE_LIVE_TARGET_TVF', N'XE_TIMER_EVENT',
		N'SOS_WORK_DISPATCHER','RESERVED_MEMORY_ALLOCATION_EXT'
	)
	AND ws.[waiting_tasks_count] > 0
	AND ws.[wait_time_ms] > 100
`

	sqlServerVolumeSpace = `
IF SERVERPROPERTY('EngineEdition') NOT IN (2,3,4) BEGIN /*NOT IN Standard,Enterpris,Express*/
	DECLARE @ErrorMessage AS nvarchar(500) = 'DataKit - Connection string Server:'+ @@ServerName + ',Database:' + DB_NAME() +' is not a SQL Server Standard,Enterprise or Express. Check the database_type parameter in the datakit configuration.';
	RAISERROR (@ErrorMessage,11,1)
	RETURN
END

DECLARE
	@MajorMinorVersion AS int = CAST(PARSENAME(CAST(SERVERPROPERTY('ProductVersion') AS nvarchar),4) AS int)*100 + CAST(PARSENAME(CAST(SERVERPROPERTY('ProductVersion') AS nvarchar),3) AS int)

IF @MajorMinorVersion > 1050 BEGIN
	SELECT DISTINCT
		'sqlserver_volumespace' AS [measurement]
		,REPLACE(@@SERVERNAME,'\',':') AS [sqlserver_host]
		/*in [volume_mount_point] any trailing "\" char will be removed by datakit */
		,vs.[volume_mount_point]
		,vs.[total_bytes] AS [volume_total_space_bytes]
		,vs.[available_bytes] AS [volume_available_space_bytes]
		,vs.[total_bytes] - vs.[available_bytes] AS [volume_used_space_bytes]
	FROM sys.master_files AS mf
	CROSS APPLY sys.dm_os_volume_stats(mf.[database_id], mf.[file_id]) AS vs
END`

	sqlServerLockRow = `
use master;

WITH CTE_SID ( BSID, SID, sql_handle )        
    AS ( SELECT [Blocking_Session_ID],
                        [Session_ID] ,
                        sql_handle
               FROM     sys.dm_exec_requests
               WHERE    [Blocking_Session_ID] <> 0
               UNION ALL
               SELECT   A.[Blocking_Session_ID] ,
                        A.[Session_ID] ,
                        A.sql_handle
               FROM     sys.dm_exec_requests A
                        JOIN CTE_SID B ON A.[Session_ID] = B.BSID
             )
    SELECT  'sqlserver_lock_row' AS [measurement],
            C.BSID AS blocking_session_id,
            C.SID AS session_id,
            S.login_name ,
            S.host_name ,
            S.status as session_status,
            S.cpu_time ,
            S.memory_usage ,
            S.last_request_start_time ,
            S.last_request_end_time ,
            S.logical_reads ,
            S.row_count ,
            q.text as message
    FROM CTE_SID C
            JOIN sys.dm_exec_sessions S ON C.sid = s.[Session_ID]
            CROSS APPLY sys.dm_exec_sql_text(C.sql_handle) Q
    ORDER BY sid
`

	sqlServerLockTable = `
  SELECT 'sqlserver_lock_table' as [measurement],
		request_session_id,
		resource_type,
		convert(varchar(20), db_name(resource_database_id)) as db_name,
		OBJECT_NAME(resource_associated_entity_id, resource_database_id) object_name,
		request_mode,
		request_status
	from sys.dm_tran_locks
	where resource_type in ('OBJECT')

`

	sqlServerLockDead = `

  SELECT
	  'sqlserver_lock_dead' as [measurement],
    db.name db_name,
    tl.request_session_id,
    wt.blocking_session_id,
    OBJECT_NAME(p.OBJECT_ID) blocking_object_name,
    tl.resource_type,
    h1.TEXT AS requesting_text,
    h2.TEXT AS blocking_text,
		h2.TEXT AS message,
    tl.request_mode
  FROM sys.dm_tran_locks AS tl
    INNER JOIN sys.databases db ON db.database_id = tl.resource_database_id
    INNER JOIN sys.dm_os_waiting_tasks AS wt ON tl.lock_owner_address = wt.resource_address
    INNER JOIN sys.partitions AS p ON p.hobt_id = tl.resource_associated_entity_id
    INNER JOIN sys.dm_exec_connections ec1 ON ec1.session_id = tl.request_session_id
    INNER JOIN sys.dm_exec_connections ec2 ON ec2.session_id = wt.blocking_session_id
    CROSS APPLY sys.dm_exec_sql_text(ec1.most_recent_sql_handle) AS h1
    CROSS APPLY sys.dm_exec_sql_text(ec2.most_recent_sql_handle) AS h2
	WHERE
		tl.request_session_id <> @@SPID 
		AND tl.request_mode <> 'SCH_M'
		AND tl.request_session_id IS NOT NULL;
`

	sqlServerLogicIO = `
SELECT
	'sqlserver_logical_io' AS [measurement],
	creation_time,
	last_execution_time,
	total_logical_reads,
	total_logical_writes,
	(total_logical_reads + total_logical_writes) AS total_logical_io,
	execution_count,
	(total_logical_reads + total_logical_writes)/ Execution_count AS 'avg_logical_io',
		CAST(SUBSTRING(sql_text.text, (statement_start_offset / 2) + 1,
        CASE
            WHEN (statement_end_offset - statement_start_offset) / 2 <= 0 THEN 640000
            ELSE (statement_end_offset - statement_start_offset) / 2
        END) AS VARCHAR(MAX)) AS message
FROM sys.dm_exec_query_stats
CROSS APPLY sys.dm_exec_sql_text(sql_handle) AS sql_text
WHERE
	(total_logical_reads + total_logical_writes)/ Execution_count > 50000
	AND
   creation_time > dateadd(second, -__COLLECT_INTERVAL_SECONDS__, GETDATE())
ORDER BY
	(total_logical_reads + total_logical_writes) DESC
`

	sqlServerWorkerTime = `
	select 
	  'sqlserver_worker_time' as [measurement],
		creation_time,
		last_execution_time,
		total_worker_time/1000 total_worker_time,
		execution_count,
		((total_worker_time / execution_count)/1000) as 'avg_worker_time',
		substring(sql_text.text, (statement_start_offset/2),
		case
		when (statement_end_offset -statement_start_offset)/2 <=0 then 64000
		else (statement_end_offset -statement_start_offset)/2 end) message 
	from sys.dm_exec_query_stats
	cross apply sys.dm_exec_sql_text(sql_handle) as sql_text
	cross apply sys.dm_exec_query_plan(plan_handle) as plan_text
	where ((total_worker_time / execution_count)/1000) >= 1000 
		and creation_time > dateadd(second, -__COLLECT_INTERVAL_SECONDS__, GETDATE())
	order by total_worker_time/1000 Desc
`

	sqlServerDatabaseSize = `
	with fs
	as
	(
		select database_id, type, size * 8.0 / 1024 size
		from sys.master_files
	)

	select top 10
	'sqlserver_database_size' as [measurement],
	name database_name,
	(select cast(round(sum(size),2) as numeric(15,2)) from fs where type = 0 and fs.database_id = db.database_id) data_size,
	(select cast(round(sum(size),2) as numeric(15,2)) from fs where type = 1 and fs.database_id = db.database_id) log_size 
	from sys.databases db
	order by 2 desc
`

	sqlServerDatabaseBackup = `
select 'sqlserver_database_backup' as [measurement], 
	sys.databases.name as database_name, 
	count(backup_set_id) as backup_count
from msdb.dbo.backupset right outer join sys.databases
on sys.databases.name = msdb.dbo.backupset.database_name
group by sys.databases.name
`
)
