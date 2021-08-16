// +build linux

package conntrack

import (
	"fmt"
	"io/ioutil"
	"strconv"
	"strings"

	"github.com/prometheus/procfs"
)

type conntrackStatistics struct {
	found         uint64
	invalid       uint64
	ignore        uint64
	insert        uint64
	insertFailed  uint64
	drop          uint64
	earlyDrop     uint64
	searchRestart uint64
}

func readIntFromFile(path string) (int64, error) {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return 0, err
	}
	value, err := strconv.ParseInt(strings.TrimSpace(string(data)), 10, 64)
	if err != nil {
		return 0, err
	}
	return value, nil
}

func ConntrackCollect() *ConntrackInfo {
	info := &ConntrackInfo{}

	value, err := readIntFromFile("/proc/sys/net/netfilter/nf_conntrack_count")
	if err != nil {
		// l.Warn(err)
		info.Current = -1 // set default -1
	} else {
		info.Current = value
	}

	value, err = readIntFromFile("/proc/sys/net/netfilter/nf_conntrack_max")
	if err != nil {
		// l.Warn(err)
		info.Limit = -1
	} else {
		info.Limit = value
	}

	conntrackStats, err := getConntrackStatistics()
	if err != nil {
		// l.Warn(err)
	} else {
		info.Found = int64(conntrackStats.found)
		info.Invalid = int64(conntrackStats.invalid)
		info.Ignore = int64(conntrackStats.ignore)
		info.Insert = int64(conntrackStats.insert)
		info.InsertFailed = int64(conntrackStats.insertFailed)
		info.Drop = int64(conntrackStats.drop)
		info.EarlyDrop = int64(conntrackStats.earlyDrop)
		info.SearchRestart = int64(conntrackStats.searchRestart)
	}

	return info
}

func getConntrackStatistics() (*conntrackStatistics, error) {
	c := conntrackStatistics{}

	fs, err := procfs.NewFS("/proc")
	if err != nil {
		return nil, fmt.Errorf("failed to open procfs: %w", err)
	}

	connStats, err := fs.ConntrackStat()
	if err != nil {
		return nil, err
	}

	for _, connStat := range connStats {
		c.found += connStat.Found
		c.invalid += connStat.Invalid
		c.ignore += connStat.Ignore
		c.insert += connStat.Insert
		c.insertFailed += connStat.InsertFailed
		c.drop += connStat.Drop
		c.earlyDrop += connStat.EarlyDrop
		c.searchRestart += connStat.SearchRestart
	}

	return &c, nil
}
