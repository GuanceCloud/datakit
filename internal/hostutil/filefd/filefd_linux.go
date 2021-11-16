// +build linux

package filefd

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
)

func parseFileFDStats(filename string) (map[string]string, error) {
	file, err := os.Open(filepath.Clean(filename))
	if err != nil {
		return nil, err
	}
	defer file.Close() //nolint:errcheck,gosec

	content, err := ioutil.ReadAll(file)
	if err != nil {
		return nil, err
	}
	parts := bytes.Split(bytes.TrimSpace(content), []byte("\u0009"))
	if len(parts) < 3 { //nolint:gomnd
		return nil, fmt.Errorf("unexpected number of file stats in %q", filename)
	}

	fileFDStat := map[string]string{}
	// The file-nr proc is only 1 line with 3 values.
	fileFDStat["allocated"] = string(parts[0])
	// The second value is skipped as it will always be zero in linux 2.6.
	fileFDStat["maximum"] = string(parts[2])

	return fileFDStat, nil
}

func collect() (map[string]int64, error) {
	info := make(map[string]int64)
	fileFDStat, err := parseFileFDStats("/proc/sys/fs/file-nr")
	if err != nil {
		return nil, fmt.Errorf("couldn't get file-nr: %w", err)
	}
	for name, value := range fileFDStat {
		v, err := strconv.ParseInt(value, 10, 64) //nolint:gomnd
		if err != nil {
			return nil, fmt.Errorf("invalid value %s in file-nr: %w", value, err)
		}
		info[name] = v
	}
	return info, nil
}
