// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package apache

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/lineproto"
	tu "gitlab.jiagouyun.com/cloudcare-tools/cliutils/testutil"
)

var testdata = `127.0.0.1
ServerVersion: Apache/2.4.29 (Ubuntu)
ServerMPM: event
Server Built: 2020-08-12T21:33:25
CurrentTime: Thursday, 13-May-2021 05:40:08 UTC
RestartTime: Thursday, 13-May-2021 03:26:29 UTC
ParentServerConfigGeneration: 64
ParentServerMPMGeneration: 63
ServerUptimeSeconds: 8018
ServerUptime: 2 hours 13 minutes 38 seconds
Load1: 0.43
Load5: 0.46
Load15: 0.39
Total Accesses: 26
Total kBytes: 18
CPUUser: .02
CPUSystem: .01
CPUChildrenUser: 0
CPUChildrenSystem: 0
CPULoad: .000374158
Uptime: 8018
ReqPerSec: .0032427
BytesPerSec: 2.29883
BytesPerReq: 708.923
BusyWorkers: 1
IdleWorkers: 49
ConnsTotal: 1
ConnsAsyncWriting: 0
ConnsAsyncKeepAlive: 0
ConnsAsyncClosing: 0
Scoreboard: W_________________________________________________....................................................................................................`

func TestParse(t *testing.T) {
	body := strings.NewReader(testdata)
	n := Input{}
	if _, err := n.parse(body); err != nil {
		t.Error(err)
	}
}

func TestGetMetric(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, testdata)
	}))

	defer ts.Close()

	n := Input{
		URL: ts.URL + "/server_status",
	}

	client, err := n.createHTTPClient()
	tu.Ok(t, err)
	n.client = client

	m, err := n.getMetric()
	tu.Ok(t, err)

	tu.Assert(t, m != nil, "Measurement should not nil")

	p, err := m.LineProto()
	tu.Ok(t, err)
	encoder := lineproto.NewLineEncoder()
	if err := encoder.AppendPoint(p.Point); err != nil {
		t.Fatal(err)
	}
	line, err := encoder.UnsafeStringWithoutLn()
	if err != nil {
		t.Fatal(err)
	}
	t.Log(line)
}
