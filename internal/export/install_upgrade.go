// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package export

import (
	"fmt"
	"sort"
	"strings"
)

var (
	windowsInstallCmdTemplate = `Remove-Item -ErrorAction SilentlyContinue Env:DK_*;
%s
Set-ExecutionPolicy Bypass -scope Process -Force;
Import-Module bitstransfer;
start-bitstransfer %s -source %s/install%s.ps1 -destination .install.ps1;
powershell ./.install.ps1;`

	windowsUpgradeCmdTemplate = `Remove-Item -ErrorAction SilentlyContinue Env:DK_*;
%s
Set-ExecutionPolicy Bypass -scope Process -Force;
Import-Module bitstransfer;
start-bitstransfer %s -source %s/install%s.ps1 -destination .install.ps1;
powershell ./.install.ps1;`

	unixInstallCmdTemplate = `%s %s -c "$(curl -L %s/install%s.sh)"`

	unixUpgradeCmdTemplate = `%s %s -c "$(curl -L %s/install%s.sh)"`
)

type installCmd struct {
	upgrade,
	inJSON,
	lite,
	oneline bool
	indent int
	temp,
	platform,
	shell,
	version,
	dwURL,
	bitstransferOpts,
	sourceURL string
	envs map[string]string
}

// String get install command string format.
func (x *installCmd) String() (out string) {
	sourceURL := strings.TrimRight(x.sourceURL, "/")

	if len(x.shell) == 0 {
		x.shell = "bash"
	}

	if x.upgrade {
		x.envs["DK_UPGRADE"] = "1"
		if x.lite {
			x.envs["DK_LITE"] = "1"
		}

		switch x.platform {
		case "windows":
			x.temp = windowsUpgradeCmdTemplate
			out = fmt.Sprintf(x.temp, x.envsStr(), x.bitstransferOpts, sourceURL, x.version)
		case "unix":
			x.temp = unixUpgradeCmdTemplate
			out = fmt.Sprintf(x.temp, x.envsStr(), x.shell, sourceURL, x.version)
		}
	} else {
		if _, ok := x.envs["DK_DATAWAY"]; !ok {
			x.envs["DK_DATAWAY"] = fmt.Sprintf("%s?token=<TOKEN>", x.dwURL)
		}

		switch x.platform {
		case "windows":
			x.temp = windowsInstallCmdTemplate
			out = fmt.Sprintf(x.temp, x.envsStr(), x.bitstransferOpts, sourceURL, x.version)
		case "unix":
			x.temp = unixInstallCmdTemplate
			out = fmt.Sprintf(x.temp, x.envsStr(), x.shell, sourceURL, x.version)
		}
	}

	if x.oneline { // convert multiline command to single line
		out = strings.Join(strings.Split(out, "\n"), " ")
	}

	// we should escape " with \" when in JSON string
	if x.inJSON {
		out = strings.ReplaceAll(out, `"`, `\"`)
	}

	if x.indent > 0 {
		out = codeBlock(out, x.indent)
	}

	return out
}

func (x *installCmd) envsStr() (out string) {
	var arr []string

	// make keys sorted to keep same output among String() callings.
	var keys []string
	for k := range x.envs {
		keys = append(keys, k)
	}

	sort.Strings(keys)

	switch x.platform {
	case "windows":
		for _, k := range keys {
			arr = append(arr, fmt.Sprintf(`$env:%s="%s";`, k, x.envs[k]))
		}
		out = strings.Join(arr, "\n")
	case "unix":
		for _, k := range keys {
			arr = append(arr, fmt.Sprintf(`%s=%s`, k, x.envs[k]))
		}
		out = strings.Join(arr, " ")
	}

	return
}

// InstallOpt used to select different datakit install command.
type InstallOpt func(x *installCmd)

// WithEnvs set optional install command with environments k/v pair.
func (p *Params) WithEnvs(k, v string) InstallOpt {
	return func(x *installCmd) {
		x.envs[k] = v
	}
}

// WithJSON used to output the command in JSON and reserved characters escaped.
func (p *Params) WithJSON(on bool) InstallOpt {
	return func(x *installCmd) {
		x.inJSON = on
	}
}

// WithOneline output datakit install command in single line.
func (p *Params) WithOneline(on bool) InstallOpt {
	return func(x *installCmd) {
		x.oneline = on
	}
}

// WithBitstransferOpts used to control bitstransfer options for windows powershell.
func (p *Params) WithBitstransferOpts(opts string) InstallOpt {
	return func(x *installCmd) {
		x.bitstransferOpts = opts
	}
}

// WithVersion used to select install version.
func (p *Params) WithVersion(ver string) InstallOpt {
	return func(x *installCmd) {
		x.version = ver
	}
}

// WithDatawayURL used to select dataway host.
func (p *Params) WithDatawayURL(dwurl string) InstallOpt {
	return func(x *installCmd) {
		x.dwURL = dwurl
	}
}

// WithPlatform used to select datakit install platform(unix or windows).
func (p *Params) WithPlatform(platform string) InstallOpt {
	return func(x *installCmd) {
		x.platform = platform
	}
}

// WithShell used to set shell for unix.
func (p *Params) WithShell(shell string) InstallOpt {
	return func(x *installCmd) {
		x.shell = shell
	}
}

// WithUpgrade used to set upgrade flag.
func (p *Params) WithUpgrade(on bool) InstallOpt {
	return func(x *installCmd) {
		x.upgrade = on
	}
}

// WithIndent set indent to every command lines.
func (p *Params) WithIndent(n int) InstallOpt {
	return func(x *installCmd) {
		x.indent = n
	}
}

// WithSourceURL used to select install source URL.
func (p *Params) WithSourceURL(url string) InstallOpt {
	return func(x *installCmd) {
		url = strings.TrimSuffix(url, "/")
		x.sourceURL = url
	}
}

// DefaultInstallCmd used to get a default install command with default dataway
// and source URL set.
func DefaultInstallCmd() *installCmd {
	return &installCmd{
		dwURL:     "https://openway.guance.com",
		sourceURL: "https://static.guance.com/datakit",
		envs:      map[string]string{},
		shell:     "bash",
	}
}

// InstallCommand used to get datakit install command with optional settings.
func InstallCommand(opts ...InstallOpt) *installCmd {
	x := DefaultInstallCmd()

	for _, opt := range opts {
		if opt != nil {
			opt(x)
		}
	}

	return x
}
