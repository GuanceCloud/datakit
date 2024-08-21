//go:build linux && with_pcap
// +build linux,with_pcap

package protodec

import (
	"bytes"
	"encoding/json"
	"os"
	"strconv"
	"testing"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/externals/ebpf/internal/l7flow/comm"
)

type multiStream struct {
	conn    comm.ConnectionInfo
	netdata []comm.NetwrkData
}

func TestJsonDump(t *testing.T) {
	name := "./pcapdata/mysql.data"
	file, err := os.ReadFile(name)
	if err != nil {
		t.Fatal(err)
	}

	d := bytes.Split(file, []byte("\n"))
	cases := []*multiStream{}
	for i := range d {
		if len(d[i]) == 0 {
			continue
		}
		n := comm.NetwrkData{}
		json.Unmarshal(d[i], &n)
		var find bool
		for _, v := range cases {
			if v.conn == n.Conn {
				v.netdata = append(v.netdata, n)
				find = true
				break
			}
		}
		if !find {
			cases = append(cases, &multiStream{
				netdata: []comm.NetwrkData{n},
				conn:    n.Conn,
			})
		}
	}

	var impl ProtoDecPipe

	for i, oneCase := range cases {
		t.Run(strconv.FormatInt(int64(i), 10), func(t *testing.T) {
			for _, data := range oneCase.netdata {
				if impl == nil {
					_, impl, _ = MysqlProtoDetect(data.Payload, data.CaptureSize)
				}
				if impl != nil {
					impl.Decode(comm.FnInOut(data.Fn), &data, 0, nil)
				}
			}
			if impl == nil {
				t.Fatal("not found")
			}

			for _, v := range impl.Export(true) {
				t.Log(v.KVs.Pretty())
			}
		})
	}

}
