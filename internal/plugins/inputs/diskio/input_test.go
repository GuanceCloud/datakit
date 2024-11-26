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
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs"
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
			name: "ok-with-devices",
			fields: fields{
				diskIO:     DiskIO4Test,
				Devices:    []string{"^sda", "vd."},
				Tags:       map[string]string{},
				tagger:     testutils.DefaultMockTagger(),
				mergedTags: make(map[string]string),
			},
			want: []string{
				"diskio,host=HOST,name=/dev/sda,serial=unknown io_time=608u,iops_in_progress=0u,merged_reads=0u,merged_writes=0u,read_bytes=7208448u,read_time=465u,reads=419u,weighted_io_time=465u,write_bytes=0u,write_time=0u,writes=0u",
				"diskio,host=HOST,name=/dev/sda1,serial=unknown io_time=36u,iops_in_progress=0u,merged_reads=0u,merged_writes=0u,read_bytes=626688u,read_time=20u,reads=53u,weighted_io_time=20u,write_bytes=0u,write_time=0u,writes=0u",
				"diskio,host=HOST,name=/dev/sda2,serial=unknown io_time=28u,iops_in_progress=0u,merged_reads=0u,merged_writes=0u,read_bytes=2173952u,read_time=26u,reads=56u,weighted_io_time=26u,write_bytes=0u,write_time=0u,writes=0u",
			},
			wantErr: false,
		},
		{
			name: "ok-no-tag",
			fields: fields{
				diskIO:     DiskIO4Test,
				Tags:       map[string]string{},
				tagger:     testutils.DefaultMockTagger(),
				mergedTags: make(map[string]string),
			},
			want: []string{
				"diskio,host=HOST,name=/dev/loop0,serial=unknown io_time=40u,iops_in_progress=0u,merged_reads=0u,merged_writes=0u,read_bytes=1110016u,read_time=36u,reads=52u,weighted_io_time=36u,write_bytes=0u,write_time=0u,writes=0u",
				"diskio,host=HOST,name=/dev/loop1,serial=unknown io_time=20u,iops_in_progress=0u,merged_reads=0u,merged_writes=0u,read_bytes=355328u,read_time=8u,reads=43u,weighted_io_time=8u,write_bytes=0u,write_time=0u,writes=0u",
				"diskio,host=HOST,name=/dev/sda,serial=unknown io_time=608u,iops_in_progress=0u,merged_reads=0u,merged_writes=0u,read_bytes=7208448u,read_time=465u,reads=419u,weighted_io_time=465u,write_bytes=0u,write_time=0u,writes=0u",
				"diskio,host=HOST,name=/dev/sda1,serial=unknown io_time=36u,iops_in_progress=0u,merged_reads=0u,merged_writes=0u,read_bytes=626688u,read_time=20u,reads=53u,weighted_io_time=20u,write_bytes=0u,write_time=0u,writes=0u",
				"diskio,host=HOST,name=/dev/sda2,serial=unknown io_time=28u,iops_in_progress=0u,merged_reads=0u,merged_writes=0u,read_bytes=2173952u,read_time=26u,reads=56u,weighted_io_time=26u,write_bytes=0u,write_time=0u,writes=0u",
				"diskio,host=HOST,name=/dev/sdb,serial=unknown io_time=643584u,iops_in_progress=0u,merged_reads=392242u,merged_writes=1217830u,read_bytes=26472674816u,read_time=409419u,reads=1000749u,weighted_io_time=904930u,write_bytes=39289734144u,write_time=488838u,writes=552776u",
				"diskio,host=HOST,name=/dev/sdb1,serial=unknown io_time=184u,iops_in_progress=0u,merged_reads=29u,merged_writes=0u,read_bytes=6110208u,read_time=202u,reads=198u,weighted_io_time=206u,write_bytes=1024u,write_time=3u,writes=2u",
				"diskio,host=HOST,name=/dev/sdb2,serial=unknown io_time=36u,iops_in_progress=0u,merged_reads=0u,merged_writes=0u,read_bytes=2138112u,read_time=31u,reads=58u,weighted_io_time=31u,write_bytes=0u,write_time=0u,writes=0u",
				"diskio,host=HOST,name=/dev/sdb3,serial=unknown io_time=500292u,iops_in_progress=0u,merged_reads=253655u,merged_writes=1079685u,read_bytes=11279713280u,read_time=191469u,reads=503575u,weighted_io_time=632923u,write_bytes=35390902272u,write_time=441453u,writes=437354u",
			},
			wantErr: false,
		},
		{
			name: "ok-with-tag",
			fields: fields{
				diskIO:     DiskIO4Test,
				Tags:       map[string]string{"foo": "bar"},
				tagger:     testutils.DefaultMockTagger(),
				mergedTags: make(map[string]string),
			},
			want: []string{
				"diskio,foo=bar,host=HOST,name=/dev/loop0,serial=unknown io_time=40u,iops_in_progress=0u,merged_reads=0u,merged_writes=0u,read_bytes=1110016u,read_time=36u,reads=52u,weighted_io_time=36u,write_bytes=0u,write_time=0u,writes=0u",
				"diskio,foo=bar,host=HOST,name=/dev/loop1,serial=unknown io_time=20u,iops_in_progress=0u,merged_reads=0u,merged_writes=0u,read_bytes=355328u,read_time=8u,reads=43u,weighted_io_time=8u,write_bytes=0u,write_time=0u,writes=0u",
				"diskio,foo=bar,host=HOST,name=/dev/sda,serial=unknown io_time=608u,iops_in_progress=0u,merged_reads=0u,merged_writes=0u,read_bytes=7208448u,read_time=465u,reads=419u,weighted_io_time=465u,write_bytes=0u,write_time=0u,writes=0u",
				"diskio,foo=bar,host=HOST,name=/dev/sda1,serial=unknown io_time=36u,iops_in_progress=0u,merged_reads=0u,merged_writes=0u,read_bytes=626688u,read_time=20u,reads=53u,weighted_io_time=20u,write_bytes=0u,write_time=0u,writes=0u",
				"diskio,foo=bar,host=HOST,name=/dev/sda2,serial=unknown io_time=28u,iops_in_progress=0u,merged_reads=0u,merged_writes=0u,read_bytes=2173952u,read_time=26u,reads=56u,weighted_io_time=26u,write_bytes=0u,write_time=0u,writes=0u",
				"diskio,foo=bar,host=HOST,name=/dev/sdb,serial=unknown io_time=643584u,iops_in_progress=0u,merged_reads=392242u,merged_writes=1217830u,read_bytes=26472674816u,read_time=409419u,reads=1000749u,weighted_io_time=904930u,write_bytes=39289734144u,write_time=488838u,writes=552776u",
				"diskio,foo=bar,host=HOST,name=/dev/sdb1,serial=unknown io_time=184u,iops_in_progress=0u,merged_reads=29u,merged_writes=0u,read_bytes=6110208u,read_time=202u,reads=198u,weighted_io_time=206u,write_bytes=1024u,write_time=3u,writes=2u",
				"diskio,foo=bar,host=HOST,name=/dev/sdb2,serial=unknown io_time=36u,iops_in_progress=0u,merged_reads=0u,merged_writes=0u,read_bytes=2138112u,read_time=31u,reads=58u,weighted_io_time=31u,write_bytes=0u,write_time=0u,writes=0u",
				"diskio,foo=bar,host=HOST,name=/dev/sdb3,serial=unknown io_time=500292u,iops_in_progress=0u,merged_reads=253655u,merged_writes=1079685u,read_bytes=11279713280u,read_time=191469u,reads=503575u,weighted_io_time=632923u,write_bytes=35390902272u,write_time=441453u,writes=437354u",
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
			intervalMillSec := ipt.Interval.Milliseconds()
			var lastAlignTime int64
			tn := time.Now()
			lastAlignTime = inputs.AlignTimeMillSec(tn, lastAlignTime, intervalMillSec)

			err := ipt.collect(tn, lastAlignTime*1e6)
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
				"diskio,host=HOST,name=/dev/sdb,serial=unknown io_time=643584u,iops_in_progress=0u,merged_reads=392242u,merged_writes=1217830u,read_bytes=26472676050u,read_bytes/sec=1234i,read_time=409419u,reads=1000749u,weighted_io_time=904930u,write_bytes=39289734144u,write_bytes/sec=0i,write_time=488838u,writes=552776u",
				"diskio,host=HOST,name=/dev/sdb1,serial=unknown io_time=184u,iops_in_progress=0u,merged_reads=29u,merged_writes=0u,read_bytes=6110208u,read_bytes/sec=0i,read_time=202u,reads=198u,weighted_io_time=206u,write_bytes=1024u,write_bytes/sec=0i,write_time=3u,writes=2u",
				"diskio,host=HOST,name=/dev/sdb2,serial=unknown io_time=36u,iops_in_progress=0u,merged_reads=0u,merged_writes=0u,read_bytes=2138112u,read_bytes/sec=0i,read_time=31u,reads=58u,weighted_io_time=31u,write_bytes=5678u,write_bytes/sec=5678i,write_time=0u,writes=0u",
				"diskio,host=HOST,name=/dev/sdb3,serial=unknown io_time=500292u,iops_in_progress=0u,merged_reads=253655u,merged_writes=1079685u,read_bytes=11279713280u,read_bytes/sec=0i,read_time=191469u,reads=503575u,weighted_io_time=632923u,write_bytes=35390902272u,write_bytes/sec=0i,write_time=441453u,writes=437354u",
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
			intervalMillSec := ipt.Interval.Milliseconds()
			var lastAlignTime int64
			tn := time.Now()
			lastAlignTime = inputs.AlignTimeMillSec(tn, lastAlignTime, intervalMillSec)

			err := ipt.collect(tn, lastAlignTime*1e6)
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
			tn = time.Now()
			lastAlignTime = inputs.AlignTimeMillSec(tn, lastAlignTime, intervalMillSec)
			err = ipt.collect(tn, lastAlignTime*1e6)
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
