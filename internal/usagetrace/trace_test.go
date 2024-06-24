// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package usagetrace

import (
	"encoding/json"
	"os"
	"runtime"
	"sync"
	T "testing"
	"time"

	"github.com/GuanceCloud/cliutils"
	"github.com/stretchr/testify/assert"
)

type refresherMock struct {
	ut *usageTrace
}

func (r *refresherMock) UsageTrace(body []byte) error {
	var ut usageTrace
	if err := json.Unmarshal(body, &ut); err != nil {
		return err
	} else {
		r.ut = &ut
		return nil
	}
}

func TestUsageTrace(t *T.T) {
	t.Run("update-trace-options", func(t *T.T) {
		r := &refresherMock{}
		ch := make(chan any)
		wg := sync.WaitGroup{}

		wg.Add(1)
		go func() {
			defer wg.Done()
			doStart(
				WithReservedInputs("rum", "kafkamq"),
				WithRefresher(r),
				WithRefreshDuration(time.Second),
				WithExitChan(ch),
			)
		}()

		time.Sleep(3 * time.Second) // wait refresher ok
		assert.Empty(t, r.ut.Inputs)
		assert.Empty(t, r.ut.ServerListens)

		UpdateTraceOptions(WithInputNames("rum", "some-xxx"), WithServerListens("1.2.3.4:1234"))

		time.Sleep(3 * time.Second) // wait refresher ok

		assert.Equal(t, "rum", r.ut.Inputs[0])
		assert.Len(t, r.ut.Inputs, 1)
		assert.Equal(t, "1.2.3.4:1234", r.ut.ServerListens[0])

		t.Logf("usage: %+#v", r.ut)

		close(ch)
		wg.Wait()
	})

	t.Run("clear-trace-options", func(t *T.T) {
		r := &refresherMock{}
		ch := make(chan any)
		wg := sync.WaitGroup{}

		wg.Add(1)
		go func() {
			defer wg.Done()
			doStart(
				WithReservedInputs("rum", "kafkamq"),
				WithRefresher(r),
				WithRefreshDuration(time.Second),
				WithExitChan(ch),

				WithInputNames("rum"),
				WithServerListens("1.2.3.4:1234"),
			)
		}()

		time.Sleep(3 * time.Second) // wait refresher ok

		assert.Equal(t, "rum", r.ut.Inputs[0])
		assert.Equal(t, "1.2.3.4:1234", r.ut.ServerListens[0])

		ClearInputNames()
		ClearServerListens()

		time.Sleep(3 * time.Second) // wait refresher ok

		assert.Empty(t, r.ut.Inputs)
		assert.Empty(t, r.ut.ServerListens)

		t.Logf("usage: %+#v", r.ut)

		close(ch)
		wg.Wait()
	})

	t.Run(`basic`, func(t *T.T) {
		r := &refresherMock{}
		ch := make(chan any)
		wg := sync.WaitGroup{}
		runtimeID := cliutils.XID("run_")

		wg.Add(1)
		go func() {
			defer wg.Done()
			doStart(
				WithReservedInputs("rum", "kafkamq"),
				WithRefresher(r),
				WithRefreshDuration(time.Second),
				WithExitChan(ch),
				WithDatakitRuntimeID(runtimeID),
				WithDatakitHostname("my-host"),
				WithWorkspaceToken("tkn_xxx"),
				WithMainIP("1.2.3.4"),
				WithDatakitPodname("datakit-xyz"),
				WithCPULimits(3.14),
				WithInputNames("rum"),
				WithDatakitStartTime(0),
				WithDCAAPIServer("dca-host:9530"),
				WithDatakitPodname("dk-pod-xxx"),
			)
		}()

		time.Sleep(3 * time.Second) // wait refresher ok

		t.Logf("usage: %+#v", r.ut)

		assert.Equal(t, 3, r.ut.UsageCores) // with rum input enabled, consume 3(3.14 -> int) cores
		assert.Equal(t, runtimeID, r.ut.RuntimeID)
		assert.Equal(t, "my-host", r.ut.Host)
		assert.Equal(t, "tkn_xxx", r.ut.Token)
		assert.Equal(t, "dca-host:9530", r.ut.DCAServer)
		assert.Equal(t, runtime.GOOS, r.ut.OS)
		assert.Equal(t, runtime.GOARCH, r.ut.Arch)
		assert.Equal(t, "dk-pod-xxx", r.ut.PodName)

		close(ch)
		wg.Wait()
	})
}

