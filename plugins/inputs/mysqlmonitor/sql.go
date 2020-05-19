package mysqlmonitor

const mysql_innodb_blocking_trx_id_sql = `
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
    information_schema.processlist a1 ON a1.id = a3.trx_mysql_thread_id;
`

const mysql_innodb_lock_waits_sql = `
select * from information_schema.innodb_locks;
`

const mysql_metadatalock_info_sql = `
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
  )b;
`

const mysql_metadatalock_session_sql = `
select * from information_schema.processlist where State='Waiting for table metadata lock';
`
const mysql_metadatalock_trx_id_sql = `
select i.trx_mysql_thread_id from information_schema.innodb_trx i,
         (select id, time
               from information_schema.processlist
               where time = (select max(time)
                               from information_schema.processlist
                               where state = 'Waiting for table metadata lock'
                                      and substring(info, 1, 5)
                                      in ('alter' , 'optim', 'repai', 'lock ', 'drop ', 'creat'))) p
         where timestampdiff(second, i.trx_started, now()) > p.time
         and i.trx_mysql_thread_id  not in (connection_id(),p.id);
`

var metricMap = map[string]string{
	"mysql_innodb_blocking_trx_id": mysql_innodb_blocking_trx_id_sql,
	"mysql_innodb_lock_waits":      mysql_innodb_lock_waits_sql,
	"mysql_metadatalock_info":      mysql_metadatalock_info_sql,
	"mysql_metadatalock_session":   mysql_metadatalock_session_sql,
	"mysql_metadatalock_trx_id":    mysql_metadatalock_trx_id_sql,
}
