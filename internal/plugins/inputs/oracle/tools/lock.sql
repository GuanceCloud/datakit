---------------------------------------------------
-- lock(optimized)
---------------------------------------------------
SELECT
    l.session_id,       -- 持有锁的会话 SID
    s.serial#,          -- 会话的序列号
    s.username,         -- Oracle 用户名 (连接用户)
    s.osuser,           -- 操作系统用户名
    s.machine,          -- 客户端机器名
    s.program,          -- 客户端程序名
    o.owner AS object_owner, -- 对象所有者
    o.object_name,      -- 被锁定的对象名
    o.object_type,      -- 被锁定的对象类型
    l.locked_mode,      -- 锁模式 (数字代码)
    -- 可以将锁模式转换为文本描述，更易读
    DECODE(l.locked_mode,
        0, 'None',             -- Not locked 
        1, 'Null',             -- Null 
        2, 'Row-S (SS)',       -- Row Share 
        3, 'Row-X (SX)',       -- Row Exclusive 
        4, 'Share',            -- Share 
        5, 'S/Row-X (SSX)',    -- Share Row Exclusive 
        6, 'Exclusive (X)',    -- Exclusive 
        TO_CHAR(l.locked_mode) 
    ) AS locked_mode_desc,
    s.sql_id AS current_sql_id, -- 当前会话正在执行的 SQL ID (可能不是导致锁的 SQL)
    s.prev_sql_id AS blocking_sql_id, -- 前一个执行的 SQL ID (更有可能是导致锁的 SQL)
    s.sid AS holding_sid, -- 持有锁的会话 SID (与 l.session_id 相同，用于明确)
    s.blocking_session AS waiting_session_sid, -- 如果此会话正在阻塞其他会话，这里是被阻塞会话的 SID
    s.blocking_session_status AS waiting_session_status -- 被阻塞会话的状态
FROM
    v$locked_object l
JOIN
    dba_objects o ON l.object_id = o.object_id
JOIN
    v$session s ON l.session_id = s.sid
ORDER BY
    l.session_id, o.object_name
;;;

---------------------------------------------------
-- lock
---------------------------------------------------
SELECT l.session_id, 
s.serial#, 
s.username, 
s.osuser, 
s.machine, 
o.object_name, 
o.object_type, 
l.oracle_username, 
l.process, 
l.locked_mode 
FROM   v$locked_object l, 
dba_objects o, 
v$session s 
WHERE  l.object_id = o.object_id 
AND    l.session_id = s.sid 
ORDER BY l.session_id;;;

---------------------------------------------------
-- 如果你更关心哪些会话正在等待锁，以及它们在等待什么锁，可以使用 v$lock 和 v$session
---------------------------------------------------
SELECT
    s_wait.sid AS waiting_session_id,
    s_wait.serial# AS waiting_serial#,
    s_wait.username AS waiting_username,
    s_wait.osuser AS waiting_osuser,
    s_wait.machine AS waiting_machine,
    s_wait.program AS waiting_program,
    s_wait.sql_id AS waiting_sql_id,          -- 等待会话当前执行的 SQL
    s_wait.event AS waiting_event,            -- 等待事件，通常是 'enq: TX - row lock contention' 等
    s_hold.sid AS holding_session_id,
    s_hold.serial# AS holding_serial#,
    s_hold.username AS holding_username,
    s_hold.osuser AS holding_osuser,
    s_hold.machine AS holding_machine,
    s_hold.program AS holding_program,
    s_hold.sql_id AS holding_sql_id,          -- 持有锁会话当前执行的 SQL
    s_hold.prev_sql_id AS holding_prev_sql_id, -- 持有锁会话上一个执行的 SQL (可能更相关)
    o.owner AS locked_object_owner,
    o.object_name AS locked_object_name,
    o.object_type AS locked_object_type,
    l_hold.type AS lock_type,                  -- 锁类型 (例如 'TX', 'TM')
    DECODE(l_hold.lmode,
        0, 'None', 1, 'Null', 2, 'Row-S (SS)', 3, 'Row-X (SX)',
        4, 'Share', 5, 'S/Row-X (SSX)', 6, 'Exclusive (X)',
        TO_CHAR(l_hold.lmode)
    ) AS holding_mode_desc,
    DECODE(l_wait.request,
        0, 'None', 1, 'Null', 2, 'Row-S (SS)', 3, 'Row-X (SX)',
        4, 'Share', 5, 'S/Row-X (SSX)', 6, 'Exclusive (X)',
        TO_CHAR(l_wait.request)
    ) AS requested_mode_desc,
    l_hold.ctime AS lock_hold_time_seconds -- 锁持有时间 (秒)
