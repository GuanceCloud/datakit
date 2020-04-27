package oraclemonitor

const oracle_hostinfo_sql = `
SELECT stat_name, value
FROM v$osstat
WHERE stat_name IN ('PHYSICAL_MEMORY_BYTES', 'NUM_CPUS')
`

const oracle_dbinfo_sql = `
SELECT dbid AS ora_db_id, name AS db_name, db_unique_name
	, to_char(created, 'yyyy-mm-dd hh24:mi:ss') AS db_create_time, log_mode AS log_mod
	, flashback_on AS flashback_mod, database_role, platform_name AS platform, open_mode, protection_mode
	, protection_level, switchover_status
FROM v$database
`

const oracle_instinfo_sql = `
SELECT instance_number, instance_name AS ora_sid, host_name, version
	, to_char(startup_time, 'yyyy-mm-dd hh24:mi:ss') AS startup_time, status
	, CASE 
		WHEN parallel = 'YES' THEN 1
		ELSE 0
	END AS is_rac
FROM v$instance
`

const oracle_psu_sql = `
select  nvl(max(id),0) max_id from  Dba_Registry_History""",
    "oracle_key_params": """SELECT
    name,
    value
FROM
    v$parameter
WHERE
    name IN (
        'audit_trail',
        'sessions',
        'processes'
    )
`

const oracle_blocking_sessions_sql = `
WITH sessions AS (
	SELECT sid, serial# AS serial, logon_time, status, event
		, p1, p2, p3, username, terminal
		, program, sql_id, prev_sql_id, last_call_et, blocking_session
		, blocking_instance, row_wait_obj# AS row_wait_obj
	FROM v$session
)
SELECT *
FROM sessions
WHERE sid IN (
	SELECT blocking_session
	FROM sessions
)
OR blocking_session IS NOT NULL
`

const oracle_undo_stat_sql = `
SELECT
    to_char(begin_time, 'yyyy-mm-dd hh24:mi:ss') begin_time,
    to_char(end_time, 'yyyy-mm-dd hh24:mi:ss') end_time,
    undoblks,
    txncount,
    activeblks,
    unexpiredblks,
    expiredblks
FROM
    v$undostat
WHERE
    ROWNUM < 2
`

const oracle_redo_info_sql = `
SELECT
    group#      group_no,
    thread#     thread_no,
    sequence#   sequence_no,
    bytes,
    members,
    archived,
    status
FROM
    v$log
`

const oracle_standby_log_sql = `
SELECT
    *
FROM
    (
        SELECT
            severity,
            message_num,
            error_code,
            to_char(timestamp, 'yyyy-mm-dd hh24:mi:ss') AS log_time,
            message
        FROM
            v$dataguard_status
        ORDER BY
            message_num DESC
    )
WHERE
    ROWNUM < 10
`

const oracle_standby_process_sql = `
SELECT
    process,
    ROW_NUMBER() OVER(
        PARTITION BY process
        ORDER BY
            process
    ) process_seq,
    status,
    client_process,
    client_dbid,
    group#      group_no,
    thread#     thread_no,
    sequence#   sequence_no,
    blocks,
    delay_mins
FROM
    v$managed_standby
`

var oracle_asm_diskgroups_sql = `
SELECT
    group_number,
    name       AS group_name,
    sector_size,
    block_size,
    allocation_unit_size,
    state,
    type,
    total_mb   AS space_total,
    free_mb    AS space_free,
    total_mb - free_mb AS space_used,
    required_mirror_free_mb,
    usable_file_mb,
    offline_disks,
    compatibility,
    database_compatibility
FROM
    v$asm_diskgroup
`

var oracle_flash_area_info_sql = `
SELECT
    substr(name, 1, 64) AS name,
    space_limit,
    space_used,
    space_reclaimable,
    number_of_files
FROM
    v$recovery_file_dest
`

var oracle_tbs_free_space_sql = `
SELECT
    tablespace_name,
    SUM(bytes) as space_free
FROM
    dba_free_space
GROUP BY
    tablespace_name
`

const oracle_tbs_space_sql = `
SELECT
    tablespace_name,
    SUM(bytes)  AS space_total,
    SUM(
        CASE
            WHEN autoextensible = 'YES' THEN
                maxbytes - bytes
            ELSE
                0
        END
    )  AS space_extensible,
    COUNT(*) AS num_files
FROM
    dba_data_files
WHERE
    status = 'AVAILABLE'
GROUP BY
    tablespace_name
UNION ALL
SELECT
    tablespace_name,
    SUM(bytes) AS space_total,
    SUM(
        CASE
            WHEN autoextensible = 'YES' THEN
                maxbytes - bytes
            ELSE
                0
        END
    ) AS space_extensible,
    COUNT(*) AS num_files
FROM
    dba_temp_files
WHERE
    status = 'ONLINE'
GROUP BY
    tablespace_name
`

