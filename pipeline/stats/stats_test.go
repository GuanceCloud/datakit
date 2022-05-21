package stats

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
)

func TestStats(t *testing.T) {
	stats := Stats{}
	event := &ChangeEvent{
		Name:         "abc.p",
		NS:           "default",
		Category:     datakit.Metric,
		CompileError: "",
		Time:         time.Now(),
	}
	stats.WriteEvent(event)
	e := stats.ReadEvent()
	if len(e) == 1 {
		assert.Equal(t, *event, e[0])
	} else {
		t.Fatal("len(events): ", len(e))
	}

	statsR := ScriptStatsROnly{
		Name:     "abc.p",
		NS:       "default",
		Category: datakit.Metric,
	}
	stats.WriteScriptStats(statsR.Category, statsR.NS, statsR.Name, 0, 0, 0, nil)
	statsRL := stats.ReadStats()
	if len(statsRL) != 0 {
		t.Fatal("len(stats)", len(statsRL))
	}

	stats.UpdateScriptStatsMeta(statsR.Category, statsR.NS, statsR.Name, true, false)
	statsRL = stats.ReadStats()
	if len(statsRL) != 1 {
		t.Fatal("len(stats)", len(statsRL))
	}
	stats.UpdateScriptStatsMeta(statsR.Category, statsR.NS, statsR.Name, true, false)
	statsRL = stats.ReadStats()
	if len(statsRL) != 1 {
		t.Fatal("len(stats)", len(statsRL))
	}
}
