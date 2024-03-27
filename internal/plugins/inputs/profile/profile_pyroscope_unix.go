// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

//go:build !windows && !arm && !386
// +build !windows,!arm,!386

package profile

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/pyroscope-io/pyroscope/pkg/agent/types"
	pyroscopeConfig "github.com/pyroscope-io/pyroscope/pkg/config"
	"github.com/pyroscope-io/pyroscope/pkg/convert/jfr"
	"github.com/pyroscope-io/pyroscope/pkg/convert/pprof"
	"github.com/pyroscope-io/pyroscope/pkg/convert/profile"
	"github.com/pyroscope-io/pyroscope/pkg/convert/speedscope"
	"github.com/pyroscope-io/pyroscope/pkg/exporter"
	"github.com/pyroscope-io/pyroscope/pkg/ingestion"
	"github.com/pyroscope-io/pyroscope/pkg/parser"
	"github.com/pyroscope-io/pyroscope/pkg/server/httputils"
	"github.com/pyroscope-io/pyroscope/pkg/storage"
	"github.com/pyroscope-io/pyroscope/pkg/storage/metadata"
	"github.com/pyroscope-io/pyroscope/pkg/storage/segment"
	"github.com/pyroscope-io/pyroscope/pkg/storage/tree"
	"github.com/pyroscope-io/pyroscope/pkg/util/attime"
	"github.com/pyroscope-io/pyroscope/pkg/util/cumulativepprof"
	"github.com/sirupsen/logrus"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/datakit"
	"google.golang.org/protobuf/proto"
)

var g = datakit.G("pyroscope")

const (
	pyroscopeFilename = "prof"

	nodeSpyName = "nodespy"
	eBPFSpyName = "ebpfspy"
)

// init check config and set config.
func (pyrs *pyroscopeOpts) init() error {
	// tags set
	pyrs.tags = map[string]string{
		"service": pyrs.Service,
		"version": pyrs.Version,
		"env":     pyrs.Env,
	}
	for k, v := range pyrs.Tags {
		pyrs.tags[k] = v
	}

	return nil
}

// run pyroscope profiling.
func (pyrs *pyroscopeOpts) run(i *Input) error {
	if i == nil {
		return fmt.Errorf("input expected not to be nil")
	}

	if i.pause {
		log.Debugf("not leader, skipped")
		return nil
	}

	pyrs.input = i
	if err := pyrs.init(); err != nil {
		return fmt.Errorf("init pyroscope profiler error: %w", err)
	}

	router := gin.New()
	router.Use(APIMiddleware(pyrs))
	router.POST("/ingest", ingestHandle)

	log.Debugf("HTTP bind addr:%s", pyrs.URL)

	srv := &http.Server{
		Addr:    pyrs.URL,
		Handler: router,
	}

	g.Go(func(ctx context.Context) error {
		// service connections.
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Errorf("ListenAndServe failed: %v\n", err)
			return err
		}
		return nil
	})

	stopFunc := func() {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		if err := srv.Shutdown(ctx); err != nil && errors.Is(err, http.ErrServerClosed) {
			log.Errorf("pyroscope profiling server Shutdown failed: %v", err)
		}

		// catching ctx.Done(). timeout of 1 seconds.
		<-ctx.Done()
	}

	for {
		select {
		case <-datakit.Exit.Wait():
			log.Info("pyroscope profiling exit")
			stopFunc()
			return nil
		case <-i.semStop.Wait():
			log.Info("pyroscope profiling stop")
			stopFunc()
			return nil
		}
	}
}

// APIMiddleware pass pyroscope struct point to the context.
func APIMiddleware(pyrs *pyroscopeOpts) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Set("pyroscope_pointer", pyrs)
		c.Next()
	}
}