const oracle_tbs_meta_info_sql = `
SELECT
    tablespace_name,
    block_size,
    initial_extent,
    next_extent,
    min_extents,
    max_extents,
    pct_increase,
    min_extlen,
    status,
    contents,
    logging,
    force_logging,
    extent_management,
    allocation_type,
    plugged_in,
    segment_space_management,
    def_tab_compression,
    retention,
    bigfile
FROM
    dba_tablespaces
`
const oracle_temp_segment_usage_sql = `
SELECT
    tablespace_name,
    SUM(total_blocks) sum_total_blocks,
    SUM(max_blocks) sum_max_blocks,
    SUM(used_blocks) sum_used_blocks,
    SUM(free_blocks) sum_free_blocks
FROM
    v$sort_segment
GROUP BY
    tablespace_name
`

const oracle_trans_sql = `
select
            count(*) num_trans,
            nvl(round(max(used_ublk * 8192 / 1024 / 1024), 2),0) max_undo_size,
            nvl(round(avg(used_ublk * 8192 / 1024 / 1024), 2),0) avg_undo_size,
            round(nvl((sysdate - min(to_date(start_time, 'mm/dd/yy hh24:mi:ss'))),0) * 1440 * 60,0) longest_trans
        FROM v$transaction
`

const oracle_archived_log_sql = `
select count(*) value from v$archived_log where archived='YES' and deleted='NO'
`

const oracle_pgastat_sql = `
select name,value,unit from v$pgastat
`

var oracle_accounts_sql = `
select 
username
,user_id
,password
,account_status
,to_char(lock_date, 'yyyy-mm-dd hh24:mi:ss') AS lock_date
,to_char(expiry_date, 'yyyy-mm-dd hh24:mi:ss') AS expiry_date
,default_tablespace
,temporary_tablespace
,to_char(created, 'yyyy-mm-dd hh24:mi:ss') AS created
,profile
,initial_rsrc_consumer_group
,external_name
,password_versions
,editions_enabled
,authentication_type
from dba_users
`

var oracle_locks_sql = `
SELECT b.session_id AS session_id,
       NVL(b.oracle_username, '(oracle)') AS oracle_username,
       a.owner AS object_owner,
       a.object_name,
       Decode(b.locked_mode, 0, 'None',
                             1, 'Null (NULL)',
                             2, 'Row-S (SS)',
                             3, 'Row-X (SX)',
                             4, 'Share (S)',
                             5, 'S/Row-X (SSX)',
                             6, 'Exclusive (X)',
                             b.locked_mode) locked_mode,
       b.os_user_name
FROM   dba_objects a,
       v$locked_object b
WHERE  a.object_id = b.object_id
ORDER BY 1, 2, 3, 4
`

const oracle_session_ratio_sql = `
SELECT 'session_cached_cursors' parameter,  
         LPAD(VALUE, 5) value,  
         DECODE(VALUE, 0, ' n/a', TO_CHAR(100 * USED / VALUE, '990') ) usage  
   FROM (SELECT MAX(S.VALUE) USED  
            FROM V$STATNAME N, V$SESSTAT S  
           WHERE N.NAME = 'session cursor cache count'  
             AND S.STATISTIC# = N.STATISTIC#),  
         (SELECT VALUE FROM V$PARAMETER WHERE NAME = 'session_cached_cursors')  
  UNION ALL  
SELECT 'open_cursors' parameter,  
         LPAD(VALUE, 5) value,  
         TO_CHAR(100 * USED / VALUE, '990')   usage
   FROM (SELECT MAX(SUM(S.VALUE)) USED  
            FROM V$STATNAME N, V$SESSTAT S  
           WHERE N.NAME IN  
                 ('opened cursors current', 'session cursor cache count')  
             AND S.STATISTIC# = N.STATISTIC#  
           GROUP BY S.SID),  
         (SELECT VALUE FROM V$PARAMETER WHERE NAME = 'open_cursors')
`

const oracle_sessions_sql = `
SELECT
        sid,
        serial#         serial,
        to_char(logon_time, 'yyyy-mm-dd hh24:mi:ss') logon_time,
        status,
        event,
        p1,
        p2,
        p3,
        username,
        terminal,
        program,
        sql_id,
        prev_sql_id,
        last_call_et,
        blocking_session,
        blocking_instance,
        row_wait_obj#   row_wait_obj
    FROM
        v$session
`

const oracle_pdb_backup_set_info_sql = `
SELECT
    bs_key,
    recid,
    stamp,
    to_char(start_time, 'yyyy-mm-dd hh24:mi:ss') start_time,
    to_char(completion_time, 'yyyy-mm-dd hh24:mi:ss') completion_time,
    elapsed_seconds,
    output_bytes,
    CASE
        WHEN backup_type = 'D' THEN
            'DB FULL'
        WHEN backup_type = 'L' THEN
            'ARCHIVELOG'
        WHEN backup_type = 'I' THEN
            'DB INCR'
        ELSE
            backup_type
    END AS backup_type,
    'SUCCESS' AS status
FROM
    v$backup_set_details
WHERE
    completion_time >= sysdate - 1 / 2
`

