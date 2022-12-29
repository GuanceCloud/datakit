// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package profile

import (
	"bytes"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"net/textproto"
	"net/url"
	"path/filepath"
	"strconv"
	"time"

	pprofile "github.com/google/pprof/profile"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	dkio "gitlab.jiagouyun.com/cloudcare-tools/datakit/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io/point"
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
	fileName string
	buf      *bytes.Buffer
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
	profileDatas := []*profileData{}
	startTime := time.Now()
	for _, k := range g.EnabledTypes {
		if p, ok := profileConfigMap[k]; ok {
			if profileData, err := g.pullProfileItem(k, p); err != nil {
				log.Warnf("profile for %s error: %s", k, err.Error())
			} else if profileData != nil {
				profileDatas = append(profileDatas, profileData)
			}
		} else {
			log.Warnf("invalid profile type: %s", k)
		}
	}
	endTime := time.Now()

	if len(profileDatas) > 0 {
		if err := g.pushProfileData(startTime, endTime, profileDatas); err != nil {
			log.Warnf("push profile data error: %s", err.Error())
		}
	}
}

func (g *GoProfiler) pullProfileItem(profileType string, item ProfileItem) (*profileData, error) {
	buf, err := g.pullProfileData(item.path, item.params)
	if err != nil {
		return nil, fmt.Errorf("pull profile data error: %w", err)
	}

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
			fileName: item.fileName,
			buf:      deltaBuf,
		}, nil
	}

	return &profileData{
		fileName: item.fileName,
		buf:      buf,
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

func (g *GoProfiler) pushProfileData(startTime, endTime time.Time, profiledatas []*profileData) error {
	b := new(bytes.Buffer)
	mw := multipart.NewWriter(b)

	for _, profileData := range profiledatas {
		if ff, err := mw.CreateFormFile(profileData.fileName, profileData.fileName); err != nil {
			continue
		} else {
			if _, err = io.Copy(ff, profileData.buf); err != nil {
				continue
			}
		}
	}

	h := make(textproto.MIMEHeader)
	h.Set("Content-Disposition", "form-data; name=\"event\"; filename=\"event.json\"")
	h.Set("Content-Type", "application/json")
	f, err := mw.CreatePart(h)
	if err != nil {
		return err
	}
	eventJSONString := `
		{
			"family": "golang",
			"format": "pprof"
	  }
	`
	if _, err := io.Copy(f, bytes.NewReader([]byte(eventJSONString))); err != nil {
		return err
	}
	if err := mw.Close(); err != nil {
		return err
	}

	profileID := randomProfileID()

	URL, err := profilingProxyURL()
	if err != nil {
		return err
	}

	req, err := http.NewRequest("POST", URL.String(), b)
	if err != nil {
		return err
	}

	wsID, err := queryWorkSpaceUUID()
	if err != nil {
		log.Errorf("query workspace id fail: %s", err)
	}
	req.Header.Set("Content-Type", mw.FormDataContentType())
	req.Header.Set("X-Datakit-Workspace", wsID)
	req.Header.Set("X-Datakit-Profileid", profileID)
	req.Header.Set("X-Datakit-Unixnano", strconv.FormatInt(startTime.UnixNano(), 10))

	client := &http.Client{Timeout: 15 * time.Second}

	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close() //nolint:errcheck

	bo, _ := ioutil.ReadAll(resp.Body)
	if resp.StatusCode == http.StatusOK {
		var resp uploadResponse

		if err := json.Unmarshal(bo, &resp); err != nil {
			return fmt.Errorf("json unmarshal upload profile binary response err: %w", err)
		}

		if resp.Content == nil || resp.Content.ProfileID == "" {
			return fmt.Errorf("fetch profile upload response profileID fail")
		}

		if err := g.writeProfilePoint(profileID, startTime, endTime); err != nil {
			return fmt.Errorf("write profile point failed: %w", err)
		}
	} else {
		return fmt.Errorf("push profile data failed, response status: %s", resp.Status)
	}
	return nil
}

func (g *GoProfiler) writeProfilePoint(profileID string, startTime, endTime time.Time) error {
	pointTags := map[string]string{
		TagEndPoint: g.url.String(),
		TagLanguage: "golang",
	}

	// extend custom tags
	for k, v := range g.tags {
		if _, ok := pointTags[k]; !ok {
			pointTags[k] = v
		}
	}

	pointFields := map[string]interface{}{
		FieldProfileID:  profileID,
		FieldFormat:     "pprof",
		FieldDatakitVer: datakit.Version,
		FieldStart:      startTime.UnixNano(),
		FieldEnd:        endTime.UnixNano(),
		FieldDuration:   endTime.Sub(startTime).Nanoseconds(),
	}

	pt, err := point.NewPoint(inputName, pointTags, pointFields, &point.PointOption{
		Time:               startTime,
		Category:           datakit.Profiling,
		Strict:             false,
		GlobalElectionTags: g.input.Election,
	})
	if err != nil {
		return fmt.Errorf("build profile point fail: %w", err)
	}

	if err := dkio.Feed(inputName,
		datakit.Profiling,
		[]*point.Point{pt},
		&dkio.Option{CollectCost: time.Since(pt.Time())}); err != nil {
		return err
	}

	return nil
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