FROM
    v$lock l_wait
JOIN
    v$session s_wait ON l_wait.sid = s_wait.sid
JOIN
    v$lock l_hold ON l_wait.id1 = l_hold.id1 AND l_wait.id2 = l_hold.id2 AND l_hold.request = 0 -- l_hold 是持有者
LEFT JOIN
    v$session s_hold ON l_hold.sid = s_hold.sid
LEFT JOIN
    dba_objects o ON s_wait.row_wait_obj# = o.object_id  -- 对于行锁，通过 row_wait_obj# 关联对象
WHERE
    l_wait.request > 0 -- l_wait 是等待者

;;;

---------------------------------------------------
-- 计算当前有多少个锁
---------------------------------------------------
SELECT COUNT(*) AS total_locks FROM v$locked_object;;;

---------------------------------------------------
-- 或者按对象类型统计锁数量：
---------------------------------------------------

SELECT o.object_type, COUNT(*) AS lock_count
FROM v$locked_object l
JOIN dba_objects o ON l.object_id = o.object_id
GROUP BY o.object_type
;;;

---------------------------------------------------
-- 统计当前不同类型对象的锁数量
---------------------------------------------------
SELECT
    o.owner AS object_owner,
    o.object_type,
    DECODE(l.locked_mode, 0,'None',1,'Null',2,'Row-S (SS)',3,'Row-X (SX)',4,'Share',5,'S/Row-X (SSX)',6,'Exclusive (X)',TO_CHAR(l.locked_mode)) AS locked_mode_desc,
    COUNT(*) as lock_instance_count
FROM
    v$locked_object l
JOIN
    dba_objects o ON l.object_id = o.object_id
GROUP BY
    o.owner, o.object_type, DECODE(l.locked_mode, 0,'None',1,'Null',2,'Row-S (SS)',3,'Row-X (SX)',4,'Share',5,'S/Row-X (SSX)',6,'Exclusive (X)',TO_CHAR(l.locked_mode));;;

---------------------------------------------------
--- 锁等待快照
---------------------------------------------------
SELECT
    s_wait.sid AS waiting_session_id, s_wait.serial# AS waiting_serial#, s_wait.username AS waiting_username, s_wait.osuser AS waiting_osuser, s_wait.machine AS waiting_machine, s_wait.program AS waiting_program, s_wait.sql_id AS waiting_sql_id, s_wait.event AS waiting_event,
    s_hold.sid AS holding_session_id, s_hold.serial# AS holding_serial#, s_hold.username AS holding_username, s_hold.osuser AS holding_osuser, s_hold.machine AS holding_machine, s_hold.program AS holding_program, s_hold.sql_id AS holding_sql_id, s_hold.prev_sql_id AS holding_prev_sql_id,
    o.owner AS locked_object_owner, o.object_name AS locked_object_name, o.object_type AS locked_object_type,
    l_hold.type AS lock_type,
    DECODE(l_hold.lmode, 0,'None',1,'Null',2,'Row-S (SS)',3,'Row-X (SX)',4,'Share',5,'S/Row-X (SSX)',6,'Exclusive (X)',TO_CHAR(l_hold.lmode)) AS holding_mode_desc,
    DECODE(l_wait.request, 0,'None',1,'Null',2,'Row-S (SS)',3,'Row-X (SX)',4,'Share',5,'S/Row-X (SSX)',6,'Exclusive (X)',TO_CHAR(l_wait.request)) AS requested_mode_desc,
    l_hold.ctime AS lock_hold_time_seconds,
    (SELECT sql_text FROM v$sqlarea WHERE sql_id = s_wait.sql_id AND ROWNUM = 1) AS waiting_sql_text, -- 获取等待SQL文本
    (SELECT sql_text FROM v$sqlarea WHERE sql_id = s_hold.prev_sql_id AND ROWNUM = 1) AS holding_sql_text -- 获取持有锁SQL文本
