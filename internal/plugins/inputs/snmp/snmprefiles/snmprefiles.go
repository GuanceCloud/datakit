// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.
// Some code modified from project Datadog (https://www.datadoghq.com/).

// Package snmprefiles contains snmp pre prepared files, including default profiles, traps database etc.
package snmprefiles

import (
	"embed"
	"io/fs"
	"io/ioutil"
	"os"
	"path/filepath"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs/snmp/snmpmeasurement"
)

const (
	profilesDir = "profiles"
	trapsDBDir  = "traps_db"
)

func ReleaseFiles() error {
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

func GetTrapsDBRoot() string {
	return filepath.Join(datakit.ConfdDir, snmpmeasurement.InputName, trapsDBDir)
}

//------------------------------------------------------------------------------

//go:embed profiles/*.yaml
var profileFiles embed.FS

func releaseProfiles(targetRoot string) error {
	releaseDir := filepath.Join(targetRoot, profilesDir)
	if err := os.MkdirAll(releaseDir, datakit.ConfPerm); err != nil {
		return err
	}
	if err := releaseEMDFiles(releaseDir, profilesDir, &profileFiles); err != nil {
		return err
	}
	return nil
}

//------------------------------------------------------------------------------

//go:embed traps_db/*.gz
var trapsdbFiles embed.FS

func releaseTrapsDB(targetRoot string) error {
	releaseDir := filepath.Join(targetRoot, trapsDBDir)
	if err := os.MkdirAll(releaseDir, datakit.ConfPerm); err != nil {
		return err
	}
	if err := releaseEMDFiles(releaseDir, trapsDBDir, &trapsdbFiles); err != nil {
		return err
	}
	return nil
}

//------------------------------------------------------------------------------

// example:
//
//	targetRoot = conf.d    emdDir = profiles
//	targetRoot = conf.d    emdDir = traps_db
func releaseEMDFiles(targetRoot, emdDir string, emdFS *embed.FS) error {
	files, _ := fs.ReadDir(emdFS, emdDir)
	for _, file := range files {
		releaseFullPath := filepath.Join(targetRoot, file.Name())
		insideFullName := filepath.Join(emdDir, file.Name())
		bys, err := fs.ReadFile(emdFS, insideFullName)
		if err != nil {
			return err
		}
		if err := ioutil.WriteFile(releaseFullPath, bys, datakit.ConfPerm); err != nil {
			return err
		}
	}

	return nil
}
