// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package snmp

import (
	"fmt"
	"io/ioutil"
	"net"
	"os"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/BurntSushi/toml"
	"github.com/GuanceCloud/cliutils/point"
	"github.com/gosnmp/gosnmp"
	dockertest "github.com/ory/dockertest/v3"
	"github.com/ory/dockertest/v3/docker"
	"github.com/stretchr/testify/require"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs/snmp/snmpmeasurement"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/testutils"
)

// ATTENTION: Docker version should use v20.10.18 in integrate tests. Other versions are not tested.

func TestIntegrate(t *testing.T) {
	if !testutils.CheckIntegrationTestingRunning() {
		t.Skip()
	}

	testutils.PurgeRemoteByName(snmpmeasurement.InputName)       // purge at first.
	defer testutils.PurgeRemoteByName(snmpmeasurement.InputName) // purge at last.

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

					require.NoError(t, tc.pool.Purge(tc.resource))
				})
			})
		}(tc)
	}
}

func buildCases(t *testing.T) ([]*caseSpec, error) {
	t.Helper()

	remote := testutils.GetRemote()

	bases := []struct {
		name           string // Also used as build image name:tag.
		conf           string
		dockerFileText string // Empty if not build image.
		exposedPorts   []string
		optsObject     []inputs.PointCheckOption
		optsMetric     []inputs.PointCheckOption
	}{
		{
			name: "pubrepo.jiagouyun.com/image-repo-for-testing/snmp:inexio-snmpsim:v2",
			conf: fmt.Sprintf(`specific_devices = ["%s"]
	snmp_version = 2
	v2_community_string = "recorded/cisco-catalyst-3750"
[tags]
	tag1 = "val1"
	tag2 = "val2"`, remote.Host),
			exposedPorts: []string{"161/udp"},
			optsMetric: []inputs.PointCheckOption{
				inputs.WithOptionalTags("interface", "interface_alias", "mac_addr", "entity_name", "power_source", "power_status_descr", "temp_index", "temp_state", "cpu", "mem", "mem_pool_name", "sensor_id", "sensor_type"),
				inputs.WithOptionalFields("ifNumber", "sysUpTimeInstance", "tcpActiveOpens", "tcpAttemptFails", "tcpCurrEstab", "tcpEstabResets", "tcpInErrs", "tcpOutRsts", "tcpPassiveOpens", "tcpRetransSegs", "udpInErrors", "udpNoPorts", "ifAdminStatus", "ifHCInBroadcastPkts", "ifHCInMulticastPkts", "ifHCInOctets", "ifHCInOctetsRate", "ifHCInUcastPkts", "ifHCOutBroadcastPkts", "ifHCOutMulticastPkts", "ifHCOutOctets", "ifHCOutOctetsRate", "ifHCOutUcastPkts", "ifHighSpeed", "ifInDiscards", "ifInDiscardsRate", "ifInErrors", "ifInErrorsRate", "ifOperStatus", "ifOutDiscards", "ifOutDiscardsRate", "ifOutErrors", "ifOutErrorsRate", "ifSpeed", "ifBandwidthInUsageRate", "ifBandwidthOutUsageRate", "cpuUsage", "memoryUsed", "memoryUsage", "memoryFree", "cieIfLastOutTime", "cieIfOutputQueueDrops", "ciscoMemoryPoolUsed", "cpmCPUTotalMonIntervalValue", "cieIfLastInTime", "cieIfResetCount", "ciscoMemoryPoolLargestFree", "ciscoEnvMonTemperatureStatusValue", "ciscoEnvMonSupplyState", "cswStackPortOperStatus", "cpmCPUTotal1minRev", "ciscoMemoryPoolFree", "cieIfInputQueueDrops", "ciscoEnvMonFanState", "cswSwitchState", "entSensorValue"), // nolint:lll
			},
		},
		{
			name: "pubrepo.jiagouyun.com/image-repo-for-testing/snmp:inexio-snmpsim:v3",
			conf: fmt.Sprintf(`specific_devices = ["%s"]
	snmp_version = 3
	v3_user = "testing"
	v3_auth_protocol = "MD5"
	v3_auth_key = "testing123"
	v3_priv_protocol = "DES"
	v3_priv_key = "12345678"
	v3_context_name = "recorded/cisco-catalyst-3750"
[tags]
	tag1 = "val1"
	tag2 = "val2"`, remote.Host),
			exposedPorts: []string{"161/udp"},
			optsMetric: []inputs.PointCheckOption{
				inputs.WithOptionalTags("interface", "interface_alias", "mac_addr", "entity_name", "power_source", "power_status_descr", "temp_index", "temp_state", "cpu", "mem", "mem_pool_name", "sensor_id", "sensor_type"),
				inputs.WithOptionalFields("ifNumber", "sysUpTimeInstance", "tcpActiveOpens", "tcpAttemptFails", "tcpCurrEstab", "tcpEstabResets", "tcpInErrs", "tcpOutRsts", "tcpPassiveOpens", "tcpRetransSegs", "udpInErrors", "udpNoPorts", "ifAdminStatus", "ifHCInBroadcastPkts", "ifHCInMulticastPkts", "ifHCInOctets", "ifHCInOctetsRate", "ifHCInUcastPkts", "ifHCOutBroadcastPkts", "ifHCOutMulticastPkts", "ifHCOutOctets", "ifHCOutOctetsRate", "ifHCOutUcastPkts", "ifHighSpeed", "ifInDiscards", "ifInDiscardsRate", "ifInErrors", "ifInErrorsRate", "ifOperStatus", "ifOutDiscards", "ifOutDiscardsRate", "ifOutErrors", "ifOutErrorsRate", "ifSpeed", "ifBandwidthInUsageRate", "ifBandwidthOutUsageRate", "cpuUsage", "memoryUsed", "memoryUsage", "memoryFree", "cieIfLastOutTime", "cieIfOutputQueueDrops", "ciscoMemoryPoolUsed", "cpmCPUTotalMonIntervalValue", "cieIfLastInTime", "cieIfResetCount", "ciscoMemoryPoolLargestFree", "ciscoEnvMonTemperatureStatusValue", "ciscoEnvMonSupplyState", "cswStackPortOperStatus", "cpmCPUTotal1minRev", "ciscoMemoryPoolFree", "cieIfInputQueueDrops", "ciscoEnvMonFanState", "cswSwitchState", "entSensorValue"), // nolint:lll
			},
		},
	}

	var cases []*caseSpec

	// compose cases
	for _, base := range bases {
		feeder := io.NewMockedFeeder()

		ipt := defaultInput()
		ipt.feeder = feeder
		// ipt.EnablePickingData = true // If uncomment this, you must adjust point count parameters for performance(time cost) in the function call "cs.feeder.NPoints" which inside in the func "run".
		// ipt.PickingCPU = []string{"cpuUsage"}

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

			dockerFileText: base.dockerFileText,
			exposedPorts:   base.exposedPorts,
			optsObject:     base.optsObject,
			optsMetric:     base.optsMetric,

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
	optsObject     []inputs.PointCheckOption
	optsMetric     []inputs.PointCheckOption
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

	for _, pt := range pts {
		measurement := string(pt.Name())

		switch measurement {
		case snmpmeasurement.SNMPObjectName:
			opts = append(opts, inputs.WithDoc(&snmpmeasurement.SNMPObject{}))
			opts = append(opts, cs.optsObject...)

			msgs := inputs.CheckPoint(pt, opts...)

			for _, msg := range msgs {
				cs.t.Logf("check measurement %s failed: %+#v", measurement, msg)
			}

			// TODO: error here
			if len(msgs) > 0 {
				return fmt.Errorf("check measurement %s failed: %+#v", measurement, msgs)
			}

			cs.mCount[snmpmeasurement.SNMPObjectName] = struct{}{}

		case snmpmeasurement.SNMPMetricName:
			opts = append(opts, inputs.WithDoc(&snmpmeasurement.SNMPMetric{}))
			opts = append(opts, cs.optsMetric...)

			msgs := inputs.CheckPoint(pt, opts...)

			for _, msg := range msgs {
				cs.t.Logf("check measurement %s failed: %+#v", measurement, msg)
			}

			// TODO: error here
			if len(msgs) > 0 {
				return fmt.Errorf("check measurement %s failed: %+#v", measurement, msgs)
			}

			cs.mCount[snmpmeasurement.SNMPMetricName] = struct{}{}

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

	dockerFileDir, dockerFilePath, err := cs.getDockerFilePath()
	if err != nil {
		return err
	}
	defer os.RemoveAll(dockerFileDir)

	uniqueContainerName := testutils.GetUniqueContainerName(snmpmeasurement.InputName)

	var resource *dockertest.Resource

	if len(cs.dockerFileText) == 0 {
		// Just run a container from existing docker image.
		resource, err = p.RunWithOptions(
			&dockertest.RunOptions{
				Name: uniqueContainerName, // ATTENTION: not cs.name.

				Repository: cs.repo,
				Tag:        cs.repoTag,
				Env:        []string{"EXTRA_FLAGS=--v3-user=testing --v3-auth-key=testing123 --v3-auth-proto=MD5 --v3-priv-key=12345678 --v3-priv-proto=DES"},

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
				Env:        []string{"EXTRA_FLAGS=--v3-user=testing --v3-auth-key=testing123 --v3-auth-proto=MD5 --v3-priv-key=12345678 --v3-priv-proto=DES"},

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
	if port, err := getPortFromString(cs.serverPorts[0]); err != nil {
		return err
	} else {
		cs.ipt.Port = port // set conf URL here.
	}

	cs.t.Logf("check service(%s:%v)...", r.Host, cs.exposedPorts)

	if err := cs.checkSNMPPortOK(r); err != nil {
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
	pts, err := cs.feeder.NPoints(100, 5*time.Minute)
	if err != nil {
		return err
	}

	cs.cr.AddField("point_latency", int64(time.Since(start)))
	cs.cr.AddField("point_count", len(pts))

	cs.t.Logf("get %d points", len(pts))

	for _, v := range pts {
		cs.t.Logf(v.LPPoint().String() + "\n")
	}

	cs.mCount = make(map[string]struct{})
	if err := cs.checkPoint(pts); err != nil {
		return err
	}

	cs.t.Logf("stop input...")
	cs.ipt.Terminate()

	require.Equal(cs.t, 2, len(cs.mCount))

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

func (cs *caseSpec) checkSNMPPortOK(r *testutils.RemoteInfo) error {
	tick := time.NewTicker(time.Minute)
	var errReturn error

	out := false
	for {
		if out {
			break
		}

		select {
		case <-tick.C:
			out = true
		default:
			if err := checkSNMPPort(r.Host, cs.serverPorts[0]); err != nil {
				cs.t.Logf("checkSNMPPort failed: %v", err)
				errReturn = err
				continue
			}

			errReturn = nil
			out = true
		}
	}

	return errReturn
}

func getPortFromString(str string) (uint16, error) {
	ui64, err := strconv.ParseUint(str, 10, 64)
	if err != nil {
		return 0, err
	}
	return uint16(ui64), nil
}

func checkSNMPPort(host, portStr string) error {
	// Default is a pointer to a GoSNMP struct that contains sensible defaults
	// eg port 161, community public, etc
	gosnmp.Default.Target = host

	port, err := getPortFromString(portStr)
	if err != nil {
		return err
	}
	gosnmp.Default.Port = port

	if err = gosnmp.Default.Connect(); err != nil {
		return err
	}
	defer gosnmp.Default.Conn.Close()

	oids := []string{"1.3.6.1.2.1.1.4.0", "1.3.6.1.2.1.1.7.0"}
	_, err2 := gosnmp.Default.Get(oids) // Get() accepts up to g.MAX_OIDS
	if err2 != nil {
		return err2
	}

	return nil
}

////////////////////////////////////////////////////////////////////////////////
