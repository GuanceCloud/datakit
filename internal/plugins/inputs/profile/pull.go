// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package profile

import (
	"bytes"
	"context"
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"path"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs/profile/metrics"

	pprofile "github.com/google/pprof/profile"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/datakit"
)

const PullInputMode = "profile-pull"

// GoProfiler pull go pprof data.
type GoProfiler struct {
	URL          string            `toml:"url"`
	Interval     string            `toml:"interval"`
	Service      string            `toml:"service"`
	Env          string            `toml:"env"`
	Version      string            `toml:"version"`
	Tags         map[string]string `toml:"tags"`
	EnabledTypes []string          `toml:"enabled_types"` // cpu,goroutine,heap,mutex,block
	// ProfileDuration controls the duration of CPU profiles. The default is 10s.
	ProfileDuration string `toml:"profile_duration"`
	// HTTPTimeout controls each pprof request. It must be longer than ProfileDuration.
	HTTPTimeout string `toml:"http_timeout"`

	TLSOpen            bool   `toml:"tls_open"`
	CacertFile         string `toml:"tls_ca"`
	CertFile           string `toml:"tls_cert"`
	KeyFile            string `toml:"tls_key"`
	InsecureSkipVerify bool   `toml:"insecure_skip_verify"`

	url              *url.URL
	interval         time.Duration
	duration         time.Duration
	timeout          time.Duration
	tags             map[string]string
	client           *http.Client
	deltas           map[string]*pprofile.Profile
	deltaCache       *profileDeltaCache
	deltaClosed      bool
	validateResponse bool
	input            *Input
}

type profileData struct {
	fileName  string
	buf       *bytes.Buffer
	startTime time.Time
	endTime   time.Time
}

type valueType struct {
	Type string
	Unit string
}

type Item struct {
	path        string
	params      url.Values
	fileName    string
	deltaValues []valueType
}

var profileConfigMap = map[string]Item{
	"cpu": {
		path: "/profile",
		params: url.Values{
			"seconds": []string{"10"},
		},
		fileName: "cpu.pprof",
	},
	"goroutine": {
		path:     "/goroutine",
		fileName: "goroutines.pprof",
	},
	"heap": {
		path:     "/heap",
		fileName: "delta-heap.pprof",
		deltaValues: []valueType{
			{Type: "alloc_objects", Unit: "count"},
			{Type: "alloc_space", Unit: "bytes"},
		},
	},
	"mutex": {
		path:     "/mutex",
		fileName: "delta-mutex.pprof",
		deltaValues: []valueType{
			{Type: "contentions", Unit: "count"},
			{Type: "delay", Unit: "nanoseconds"},
		},
	},
	"block": {
		path:     "/block",
		fileName: "delta-block.pprof",
		deltaValues: []valueType{
			{Type: "contentions", Unit: "count"},
			{Type: "delay", Unit: "nanoseconds"},
		},
	},
}

// init check config and set config.
func (g *GoProfiler) init() error {
	var (
		duration time.Duration
		err      error
	)
	if g.Interval != "" {
		duration, err = time.ParseDuration(g.Interval)
		if err != nil {
			return err
		}
	}
	// duration should be larger than 10s
	if duration < 10*time.Second {
		duration = 10 * time.Second
	}
	g.interval = duration

	g.duration = 10 * time.Second
	if g.ProfileDuration != "" {
		g.duration, err = time.ParseDuration(g.ProfileDuration)
		if err != nil {
			return fmt.Errorf("invalid profile_duration: %w", err)
		}
	}
	if g.duration < time.Second {
		return fmt.Errorf("profile_duration must be at least 1s")
	}

	g.timeout = g.duration + 5*time.Second
	if g.timeout < 15*time.Second {
		g.timeout = 15 * time.Second
	}
	if g.HTTPTimeout != "" {
		g.timeout, err = time.ParseDuration(g.HTTPTimeout)
		if err != nil {
			return fmt.Errorf("invalid http_timeout: %w", err)
		}
	}
	if g.timeout <= g.duration {
		return fmt.Errorf("http_timeout must be greater than profile_duration")
	}

	// url parse
	g.url, err = url.Parse(g.URL)
	if err != nil {
		return fmt.Errorf("invalid pprof URL")
	}
	if g.url.Scheme != "http" && g.url.Scheme != "https" {
		return fmt.Errorf("unsupported pprof URL scheme %q", g.url.Scheme)
	}
	if g.url.Host == "" {
		return fmt.Errorf("pprof URL host cannot be empty")
	}
	if g.url.Path == "" || g.url.Path == "/" {
		g.url.Path = "/debug/pprof"
	}

	// tags set
	g.tags = map[string]string{
		"service": g.Service,
		"version": g.Version,
		"env":     g.Env,
	}
	for k, v := range g.Tags {
		g.tags[k] = v
	}

	if client, err := g.createHTTPClient(); err != nil {
		return err
	} else {
		g.client = client
	}

	g.deltas = make(map[string]*pprofile.Profile)

	return nil
}

