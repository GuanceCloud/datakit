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
)

func TestUpdateDownloadURLs(t *T.T) {
	t.Run(`basic`, func(t *T.T) {
		arg := DefaultInstallArgs()

		arg.DataKitBaseURL = "zhuyun-static-files-testing.oss-cn-hangzhou.aliyuncs.com/datakit-tw"
		arg.DataKitVersion = "1.2.3"
		assert.NoError(t, arg.UpdateDownloadURLs())

		t.Logf("args: %+#v", arg)

		assert.Equal(t,
			fmt.Sprintf("https://%s/datakit-%s-%s-%s.tar.gz",
				arg.DataKitBaseURL, runtime.GOOS, runtime.GOARCH, arg.DataKitVersion),
			arg.DistDatakitURL)

		assert.Equal(t, "https://static.guance.com/dd-image/dd-java-agent.jar", arg.DistDatakitAPMInjJavaLibURL)
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
