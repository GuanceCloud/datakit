// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package installer

import (
	T "testing"

	"github.com/stretchr/testify/assert"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/config"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/io/dataway"
)

func Test_mergeDefaultInputs(t *T.T) {
	t.Run("empty", func(t *T.T) {
		opt := DefaultInstallArgs()

		expect := []string{
			"cpu",
			"mem",
		}

		assert.Equal(t, expect,
			opt.mergeDefaultInputs([]string{"cpu", "mem"}, nil, false))
	})

	t.Run("disable-all", func(t *T.T) {
		opt := DefaultInstallArgs()
		expect := []string{
			"-cpu",
			"-mem",
		}
		assert.Equal(t, expect,
			opt.mergeDefaultInputs([]string{"cpu", "mem"}, []string{"-"}, false))
	})

	t.Run("enable-some", func(t *T.T) {
		opt := DefaultInstallArgs()
		expect := []string{
			"-mem",
			"cpu",
		}
		assert.Equal(t, expect,
			opt.mergeDefaultInputs([]string{"cpu", "mem"}, []string{"cpu"}, false))
	})

	t.Run("disable-some", func(t *T.T) {
		opt := DefaultInstallArgs()
		expect := []string{
			"-cpu",
			"mem",
		}
		assert.Equal(t, expect,
			opt.mergeDefaultInputs([]string{"cpu", "mem"}, []string{"-cpu"}, false))
	})

	t.Run("disable-and-enable-some", func(t *T.T) {
		opt := DefaultInstallArgs()
		defaultList := []string{
			"cpu",
			"disk",
			"mem",
			"system",
		}
		expect := []string{
			"-cpu",
			"disk",
			"mem",
			"system",
		}

		assert.Equal(t,
			expect,
			opt.mergeDefaultInputs(defaultList, []string{"-cpu", "mem", "disk"}, false))
	})

	t.Run("disable-all-default", func(t *T.T) {
		opt := DefaultInstallArgs()
		defaultList := []string{
			"cpu",
			"disk",
			"mem",
			"system",
		}
		expect := []string{
			"-cpu",
			"-disk",
			"-mem",
			"-system",
			"ddtrace",
		}

		assert.Equal(t,
			expect,
			opt.mergeDefaultInputs(defaultList, []string{"ddtrace"}, false))
	})
}

