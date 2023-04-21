// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package tomcat

import (
	"fmt"
	"io/ioutil"
	"net"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/BurntSushi/toml"
	"github.com/GuanceCloud/cliutils/point"
	dockertest "github.com/ory/dockertest/v3"
	"github.com/ory/dockertest/v3/docker"
	"github.com/stretchr/testify/require"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/testutils"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
)

// ATTENTION: Docker version should use v20.10.18 in integrate tests. Other versions are not tested.
// Reference: https://jolokia.org/reference/html/agents.html#jvm-agent

var mCount map[string]struct{} = make(map[string]struct{}) // Length of got measurements.

func TestTomcatInput(t *testing.T) {
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
		t.Run(tc.name, func(t *testing.T) {
			caseStart := time.Now()

			t.Logf("testing %s...", tc.name)

			if err := tc.run(); err != nil {
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

				require.NoError(t, tc.pool.Purge(tc.resource))
			})
		})
	}
}

func buildCases(t *testing.T) ([]*caseSpec, error) {
	t.Helper()

	remote := testutils.GetRemote()

	bases := []struct {
		name                             string // Also used as build image name:tag.
		conf                             string
		dockerFileText                   string // Empty if not build image.
		exposedPorts                     []string
		optsTomcatGlobalRequestProcessor []inputs.PointCheckOption
		optsTomcatJspMonitor             []inputs.PointCheckOption
		optsTomcatThreadPool             []inputs.PointCheckOption
		optsTomcatServlet                []inputs.PointCheckOption
		optsTomcatCache                  []inputs.PointCheckOption
		mPathCount                       map[string]int
	}{
		{
			name: "pubrepo.jiagouyun.com/image-repo-for-testing/tomcat:8-jolokia",
			conf: fmt.Sprintf(`username = "jolokia_user"
			password = "123456@secPassWd"
			urls = ["http://%s/jolokia"]
			[[metric]]
			  name     = "tomcat_global_request_processor"
			  mbean    = '''Catalina:name="*",type=GlobalRequestProcessor'''
			  paths    = ["requestCount","bytesReceived","bytesSent","processingTime","errorCount"]
			  tag_keys = ["name"]
			[[metric]]
			  name     = "tomcat_jsp_monitor"
			  mbean    = "Catalina:J2EEApplication=*,J2EEServer=*,WebModule=*,name=jsp,type=JspMonitor"
			  paths    = ["jspReloadCount","jspCount","jspUnloadCount"]
			  tag_keys = ["J2EEApplication","J2EEServer","WebModule"]
			[[metric]]
			  name     = "tomcat_thread_pool"
			  mbean    = "Catalina:name=\"*\",type=ThreadPool"
			  paths    = ["maxThreads","currentThreadCount","currentThreadsBusy"]
			  tag_keys = ["name"]
			[[metric]]
			  name     = "tomcat_servlet"
			  mbean    = "Catalina:J2EEApplication=*,J2EEServer=*,WebModule=*,j2eeType=Servlet,name=*"
			  paths    = ["processingTime","errorCount","requestCount"]
			  tag_keys = ["name","J2EEApplication","J2EEServer","WebModule"]
			[[metric]]
			  name     = "tomcat_cache"
			  mbean    = "Catalina:context=*,host=*,name=Cache,type=WebResourceRoot"
			  paths    = ["hitCount","lookupCount"]
			  tag_keys = ["context","host"]
			  tag_prefix = "tomcat_"`, net.JoinHostPort(remote.Host, fmt.Sprintf("%d", testutils.RandPort("tcp")))),
			exposedPorts: []string{"8080/tcp"},
			mPathCount: map[string]int{
				"/": 10,
			},
		},

		{
			name: "pubrepo.jiagouyun.com/image-repo-for-testing/tomcat:9-jolokia",
			conf: fmt.Sprintf(`username = "jolokia_user"
			password = "123456@secPassWd"
			urls = ["http://%s/jolokia"]
			[[metric]]
			  name     = "tomcat_global_request_processor"
			  mbean    = '''Catalina:name="*",type=GlobalRequestProcessor'''
			  paths    = ["requestCount","bytesReceived","bytesSent","processingTime","errorCount"]
			  tag_keys = ["name"]
			[[metric]]
			  name     = "tomcat_jsp_monitor"
			  mbean    = "Catalina:J2EEApplication=*,J2EEServer=*,WebModule=*,name=jsp,type=JspMonitor"
			  paths    = ["jspReloadCount","jspCount","jspUnloadCount"]
			  tag_keys = ["J2EEApplication","J2EEServer","WebModule"]
			[[metric]]
			  name     = "tomcat_thread_pool"
			  mbean    = "Catalina:name=\"*\",type=ThreadPool"
			  paths    = ["maxThreads","currentThreadCount","currentThreadsBusy"]
			  tag_keys = ["name"]
			[[metric]]
			  name     = "tomcat_servlet"
			  mbean    = "Catalina:J2EEApplication=*,J2EEServer=*,WebModule=*,j2eeType=Servlet,name=*"
			  paths    = ["processingTime","errorCount","requestCount"]
			  tag_keys = ["name","J2EEApplication","J2EEServer","WebModule"]
			[[metric]]
			  name     = "tomcat_cache"
			  mbean    = "Catalina:context=*,host=*,name=Cache,type=WebResourceRoot"
			  paths    = ["hitCount","lookupCount"]
			  tag_keys = ["context","host"]
			  tag_prefix = "tomcat_"`, net.JoinHostPort(remote.Host, fmt.Sprintf("%d", testutils.RandPort("tcp")))),
			exposedPorts: []string{"8080/tcp"},
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
		ipt.Feeder = feeder

		_, err := toml.Decode(base.conf, ipt)
		require.NoError(t, err)

		uURL, err := url.Parse(ipt.URLs[0])
		require.NoError(t, err, "parse %s failed: %s", ipt.URLs[0], err)

		repoTag := strings.Split(base.name, ":")

		cases = append(cases, &caseSpec{
			t:       t,
			ipt:     ipt,
			name:    base.name,
			feeder:  feeder,
			repo:    repoTag[0],
			repoTag: repoTag[1],

			dockerFileText: base.dockerFileText,
			exposedPorts:   base.exposedPorts,
			serverPorts:    []string{uURL.Port()},

			optsTomcatGlobalRequestProcessor: base.optsTomcatGlobalRequestProcessor,
			optsTomcatJspMonitor:             base.optsTomcatJspMonitor,
			optsTomcatThreadPool:             base.optsTomcatThreadPool,
			optsTomcatServlet:                base.optsTomcatServlet,
			optsTomcatCache:                  base.optsTomcatCache,

			mPathCount: base.mPathCount,

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

	name                             string
	repo                             string
	repoTag                          string
	dockerFileText                   string
	exposedPorts                     []string
	serverPorts                      []string
	optsTomcatGlobalRequestProcessor []inputs.PointCheckOption
	optsTomcatJspMonitor             []inputs.PointCheckOption
	optsTomcatThreadPool             []inputs.PointCheckOption
	optsTomcatServlet                []inputs.PointCheckOption
	optsTomcatCache                  []inputs.PointCheckOption
	mPathCount                       map[string]int

	ipt    *Input
	feeder *io.MockedFeeder

	pool     *dockertest.Pool
	resource *dockertest.Resource

	cr *testutils.CaseResult
}

func (cs *caseSpec) checkPoint(pts []*point.Point) error {
	var opts []inputs.PointCheckOption
	opts = append(opts, inputs.WithExtraTags(cs.ipt.Tags))

	for _, pt := range pts {
		measurement := string(pt.Name())

		switch measurement {
		case TomcatGlobalRequestProcessor:
			opts = append(opts, cs.optsTomcatGlobalRequestProcessor...)
			opts = append(opts, inputs.WithDoc(&TomcatGlobalRequestProcessorM{}))

			msgs := inputs.CheckPoint(pt, opts...)

			for _, msg := range msgs {
				cs.t.Logf("check measurement %s failed: %+#v", measurement, msg)
			}

			// TODO: error here
			if len(msgs) > 0 {
				return fmt.Errorf("check measurement %s failed: %+#v", measurement, msgs)
			}

			mCount[TomcatGlobalRequestProcessor] = struct{}{}

		case TomcatJspMonitor:
			opts = append(opts, cs.optsTomcatJspMonitor...)
			opts = append(opts, inputs.WithDoc(&TomcatJspMonitorM{}))

			msgs := inputs.CheckPoint(pt, opts...)

			for _, msg := range msgs {
				cs.t.Logf("check measurement %s failed: %+#v", measurement, msg)
			}

			// TODO: error here
			if len(msgs) > 0 {
				return fmt.Errorf("check measurement %s failed: %+#v", measurement, msgs)
			}

			mCount[TomcatJspMonitor] = struct{}{}

		case TomcatThreadPool:
			opts = append(opts, cs.optsTomcatThreadPool...)
			opts = append(opts, inputs.WithDoc(&TomcatThreadPoolM{}))

			msgs := inputs.CheckPoint(pt, opts...)

			for _, msg := range msgs {
				cs.t.Logf("check measurement %s failed: %+#v", measurement, msg)
			}

			// TODO: error here
			if len(msgs) > 0 {
				return fmt.Errorf("check measurement %s failed: %+#v", measurement, msgs)
			}

			mCount[TomcatThreadPool] = struct{}{}

		case TomcatServlet:
			opts = append(opts, cs.optsTomcatServlet...)
			opts = append(opts, inputs.WithDoc(&TomcatServletM{}))

			msgs := inputs.CheckPoint(pt, opts...)

			for _, msg := range msgs {
				cs.t.Logf("check measurement %s failed: %+#v", measurement, msg)
			}

			// TODO: error here
			if len(msgs) > 0 {
				return fmt.Errorf("check measurement %s failed: %+#v", measurement, msgs)
			}

			mCount[TomcatServlet] = struct{}{}

		case TomcatCache:
			opts = append(opts, cs.optsTomcatCache...)
			opts = append(opts, inputs.WithDoc(&TomcatCacheM{}))

			msgs := inputs.CheckPoint(pt, opts...)

			for _, msg := range msgs {
				cs.t.Logf("check measurement %s failed: %+#v", measurement, msg)
			}

			// TODO: error here
			if len(msgs) > 0 {
				return fmt.Errorf("check measurement %s failed: %+#v", measurement, msgs)
			}

			mCount[TomcatCache] = struct{}{}

		default: // TODO: check other measurement
			panic("not implement")
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

	containerName := cs.getContainterName()

	// Remove the container if exist.
	if err := p.RemoveContainerByName(containerName); err != nil {
		return err
	}

	dockerFileDir, dockerFilePath, err := cs.getDockerFilePath()
	if err != nil {
		return err
	}
	defer os.RemoveAll(dockerFileDir)

	var resource *dockertest.Resource

	if len(cs.dockerFileText) == 0 {
		// Just run a container from existing docker image.
		resource, err = p.RunWithOptions(
			&dockertest.RunOptions{
				Name: containerName, // ATTENTION: not cs.name.

				Repository: cs.repo,
				Tag:        cs.repoTag,
				// Env:        []string{"JOLOKIA_PORT=59090"},

				ExposedPorts: cs.exposedPorts,
				PortBindings: cs.getPortBindings(),
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
				Name: cs.name,

				Repository: cs.repo,
				Tag:        cs.repoTag,
				// Env:        []string{"JOLOKIA_PORT=59090"},

				ExposedPorts: cs.exposedPorts,
				PortBindings: cs.getPortBindings(),
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

	cs.t.Logf("check service(%s:%v)...", r.Host, cs.serverPorts)

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
	pts, err := cs.feeder.NPoints(50, 5*time.Minute)
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

	require.Equal(cs.t, 5, len(mCount))

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

func (cs *caseSpec) getContainterName() string {
	nameTag := strings.Split(cs.name, ":")
	name := filepath.Base(nameTag[0])
	return name
}

func (cs *caseSpec) getPortBindings() map[docker.Port][]docker.PortBinding {
	portBindings := make(map[docker.Port][]docker.PortBinding)

	// check ports' mapping.
	require.Equal(cs.t, len(cs.exposedPorts), len(cs.serverPorts))

	for k, v := range cs.exposedPorts {
		portBindings[docker.Port(v)] = []docker.PortBinding{{HostPort: docker.Port(cs.serverPorts[k]).Port()}}
	}

	return portBindings
}

func (cs *caseSpec) portsOK(r *testutils.RemoteInfo) error {
	for _, v := range cs.serverPorts {
		if !r.PortOK(docker.Port(v).Port(), time.Minute) {
			return fmt.Errorf("service checking failed")
		}
	}
	return nil
}
