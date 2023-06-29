// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

// Package chrony collects chrony metrics.
package chrony

import (
	"fmt"
	"net"
	"net/netip"
	"net/url"
	"os"
	"strings"
	"sync"
	T "testing"
	"time"

	"github.com/BurntSushi/toml"
	"github.com/GuanceCloud/cliutils/point"
	dt "github.com/ory/dockertest/v3"
	"github.com/ory/dockertest/v3/docker"
	"github.com/stretchr/testify/assert"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs"
	tu "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/testutils"
)

type caseSpec struct {
	t *T.T

	name        string
	repo        string // docker name
	repoTag     string // docker tag
	envs        []string
	servicePort string // port (rand)）

	opts []inputs.PointCheckOption

	ipt    *Input // This is real prom
	feeder *io.MockedFeeder

	pool     *dt.Pool
	resource *dt.Resource

	cr *tu.CaseResult // collect `go test -run` metric
}

func (cs *caseSpec) checkPoint(pts []*point.Point) error {
	for _, pt := range pts {
		measurement := string(pt.Name())

		switch measurement {
		case "chrony":
			msgs := inputs.CheckPoint(pt, inputs.WithDoc(&ChronyMeasurement{}), inputs.WithExtraTags(cs.ipt.Tags))

			for _, msg := range msgs {
				cs.t.Logf("check measurement %s failed: %+#v", measurement, msg)
			}

			if len(msgs) > 0 {
				return fmt.Errorf("check measurement %s failed: %+#v", measurement, msgs)
			}
		default: // TODO: check other measurement
			return nil
		}

		// check if tag appended
		if len(cs.ipt.Tags) != 0 {
			cs.t.Logf("checking tags %+#v...", cs.ipt.Tags)

			tags := pt.Tags()
			for k, expect := range cs.ipt.Tags {
				if v := tags.Get([]byte(k)); v != nil {
					got := string(v.GetD())
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

	return nil
}

func (cs *caseSpec) run() error {
	// start remote image server
	r := tu.GetRemote()
	dockerTCP := r.TCPURL() // got "tcp://" + net.JoinHostPort(i.Host, i.Port)

	cs.t.Logf("get remote: %+#v, TCP: %s", r, dockerTCP)

	start := time.Now()

	p, err := dt.NewPool(dockerTCP)
	if err != nil {
		return err
	}

	hostname, err := os.Hostname()
	if err != nil {
		cs.t.Logf("get hostname failed: %s, ignored", err)
		hostname = "unknown-hostname"
	}

	containerName := fmt.Sprintf("%s.%s", hostname, cs.name)

	// remove the container if exist.
	if err := p.RemoveContainerByName(containerName); err != nil {
		return err
	}

	resource, err := p.RunWithOptions(&dt.RunOptions{
		// specify container image & tag
		Repository: cs.repo,
		Tag:        cs.repoTag,

		// port binding
		PortBindings: map[docker.Port][]docker.PortBinding{
			"22/tcp": {{HostIP: "0.0.0.0", HostPort: cs.servicePort}},
		},

		Name: containerName,

		// container run-time envs
		Env: cs.envs,
	}, func(c *docker.HostConfig) {
		c.RestartPolicy = docker.RestartPolicy{Name: "no"}
	})
	if err != nil {
		return err
	}

	cs.pool = p
	cs.resource = resource

	cs.t.Logf("check service(%s:%s)...", r.Host, cs.servicePort)
	if !r.PortOK(cs.servicePort, time.Minute) {
		return fmt.Errorf("service checking failed")
	}

	cs.cr.AddField("container_ready_cost", int64(time.Since(start)))

	var wg sync.WaitGroup

	// start input
	cs.t.Logf("start input...")
	wg.Add(1)
	go func() {
		defer wg.Done()
		cs.ipt.Run()
	}()

	// wait data
	start = time.Now()
	cs.t.Logf("wait points...")
	pts, err := cs.feeder.AnyPoints()
	if err != nil {
		return err
	}

	cs.cr.AddField("point_latency", int64(time.Since(start)))
	cs.cr.AddField("point_count", len(pts))

	cs.t.Logf("get %d points", len(pts))
	if err := cs.checkPoint(pts); err != nil {
		return err
	}

	cs.t.Logf("stop input...")
	cs.ipt.Terminate()

	cs.t.Logf("exit...")
	wg.Wait()

	return nil
}

func buildCases(t *T.T) ([]*caseSpec, error) {
	t.Helper()

	remote := tu.GetRemote()

	bases := []struct {
		name    string
		repo    string // docker name
		repoTag string // docker tag
		conf    string
		opts    []inputs.PointCheckOption
	}{
		{
			name:    "remote-chrony",
			repo:    "pubrepo.jiagouyun.com/image-repo-for-testing/chrony/chrony",
			repoTag: "ssh",
			conf: fmt.Sprintf(`
interval = "10s"
bin_path = "chronyc"
remote_addrs = ["http://%s"]
remote_users = ["root"]
remote_passwords = ["pw123"]
remote_command = "chronyc -n tracking"
election = false
[tags]
  tag1 = "some_value"
  tag2 = "some_other_value"`, net.JoinHostPort(remote.Host, fmt.Sprintf("%d", tu.RandPort("tcp")))),
			opts: []inputs.PointCheckOption{
				inputs.WithTypeChecking(false),
				inputs.WithExtraTags(map[string]string{"instance": ""}),
				inputs.WithOptionalTags(),
				inputs.WithOptionalFields(), // nolint:lll
			},
		},
	}

	// TODO: add per-image configs
	perImageCfgs := []interface{}{}
	_ = perImageCfgs

	var cases []*caseSpec

	// compose cases
	for _, base := range bases {
		feeder := io.NewMockedFeeder()

		ipt := newDefaultInput()
		ipt.feeder = feeder

		// URL from ENV.
		_, err := toml.Decode(base.conf, ipt)
		assert.NoError(t, err)

		url, err := url.Parse(ipt.RemoteAddrs[0])
		assert.NoError(t, err)

		ipt.RemoteAddrs[0] = strings.TrimPrefix(ipt.RemoteAddrs[0], "http://")

		ipport, err := netip.ParseAddrPort(url.Host)
		assert.NoError(t, err, "parse %s failed: %s", ipt.RemoteAddrs[0], err)

		cases = append(cases, &caseSpec{
			t:      t,
			ipt:    ipt,
			name:   base.name,
			feeder: feeder,
			// envs:   envs,

			repo:    base.repo,    // docker name
			repoTag: base.repoTag, // docker tag

			servicePort: fmt.Sprintf("%d", ipport.Port()),

			opts: base.opts,

			// Test case result.
			cr: &tu.CaseResult{
				Name:        t.Name(),
				Case:        base.name,
				ExtraFields: map[string]any{},
				ExtraTags: map[string]string{
					"image":         base.repo,
					"image_tag":     base.repoTag,
					"remote_server": ipt.RemoteAddrs[0],
				},
			},
		})
	}
	return cases, nil
}

func TestInput(t *T.T) {
	if !tu.CheckIntegrationTestingRunning() {
		t.Skip()
	}
	start := time.Now()
	cases, err := buildCases(t)
	if err != nil {
		cr := &tu.CaseResult{
			Name:          t.Name(),
			Status:        tu.TestPassed,
			FailedMessage: err.Error(),
			Cost:          time.Since(start),
		}

		_ = tu.Flush(cr)
		return
	}

	t.Logf("testing %d cases...", len(cases))

	for _, tc := range cases {
		t.Run(tc.name, func(t *T.T) {
			caseStart := time.Now()

			t.Logf("testing %s...", tc.name)

			// Run a test case.
			if err := tc.run(); err != nil {
				tc.cr.Status = tu.TestFailed
				tc.cr.FailedMessage = err.Error()

				assert.NoError(t, err)
			} else {
				tc.cr.Status = tu.TestPassed
			}

			tc.cr.Cost = time.Since(caseStart)

			assert.NoError(t, tu.Flush(tc.cr))

			t.Cleanup(func() {
				// clean remote docker resources
				if tc.resource == nil {
					return
				}

				assert.NoError(t, tc.pool.Purge(tc.resource))
			})
		})
	}
}
