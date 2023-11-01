// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package elasticsearch

import (
	"context"
	"crypto/tls"
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
	dt "github.com/ory/dockertest/v3"
	"github.com/ory/dockertest/v3/docker"
	"github.com/stretchr/testify/assert"

	"github.com/GuanceCloud/cliutils/point"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/testutils"
)

const (
	User         = "elastic"
	UserPassword = "Abc123!"

	RepoURL = "pubrepo.guance.com/image-repo-for-testing/elasticsearch/"
)

type (
	validateFunc     func(pts []*point.Point, cs *caseSpec) error
	getConfFunc      func(c containerInfo) string
	serviceReadyFunc func(ipt *Input) error
	serviceOKFunc    func(port string) bool
)

type caseSpec struct {
	t *testing.T

	name  string
	image string
	envs  []string

	validate     validateFunc
	getConf      getConfFunc
	serviceOK    serviceOKFunc
	serviceReady serviceReadyFunc
	bindingPort  docker.Port

	ipt    *Input
	feeder *io.MockedFeeder

	pool     *dt.Pool
	resource *dt.Resource

	cr *testutils.CaseResult
}

type caseItem struct {
	name         string
	getConf      getConfFunc
	validate     validateFunc
	serviceReady serviceReadyFunc
	serviceOK    serviceOKFunc
	envs         []string
	images       []string

	bindingPort docker.Port
}

var cases = []caseItem{
	{
		name: "elasticsearch-ok",
		images: []string{
			fmt.Sprintf("%selasticsearch:7.0.0", RepoURL),
			fmt.Sprintf("%selasticsearch:6.4.0", RepoURL),
		},
		bindingPort: "9200/tcp",
		envs: []string{
			fmt.Sprintf("ELASTIC_PASSWORD=%s", UserPassword),
			"discovery.type=single-node",
		},
		getConf: func(c containerInfo) string {
			return fmt.Sprintf(`
servers = ["http://%s:%s@%s:%s"]
interval = "2s"
local = false
cluster_health = true
cluster_stats = true
indices_include = ["_all"]
indices_level = "shards"
`, c.User, c.Password, c.Host, c.Port)
		},
		serviceReady: func(ipt *Input) error {
			if len(ipt.Servers) == 0 {
				return nil
			}

			if req, err := http.NewRequest(http.MethodPut, fmt.Sprintf("%s/demo_index", ipt.Servers[0]), nil); err != nil {
				return err
			} else if res, err := http.DefaultClient.Do(req); err != nil {
				return err
			} else {
				body, err := ioutil.ReadAll(res.Body)
				if err != nil {
					return err
				}
				defer res.Body.Close()
				bodyContent := string(body)

				if strings.Contains(bodyContent, "acknowledged") && !strings.Contains(bodyContent, "error") {
					return nil
				} else {
					return fmt.Errorf("create index failed: %s", bodyContent)
				}
			}
		},
		validate: assertMeasurements,
	},
	{
		name: "elasticsearch-ok",
		images: []string{
			fmt.Sprintf("%selasticsearch:8.6.0", RepoURL),
		},
		bindingPort: "9200/tcp",
		envs: []string{
			fmt.Sprintf("ELASTIC_PASSWORD=%s", UserPassword),
			"discovery.type=single-node",
			"xpack.security.enabled=false",
			"ES_JAVA_OPTS=-Xms1g -Xmx1g",
		},
		getConf: func(c containerInfo) string {
			return fmt.Sprintf(`
servers = ["http://%s:%s@%s:%s"]
interval = "2s"
local = false
cluster_health = true
cluster_stats = true
indices_include = ["_all"]
indices_level = "shards"
`, c.User, c.Password, c.Host, c.Port)
		},
		serviceReady: func(ipt *Input) error {
			if len(ipt.Servers) == 0 {
				return nil
			}
			client := &http.Client{
				Transport: &http.Transport{
					TLSClientConfig: &tls.Config{
						InsecureSkipVerify: true,
					},
				},
			}
			if req, err := http.NewRequest(http.MethodPut, fmt.Sprintf("%s/demo_index", ipt.Servers[0]), nil); err != nil {
				return err
			} else if res, err := client.Do(req); err != nil {
				return err
			} else {
				body, err := ioutil.ReadAll(res.Body)
				if err != nil {
					return err
				}
				defer res.Body.Close()
				bodyContent := string(body)

				if strings.Contains(bodyContent, "acknowledged") && !strings.Contains(bodyContent, "error") {
					return nil
				} else {
					return fmt.Errorf("create index failed: %s", bodyContent)
				}
			}
		},
		validate: assertMeasurements,
	},
}

