// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package nginx

import (
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"sync"
	T "testing"
	"time"

	"github.com/BurntSushi/toml"
	"github.com/GuanceCloud/cliutils/point"
	dt "github.com/ory/dockertest/v3"
	"github.com/ory/dockertest/v3/docker"
	dc "github.com/ory/dockertest/v3/docker"
	"github.com/stretchr/testify/assert"
	tu "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/testutils"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
)

func TestNginxInput(t *T.T) {
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

func buildCases(t *T.T) ([]*caseSpec, error) {
	t.Helper()

	remote := tu.GetRemote()

	bases := []struct {
		name           string
		conf           string
		dockerFileText string
		exposedPorts   []string
	}{
		{
			name:           "nginx:http_stub_status_module",
			conf:           fmt.Sprintf(`url = "http://%s:80/server_status"`, remote.Host),
			dockerFileText: dockerFileHTTPStubStatusModule,
			exposedPorts:   []string{"80/tcp"},
		},
	}

	var cases []*caseSpec

	// compose cases
	for _, base := range bases {
		feeder := io.NewMockedFeeder()

		ipt := defaultInput()
		ipt.feeder = feeder

		_, err := toml.Decode(base.conf, ipt)
		assert.NoError(t, err)

		repoTag := strings.Split(base.name, ":")

		cases = append(cases, &caseSpec{
			t:       t,
			ipt:     ipt,
			name:    base.name,
			feeder:  feeder,
			repo:    repoTag[0],
			repoTag: repoTag[1],

			exposedPorts:   base.exposedPorts,
			dockerFileText: base.dockerFileText,

			cr: &tu.CaseResult{
				Name:        t.Name(),
				Case:        base.name,
				ExtraFields: map[string]any{},
				ExtraTags: map[string]string{
					"image":       repoTag[0],
					"image_tag":   repoTag[1],
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
	t *T.T

	name           string
	repo           string
	repoTag        string
	envs           []string
	exposedPorts   []string
	dockerFileText string

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
		case "nginx":
			msgs := inputs.CheckPoint(pt, &NginxMeasurement{}, inputs.WithAllowExtraTags(len(cs.ipt.Tags) > 0))

			for _, msg := range msgs {
				cs.t.Logf("check measurement %s failed: %+#v", measurement, msg)
			}

			// TODO: error here
			if len(msgs) > 0 {
				switch cs.name {
				case "nginx:http_stub_status_module":
					if !assert.ElementsMatch(cs.t, []string{"tag nginx_version not found", "field load_timestamp not found"}, msgs) {
						return fmt.Errorf("check measurement %s failed: %+#v", measurement, msgs)
					}
				}
			}

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
	r := tu.GetRemote()
	dockerTCP := r.TCPURL()

	cs.t.Logf("get remote: %+#v, TCP: %s", r, dockerTCP)

	start := time.Now()

	p, err := cs.getPool(dockerTCP)
	if err != nil {
		return err
	}

	containerName := cs.getContainterName()

	// remove the container if exist.
	if err := p.RemoveContainerByName(containerName); err != nil {
		return err
	}

	dockerFileDir, dockerFilePath, err := cs.getDockerFilePath()
	if err != nil {
		return err
	}
	defer os.RemoveAll(dockerFileDir)

	resource, err := p.BuildAndRunWithOptions(
		dockerFilePath,

		&dt.RunOptions{
			Name: cs.name,

			// specify container image & tag
			Repository: cs.repo,
			Tag:        cs.repoTag,

			ExposedPorts: cs.exposedPorts,
			PortBindings: cs.getPortBindings(),
		},

		func(c *dc.HostConfig) {
			c.RestartPolicy = dc.RestartPolicy{Name: "no"}
			c.AutoRemove = true
		},
	)
	if err != nil {
		return err
	}

	cs.pool = p
	cs.resource = resource

	cs.t.Logf("check service(%s:%v)...", r.Host, cs.exposedPorts)

	if err := cs.portsOK(r); err != nil {
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

func (cs *caseSpec) getDockerFilePath() (dirName string, fileName string, err error) {
	tmpDir, err := ioutil.TempDir("", "dockerfiles_")
	if err != nil {
		cs.t.Logf("ioutil.TempDir failed: %s", err.Error())
		return "", "", err
	}

	tmpFile, err := ioutil.TempFile(tmpDir, "dockerfile_")
	if err != nil {
		cs.t.Logf("ioutil.TempFile failed: %s", err.Error())
		return "", "", err
	}

	_, err = tmpFile.WriteString(cs.dockerFileText)
	if err != nil {
		cs.t.Logf("TempFile.WriteString failed: %s", err.Error())
		return "", "", err
	}

	if err := os.Chmod(tmpFile.Name(), os.ModePerm); err != nil {
		cs.t.Logf("os.Chmod failed: %s", err.Error())
		return "", "", err
	}

	if err := tmpFile.Close(); err != nil {
		cs.t.Logf("Close failed: %s", err.Error())
		return "", "", err
	}

	return tmpDir, tmpFile.Name(), nil
}

func (cs *caseSpec) getContainterName() string {
	nameTag := strings.Split(cs.name, ":")
	return nameTag[0]
}

func (cs *caseSpec) getPortBindings() map[dc.Port][]dc.PortBinding {
	portBindings := make(map[dc.Port][]dc.PortBinding)

	for _, v := range cs.exposedPorts {
		portBindings[dc.Port(v)] = []docker.PortBinding{{HostIP: "0.0.0.0", HostPort: dc.Port(v).Port()}}
	}

	return portBindings
}

func (cs *caseSpec) portsOK(r *tu.RemoteInfo) error {
	for _, v := range cs.exposedPorts {
		if !r.PortOK(dc.Port(v).Port(), time.Minute) {
			return fmt.Errorf("service checking failed")
		}
	}
	return nil
}

////////////////////////////////////////////////////////////////////////////////

// Dockerfiles.

const dockerFileHTTPStubStatusModule = `FROM nginx:latest

RUN sed -i "/location \/ {/i\    location = /server_status {" /etc/nginx/conf.d/default.conf \
    && sed -i "/location \/ {/i\        stub_status;" /etc/nginx/conf.d/default.conf \
    && sed -i "/location \/ {/i\    }\n" /etc/nginx/conf.d/default.conf`
