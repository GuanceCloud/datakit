// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package dialtesting

import (
	"context"
	"fmt"
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
	UserPassword = "Abc123!"
	RegionID     = "regionID"
	DialAK       = "dialak"
	DialSK       = "dialsk"
	RepoURL      = "pubrepo.jiagouyun.com/image-repo-for-testing/dialtesting/"
)

type (
	validateFunc     func(pts []*point.Point, cs *caseSpec) error
	getConfFunc      func(c containerInfo) string
	serviceReadyFunc func(ipt *Input) error
	serviceOKFunc    func(t *testing.T, port int) bool
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

	ipt           *Input
	feeder        *io.MockedFeeder
	collectPoints func(*caseSpec) []*point.Point

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

type caseConfig struct {
	name               string
	images             []string
	checkedMeasurement []string
}

func generateCase(config *caseConfig) caseItem {
	images := []string{}
	for _, image := range config.images {
		images = append(images, fmt.Sprintf("%s%s", RepoURL, image))
	}
	return caseItem{
		name:        config.name,
		images:      images,
		bindingPort: "9538/tcp",
		envs: []string{
			fmt.Sprintf("MYSQL_ROOT_PASSWORD=%s", UserPassword),
			fmt.Sprintf("DIAL_AK=%s", DialAK),
			fmt.Sprintf("DIAL_SK=%s", DialSK),
			fmt.Sprintf("REGION_ID=%s", RegionID),
		},
		getConf: func(c containerInfo) string {
			return fmt.Sprintf(`
server = "http://%s:%s"
pull_interval = "10s"
time_out = "1m"
workers = 6
region_id = "%s"
ak = "%s"
sk = "%s"
`, c.Host, c.Port, RegionID, DialAK, DialSK)
		},
		serviceReady: func(ipt *Input) error {
			return nil
		},
		validate: assertSelectedMeasurments(config.checkedMeasurement),
	}
}

var cases = []caseItem{
	generateCase(&caseConfig{
		name:               "http-test-ok",
		images:             []string{"dialtesting:0.0.1"},
		checkedMeasurement: []string{""},
	}),
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

	// get port
	port := testutils.RandPort("tcp")

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
		Repository: images[0],
		Tag:        images[1],

		// port binding
		PortBindings: map[docker.Port][]docker.PortBinding{
			cs.bindingPort: {{HostIP: "0.0.0.0", HostPort: fmt.Sprintf("%d", port)}},
		},

		Name: containerName,

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

	cs.t.Logf("check service(%s:%d)...", r.Host, port)
	if cs.serviceOK != nil {
		if !cs.serviceOK(cs.t, port) {
			return fmt.Errorf("service failed to serve")
		}
	} else if !r.PortOK(fmt.Sprintf("%d", port), 5*time.Minute) {
		return fmt.Errorf("service port checking failed")
	}

	info := containerInfo{
		Password: UserPassword,
		Host:     r.Host,
		Port:     fmt.Sprintf("%d", port),
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

	var ps []*point.Point
	if cs.collectPoints != nil {
		ps = cs.collectPoints(cs)
	} else {
		ps, err = cs.feeder.AnyPoints()
		if err != nil {
			return err
		}
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
				name:  fmt.Sprintf("%s.%s", config.name, fmt.Sprintf("%s.%s", inputName, tag)),
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
				collectPoints: func(cs *caseSpec) []*point.Point {
					ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
					defer cancel()
					pts := make([]*point.Point, 0)
				outer:
					for {
						if ps, err := cs.feeder.AnyPoints(); err != nil {
							continue
						} else {
							pts = append(pts, ps...)
						}

						if len(pts) >= 4 {
							break
						}

						select {
						case <-ctx.Done():
							break outer
						default:
						}
					}
					return pts
				},
			}

			if config.serviceOK != nil {
				caseSpecItem.serviceOK = config.serviceOK
			} else {
				caseSpecItem.serviceOK = func(t *testing.T, port int) bool {
					t.Helper()
					return true
				}
			}

			cases = append(cases, caseSpecItem)
		}
	}

	return cases, nil
}

func assertSelectedMeasurments(selected []string) func(pts []*point.Point, cs *caseSpec) error {
	mtMap := map[string]struct {
		measurement    inputs.Measurement
		optionalFields []string
		optionalTags   []string
		extraTags      map[string]string
	}{
		"http_dial_testing": {
			measurement:    &httpMeasurement{},
			optionalTags:   []string{},
			optionalFields: []string{},
		},
		"tcp_dial_testing": {
			measurement:    &tcpMeasurement{},
			optionalFields: []string{"traceroute"},
		},
		"icmp_dial_testing": {
			measurement:    &icmpMeasurement{},
			optionalFields: []string{"traceroute"},
		},
		"websocket_dial_testing": {
			measurement: &websocketMeasurement{},
		},
	}

	return func(pts []*point.Point, cs *caseSpec) error {
		pointMap := map[string]bool{}
		for _, pt := range pts {
			name := string(pt.Name())
			if _, ok := pointMap[name]; ok {
				continue
			}

			if m, ok := mtMap[name]; ok {
				extraTags := map[string]string{}

				for k, v := range m.extraTags {
					extraTags[k] = v
				}
				msgs := inputs.CheckPoint(pt,
					inputs.WithDoc(m.measurement),
					inputs.WithOptionalFields(m.optionalFields...),
					inputs.WithExtraTags(cs.ipt.Tags),
					inputs.WithOptionalTags(m.optionalTags...),
					inputs.WithExtraTags(extraTags),
				)
				for _, msg := range msgs {
					cs.t.Logf("check measurement %s failed: %+#v", name, msg)
				}
				pointMap[name] = true
			} else {
				continue
			}
		}

		for m := range mtMap {
			for _, item := range selected {
				if m == item {
					if _, ok := pointMap[m]; !ok {
						return fmt.Errorf("measurement %s not found", m)
					}
				}
			}
		}

		return nil
	}
}

// setupContainer sets up the container for the given Pool and Resource.
func setupContainer(p *dt.Pool, resource *dt.Resource) error {
	ctx, cancel := context.WithTimeout(context.TODO(), 10*time.Second)
	var err error
	defer cancel()
	go func() {
		cmd := `bash /start.sh`
		_, err = resource.Exec([]string{
			"/bin/sh", "-c", cmd,
		}, dt.ExecOptions{
			StdOut: os.Stdout,
			StdErr: os.Stderr,
		})
	}()

	<-ctx.Done()

	return err
}

func TestInput(t *testing.T) {
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

					assert.NoError(t, tc.pool.Purge(tc.resource))
				})
			})
		}(tc)
	}
}
