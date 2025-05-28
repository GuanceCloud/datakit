-- 按等待事件类型聚合
-- 这个查询统计当前有多少会话在等待特定类型的锁/资源
SELECT
    s.event AS waiting_event,         -- Tag: 等待事件 (例如 'enq: TX - row lock contention', 'latch: shared pool')
    COUNT(*) AS waiting_session_count -- Field: 等待该事件的会话数量
FROM
    v$session s
WHERE
    s.wait_time = 0                -- 表示当前正在等待
GROUP BY
    s.event
ORDER BY
    waiting_session_count DESC