FROM v$lock l_wait
JOIN v$session s_wait ON l_wait.sid = s_wait.sid
JOIN v$lock l_hold ON l_wait.id1 = l_hold.id1 AND l_wait.id2 = l_hold.id2 AND l_hold.request = 0
JOIN v$session s_hold ON l_hold.sid = s_hold.sid
LEFT JOIN dba_objects o ON s_wait.row_wait_obj# = o.object_id
WHERE l_wait.request > 0;;;

---------------------------------------------------
-- Top N 被锁对象 (按被等待次数)
---------------------------------------------------
SELECT * FROM (
    SELECT
        o.owner AS locked_object_owner,
        o.object_name AS locked_object_name,
        o.object_type AS locked_object_type,
        COUNT(DISTINCT l_wait.sid) AS waiting_on_object_count -- 有多少不同会话在等待这个对象上的锁
    FROM
        v$lock l_wait
    JOIN
        v$session s_wait ON l_wait.sid = s_wait.sid
    LEFT JOIN
        dba_objects o ON s_wait.row_wait_obj# = o.object_id
    WHERE
        l_wait.request > 0 AND s_wait.row_wait_obj# IS NOT NULL
    GROUP BY
        o.owner, o.object_name, o.object_type
    ORDER BY
        waiting_on_object_count DESC
) WHERE ROWNUM <= 5;;;

SELECT
    o.object_type,                             -- Tag: 对象类型 (基数低)
    DECODE(l.locked_mode,
        0, 'None', 1, 'Null', 2, 'Row-S (SS)', 3, 'Row-X (SX)',
        4, 'Share', 5, 'S/Row-X (SSX)', 6, 'Exclusive (X)',
        TO_CHAR(l.locked_mode)
    ) AS locked_mode_desc,                    -- Tag: 锁模式描述 (基数低)
    COUNT(*) AS lock_instance_count            -- Field: 该类型和模式下的锁实例数量
FROM
    v$locked_object l
JOIN
    dba_objects o ON l.object_id = o.object_id
GROUP BY
    o.object_type,
    DECODE(l.locked_mode,
        0, 'None', 1, 'Null', 2, 'Row-S (SS)', 3, 'Row-X (SX)',
        4, 'Share', 5, 'S/Row-X (SSX)', 6, 'Exclusive (X)',
        TO_CHAR(l.locked_mode)
    );;;


---
-- 这个查询统计当前有多少会话在等待锁，并按等待的事件（通常能反映锁的类型）进行聚合。
--
SELECT
    s.event AS waiting_event,                 -- Tag: 等待事件 (基数中等，但很有意义)
    COUNT(DISTINCT s.sid) AS waiting_session_count -- Field: 等待该事件的会话数量
FROM
    v$session s
WHERE
    s.wait_time = 0                           -- 表示当前正在等待
    AND s.event LIKE 'enq: %'                 -- 筛选与锁相关的 enqueue 等待事件
    -- 可以根据需要添加其他锁相关的等待事件，例如 'latch: %' 如果也想监控 Latch
GROUP BY
    s.event;;;


-------------
-- 这个查询会列出所有当前持有的锁（REQUEST = 0），无论它们是否正在阻塞其他会话
-------------
SELECT
    s.sid AS locking_session_id,
    s.serial# AS locking_serial#,
    s.username AS locking_oracle_user,
    s.osuser AS locking_os_user,
    s.machine AS locking_machine,
    s.program AS locking_program,
    s.status AS session_status,
    s.sql_id AS current_sql_id,
    s.prev_sql_id AS previous_sql_id,
    lk.type AS lock_type,                      -- 锁类型 (e.g., 'TM', 'TX', 'UL')
    DECODE(lk.lmode,                           -- 持有模式
        0, 'None (0)', 1, 'Null (1)', 2, 'Row Share (SS) (2)', 3, 'Row Exclusive (SX) (3)',
        4, 'Share (S) (4)', 5, 'Share Row Exclusive (SSX) (5)', 6, 'Exclusive (X) (6)',
        TO_CHAR(lk.lmode)
    ) AS lock_mode_held,
    lk.id1,                                    -- 锁标识符1 (对于TM锁是object_id, 对于TX锁是undo段号和槽号)
    lk.id2,                                    -- 锁标识符2 (对于TX锁是序列号)
    lk.ctime AS lock_hold_time_seconds,        -- 锁已持有的时间 (秒)
    -- 尝试获取被锁对象信息 (对于 TM 锁可以直接用 id1)
    (SELECT o.owner || '.' || o.object_name || ' (' || o.object_type || ')'
     FROM dba_objects o
     WHERE lk.type = 'TM' AND o.object_id = lk.id1 AND ROWNUM = 1) AS tm_locked_object,
    -- 对于 TX 锁，如果想看具体的行，需要更复杂的查询，通常通过 v$session.row_wait_obj# 等
    s.blocking_session AS is_blocking_other_session, -- 如果该会话正在阻塞其他会话，这里是被阻塞会话的SID
    s.blocking_session_status AS is_blocking_status,
    TO_CHAR(s.logon_time, 'YYYY-MM-DD HH24:MI:SS') AS logon_time
