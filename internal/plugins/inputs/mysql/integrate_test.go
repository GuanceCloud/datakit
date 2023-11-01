// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package mysql

import (
	"context"
	"database/sql"
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/BurntSushi/toml"
	"github.com/go-sql-driver/mysql"
	dt "github.com/ory/dockertest/v3"
	"github.com/ory/dockertest/v3/docker"
	"github.com/stretchr/testify/assert"

	"github.com/GuanceCloud/cliutils/point"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/testutils"
)

const RepoURL = "pubrepo.guance.com/image-repo-for-testing/mysql/"

var (
	MySQLPassword = "Abc123!"
	MySQL5UserSQL = fmt.Sprintf("CREATE USER 'datakit'@'%%' IDENTIFIED BY '%s';", MySQLPassword)
	MySQL8UserSQL = fmt.Sprintf("CREATE USER 'datakit'@'%%' IDENTIFIED WITH caching_sha2_password by '%s';", MySQLPassword)
	MySQLGrantSQL = `
CREATE DATABASE test;
CREATE TABLE test.user (id int, name varchar(50), value float(5,2));
INSERT INTO test.user(id, name, value) values(1, 'ross', 22.80);
GRANT PROCESS ON *.* TO 'datakit'@'%';
GRANT SELECT ON *.* TO 'datakit'@'%';
show databases like 'performance_schema';
GRANT SELECT ON performance_schema.* TO 'datakit'@'%';
GRANT SELECT ON mysql.user TO 'datakit'@'%';
GRANT replication client on *.*  to 'datakit'@'%';
`
)

type (
	validateFunc  func(pts []*point.Point, cs *caseSpec) error
	serviceOKFunc func(t *testing.T, port string) bool
)