func TestLoadInstallerArgs(t *T.T) {
	t.Run(`dataway-invalid-token`, func(t *T.T) {
		args := &InstallerArgs{
			DatawayURLs: "https://abc.com?token=tkn_1234",
		}

		mc := config.DefaultConfig()
		_, err := args.LoadInstallerArgs(mc)
		assert.ErrorIs(t, err, dataway.ErrInvalidToken)
		t.Logf("LoadInstallerArgs: %s", err)
	})

	t.Run(`dataway-empty-token`, func(t *T.T) {
		args := &InstallerArgs{
			DatawayURLs: "https://abc.com",
		}

		mc := config.DefaultConfig()
		_, err := args.LoadInstallerArgs(mc)
		assert.ErrorIs(t, err, dataway.ErrEmptyToken)

		t.Logf("LoadInstallerArgs: %s", err)
	})

	t.Run(`multi-dataway`, func(t *T.T) {
		args := &InstallerArgs{
			DatawayURLs: "https://abc.com?token=tkn_11111111111111111111111111111111,https://abc.com?token=tkn_22222222222222222222222222222222",
		}

		mc := config.DefaultConfig()
		_, err := args.LoadInstallerArgs(mc)
		assert.NoError(t, err)

		assert.Equal(t, "tkn_11111111111111111111111111111111", mc.Dataway.Token)

		tkns := mc.Dataway.GetTokens()
		assert.Equal(t, []string{"tkn_11111111111111111111111111111111", "tkn_22222222222222222222222222222222"}, tkns)
	})

	t.Run(`apm-inject`, func(t *T.T) {
		args := &InstallerArgs{
			InstrumentationEnabled: "yes",
		}
		mc := config.DefaultConfig()
		mc, err := args.LoadInstallerArgs(mc)
		assert.NoError(t, err)
		assert.Equal(t, "/var/run/datakit/datakit.sock", mc.HTTPAPI.ListenSocket)
		assert.Equal(t, "yes", mc.APMInject.InstrumentationEnabled)
	})

	t.Run(`resource-limit-cpu-cores`, func(t *T.T) {
		args := &InstallerArgs{
			LimitCPUCores: 1,
		}

		mc := config.DefaultConfig()
		memMax := mc.ResourceLimitOptions.MemMax

		mc, err := args.LoadInstallerArgs(mc)
		assert.NoError(t, err)

		assert.Equal(t, 1.0, mc.ResourceLimitOptions.CPUCores)

		assert.Equal(t, memMax, mc.ResourceLimitOptions.MemMax)
	})

	t.Run(`resource-limit-cpu-max`, func(t *T.T) {
		args := &InstallerArgs{
			LimitCPUMax: 10.0,
		}
		mc := config.DefaultConfig()
		mc, err := args.LoadInstallerArgs(mc)
		assert.NoError(t, err)
		memMax := mc.ResourceLimitOptions.MemMax

		assert.Equal(t, 10.0, mc.ResourceLimitOptions.CPUMax)
		assert.Equal(t, memMax, mc.ResourceLimitOptions.MemMax)
	})

	t.Run(`resource-limit-cpu-args-not-set`, func(t *T.T) {
		args := &InstallerArgs{
			LimitCPUMax: 0,
		}

		mc := config.DefaultConfig()
		memMax := mc.ResourceLimitOptions.MemMax

		mc.ResourceLimitOptions.CPUMax = 7.0

		mc, err := args.LoadInstallerArgs(mc)
		assert.NoError(t, err)

		assert.Equal(t, 7.0, mc.ResourceLimitOptions.CPUMax)
		assert.Equal(t, memMax, mc.ResourceLimitOptions.MemMax) // max-mem not changed
	})

	t.Run(`double-load`, func(t *T.T) {
		var (
			args = &InstallerArgs{
				LimitCPUMax: 0,
			}
			err error
			mc  = config.DefaultConfig()
		)

		defMemMax := mc.ResourceLimitOptions.MemMax
		defCpuCores := mc.ResourceLimitOptions.CPUCores

		mc, err = args.LoadInstallerArgs(mc) // 1st load within  main.go/applyFlags()
		assert.NoError(t, err)
		assert.Equal(t, defMemMax, mc.ResourceLimitOptions.MemMax)     // not changed
		assert.Equal(t, defCpuCores, mc.ResourceLimitOptions.CPUCores) // not changed

		mc.ResourceLimitOptions.CPUMax = 7.0
		mc.ResourceLimitOptions.MemMax = 11

		mc, err = args.LoadInstallerArgs(mc) // 2nd load within upgrade.go/Upgrade()
		assert.NoError(t, err)

		t.Logf("ResourceLimitOptions: %+#v", mc.ResourceLimitOptions)

		assert.Equal(t, 7.0, mc.ResourceLimitOptions.CPUMax)
		assert.Equal(t, int64(11), mc.ResourceLimitOptions.MemMax)
	})

	t.Run(`resource-limit-mem`, func(t *T.T) {
		args := &InstallerArgs{
			LimitMemMax: 100,
		}
		mc := config.DefaultConfig()
		mc, err := args.LoadInstallerArgs(mc)
		assert.NoError(t, err)

		assert.Equal(t, int64(100), mc.ResourceLimitOptions.MemMax)
	})

	t.Run(`9529-http-listen-only-port`, func(t *T.T) {
		args := &InstallerArgs{
			HTTPPort: 4321,
		}
		mc := config.DefaultConfig()
		mc, err := args.LoadInstallerArgs(mc)
		assert.NoError(t, err)
		assert.Equal(t, "127.0.0.1:4321", mc.HTTPAPI.Listen)
	})

	t.Run(`9529-http-listen-localhost`, func(t *T.T) {
		args := &InstallerArgs{
			HTTPListen: "localhost",
			HTTPPort:   4321,
		}
		mc := config.DefaultConfig()
		mc, err := args.LoadInstallerArgs(mc)
		assert.NoError(t, err)
		assert.Equal(t, "localhost:4321", mc.HTTPAPI.Listen)
	})

	t.Run(`9529-http-listen-ipv6`, func(t *T.T) {
		args := &InstallerArgs{
			HTTPListen: "::1",
			HTTPPort:   4321,
		}
		mc := config.DefaultConfig()
		mc, err := args.LoadInstallerArgs(mc)
		assert.NoError(t, err)
		assert.Equal(t, "::1:4321", mc.HTTPAPI.Listen)
	})

	t.Run(`set-user`, func(t *T.T) {
		args := DefaultInstallArgs()

		mc := config.DefaultConfig()
		mc.DatakitUser = "not-admin"

		mc, err := args.LoadInstallerArgs(mc)
		assert.NoError(t, err)
		assert.Equal(t, "not-admin", mc.DatakitUser)

		// set new user
		args.FlagUserName = "another-admin"
		mc, err = args.LoadInstallerArgs(mc)
		assert.NoError(t, err)
		assert.Equal(t, "another-admin", mc.DatakitUser)
	})
}
