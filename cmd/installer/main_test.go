// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package main

import (
	"net/http/httptest"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	T "testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/cmd/upgrader/upgrader"
	apminjUtils "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/apminject/dkrunc/utils"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/datakit"
)

func Test_trimFileName(t *T.T) {
	t.Run(`basic`, func(t *T.T) {
		src := " ./abc.1,  ./def.1   "
		arr := strings.Split(src, ",")

		assert.Equal(t, `abc.1`, trimFileName(arr[0], "./"))
		assert.Equal(t, `def.1`, trimFileName(arr[1], "./"))

		src = ` .\abc.1,  .\def.1   `
		arr = strings.Split(src, ",")

		assert.Equal(t, `abc.1`, trimFileName(arr[0], `.\`))
		assert.Equal(t, `def.1`, trimFileName(arr[1], `.\`))

		// absolute paths must be left untouched: the old implementation trimmed
		// a *cutset*, so it also ate the leading "/" and the offline install
		// could not open its packages any more
		assert.Equal(t, `/tmp/x/.dk_upgrader-linux-amd64.tar.gz`,
			trimFileName(`/tmp/x/.dk_upgrader-linux-amd64.tar.gz`, "./"))
		assert.Equal(t, `C:\Users\Li\AppData\Local\Temp\Temp_dk_installer_files_x\.dk_upgrader-windows-amd64-1.93.0.tar.gz`,
			trimFileName(`C:\Users\Li\AppData\Local\Temp\Temp_dk_installer_files_x\.dk_upgrader-windows-amd64-1.93.0.tar.gz`, `.\`))
	})
}

func TestCheckIsVersion(t *T.T) {
	r := gin.New()
	r.GET("/v1/ping", func(c *gin.Context) {
		c.Data(200, "", []byte(`{ "content":{ "version": "1.2.3", "uptime": "30m", "host": "wtf" }}`))
	})

	_ = r

	ts := httptest.NewServer(r)
	time.Sleep(time.Second)
	defer ts.Close()

	cases := []struct {
		ver  string
		fail bool
	}{
		{
			ver: "1.2.3",
		},
		{
			ver:  "1.2.4",
			fail: true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.ver, func(t *T.T) {
			err := checkIsNewVersion(ts.URL, tc.ver)
			if tc.fail {
				assert.Error(t, err)
				t.Logf("expect err: %s", err)

				return
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestCopyFileAtomic(t *T.T) {
	dir := t.TempDir()
	src := filepath.Join(dir, "dd-java-agent.jar")
	dst := filepath.Join(dir, "apm_inject", "lib", "java", "dd-java-agent.jar")

	assert.NoError(t, os.MkdirAll(filepath.Dir(dst), 0o755))
	assert.NoError(t, os.WriteFile(dst, []byte("old"), 0o755))
	assert.NoError(t, os.WriteFile(src, []byte("new"), 0o644))

	assert.NoError(t, copyFileAtomic(src, dst, 0o755))

	data, err := os.ReadFile(dst)
	assert.NoError(t, err)
	assert.Equal(t, []byte("new"), data)

	info, err := os.Stat(dst)
	assert.NoError(t, err)
	assert.Equal(t, os.FileMode(0o755), info.Mode().Perm())
}

func TestInstallTemplateDoesNotInstallCompletion(t *T.T) {
	data, err := os.ReadFile(filepath.Join("..", "..", "templates", "install.template.sh"))
	assert.NoError(t, err)

	script := string(data)
	assert.NotContains(t, script, "datakit tool --setup-completer-script")
	assert.NotContains(t, script, "datakit tool --completer-script")
	assert.NotContains(t, script, "datakit completion")
}

func TestEnsureDatakitCLIExecutable(t *T.T) {
	if runtime.GOOS == datakit.OSWindows {
		t.Skip("Windows does not use unix executable bits")
	}

	origInstallDir := datakit.InstallDir
	t.Cleanup(func() {
		datakit.SetupWorkDir(origInstallDir)
	})

	dir := t.TempDir()
	datakit.SetupWorkDir(dir)

	binaryPath := datakit.DatakitBinaryPath()
	require.NoError(t, os.WriteFile(binaryPath, []byte("datakit"), 0o750))
	require.NoError(t, os.Chmod(dir, 0o750))

	require.NoError(t, ensureDatakitCLIExecutable())

	dirInfo, err := os.Stat(dir)
	require.NoError(t, err)
	assert.Equal(t, os.FileMode(0o755), dirInfo.Mode().Perm())

	binInfo, err := os.Stat(binaryPath)
	require.NoError(t, err)
	assert.Equal(t, os.FileMode(0o755), binInfo.Mode().Perm())
}

// TestOfflineExtractDestDir: the offline one-liner hands the packages over as
// absolute paths from a temp dir, with a leading dot in the file name
// (C:\Users\Li\AppData\Local\Temp\Temp_dk_installer_files_xxx\.dk_upgrader-windows-amd64-1.93.0.tar.gz).
// The upgrader package must still land in the upgrader install dir, otherwise
// its service cannot be installed ("The system cannot find the file specified").
func TestOfflineExtractDestDir(t *T.T) {
	upgraderDir := upgrader.InstallDir
	datakitDir := datakit.InstallDir
	injectDir := filepath.Join(datakit.InstallDir,
		apminjUtils.DirInject, apminjUtils.DirInjectSubInject)

	cases := map[string]string{
		// the customer's offline install command
		`C:\Users\Li\AppData\Local\Temp\Temp_dk_installer_files_20260918013158_8742\.dk_upgrader-windows-amd64-1.93.0.tar.gz`: upgraderDir,
		`C:\Users\Li\AppData\Local\Temp\Temp_dk_installer_files_20260918013158_8742\.datakit-windows-amd64-1.93.0.tar.gz`:     datakitDir,
		`C:\Users\Li\AppData\Local\Temp\Temp_dk_installer_files_20260918013158_8742\.data.tar.gz`:                             datakitDir,

		// relative forms used by the docs
		`./dk_upgrader-linux-amd64.tar.gz`:              upgraderDir,
		`.\dk_upgrader-windows-amd64.tar.gz`:            upgraderDir,
		`./.dk_upgrader-linux-amd64.tar.gz`:             upgraderDir,
		`dk_upgrader-linux-amd64.tar.gz`:                upgraderDir,
		`./datakit-linux-amd64.tar.gz`:                  datakitDir,
		`./.data.tar.gz`:                                datakitDir,
		`./datakit-apm-inject-linux-amd64.tar.gz`:       injectDir,
		`/tmp/x/.datakit-apm-inject-linux-amd64.tar.gz`: injectDir,
	}

	for in, want := range cases {
		assert.Equalf(t, want, offlineExtractDestDir(in), "offlineExtractDestDir(%q)", in)
	}
}
