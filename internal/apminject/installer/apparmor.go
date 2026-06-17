// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

//go:build (linux && amd64) || (linux && arm64)
// +build linux,amd64 linux,arm64

package installer

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/GuanceCloud/cliutils/logger"
)

const (
	appArmorAbstractionName = "datakit-apm-inject"
)

var (
	appArmorEnabledPath     = "/sys/module/apparmor/parameters/enabled"
	appArmorProfilesPath    = "/sys/kernel/security/apparmor/profiles"
	appArmorAbstractionPath = "/etc/apparmor.d/abstractions/" + appArmorAbstractionName
)

func appArmorEnabled() bool {
	if data, err := os.ReadFile(appArmorEnabledPath); err == nil {
		return strings.EqualFold(strings.TrimSpace(string(data)), "Y")
	}

	if _, err := os.Stat(appArmorProfilesPath); err == nil {
		return true
	}

	return false
}

func appArmorPath(path string) string {
	return filepath.ToSlash(filepath.Clean(path))
}

func appArmorAbstractionContent(installDir string) string {
	installDir = appArmorPath(installDir)

	injectDir := appArmorPath(filepath.Join(installDir, DirInject))
	injectBinDir := appArmorPath(filepath.Join(injectDir, DirInjectSubInject))
	injectLibDir := appArmorPath(filepath.Join(injectDir, DirInjectSubLib))
	injectLogDir := appArmorPath(filepath.Join(injectDir, "log"))

	return fmt.Sprintf(`# DataKit APM auto-injection AppArmor abstraction.
#
# Include this file in every enforced profile that starts applications
# requiring DataKit APM auto-injection, then reload that profile:
#
#   #include <abstractions/%s>
#
# Example:
#   echo '#include <abstractions/%s>' >> /etc/apparmor.d/local/<profile-file>
#   apparmor_parser -r /etc/apparmor.d/<profile-file>

/etc/ld.so.preload r,
/dev/urandom r,
/tmp/dk_inject_rewrite_* rw,

%s/apm_launcher.so mr,
%s/apm_launcher_musl.so mr,
%s/rewriter ix,

%s/** r,
%s/** mr,

%s/ rw,
%s/rewriter.log rwk,

%s/conf.d/datakit.conf r,
%s/conf.d/statsd/statsd.conf r,
/var/run/datakit/datakit.sock rw,
/var/run/datakit/statsd.sock rw,
`, appArmorAbstractionName, appArmorAbstractionName,
		injectBinDir, injectBinDir, injectBinDir,
		injectLibDir, injectLibDir,
		injectLogDir, injectLogDir,
		installDir, installDir)
}

func prepareAppArmorAbstraction(installDir string) (bool, error) {
	if !appArmorEnabled() {
		return false, nil
	}

	content := []byte(appArmorAbstractionContent(installDir))
	if old, err := os.ReadFile(appArmorAbstractionPath); err == nil && string(old) == string(content) {
		return true, nil
	}

	if err := os.MkdirAll(filepath.Dir(appArmorAbstractionPath), 0o755); err != nil {
		return true, fmt.Errorf("create AppArmor abstraction dir: %w", err)
	}

	if err := os.WriteFile(appArmorAbstractionPath, content, 0o644); err != nil {
		return true, fmt.Errorf("write AppArmor abstraction %s: %w", appArmorAbstractionPath, err)
	}

	return true, nil
}

func prepareAndWarnAppArmor(log *logger.Logger, installDir string) {
	enabled, err := prepareAppArmorAbstraction(installDir)
	if err != nil {
		if log != nil {
			log.Warnf("prepare AppArmor abstraction for APM auto-inject failed: %s", err.Error())
		}
		return
	}

	if enabled && log != nil {
		log.Warnf("AppArmor is enabled; if target applications run under enforced profiles, include <abstractions/%s> in those profiles and reload them, otherwise APM auto-inject may be denied", appArmorAbstractionName)
	}
}
