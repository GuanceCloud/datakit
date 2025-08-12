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

const colLockedSession = `
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
`

type lockedSession struct {
	Event sql.NullString `db:"WAITING_EVENT"`
	Count int64          `db:"WAITING_SESSION_COUNT"`
}

func (ipt *Input) collectLockedSession() {
	var (
		name  = measurementLockedSession
		rows  = []lockedSession{}
		pts   []*point.Point
		start = time.Now()
	)

	if ipt.isMetricExclude(name) {
		l.Debugf("metric [%s] is excluded, ignored", name)
		return
	}

	if err := selectWrapper(ipt, &rows, colLockedSession, name); err != nil {
		l.Errorf("failed to collect %q: %s", name, err)
		return
	}

	for _, r := range rows {
		kvs := ipt.getKVs()
		kvs = kvs.AddTag("event", r.Event.String).
			Set("waiting_session_count", r.Count)
		pts = append(pts, point.NewPoint(name, kvs, ipt.getKVsOpts(point.Metric)...))
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
