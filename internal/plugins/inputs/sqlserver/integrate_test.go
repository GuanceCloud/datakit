// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package sqlserver

import (
	"context"
	"database/sql"
	"fmt"
	"net"
	"os"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/BurntSushi/toml"
	"github.com/GuanceCloud/cliutils/point"
	mssql "github.com/microsoft/go-mssqldb"
	"github.com/microsoft/go-mssqldb/msdsn"
	dt "github.com/ory/dockertest/v3"
	"github.com/ory/dockertest/v3/docker"
	"github.com/stretchr/testify/assert"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/testutils"
)

const (
	User         = "sa"
	UserPassword = "Abc123abC$"
	Database     = "master"
	Table        = ""

	RepoURL = "mcr.microsoft.com/mssql/"
)

var mtMap = map[string]struct {
	measurement    inputs.Measurement
	optionalFields []string
	optionalTags   []string
	extraTags      map[string]string
}{
	"sqlserver_database_files": {
		measurement:    &DatabaseFilesMeasurement{},
		optionalFields: []string{},
	},
	"sqlserver": {
		measurement: &SqlserverMeasurment{},
	},
	"sqlserver_performance": {
		measurement: &Performance{},
	},
	"sqlserver_waitstats": {
		measurement: &WaitStatsCategorized{},
	},
	"sqlserver_database_io": {
		measurement: &DatabaseIO{},
	},
	"sqlserver_schedulers": {
		measurement: &Schedulers{},
	},
	"sqlserver_volumespace": {
		measurement:  &VolumeSpace{},
		optionalTags: []string{"volume_mount_point"},
	},
	"sqlserver_lock_row": {
		measurement: &LockRow{},
		extraTags: map[string]string{
			"status": "",
		},
	},
	"sqlserver_lock_table": {
		measurement: &LockTable{},
		extraTags: map[string]string{
			"status": "",
		},
	},
	"sqlserver_lock_dead": {
		measurement: &LockDead{},
		extraTags: map[string]string{
			"status": "",
		},
	},
	"sqlserver_logical_io": {
		measurement: &LogicalIO{},
		extraTags: map[string]string{
			"status": "",
		},
	},
	"sqlserver_worker_time": {
		measurement: &WorkerTime{},
		extraTags: map[string]string{
			"status": "",
		},
		optionalTags: []string{
			"message",
		},
	},
	"sqlserver_database_size": {
		measurement: &DatabaseSize{},
	},
}

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
		bindingPort: "1433/tcp",
		envs: []string{
			"ACCEPT_EULA=Y",
			fmt.Sprintf("MSSQL_SA_PASSWORD=%s", UserPassword),
		},
		getConf: func(c containerInfo) string {
			return fmt.Sprintf(`
host = "%s:%s"
user = "%s"
password = "%s"
interval = "2s"
database = "%s"
`, c.Host, c.Port, c.User, c.Password, Database)
		},
		serviceReady: func(ipt *Input) error {
			err := ipt.initDB()
			if err != nil {
				return err
			}
			_, err = ipt.query(`
			CREATE TABLE Persons ( PersonID int, LastName varchar(255), FirstName varchar(255), Address varchar(255), City varchar(255));
			`)
			if err != nil {
				return err
			}

			_, err = ipt.query("insert into Persons (LastName, FirstName) values('a', 'b')")
			if err != nil {
				return err
			}

			for i := 0; i < 2; i++ {
				go func() {
					_, err = ipt.query(`
					begin transaction;
					update Persons set FirstName='bb' where LastName='a';
				`)
				}()
			}

			time.Sleep(2 * time.Second)
			return err
		},
		validate: assertSelectedMeasurments(config.checkedMeasurement),
	}
}

// TODO: check lock measurement: sqlserver_worker_time,sqlserver_logical_io
var cases = []caseItem{
	generateCase(&caseConfig{
		name:   "sqlserver-ok",
		images: []string{"server:2017-latest", "server:2019-latest", "server:2022-latest"},
		checkedMeasurement: []string{
			"sqlserver_database_files",
			"sqlserver",
			"sqlserver_performance",
			"sqlserver_waitstats",
			"sqlserver_database_io",
			"sqlserver_schedulers",
			"sqlserver_volumespace",
			"sqlserver_database_size",
			"sqlserver_lock_row",
			"sqlserver_lock_table",
			"sqlserver_lock_dead",
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

		ExposedPorts: []string{cs.bindingPort.Port()},
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

	// collect metric points
	ps, err := cs.feeder.NPoints(400, 5*time.Minute)
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

					connStr := fmt.Sprintf("sqlserver://%s:%s@%s?dial+timeout=3", User, UserPassword, host)
					cfg, _ := msdsn.Parse(connStr)
					conn := mssql.NewConnectorConfig(cfg)
					db := sql.OpenDB(conn)
					defer db.Close()
					ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
					defer cancel()
					ticker := time.NewTicker(time.Second)
					defer ticker.Stop()
					for {
						select {
						case <-ctx.Done():
							return false
						case <-ticker.C:
							if err := db.Ping(); err == nil {
								return true
							} else {
								t.Logf("service check failed, %s, try again", err.Error())
								continue
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

func assertSelectedMeasurments(selected []string) func(pts []*point.Point, cs *caseSpec) error {
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
					cs.t.Logf("[%s] check measurement %s failed: %+#v", cs.t.Name(), name, msg)
				}
				if len(msgs) > 0 {
					return fmt.Errorf("check measurement %s failed: collected points are not as expected ", name)
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

		missingMeasurements := []string{}
		for m := range mtMap {
			if selected == nil {
				if _, ok := pointMap[m]; !ok {
					missingMeasurements = append(missingMeasurements, m)
				}
			} else {
				for _, item := range selected {
					if m == item {
						if _, ok := pointMap[m]; !ok {
							missingMeasurements = append(missingMeasurements, m)
						}
					}
				}
			}
		}

		if len(missingMeasurements) > 0 {
			return fmt.Errorf("measurements not found: %s", strings.Join(missingMeasurements, ","))
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
