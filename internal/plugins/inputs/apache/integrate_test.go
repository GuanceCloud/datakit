// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package apache

import (
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"os"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/BurntSushi/toml"
	"github.com/GuanceCloud/cliutils/point"
	dockertest "github.com/ory/dockertest/v3"
	"github.com/ory/dockertest/v3/docker"
	"github.com/stretchr/testify/require"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/testutils"
)

// ATTENTION: Docker version should use v20.10.18 in integrate tests. Other versions are not tested.

func TestIntegrate(t *testing.T) {
	if !testutils.CheckIntegrationTestingRunning() {
		t.Skip()
	}

	testutils.PurgeRemoteByName(inputName)       // purge at first.
	defer testutils.PurgeRemoteByName(inputName) // purge at last.

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
				t.Parallel()
				caseStart := time.Now()

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

func getConfAccessPoint(host, port string) string {
	return fmt.Sprintf("http://%s/server-status?auto", net.JoinHostPort(host, port))
}

// httpd:2.4.54-alpine-server-status --> 2.4.54
func getVersion(name string) string {
	preDefined1 := "httpd:"
	preDefined2 := "-"

	ndx1 := strings.Index(name, preDefined1)
	if ndx1 == -1 {
		return ""
	}

	name2 := name[ndx1+len(preDefined1):]
	ndx2 := strings.Index(name2, preDefined2)
	if ndx2 == -1 {
		return name2
	}

	return name2[:ndx2]
}

func buildCases(t *testing.T) ([]*caseSpec, error) {
	t.Helper()

	remote := testutils.GetRemote()

	bases := []struct {
		name           string // Also used as build image name:tag.
		selfBuild      bool   // Indicates the image was customized built.
		conf           string
		dockerFileText string // Empty if not build image.
		exposedPorts   []string
		opts           []inputs.PointCheckOption
		mPathCount     map[string]int
	}{
		{
			name:      "httpd:2.4.56-alpine-server-status",
			selfBuild: false,
			conf: `url = ""
			interval = "1s"`, // set conf URL later.
			exposedPorts: []string{"80/tcp"},
			mPathCount: map[string]int{
				"/": 10,
			},
		},

		{
			name:      "httpd:2.4.54-alpine-server-status",
			selfBuild: false,
			conf: `url = ""
			interval = "1s"`, // set conf URL later.
			exposedPorts: []string{"80/tcp"},
			mPathCount: map[string]int{
				"/": 10,
			},
		},

		{
			name:      "httpd:2.4.41-alpine-server-status",
			selfBuild: false,
			conf: `url = ""
			interval = "1s"`, // set conf URL later.
			exposedPorts: []string{"80/tcp"},
			mPathCount: map[string]int{
				"/": 10,
			},
		},

		{
			name:      "httpd:2.4.38-alpine-server-status",
			selfBuild: false,
			conf: `url = ""
			interval = "1s"`, // set conf URL later.
			exposedPorts: []string{"80/tcp"},
			mPathCount: map[string]int{
				"/": 10,
			},
		},

		{
			name:      "httpd:2.4.29-alpine-server-status",
			selfBuild: false,
			conf: `url = ""
			interval = "1s"`, // set conf URL later.
			exposedPorts: []string{"80/tcp"},
			opts: []inputs.PointCheckOption{
				inputs.WithOptionalFields("cpu_load"), //nolint:lll
			},
			mPathCount: map[string]int{
				"/": 10,
			},
		},

		{
			name:      "httpd:2.4.6-alpine-server-status",
			selfBuild: true,
			conf: `url = ""
			interval = "1s"`, // set conf URL later.
			exposedPorts: []string{"80/tcp"},
			opts: []inputs.PointCheckOption{
				inputs.WithOptionalFields("cpu_load"),                   //nolint:lll
				inputs.WithOptionalTags("server_version", "server_mpm"), //nolint:lll
			},
			mPathCount: map[string]int{
				"/": 10,
			},
		},
	}

	var cases []*caseSpec

	// compose cases
	for _, base := range bases {
		feeder := io.NewMockedFeeder()

		ipt := defaultInput()
		ipt.feeder = feeder

		_, err := toml.Decode(base.conf, ipt)
		require.NoError(t, err)

		repoTag := strings.Split(base.name, ":")

		cases = append(cases, &caseSpec{
			t:       t,
			ipt:     ipt,
			name:    base.name,
			feeder:  feeder,
			repo:    repoTag[0],
			repoTag: repoTag[1],

			dockerFileText: getDockerfile(getVersion(base.name), base.selfBuild),
			exposedPorts:   base.exposedPorts,
			mPathCount:     base.mPathCount,
			opts:           base.opts,

			cr: &testutils.CaseResult{
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
	t *testing.T

	name           string
	repo           string
	repoTag        string
	dockerFileText string
	exposedPorts   []string
	serverPorts    []string
	opts           []inputs.PointCheckOption
	mPathCount     map[string]int
	mCount         map[string]struct{}

	ipt    *Input
	feeder *io.MockedFeeder

	pool     *dockertest.Pool
	resource *dockertest.Resource

	cr *testutils.CaseResult
}

func (cs *caseSpec) checkPoint(pts []*point.Point) error {
	var opts []inputs.PointCheckOption
	opts = append(opts, inputs.WithExtraTags(cs.ipt.Tags))
	opts = append(opts, cs.opts...)

	for _, pt := range pts {
		measurement := string(pt.Name())

		switch measurement {
		case inputName:
			opts = append(opts, inputs.WithDoc(&Measurement{}))

			msgs := inputs.CheckPoint(pt, opts...)

			for _, msg := range msgs {
				cs.t.Logf("check measurement %s failed: %+#v", measurement, msg)
			}

			// TODO: error here
			if len(msgs) > 0 {
				return fmt.Errorf("check measurement %s failed: %+#v", measurement, msgs)
			}

			cs.mCount[inputName] = struct{}{}

		default: // TODO: check other measurement
			panic("unknown measurement")
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
	r := testutils.GetRemote()
	dockerTCP := r.TCPURL()

	cs.t.Logf("get remote: %+#v, TCP: %s", r, dockerTCP)

	start := time.Now()

	p, err := cs.getPool(dockerTCP)
	if err != nil {
		return err
	}

	dockerFileDir, dockerFilePath, err := cs.getDockerFilePath()
	if err != nil {
		return err
	}
	defer os.RemoveAll(dockerFileDir)

	extIP, err := testutils.ExternalIP()
	if err != nil {
		return err
	}

	uniqueContainerName := testutils.GetUniqueContainerName(inputName)

	var resource *dockertest.Resource

	if len(cs.dockerFileText) == 0 {
		// Just run a container from existing docker image.
		resource, err = p.RunWithOptions(
			&dockertest.RunOptions{
				Name: uniqueContainerName, // ATTENTION: not cs.name.

				Repository: cs.repo,
				Tag:        cs.repoTag,
				Env:        []string{fmt.Sprintf("DATAKIT_HOST=%s", extIP)},

				ExposedPorts: cs.exposedPorts,
			},

			func(c *docker.HostConfig) {
				c.RestartPolicy = docker.RestartPolicy{Name: "no"}
				c.AutoRemove = true
			},
		)
	} else {
		// Build docker image from Dockerfile and run a container from it.
		resource, err = p.BuildAndRunWithOptions(
			dockerFilePath,

			&dockertest.RunOptions{
				ContainerName: uniqueContainerName,
				Name:          cs.name, // ATTENTION: not uniqueContainerName.

				Repository: cs.repo,
				Tag:        cs.repoTag,
				Env:        []string{fmt.Sprintf("DATAKIT_HOST=%s", extIP)},

				ExposedPorts: cs.exposedPorts,
			},

			func(c *docker.HostConfig) {
				c.RestartPolicy = docker.RestartPolicy{Name: "no"}
				c.AutoRemove = true
			},
		)
	}

	if err != nil {
		return err
	}

	cs.pool = p
	cs.resource = resource

	if err := cs.getMappingPorts(); err != nil {
		return err
	}
	cs.ipt.URL = getConfAccessPoint(r.Host, cs.serverPorts[0]) // set conf URL here.

	cs.t.Logf("check service(%s:%v)...", r.Host, cs.serverPorts)

	if err := cs.portsOK(r); err != nil {
		return err
	}

	cs.cr.AddField("container_ready_cost", int64(time.Since(start)))

	cs.runHTTPTests(r)

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
	pts, err := cs.feeder.AnyPoints(5 * time.Minute)
	if err != nil {
		return err
	}

	cs.cr.AddField("point_latency", int64(time.Since(start)))
	cs.cr.AddField("point_count", len(pts))

	cs.t.Logf("get %d points", len(pts))
	cs.mCount = make(map[string]struct{})
	if err := cs.checkPoint(pts); err != nil {
		return err
	}

	cs.t.Logf("stop input...")
	cs.ipt.Terminate()

	require.Equal(cs.t, 1, len(cs.mCount))

	cs.t.Logf("exit...")
	wg.Wait()

	return nil
}

func (cs *caseSpec) getPool(endpoint string) (*dockertest.Pool, error) {
	p, err := dockertest.NewPool(endpoint)
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
	if len(cs.dockerFileText) == 0 {
		return
	}

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

func (cs *caseSpec) getMappingPorts() error {
	cs.serverPorts = make([]string, len(cs.exposedPorts))
	for k, v := range cs.exposedPorts {
		mapStr := cs.resource.GetHostPort(v)
		_, port, err := net.SplitHostPort(mapStr)
		if err != nil {
			return err
		}
		cs.serverPorts[k] = port
	}
	return nil
}

func (cs *caseSpec) portsOK(r *testutils.RemoteInfo) error {
	for _, v := range cs.serverPorts {
		if !r.PortOK(docker.Port(v).Port(), time.Minute) {
			return fmt.Errorf("service checking failed")
		}
	}
	return nil
}

// Launch large amount of HTTP requests to remote nginx.
func (cs *caseSpec) runHTTPTests(r *testutils.RemoteInfo) {
	for _, v := range cs.serverPorts {
		for path, count := range cs.mPathCount {
			newURL := fmt.Sprintf("http://%s%s", net.JoinHostPort(r.Host, v), path)

			var wg sync.WaitGroup
			wg.Add(count)

			for i := 0; i < count; i++ {
				go func() {
					defer wg.Done()

					netTransport := &http.Transport{
						Dial: (&net.Dialer{
							Timeout: 10 * time.Second,
						}).Dial,
						TLSHandshakeTimeout: 10 * time.Second,
					}
					netClient := &http.Client{
						Timeout:   time.Second * 20,
						Transport: netTransport,
					}

					resp, err := netClient.Get(newURL)
					if err != nil {
						fmt.Printf("HTTP GET failed: %v\n", err)
						return
					}
					defer resp.Body.Close()
				}()
			}

			wg.Wait()
		}
	}
}

////////////////////////////////////////////////////////////////////////////////

// selfBuild indicates the image was customized built.
func getDockerfile(version string, selfBuild bool) string {
	if len(version) == 0 {
		panic("version is empty")
	}

	replacePair := map[string]string{
		"VERSION": version,
		"a":       "$a",
	}

	if selfBuild {
		return os.Expand(dockerFileServerStatusSelfBuild, func(k string) string { return replacePair[k] })
	}

	return os.Expand(dockerFileServerStatus, func(k string) string { return replacePair[k] })
}

////////////////////////////////////////////////////////////////////////////////

// Dockerfiles.

const dockerFileServerStatus = `FROM httpd:${VERSION}-alpine

  RUN sed -i '$a <Location /server-status>' /usr/local/apache2/conf/httpd.conf \
	&& sed -i '$a SetHandler server-status' /usr/local/apache2/conf/httpd.conf \
	&& sed -i '$a Order Allow,Deny' /usr/local/apache2/conf/httpd.conf \
	&& sed -i '$a Allow from all' /usr/local/apache2/conf/httpd.conf \
	&& sed -i '$a </Location>' /usr/local/apache2/conf/httpd.conf`

const dockerFileServerStatusSelfBuild = `FROM pubrepo.jiagouyun.com/image-repo-for-testing/httpd:${VERSION}-alpine

  RUN sed -i '$a <Location /server-status>' /usr/local/apache2/conf/httpd.conf \
	&& sed -i '$a SetHandler server-status' /usr/local/apache2/conf/httpd.conf \
	&& sed -i '$a Order Allow,Deny' /usr/local/apache2/conf/httpd.conf \
	&& sed -i '$a Allow from all' /usr/local/apache2/conf/httpd.conf \
	&& sed -i '$a </Location>' /usr/local/apache2/conf/httpd.conf`