type caseSpec struct {
	t *testing.T

	name        string
	repo        string
	repoTag     string
	envs        []string
	servicePort string

	validate  validateFunc
	serviceOK serviceOKFunc

	ipt    *Input
	feeder *io.MockedFeeder

	pool     *dt.Pool
	resource *dt.Resource

	cr *testutils.CaseResult
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

func (cs *caseSpec) run() error {
	r := testutils.GetRemote()
	start := time.Now()
	p, err := cs.getPool(r)
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

	resource, err := p.RunWithOptions(&dt.RunOptions{
		// specify container image & tag
		Repository: cs.repo,
		Tag:        cs.repoTag,

		ExposedPorts: []string{cs.servicePort},
		Name:         containerName,

		User: "root",
		// container run-time envs
		Env: cs.envs,
	}, func(c *docker.HostConfig) {
		c.RestartPolicy = docker.RestartPolicy{Name: "no"}
	})
	if err != nil {
		return err
	}

	hostPort := resource.GetHostPort(cs.servicePort)
	_, port, err := net.SplitHostPort(hostPort)
	if err != nil {
		return fmt.Errorf("get host port error: %w", err)
	}

	cs.pool = p
	cs.resource = resource
	if portNumber, err := strconv.Atoi(port); err != nil {
		return fmt.Errorf("get host port error: %w", err)
	} else {
		cs.ipt.Port = portNumber
	}
	cs.t.Logf("check service(%s:%s)...", r.Host, port)
	if cs.serviceOK != nil {
		if !cs.serviceOK(cs.t, port) {
			return fmt.Errorf("service failed to serve")
		}
	} else if !r.PortOK(port, 5*time.Minute) {
		return fmt.Errorf("service port checking failed")
	}

	if err := initMySQL(resource, cs); err != nil {
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

	wg.Add(1)
	go func() {
		defer wg.Done()
		for {
			select {
			case <-cs.ipt.semStop.Wait():
				return
			default:
				resource.Exec([]string{
					"/bin/sh", "-c", fmt.Sprintf(`mysql -uroot -p%s -e "%s"`, MySQLPassword, "select * from test.user"),
				}, dt.ExecOptions{})
			}
			time.Sleep(time.Millisecond)
		}
	}()

	// wait data
	start = time.Now()
	cs.t.Logf("wait points...")
	pts := []*point.Point{}
	// merge pts
	for i := 0; i < 4; i++ {
		ps, err := cs.feeder.AnyPoints(10 * time.Second)
		if err != nil {
			return err
		}
		pts = append(pts, ps...)
	}

	cs.cr.AddField("point_latency", int64(time.Since(start)))
	cs.cr.AddField("point_count", len(pts))

	cs.t.Logf("get %d points", len(pts))
	if cs.validate != nil {
		if err := cs.validate(pts, cs); err != nil {
			return err
		}
	}

	cs.t.Logf("stop input...")
	cs.ipt.Terminate()

	cs.t.Logf("exit...")
	wg.Wait()

	return nil
}

func buildCases(t *testing.T) ([]*caseSpec, error) {
	t.Helper()

	bases := []struct {
		name     string
		conf     string
		validate validateFunc
	}{
		{
			name: "mysql-ok",
			conf: fmt.Sprintf(`
host = "%s"
user = "datakit"
pass = "%s"
port = 0
innodb = true
interval = "1s"
dbm = true
[dbm_metric]
  enabled = true
[dbm_sample]
  enabled = true
[dbm_activity]
  enabled = true
[[custom_queries]]
  sql = "SELECT id, name, value FROM test.user;"
  metric = "mysql_custom"
  tags = ["name"]
  fields = ["id", "value"]
`, testutils.GetRemote().Host, MySQLPassword),
			validate: assertMeasurements,
		},
	}

	images := [][]string{
		{RepoURL + "mysql", "5.7"},
		{RepoURL + "mysql", "8.0"},
	}

	var cases []*caseSpec

	for _, img := range images {
		for _, base := range bases {
			feeder := io.NewMockedFeeder()
			ipt := defaultInput()
			ipt.feeder = feeder

			_, err := toml.Decode(base.conf, ipt)
			assert.NoError(t, err)

			envs := []string{
				fmt.Sprintf("MYSQL_ROOT_PASSWORD=%s", MySQLPassword),
			}

			cases = append(cases, &caseSpec{
				t:      t,
				ipt:    ipt,
				name:   fmt.Sprintf("%s.%s", base.name, "mysql"+img[1]),
				feeder: feeder,
				envs:   envs,

				repo:    img[0],
				repoTag: img[1],

				servicePort: "3306/tcp",

				validate: base.validate,

				cr: &testutils.CaseResult{
					Name: t.Name(),
					Case: base.name,
					ExtraTags: map[string]string{
						"image":         img[0],
						"image_tag":     img[1],
						"remote_server": ipt.Host,
					},
				},
				serviceOK: func(t *testing.T, port string) bool {
					t.Helper()
					cfg := mysql.Config{
						AllowNativePasswords: true,
						User:                 "root",
						Passwd:               MySQLPassword,
						Net:                  "tcp",
						Addr:                 fmt.Sprintf("%s:%s", testutils.GetRemote().Host, port),
					}

					ctx, cancel := context.WithTimeout(context.Background(), 1*time.Minute)
					defer cancel()
					ticker := time.NewTicker(time.Second)
					defer ticker.Stop()
					dsn := cfg.FormatDSN()
					for {
						select {
						case <-ctx.Done():
							return false
						case <-ticker.C:
							db, err := sql.Open("mysql", dsn)
							if err != nil {
								continue
							}
							if err := db.Ping(); err != nil {
								t.Logf("ping db error: %s", err.Error())
								continue
							} else {
								return true
							}
						}
					}
				},
			})
		}
	}

	return cases, nil
}

type measurementInfo struct {
	measurement    inputs.Measurement
	optionalFields []string
	optionalTags   []string
	extraTags      map[string]string
}

type customMeasurement struct {
	name     string
	tags     map[string]string
	fields   map[string]interface{}
	election bool
}

// Point implement MeasurementV2.
func (m *customMeasurement) Point() *point.Point {
	opts := point.DefaultMetricOptions()

	return point.NewPointV2(m.name,
		append(point.NewTags(m.tags), point.NewKVs(m.fields)...),
		opts...)
}

//nolint:lll,funlen
func (m *customMeasurement) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name: "mysql_custom",
		Type: "metric",
		Fields: map[string]interface{}{
			"id": &inputs.FieldInfo{
				DataType: inputs.Float,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     "id",
			},
			"value": &inputs.FieldInfo{
				DataType: inputs.Float,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     "value",
			},
		},
		Tags: map[string]interface{}{
			"name":   &inputs.TagInfo{Desc: "name"},
			"server": &inputs.TagInfo{Desc: "server name"},
		},
	}
}