FROM
    v$lock lk
JOIN
    v$session s ON lk.sid = s.sid
WHERE
    lk.lmode > 0  -- lmode > 0 表示持有锁 (request = 0 也可以，但 lmode > 0 更直接)
    -- AND lk.type IN ('TM', 'TX') -- 可以根据需要筛选特定类型的锁
ORDER BY
    s.sid, lk.type, lk.id1;;;

------------------
-- 直接显示被锁定的对象信息
-- 包含了会话的详细信息
------------------
SELECT
    s.sid AS locking_session_id,
    s.serial# AS locking_serial#,
    s.username AS locking_oracle_user,
    s.osuser AS locking_os_user,
    s.machine AS locking_machine,
    s.program AS locking_program,
    s.status AS session_status,
    s.sql_id AS current_sql_id,          -- 会话当前正在执行的SQL ID
    s.prev_sql_id AS previous_sql_id,    -- 会话上一个执行的SQL ID (可能更相关)
    o.owner AS locked_object_owner,
    o.object_name AS locked_object_name,
    o.object_type AS locked_object_type,
    DECODE(l.locked_mode,
        0, 'None (0)',
        1, 'Null (1)',
        2, 'Row Share (SS) (2)',
        3, 'Row Exclusive (SX) (3)',
        4, 'Share (S) (4)',
        5, 'Share Row Exclusive (SSX) (5)',
        6, 'Exclusive (X) (6)',
        TO_CHAR(l.locked_mode)
    ) AS lock_mode_held,
    l.oracle_username AS locked_by_oracle_user, -- 通常与 s.username 相同
    l.process AS os_process_id_client,    -- 客户端进程ID (可能不准确或为空)
    TO_CHAR(s.logon_time, 'YYYY-MM-DD HH24:MI:SS') AS logon_time,
    TRUNC(s.last_call_et / 3600) || 'h ' ||
    TRUNC(MOD(s.last_call_et, 3600) / 60) || 'm ' ||
    MOD(s.last_call_et, 60) || 's' AS last_call_et_readable -- 自上次活动以来的时间
FROM
    v$locked_object l
JOIN
    dba_objects o ON l.object_id = o.object_id
JOIN
    v$session s ON l.session_id = s.sid
ORDER BY
    s.sid, o.object_name

;;;

----------------------
-- 方案一：按等待事件类型聚合 (最常用)
-- 这个查询统计当前有多少会话在等待特定类型的锁/资源。

SELECT
    s.event AS waiting_event,      -- Tag: 等待事件 (例如 'enq: TX - row lock contention', 'latch: shared pool')
    COUNT(*) AS waiting_session_count -- Field: 等待该事件的会话数量
FROM
    v$session s
WHERE
    s.wait_time = 0                -- 表示当前正在等待
    --AND (
    --        s.event LIKE 'enq: %'  -- 等待 enqueue (常见的行锁、表锁等)
    --     OR s.event LIKE 'latch:%' -- 等待 latch
    --     OR s.event LIKE 'buffer busy waits' -- 等待数据块
    --     OR s.event LIKE 'row cache lock'    -- 等待行缓存锁
    --     -- 可以根据需要添加其他你关心的与“锁”或资源争用相关的等待事件
    --    )
GROUP BY
    s.event
ORDER BY
    waiting_session_count DESC;;;
