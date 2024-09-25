package storage

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/GuanceCloud/cliutils/pprofparser/cfg"
	"github.com/GuanceCloud/cliutils/pprofparser/domain/parameter"
)

var _ Storage = &Disk{}

type Disk struct {
}

func (d *Disk) GetProfilePathOld(workspaceUUID string, profileID string, unixTimeNS int64, ossFilename string) string {
	return filepath.Join(d.GetProfileDirOld(workspaceUUID, profileID, unixTimeNS), ossFilename)
}

func (d *Disk) GetProfileDirOld(workspaceUUID string, profileID string, unixTimeNS int64) string {
	if unixTimeNS >= parameter.MinTimestampMicro && unixTimeNS <= parameter.MaxTimestampMicro {
		unixTimeNS *= 1000
	}
	date := time.Unix(0, unixTimeNS).In(timeZoneCST).Format("20060102")
	return filepath.Join(cfg.Cfg.Storage.Disk.ProfileDir, date, workspaceUUID, profileID)
}

func (d *Disk) GetProfilePath(workspaceUUID string, profileID string, unixTimeNS int64, ossFilename string) string {
	return d.GetProfileDir(workspaceUUID, profileID, unixTimeNS) + "/" + ossFilename
}

func (d *Disk) GetProfileDir(workspaceUUID string, profileID string, unixTimeNS int64) string {
	if unixTimeNS >= parameter.MinTimestampMicro && unixTimeNS <= parameter.MaxTimestampMicro {
		unixTimeNS *= 1000
	}
	date := time.Unix(0, unixTimeNS).In(timeZoneCST).Format("20060102")
	return filepath.Join(cfg.Cfg.Storage.Disk.ProfileDir, date, workspaceUUID, profileID[:2], profileID)
}

func (d *Disk) IsFileExists(path string) (bool, error) {
	stat, err := os.Stat(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return false, nil
		}
		return false, fmt.Errorf("stat file [%s] err: %w", path, err)
	}
	return stat.Mode().IsRegular(), nil
}

func (d *Disk) ReadFile(path string) (io.ReadCloser, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("open file [%s] fail: %w", path, err)
	}
	return f, nil
}
