// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.
// Some code modified from project Datadog (https://www.datadoghq.com/).

// Package snmprefiles contains snmp pre prepared files, including default profiles, traps database etc.
package snmprefiles

import (
	"embed"
	"fmt"
	"io/fs"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/GuanceCloud/cliutils/logger"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs/snmp/snmpmeasurement"
)

const (
	profilesDir = "profiles"
	trapsDBDir  = "traps_db"
)

var l = logger.DefaultSLogger("snmprefiles")

func ReleaseFiles() error {
	l = logger.SLogger("snmprefiles")
	snmpPreFilesRoot := filepath.Join(datakit.ConfdDir, snmpmeasurement.InputName)
	if err := releaseProfiles(snmpPreFilesRoot); err != nil {
		return err
	}
	if err := releaseTrapsDB(snmpPreFilesRoot); err != nil {
		return err
	}
	return nil
}

func GetProfilesRoot() string {
	return filepath.Join(datakit.ConfdDir, snmpmeasurement.InputName, profilesDir)
}

func GetUserProfilesRoot() string {
	return filepath.Join(datakit.ConfdDir, snmpmeasurement.InputName, "user"+profilesDir)
}

func GetTrapsDBRoot() string {
	return filepath.Join(datakit.ConfdDir, snmpmeasurement.InputName, trapsDBDir)
}

//------------------------------------------------------------------------------

//go:embed profiles/*.yaml
var profileFiles embed.FS

func releaseProfiles(dest string) error {
	releaseDir := filepath.Join(dest, profilesDir)
	if err := os.MkdirAll(releaseDir, datakit.ConfPerm); err != nil {
		return err
	}
	if err := releaseEmbedFiles(releaseDir, profilesDir, &profileFiles); err != nil {
		return err
	}
	return nil
}

//go:embed traps_db/*.gz
var trapsdbFiles embed.FS

func releaseTrapsDB(dest string) error {
	releaseDir := filepath.Join(dest, trapsDBDir)
	if err := os.MkdirAll(releaseDir, datakit.ConfPerm); err != nil {
		return err
	}
	if err := releaseEmbedFiles(releaseDir, trapsDBDir, &trapsdbFiles); err != nil {
		return err
	}
	return nil
}

// example:
//
//	dest = conf.d, embDir = profiles
//	dest = conf.d, embDir = traps_db
func releaseEmbedFiles(dest, embDir string, emdFS *embed.FS) error {
	files, err := fs.ReadDir(emdFS, embDir)
	if err != nil {
		return fmt.Errorf("ReadDir(%q): %w", embDir, err)
	}

	for _, file := range files {
		var (
			releasePath = filepath.Join(dest, file.Name())
			embPath     = filepath.Join(embDir, file.Name())
		)

		l.Debugf("read %q to %q...", embPath, releasePath)

		if data, err := fs.ReadFile(emdFS, embPath); err != nil {
			l.Warnf("ReadFile(%q): %s, ignored", embPath, err)
			continue
		} else if err := ioutil.WriteFile(releasePath, data, datakit.ConfPerm); err != nil {
			l.Warnf("ioutil.WriteFile(%q): %s", releasePath, err)
			continue
		}
	}

	return nil
}
