// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package oracle

import (
	"database/sql"
	"time"

	"github.com/GuanceCloud/cliutils/point"

	dkio "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/io"
)

const colWaitingEvent = `
SELECT
    p.spid AS os_process_id,          -- Tag: 服务器进程的操作系统 PID
    s.username AS oracle_username,    -- Tag: (可选) Oracle 用户名，帮助识别来源
    s.program AS client_program,      -- Tag: (可选) 客户端程序名
    s.event AS waiting_event_name,    -- Tag: 等待事件的名称
    s.type AS type,                   -- Tag: 等待事件的类型
    COUNT(*) AS current_waiting_sessions_count -- Field: 当前有多少会话（及其关联进程）正在等待这个特定事件
FROM
    v$session s
JOIN
    v$process p ON s.paddr = p.addr -- 通过 paddr 关联 v$session 和 v$process
WHERE
    s.wait_time = 0                -- 表示当前正在等待
    AND s.status = 'ACTIVE'        -- 通常只关心活动会话的等待
    -- AND s.type != 'BACKGROUND'     -- 只关心用户/应用会话，排除后台进程的会话视图
    AND s.event IS NOT NULL        -- 确保事件不是 NULL (例如，CPU on)
GROUP BY
    p.spid,
    s.username,
    s.program,
    s.event,
		s.type
HAVING
    COUNT(*) > 0 -- 只输出当前确实有等待的组合
ORDER BY
    current_waiting_sessions_count DESC, p.spid, s.event
`

type waitingEvent struct {
	Event     sql.NullString `db:"WAITING_EVENT_NAME"`
	EventType sql.NullString `db:"TYPE"`
	Username  sql.NullString `db:"ORACLE_USERNAME"`
	ProcessID int64          `db:"OS_PROCESS_ID"`
	Program   sql.NullString `db:"CLIENT_PROGRAM"`
	Count     int64          `db:"CURRENT_WAITING_SESSIONS_COUNT"`
}

func (ipt *Input) collectWaitingEvent() {
	var (
		name  = measurementWaitingEvent
		rows  = []waitingEvent{}
		pts   []*point.Point
		start = time.Now()
	)

	if ipt.isMetricExclude(name) {
		l.Debugf("metric [%s] is excluded, ignored", name)
		return
	}

	if err := selectWrapper(ipt, &rows, colWaitingEvent, name); err != nil {
		l.Errorf("failed to collect %q: %s", name, err)
		return
	}

	for _, r := range rows {
		kvs := ipt.getKVs()
		kvs = kvs.AddTag("event", r.Event.String).
			AddTag("program", r.Program.String).
			AddTag("username", r.Username.String).
			AddTag("event_type", r.EventType.String).
			Set("count", r.Count)
		pts = append(pts, point.NewPoint(name, kvs, ipt.getKVsOpts()...))
	}

	l.Debugf("%s: get %d points", name, len(pts))
	if err := ipt.feeder.Feed(point.Metric,
		pts,
		dkio.WithCollectCost(time.Since(start)),
		dkio.WithElection(ipt.Election),
		dkio.WithSource(inputName)); err != nil {
		l.Warnf("feeder.Feed: %s, ignored", err)
	}
}