// getPool generates pool to connect to Docker.
func (cs *caseSpec) getPool(r *testutils.RemoteInfo) (*dt.Pool, error) {
	dockerTCP := r.TCPURL()

	cs.t.Logf("get remote: %+#v, TCP: %s", r, dockerTCP)

	p, err := dt.NewPool(dockerTCP)
	if err != nil {
		return nil, err
	}

	err = p.Client.Ping()
	if err != nil {
		if r.Host != "0.0.0.0" {
			return nil, err
		}
		// use default docker service
		cs.t.Log("try default docker")
		p, err = dt.NewPool("")
		if err != nil {
			return nil, err
		} else {
			if err = p.Client.Ping(); err != nil {
				return nil, err
			}
		}
	}

	return p, nil
}

func (cs *caseSpec) getInput(c containerInfo) (*Input, error) {
	ipt := defaultInput()

	if _, err := toml.Decode(cs.getConf(c), ipt); err != nil {
		return nil, err
	}

	return ipt, nil
}

func (cs *caseSpec) run() error {
	var err error
	var containerName string
	r := testutils.GetRemote()
	start := time.Now()

	// set pool
	if cs.pool, err = cs.getPool(r); err != nil {
		return err
	}

	hostname := "unknown-hostname"
	// set containerName
	if name, err := os.Hostname(); err != nil {
		cs.t.Logf("get hostname failed: %s, ignored", err)
	} else {
		hostname = name
	}
	containerName = fmt.Sprintf("%s.%s", hostname, cs.name)

	// remove the container if exist.
	if err := cs.pool.RemoveContainerByName(containerName); err != nil {
		return err
	}

	// check image valid
	images := strings.Split(cs.image, ":")
	if len(images) != 2 {
		return fmt.Errorf("invalid image %s", cs.image)
	}

	// check binding port
	if len(cs.bindingPort) == 0 {
		return fmt.Errorf("binding port is empty")
	}

	// run a container
	if cs.resource, err = cs.pool.RunWithOptions(&dt.RunOptions{
		// specify container image & tag
		Repository:   images[0],
		Tag:          images[1],
		ExposedPorts: []string{string(cs.bindingPort)},
		Name:         containerName,

		// container run-time envs
		Env: cs.envs,
	}, func(c *docker.HostConfig) {
		c.RestartPolicy = docker.RestartPolicy{Name: "no"}
	}); err != nil {
		return err
	}

	// setup container
	if err := setupContainer(cs.pool, cs.resource); err != nil {
		return err
	}

	hostPort := cs.resource.GetHostPort(string(cs.bindingPort))
	_, port, err := net.SplitHostPort(hostPort)
	if err != nil {
		return fmt.Errorf("get host port error: %w", err)
	}

	cs.t.Logf("check service(%s:%s)...", r.Host, port)
	if cs.serviceOK != nil {
		if !cs.serviceOK(port) {
			return fmt.Errorf("service failed to serve")
		}
	} else if !r.PortOK(port, 5*time.Minute) {
		return fmt.Errorf("service port checking failed")
	}

	info := containerInfo{
		User:     User,
		Password: UserPassword,
		Host:     r.Host,
		Port:     port,
	}

	// set input
	if cs.ipt, err = cs.getInput(info); err != nil {
		return err
	}

	cs.feeder = io.NewMockedFeeder()
	cs.ipt.feeder = cs.feeder

	if cs.serviceReady != nil {
		if err := cs.serviceReady(cs.ipt); err != nil {
			return err
		}
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
	ps, err := cs.feeder.AnyPoints()
	if err != nil {
		return err
	}

	cs.cr.AddField("point_latency", int64(time.Since(start)))
	cs.cr.AddField("point_count", len(ps))

	cs.t.Logf("get %d points", len(ps))
	if cs.validate != nil {
		if err := cs.validate(ps, cs); err != nil {
			return err
		}
	}

	cs.t.Logf("stop input...")
	cs.ipt.Terminate()

	cs.t.Logf("exit...")
	wg.Wait()

	return nil
}

type containerInfo struct {
	Host     string
	Port     string
	User     string
	Password string
}

// build test cases based on case item
func buildCases(t *testing.T, configs []caseItem) ([]*caseSpec, error) {
	t.Helper()

	var cases []*caseSpec

	for _, config := range configs {
		for _, img := range config.images {
			parts := strings.Split(img, ":")
			tag := "latest"
			if len(parts) == 2 {
				tag = parts[1]
			}

			caseSpecItem := &caseSpec{
				t:     t,
				name:  fmt.Sprintf("%s.%s", config.name, fmt.Sprintf("elasticsearch.%s", tag)),
				envs:  config.envs,
				image: img,

				validate:     config.validate,
				getConf:      config.getConf,
				serviceReady: config.serviceReady,
				bindingPort:  config.bindingPort,
				cr: &testutils.CaseResult{
					Name: t.Name(),
					Case: config.name,
					ExtraTags: map[string]string{
						"image": img,
					},
				},
			}

			if config.serviceOK != nil {
				caseSpecItem.serviceOK = config.serviceOK
			} else {
				caseSpecItem.serviceOK = func(port string) bool {
					host := testutils.GetRemote().Host
					ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
					defer cancel()
					ticker := time.NewTicker(time.Second)
					defer ticker.Stop()
					for {
						select {
						case <-ctx.Done():
							return false
						case <-ticker.C:
							url := net.JoinHostPort(host, port)
							if res, err := http.Get("http://" + url); err == nil && res.StatusCode == http.StatusOK {
								return true
							}
						}
					}
				}
			}

			cases = append(cases, caseSpecItem)
		}
	}

	return cases, nil
}

func assertMeasurements(pts []*point.Point, cs *caseSpec) error {
	pointMap := map[string]bool{}
	mtMap := map[string]struct {
		measurement    inputs.Measurement
		optionalFields []string
		optionalTags   []string
		extraTags      map[string]string
	}{
		"elasticsearch_node_stats": {
			measurement: nodeStatsMeasurement{},
			extraTags: map[string]string{
				"host":                                   "",
				"node_attribute_ml.allocated_processors": "",
				"node_attribute_ml.allocated_processors_double": "",
				"node_attribute_ml.max_jvm_size":                "",
				"node_roles":                                    "",
			},
			optionalFields: []string{
				"fs_io_stats_total_write_operations",
				"thread_pool_transform_indexing_queue",
				"fs_io_stats_devices_0_write_kilobytes",
				"thread_pool_transform_indexing_rejected",
				"fs_io_stats_devices_0_write_operations",
				"fs_io_stats_total_operations",
				"fs_io_stats_devices_0_read_operations",
				"fs_io_stats_total_read_kilobytes",
				"fs_io_stats_total_read_operations",
				"fs_io_stats_total_write_kilobytes",
				"fs_io_stats_devices_0_operations",
				"fs_io_stats_devices_0_read_kilobytes",
			},
			optionalTags: []string{
				"node_attribute_ml.enabled",
				"node_attribute_ml.max_open_jobs",
				"node_attribute_ml.allocated_processors",
				"node_attribute_ml.allocated_processors_double",
				"node_attribute_ml.max_jvm_size",
				"node_roles",
			},
		},

		"elasticsearch_indices_stats": {
			measurement: indicesStatsMeasurement{},
			extraTags: map[string]string{
				"host": "",
			},
		},
		"elasticsearch_cluster_stats": {
			measurement: clusterStatsMeasurement{},
			extraTags: map[string]string{
				"host": "",
			},
		},
		"elasticsearch_cluster_health": {
			measurement: clusterHealthMeasurement{},
			extraTags: map[string]string{
				"host": "",
				"name": "",
			},
		},
	}

	for _, pt := range pts {
		name := pt.Name()
		if _, ok := pointMap[name]; ok {
			continue
		}
		if m, ok := mtMap[name]; ok {
			msgs := inputs.CheckPoint(pt,
				inputs.WithDoc(m.measurement),
				inputs.WithOptionalFields(m.optionalFields...),
				inputs.WithExtraTags(cs.ipt.Tags),
				inputs.WithOptionalTags(m.optionalTags...),
				inputs.WithExtraTags(m.extraTags),
			)
			for _, msg := range msgs {
				cs.t.Logf("check measurement %s failed: %+#v", name, msg)
			}
			pointMap[name] = true
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

	for m := range mtMap {
		if _, ok := pointMap[m]; !ok {
			return fmt.Errorf("measurement %s not found", m)
		}
	}

	return nil
}

// setupContainer sets up the container for the given Pool and Resource.
func setupContainer(p *dt.Pool, resource *dt.Resource) error {
	return nil
}

func TestIntegrate(t *testing.T) {
	if !testutils.CheckIntegrationTestingRunning() {
		t.Skip()
	}

	t.Helper()
	start := time.Now()
	cases, err := buildCases(t, cases)
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
				// t.Parallel()
				tc.t = t
				caseStart := time.Now()

				t.Logf("testing %s...", tc.name)

				if err := tc.run(); err != nil {
					tc.cr.Status = testutils.TestFailed
					tc.cr.FailedMessage = err.Error()

					assert.NoError(t, err)
				} else {
					tc.cr.Status = testutils.TestPassed
				}

				tc.cr.Cost = time.Since(caseStart)

				assert.NoError(t, testutils.Flush(tc.cr))

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
