// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package profile

import (
	"bytes"
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"path/filepath"
	"time"

	pprofile "github.com/google/pprof/profile"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/datakit"
)

const (
	goReportFamily = "golang"
	goReportFormat = "pprof"
)

// GoProfiler pull go pprof data.
type GoProfiler struct {
	URL          string            `toml:"url"`
	Interval     string            `toml:"interval"`
	Service      string            `toml:"service"`
	Env          string            `toml:"env"`
	Version      string            `toml:"version"`
	Tags         map[string]string `toml:"tags"`
	EnabledTypes []string          `toml:"enabled_types"` // cpu,goroutine,heap,mutex,block

	TLSOpen            bool   `toml:"tls_open"`
	CacertFile         string `toml:"tls_ca"`
	CertFile           string `toml:"tls_cert"`
	KeyFile            string `toml:"tls_key"`
	InsecureSkipVerify bool   `toml:"insecure_skip_verify"`

	url      *url.URL
	interval time.Duration
	tags     map[string]string
	client   *http.Client
	deltas   map[string]*pprofile.Profile
	input    *Input
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

type ProfileItem struct {
	path        string
	params      url.Values
	fileName    string
	deltaValues []valueType
}

var profileConfigMap = map[string]ProfileItem{
	"cpu": {
		path: "/debug/pprof/profile",
		params: url.Values{
			"seconds": []string{"10"},
		},
		fileName: "cpu.pprof",
	},
	"goroutine": {
		path:     "/debug/pprof/goroutine",
		fileName: "goroutines.pprof",
	},
	"heap": {
		path:     "/debug/pprof/heap",
		fileName: "delta-heap.pprof",
		deltaValues: []valueType{
			{Type: "alloc_objects", Unit: "count"},
			{Type: "alloc_space", Unit: "bytes"},
		},
	},
	"mutex": {
		path:     "/debug/pprof/mutex",
		fileName: "delta-mutex.pprof",
		deltaValues: []valueType{
			{Type: "contentions", Unit: "count"},
			{Type: "delay", Unit: "nanoseconds"},
		},
	},
	"block": {
		path:     "/debug/pprof/block",
		fileName: "delta-block.pprof",
		deltaValues: []valueType{
			{Type: "contentions", Unit: "count"},
			{Type: "delay", Unit: "nanoseconds"},
		},
	},
}

// init check config and set config.
func (g *GoProfiler) init() error {
	duration, err := time.ParseDuration(g.Interval)
	if err != nil {
		return err
	}
	// duration should be larger than 10s
	if duration < 10*time.Second {
		duration = 10 * time.Second
	}
	g.interval = duration

	// url parse
	g.url, err = url.Parse(g.URL)
	if err != nil {
		return err
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
	if i == nil {
		return fmt.Errorf("input expected not to be nil")
	}
	g.input = i

	if err := g.init(); err != nil {
		return fmt.Errorf("init go profiler error: %w", err)
	}

	tick := time.NewTicker(g.interval)
	defer tick.Stop()

	for {
		if i.pause {
			log.Debugf("not leader, skipped")
		} else {
			g.pullProfile()
		}

		select {
		case <-datakit.Exit.Wait():
			return nil
		case <-i.semStop.Wait():
			log.Info("go profiler exit")
			return nil
		case <-tick.C:

		case i.pause = <-i.pauseCh:
		}
	}
}

func (g *GoProfiler) pullProfile() {
	deletaDatas := []*profileData{}
	for _, k := range g.EnabledTypes {
		if p, ok := profileConfigMap[k]; ok {
			if pData, err := g.pullProfileItem(k, p); err != nil {
				log.Warnf("profile for %s error: %s", k, err.Error())
			} else if pData != nil {
				if p.deltaValues != nil {
					deletaDatas = append(deletaDatas, pData)
				} else if err := pushProfileData(
					&pushProfileDataOpt{
						startTime:       pData.startTime,
						endTime:         pData.endTime,
						profiledatas:    []*profileData{pData},
						reportFamily:    goReportFamily,
						reportFormat:    goReportFormat,
						endPoint:        g.url.String(),
						inputTags:       g.tags,
						election:        g.input.Election,
						inputNameSuffix: "/go",
					},
				); err != nil {
					log.Warnf("push profile data error: %s", err.Error())
				}
			}
		} else {
			log.Warnf("invalid profile type: %s", k)
		}
	}

	// push delta profiles together
	if len(deletaDatas) > 0 {
		pData := deletaDatas[0]
		if err := pushProfileData(
			&pushProfileDataOpt{
				startTime:       pData.startTime,
				endTime:         pData.endTime,
				profiledatas:    deletaDatas,
				reportFamily:    goReportFamily,
				reportFormat:    goReportFormat,
				endPoint:        g.url.String(),
				inputTags:       g.tags,
				election:        g.input.Election,
				inputNameSuffix: "/go",
			},
		); err != nil {
			log.Warnf("push delta profile data error: %s", err.Error())
		}
	}
}

func (g *GoProfiler) pullProfileItem(profileType string, item ProfileItem) (*profileData, error) {
	startTime := time.Now()
	buf, err := g.pullProfileData(item.path, item.params)
	if err != nil {
		return nil, fmt.Errorf("pull profile data error: %w", err)
	}
	endTime := time.Now()

	if len(item.deltaValues) > 0 {
		curProf, err := pprofile.ParseData(buf.Bytes())
		if err != nil {
			return nil, fmt.Errorf("parse prof error:%w", err)
		}

		prevProf, ok := g.deltas[profileType]
		g.deltas[profileType] = curProf

		// ignore first profile
		if !ok {
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

func (g *GoProfiler) pullProfileData(path string, params url.Values) (*bytes.Buffer, error) {
	u := url.URL{
		Path:   path,
		Scheme: g.url.Scheme,
		Host:   g.url.Host,
	}

	if params != nil {
		u.RawQuery = params.Encode()
	}

	req, err := http.NewRequest("GET", u.String(), nil)
	if err != nil {
		return nil, err
	}

	if g.client == nil {
		return nil, fmt.Errorf("http client should be initialized")
	}

	resp, err := g.client.Do(req)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close() //nolint:errcheck

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("invalid response status: %s(%s)", resp.Status, u.String())
	}

	dst := new(bytes.Buffer)
	n, err := io.Copy(dst, LimitReader(resp.Body, profileMaxSize))
	if err != nil {
		return nil, err
	}

	if n >= profileMaxSize {
		return nil, fmt.Errorf("exceed body max size")
	}

	return dst, nil
}

func (g *GoProfiler) createHTTPClient() (*http.Client, error) {
	timeout := 15 * time.Second
	client := &http.Client{Timeout: timeout}

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

	return client, nil
}

type Reader interface {
	Read(p []byte) (n int, err error)
}

var ErrEOF = errors.New("EOF")

func LimitReader(r Reader, n int64) Reader { return &LimitedReader{r, n} }

// A LimitedReader reads from R but limits the amount of
// data returned to just N bytes. Each call to Read
// updates N to reflect the new amount remaining.
// Read returns EOF when N <= 0 or when the underlying R returns EOF.
type LimitedReader struct {
	R Reader // underlying reader
	N int64  // max bytes remaining
}

func (l *LimitedReader) Read(p []byte) (n int, err error) {
	if l.N <= 0 {
		return 0, ErrEOF
	}
	if int64(len(p)) > l.N {
		p = p[0:l.N]
	}
	n, err = l.R.Read(p)
	l.N -= int64(n)
	return
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
