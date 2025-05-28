SELECT
    p.spid AS os_process_id,          -- Tag: 服务器进程的操作系统 PID
    s.username AS oracle_username,    -- Tag: (可选) Oracle 用户名，帮助识别来源
    s.program AS client_program,      -- Tag: (可选) 客户端程序名
    s.event AS waiting_event_name,    -- Tag: 等待事件的名称
		s.type AS session_type,
    COUNT(*) AS current_waiting_sessions_count -- Field: 当前有多少会话（及其关联进程）正在等待这个特定事件
FROM
    v$session s
JOIN
    v$process p ON s.paddr = p.addr -- 通过 paddr 关联 v$session 和 v$process
WHERE
    s.wait_time = 0                -- 表示当前正在等待
    AND s.status = 'ACTIVE'        -- 通常只关心活动会话的等待
    --AND s.type != 'BACKGROUND'   -- 只关心用户/应用会话，排除后台进程的会话视图
    AND s.event IS NOT NULL        -- 确保事件不是 NULL (例如，CPU on)
    AND s.event NOT IN (           -- 排除常见的空闲等待事件
        'SQL*Net message from client',
        'SQL*Net message to client',
        'pipe get',
        'rdbms ipc message',
        'pmon timer',
        'smon timer',
        'PX Idle Wait',
        'jobq slave wait',
        'Streams AQ: waiting for messages in the queue',
      'Streams AQ: qmn coordinator waiting for slave to start',
      'Streams AQ: qmn slave idle wait',
      'DIAG idle wait',
      'class slave wait',
      'wait for unread message on broadcast channel',
      'ges remote message',
      'gcs remote message',
      'gcs_wait_for_msg',
      'master wait',
      'KSV master wait'
    --    -- 根据你的环境和关注点，可能需要调整这个排除列表
    )
GROUP BY
    p.spid,
    s.username,
    s.program,
		s.type,
    s.event
HAVING
    COUNT(*) > 0 -- 只输出当前确实有等待的组合
ORDER BY
    current_waiting_sessions_count DESC, p.spid, s.event;;;
