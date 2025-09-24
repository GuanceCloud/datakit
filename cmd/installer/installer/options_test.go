// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package installer

import (
	"fmt"
	"runtime"
	T "testing"

	"github.com/stretchr/testify/assert"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/config"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/datakit"
)

func TestUpdateDownloadURLs(t *T.T) {
	t.Run(`basic`, func(t *T.T) {
		arg := DefaultInstallArgs()

		arg.DataKitBaseURL = "zhuyun-static-files-testing.oss-cn-hangzhou.aliyuncs.com/datakit-tw"
		arg.BrandURL = "zhuyun-static-files-testing.oss-cn-hangzhou.aliyuncs.com"
		arg.DataKitVersion = "1.2.3"
		assert.NoError(t, arg.UpdateDownloadURLs())

		t.Logf("args: %+#v", arg)

		assert.Equal(t,
			fmt.Sprintf("https://%s/datakit-%s-%s-%s.tar.gz",
				arg.DataKitBaseURL, runtime.GOOS, runtime.GOARCH, arg.DataKitVersion),
			arg.DistDatakitURL)

		assert.Equal(t,
			fmt.Sprintf("https://%s/dd-image/dd-java-agent.jar", arg.BrandURL),
			arg.DistDatakitAPMInjJavaLibURL)
	})

	t.Run(`with-dist-url`, func(t *T.T) {
		arg := DefaultInstallArgs()

		arg.DataKitBaseURL = "zhuyun-static-files-testing.oss-cn-hangzhou.aliyuncs.com/datakit-tw"
		arg.DataKitVersion = "1.2.3"
		arg.DistBaseURL = "http://some.nginx.com"
		assert.NoError(t, arg.UpdateDownloadURLs())

		t.Logf("args: %+#v", arg)

		assert.Equal(t,
			fmt.Sprintf("%s/datakit-%s-%s-%s.tar.gz",
				arg.DistBaseURL, runtime.GOOS, runtime.GOARCH, arg.DataKitVersion),
			arg.DistDatakitURL)

		assert.Equal(t, fmt.Sprintf("%s/apm_lib/dd-java-agent.jar", arg.DistBaseURL), arg.DistDatakitAPMInjJavaLibURL)
	})
}

func TestSetupService(t *T.T) {
	t.Run(`setup-service`, func(t *T.T) {
		args := DefaultInstallArgs()

		svc, err := args.SetupService()
		assert.NoError(t, err)
		_ = svc
		t.Logf("service: %s", svc.String())
	})

	t.Run(`no-cpu-mem-limit:use-default`, func(t *T.T) {
		if runtime.GOOS != datakit.OSLinux {
			t.Skip("only test under linux")
		}

		args := DefaultInstallArgs()

		args.FlagUserName = "not-admin"

		scfg := args.setupServiceOptions()
		t.Logf("service config: %+#v", scfg)

		rl := config.DefaultConfig().ResourceLimitOptions
		rl.Setup()

		assert.Equal(t, fmt.Sprintf("%f%%", rl.CPUMax()), scfg.Option["CPUQuota"])
		assert.Equal(t, fmt.Sprintf("%dM", rl.MemMax), scfg.Option["MemoryLimit"])
	})

	t.Run(`cpu-mem-limit:use-args`, func(t *T.T) {
		if runtime.GOOS != datakit.OSLinux {
			t.Skip("only test under linux")
		}

		args := DefaultInstallArgs()
		args.FlagUserName = "not-admin"

		args.LimitMemMax = 1024
		args.LimitCPUMax = 10.0

		scfg := args.setupServiceOptions()
		t.Logf("service config: %+#v", scfg)

		assert.Equal(t, fmt.Sprintf("%f%%", args.LimitCPUMax), scfg.Option["CPUQuota"])
		assert.Equal(t, "1024M", scfg.Option["MemoryLimit"])
	})

	t.Run(`cpu-mem-limit:admin-user`, func(t *T.T) {
		if runtime.GOOS != datakit.OSLinux {
			t.Skip("only test under linux")
		}

		args := DefaultInstallArgs()
		args.FlagUserName = "root"

		args.LimitMemMax = 1024
		args.LimitCPUMax = 10.0

		scfg := args.setupServiceOptions()
		t.Logf("service config: %+#v", scfg)

		assert.Equal(t, nil, scfg.Option["CPUQuota"])
		assert.Equal(t, nil, scfg.Option["MemoryLimit"])
	})

	t.Run(`cpu-mem-limit:no-user`, func(t *T.T) {
		if runtime.GOOS != datakit.OSLinux {
			t.Skip("only test under linux")
		}

		args := DefaultInstallArgs()
		args.FlagUserName = ""

		args.LimitMemMax = 1024
		args.LimitCPUMax = 10.0

		scfg := args.setupServiceOptions()
		t.Logf("service config: %+#v", scfg)

		assert.Equal(t, nil, scfg.Option["CPUQuota"])
		assert.Equal(t, nil, scfg.Option["MemoryLimit"])
	})

	t.Run(`cpu-mem-limit-updated-during-non-root-upgrade`, func(t *T.T) {
		if runtime.GOOS != datakit.OSLinux {
			t.Skip("only test under linux")
		}

		args := DefaultInstallArgs()
		args.FlagUserName = "non-admin"

		args.LimitMemMax = 1024
		args.LimitCPUMax = 10.0
		args.FlagDKUpgrade = true

		scfg := args.setupServiceOptions()
		t.Logf("service config: %+#v", scfg)

		assert.Equal(t, fmt.Sprintf("%f%%", args.LimitCPUMax), scfg.Option["CPUQuota"])
		assert.Equal(t, "1024M", scfg.Option["MemoryLimit"])
		assert.True(t, args.shouldReinstallService)
	})
}
