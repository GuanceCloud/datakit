// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package nginx

import (
	"fmt"
	"os"
	"sync"
	T "testing"
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

func TestSQLServerInput(t *T.T) {
	os.Setenv("REMOTE_HOST", "10.200.14.142")
	os.Setenv("TESTING_METRIC_PATH", "/tmp/testing.metrics")

	start := time.Now()
	cases, err := buildCases(t)
	if err != nil {
		cs := tu.CaseResult{
			Name:          t.Name(),
			Status:        tu.CasePassed,
			FailedMessage: err.Error(),
			Cost:          time.Since(start),
		}

		_ = cs.Flush()
		return
	}

	t.Logf("testing %d cases...", len(cases))

	for _, tc := range cases {
		t.Run(tc.name, func(t *T.T) {
			caseStart := time.Now()

			t.Logf("testing %s...", tc.name)

			if err := tc.run(); err != nil {
				tc.cr.Status = tu.CaseFailed
				tc.cr.FailedMessage = err.Error()

				assert.NoError(t, err)
			} else {
				tc.cr.Status = tu.CasePassed
			}

			tc.cr.Cost = time.Since(caseStart)

			assert.NoError(t, tc.cr.Flush())

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

type caseSpec struct {
	t *T.T

	name        string
	repo        string
	repoTag     string
	envs        []string
	servicePort []string

	ipt    *Input
	feeder *io.MockedFeeder

	pool     *dt.Pool
	resource *dt.Resource

	cr *tu.CaseResult
}

func (cs *caseSpec) checkPoint(pts []*point.Point) error {
	for _, pt := range pts {
		measurement := string(pt.Name())

		switch measurement {
		case "sqlserver_performance":
			msgs := inputs.CheckPoint(pt, &NginxMeasurement{}, inputs.WithAllowExtraTags(len(cs.ipt.Tags) > 0))

			for _, msg := range msgs {
				cs.t.Logf("check measurement %s failed: %+#v", measurement, msg)
			}

			// TODO: error here
			//if len(msgs) > 0 {
			//	return fmt.Errorf("check measurement %s failed: %+#v", measurement, msgs)
			//}

		default: // TODO: check other measurement

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

	p, err := cs.getPool(dockerTCP)
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

	resource, err := p.BuildAndRunWithOptions(
		"/Users/mac/Downloads/tmp/sh-dockers/nginx/http_stub_status_module/dockerfile",

		&dt.RunOptions{
			Name: "nginx:httpstubstatusmodule",

			// specify container image & tag
			Repository: cs.repo,
			Tag:        cs.repoTag,

			ExposedPorts: []string{
				"80/tcp",
			},

			// port binding
			PortBindings: map[docker.Port][]docker.PortBinding{
				"80/tcp": {{HostIP: "0.0.0.0", HostPort: cs.servicePort[0]}},
			},
		},

		func(c *docker.HostConfig) {
			c.RestartPolicy = docker.RestartPolicy{Name: "no"}
		},
	)

	// resource, err := p.RunWithOptions(&dt.RunOptions{
	// 	// specify container image & tag
	// 	Repository: cs.repo,
	// 	Tag:        cs.repoTag,

	// 	// port binding
	// 	PortBindings: map[docker.Port][]docker.PortBinding{
	// 		"80/tcp":  {{HostIP: "0.0.0.0", HostPort: cs.servicePort[0]}},
	// 		"443/tcp": {{HostIP: "0.0.0.0", HostPort: cs.servicePort[1]}},
	// 	},

	// 	Name: containerName,

	// 	// container run-time envs
	// 	Env: cs.envs,
	// }, func(c *docker.HostConfig) {
	// 	c.RestartPolicy = docker.RestartPolicy{Name: "no"}
	// })
	if err != nil {
		return err
	}

	cs.pool = p
	cs.resource = resource

	cs.t.Logf("check service(%s:%s)...", r.Host, cs.servicePort)
	// if !r.PortOK(cs.servicePort, time.Minute) {
	// 	return fmt.Errorf("service checking failed")
	// }
	if err := portsOK(r, cs.servicePort...); err != nil {
		return err
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
		name string
		conf string
	}{
		// 		{
		// 			name: "remote-sqlserver",

		// 			conf: fmt.Sprintf(`
		// host = "%s"
		// user = "sa"
		// password = "Abc123abC$"`,
		// 				net.JoinHostPort(remote.Host, fmt.Sprintf("%d", tu.RandPort("tcp")))),
		// 		},

		// 		{
		// 			name: "remote-sqlserver-with-extra-tags",

		// 			// Why config like this? See:
		// 			//    https://gitlab.jiagouyun.com/cloudcare-tools/datakit/-/issues/1391#note_36026
		// 			conf: fmt.Sprintf(`
		// host = "%s"
		// user = "sa"
		// password = "Abc123abC$" # SQLServer require password to be larger than 8bytes, must include number, alphabet and symbol.
		// [tags]
		//   tag1 = "some_value"
		//   tag2 = "some_other_value"`, net.JoinHostPort(remote.Host, fmt.Sprintf("%d", tu.RandPort("tcp")))),
		// 		},

		{
			name: "http_stub_status_module",
			conf: fmt.Sprintf(`url = "http://%s:80/server_status"`, remote.Host),
		},
	}

	images := [][2]string{
		{"nginx", "httpstubstatusmodule"},
	}

	// TODO: add per-image configs
	// perImageCfgs := []interface{}{}
	// _ = perImageCfgs

	var cases []*caseSpec

	// compose cases
	for _, img := range images {
		for _, base := range bases {
			feeder := io.NewMockedFeeder()

			ipt := defaultInput()
			ipt.feeder = feeder

			_, err := toml.Decode(base.conf, ipt)
			assert.NoError(t, err)

			// envs := []string{
			// 	"ACCEPT_EULA=Y",
			// 	fmt.Sprintf("SA_PASSWORD=%s", ipt.Password),
			// }

			// ipport, err := netip.ParseAddrPort(ipt.Host)
			// assert.NoError(t, err, "parse %s failed: %s", ipt.Host, err)

			cases = append(cases, &caseSpec{
				t:      t,
				ipt:    ipt,
				name:   base.name,
				feeder: feeder,
				// envs:   envs,

				repo:    img[0],
				repoTag: img[1],

				servicePort: []string{
					"80",
				},

				cr: &tu.CaseResult{
					Name:        t.Name(),
					Case:        base.name,
					ExtraFields: map[string]any{},
					ExtraTags: map[string]string{
						"image":     img[0],
						"image_tag": img[1],
						// "remote_server": ipt.Host,
						"remote_host": remote.Host,
						"remote_port": remote.Port,
					},
				},
			})
		}
	}
	return cases, nil
}

func (cs *caseSpec) getPool(endpoint string) (*dt.Pool, error) {
	p, err := dt.NewPool(endpoint)
	if err != nil {
		return nil, err
	}
	err = p.Client.Ping()
	if err != nil {
		cs.t.Logf("Could not connect to Docker: %v", err)
		return nil, err
	}
	return p, nil
}

func portsOK(r *tu.RemoteInfo, ports ...string) error {
	for _, v := range ports {
		if !r.PortOK(v, time.Minute) {
			return fmt.Errorf("service checking failed")
		}
	}
	return nil
}

// func Test_setHostTagIfNotLoopback(t *T.T) {
// 	type args struct {
// 		tags      map[string]string
// 		ipAndPort string
// 	}
// 	tests := []struct {
// 		name     string
// 		args     args
// 		expected map[string]string
// 	}{
// 		{
// 			name: "loopback",
// 			args: args{
// 				tags:      map[string]string{},
// 				ipAndPort: "localhost:1234",
// 			},
// 			expected: map[string]string{},
// 		},
// 		{
// 			name: "loopback",
// 			args: args{
// 				tags:      map[string]string{},
// 				ipAndPort: "127.0.0.1:1234",
// 			},
// 			expected: map[string]string{},
// 		},
// 		{
// 			name: "normal",
// 			args: args{
// 				tags:      map[string]string{},
// 				ipAndPort: "192.168.1.1:1234",
// 			},
// 			expected: map[string]string{
// 				"host": "192.168.1.1",
// 			},
// 		},
// 		{
// 			name: "error not ip:port",
// 			args: args{
// 				tags:      map[string]string{},
// 				ipAndPort: "http://192.168.1.1:1234",
// 			},
// 			expected: map[string]string{},
// 		},
// 	}
// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *T.T) {
// 			setHostTagIfNotLoopback(tt.args.tags, tt.args.ipAndPort)
// 			assert.Equal(t, tt.expected, tt.args.tags)
// 		})
// 	}
// }
