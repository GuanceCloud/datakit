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
}

func Test_loadInstallerArgs(t *T.T) {
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
		mc, err := args.LoadInstallerArgs(mc)
		assert.NoError(t, err)

		assert.Equal(t, 1.0, mc.ResourceLimitOptions.CPUCores)

		assert.Equal(t, args.LimitMemMax, mc.ResourceLimitOptions.MemMax)
	})

	t.Run(`resource-limit-cpu-max`, func(t *T.T) {
		args := &InstallerArgs{
			LimitCPUMax: 10.0,
		}
		mc := config.DefaultConfig()
		mc, err := args.LoadInstallerArgs(mc)
		assert.NoError(t, err)

		assert.Equal(t, 10.0, mc.ResourceLimitOptions.CPUMax)
		assert.Equal(t, args.LimitMemMax, mc.ResourceLimitOptions.MemMax)
	})

	t.Run(`resource-limit-mem`, func(t *T.T) {
		args := &InstallerArgs{
			LimitMemMax: 100,
		}
		mc := config.DefaultConfig()
		mc, err := args.LoadInstallerArgs(mc)
		assert.NoError(t, err)

		assert.Equal(t, int64(100), mc.ResourceLimitOptions.MemMax)

		assert.Equal(t, args.LimitCPUCores, mc.ResourceLimitOptions.CPUCores)
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
}
