package mysqlmonitor

const (
	mysql_innodb_blocking_trx_id_sql = `
SELECT 
    a1.ID,
    a1.USER,
    a1.HOST,
    a1.DB,
    a1.COMMAND,
    a1.TIME AS conn_time,
    a1.STATE,
    IFNULL(a1.INFO, '') INFO,
    a3.trx_id,
    a3.trx_state,
    unix_timestamp(a3.trx_started) trx_started,
    IFNULL(a3.trx_requested_lock_id, '') trx_requested_lock_id,
    IFNULL(a3.trx_wait_started, '') trx_wait_started,
    a3.trx_weight,
    a3.trx_mysql_thread_id,
    IFNULL(a3.trx_query, '') trx_query,
    IFNULL(a3.trx_operation_state, '') trx_operation_state,
    a3.trx_tables_in_use,
    a3.trx_tables_locked,
    a3.trx_lock_structs,
    a3.trx_lock_memory_bytes,
    a3.trx_rows_locked,
    a3.trx_rows_modified,
    a3.trx_concurrency_tickets,
    a3.trx_isolation_level,
    a3.trx_unique_checks,
    IFNULL(a3.trx_foreign_key_checks, '') trx_foreign_key_checks,
    IFNULL(a3.trx_last_foreign_key_error, '') trx_last_foreign_key_error,
    a3.trx_adaptive_hash_latched,
    a3.trx_adaptive_hash_timeout,
    a3.trx_is_read_only,
    a3.trx_autocommit_non_locking,
    a2.countnum
FROM
    (SELECT 
        min_blocking_trx_id AS blocking_trx_id,
            COUNT(trx_mysql_thread_id) countnum
    FROM
        (SELECT 
        trx_mysql_thread_id,
            MIN(blocking_trx_id) AS min_blocking_trx_id
    FROM
        (SELECT 
        a.trx_id,
            a.trx_state,
            b.requesting_trx_id,
            b.blocking_trx_id,
            a.trx_mysql_thread_id
    FROM
        information_schema.innodb_lock_waits AS b
    LEFT JOIN information_schema.innodb_trx AS a ON a.trx_id = b.requesting_trx_id) AS t1
    GROUP BY trx_mysql_thread_id
    ORDER BY min_blocking_trx_id) c
    GROUP BY min_blocking_trx_id) a2
        JOIN
    information_schema.innodb_trx a3 ON a2.blocking_trx_id = a3.trx_id
        JOIN
    information_schema.processlist a1 ON a1.id = a3.trx_mysql_thread_id
`

	mysql_innodb_lock_waits_sql = `
select * from information_schema.innodb_locks
`

	mysql_metadatalock_info_sql = `
select a.count, b.id
from
(select count(*) count from information_schema.processlist where State='Waiting for table metadata lock') a
join
(
select max(id) id from
  (select i.trx_mysql_thread_id id from information_schema.innodb_trx i,
  (select
         id, time
     from
         information_schema.processlist
     where
         time = (select
                 max(time)
             from
                 information_schema.processlist
             where
                 state = 'Waiting for table metadata lock'
                     and substring(info, 1, 5) in ('alter' , 'optim', 'repai', 'lock ', 'drop ', 'creat'))) p
  where timestampdiff(second, i.trx_started, now()) > p.time
  and i.trx_mysql_thread_id  not in (connection_id(),p.id)
  union select 0 id) t1
  )b
`

	mysql_metadatalock_session_sql = `
select * from information_schema.processlist where State='Waiting for table metadata lock'
`

	mysql_metadatalock_trx_id_sql = `
select i.trx_mysql_thread_id from information_schema.innodb_trx i,
         (select id, time
               from information_schema.processlist
               where time = (select max(time)
                               from information_schema.processlist
                               where state = 'Waiting for table metadata lock'
                                      and substring(info, 1, 5)
                                      in ('alter' , 'optim', 'repai', 'lock ', 'drop ', 'creat'))) p
         where timestampdiff(second, i.trx_started, now()) > p.time
         and i.trx_mysql_thread_id  not in (connection_id(),p.id)
`

	globalStatusQuery          = `SHOW GLOBAL STATUS`
	globalVariablesQuery       = `SHOW GLOBAL VARIABLES`
	slaveStatusQuery           = `SHOW SLAVE STATUS`
	binaryLogsQuery            = `SHOW BINARY LOGS`
	infoSchemaProcessListQuery = `
        SELECT COALESCE(command,''),COALESCE(state,''),count(*)
        FROM information_schema.processlist
        WHERE ID != connection_id()
        GROUP BY command,state
        ORDER BY null`
	infoSchemaUserStatisticsQuery = `
        SELECT *
        FROM information_schema.user_statistics`
	infoSchemaAutoIncQuery = `
        SELECT table_schema, table_name, column_name, auto_increment,
          CAST(pow(2, case data_type
            when 'tinyint'   then 7
            when 'smallint'  then 15
            when 'mediumint' then 23
            when 'int'       then 31
            when 'bigint'    then 63
            end+(column_type like '% unsigned'))-1 as decimal(19)) as max_int
          FROM information_schema.tables t
          JOIN information_schema.columns c USING (table_schema,table_name)
          WHERE c.extra = 'auto_increment' AND t.auto_increment IS NOT NULL
    `
	innoDBMetricsQuery = `
        SELECT NAME, COUNT
        FROM information_schema.INNODB_METRICS
        WHERE status='enabled'
    `
	perfTableIOWaitsQuery = `
        SELECT OBJECT_SCHEMA, OBJECT_NAME, COUNT_FETCH, COUNT_INSERT, COUNT_UPDATE, COUNT_DELETE,
        SUM_TIMER_FETCH, SUM_TIMER_INSERT, SUM_TIMER_UPDATE, SUM_TIMER_DELETE
        FROM performance_schema.table_io_waits_summary_by_table
        WHERE OBJECT_SCHEMA NOT IN ('mysql', 'performance_schema')
    `
	perfIndexIOWaitsQuery = `
        SELECT OBJECT_SCHEMA, OBJECT_NAME, ifnull(INDEX_NAME, 'NONE') as INDEX_NAME,
        COUNT_FETCH, COUNT_INSERT, COUNT_UPDATE, COUNT_DELETE,
        SUM_TIMER_FETCH, SUM_TIMER_INSERT, SUM_TIMER_UPDATE, SUM_TIMER_DELETE
        FROM performance_schema.table_io_waits_summary_by_index_usage
        WHERE OBJECT_SCHEMA NOT IN ('mysql', 'performance_schema')
    `
	perfTableLockWaitsQuery = `
        SELECT
            OBJECT_SCHEMA,
            OBJECT_NAME,
            COUNT_READ_NORMAL,
            COUNT_READ_WITH_SHARED_LOCKS,
            COUNT_READ_HIGH_PRIORITY,
            COUNT_READ_NO_INSERT,
            COUNT_READ_EXTERNAL,
            COUNT_WRITE_ALLOW_WRITE,
            COUNT_WRITE_CONCURRENT_INSERT,
            COUNT_WRITE_LOW_PRIORITY,
            COUNT_WRITE_NORMAL,
            COUNT_WRITE_EXTERNAL,
            SUM_TIMER_READ_NORMAL,
            SUM_TIMER_READ_WITH_SHARED_LOCKS,
            SUM_TIMER_READ_HIGH_PRIORITY,
            SUM_TIMER_READ_NO_INSERT,
            SUM_TIMER_READ_EXTERNAL,
            SUM_TIMER_WRITE_ALLOW_WRITE,
            SUM_TIMER_WRITE_CONCURRENT_INSERT,
            SUM_TIMER_WRITE_LOW_PRIORITY,
            SUM_TIMER_WRITE_NORMAL,
            SUM_TIMER_WRITE_EXTERNAL
        FROM performance_schema.table_lock_waits_summary_by_table
        WHERE OBJECT_SCHEMA NOT IN ('mysql', 'performance_schema', 'information_schema')
    `
	perfEventsStatementsQuery = `
        SELECT
            ifnull(SCHEMA_NAME, 'NONE') as SCHEMA_NAME,
            DIGEST,
            LEFT(DIGEST_TEXT, %d) as DIGEST_TEXT,
            COUNT_STAR,
            SUM_TIMER_WAIT,
            SUM_ERRORS,
            SUM_WARNINGS,
            SUM_ROWS_AFFECTED,
            SUM_ROWS_SENT,
            SUM_ROWS_EXAMINED,
            SUM_CREATED_TMP_DISK_TABLES,
            SUM_CREATED_TMP_TABLES,
            SUM_SORT_MERGE_PASSES,
            SUM_SORT_ROWS,
            SUM_NO_INDEX_USED
        FROM performance_schema.events_statements_summary_by_digest
        WHERE SCHEMA_NAME NOT IN ('mysql', 'performance_schema', 'information_schema')
            AND last_seen > DATE_SUB(NOW(), INTERVAL %d SECOND)
        ORDER BY SUM_TIMER_WAIT DESC
        LIMIT %d
    `
	perfEventWaitsQuery = `
        SELECT EVENT_NAME, COUNT_STAR, SUM_TIMER_WAIT
        FROM performance_schema.events_waits_summary_global_by_event_name
    `
	perfFileEventsQuery = `
        SELECT
            EVENT_NAME,
            COUNT_READ, SUM_TIMER_READ, SUM_NUMBER_OF_BYTES_READ,
            COUNT_WRITE, SUM_TIMER_WRITE, SUM_NUMBER_OF_BYTES_WRITE,
            COUNT_MISC, SUM_TIMER_MISC
        FROM performance_schema.file_summary_by_event_name
    `
	tableSchemaQuery = `
        SELECT
            TABLE_SCHEMA,
            TABLE_NAME,
            TABLE_TYPE,
            ifnull(ENGINE, 'NONE') as ENGINE,
            ifnull(VERSION, '0') as VERSION,
            ifnull(ROW_FORMAT, 'NONE') as ROW_FORMAT,
            ifnull(TABLE_ROWS, '0') as TABLE_ROWS,
            ifnull(DATA_LENGTH, '0') as DATA_LENGTH,
            ifnull(INDEX_LENGTH, '0') as INDEX_LENGTH,
            ifnull(DATA_FREE, '0') as DATA_FREE,
            ifnull(CREATE_OPTIONS, 'NONE') as CREATE_OPTIONS
        FROM information_schema.tables
        WHERE TABLE_SCHEMA = '%s'
    `
	dbListQuery = `
        SELECT
            SCHEMA_NAME
            FROM information_schema.schemata
        WHERE SCHEMA_NAME NOT IN ('mysql', 'performance_schema', 'information_schema')
    `
	perfSchemaTablesQuery = `
        SELECT
            table_name
            FROM information_schema.tables
        WHERE table_schema = 'performance_schema' AND table_name = ?
    `
)

var metricMap = map[string]string{
	"mysql_innodb_blocking_trx_id": mysql_innodb_blocking_trx_id_sql,
	"mysql_innodb_lock_waits":      mysql_innodb_lock_waits_sql,
	"mysql_metadatalock_info":      mysql_metadatalock_info_sql,
	"mysql_metadatalock_session":   mysql_metadatalock_session_sql,
	"mysql_metadatalock_trx_id":    mysql_metadatalock_trx_id_sql,
}