// run pull profile.
func (g *GoProfiler) run(i *Input) error {
	defer func() {
		log.Warnf("%s handler stopped", PullInputMode)
	}()

	if i == nil {
		return fmt.Errorf("input expected not to be nil")
	}
	g.input = i

	if err := g.init(); err != nil {
		return fmt.Errorf("init go profiler error: %w", err)
	}

	tick := time.NewTicker(g.interval)
	defer tick.Stop()

	once := new(sync.Once)

	for {
		if i.pause.Load() {
			log.Debugf("not leader, skipped")
		} else {
			once.Do(func() {
				log.Infof("profiling pull mode start....")
			})
			g.pullProfile()
		}

		select {
		case <-datakit.Exit.Wait():
			return nil
		case <-i.semStop.Wait():
			log.Info("go profiler exit")
			return nil
		case <-tick.C:
		}
	}
}

func withExtName(f, ext string) string {
	if !strings.HasSuffix(f, ext) {
		return f + ext
	}
	return f
}

func (g *GoProfiler) pullProfile() {
	if err := g.collect(context.Background(), g.EnabledTypes, g.duration, nil); err != nil {
		log.Warnf("pull profile from %s failed: %s", safeProfileURL(g.url), err)
	}
}

// collect pulls and queues a set of Go profiles. A GoProfiler is stateful because
// heap, mutex, and block profiles are converted to deltas, so callers must not
// invoke collect concurrently for the same profiler.
func (g *GoProfiler) collect(
	ctx context.Context,
	enabledTypes []string,
	duration time.Duration,
	extraTags map[string]string,
) error {
	var (
		deltaData      []*profileData
		deltaFileNames []string
		firstErr       error
	)

	allTags := copyTags(g.tags)
	customTags := copyTags(g.Tags)
	for k, v := range extraTags {
		allTags[k] = v
		customTags[k] = v
	}

	for _, k := range enabledTypes {
		if err := ctx.Err(); err != nil {
			return err
		}
		if p, ok := profileConfigMap[k]; ok {
			if pData, err := g.pullProfileItem(ctx, k, p, duration); err != nil {
				log.Warnf("profile for %s error: %s", k, err.Error())
				if firstErr == nil {
					firstErr = err
				}
			} else if pData != nil {
				if p.deltaValues != nil {
					deltaData = append(deltaData, pData)
					deltaFileNames = append(deltaFileNames, withExtName(pData.fileName, ".pprof"))
				} else if err := pushProfileData(
					&pushProfileDataOpt{
						startTime:       pData.startTime,
						endTime:         pData.endTime,
						profiledatas:    []*profileData{pData},
						endPoint:        safeProfileURL(g.url),
						inputTags:       allTags,
						inputNameSuffix: "/go",
						Input:           g.input,
					},
					&metrics.Metadata{
						Language:      metrics.Golang,
						Format:        metrics.PPROF,
						Profiler:      metrics.GoPProf,
						Start:         metrics.NewRFC3339Time(pData.startTime),
						End:           metrics.NewRFC3339Time(pData.endTime),
						Attachments:   []string{withExtName(pData.fileName, ".pprof")},
						TagsProfiler:  metrics.JoinTags(allTags),
						SubCustomTags: metrics.JoinTags(customTags),
					},
					g.input.GetBodySizeLimit(),
				); err != nil {
					log.Warnf("push profile data error: %s", err.Error())
					if firstErr == nil {
						firstErr = err
					}
				}
			}
		} else {
			log.Warnf("invalid profile type: %s", k)
			if firstErr == nil {
				firstErr = fmt.Errorf("invalid profile type %q", k)
			}
		}
	}

	// push delta profiles together
	if err := ctx.Err(); err != nil {
		return err
	}
	if len(deltaData) > 0 {
		pData := deltaData[0]
		if err := pushProfileData(
			&pushProfileDataOpt{
				startTime:       pData.startTime,
				endTime:         pData.endTime,
				profiledatas:    deltaData,
				endPoint:        safeProfileURL(g.url),
				inputTags:       allTags,
				inputNameSuffix: "/go",
				Input:           g.input,
			},
			&metrics.Metadata{
				Language:      metrics.Golang,
				Format:        metrics.PPROF,
				Profiler:      metrics.GoPProf,
				Start:         metrics.NewRFC3339Time(pData.startTime),
				End:           metrics.NewRFC3339Time(pData.endTime),
				Attachments:   deltaFileNames,
				TagsProfiler:  metrics.JoinTags(allTags),
				SubCustomTags: metrics.JoinTags(customTags),
			},
			g.input.GetBodySizeLimit(),
		); err != nil {
			log.Warnf("push delta profile data error: %s", err.Error())
			if firstErr == nil {
				firstErr = err
			}
		}
	}

	return firstErr
}

