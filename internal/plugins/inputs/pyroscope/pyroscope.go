// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package pyroscope

import (
	"bytes"
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/pyroscope-io/pyroscope/pkg/storage"
	"github.com/pyroscope-io/pyroscope/pkg/storage/tree"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs/profile/metrics"
	"google.golang.org/protobuf/proto"
)

const (
	pyroscopeFilename = "prof"
	nodeSpyName       = "nodespy"
	eBPFSpyName       = "ebpfspy"
)

type pyroscopeData struct {
	fileName string
	buf      *bytes.Buffer
}

func withExtName(f, ext string) string {
	if !strings.HasSuffix(f, ext) {
		return f + ext
	}
	return f
}

type cacheDetail struct {
	CPU          *storage.PutInput
	InuseObjects *storage.PutInput
	InuseSpace   *storage.PutInput
}

func (cd *cacheDetail) Len() int {
	length := 0

	if cd.CPU != nil {
		length++
	}
	if cd.InuseObjects != nil {
		length++
	}
	if cd.InuseSpace != nil {
		length++
	}

	return length
}

const splitSymbol = "{}"

func getReportCacheKeyName(putInput *storage.PutInput) string {
	// [app name] + [spy name] + [unit name] + [start time num]
	appName := putInput.Key.AppName()
	appName = strings.TrimSuffix(appName, ".cpu")
	appName = strings.TrimSuffix(appName, ".inuse_objects")
	appName = strings.TrimSuffix(appName, ".inuse_space")
	return appName + splitSymbol + putInput.SpyName
}

func getReportCacheKeyInt(name string) (int64, error) {
	arr := strings.Split(name, splitSymbol)
	if len(arr) != 3 {
		// impossible.
		return 0, fmt.Errorf("length not 3")
	}

	val, err := strconv.ParseInt(arr[2], 10, 64)
	if err != nil {
		return 0, err
	}

	return val, nil
}

type pyroscopeDatakitReport struct {
	endPoint  string
	inputTags map[string]string

	pyrs *pyroscopeOpts
}

func (report *pyroscopeDatakitReport) SetVar(opt *pyroscopeDatakitReport) {
	report.endPoint = opt.endPoint
	report.inputTags = opt.inputTags

	report.pyrs = opt.pyrs
}

func (report *pyroscopeDatakitReport) AutoClean() {
	length := 0
	report.pyrs.cacheData.Range(func(k, v interface{}) bool {
		length++
		return true
	})

	if length > 128 {
		report.pyrs.cacheData.Range(func(k, v interface{}) bool {
			name, ok := k.(string)
			if !ok {
				// impossible.
				report.pyrs.cacheData.Delete(k)
				return true
			}

			val, err := getReportCacheKeyInt(name)
			if err != nil {
				report.pyrs.cacheData.Delete(k)
				return true
			}

			tm := time.Unix(val, 0)

			duration := time.Since(tm)
			if duration > time.Minute {
				report.pyrs.cacheData.Delete(k)
			}

			return true
		})
	}
}

func (report *pyroscopeDatakitReport) LoadAndStore(name string, putInput *storage.PutInput) *cacheDetail {
	storefunc := func(cd *cacheDetail) *cacheDetail {
		switch putInput.Units.String() {
		case "samples": // cpu
			cd.CPU = putInput // Save and use the latest data.
		case "objects": // inuse_objects
			cd.InuseObjects = putInput // Save and use the latest data.
		case "bytes": // inuse_space
			cd.InuseSpace = putInput // Save and use the latest data.
		}
		report.pyrs.cacheData.Store(name, cd)
		if cd.Len() >= 3 {
			// Data is ready to send.
			return cd
		}
		return nil
	}

	actual, ok := report.pyrs.cacheData.Load(name)
	if !ok {
		// not found, store new.
		cd := &cacheDetail{}
		return storefunc(cd)
	}

	val, ok := actual.(*cacheDetail)
	if !ok {
		// impossible, cover older.
		cd := &cacheDetail{}
		return storefunc(cd)
	}

	// found, modify and store.
	return storefunc(val)
}

func (report *pyroscopeDatakitReport) Delete(name string) {
	report.pyrs.cacheData.Delete(name)
}