const oracle_cdb_backup_job_info_sql = `
SELECT
    session_key,
    session_recid,
    session_stamp,
    to_char(start_time, 'yyyy-mm-dd hh24:mi:ss') start_time,
    to_char(end_time, 'yyyy-mm-dd hh24:mi:ss') end_time,
    elapsed_seconds,
    output_bytes,
    input_type,
    CASE
        WHEN status LIKE 'COMPLETED%' THEN
            'SUCCESS'
        ELSE
            'FAILED'
    END AS status
FROM
    v$rman_backup_job_details
WHERE
    end_time >= sysdate - 1 / 2
    AND status NOT LIKE 'RUNNING%'
`

const oracle_snap_info_sql = `
SELECT dbid,
to_char(sys_extract_utc(s.startup_time), 'yyyy-mm-dd hh24:mi:ss') snap_startup_time,
to_char(sys_extract_utc(s.begin_interval_time),
   'yyyy-mm-dd hh24:mi:ss') begin_interval_time,
to_char(sys_extract_utc(s.end_interval_time), 'yyyy-mm-dd hh24:mi:ss') end_interval_time,
s.snap_id, s.instance_number,
(cast(s.end_interval_time as date) - cast(s.begin_interval_time as date))*86400 as span_in_second
from dba_hist_snapshot  s, v$instance b
where s.end_interval_time >= sysdate - interval '2' hour
and s.INSTANCE_NUMBER = b.INSTANCE_NUMBER
`

var metricMap = map[string]string{
	"oracle_hostinfo":            oracle_hostinfo_sql,
	"oracle_dbinfo":              oracle_dbinfo_sql,
	"oracle_instinfo":            oracle_instinfo_sql,
	"oracle_psu":                 oracle_psu_sql,
	"oracle_blocking_sessions":   oracle_blocking_sessions_sql,
	"oracle_undo_stat":           oracle_undo_stat_sql,
	"oracle_redo_info":           oracle_redo_info_sql,
	"oracle_standby_log":         oracle_standby_log_sql,
	"oracle_standby_process":     oracle_standby_process_sql,
	"oracle_asm_diskgroups":      oracle_asm_diskgroups_sql,
	"oracle_flash_area_info":     oracle_flash_area_info_sql,
	"oracle_tbs_free_space":      oracle_tbs_free_space_sql,
	"oracle_tbs_space":           oracle_tbs_space_sql,
	"oracle_tbs_meta_info":       oracle_tbs_meta_info_sql,
	"oracle_temp_segment_usage":  oracle_temp_segment_usage_sql,
	"oracle_trans":               oracle_trans_sql,
	"oracle_archived_log":        oracle_archived_log_sql,
	"oracle_pgastat":             oracle_pgastat_sql,
	"oracle_accounts":            oracle_accounts_sql,
	"oracle_locks":               oracle_locks_sql,
	"oracle_session_ratio":       oracle_session_ratio_sql,
	"oracle_sessions":            oracle_sessions_sql,
	"oracle_pdb_backup_set_info": oracle_pdb_backup_set_info_sql,
	"oracle_cdb_backup_job_info": oracle_cdb_backup_job_info_sql,
	"oracle_snap_info":           oracle_snap_info_sql,
}

var tagsMap = map[string][]string{
	"oracle_pdb_backup_set_info": []string{"bs_key"},
	"oracle_cdb_backup_job_info": []string{"bs_key"},
	"oracle_hostinfo":            []string{"stat_name"},
	"oracle_dbinfo":              []string{"ora_db_id"},
	"oracle_key_params":          []string{"name"},
	"oracle_blocking_sessions":   []string{"serial"},
	"oracle_redo_info":           []string{"group_no", "sequence_no"},
	"oracle_standby_log":         []string{"message_num"},
	"oracle_standby_process":     []string{"process_seq"},
	"oracle_asm_diskgroups":      []string{"group_number", "group_name"},
	"oracle_flash_area_info":     []string{"name"},
	"oracle_tbs_free_space":      []string{"tablespace_name"},
	"oracle_tbs_space":           []string{"tablespace_name"},
	"oracle_tbs_meta_info":       []string{"tablespace_name"},
	"oracle_temp_segment_usage":  []string{"tablespace_name"},
	"oracle_pgastat":             []string{"name"},
	"oracle_accounts":            []string{"username", "user_id"},
	"oracle_locks":               []string{"session_id"},
	"oracle_session_ratio":       []string{"parameter"},
	"oracle_sessions":            []string{"serial", "username"},
	"oracle_snap_info":           []string{"dbid", "snap_id"},
}
