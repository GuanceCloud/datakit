// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package postgresql

import (
	"context"
	"fmt"
	"net"
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
	User         = "postgres"
	UserPassword = "Abc123!"
	Database     = "demo"
	Table        = "demo"

	RepoURL = "pubrepo.guance.com/image-repo-for-testing/postgres/"
)

type (
	validateFunc     func(pts []*point.Point, cs *caseSpec) error
	getConfFunc      func(c containerInfo) string
	serviceReadyFunc func(ipt *Input) error
	serviceOKFunc    func(t *testing.T, port string) bool
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
		bindingPort: "5432/tcp",
		envs: []string{
			fmt.Sprintf("POSTGRES_PASSWORD=%s", UserPassword),
		},
		getConf: func(c containerInfo) string {
			return fmt.Sprintf(`
address = "postgres://%s:%s@%s:%s/%s?sslmode=disable"
interval = "2s"
[[relations]]
	relation_name = "%s"
`, c.User, c.Password, c.Host, c.Port, Database, Table)
		},
		serviceReady: func(ipt *Input) error {
			service := &SQLService{
				MaxIdle: 1,
				MaxOpen: 1,
			}
			service.SetAddress(ipt.Address)
			service.Start()

			rows, err := service.Query(fmt.Sprintf("create table %s(id int, primary key(id))", Table))
			if err != nil {
				return err
			}
			rows.Close()
			rows, err = service.Query(fmt.Sprintf("insert into %s(id) values(1)", Table))
			if err != nil {
				return err
			}
			rows.Close()

			rows, err = service.Query("begin")
			if err != nil {
				return err
			}
			rows.Close()

			rows, err = service.Query(fmt.Sprintf("lock table %s in share mode nowait", Table))
			if err != nil {
				return err
			}
			rows.Close()

			// stop service when test is done
			go func() {
				<-ipt.semStop.Wait()
				service.Stop()
			}()

			return nil
		},
		validate: assertSelectedMeasurments(config.checkedMeasurement),
	}
}

