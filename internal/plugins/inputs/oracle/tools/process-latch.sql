------------------------
-- 统计有多少进程当前在等待 Latch
------------------------
SELECT COUNT(*) AS latch_wait_processes
FROM V$PROCESS p
WHERE p.LATCHWAIT IS NOT NULL

;;;

------------------------
-- (更细致的) 统计等待特定Latch的进程数
------------------------
SELECT l.name AS latch_name, COUNT(DISTINCT p.ADDR) AS waiting_processes_count
FROM V$PROCESS p
JOIN V$LATCHHOLDER lh ON p.LATCHWAIT = lh.LADDR -- 需要确认关联方式，或直接从 V$LATCH, V$SESSION_WAIT 获取
JOIN V$LATCHNAME l ON lh.LADDR = l.LATCH# -- 这是一个简化的概念，实际获取等待的latch名更复杂
WHERE p.LATCHWAIT IS NOT NULL
GROUP BY l.name
-- 注意：直接从 V$PROCESS 获取等待的具体 latch 名称比较困难，通常结合 V$SESSION_WAIT 或 V$LATCH 分析。
-- 一个更简单的方式是，如果 LATCHWAIT 非空，就认为进程在等待某种latch。
;;;

SELECT
    ln.name AS latch_name,
    COUNT(DISTINCT s.sid) AS waiting_sessions_count
FROM
    V$SESSION s
JOIN
    V$LATCHNAME ln ON s.p2 = ln.latch# -- s.p2 对于 'latch free' 事件通常是 latch number
WHERE
    s.event = 'latch free' -- 'latch free' 是旧的事件名，现在通常是 'latch: <specific latch>'
    AND s.wait_time = 0  -- 表示当前正在等待
    AND s.p2text = 'number' -- 确认 p2 是 latch number
GROUP BY
    ln.name
ORDER BY
    waiting_sessions_count DESC
;;;

-- 对于现代 Oracle 版本，'latch: <specific latch>' 事件更常见
SELECT
    SUBSTR(s.event, INSTR(s.event, ':') + 2) AS latch_name,
    COUNT(DISTINCT s.sid) AS waiting_sessions_count
FROM
    V$SESSION s
WHERE
    s.event LIKE 'latch:%'
    AND s.wait_time = 0  -- 表示当前正在等待
GROUP BY
    SUBSTR(s.event, INSTR(s.event, ':') + 2)
ORDER BY
    waiting_sessions_count DESC
