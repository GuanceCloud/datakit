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
	"github.com/pyroscope-io/pyroscope/pkg/convert"
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
	"github.com/pyroscope-io/pyroscope/pkg/structs/transporttrie"
	"github.com/pyroscope-io/pyroscope/pkg/util/attime"
	"github.com/pyroscope-io/pyroscope/pkg/util/cumulativepprof"
	"github.com/sirupsen/logrus"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
)

var g = datakit.G("pyroscope")

const (
	pyroscopeReportFormat = "rawflamegraph"
	pyroscopeFilename     = "prof"
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
	ingester := parser.New(nil, nil, metricsExporter) // if !svc.config.RemoteWrite.Enabled || !svc.config.RemoteWrite.DisableLocalWrites

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
	}
	igHdr.ServeHTTP(c.Writer, c.Request)
}

//------------------------------------------------------------------------------

type ingestHandler struct {
	ingester               ingestion.Ingester
	httpUtils              httputils.Utils
	disableCumulativeMerge bool
	exporter               storage.MetricsExporter

	pyrs *pyroscopeOpts
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

func (h ingestHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	input, err := h.ingestInputFromRequest(r)
	if err != nil {
		writeClientError(h.httpUtils, r, w, "invalid parameter", err, http.StatusBadRequest)
		return
	}

	putInput := &storage.PutInput{
		StartTime:       input.Metadata.StartTime,
		EndTime:         input.Metadata.EndTime,
		Key:             input.Metadata.Key,
		SpyName:         input.Metadata.SpyName,
		SampleRate:      input.Metadata.SampleRate,
		Units:           input.Metadata.Units,
		AggregationType: input.Metadata.AggregationType,
		Val:             tree.New(),
	}

	cb := createParseCallback(putInput, h.exporter)
	bys, err := input.Profile.Bytes()
	if err != nil {
		writeClientError(h.httpUtils, r, w, "get profile RawData failed", err, http.StatusBadRequest)
		return
	}

	rd := bytes.NewReader(bys)

	switch input.Format { //nolint:exhaustive
	case ingestion.FormatTrie:
		err = transporttrie.IterateRaw(rd, make([]byte, 0, 256), cb)
	case ingestion.FormatTree:
		err = convert.ParseTreeNoDict(rd, cb)
	case ingestion.FormatLines:
		err = convert.ParseIndividualLines(rd, cb)
	case ingestion.FormatGroups:
		err = convert.ParseGroups(rd, cb)
	default:
		err = fmt.Errorf("unknown format %q", input.Format)
	}

	if err != nil {
		writeClientError(h.httpUtils, r, w, "invalid input format", err, http.StatusBadRequest)
		return
	}

	spyName := input.Metadata.SpyName

	addTags(h.pyrs.tags, "app_name", input.Metadata.Key.AppName())
	addTags(h.pyrs.tags, "input_format", string(input.Format))
	addTags(h.pyrs.tags, "spy_name", spyName)
	addTags(h.pyrs.tags, "sample_rate", fmt.Sprintf("%d", input.Metadata.SampleRate))
	addTags(h.pyrs.tags, "units", input.Metadata.Units.String())
	addTags(h.pyrs.tags, "aggregation_type", input.Metadata.AggregationType.String())
	labels := getPyroscopeTagFromLabels(input.Metadata.Key.Labels())
	for k, v := range labels {
		addTags(h.pyrs.tags, k, v)
	}

	collapsed := putInput.Val.Collapsed()

	var b bytes.Buffer
	b.WriteString(collapsed) // Write strings to the Buffer.

	langFamily := getLangFamilyFromSpyName(spyName)

	if err := pushProfileData(
		&pushProfileDataOpt{
			startTime: input.Metadata.StartTime,
			endTime:   input.Metadata.EndTime,
			profiledatas: []*profileData{
				{
					fileName: pyroscopeFilename,
					buf:      &b,
				},
			},
			reportFamily:    langFamily,
			reportFormat:    pyroscopeReportFormat,
			endPoint:        h.pyrs.URL,
			inputTags:       h.pyrs.tags,
			election:        false,
			inputNameSuffix: "/pyroscope/" + input.Metadata.SpyName,
		},
	); err != nil {
		writeClientError(h.httpUtils, r, w, "pushProfileData failed", err, http.StatusBadRequest)
		return
	}
}

func getLangFamilyFromSpyName(spyName string) string {
	if length := len(spyName); length > 3 {
		return spyName[:length-3]
	}
	return spyName
}

func copyBody(r *http.Request) ([]byte, error) {
	buf := bytes.NewBuffer(make([]byte, 0, 64<<10))
	if _, err := io.Copy(buf, r.Body); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func createParseCallback(pi *storage.PutInput, e storage.MetricsExporter) func([]byte, int) {
	o, ok := e.Evaluate(pi)
	if !ok {
		return pi.Val.InsertInt
	}
	return func(k []byte, v int) {
		o.Observe(k, v)
		pi.Val.InsertInt(k, v)
	}
}

//------------------------------------------------------------------------------

func writeClientError(hus httputils.Utils, r *http.Request, w http.ResponseWriter, action string, err error, code int) {
	log.Errorf("%s: %v", action, err)
	hus.WriteError(r, w, code, err, action)
}
