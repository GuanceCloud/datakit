-- 计算过去 N 分钟的平均活动会话数 (AAS)，即近似的 DB Time Per Sec
-- 这个查询更适合在AWR数据或ASH历史数据上运行以获得平滑值
-- 对于实时V$ACTIVE_SESSION_HISTORY，它反映的是非常近期的活动

SELECT
    COUNT(*) /
    (
        EXTRACT(DAY FROM (MAX(sample_time) - MIN(sample_time))) * 24 * 60 * 60 +
        EXTRACT(HOUR FROM (MAX(sample_time) - MIN(sample_time))) * 60 * 60 +
        EXTRACT(MINUTE FROM (MAX(sample_time) - MIN(sample_time))) * 60 +
        EXTRACT(SECOND FROM (MAX(sample_time) - MIN(sample_time)))
    ) AS average_active_sessions
FROM
    v$active_session_history
WHERE
    sample_time >= SYSTIMESTAMP - INTERVAL '1' MINUTE -- 示例：过去1分钟

;;;

SELECT
    COUNT(*) /
    ( (MAX(CAST(sample_time AS DATE)) - MIN(CAST(sample_time AS DATE))) * 24 * 60 * 60 ) AS average_active_sessions
FROM
    v$active_session_history
WHERE
    sample_time >= SYSTIMESTAMP - INTERVAL '1' MINUTE;;;

SELECT
    CASE
        WHEN ( -- 除数不能为0
            EXTRACT(DAY FROM (MAX(sample_time) - MIN(sample_time))) * 24 * 60 * 60 +
            EXTRACT(HOUR FROM (MAX(sample_time) - MIN(sample_time))) * 60 * 60 +
            EXTRACT(MINUTE FROM (MAX(sample_time) - MIN(sample_time))) * 60 +
            EXTRACT(SECOND FROM (MAX(sample_time) - MIN(sample_time)))
        ) = 0 THEN NULL -- 或者返回 0，取决于你希望如何处理这种情况
        ELSE
            COUNT(*) /
            (
                EXTRACT(DAY FROM (MAX(sample_time) - MIN(sample_time))) * 24 * 60 * 60 +
                EXTRACT(HOUR FROM (MAX(sample_time) - MIN(sample_time))) * 60 * 60 +
                EXTRACT(MINUTE FROM (MAX(sample_time) - MIN(sample_time))) * 60 +
                EXTRACT(SECOND FROM (MAX(sample_time) - MIN(sample_time)))
            )
    END AS average_active_sessions
FROM
    v$active_session_history ash
WHERE
    ash.sample_time >= SYSTIMESTAMP - INTERVAL '1' MINUTE -- 监控过去1分钟的ASH数据
    -- 可以根据需要添加其他过滤条件，例如：
    -- AND ash.session_state = 'ON CPU' -- 只看在CPU上的会话
    -- AND ash.session_type = 'FOREGROUND' -- 只看前台用户会话
HAVING -- 确保 MIN 和 MAX sample_time 不同，避免除以零
    MAX(sample_time) > MIN(sample_time)
