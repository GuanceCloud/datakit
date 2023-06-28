// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package beats_output

import (
	"fmt"
	"io/ioutil"
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

// elastic/filebeat:7.17.6-logstash --> 7.17.6
func getVersion(name string) string {
	preDefined1 := "filebeat:"
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
		name    string // Also used as build image name:tag.
		verConf BUILD_CONFIG_VER
		conf    string
		opts    []inputs.PointCheckOption
	}{
		{
			name:    "elastic/filebeat:8.6.2-logstash",
			verConf: VER_7,
			conf:    `listen = "tcp://0.0.0.0:"`, // tcp://0.0.0.0:5044
		},

		{
			name:    "elastic/filebeat:7.17.9-logstash",
			verConf: VER_7,
			conf:    `listen = "tcp://0.0.0.0:"`, // tcp://0.0.0.0:5044
		},

		{
			name:    "elastic/filebeat:7.17.6-logstash",
			verConf: VER_7,
			conf:    `listen = "tcp://0.0.0.0:"`, // tcp://0.0.0.0:5044
		},

		{
			name:    "elastic/filebeat:6.0.0-logstash",
			verConf: VER_6,
			conf:    `listen = "tcp://0.0.0.0:"`, // tcp://0.0.0.0:5044
			opts: []inputs.PointCheckOption{
				inputs.WithOptionalTags("filepath", "host"), //nolint:lll
			},
		},

		{
			name:    "pubrepo.jiagouyun.com/image-repo-for-testing/filebeat:5.0.0-logstash",
			verConf: VER_5,
			conf:    `listen = "tcp://0.0.0.0:"`, // tcp://0.0.0.0:5044
			opts: []inputs.PointCheckOption{
				inputs.WithOptionalTags("filepath", "host"), //nolint:lll
			},
		},

		{
			name:    "pubrepo.jiagouyun.com/image-repo-for-testing/filebeat:1.3.0-logstash",
			verConf: VER_1,
			conf:    `listen = "tcp://0.0.0.0:"`, // tcp://0.0.0.0:5044
			opts: []inputs.PointCheckOption{
				inputs.WithOptionalTags("filepath", "host"), //nolint:lll
			},
		},

		{
			name:    "pubrepo.jiagouyun.com/image-repo-for-testing/filebeat:1.2.0-logstash",
			verConf: VER_1,
			conf:    `listen = "tcp://0.0.0.0:"`, // tcp://0.0.0.0:5044
			opts: []inputs.PointCheckOption{
				inputs.WithOptionalTags("filepath", "host"), //nolint:lll
			},
		},

		{
			name:    "pubrepo.jiagouyun.com/image-repo-for-testing/filebeat:1.1.0-logstash",
			verConf: VER_1,
			conf:    `listen = "tcp://0.0.0.0:"`, // tcp://0.0.0.0:5044
			opts: []inputs.PointCheckOption{
				inputs.WithOptionalTags("filepath", "host"), //nolint:lll
			},
		},

		{
			name:    "pubrepo.jiagouyun.com/image-repo-for-testing/filebeat:1.0.0-logstash",
			verConf: VER_1,
			conf:    `listen = "tcp://0.0.0.0:"`, // tcp://0.0.0.0:5044
			opts: []inputs.PointCheckOption{
				inputs.WithOptionalTags("filepath", "host"), //nolint:lll
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

		randPort := testutils.RandPort("tcp")
		randPortStr := fmt.Sprintf("%d", randPort)
		ipt.Listen += randPortStr // tcp://0.0.0.0:5044

		repoTag := strings.Split(base.name, ":")

		filebeatVersion := getVersion(base.name)

		cases = append(cases, &caseSpec{
			t:       t,
			ipt:     ipt,
			name:    base.name,
			feeder:  feeder,
			repo:    repoTag[0],
			repoTag: repoTag[1],

			dockerFileText: getDockerfile(filebeatVersion, randPortStr, base.verConf),
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
	opts           []inputs.PointCheckOption
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
		case measurementName:
			opts = append(opts, inputs.WithDoc(&loggingMeasurement{}))

			msgs := inputs.CheckPoint(pt, opts...)

			for _, msg := range msgs {
				cs.t.Logf("check measurement %s failed: %+#v", measurement, msg)
			}

			// TODO: error here
			if len(msgs) > 0 {
				return fmt.Errorf("check measurement %s failed: %+#v", measurement, msgs)
			}

			cs.mCount[measurementName] = struct{}{}

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
				ContainerName: uniqueContainerName,
				Name:          cs.name, // ATTENTION: not uniqueContainerName.

				Repository: cs.repo,
				Tag:        cs.repoTag,
				Env:        []string{fmt.Sprintf("DATAKIT_HOST=%s", extIP)},

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
		cs.t.Logf("%s", err.Error())
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

func (cs *caseSpec) getPortBindings() map[docker.Port][]docker.PortBinding {
	portBindings := make(map[docker.Port][]docker.PortBinding)

	for _, v := range cs.exposedPorts {
		portBindings[docker.Port(v)] = []docker.PortBinding{{HostIP: "0.0.0.0", HostPort: docker.Port(v).Port()}}
	}

	return portBindings
}

func (cs *caseSpec) portsOK(r *testutils.RemoteInfo) error {
	for _, v := range cs.exposedPorts {
		if !r.PortOK(docker.Port(v).Port(), time.Minute) {
			return fmt.Errorf("service checking failed")
		}
	}
	return nil
}

////////////////////////////////////////////////////////////////////////////////

// selfBuild indicates the image was customized built.
func getDockerfile(version, listenPort string, buildConfVer BUILD_CONFIG_VER) string {
	if len(version) == 0 {
		panic("version is empty")
	}

	if len(listenPort) == 0 {
		panic("listenPort is empty")
	}

	replacePair := map[string]string{
		"VERSION":      version,
		"a":            "$a",
		"DATAKIT_HOST": "${DATAKIT_HOST}",
		"LISTEN_PORT":  listenPort,
	}

	switch buildConfVer {
	case VER_1:
		return os.Expand(dockerFileLogstashV1, func(k string) string { return replacePair[k] })
	case VER_5:
		return os.Expand(dockerFileLogstashV5, func(k string) string { return replacePair[k] })
	case VER_6:
		return os.Expand(dockerFileLogstashV6, func(k string) string { return replacePair[k] })
	case VER_7:
		return os.Expand(dockerFileLogstashV7, func(k string) string { return replacePair[k] })
	}

	panic("should not been here.")
}

////////////////////////////////////////////////////////////////////////////////

// Dockerfiles.

type BUILD_CONFIG_VER int

const (
	VER_1 BUILD_CONFIG_VER = iota + 1 // Filebeat: v1.0 ~ v1.3.0
	VER_5                             // Filebeat: v5.0.0
	VER_6                             // Filebeat: v6.0.0
	VER_7                             // Filebeat: v7.0.0+
)

// Filebeat: v7.0.0+
/*
output.logstash:
  hosts: ["${DATAKIT_HOST}:${LISTEN_PORT}"]

filebeat.inputs:
- type: filestream
  id: my-filestream-id
  enabled: true
  paths:
    - /var/log/*.log
*/
const dockerFileLogstashV7 = `FROM elastic/filebeat:${VERSION}

USER root

RUN sed -i '10,13d' /usr/share/filebeat/filebeat.yml \
    && sed -i '$a output.logstash:' /usr/share/filebeat/filebeat.yml \
    && sed -i '$a \  hosts: ["${DATAKIT_HOST}:${LISTEN_PORT}"]' /usr/share/filebeat/filebeat.yml \
    && echo "" >> /usr/share/filebeat/filebeat.yml \
    && sed -i '$a filebeat.inputs:' /usr/share/filebeat/filebeat.yml \
    && sed -i '$a - type: filestream' /usr/share/filebeat/filebeat.yml \
    && sed -i '$a \  id: my-filestream-id' /usr/share/filebeat/filebeat.yml \
    && sed -i '$a \  enabled: true' /usr/share/filebeat/filebeat.yml \
    && sed -i '$a \  paths:' /usr/share/filebeat/filebeat.yml \
    && sed -i '$a \    - /var/log/*.log' /usr/share/filebeat/filebeat.yml

USER filebeat`

// Filebeat: v6.0.0
/*
filebeat.prospectors:
- type: log
  enabled: true
  paths:
    - /var/log/*.log

output.logstash:
  enabled: true
  hosts: ["10.100.65.61:16209"]

logging.to_files: true
*/
const dockerFileLogstashV6 = `FROM elastic/filebeat:${VERSION}

USER root

RUN    echo "" > /usr/share/filebeat/filebeat.yml \
    && sed -i '$a filebeat.prospectors:' /usr/share/filebeat/filebeat.yml \
    && sed -i '$a  - type: log' /usr/share/filebeat/filebeat.yml \
    && sed -i '$a \  enabled: true' /usr/share/filebeat/filebeat.yml \
    && sed -i '$a \  paths:' /usr/share/filebeat/filebeat.yml \
    && sed -i '$a \    - /var/log/*.log' /usr/share/filebeat/filebeat.yml \
    && echo "" >> /usr/share/filebeat/filebeat.yml \
    && sed -i '$a output.logstash:' /usr/share/filebeat/filebeat.yml \
    && sed -i '$a \  enabled: true' /usr/share/filebeat/filebeat.yml \
    && sed -i '$a \  hosts: ["${DATAKIT_HOST}:${LISTEN_PORT}"]' /usr/share/filebeat/filebeat.yml \
    && sed -i '$a logging.to_files: true' /usr/share/filebeat/filebeat.yml

WORKDIR /usr/share/filebeat`

// Filebeat: v5.0.0
/*
filebeat.prospectors:
- input_type: log
  paths:
    - /var/log/*.log

output.logstash:
  hosts: ["${DATAKIT_HOST}:${LISTEN_PORT}"]
*/
const dockerFileLogstashV5 = `FROM pubrepo.jiagouyun.com/image-repo-for-testing/filebeat:${VERSION}

RUN sed -i '$a filebeat.prospectors:' /usr/share/filebeat/filebeat.yml \
    && sed -i '$a - input_type: log' /usr/share/filebeat/filebeat.yml \
    && sed -i '$a \  paths:' /usr/share/filebeat/filebeat.yml \
    && sed -i '$a \    - /usr/share/filebeat/logs/*' /usr/share/filebeat/filebeat.yml \
    && echo "" >> /usr/share/filebeat/filebeat.yml \
    && sed -i '$a output.logstash:' /usr/share/filebeat/filebeat.yml \
    && sed -i '$a \  hosts: ["${DATAKIT_HOST}:${LISTEN_PORT}"]' /usr/share/filebeat/filebeat.yml`

// Filebeat: v1.0 ~ v1.3.0
/*
filebeat:
  prospectors:
    -
      paths:
        - /var/log/*.log

output:
  logstash:
    hosts: ["${DATAKIT_HOST}:${LISTEN_PORT}"]

logging:
  to_syslog: false
*/
const dockerFileLogstashV1 = `FROM pubrepo.jiagouyun.com/image-repo-for-testing/filebeat:${VERSION}

RUN    touch /var/log/1.log /var/log/2.log \
    && echo "123" >> /var/log/1.log \
    && echo "456" >> /var/log/2.log \
    && sed -i '$a filebeat:' /usr/share/filebeat/filebeat.yml \
    && sed -i '$a \  prospectors:' /usr/share/filebeat/filebeat.yml \
    && sed -i '$a \    -' /usr/share/filebeat/filebeat.yml \
    && sed -i '$a \      paths:' /usr/share/filebeat/filebeat.yml \
    && sed -i '$a \        - /var/log/*.log' /usr/share/filebeat/filebeat.yml \
    && echo "" >> /usr/share/filebeat/filebeat.yml \
    && sed -i '$a output:' /usr/share/filebeat/filebeat.yml \
    && sed -i '$a \  logstash:' /usr/share/filebeat/filebeat.yml \
    && sed -i '$a \    hosts: ["${DATAKIT_HOST}:${LISTEN_PORT}"]' /usr/share/filebeat/filebeat.yml \
    && echo "" >> /usr/share/filebeat/filebeat.yml \
    && sed -i '$a logging:' /usr/share/filebeat/filebeat.yml \
    && sed -i '$a \  to_syslog: false' /usr/share/filebeat/filebeat.yml

WORKDIR /usr/share/filebeat`