/*
Put implements storage.Putter interface

// pkg/storage/types.go

	type Putter interface {
		Put(context.Context, *PutInput) error
	}

	type PutInput struct {
		StartTime       time.Time
		EndTime         time.Time
		Key             *segment.Key
		Val             *tree.Tree
		SpyName         string
		SampleRate      uint32
		Units           metadata.Units
		AggregationType metadata.AggregationType
	}.
*/
func (report *pyroscopeDatakitReport) Put(ctx context.Context, putInput *storage.PutInput) error {
	var profileDataSet []*pyroscopeData
	var reportFamily metrics.Language
	var reportFormat metrics.Format
	var startTime, endTime time.Time

	spyName := putInput.SpyName

	switch spyName {
	case nodeSpyName:
		{
			// Must be these three, if find anything else we should change the code to adapt to it.
			appFullName := putInput.Key.AppName()
			checkOK := false
			switch {
			case strings.HasSuffix(appFullName, ".cpu"):
				checkOK = true
			case strings.HasSuffix(appFullName, ".inuse_objects"):
				checkOK = true
			case strings.HasSuffix(appFullName, ".inuse_space"):
				checkOK = true
			}
			if !checkOK {
				log.Errorf("putInput AppName checked failed: %s", appFullName)
			}
		}

		name := getReportCacheKeyName(putInput)
		if len(name) == 0 {
			return fmt.Errorf("getReportCacheKeyName == 0")
		}

		detail := report.LoadAndStore(name, putInput)
		if detail == nil {
			// Don't send yet.
			return nil
		}

		// Having cpu, inuse_objects, and inuse_space. Data is ready, start send.

		reportFamily = metrics.NodeJS
		reportFormat = metrics.PPROF

		// cpu
		cpuData, err := getBytesBufferByPut(detail.CPU)
		if err != nil {
			return err
		}
		// inuse_objects
		inuseObjectsData, err := getBytesBufferByPut(detail.InuseObjects)
		if err != nil {
			return err
		}
		// inuse_space
		inuseSpaceData, err := getBytesBufferByPut(detail.InuseSpace)
		if err != nil {
			return err
		}

		profileDataSet = []*pyroscopeData{
			{
				fileName: "cpu.pprof",
				buf:      cpuData,
			},
			{
				fileName: "inuse_objects.pprof",
				buf:      inuseObjectsData,
			},
			{
				fileName: "inuse_space.pprof",
				buf:      inuseSpaceData,
			},
		}

		// Use CPU sample collect time and set the collecting duration at fixed 10 seconds.
		// https://github.com/grafana/pyroscope-nodejs/blob/f0f5a777536fa010920e76c8112d3c62b5e04295/src/index.ts#L227
		// https://github.com/google/pprof-nodejs/blob/0eabf2d9a4e13456e642c41786fcb880a9119f28/ts/src/time-profiler.ts#L35-L36
		endTime = detail.CPU.StartTime
		startTime = detail.CPU.StartTime.Add(-10 * time.Second)

		report.Delete(name)

	case eBPFSpyName:
		reportFamily = metrics.CPP
		reportFormat = metrics.Collapsed

		report.inputTags["sample_rate"] = fmt.Sprintf("%d", putInput.SampleRate)
		report.inputTags["units"] = putInput.Units.String()
		report.inputTags["aggregation_type"] = putInput.AggregationType.String()

		collapsed := putInput.Val.Collapsed()
		var b bytes.Buffer
		b.WriteString(collapsed) // Write strings to the Buffer.

		profileDataSet = []*pyroscopeData{
			{
				fileName: pyroscopeFilename,
				buf:      &b,
			},
		}

		startTime = putInput.StartTime
		endTime = putInput.EndTime

	default:
		return fmt.Errorf("not supported format")
	}

	event := &metrics.Metadata{
		Language: reportFamily,
		Format:   reportFormat,
		Profiler: metrics.Pyroscope,
		Start:    metrics.NewRFC3339Time(startTime),
		End:      metrics.NewRFC3339Time(endTime),
		Attachments: []string{
			withExtName(pyroscopeFilename, ".pprof"),
		},
		TagsProfiler:  metrics.JoinTags(report.inputTags),
		SubCustomTags: metrics.JoinTags(report.pyrs.Tags),
	}

	if err := pushPyroscopeData(
		&pushPyroscopeDataOpt{
			startTime:       startTime,
			endTime:         endTime,
			pyroscopeData:   profileDataSet,
			endPoint:        report.endPoint,
			inputTags:       report.inputTags,
			inputNameSuffix: "/pyroscope/" + spyName,
			Input:           report.pyrs.input,
		},
		event,
		report.pyrs.input.GetBodySizeLimit(),
	); err != nil {
		log.Errorf("unable to push pyroscope profile data: %s", err)
		return err
	}

	return nil
}

func getBytesBufferByPut(putInput *storage.PutInput) (*bytes.Buffer, error) {
	bys, err := toPProf(putInput)
	if err != nil {
		return nil, err
	}

	var b bytes.Buffer
	_, err = b.Write(bys)
	if err != nil {
		return nil, err
	}

	return &b, nil
}

func toPProf(putInput *storage.PutInput) ([]byte, error) {
	pprof := putInput.Val.Pprof(&tree.PprofMetadata{
		// TODO(petethepig): not sure if this conversion is right
		Unit:      string(putInput.Units),
		StartTime: putInput.StartTime,
	})
	return proto.Marshal(pprof)
}