func ingestHandle(c *gin.Context) {
	logger := logrus.New()
	logger.Out = ioutil.Discard
	hus := httputils.NewDefaultHelper(logger)

	exportedMetricsRegistry := prometheus.NewRegistry()
	metricsExporter, err := exporter.NewExporter(pyroscopeConfig.MetricsExportRules{}, exportedMetricsRegistry)
	if err != nil {
		writeClientError(hus, c.Request, c.Writer, "new metric exporter failed", err, http.StatusInternalServerError)
		return
	}

	report := &pyroscopeDatakitReport{}
	ingester := parser.New(nil, report, metricsExporter) // if !svc.config.RemoteWrite.Enabled || !svc.config.RemoteWrite.DisableLocalWrites

	pyrs, ok := c.MustGet("pyroscope_pointer").(*pyroscopeOpts)
	if !ok {
		writeClientError(hus, c.Request, c.Writer, "get pyroscope pointer failed", err, http.StatusInternalServerError)
		return
	}

	igHdr := ingestHandler{
		ingester:               ingester,
		httpUtils:              hus,
		disableCumulativeMerge: true, // disableCumulativeMerge := !ctrl.config.RemoteWrite.Enabled
		exporter:               metricsExporter,
		pyrs:                   pyrs,
		report:                 report,
	}
	igHdr.ServeHTTP(c.Writer, c.Request)
}

////////////////////////////////////////////////////////////////////////////////

type ingestHandler struct {
	ingester               ingestion.Ingester
	httpUtils              httputils.Utils
	disableCumulativeMerge bool
	exporter               storage.MetricsExporter

	pyrs   *pyroscopeOpts
	report *pyroscopeDatakitReport
}

func (h ingestHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	input, err := h.ingestInputFromRequest(r)
	if err != nil {
		writeClientError(h.httpUtils, r, w, "invalid parameter", err, http.StatusBadRequest)
		return
	}

	newTags := internal.CopyMapString(h.pyrs.tags)
	originAddTagsSafe(newTags, "app_name", input.Metadata.Key.AppName())
	originAddTagsSafe(newTags, "input_format", string(input.Format))
	originAddTagsSafe(newTags, "spy_name", input.Metadata.SpyName)
	labels := getPyroscopeTagFromLabels(input.Metadata.Key.Labels())
	for k, v := range labels {
		originAddTagsSafe(newTags, k, v)
	}

	h.report.SetVar(&pyroscopeDatakitReport{
		endPoint:  h.pyrs.URL,
		inputTags: newTags,
		pyrs:      h.pyrs,
	})

	err = h.ingester.Ingest(r.Context(), input)
	switch {
	case err == nil:
		h.report.AutoClean()
	case ingestion.IsIngestionError(err):
		h.httpUtils.WriteError(r, w, http.StatusInternalServerError, err, "error happened while ingesting data")
	default:
		h.httpUtils.WriteError(r, w, http.StatusUnprocessableEntity, err, "error happened while parsing request body")
	}
}

