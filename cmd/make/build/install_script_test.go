// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2026-present Guance, Inc.

package build

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"text/template"

	"github.com/stretchr/testify/require"
)

func TestInstallScriptArchitecture(t *testing.T) {
	bash, err := exec.LookPath("bash")
	if err != nil {
		t.Skip("bash is required to test the Unix install script")
	}

	source, err := os.ReadFile(filepath.Join("..", "..", "..", "templates", "install.template.sh"))
	require.NoError(t, err)
	tmpl, err := template.New("install").Parse(string(source))
	require.NoError(t, err)
	var rendered bytes.Buffer
	require.NoError(t, tmpl.Execute(&rendered, map[string]string{
		"InstallBaseURL": "example.invalid/datakit-v2",
		"Version":        "2.12.0",
		"BrandDomain":    "example.invalid",
	}))

	// Exercise the rendered script through URL selection, before any temporary
	// files, downloads, privilege escalation or installer execution.
	prefix, _, found := strings.Cut(rendered.String(), "\ntmpdir=")
	require.True(t, found, "install script must have a temporary-directory boundary")

	const mocks = `
OSTYPE="$1"
machine="$2"
translated="$3"
uname() {
  case "$1" in
    -m) printf '%s\n' "$machine" ;;
    -r) printf '%s\n' '6.8.0' ;;
    *) return 1 ;;
  esac
}
sw_vers() { printf '%s\n' '14.0'; }
sysctl() {
  [[ "$*" == '-n sysctl.proc_translated' ]] || return 1
  [[ "$translated" != 'missing' ]] || return 1
  printf '%s\n' "$translated"
}
`
	cases := []struct {
		name, ostype, machine, translated, want string
	}{
		{"apple-silicon", "darwin23", "arm64", "0", "darwin-arm64"},
		{"rosetta", "darwin23", "x86_64", "1", "darwin-arm64"},
		{"intel", "darwin23", "x86_64", "0", "darwin-amd64"},
		{"intel-without-rosetta-sysctl", "darwin23", "x86_64", "missing", "darwin-amd64"},
		{"linux-arm64", "linux-gnu", "aarch64", "1", "linux-arm64"},
		{"linux-amd64", "linux-gnu", "x86_64", "1", "linux-amd64"},
		{"linux-arm", "linux-gnu", "armv7l", "missing", "linux-arm"},
		{"linux-386", "linux-gnu", "i686", "missing", "linux-386"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			script := mocks + prefix + "\nprintf 'Selected URL: %s\\n' \"$installer_url\"\n"
			cmd := exec.Command(bash, "--noprofile", "--norc", "-c", script,
				"install-test", tc.ostype, tc.machine, tc.translated)
			// Do not inherit installation options, proxies or Bash startup hooks.
			cmd.Env = []string{"PATH=" + os.Getenv("PATH"), "LC_ALL=C"}
			cmd.Dir = t.TempDir()
			output, err := cmd.CombinedOutput()
			require.NoError(t, err, "%s", output)
			require.Contains(t, string(output),
				"Selected URL: https://example.invalid/datakit-v2/installer-"+tc.want+"-2.12.0\n")
			files, err := os.ReadDir(cmd.Dir)
			require.NoError(t, err)
			require.Empty(t, files, "architecture detection must not create installation files")
		})
	}
}

func TestDarwinReleaseArchitectures(t *testing.T) {
	t.Setenv("ALL_ARCHS", "")
	archs := ParseArchs(ALL)
	require.Contains(t, archs, "darwin/amd64")
	require.Contains(t, archs, "darwin/arm64")
}

func TestDarwinARM64NotifyContent(t *testing.T) {
	content := buildNotifyContent("2.12.0", "example.invalid/datakit-v2", ReleaseTesting,
		[]string{"darwin/amd64", "darwin/arm64"})
	require.Contains(t, content, "### darwin/amd64 Install/Upgrade")
	require.Contains(t, content, "### darwin/arm64 Install/Upgrade")
	require.Contains(t, content, "install-2.12.0.sh")
}
