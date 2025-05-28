SELECT
    SUBSTR(s.event, INSTR(s.event, ':') + 2) AS latch_name,
    COUNT(DISTINCT s.sid) AS waiting_sessions_count,
    s.p1text, s.p1, s.p2text, s.p2, s.p3text, s.p3 -- 查看参数，p2 通常是 latch number
FROM
    V$SESSION s
WHERE
    s.event LIKE 'latch:%'
    AND s.wait_time = 0  -- 表示当前正在等待
GROUP BY
    SUBSTR(s.event, INSTR(s.event, ':') + 2), s.p1text, s.p1, s.p2text, s.p2, s.p3text, s.p3
ORDER BY
    waiting_sessions_count DESC

;;;
-- 或者更详细的，查看具体等待的 Latch 地址和编号
SELECT sid, serial#, event, p1raw, p2, p3
FROM v$session_wait
WHERE event LIKE 'latch:%' AND wait_time=0
