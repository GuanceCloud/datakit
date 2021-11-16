package time

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestDuration(t *testing.T) {
	d := Duration{Duration: time.Second}
	assert.Equal(t, "1s", d.UnitString(time.Second))
	assert.Equal(t, "1000000000ns", d.UnitString(time.Nanosecond))
	assert.Equal(t, "1000000mics", d.UnitString(time.Microsecond))
	assert.Equal(t, "1000ms", d.UnitString(time.Millisecond))
	assert.Equal(t, "0m", d.UnitString(time.Minute))
	assert.Equal(t, "0h", d.UnitString(time.Hour))
}
