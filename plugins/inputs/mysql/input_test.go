// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package mysql

import (
	"fmt"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/BurntSushi/toml"
	dt "github.com/ory/dockertest/v3"
	"github.com/ory/dockertest/v3/docker"
	"github.com/stretchr/testify/assert"

	"github.com/GuanceCloud/cliutils/point"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/testutils"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
)

const RepoURL = "pubrepo.jiagouyun.com/image-repo-for-testing/mysql/"

var (
	MySQLPassword = "Abc123!"
	MySQL5UserSQL = fmt.Sprintf("CREATE USER 'datakit'@'%%' IDENTIFIED BY '%s';", MySQLPassword)
	MySQL8UserSQL = fmt.Sprintf("CREATE USER 'datakit'@'%%' IDENTIFIED WITH mysql_native_password by '%s';", MySQLPassword)
	MySQLGrantSQL = `
CREATE DATABASE test;
CREATE TABLE test.user (id int, name varchar(50));
GRANT PROCESS ON *.* TO 'datakit'@'%';
GRANT SELECT ON *.* TO 'datakit'@'%';
show databases like 'performance_schema';
GRANT SELECT ON performance_schema.* TO 'datakit'@'%';
GRANT SELECT ON mysql.user TO 'datakit'@'%';
GRANT replication client on *.*  to 'datakit'@'%';
`
)

type validateFunc func(pts []*point.Point, cs *caseSpec) error

type caseSpec struct {
	t *testing.T

	name        string
	repo        string
	repoTag     string
	envs        []string
	servicePort string

	validate validateFunc

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
	port := testutils.RandPort("tcp")
	resource, err := p.RunWithOptions(&dt.RunOptions{
		// specify container image & tag
		Repository: cs.repo,
		Tag:        cs.repoTag,

		// port binding
		PortBindings: map[docker.Port][]docker.PortBinding{
			"3306/tcp": {{HostIP: "0.0.0.0", HostPort: fmt.Sprintf("%d", port)}},
		},

		Name: containerName,

		// container run-time envs
		Env: cs.envs,
	}, func(c *docker.HostConfig) {
		c.RestartPolicy = docker.RestartPolicy{Name: "no"}
	})
	if err != nil {
		return err
	}

	cs.pool = p
	cs.resource = resource
	cs.ipt.Port = port

	time.Sleep(10 * time.Second)

	if err := setupContainer(p, resource); err != nil {
		return err
	}

	cs.t.Logf("check service(%s:%d)...", r.Host, port)
	if !r.PortOK(fmt.Sprintf("%d", port), time.Minute) {
		return fmt.Errorf("service checking failed")
	}

	// wait a period of time to ensure that the MySQL service is available.
	time.Sleep(10 * time.Second)

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
				}, dt.ExecOptions{
					StdOut: os.Stdout,
					StdErr: os.Stderr,
				})
			}
			time.Sleep(1 * time.Second)
		}
	}()

	// wait data
	start = time.Now()
	cs.t.Logf("wait points...")
	pts := []*point.Point{}
	// merge pts
	for i := 0; i < 5; i++ {
		ps, err := cs.feeder.AnyPoints()
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
`, testutils.GetRemote().Host, MySQLPassword),
			validate: assertMeasurements,
		},
		{
			name: "mysql-no-dbm",
			conf: fmt.Sprintf(`
host = "%s"
user = "datakit"
pass = "%s"
port = 0 
innodb = true
interval = "1s"
dbm = false 
[dbm_metric]
  enabled = true
[dbm_sample]
  enabled = true  
[dbm_activity]
  enabled = true  
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

				servicePort: fmt.Sprintf("%d", ipt.Port),

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

func assertMeasurements(pts []*point.Point, cs *caseSpec) error {
	pointMap := map[string]bool{}
	mtMap := map[string]measurementInfo{
		"mysql": {
			measurement: &baseMeasurement{},
			optionalFields: []string{
				"Key_buffer_bytes_used",
				"Binlog_space_usage_bytes",
				"Key_buffer_size",
				"Key_cache_utilization",
				"Key_buffer_bytes_unflushed",
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
			measurement:    &innodbMeasurement{},
			optionalFields: []string{"log_padded"},
		},
		"mysql_table_schema": {
			measurement: &tbMeasurement{},
		},
		"mysql_user_status": {
			measurement: &userMeasurement{},
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
		name := string(pt.Name())
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

	for m := range mtMap {
		if _, ok := pointMap[m]; !ok {
			return fmt.Errorf("measurement %s not found", m)
		}
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

// setupContainer sets up the container for the given Pool and Resource.
func setupContainer(p *dt.Pool, resource *dt.Resource) error {
	mysqlConfCmd := `cat > /etc/mysql/conf.d/mysql.cnf <<EOF
[mysqld]
	performance_schema = on
	max_digest_length = 4096
	performance_schema_max_digest_length = 4096
	performance_schema_max_sql_text_length = 4096
	performance-schema-consumer-events-statements-current = on
	performance-schema-consumer-events-waits-current = on
	performance-schema-consumer-events-statements-history-long = on
	performance-schema-consumer-events-statements-history = on	
EOF
`

	resource.Exec([]string{
		"/bin/sh", "-c", mysqlConfCmd,
	}, dt.ExecOptions{
		StdOut: os.Stdout,
		StdErr: os.Stderr,
	})

	if err := p.Client.RestartContainer(resource.Container.ID, 30); err != nil {
		return err
	}

	return nil
}

func TestMySQLInput(t *testing.T) {
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
