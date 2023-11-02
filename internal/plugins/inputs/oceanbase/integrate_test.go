// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package oceanbase

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/BurntSushi/toml"
	lp "github.com/GuanceCloud/cliutils/lineproto"
	"github.com/GuanceCloud/cliutils/point"
	"github.com/gin-gonic/gin"
	influxdb "github.com/influxdata/influxdb1-client/v2"
	"github.com/ory/dockertest/v3"
	"github.com/stretchr/testify/require"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/datakit"
	dkio "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/testutils"
)

// ATTENTION: Docker version should use v20.10.18 in integrate tests. Other versions are not tested.

func TestIntegrate(t *testing.T) {
	if !testutils.CheckIntegrationTestingRunning() {
		t.Skip()
	}

	obLsnPort := os.Getenv("OCEANBASE_LISTEN_PORT")
	if len(obLsnPort) == 0 {
		t.Skip()
	}

	start := time.Now()
	cases, err := buildCases(t)
	if err != nil {
		cr := &testutils.CaseResult{
			Name:          t.Name(),
			Status:        testutils.TestPassed,
			FailedMessage: err.Error(),
			Cost:          time.Since(start),
		}

		_ = testutils.Flush(cr)
		return
	}

	t.Logf("testing %d cases...", len(cases))

	for _, tc := range cases {
		func(tc *caseSpec) {
			t.Run(tc.name, func(t *testing.T) {
				// t.Parallel() // Should not be parallel, if so, it would dead and timeout due to junk machine.
				caseStart := time.Now()
				tc.listenPort = obLsnPort

				t.Logf("testing %s...", tc.name)

				if err := testutils.RetryTestRun(tc.run); err != nil {
					tc.cr.Status = testutils.TestFailed
					tc.cr.FailedMessage = err.Error()

					panic(err)
				} else {
					tc.cr.Status = testutils.TestPassed
				}

				tc.cr.Cost = time.Since(caseStart)

				require.NoError(t, testutils.Flush(tc.cr))

				t.Cleanup(func() {
					// clean remote docker resources
					if tc.resource == nil {
						return
					}

					tc.pool.Purge(tc.resource)
				})
			})
		}(tc)
	}
}

func buildCases(t *testing.T) ([]*caseSpec, error) {
	t.Helper()

	remote := testutils.GetRemote()

	bases := []struct {
		name        string // Also used as build image name:tag.
		conf        string
		optsMetric  []inputs.PointCheckOption
		optsLogging []inputs.PointCheckOption
	}{
		{
			name: "normal",
			optsMetric: []inputs.PointCheckOption{
				inputs.WithOptionalTags("oceanbase_service"),
			},
		},
	}

	var cases []*caseSpec

	// compose cases
	for _, base := range bases {
		feeder := dkio.NewMockedFeeder()

		ipt := defaultInput()
		// ipt.feeder = feeder // no need.

		_, err := toml.Decode(base.conf, ipt)
		require.NoError(t, err)

		cases = append(cases, &caseSpec{
			t:           t,
			ipt:         ipt,
			name:        base.name,
			feeder:      feeder,
			optsMetric:  base.optsMetric,
			optsLogging: base.optsLogging,

			cr: &testutils.CaseResult{
				Name:        t.Name(),
				Case:        base.name,
				ExtraFields: map[string]any{},
				ExtraTags: map[string]string{
					"docker_host": remote.Host,
					"docker_port": remote.Port,
				},
			},
		})
	}

	return cases, nil
}

////////////////////////////////////////////////////////////////////////////////

// caseSpec.

type caseSpec struct {
	t *testing.T

	name        string
	optsMetric  []inputs.PointCheckOption
	optsLogging []inputs.PointCheckOption
	done        chan struct{}
	mCount      map[string]struct{}
	listenPort  string

	ipt    *Input
	feeder *dkio.MockedFeeder

	pool     *dockertest.Pool
	resource *dockertest.Resource

	cr *testutils.CaseResult
}

type FeedMeasurementBody []struct {
	Measurement string                 `json:"measurement"`
	Tags        map[string]string      `json:"tags"`
	Fields      map[string]interface{} `json:"fields"`
}

func (cs *caseSpec) handler(c *gin.Context) {
	uri, err := url.ParseRequestURI(c.Request.URL.RequestURI())
	if err != nil {
		cs.t.Logf("%s", err.Error())
		return
	}

	body, err := io.ReadAll(c.Request.Body)
	defer c.Request.Body.Close()
	if err != nil {
		cs.t.Logf("%s", err.Error())
		return
	}

	switch uri.Path {
	case datakit.Metric, datakit.Logging:
		pts, err := lp.ParsePoints(body, nil)
		if err != nil {
			cs.t.Logf("ParsePoints failed: %s", err.Error())
			return
		}

		newPts := dkpt2point(pts...)

		for _, pt := range newPts {
			fmt.Println(pt.LineProto())
		}

		if err := cs.checkPoint(newPts); err != nil {
			cs.t.Logf("%s", err.Error())
			require.NoError(cs.t, err)
			return
		}

	default:
		panic("unknown path: " + uri.Path)
	}
}