// TODO: test postgresql_replication
var cases = []caseItem{
	generateCase(&caseConfig{
		name:   "postgresql-ok-version-larger-than-13",
		images: []string{"postgres:13.11", "postgres:15.3"},
		checkedMeasurement: []string{
			"postgresql",
			"postgresql_lock",
			"postgresql_stat",
			"postgresql_index",
			"postgresql_size",
			"postgresql_statio",
			"postgresql_slru",
			"postgresql_bgwriter",
			"postgresql_connection",
			"postgresql_conflict",
			"postgresql_archiver",
		},
	}),
	generateCase(&caseConfig{
		name:   "postgresql-ok-version-larger-than-9.4",
		images: []string{"postgres:9.4", "postgres:12.15"},
		checkedMeasurement: []string{
			"postgresql",
			"postgresql_lock",
			"postgresql_stat",
			"postgresql_index",
			"postgresql_size",
			"postgresql_statio",
			"postgresql_bgwriter",
			"postgresql_connection",
			"postgresql_conflict",
			"postgresql_archiver",
		},
	}),
	generateCase(&caseConfig{
		name:   "postgresql-ok-version-larger-than-9.2",
		images: []string{"postgres:9.3"},
		checkedMeasurement: []string{
			"postgresql",
			"postgresql_lock",
			"postgresql_stat",
			"postgresql_index",
			"postgresql_size",
			"postgresql_statio",
			"postgresql_bgwriter",
			"postgresql_connection",
			"postgresql_conflict",
		},
	}),
	generateCase(&caseConfig{
		name:   "postgresql-ok-version-larger-than-9.0",
		images: []string{"postgres:9.0"},
		checkedMeasurement: []string{
			"postgresql",
			"postgresql_lock",
			"postgresql_stat",
			"postgresql_index",
			"postgresql_size",
			"postgresql_statio",
			"postgresql_bgwriter",
			"postgresql_connection",
		},
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

		ExposedPorts: []string{string(cs.bindingPort)},

		Name: containerName,

		// container run-time envs
		Env: cs.envs,
	}, func(c *docker.HostConfig) {
		c.RestartPolicy = docker.RestartPolicy{Name: "no"}
	}); err != nil {
		return err
	}

	hostPort := cs.resource.GetHostPort(string(cs.bindingPort))

	_, port, err := net.SplitHostPort(hostPort)
	if err != nil {
		return fmt.Errorf("get host port error: %w", err)
	}

	// setup container
	if err := setupContainer(cs.pool, cs.resource); err != nil {
		return err
	}

	cs.t.Logf("check service(%s:%s)...", r.Host, port)
	if cs.serviceOK != nil {
		if !cs.serviceOK(cs.t, port) {
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
			}

			if config.serviceOK != nil {
				caseSpecItem.serviceOK = config.serviceOK
			} else {
				caseSpecItem.serviceOK = func(t *testing.T, port string) bool {
					t.Helper()
					host := net.JoinHostPort(testutils.GetRemote().Host, fmt.Sprint(port))
					address := fmt.Sprintf("postgres://%s:%s@%s?sslmode=disable", User, UserPassword, host)
					service := &SQLService{Address: address}
					service.Start()
					defer service.Stop() //nolint:errcheck

					ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
					defer cancel()
					ticker := time.NewTicker(time.Second)
					defer ticker.Stop()
					for {
						select {
						case <-ctx.Done():
							return false
						case <-ticker.C:
							rows, err := service.Query(fmt.Sprintf("create database %s", Database))
							if err != nil {
								t.Logf("service check failed: %s, try again", err.Error())
								continue
							}
							defer rows.Close()
							return true
						}
					}
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
		"postgresql": {
			measurement: inputMeasurement{},
			optionalFields: []string{
				"session_time",
				"active_time",
				"idle_in_transaction_time",
				"sessions",
				"sessions_abandoned",
				"sessions_fatal",
				"sessions_killed",
				"temp_bytes",
				"temp_files",
				"deadlocks",
			},
		},
		"postgresql_lock": {measurement: lockMeasurement{}},
		"postgresql_stat": {
			measurement: statMeasurement{},
			optionalFields: []string{
				"vacuum_count",
				"autovacuum_count",
				"analyze_count",
				"autoanalyze_count",
			},
		},
		"postgresql_index": {measurement: indexMeasurement{}},
		"postgresql_size":  {measurement: sizeMeasurement{}},
		"postgresql_statio": {
			measurement: statIOMeasurement{},
			optionalFields: []string{
				"toast_blks_hit",
				"tidx_blks_read",
				"tidx_blks_hit",
				"toast_blks_read",
			},
		},
		"postgresql_replication": {measurement: replicationMeasurement{}},
		"postgresql_slru":        {measurement: slruMeasurement{}},
		"postgresql_bgwriter": {
			measurement: bgwriterMeasurement{},
			optionalFields: []string{
				"checkpoint_write_time",
				"checkpoint_sync_time",
				"buffers_backend_fsync",
			},
		},
		"postgresql_connection": {measurement: connectionMeasurement{}},
		"postgresql_conflict":   {measurement: conflictMeasurement{}},
		"postgresql_archiver":   {measurement: archiverMeasurement{}},
	}

	return func(pts []*point.Point, cs *caseSpec) error {
		pointMap := map[string]bool{}
		for _, pt := range pts {
			name := pt.Name()
			if _, ok := pointMap[name]; ok {
				continue
			}

			if m, ok := mtMap[name]; ok {
				extraTags := map[string]string{
					"host": "host",
				}

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
			// check if tag appended
			if len(cs.ipt.Tags) != 0 {
				cs.t.Logf("checking tags %+#v...", cs.ipt.Tags)

				tags := pt.Tags()
				for k := range cs.ipt.Tags {
					if v := tags.Get(k); v == nil {
						return fmt.Errorf("tag %s not found, got %v", k, tags)
					}
				}
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
