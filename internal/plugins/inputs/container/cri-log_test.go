// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package container

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseCriInfo(t *testing.T) {
	t.Run("parse-cri-info", func(t *testing.T) {
		in := "{\"sandboxID\":\"5f74c00898d3750acd1a0dadacc81ce94354b4ff9fb42ad0eb333acdb10ace03\",\"pid\":20622,\"runtimeType\":\"io.containerd.runc.v2\",\"runtimeOptions\":{\"systemd_cgroup\":true},\"config\":{\"metadata\":{\"name\":\"log-output\"},\"image\":{\"image\":\"sha256:d1a528908992e9b5bcff8329a22de1749007d0eeeccb93ab85dd5a822b8d46a0\"},\"args\":[\"/bin/sh\",\"-c\",\"i=0; while true; do\\n  echo \\\"$(date +'%F %H:%M:%S')  [$i]  Bash For Loop Examples. Hello, world! Testing output.\\\";\\n  i=$((i+1));\\n  sleep 1;\\ndone\\n\"],\"envs\":[{\"key\":\"DATAKIT_LOGS_CONFIG\",\"value\":\"[{\\\"disable\\\":false,\\\"source\\\":\\\"testing-source\\\"}]\"},{\"key\":\"KUBERNETES_PORT_443_TCP\",\"value\":\"tcp://192.168.0.2:443\"}]}}"

		out := &criInfo{
			SandboxID:   "5f74c00898d3750acd1a0dadacc81ce94354b4ff9fb42ad0eb333acdb10ace03",
			Pid:         20622,
			RuntimeType: "io.containerd.runc.v2",
			Config: criInfoConfig{
				Envs: envVars{
					{
						Key:   "DATAKIT_LOGS_CONFIG",
						Value: "[{\"disable\":false,\"source\":\"testing-source\"}]",
					},
					{
						Key:   "KUBERNETES_PORT_443_TCP",
						Value: "tcp://192.168.0.2:443",
					},
				},
			},
		}

		res, err := parseCriInfo(in)
		assert.NoError(t, err)
		assert.Equal(t, out, res)

		t.Logf("config: %s", res.Config.Envs.Find("DATAKIT_LOGS_CONFIG"))
	})
}
