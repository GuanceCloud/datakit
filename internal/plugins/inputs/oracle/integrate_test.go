// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package oracle

import (
	"context"
	"fmt"
	"net"
	"os"
	"strconv"
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

const RepoURL = "pubrepo.guance.com/image-repo-for-testing/oracle/"

var (
	user     = "datakit"
	password = "Abc123!"
	initSQL  = fmt.Sprintf(`
-- Create the datakit user. Replace the password placeholder with a secure password.
    CREATE USER %s IDENTIFIED BY "%s";

    -- Grant access to the datakit user.
    GRANT CONNECT, CREATE SESSION TO datakit;
    GRANT SELECT_CATALOG_ROLE to datakit;
    GRANT SELECT ON DBA_TABLESPACE_USAGE_METRICS TO datakit;
    GRANT SELECT ON DBA_TABLESPACES TO datakit;
    GRANT SELECT ON DBA_USERS TO datakit;
    GRANT SELECT ON SYS.DBA_DATA_FILES TO datakit;
    GRANT SELECT ON V_\$ACTIVE_SESSION_HISTORY TO datakit;
    GRANT SELECT ON V_\$ARCHIVE_DEST TO datakit;
    GRANT SELECT ON V_\$ASM_DISKGROUP TO datakit;
    GRANT SELECT ON V_\$DATABASE TO datakit;
    GRANT SELECT ON V_\$DATAFILE TO datakit;
    GRANT SELECT ON V_\$INSTANCE TO datakit;
    GRANT SELECT ON V_\$LOG TO datakit;
    GRANT SELECT ON V_\$OSSTAT TO datakit;
    GRANT SELECT ON V_\$PGASTAT TO datakit;
    GRANT SELECT ON V_\$PROCESS TO datakit;
    GRANT SELECT ON V_\$RECOVERY_FILE_DEST TO datakit;
    GRANT SELECT ON V_\$RESTORE_POINT TO datakit;
    GRANT SELECT ON V_\$SESSION TO datakit;
    GRANT SELECT ON V_\$SGASTAT TO datakit;
    GRANT SELECT ON V_\$SYSMETRIC TO datakit;
    GRANT SELECT ON V_\$SYSTEM_PARAMETER TO datakit;

    -- Initialize testing data.
    CREATE TABLE students  ( student_id number(10) NOT NULL,  student_name varchar2(40) NOT NULL,  student_age varchar2(10)  );
    INSERT INTO students  (student_id, student_name, student_age)  VALUES  (3, 'Happy', '11');
    exit;
`, user, password)
)

type (
	validateFunc  func(pts []*point.Point, measurementsInfo map[string]measurementInfo, cs *caseSpec) error
	serviceOKFunc func(t *testing.T, cs *caseSpec) bool
)

type caseSpec struct {
	t *testing.T

	name        string
	repo        string
	repoTag     string
	envs        []string
	servicePort string

	validate      validateFunc
	serviceOK     serviceOKFunc
	postServiceOK serviceOKFunc

	measurementsInfo map[string]measurementInfo

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

	// check port
	if !r.PortOK(port, 5*time.Minute) {
		return fmt.Errorf("service port checking failed")
	}

	cs.t.Logf("check service(%s:%s)...", r.Host, port)
	if cs.serviceOK != nil {
		if !cs.serviceOK(cs.t, cs) {
			return fmt.Errorf("service failed to serve")
		}
	}

	if cs.postServiceOK != nil && !cs.postServiceOK(cs.t, cs) {
		return fmt.Errorf("post service ok failed")
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
	pts := []*point.Point{}
	m := map[string]interface{}{}
	for k := range cs.measurementsInfo {
		m[k] = struct{}{}
	}
	// merge pts, one pt per measurement
	for len(m) > 0 {
		ps, err := cs.feeder.AnyPoints(10 * time.Second)
		if err != nil {
			cs.t.Log("got points error: ", err.Error())
			continue
		}
		for _, p := range ps {
			if _, ok := m[p.Name()]; ok {
				pts = append(pts, p)
				delete(m, p.Name())
			}
		}
	}

	cs.cr.AddField("point_latency", int64(time.Since(start)))
	cs.cr.AddField("point_count", len(pts))

	cs.t.Logf("get %d points", len(pts))
	if cs.validate != nil {
		if err := cs.validate(pts, cs.measurementsInfo, cs); err != nil {
			return err
		}
	}

	cs.t.Logf("stop input...")
	cs.ipt.Terminate()

	cs.t.Logf("exit...")
	wg.Wait()

	return nil
}

type imageInfo struct {
	images    []string
	serviceOK serviceOKFunc
}

func buildCases(t *testing.T) ([]*caseSpec, error) {
	t.Helper()
	defaultServiceOK := func(t *testing.T, cs *caseSpec) bool {
		t.Helper()
		ipt := cs.ipt
		ipt.User = user
		ipt.Password = password
		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Minute)
		defer cancel()
		ticker := time.NewTicker(time.Second)
		defer ticker.Stop()
		initSQL := "$ORACLE_HOME/bin/sqlplus / as sysdba <<EOF\n" + initSQL + "EOF\n"
		for {
			// init sql
			if code, err := cs.resource.Exec(
				[]string{"/bin/sh", "-c", fmt.Sprintf("su oracle && su oracle -c '%s' || %s", initSQL, initSQL)},
				dt.ExecOptions{StdOut: os.Stdout, StdErr: os.Stderr},
			); err != nil {
				t.Logf("run command in container failed(errCode: %d): %s", code, err.Error())
			}
			select {
			case <-ctx.Done():
				return false
			case <-ticker.C:
				err := ipt.setupDB()
				if err != nil {
					continue
				} else {
					return true
				}
			}
		}
	}

	bases := []struct {
		name             string
		conf             string
		validate         validateFunc
		measurementsInfo map[string]measurementInfo
	}{
		{
			name: "oracle-ok",
			conf: fmt.Sprintf(`
host = "%s"
user = "%s"
pass = "%s"
port = 1521
interval = "1s"
service = "XE"
timeout = "30s"
[[custom_queries]]
  sql = "SELECT GROUP_ID, METRIC_NAME, VALUE FROM GV$SYSMETRIC"
  metric = "oracle_custom"
  tags = ["GROUP_ID", "METRIC_NAME"]
  fields = ["VALUE"]
`, testutils.GetRemote().Host, user, password),
			validate: assertMeasurements,
		},
	}

	// TODO: 19c
	images := []imageInfo{
		{
			images: []string{RepoURL + "oracle", "11g"},
		},
		{
			images: []string{RepoURL + "oracle", "12c"},
		},
	}

	measurementsInfo := map[string]measurementInfo{
		"oracle_process": {
			measurement:    &processMeasurement{},
			optionalFields: []string{"pid"},
			optionalTags:   []string{"pdb_name"},
		},
		"oracle_tablespace": {
			measurement:  &tablespaceMeasurement{},
			optionalTags: []string{"pdb_name"},
		},
		"oracle_system": {
			measurement:  &systemMeasurement{},
			optionalTags: []string{"pdb_name"},
			optionalFields: []string{
				"cache_blocks_corrupt",
				"cache_blocks_lost",
				"cursor_cachehit_ratio",
				"database_wait_time_ratio",
				"disk_sorts",
				"enqueue_timeouts",
				"gc_cr_block_received",
				"memory_sorts_ratio",
				"rows_per_sort",
				"service_response_time",
				"session_count",
				"session_limit_usage",
				"sorts_per_user_call",
				"temp_space_used",
				"user_rollbacks",
				"pga_over_allocation_count",
			},
		},
		"oracle_custom": {
			measurement: &customMeasurement{},
		},
	}

	var cases []*caseSpec

	for _, img := range images {
		for _, base := range bases {
			feeder := io.NewMockedFeeder()
			ipt := defaultInput()
			ipt.feeder = feeder

			ipt.Service = "XE"
			ms := measurementsInfo
			if base.measurementsInfo != nil {
				ms = base.measurementsInfo
			}
			_, err := toml.Decode(base.conf, ipt)
			assert.NoError(t, err)

			envs := []string{
				fmt.Sprintf("ORACLE_PASSWORD=%s", password),
			}

			cases = append(cases, &caseSpec{
				t:      t,
				ipt:    ipt,
				name:   fmt.Sprintf("%s.%s", base.name, "oracle"+img.images[1]),
				feeder: feeder,
				envs:   envs,

				repo:    img.images[0],
				repoTag: img.images[1],

				servicePort: "1521/tcp",

				validate:         base.validate,
				measurementsInfo: ms,

				cr: &testutils.CaseResult{
					Name: t.Name(),
					Case: base.name,
					ExtraTags: map[string]string{
						"image":         img.images[0],
						"image_tag":     img.images[1],
						"remote_server": ipt.Host,
					},
				},
				serviceOK: func(t *testing.T, cs *caseSpec) bool {
					t.Helper()
					if img.serviceOK != nil {
						return img.serviceOK(t, cs)
					} else {
						return defaultServiceOK(t, cs)
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
	name   string
	tags   map[string]string
	fields map[string]interface{}
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
		Name: "oracle_custom",
		Type: "metric",
		Fields: map[string]interface{}{
			"VALUE": &inputs.FieldInfo{
				DataType: inputs.Float,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     "value",
			},
		},
		Tags: map[string]interface{}{
			"GROUP_ID":       &inputs.TagInfo{Desc: "group_id"},
			"METRIC_NAME":    &inputs.TagInfo{Desc: "metric_name"},
			"oracle_server":  &inputs.TagInfo{Desc: "Server addr"},
			"oracle_service": &inputs.TagInfo{Desc: "Server service"},
		},
	}
}

func assertMeasurements(pts []*point.Point, mtMap map[string]measurementInfo, cs *caseSpec) error {
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
		if _, ok := pointMap[m]; !ok {
			missingMeasurements = append(missingMeasurements, m)
		}
	}

	if len(missingMeasurements) > 0 {
		return fmt.Errorf("measurements not found: %s", strings.Join(missingMeasurements, ","))
	}

	return nil
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
