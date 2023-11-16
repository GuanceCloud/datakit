// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package diskio

import (
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/GuanceCloud/cliutils"
	"github.com/GuanceCloud/cliutils/point"
	"github.com/shirou/gopsutil/disk"
	"github.com/stretchr/testify/assert"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/datakit"
	dkio "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/testutils"
)

func TestInput_collect(t *testing.T) {
	type fields struct {
		Interval         time.Duration
		Devices          []string
		DeviceTags       []string
		NameTemplates    []string
		SkipSerialNumber bool
		Tags             map[string]string
		collectCache     []*point.Point
		lastStat         map[string]disk.IOCountersStat
		lastTime         time.Time
		diskIO           DiskIO
		feeder           dkio.Feeder
		mergedTags       map[string]string
		tagger           datakit.GlobalTagger
		infoCache        map[string]diskInfoCache
		deviceFilter     *DevicesFilter
		semStop          *cliutils.Sem
	}
	tests := []struct {
		name    string
		fields  fields
		want    []string
		wantErr bool
	}{
		{
			name: "ok with devices",
			fields: fields{
				diskIO:     DiskIO4Test,
				Devices:    []string{"^sda", "vd."},
				Tags:       map[string]string{},
				tagger:     testutils.DefaultMockTagger(),
				mergedTags: make(map[string]string),
			},
			want: []string{
				"diskio,host=HOST,name=/dev/sda,serial=unknown io_time=608i,iops_in_progress=0i,merged_reads=0i,merged_writes=0i,read_bytes=7208448i,read_time=465i,reads=419i,weighted_io_time=465i,write_bytes=0i,write_time=0i,writes=0i",
				"diskio,host=HOST,name=/dev/sda1,serial=unknown io_time=36i,iops_in_progress=0i,merged_reads=0i,merged_writes=0i,read_bytes=626688i,read_time=20i,reads=53i,weighted_io_time=20i,write_bytes=0i,write_time=0i,writes=0i",
				"diskio,host=HOST,name=/dev/sda2,serial=unknown io_time=28i,iops_in_progress=0i,merged_reads=0i,merged_writes=0i,read_bytes=2173952i,read_time=26i,reads=56i,weighted_io_time=26i,write_bytes=0i,write_time=0i,writes=0i",
			},
			wantErr: false,
		},
		{
			name: "ok no tag",
			fields: fields{
				diskIO:     DiskIO4Test,
				Tags:       map[string]string{},
				tagger:     testutils.DefaultMockTagger(),
				mergedTags: make(map[string]string),
			},
			want: []string{
				"diskio,host=HOST,name=/dev/loop0,serial=unknown io_time=40i,iops_in_progress=0i,merged_reads=0i,merged_writes=0i,read_bytes=1110016i,read_time=36i,reads=52i,weighted_io_time=36i,write_bytes=0i,write_time=0i,writes=0i",
				"diskio,host=HOST,name=/dev/loop1,serial=unknown io_time=20i,iops_in_progress=0i,merged_reads=0i,merged_writes=0i,read_bytes=355328i,read_time=8i,reads=43i,weighted_io_time=8i,write_bytes=0i,write_time=0i,writes=0i",
				"diskio,host=HOST,name=/dev/sda,serial=unknown io_time=608i,iops_in_progress=0i,merged_reads=0i,merged_writes=0i,read_bytes=7208448i,read_time=465i,reads=419i,weighted_io_time=465i,write_bytes=0i,write_time=0i,writes=0i",
				"diskio,host=HOST,name=/dev/sda1,serial=unknown io_time=36i,iops_in_progress=0i,merged_reads=0i,merged_writes=0i,read_bytes=626688i,read_time=20i,reads=53i,weighted_io_time=20i,write_bytes=0i,write_time=0i,writes=0i",
				"diskio,host=HOST,name=/dev/sda2,serial=unknown io_time=28i,iops_in_progress=0i,merged_reads=0i,merged_writes=0i,read_bytes=2173952i,read_time=26i,reads=56i,weighted_io_time=26i,write_bytes=0i,write_time=0i,writes=0i",
				"diskio,host=HOST,name=/dev/sdb,serial=unknown io_time=643584i,iops_in_progress=0i,merged_reads=392242i,merged_writes=1217830i,read_bytes=26472674816i,read_time=409419i,reads=1000749i,weighted_io_time=904930i,write_bytes=39289734144i,write_time=488838i,writes=552776i",
				"diskio,host=HOST,name=/dev/sdb1,serial=unknown io_time=184i,iops_in_progress=0i,merged_reads=29i,merged_writes=0i,read_bytes=6110208i,read_time=202i,reads=198i,weighted_io_time=206i,write_bytes=1024i,write_time=3i,writes=2i",
				"diskio,host=HOST,name=/dev/sdb2,serial=unknown io_time=36i,iops_in_progress=0i,merged_reads=0i,merged_writes=0i,read_bytes=2138112i,read_time=31i,reads=58i,weighted_io_time=31i,write_bytes=0i,write_time=0i,writes=0i",
				"diskio,host=HOST,name=/dev/sdb3,serial=unknown io_time=500292i,iops_in_progress=0i,merged_reads=253655i,merged_writes=1079685i,read_bytes=11279713280i,read_time=191469i,reads=503575i,weighted_io_time=632923i,write_bytes=35390902272i,write_time=441453i,writes=437354i",
			},
			wantErr: false,
		},
		{
			name: "ok with tag",
			fields: fields{
				diskIO:     DiskIO4Test,
				Tags:       map[string]string{"foo": "bar"},
				tagger:     testutils.DefaultMockTagger(),
				mergedTags: make(map[string]string),
			},
			want: []string{
				"diskio,foo=bar,host=HOST,name=/dev/loop0,serial=unknown io_time=40i,iops_in_progress=0i,merged_reads=0i,merged_writes=0i,read_bytes=1110016i,read_time=36i,reads=52i,weighted_io_time=36i,write_bytes=0i,write_time=0i,writes=0i",
				"diskio,foo=bar,host=HOST,name=/dev/loop1,serial=unknown io_time=20i,iops_in_progress=0i,merged_reads=0i,merged_writes=0i,read_bytes=355328i,read_time=8i,reads=43i,weighted_io_time=8i,write_bytes=0i,write_time=0i,writes=0i",
				"diskio,foo=bar,host=HOST,name=/dev/sda,serial=unknown io_time=608i,iops_in_progress=0i,merged_reads=0i,merged_writes=0i,read_bytes=7208448i,read_time=465i,reads=419i,weighted_io_time=465i,write_bytes=0i,write_time=0i,writes=0i",
				"diskio,foo=bar,host=HOST,name=/dev/sda1,serial=unknown io_time=36i,iops_in_progress=0i,merged_reads=0i,merged_writes=0i,read_bytes=626688i,read_time=20i,reads=53i,weighted_io_time=20i,write_bytes=0i,write_time=0i,writes=0i",
				"diskio,foo=bar,host=HOST,name=/dev/sda2,serial=unknown io_time=28i,iops_in_progress=0i,merged_reads=0i,merged_writes=0i,read_bytes=2173952i,read_time=26i,reads=56i,weighted_io_time=26i,write_bytes=0i,write_time=0i,writes=0i",
				"diskio,foo=bar,host=HOST,name=/dev/sdb,serial=unknown io_time=643584i,iops_in_progress=0i,merged_reads=392242i,merged_writes=1217830i,read_bytes=26472674816i,read_time=409419i,reads=1000749i,weighted_io_time=904930i,write_bytes=39289734144i,write_time=488838i,writes=552776i",
				"diskio,foo=bar,host=HOST,name=/dev/sdb1,serial=unknown io_time=184i,iops_in_progress=0i,merged_reads=29i,merged_writes=0i,read_bytes=6110208i,read_time=202i,reads=198i,weighted_io_time=206i,write_bytes=1024i,write_time=3i,writes=2i",
				"diskio,foo=bar,host=HOST,name=/dev/sdb2,serial=unknown io_time=36i,iops_in_progress=0i,merged_reads=0i,merged_writes=0i,read_bytes=2138112i,read_time=31i,reads=58i,weighted_io_time=31i,write_bytes=0i,write_time=0i,writes=0i",
				"diskio,foo=bar,host=HOST,name=/dev/sdb3,serial=unknown io_time=500292i,iops_in_progress=0i,merged_reads=253655i,merged_writes=1079685i,read_bytes=11279713280i,read_time=191469i,reads=503575i,weighted_io_time=632923i,write_bytes=35390902272i,write_time=441453i,writes=437354i",
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ipt := &Input{
				Interval:         tt.fields.Interval,
				Devices:          tt.fields.Devices,
				DeviceTags:       tt.fields.DeviceTags,
				NameTemplates:    tt.fields.NameTemplates,
				SkipSerialNumber: tt.fields.SkipSerialNumber,
				Tags:             tt.fields.Tags,
				collectCache:     tt.fields.collectCache,
				lastStat:         tt.fields.lastStat,
				lastTime:         tt.fields.lastTime,
				diskIO:           tt.fields.diskIO,
				feeder:           tt.fields.feeder,
				mergedTags:       tt.fields.mergedTags,
				tagger:           tt.fields.tagger,
				infoCache:        tt.fields.infoCache,
				deviceFilter:     tt.fields.deviceFilter,
				semStop:          tt.fields.semStop,
			}

			ipt.setup()

			err := ipt.collect()
			if (err != nil) != tt.wantErr {
				t.Errorf("error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err != nil {
				return
			}

			assert.NotEqual(t, len(ipt.collectCache), 0)

			gotStr := []string{}
			for _, v := range ipt.collectCache {
				s := v.LineProto()
				s = s[:strings.LastIndex(s, " ")]
				gotStr = append(gotStr, s)
			}
			sort.Strings(gotStr)

			assert.Equal(t, gotStr, tt.want)
		})
	}
}

func TestInput_rw_rate(t *testing.T) {
	type fields struct {
		Interval         time.Duration
		Devices          []string
		DeviceTags       []string
		NameTemplates    []string
		SkipSerialNumber bool
		Tags             map[string]string
		collectCache     []*point.Point
		lastStat         map[string]disk.IOCountersStat
		lastTime         time.Time
		diskIO           DiskIO
		feeder           dkio.Feeder
		mergedTags       map[string]string
		tagger           datakit.GlobalTagger
		infoCache        map[string]diskInfoCache
		deviceFilter     *DevicesFilter
		semStop          *cliutils.Sem
	}
	tests := []struct {
		name    string
		fields  fields
		want    []string
		wantErr bool
	}{
		{
			name: "ok with rate/sec",
			fields: fields{
				// diskIO:     DiskIO4Test,
				Devices:    []string{"^sdb[\\d]{0,2}"},
				Tags:       map[string]string{},
				tagger:     testutils.DefaultMockTagger(),
				mergedTags: make(map[string]string),
			},
			want: []string{
				"diskio,host=HOST,name=/dev/sdb,serial=unknown io_time=643584i,iops_in_progress=0i,merged_reads=392242i,merged_writes=1217830i,read_bytes=26472676050i,read_bytes/sec=1234i,read_time=409419i,reads=1000749i,weighted_io_time=904930i,write_bytes=39289734144i,write_bytes/sec=0i,write_time=488838i,writes=552776i",
				"diskio,host=HOST,name=/dev/sdb1,serial=unknown io_time=184i,iops_in_progress=0i,merged_reads=29i,merged_writes=0i,read_bytes=6110208i,read_bytes/sec=0i,read_time=202i,reads=198i,weighted_io_time=206i,write_bytes=1024i,write_bytes/sec=0i,write_time=3i,writes=2i",
				"diskio,host=HOST,name=/dev/sdb2,serial=unknown io_time=36i,iops_in_progress=0i,merged_reads=0i,merged_writes=0i,read_bytes=2138112i,read_bytes/sec=0i,read_time=31i,reads=58i,weighted_io_time=31i,write_bytes=5678i,write_bytes/sec=5678i,write_time=0i,writes=0i",
				"diskio,host=HOST,name=/dev/sdb3,serial=unknown io_time=500292i,iops_in_progress=0i,merged_reads=253655i,merged_writes=1079685i,read_bytes=11279713280i,read_bytes/sec=0i,read_time=191469i,reads=503575i,weighted_io_time=632923i,write_bytes=35390902272i,write_bytes/sec=0i,write_time=441453i,writes=437354i",
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ipt := &Input{
				Interval:         tt.fields.Interval,
				Devices:          tt.fields.Devices,
				DeviceTags:       tt.fields.DeviceTags,
				NameTemplates:    tt.fields.NameTemplates,
				SkipSerialNumber: tt.fields.SkipSerialNumber,
				Tags:             tt.fields.Tags,
				collectCache:     tt.fields.collectCache,
				lastStat:         tt.fields.lastStat,
				lastTime:         tt.fields.lastTime,
				diskIO:           tt.fields.diskIO,
				feeder:           tt.fields.feeder,
				mergedTags:       tt.fields.mergedTags,
				tagger:           tt.fields.tagger,
				infoCache:        tt.fields.infoCache,
				deviceFilter:     tt.fields.deviceFilter,
				semStop:          tt.fields.semStop,
			}

			ipt.setup()

			ipt.diskIO = DiskIO4Test
			err := ipt.collect()
			if (err != nil) != tt.wantErr {
				t.Errorf("error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err != nil {
				return
			}

			// 2nd loop
			time.Sleep(time.Second * 1)

			temp01 := testData["sdb"]
			temp01.ReadBytes += 1234
			testData["sdb"] = temp01

			temp02 := testData["sdb2"]
			temp02.WriteBytes += 5678
			testData["sdb2"] = temp02

			ipt.diskIO = DiskIO4Test
			err = ipt.collect()
			if (err != nil) != tt.wantErr {
				t.Errorf("error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err != nil {
				return
			}

			assert.NotEqual(t, len(ipt.collectCache), 0)

			gotStr := []string{}
			for _, v := range ipt.collectCache {
				s := v.LineProto()
				s = s[:strings.LastIndex(s, " ")]
				gotStr = append(gotStr, s)
			}
			sort.Strings(gotStr)

			assert.Equal(t, gotStr, tt.want)
		})
	}
}

// mock data

type IOCountersStat = disk.IOCountersStat

var testData = map[string]IOCountersStat{
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

func DiskIO4Test(names ...string) (map[string]disk.IOCountersStat, error) {
	m := map[string]IOCountersStat{}
	for k, v := range testData {
		m[k] = v
	}
	return m, nil
}
