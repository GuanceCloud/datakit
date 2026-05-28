package cli

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestInformerSyncTimeoutEnv(t *testing.T) {
	t.Setenv(informerSyncTimeoutEnv, "")
	assert.Equal(t, defaultInformerSyncTimeout, informerSyncTimeout())

	t.Setenv(informerSyncTimeoutEnv, "2s")
	assert.Equal(t, 2*time.Second, informerSyncTimeout())

	t.Setenv(informerSyncTimeoutEnv, "3")
	assert.Equal(t, 3*time.Second, informerSyncTimeout())

	t.Setenv(informerSyncTimeoutEnv, "0")
	assert.Equal(t, time.Duration(0), informerSyncTimeout())

	t.Setenv(informerSyncTimeoutEnv, "-1s")
	assert.Equal(t, defaultInformerSyncTimeout, informerSyncTimeout())
}
