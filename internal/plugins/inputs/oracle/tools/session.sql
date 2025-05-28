SELECT
    s.sid,
    s.serial#,
    s.username,
    sn.name AS stat_name,
    ss.value AS stat_value
FROM
    v$session s
JOIN
    v$sesstat ss ON s.sid = ss.sid
JOIN
    v$statname sn ON ss.statistic# = sn.statistic#
WHERE
    s.type = 'USER' AND s.username IS NOT NULL -- 只关注用户会话
    AND sn.name IN (
        'session uga memory',
        'session pga memory',
        'CPU used by this session',
        'physical reads',
        'db block gets',
        'consistent gets',
        'redo size',
        'parse count (total)',
        'execute count'
        -- ... 添加你关心的其他统计项
    );;;

SELECT
    s.sid,
    s.serial#,
    s.username,
    s.osuser,
    s.machine,
    s.terminal,
    s.program,
    s.module,                       -- DBMS_APPLICATION_INFO
    s.action,                       -- DBMS_APPLICATION_INFO
    s.client_info,                  -- DBMS_APPLICATION_INFO
    s.type,
    s.status,
    TO_CHAR(s.logon_time, 'YYYY-MM-DD HH24:MI:SS') AS logon_time,
    s.last_call_et,
    s.service_name,
    s.server,

    s.sql_address,                  -- For joining with V$SQLAREA/V$SQL
    s.sql_hash_value,               -- For joining with V$SQLAREA/V$SQL
    s.sql_id,                       -- SQL ID (10g+)
    s.sql_child_number,             -- (10g+)
    s.sql_exec_start,               -- (10gR2 or 11g+)

    -- SQL Text (safer to join with v$sqlarea using sql_address and sql_hash_value for older versions)
    -- For newer versions, sql_id join is common.
    --sqla.sql_text AS current_sql_text, -- First 1000 chars

    s.prev_sql_addr,                -- Previous SQL Address
    s.prev_hash_value,              -- Previous SQL Hash Value
    s.prev_sql_id,                  -- Previous SQL ID (10g+)
    s.prev_child_number,            -- (10g+)
    s.prev_exec_start,              -- (10gR2 or 11g+)

    s.event#,
    s.event,
    s.wait_class_id,                -- (10g+)
    s.wait_class#,                  -- (10g+)
    s.wait_class,                   -- (10g+)
    s.wait_time,                    -- If 0, means currently waiting. If >0, last wait time in centiseconds.
                                    -- For microseconds, s.wait_time_micro is newer (11g+)
    s.seconds_in_wait,              -- (10g+)
    s.state,                        -- Wait state

    s.p1,                           -- Wait event parameter 1 value
    s.p2,                           -- Wait event parameter 2 value
    s.p3,                           -- Wait event parameter 3 value
    -- P1TEXT, P2TEXT, P3TEXT are newer (10g+)

    s.blocking_session_status,      -- (9iR2+)
    s.blocking_instance,            -- (RAC, 9iR2+)
    s.blocking_session,             -- (9iR2+)

    s.row_wait_obj#,
    s.row_wait_file#,
    s.row_wait_block#,
    s.row_wait_row#,

    -- PGA Memory (PGA_MAX_MEM is newer, 10g or 11g+)
		-- s.pga_tunable_mem,
    -- s.pga_used_mem,
    -- s.pga_alloc_mem,
    -- s.pga_freeable_mem,
    -- s.pga_max_mem,               -- Uncomment if your version supports it (e.g., 11g+)

    s.process AS client_os_pid,
    sp.spid AS server_os_pid,
    s.saddr AS session_address,
    s.taddr AS transaction_address

    -- Newer columns, uncomment if your version supports them:
    -- s.ecid,                      -- Execution Context ID (11g+)
    -- s.wait_time_micro,           -- (11g+)
    -- s.p1text,
    -- s.p2text,
    -- s.p3text,
    -- s.sql_exec_id (11g+)
    -- s.plsql_entry_object_id, s.plsql_entry_subprogram_id (11g+)
    -- s.plsql_object_id, s.plsql_subprogram_id (11g+)

FROM
    v$session s
LEFT JOIN
    v$process sp ON s.paddr = sp.addr
LEFT JOIN
    v$sqlarea sqla ON s.sql_address = sqla.address AND s.sql_hash_value = sqla.hash_value
    -- For newer versions or if you prefer sql_id and are sure it's populated:
    -- LEFT JOIN v$sqlarea sqla ON s.sql_id = sqla.sql_id AND ROWNUM = 1 -- ROWNUM to avoid multiple rows if sql_id has multiple plans in v$sqlarea
    -- If using v$sql for more precise child cursor text:
    -- LEFT JOIN v$sql vsql ON s.sql_address = vsql.address AND s.sql_hash_value = vsql.hash_value AND s.sql_child_number = vsql.child_number
WHERE
    s.status = 'ACTIVE'
    AND s.type = 'USER'
    AND s.username IS NOT NULL
ORDER BY
    s.sid;;;