func Test_checkLoopbackServerListen(t *T.T) {
	t.Run("1.2.3.4:1234", func(t *T.T) {
		ok, err := checkLoopbackServerListen("1.2.3.4:1234")
		assert.NoError(t, err)
		assert.False(t, ok)
	})

	t.Run("tcp://1.2.3.4:1234", func(t *T.T) {
		ok, err := checkLoopbackServerListen("tcp://1.2.3.4:1234")
		assert.NoError(t, err)
		assert.False(t, ok)
	})

	t.Run("udp://1.2.3.4:1234", func(t *T.T) {
		ok, err := checkLoopbackServerListen("udp://1.2.3.4:1234")
		assert.NoError(t, err)
		assert.False(t, ok)
	})

	t.Run("udp://1.2.3:1234", func(t *T.T) {
		ok, err := checkLoopbackServerListen("udp://1.2.3:1234") // invalid IP
		assert.Error(t, err)
		assert.False(t, ok)
	})

	t.Run("udp://1.2.3.4:0", func(t *T.T) {
		ok, err := checkLoopbackServerListen("udp://1.2.3.4:0")
		assert.NoError(t, err)
		assert.False(t, ok)
	})

	t.Run(":1234", func(t *T.T) {
		ok, err := checkLoopbackServerListen(":1234")
		assert.NoError(t, err)
		assert.False(t, ok)
	})

	t.Run("udp://127.0.0.1:0", func(t *T.T) {
		ok, err := checkLoopbackServerListen("udp://127.0.0.1:0")
		assert.NoError(t, err)
		assert.True(t, ok)
	})

	t.Run("udp://localhost:0", func(t *T.T) {
		ok, err := checkLoopbackServerListen("udp://localhost:0") // invalid IP
		assert.NoError(t, err)
		assert.True(t, ok)
	})

	t.Run("6666", func(t *T.T) {
		ok, err := checkLoopbackServerListen("6666") // invalid port
		assert.Error(t, err)
		assert.False(t, ok)
	})

	t.Run("0.0.0.0:66666", func(t *T.T) {
		ok, err := checkLoopbackServerListen("0.0.0.0:66666") // invalid port
		assert.Error(t, err)
		assert.False(t, ok)
	})

	t.Run("0.0.0.0:1234", func(t *T.T) {
		ok, err := checkLoopbackServerListen("0.0.0.0:1234")
		assert.NoError(t, err)
		assert.False(t, ok)
	})

	t.Run("localhost:9529", func(t *T.T) {
		ok, err := checkLoopbackServerListen("localhost:9529")
		assert.NoError(t, err)
		assert.True(t, ok)
	})

	t.Run("127.0.0.1:9529", func(t *T.T) {
		ok, err := checkLoopbackServerListen("127.0.0.1:9529")
		assert.NoError(t, err)
		assert.True(t, ok)
	})

	t.Run("127.x", func(t *T.T) {
		ok, err := checkLoopbackServerListen("127.1.1.1:9529")
		assert.NoError(t, err)
		assert.True(t, ok)
	})

	t.Run("domain-socket", func(t *T.T) {
		f, err := os.CreateTemp("", "uds.sock")
		assert.NoError(t, err)
		defer os.Remove(f.Name())

		ok, err := checkLoopbackServerListen(f.Name())
		assert.NoError(t, err)
		assert.True(t, ok)
	})

	t.Run("loopback-ipv6", func(t *T.T) {
		ok, err := checkLoopbackServerListen("[::1]:9529")
		assert.NoError(t, err)
		assert.True(t, ok)
	})

	t.Run("all-ipv6", func(t *T.T) {
		ok, err := checkLoopbackServerListen("[::]:9529")
		assert.NoError(t, err)
		assert.False(t, ok)
	})
}
