package time_test

import (
	"testing"
	"time"

	dktime "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/time"
	"github.com/stretchr/testify/assert"
)

func TestNewAlignedTicker(t *testing.T) {
	t.Run("normal interval", func(t *testing.T) {
		interval := 100 * time.Millisecond

		// Get initial timestamp
		start := time.Now()

		// Create aligned ticker
		ticker := dktime.NewAlignedTicker(interval)
		defer ticker.Stop()

		// Get first tick
		firstTick := <-ticker.C

		// Verify first tick is aligned to interval
		assert.Equal(t, int64(0), firstTick.UnixNano()%int64(interval),
			"First tick should be aligned to interval")

		// Verify first tick happened after start time
		assert.True(t, firstTick.After(start),
			"First tick should be after start time")

		// Get second tick
		secondTick := <-ticker.C

		// Verify interval between ticks
		tickDiff := secondTick.Sub(firstTick)
		assert.Equal(t, interval, tickDiff,
			"Interval between ticks should match specified interval")
	})

	t.Run("minimum interval", func(t *testing.T) {
		interval := time.Nanosecond
		start := time.Now()

		ticker := dktime.NewAlignedTicker(interval)
		defer ticker.Stop()

		firstTick := <-ticker.C
		assert.Equal(t, int64(0), firstTick.UnixNano()%int64(interval))
		assert.True(t, firstTick.After(start))
	})

	t.Run("large interval", func(t *testing.T) {
		interval := time.Second
		start := time.Now()

		ticker := dktime.NewAlignedTicker(interval)
		defer ticker.Stop()

		firstTick := <-ticker.C
		assert.Equal(t, int64(0), firstTick.UnixNano()%int64(interval))
		assert.True(t, firstTick.After(start))

		secondTick := <-ticker.C
		tickDiff := secondTick.Sub(firstTick)
		assert.Equal(t, interval, tickDiff)
	})

	t.Run("multiple ticks", func(t *testing.T) {
		interval := 50 * time.Millisecond
		ticker := dktime.NewAlignedTicker(interval)
		defer ticker.Stop()

		var lastTick time.Time
		for i := 0; i < 5; i++ {
			currentTick := <-ticker.C
			if !lastTick.IsZero() {
				tickDiff := currentTick.Sub(lastTick)
				assert.Equal(t, interval, tickDiff)
			}
			lastTick = currentTick
		}
	})
}