func assertMeasurements(pts []*point.Point, cs *caseSpec) error {
	pointMap := map[string]bool{}
	mtMap := map[string]measurementInfo{
		"mysql": {
			measurement: &baseMeasurement{},
			optionalFields: []string{
				"Binlog_space_usage_bytes",
				"Qcache_not_cached",
				"Qcache_lowmem_prunes",
				"Qcache_free_blocks",
				"Qcache_hits",
				"Qcache_queries_in_cache",
				"Qcache_inserts",
				"Qcache_total_blocks",
				"query_cache_size",
				"Qcache_free_memory",
			},
		},
		"mysql_schema": {
			measurement:    &schemaMeasurement{},
			optionalFields: []string{"query_run_time_avg"},
		},
		"mysql_innodb": {
			measurement: &innodbMeasurement{},
			optionalFields: []string{
				"log_padded",
				"mem_recovery_system",
				"mem_dictionary",
				"mem_additional_pool",
				"pending_log_writes",
				"mem_lock_system",
				"mem_page_hash",
				"mem_adaptive_hash",
				"mem_file_system",
				"pending_checkpoint_writes",
				"mem_thread_hash",
				"mem_total",
				"pending_aio_sync_ios",
				"pending_aio_log_ios",
				"pending_ibuf_aio_reads",
			},
		},
		"mysql_table_schema": {
			measurement: &tbMeasurement{},
		},
		"mysql_user_status": {
			measurement: &userMeasurement{},
		},
		"mysql_custom": {
			measurement: &customMeasurement{},
		},
	}

	// dbm metric
	if cs.ipt.Dbm {
		mtMap["mysql_dbm_metric"] = measurementInfo{
			measurement: &dbmStateMeasurement{},
			extraTags: map[string]string{
				"status": "info",
			},
			optionalTags: []string{"schema_name"},
		}

		mtMap["mysql_dbm_sample"] = measurementInfo{
			measurement: &dbmSampleMeasurement{},
			extraTags: map[string]string{
				"status": "info",
			},
		}
		mtMap["mysql_dbm_activity"] = measurementInfo{
			measurement: &dbmActivityMeasurement{},
			extraTags: map[string]string{
				"status": "info",
			},
		}
	}

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
		if _, ok := pointMap[m]; !ok {
			missingMeasurements = append(missingMeasurements, m)
		}
	}

	if len(missingMeasurements) > 0 {
		return fmt.Errorf("measurements not found: %s", strings.Join(missingMeasurements, ","))
	}

	return nil
}

func initMySQL(resource *dt.Resource, cs *caseSpec) error {
	createUserSQL := MySQL5UserSQL
	if cs.repoTag[0] == '8' {
		createUserSQL = MySQL8UserSQL
	}

	configConsumerSQL := `
UPDATE performance_schema.setup_consumers SET enabled='YES' WHERE name LIKE 'events_statements_%';
UPDATE performance_schema.setup_consumers SET enabled='YES' WHERE name = 'events_waits_current';

`

	initSQL := createUserSQL + MySQLGrantSQL + configConsumerSQL
	mysqlInitCmd := fmt.Sprintf(`mysql -uroot -p%s -e "%s"`, MySQLPassword, initSQL)
	_, err := resource.Exec([]string{
		"/bin/sh", "-c", mysqlInitCmd,
	}, dt.ExecOptions{
		StdOut: os.Stdout,
		StdErr: os.Stderr,
	})
	if err != nil {
		return err
	}

	return initDbm(resource, cs.repoTag[0])
}

func initDbm(resource *dt.Resource, version byte) error {
	sql := ""
	switch version {
	case '5':
		sql = `
GRANT REPLICATION CLIENT ON *.* TO datakit@'%' WITH MAX_USER_CONNECTIONS 5;
GRANT PROCESS ON *.* TO datakit@'%';
`
	case '8':
		sql = `
ALTER USER datakit@'%' WITH MAX_USER_CONNECTIONS 5;
GRANT REPLICATION CLIENT ON *.* TO datakit@'%';
GRANT PROCESS ON *.* TO datakit@'%';
`
	default:
		return fmt.Errorf("invalid mysql version %s", string(version))
	}

	sql += `
	CREATE SCHEMA IF NOT EXISTS datakit;
	GRANT EXECUTE ON datakit.* to datakit@'%';
	GRANT CREATE TEMPORARY TABLES ON datakit.* TO datakit@'%';

	DELIMITER $$
CREATE PROCEDURE datakit.explain_statement(IN query TEXT)
    SQL SECURITY DEFINER
BEGIN
    SET @explain := CONCAT('EXPLAIN FORMAT=json ', query);
    PREPARE stmt FROM @explain;
    EXECUTE stmt;
    DEALLOCATE PREPARE stmt;
END $$
DELIMITER ;

UPDATE performance_schema.setup_consumers SET enabled='YES' WHERE name LIKE 'events_statements_%';
UPDATE performance_schema.setup_consumers SET enabled='YES' WHERE name = 'events_waits_current';

`

	sqlCmd := fmt.Sprintf(`mysql -uroot -p%s -e "%s"`, MySQLPassword, sql)
	_, err := resource.Exec([]string{
		"/bin/sh", "-c", sqlCmd,
	}, dt.ExecOptions{
		StdOut: os.Stdout,
		StdErr: os.Stderr,
	})

	return err
}

func TestIntegrate(t *testing.T) {
	if !testutils.CheckIntegrationTestingRunning() {
		t.Skip()
	}

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
				// t.Parallel() // Should not be parallel, if so, it would dead and timeout due to junk machine.
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
