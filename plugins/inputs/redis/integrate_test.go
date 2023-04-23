// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package redis

import (
	"fmt"
	"os"
	"strconv"
	"sync"
	"testing"
	"time"

	"github.com/BurntSushi/toml"
	"github.com/GuanceCloud/cliutils/point"
	dt "github.com/ory/dockertest/v3"
	"github.com/ory/dockertest/v3/docker"
	"github.com/stretchr/testify/assert"
	tu "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/testutils"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
)

type caseSpec struct {
	t *testing.T

	name        string
	repo        string
	repoTag     string
	envs        []string
	servicePort string

	ipt    *Input
	feeder *io.MockedFeeder

	pool     *dt.Pool
	resource *dt.Resource

	cr *tu.CaseResult
}

func buildCases(t *testing.T) ([]*caseSpec, error) {
	t.Helper()

	remote := tu.GetRemote()

	bases := []struct {
		name string
		conf string
	}{
		{
			name: "remote-redis",

			conf: fmt.Sprintf(`
host = "%s"
port = %s`,
				remote.Host,
				fmt.Sprintf("%d", tu.RandPort("tcp"))),
		},
	}

	images := [][2]string{
		{"redis", "7.0"},
		{"redis", "6.0"},
		{"redis", "5.0"},
	}

	// TODO: add per-image configs
	perImageCfgs := []interface{}{}
	_ = perImageCfgs

	var cases []*caseSpec

	// compose cases
	for _, img := range images {
		for _, base := range bases {
			feeder := io.NewMockedFeeder()

			ipt := defaultInput()
			ipt.Service = "some_service"
			ipt.feeder = feeder

			_, err := toml.Decode(base.conf, ipt)
			assert.NoError(t, err)

			envs := []string{}

			// ipport, err := netip.ParseAddrPort(ipt.Host)
			assert.NoError(t, err, "parse %s failed: %s", ipt.Host, err)

			cases = append(cases, &caseSpec{
				t:      t,
				ipt:    ipt,
				name:   base.name,
				feeder: feeder,
				envs:   envs,

				repo:    img[0],
				repoTag: img[1],

				servicePort: fmt.Sprintf("%d", ipt.Port),

				cr: &tu.CaseResult{
					Name:        t.Name(),
					Case:        base.name,
					ExtraFields: map[string]any{},
					ExtraTags: map[string]string{
						"image":         img[0],
						"image_tag":     img[1],
						"remote_server": ipt.Host,
					},
				},
			})
			l.Infof("using port: %s\n", cases[len(cases)-1].servicePort)
		}
	}
	return cases, nil
}

func (cs *caseSpec) checkPoint(pts []*point.Point) error {
	for _, pt := range pts {
		measurement := string(pt.Name())

		switch measurement {
		case "redis_latency":
			msgs := inputs.CheckPoint(pt, inputs.WithDoc(&latencyMeasurement{}), inputs.WithExtraTags(cs.ipt.Tags))

			for _, msg := range msgs {
				cs.t.Logf("check measurement %s failed: %+#v", measurement, msg)
			}

		case "redis_slowlog":
			msgs := inputs.CheckPoint(pt, inputs.WithDoc(&slowlogMeasurement{}), inputs.WithExtraTags(cs.ipt.Tags))

			for _, msg := range msgs {
				cs.t.Logf("check measurement %s failed: %+#v", measurement, msg)
			}

		case "redis_bigkey":
			msgs := inputs.CheckPoint(pt, inputs.WithDoc(&bigKeyMeasurement{}), inputs.WithExtraTags(cs.ipt.Tags))

			for _, msg := range msgs {
				cs.t.Logf("check measurement %s failed: %+#v", measurement, msg)
			}
		case "redis_client":
			msgs := inputs.CheckPoint(pt, inputs.WithDoc(&clientMeasurement{}), inputs.WithExtraTags(cs.ipt.Tags))

			for _, msg := range msgs {
				cs.t.Logf("check measurement %s failed: %+#v", measurement, msg)
			}
		case "redis_cluster":
			msgs := inputs.CheckPoint(pt, inputs.WithDoc(&clusterMeasurement{}), inputs.WithExtraTags(cs.ipt.Tags))

			for _, msg := range msgs {
				cs.t.Logf("check measurement %s failed: %+#v", measurement, msg)
			}
		case "redis_command_stat":
			msgs := inputs.CheckPoint(pt, inputs.WithDoc(&commandMeasurement{}), inputs.WithExtraTags(cs.ipt.Tags))

			for _, msg := range msgs {
				cs.t.Logf("check measurement %s failed: %+#v", measurement, msg)
			}
		case "redis_db":
			msgs := inputs.CheckPoint(pt, inputs.WithDoc(&dbMeasurement{}), inputs.WithExtraTags(cs.ipt.Tags))

			for _, msg := range msgs {
				cs.t.Logf("check measurement %s failed: %+#v", measurement, msg)
			}
		// Some metrics in redis_info are only present on replica
		case "redis_info":
			optionalFields := []string{
				"loading_eta_seconds", "loading_loaded_perc", "master_last_io_seconds_ago", "client_longest_output_list",
				"master_sync_left_bytes", "used_cpu_sys_percent", "aof_current_size", "loading_loaded_bytes", "loading_total_bytes", "master_sync_in_progress",
				"slave_repl_offset", "aof_buffer_length", "used_cpu_user_percent", "client_biggest_input_buf",
			}
			// optionalTags := []string{"host"}
			extra := map[string]string{"host": "some_host", "service_name": "some_service"}

			msgs := inputs.CheckPoint(pt, inputs.WithDoc(&infoMeasurement{}), inputs.WithExtraTags(extra), inputs.WithOptionalFields(optionalFields...))

			for _, msg := range msgs {
				cs.t.Logf("check measurement %s failed: %+#v", measurement, msg)
			}
		default:
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
	// start remote sqlserver
	r := tu.GetRemote()
	dockerTCP := r.TCPURL()

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

	// remove container if exist.
	if err := p.RemoveContainerByName(containerName); err != nil {
		return err
	}

	resource, err := p.RunWithOptions(&dt.RunOptions{
		// specify container image & tag
		Repository: cs.repo,
		Tag:        cs.repoTag,

		// port binding
		PortBindings: map[docker.Port][]docker.PortBinding{
			"1433/tcp": {{HostIP: "0.0.0.0", HostPort: cs.servicePort}},
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

	// set input's port to port exposed by docker
	cs.resource = resource
	exposedPort := resource.GetPort("6379/tcp")
	cs.ipt.Port, _ = strconv.Atoi(exposedPort)

	cs.t.Logf("check service(%s:%s)...", r.Host, exposedPort)
	if !r.PortOK(exposedPort, time.Minute) {
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

func TestRedisInput(t *testing.T) {
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
		t.Run(tc.name, func(t *testing.T) {
			caseStart := time.Now()

			t.Logf("testing %s...", tc.name)

			if err := tu.RetryTestRun(tc.run); err != nil {
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