func (cs *caseSpec) lasterror(c *gin.Context) {
	uri, err := url.ParseRequestURI(c.Request.URL.RequestURI())
	if err != nil {
		cs.t.Logf("%s", err.Error())
		return
	}
	fmt.Println("uri ==>", uri)

	body, err := io.ReadAll(c.Request.Body)
	defer c.Request.Body.Close()
	if err != nil {
		cs.t.Logf("%s", err.Error())
		return
	}
	fmt.Println("lasterror ==>", string(body))
}

func (cs *caseSpec) checkPoint(pts []*point.Point) error {
	for _, pt := range pts {
		var opts []inputs.PointCheckOption
		opts = append(opts, inputs.WithExtraTags(cs.ipt.Tags))

		measurement := pt.Name()

		switch measurement {
		case metricName:
			opts = append(opts, cs.optsMetric...)
			opts = append(opts, inputs.WithDoc(&metricMeasurement{}))

			msgs := inputs.CheckPoint(pt, opts...)

			for _, msg := range msgs {
				cs.t.Logf("check measurement %s failed: %+#v", measurement, msg)
			}

			// TODO: error here
			if len(msgs) > 0 {
				return fmt.Errorf("check measurement %s failed: %+#v", measurement, msgs)
			}

			cs.t.Logf(metricName + " check completed!")
			cs.mCount[metricName] = struct{}{}

		case logName:
			opts = append(opts, cs.optsMetric...)
			opts = append(opts, inputs.WithDoc(&loggingMeasurement{}))

			msgs := inputs.CheckPoint(pt, opts...)

			for _, msg := range msgs {
				cs.t.Logf("check measurement %s failed: %+#v", measurement, msg)
			}

			// TODO: error here
			if len(msgs) > 0 {
				return fmt.Errorf("check measurement %s failed: %+#v", measurement, msgs)
			}

			cs.t.Logf(logName + " check completed!")
			cs.mCount[logName] = struct{}{}

		default: // TODO: check other measurement
			panic("unknown measurement: " + measurement)
		}

		// check if tag appended
		if len(cs.ipt.Tags) != 0 {
			cs.t.Logf("checking tags %+#v...", cs.ipt.Tags)

			tags := pt.Tags()
			for k, expect := range cs.ipt.Tags {
				if v := tags.Get(k); v != nil {
					got := v.GetS()
					if got != expect {
						return fmt.Errorf("expect tag value %s, got %s", expect, got)
					}
				} else {
					return fmt.Errorf("tag %s not found, got %v", k, tags)
				}
			}
		}
	}

	// TODO: some other checking on @pts, such as `if some required measurements exist'...
	if len(cs.mCount) == 2 {
		cs.done <- struct{}{}
	}

	return nil
}

func (cs *caseSpec) run() error {
	cs.t.Helper()

	r := testutils.GetRemote()
	dockerTCP := r.TCPURL()

	cs.t.Logf("get remote: %+#v, TCP: %s", r, dockerTCP)

	cs.done = make(chan struct{})
	gin.SetMode(gin.DebugMode)
	router := gin.Default()
	router.POST(datakit.Metric, cs.handler)
	router.POST(datakit.Logging, cs.handler)
	router.POST("/v1/lasterror", cs.lasterror)

	var (
		listener    net.Listener
		randPortStr string
		err         error
	)

	for {
		listener, err = net.Listen("tcp", ":"+cs.listenPort)
		if err != nil {
			if strings.Contains(err.Error(), "bind: address already in use") {
				continue
			}
			cs.t.Logf("net.Listen failed: %v", err)
			return err
		}
		break
	}

	cs.t.Logf("listening port " + randPortStr + "...")

	srv := &http.Server{Handler: router}

	go func() {
		if err := srv.Serve(listener); err != nil && err != http.ErrServerClosed {
			panic(err)
		}
	}()

	defer func() {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		if err := srv.Shutdown(ctx); err != nil && errors.Is(err, http.ErrServerClosed) {
			cs.t.Logf("Shutdown failed: %v", err)
		}
	}()

	cs.mCount = map[string]struct{}{}

	timeout := time.Minute
	cs.t.Logf("checking in %v...", timeout)
	tick := time.NewTicker(timeout)

	select {
	case <-tick.C:
		panic("check " + inputName + " timeout: " + cs.name)
	case <-cs.done:
		cs.t.Logf("check " + inputName + " all done!")
	}

	cs.t.Logf("exit...")

	return nil
}

// nolint: deadcode,unused
// dkpt2point convert old io/point.Point to point.Point.
func dkpt2point(pts ...*influxdb.Point) (res []*point.Point) {
	for _, pt := range pts {
		fs, err := pt.Fields()
		if err != nil {
			continue
		}

		pt := point.NewPointV2(pt.Name(),
			append(point.NewTags(pt.Tags()), point.NewKVs(fs)...), nil)

		res = append(res, pt)
	}

	return res
}