func (h ingestHandler) ingestInputFromRequest(r *http.Request) (*ingestion.IngestInput, error) {
	var (
		q     = r.URL.Query()
		input ingestion.IngestInput
		err   error
	)

	input.Metadata.Key, err = segment.ParseKey(q.Get("name"))
	if err != nil {
		return nil, fmt.Errorf("name: %w", err)
	}

	if qt := q.Get("from"); qt != "" {
		input.Metadata.StartTime = attime.Parse(qt)
	} else {
		input.Metadata.StartTime = time.Now()
	}

	if qt := q.Get("until"); qt != "" {
		input.Metadata.EndTime = attime.Parse(qt)
	} else {
		input.Metadata.EndTime = time.Now()
	}

	if sr := q.Get("sampleRate"); sr != "" {
		sampleRate, err := strconv.Atoi(sr)
		if err != nil {
			// h.log.WithError(err).Errorf("invalid sample rate: %q", sr)
			input.Metadata.SampleRate = types.DefaultSampleRate
		} else {
			input.Metadata.SampleRate = uint32(sampleRate)
		}
	} else {
		input.Metadata.SampleRate = types.DefaultSampleRate
	}

	if sn := q.Get("spyName"); sn != "" {
		// TODO: error handling
		input.Metadata.SpyName = sn
	} else {
		input.Metadata.SpyName = "unknown"
	}

	if u := q.Get("units"); u != "" {
		// TODO(petethepig): add validation for these?
		input.Metadata.Units = metadata.Units(u)
	} else {
		input.Metadata.Units = metadata.SamplesUnits
	}

	if at := q.Get("aggregationType"); at != "" {
		// TODO(petethepig): add validation for these?
		input.Metadata.AggregationType = metadata.AggregationType(at)
	} else {
		input.Metadata.AggregationType = metadata.SumAggregationType
	}

	b, err := copyBody(r)
	if err != nil {
		return nil, err
	}

	format := q.Get("format")
	contentType := r.Header.Get("Content-Type")
	switch {
	default:
		input.Format = ingestion.FormatGroups
	case format == "trie", contentType == "binary/octet-stream+trie":
		input.Format = ingestion.FormatTrie
	case format == "tree", contentType == "binary/octet-stream+tree":
		input.Format = ingestion.FormatTree
	case format == "lines":
		input.Format = ingestion.FormatLines

	case format == "jfr":
		input.Format = ingestion.FormatJFR
		input.Profile = &jfr.RawProfile{
			FormDataContentType: contentType,
			RawData:             b,
		}

	case format == "pprof":
		input.Format = ingestion.FormatPprof
		input.Profile = &pprof.RawProfile{
			RawData: b,
		}

	case format == "speedscope":
		input.Format = ingestion.FormatSpeedscope
		input.Profile = &speedscope.RawProfile{
			RawData: b,
		}

	case strings.Contains(contentType, "multipart/form-data"):
		p := &pprof.RawProfile{
			FormDataContentType: contentType,
			RawData:             b,
		}
		if !h.disableCumulativeMerge {
			p.MergeCumulative(cumulativepprof.NewMergers()) //nolint:errcheck,gosec
		}
		input.Profile = p
	}

	if input.Profile == nil {
		input.Profile = &profile.RawProfile{
			Format:  input.Format,
			RawData: b,
		}
	}

	return &input, nil
}

////////////////////////////////////////////////////////////////////////////////

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
	var profiledatas []*profileData
	var reportFamily, reportFormat string
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

		reportFamily = "nodejs"
		reportFormat = "pprof"

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

		profiledatas = []*profileData{
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
		reportFamily = "ebpf"
		reportFormat = "rawflamegraph"

		report.inputTags["sample_rate"] = fmt.Sprintf("%d", putInput.SampleRate)
		report.inputTags["units"] = putInput.Units.String()
		report.inputTags["aggregation_type"] = putInput.AggregationType.String()

		collapsed := putInput.Val.Collapsed()
		var b bytes.Buffer
		b.WriteString(collapsed) // Write strings to the Buffer.

		profiledatas = []*profileData{
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

	event := &eventOpts{
		Family:   reportFamily,
		Format:   reportFormat,
		Profiler: "pyroscope",
		Start:    startTime.Format(time.RFC3339Nano),
		End:      endTime.Format(time.RFC3339Nano),
		Attachments: []string{
			withExtName(pyroscopeFilename, ".pprof"),
		},
		TagsProfiler: joinMap(report.inputTags),
	}

	if err := pushProfileData(
		&pushProfileDataOpt{
			startTime:       startTime,
			endTime:         endTime,
			profiledatas:    profiledatas,
			endPoint:        report.endPoint,
			inputTags:       report.inputTags,
			inputNameSuffix: "/pyroscope/" + spyName,
			Input:           report.pyrs.input,
		},
		event,
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

////////////////////////////////////////////////////////////////////////////////

func copyBody(r *http.Request) ([]byte, error) {
	buf := bytes.NewBuffer(make([]byte, 0, 64<<10))
	if _, err := io.Copy(buf, r.Body); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func writeClientError(hus httputils.Utils, r *http.Request, w http.ResponseWriter, action string, err error, code int) {
	log.Errorf("%s: %v", action, err)
	hus.WriteError(r, w, code, err, action)
}
