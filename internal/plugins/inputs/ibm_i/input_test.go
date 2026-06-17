// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package ibm_i

import (
	"strings"
	"testing"

	"github.com/BurntSushi/toml"
	"github.com/stretchr/testify/require"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs/external"
)

func TestSampleConfigUsesExternalCollector(t *testing.T) {
	type inputsConf struct {
		External []*external.Input `toml:"external"`
	}
	type mainConf struct {
		Inputs inputsConf `toml:"inputs"`
	}

	var conf mainConf
	_, err := toml.Decode(sampleConfig, &conf)
	require.NoError(t, err)
	require.Len(t, conf.Inputs.External, 1)

	ipt := conf.Inputs.External[0]
	require.Equal(t, inputName, ipt.Name)
	require.True(t, ipt.Daemon)
	require.True(t, ipt.Election)
	require.Equal(t, "/usr/local/datakit/externals/ibm_i", ipt.Cmd)
	require.NotContains(t, ipt.Args, "--metric-enabled")
	require.Contains(t, ipt.Args, "--host")
	require.Contains(t, ipt.Args, "--username")
	require.NotContains(t, ipt.Args, "--port")
	require.NotContains(t, ipt.Args, "--collect-history-log=true")
	require.NotContains(t, ipt.Args, "--password")
	require.Contains(t, ipt.Envs, "ENV_INPUT_IBM_I_PASSWORD=<password>")
	require.Contains(t, ipt.Envs, "LD_LIBRARY_PATH=/opt/ibm/iaccess/lib64:$LD_LIBRARY_PATH")
}

func TestSampleMeasurementsMatchCurrentCollector(t *testing.T) {
	ipt := defaultInput()
	names := []string{}
	for _, m := range ipt.SampleMeasurement() {
		info := m.Info()
		require.NotNil(t, info)
		names = append(names, info.Name)
	}

	require.ElementsMatch(t, []string{"ibm_i", "collector"}, names)

	require.NotContains(t, names, "ibm_i_system")
	require.NotContains(t, names, "ibm_i_job")
	require.NotContains(t, names, "ibm_i_sql")
	require.NotContains(t, names, "ibm_i_sql_statement")
	require.NotContains(t, names, "ibm_i_history_log")
	require.NotContains(t, names, "ibm_i_job_log")
}

func TestUnifiedMeasurementFieldsAndTaggedBy(t *testing.T) {
	info := (&ibmiMeasurement{}).Info()
	require.Equal(t, measurementName, info.Name)
	require.Len(t, info.Fields, 27)
	require.Contains(t, info.Tags, "host")
	require.NotContains(t, info.Tags, "ibmi_host")
	require.Contains(t, info.Tags, "partition_id")
	require.Contains(t, info.Tags, "job_id")

	systemField, ok := info.Fields["system_configured_cpus"].(*inputs.FieldInfo)
	require.True(t, ok)
	require.Equal(t, inputs.Float, systemField.DataType)
	require.ElementsMatch(t, []string{"partition_id"}, systemField.Taggedby)
	require.NotContains(t, systemField.Taggedby, "host")

	aspField, ok := info.Fields["asp_io_requests_per_s"].(*inputs.FieldInfo)
	require.True(t, ok)
	require.ElementsMatch(t,
		[]string{"asp_number", "resource_name", "serial_number", "unit_number", "unit_type"},
		aspField.Taggedby)

	jobField, ok := info.Fields["job_status_value"].(*inputs.FieldInfo)
	require.True(t, ok)
	require.Contains(t, jobField.Taggedby, "job_id")
	require.Contains(t, jobField.Taggedby, "job_status")

	poolField, ok := info.Fields["pool_size"].(*inputs.FieldInfo)
	require.True(t, ok)
	require.Equal(t, inputs.Float, poolField.DataType)
	require.ElementsMatch(t, []string{"pool_name", "subsystem_name"}, poolField.Taggedby)

	require.NotContains(t, info.Fields, "configured_cpus")
	require.NotContains(t, info.Fields, "status")
	require.NotContains(t, info.Fields, "critical_size")

	for name := range info.Fields {
		require.Truef(t,
			strings.HasPrefix(name, "system_") ||
				strings.HasPrefix(name, "asp_") ||
				strings.HasPrefix(name, "job_") ||
				strings.HasPrefix(name, "pool_") ||
				strings.HasPrefix(name, "subsystem_") ||
				strings.HasPrefix(name, "message_queue_"),
			"field %q does not have an IBM i metric group prefix", name)
	}
}