func copyTags(tags map[string]string) map[string]string {
	result := make(map[string]string, len(tags))
	for k, v := range tags {
		result[k] = v
	}
	return result
}

func (g *GoProfiler) pullProfileItem(
	ctx context.Context,
	profileType string,
	item Item,
	duration time.Duration,
) (*profileData, error) {
	params := cloneURLValues(item.params)
	if profileType == "cpu" {
		seconds := int(duration.Round(time.Second) / time.Second)
		if seconds < 1 {
			seconds = 1
		}
		params.Set("seconds", strconv.Itoa(seconds))
	}

	startTime := time.Now()
	buf, err := g.pullProfileData(ctx, item.path, params)
	if err != nil {
		return nil, fmt.Errorf("pull profile data error: %w", err)
	}
	endTime := time.Now()
	var current *pprofile.Profile
	var decoded []byte
	if g.validateResponse {
		current, decoded, err = parseBoundedProfile(buf.Bytes(), g.input.GetBodySizeLimit())
		if err != nil {
			return nil, err
		}
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	if len(item.deltaValues) > 0 {
		curProf := current
		if curProf == nil {
			curProf, err = pprofile.ParseData(buf.Bytes())
			if err != nil {
				return nil, fmt.Errorf("parse prof error:%w", err)
			}
		}

		var prevProf *pprofile.Profile
		if g.deltaCache != nil {
			previous := g.deltaCache.exchange(g, profileType, decoded)
			if previous != nil {
				prevProf, err = pprofile.ParseUncompressed(previous)
				if err != nil {
					return nil, fmt.Errorf("invalid cached profile")
				}
			}
		} else {
			prevProf = g.deltas[profileType]
			g.deltas[profileType] = curProf
		}

		// ignore first profile
		if prevProf == nil {
			return nil, nil
		}

		// calculate delta
		deltaProfile, err := getDeltaProfile(prevProf, curProf, item.deltaValues)
		if err != nil {
			return nil, fmt.Errorf("diff profile error: %w", err)
		}
		deltaProfile.TimeNanos = curProf.TimeNanos
		deltaProfile.DurationNanos = curProf.TimeNanos - prevProf.TimeNanos
		deltaBuf := &bytes.Buffer{}
		if err := deltaProfile.Write(deltaBuf); err != nil {
			return nil, fmt.Errorf("write delta profile failed: %w", err)
		}

		return &profileData{
			fileName:  item.fileName,
			buf:       deltaBuf,
			startTime: time.UnixMicro(prevProf.TimeNanos / 1000),
			endTime:   time.UnixMicro(curProf.TimeNanos / 1000),
		}, nil
	}

	return &profileData{
		fileName:  item.fileName,
		buf:       buf,
		startTime: startTime,
		endTime:   endTime,
	}, nil
}

func cloneURLValues(values url.Values) url.Values {
	result := make(url.Values, len(values))
	for k, values := range values {
		result[k] = append([]string(nil), values...)
	}
	return result
}

func (g *GoProfiler) pullProfileData(ctx context.Context, profilePath string, params url.Values) (*bytes.Buffer, error) {
	u := *g.url
	u.Path = path.Join(g.url.Path, profilePath)
	u.RawPath = ""

	if params != nil {
		u.RawQuery = params.Encode()
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
	if err != nil {
		return nil, &profileRequestError{endpoint: safeProfileURL(&u), cause: err}
	}

	if g.client == nil {
		return nil, fmt.Errorf("http client should be initialized")
	}

	resp, err := g.client.Do(req)
	if err != nil {
		return nil, &profileRequestError{endpoint: safeProfileURL(&u), cause: err}
	}

	defer resp.Body.Close() //nolint:errcheck

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("invalid response status: %d(%s)", resp.StatusCode, safeProfileURL(&u))
	}

	dst := new(bytes.Buffer)
	limit := g.input.GetBodySizeLimit()
	n, err := io.Copy(dst, io.LimitReader(resp.Body, limit+1))
	if err != nil {
		return nil, &profileRequestError{endpoint: safeProfileURL(&u), cause: err}
	}

	if n > limit {
		return nil, fmt.Errorf("exceed body max size")
	}

	return dst, nil
}

func (g *GoProfiler) createHTTPClient() (*http.Client, error) {
	client := &http.Client{Timeout: g.timeout}

	if g.TLSOpen {
		if g.InsecureSkipVerify {
			client.Transport = &http.Transport{
				TLSClientConfig: &tls.Config{InsecureSkipVerify: true}, // nolint:gosec
			}
		} else {
			tc, err := TLSConfig(g.CacertFile, g.CertFile, g.KeyFile)
			if err != nil {
				return nil, err
			} else {
				client.Transport = &http.Transport{
					TLSClientConfig: tc,
				}
			}
		}
	}

	if transport, ok := client.Transport.(*http.Transport); ok {
		transport.IdleConnTimeout = 90 * time.Second
		transport.MaxIdleConns = 16
		transport.TLSHandshakeTimeout = 10 * time.Second
	}
	return client, nil
}

func safeProfileURL(endpoint *url.URL) string {
	if endpoint == nil {
		return ""
	}
	return (&url.URL{Scheme: endpoint.Scheme, Host: endpoint.Host, Path: endpoint.Path}).String()
}

type profileRequestError struct {
	endpoint string
	cause    error
}

func (err *profileRequestError) Error() string {
	reason := "HTTP request failed"
	if errors.Is(err.cause, context.Canceled) {
		reason = "request canceled"
	} else if errors.Is(err.cause, context.DeadlineExceeded) {
		reason = "request timed out"
	}
	return fmt.Sprintf("%s: %s (%T)", err.endpoint, reason, err.cause)
}

func (err *profileRequestError) Unwrap() error {
	return err.cause
}

func TLSConfig(caFile, certFile, keyFile string) (*tls.Config, error) {
	// Load client cert
	cert, err := tls.LoadX509KeyPair(certFile, keyFile)
	if err != nil {
		return nil, err
	}
	// Load CA cert
	caCert, err := ioutil.ReadFile(filepath.Clean(caFile))
	if err != nil {
		return nil, err
	}
	caCertPool := x509.NewCertPool()
	if !caCertPool.AppendCertsFromPEM(caCert) {
		return nil, errors.New("failed to append certs from PEM")
	}

	tlsConfig := &tls.Config{ //nolint:gosec
		Certificates: []tls.Certificate{cert},
		RootCAs:      caCertPool,
		MinVersion:   tls.VersionTLS10,
	}

	return tlsConfig, nil
}

// getDeltaProfile computes the delta profile
// refer: https://github.com/DataDog/dd-trace-go/blob/153cabf0e3df707a9c779b43d3aec13fd79066b0/profiler/internal/pprofutils/delta.go#L32
func getDeltaProfile(a, b *pprofile.Profile, sampleTypes []valueType) (*pprofile.Profile, error) {
	ratios := make([]float64, len(a.SampleType))

	found := 0
	for i, st := range a.SampleType {
		for _, deltaSt := range sampleTypes {
			if deltaSt.Type == st.Type && deltaSt.Unit == st.Unit {
				ratios[i] = -1
				found++
			}
		}
	}
	if found != len(sampleTypes) {
		return nil, errors.New("one or more sample type(s) was not found in the profile")
	}

	if err := a.ScaleN(ratios); err != nil {
		return nil, fmt.Errorf("failed scaling profile a %w", err)
	}

	delta, err := pprofile.Merge([]*pprofile.Profile{a, b})
	if err != nil {
		return nil, err
	}
	return delta, delta.CheckValid()
}
