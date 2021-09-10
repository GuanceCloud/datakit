package diskio

import (
	"sort"
	"testing"
	"time"

	"github.com/shirou/gopsutil/disk"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
)

type IOCountersStat = disk.IOCountersStat

var (
	testData = map[string]IOCountersStat{
		"loop0": {
			ReadCount:        52,
			MergedReadCount:  0,
			WriteCount:       0,
			MergedWriteCount: 0,
			ReadBytes:        1110016,
			WriteBytes:       0,
			ReadTime:         36,
			WriteTime:        0,
			IopsInProgress:   0,
			IoTime:           40,
			WeightedIO:       36,
			Name:             "loop0",
			SerialNumber:     "",
			Label:            "",
		},
		"loop1": {
			ReadCount:        43,
			MergedReadCount:  0,
			WriteCount:       0,
			MergedWriteCount: 0,
			ReadBytes:        355328,
			WriteBytes:       0,
			ReadTime:         8,
			WriteTime:        0,
			IopsInProgress:   0,
			IoTime:           20,
			WeightedIO:       8,
			Name:             "loop1",
			SerialNumber:     "",
			Label:            "",
		},
		"sda": {
			ReadCount:        419,
			MergedReadCount:  0,
			WriteCount:       0,
			MergedWriteCount: 0,
			ReadBytes:        7208448,
			WriteBytes:       0,
			ReadTime:         465,
			WriteTime:        0,
			IopsInProgress:   0,
			IoTime:           608,
			WeightedIO:       465,
			Name:             "sda",
			SerialNumber:     "",
			Label:            "",
		},
		"sda1": {
			ReadCount:        53,
			MergedReadCount:  0,
			WriteCount:       0,
			MergedWriteCount: 0,
			ReadBytes:        626688,
			WriteBytes:       0,
			ReadTime:         20,
			WriteTime:        0,
			IopsInProgress:   0,
			IoTime:           36,
			WeightedIO:       20,
			Name:             "sda1",
			SerialNumber:     "",
			Label:            "",
		},
		"sda2": {
			ReadCount:        56,
			MergedReadCount:  0,
			WriteCount:       0,
			MergedWriteCount: 0,
			ReadBytes:        2173952,
			WriteBytes:       0,
			ReadTime:         26,
			WriteTime:        0,
			IopsInProgress:   0,
			IoTime:           28,
			WeightedIO:       26,
			Name:             "sda2",
			SerialNumber:     "",
			Label:            "",
		},
		"sdb": {
			ReadCount:        1000749,
			MergedReadCount:  392242,
			WriteCount:       552776,
			MergedWriteCount: 1217830,
			ReadBytes:        26472674816,
			WriteBytes:       39289734144,
			ReadTime:         409419,
			WriteTime:        488838,
			IopsInProgress:   0,
			IoTime:           643584,
			WeightedIO:       904930,
			Name:             "sdb",
			SerialNumber:     "",
			Label:            "",
		},
		"sdb1": {
			ReadCount:        198,
			MergedReadCount:  29,
			WriteCount:       2,
			MergedWriteCount: 0,
			ReadBytes:        6110208,
			WriteBytes:       1024,
			ReadTime:         202,
			WriteTime:        3,
			IopsInProgress:   0,
			IoTime:           184,
			WeightedIO:       206,
			Name:             "sdb1",
			SerialNumber:     "",
			Label:            "",
		},
		"sdb2": {
			ReadCount:        58,
			MergedReadCount:  0,
			WriteCount:       0,
			MergedWriteCount: 0,
			ReadBytes:        2138112,
			WriteBytes:       0,
			ReadTime:         31,
			WriteTime:        0,
			IopsInProgress:   0,
			IoTime:           36,
			WeightedIO:       31,
			Name:             "sdb2",
			SerialNumber:     "",
			Label:            "",
		},
		"sdb3": {
			ReadCount:        503575,
			MergedReadCount:  253655,
			WriteCount:       437354,
			MergedWriteCount: 1079685,
			ReadBytes:        11279713280,
			WriteBytes:       35390902272,
			ReadTime:         191469,
			WriteTime:        441453,
			IopsInProgress:   0,
			IoTime:           500292,
			WeightedIO:       632923,
			Name:             "sdb3",
			SerialNumber:     "",
			Label:            "",
		},
	}
)

func DiskIO4Test(names ...string) (map[string]disk.IOCountersStat, error) {
	return testData, nil
}

func ResultMeasurementFilterDevices(i *Input) ([]*diskioMeasurement, error) {
	m := []*diskioMeasurement{}
	filter := DevicesFilter{}
	err := filter.Compile(i.Devices)
	if err != nil {
		return m, err
	}
	for name, io := range testData {
		if len(filter.filters) < 1 || filter.Match(name) {
			tmp := &diskioMeasurement{
				fields: map[string]interface{}{
					"reads":            io.ReadCount,
					"writes":           io.WriteCount,
					"read_bytes":       io.ReadBytes,
					"write_bytes":      io.WriteBytes,
					"read_time":        io.ReadTime,
					"write_time":       io.WriteTime,
					"io_time":          io.IoTime,
					"weighted_io_time": io.WeightedIO,
					"iops_in_progress": io.IopsInProgress,
					"merged_reads":     io.MergedReadCount,
					"merged_writes":    io.MergedWriteCount,
				},
				tags: map[string]string{"name": io.Name},
				name: metricName,
			}
			if !i.SkipSerialNumber {
				if len(io.SerialNumber) != 0 {
					tmp.tags["serial"] = io.SerialNumber
				} else {
					tmp.tags["serial"] = "unknown"
				}
			}
			m = append(m, tmp)
		}
	}
	// Sort by tag
	tagname := []string{}
	tagindex := map[string]int{}
	for index, x := range m {
		tagindex[x.tags["name"]] = index
		tagname = append(tagname, x.tags["name"])
	}
	result := []*diskioMeasurement{}
	sort.Strings(tagname)
	for _, x := range tagname {
		result = append(result, m[tagindex[x]])
	}
	return result, err
}
func TestF(t *testing.T) {
	i := &Input{diskIO: disk.IOCounters}
	i.Devices = []string{"^sdb[\\d]{0,2}"}
	err := i.Collect()
	if err != nil {
		t.Error(err)
	}
	time.Sleep(time.Second * 1)
	err = i.Collect()
	if err != nil {
		t.Error(err)
	}
	// clear collectCache
	i.collectCache = make([]inputs.Measurement, 0)

	i.diskIO = DiskIO4Test
	err = i.Collect()
	if err != nil {
		t.Error(err)
	}
	result, err := ResultMeasurementFilterDevices(i)
	if err != nil {
		t.Error(err)
	}
	// Sort by tag
	tagname := []string{}
	tagindex := map[string]int{}
	for index, x := range i.collectCache {
		xM, _ := x.(*diskioMeasurement)
		tagindex[xM.tags["name"]] = index
		tagname = append(tagname, xM.tags["name"])
	}
	collectCache := []inputs.Measurement{}
	sort.Strings(tagname)

	for _, x := range tagname {
		collectCache = append(collectCache, i.collectCache[tagindex[x]])
	}
	AssertMeasurement(t, result, collectCache, FieldCompare+TagCompare+NameCompare)
	// TODO TestDiskInfo
}
