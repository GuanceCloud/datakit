//go:build linux
// +build linux

package l4log

import (
	"encoding/base64"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/vishvananda/netns"
)

func TestPortListen(t *testing.T) {
	nns, err := netns.Get()
	if err != nil {
		t.Error(err)
	}

	h := newNetNsHandle(true, true, nns)

	if v, err := h.tcpPortListen(nil); err != nil {
		t.Error(err)
	} else {
		for _, v := range v {
			t.Log(v.IP, " ", v.Port, " ", v.St, " ", v.V6)
		}
	}

	if v, err := h.nicInfo(); err != nil {
		t.Error(err)
	} else {
		for _, v := range v {
			s := strings.Builder{}
			s.WriteString(fmt.Sprintf("%s %s %d %v\n", v.Name, v.MAC, v.Index, v.VIface))
			for _, v := range v.Addrs {
				s.WriteString(fmt.Sprintf("%s\n", v.IP))
			}
			t.Log(s.String())
		}
	}
}

func TestSpanid(t *testing.T) {
	spanID := "tpIhb+V93r4="

	r := base64.NewDecoder(base64.StdEncoding, strings.NewReader(spanID))
	buf := make([]byte, 16)
	if _, err := r.Read(buf); err != nil {
		t.Error(err)
	}

	buf = buf[:8]

	d := binary.BigEndian.Uint64(buf)

	s := strconv.FormatUint(d, 16)
	t.Log(s)
}

func TestMatch(t *testing.T) {
	name := "veth"
	assert.True(t, strings.HasPrefix(name, "veth"))

	name = "veth1"
	assert.True(t, strings.HasPrefix(name, "veth"))
}

func TestDetectProto(t *testing.T) {
	h := HTTPLog{}
	if !h.DetectProto([]byte("GET / HTTP1.1\r\n")) {
		t.Fatalf("should be http proto")
	}
}

func TestPt(t *testing.T) {
	msg := map[string]any{
		"tcp": map[string]any{
			"tcp_series_col_name": []string{
				"txrx", "flags", "seq", "ack_seq", "payload_size", "win", "ts",
			},
			"tcp_series": []*PktTCPHdr{
				{
					TXRX:           "tx",
					Flags:          TCPSYN | TCPACK,
					Seq:            1,
					AckSeq:         2,
					TCPPayloadSize: 3,
					Win:            4,
					TS:             5,
				},
			},
		},
	}

	v, _ := json.Marshal(msg)
	t.Log(string(v))
}

func TestRetrans(t *testing.T) {
	rec := tcpRetransAndReorder{}

	elems := []*tcpSortElem{
		{
			idx:    0,
			txRx:   directionTX,
			seq:    100,
			len:    10,
			ackSeq: 20,
		},
		{
			idx:    1,
			txRx:   directionRX,
			seq:    20,
			len:    10,
			ackSeq: 120,
		},
		{
			idx:    2,
			txRx:   directionTX,
			seq:    130,
			len:    10,
			ackSeq: 30,
		},
		{
			idx:    3,
			txRx:   directionTX,
			seq:    120,
			len:    10,
			ackSeq: 30,
		},
		{
			idx:    4,
			txRx:   directionTX,
			seq:    120,
			len:    10,
			ackSeq: 30,
		},
		{
			idx:    5,
			txRx:   directionRX,
			seq:    30,
			len:    10,
			ackSeq: 140,
		},
		{
			idx:    6,
			txRx:   directionTX,
			seq:    140,
			len:    0,
			ackSeq: 40,
		},

		{
			idx:    7,
			txRx:   directionTX,
			seq:    140,
			len:    0,
			ackSeq: 39,
		},
		{
			idx:    8,
			txRx:   directionTX,
			seq:    140,
			len:    0,
			ackSeq: 39,
		},
	}

	var counter int
	for _, e := range elems {
		if r := rec.insert(e); r == 1 {
			counter++
		}
	}

	t.Log(elems)

	assert.Equal(t, []*tcpSortElem{elems[0], elems[3], elems[4], elems[2], elems[7], elems[8], elems[6]}, rec.txPkts) // tx
	assert.Equal(t, []*tcpSortElem{elems[1], elems[5]}, rec.rxPkts)
	assert.Equal(t, 2, counter)
}

// func TestM(t *testing.T) {
// 	k8sinfo, err := k8sinfo.NewK8sInfoFromENV()
// 	if err != nil {
// 		t.Error(err)
// 	} else {
// 		k8sinfo.AutoUpdate(context.Background())
// 	}

// 	enableNetlog = true
// 	enabledNetMetric = true
// 	initULID()

// 	SetK8sNetInfo(k8sinfo)

// 	exporter.Init(log)
// 	rt, err := cruntime.NewDockerRuntime("unix:///var/run/docker.sock", "")
// 	if err != nil {
// 		t.Error(err)
// 	}
// 	svc, err := remote.NewRemoteRuntimeService("unix:///var/run/containerd/containerd.sock", time.Second*5)
// 	if err != nil {
// 		t.Error(err)
// 	}
// 	m, err := newNetlogMonitor(map[string]string{}, fmt.Sprintf("http://%s%s?input=",
// 		"0.0.0.0:9529", point.Logging.URL())+url.QueryEscape("netlog"),
// 		fmt.Sprintf("http://%s%s?input=",
// 			"0.0.0.0:9529", point.Network.URL())+url.QueryEscape("netlog"), "udp", nil)

// 	if err != nil {
// 		t.Fatal(err)
// 	}
// 	m.Run(context.Background(), svc, rt)
// }

func TestStructHTTPLog(t *testing.T) {
	elem := HTTPLogElem{}

	s, err := json.Marshal(elem)
	if err != nil {
		t.Error(err)
	}
	t.Log(string(s))
}

var req = "\x47\x45\x54\x20\x2f\x61\x70\x69\x2f\x64\x61\x74\x61\x2f\x3f\x63" +
	"\x6f\x6e\x74\x65\x78\x74\x4b\x65\x79\x73\x3d\x62\x69\x6e\x6f\x63" +
	"\x75\x6c\x61\x72\x73\x20\x48\x54\x54\x50\x2f\x31\x2e\x31\x0d\x0a" +
	"\x68\x6f\x73\x74\x3a\x20\x6f\x70\x65\x6e\x74\x65\x6c\x65\x6d\x65" +
	"\x74\x72\x79\x2d\x64\x65\x6d\x6f\x2d\x66\x72\x6f\x6e\x74\x65\x6e" +
	"\x64\x70\x72\x6f\x78\x79\x3a\x38\x30\x38\x30\x0d\x0a\x75\x73\x65" +
	"\x72\x2d\x61\x67\x65\x6e\x74\x3a\x20\x70\x79\x74\x68\x6f\x6e\x2d" +
	"\x72\x65\x71\x75\x65\x73\x74\x73\x2f\x32\x2e\x33\x31\x2e\x30\x0d" +
	"\x0a\x61\x63\x63\x65\x70\x74\x2d\x65\x6e\x63\x6f\x64\x69\x6e\x67" +
	"\x3a\x20\x67\x7a\x69\x70\x2c\x20\x64\x65\x66\x6c\x61\x74\x65\x2c" +
	"\x20\x62\x72\x0d\x0a\x61\x63\x63\x65\x70\x74\x3a\x20\x2a\x2f\x2a" +
	"\x0d\x0a\x62\x61\x67\x67\x61\x67\x65\x3a\x20\x73\x79\x6e\x74\x68" +
	"\x65\x74\x69\x63\x5f\x72\x65\x71\x75\x65\x73\x74\x3d\x74\x72\x75" +
	"\x65\x0d\x0a\x78\x2d\x66\x6f\x72\x77\x61\x72\x64\x65\x64\x2d\x70" +
	"\x72\x6f\x74\x6f\x3a\x20\x68\x74\x74\x70\x0d\x0a\x78\x2d\x72\x65" +
	"\x71\x75\x65\x73\x74\x2d\x69\x64\x3a\x20\x64\x39\x33\x33\x36\x36" +
	"\x66\x62\x2d\x31\x63\x62\x34\x2d\x39\x38\x38\x38\x2d\x62\x32\x31" +
	"\x65\x2d\x31\x65\x36\x38\x35\x63\x32\x32\x63\x66\x37\x64\x0d\x0a" +
	"\x78\x2d\x65\x6e\x76\x6f\x79\x2d\x65\x78\x70\x65\x63\x74\x65\x64" +
	"\x2d\x72\x71\x2d\x74\x69\x6d\x65\x6f\x75\x74\x2d\x6d\x73\x3a\x20" +
	"\x31\x35\x30\x30\x30\x0d\x0a\x74\x72\x61\x63\x65\x70\x61\x72\x65" +
	"\x6e\x74\x3a\x20\x30\x30\x2d\x64\x32\x64\x32\x33\x34\x65\x37\x38" +
	"\x38\x65\x31\x37\x31\x66\x37\x63\x35\x64\x61\x38\x61\x39\x31\x31" +
	"\x34\x31\x34\x34\x33\x64\x31\x2d\x31\x64\x61\x61\x66\x33\x34\x32" +
	"\x65\x37\x35\x64\x66\x62\x30\x35\x2d\x30\x31\x0d\x0a\x74\x72\x61" +
	"\x63\x65\x73\x74\x61\x74\x65\x3a\x20\x0d\x0a\x0d\x0a"

var resp = "\x48\x54\x54\x50\x2f\x31\x2e\x31\x20\x33\x30\x38\x20\x50\x65\x72" +
	"\x6d\x61\x6e\x65\x6e\x74\x20\x52\x65\x64\x69\x72\x65\x63\x74\x0d" +
	"\x0a\x4c\x6f\x63\x61\x74\x69\x6f\x6e\x3a\x20\x2f\x61\x70\x69\x2f" +
	"\x64\x61\x74\x61\x3f\x63\x6f\x6e\x74\x65\x78\x74\x4b\x65\x79\x73" +
	"\x3d\x62\x69\x6e\x6f\x63\x75\x6c\x61\x72\x73\x0d\x0a\x52\x65\x66" +
	"\x72\x65\x73\x68\x3a\x20\x30\x3b\x75\x72\x6c\x3d\x2f\x61\x70\x69" +
	"\x2f\x64\x61\x74\x61\x3f\x63\x6f\x6e\x74\x65\x78\x74\x4b\x65\x79" +
	"\x73\x3d\x62\x69\x6e\x6f\x63\x75\x6c\x61\x72\x73\x0d\x0a\x44\x61" +
	"\x74\x65\x3a\x20\x54\x68\x75\x2c\x20\x31\x36\x20\x4e\x6f\x76\x20" +
	"\x32\x30\x32\x33\x20\x30\x36\x3a\x30\x33\x3a\x30\x35\x20\x47\x4d" +
	"\x54\x0d\x0a\x43\x6f\x6e\x6e\x65\x63\x74\x69\x6f\x6e\x3a\x20\x6b" +
	"\x65\x65\x70\x2d\x61\x6c\x69\x76\x65\x0d\x0a\x4b\x65\x65\x70\x2d" +
	"\x41\x6c\x69\x76\x65\x3a\x20\x74\x69\x6d\x65\x6f\x75\x74\x3d\x35" +
	"\x0d\x0a\x54\x72\x61\x6e\x73\x66\x65\x72\x2d\x45\x6e\x63\x6f\x64" +
	"\x69\x6e\x67\x3a\x20\x63\x68\x75\x6e\x6b\x65\x64\x0d\x0a\x0d\x0a" +
	"\x32\x30\x0d\x0a\x2f\x61\x70\x69\x2f\x64\x61\x74\x61\x3f\x63\x6f" +
	"\x6e\x74\x65\x78\x74\x4b\x65\x79\x73\x3d\x62\x69\x6e\x6f\x63\x75" +
	"\x6c\x61\x72\x73\x0d\x0a\x30\x0d\x0a\x0d\x0a"

func TestHTTPLog(t *testing.T) {
	httplog := HTTPLog{}
	_ = resp

	httplog.Handle(nil, 0, []byte(req), 10, &PktTCPHdr{
		TXRX:           "tx",
		Flags:          TCPSYN | TCPACK,
		Seq:            1,
		AckSeq:         2,
		TCPPayloadSize: 3,
		Win:            4,
		TS:             5,
	}, &PMeta{
		SrcIP:   "sd",
		DstIP:   "dd",
		SrcPort: 100,
		DstPort: 200,
	}, 0, 1)

	httplog.Handle(nil, 0, []byte(req), 10, &PktTCPHdr{
		TXRX:           "tx",
		Flags:          TCPSYN | TCPACK,
		Seq:            1,
		AckSeq:         2,
		TCPPayloadSize: 3,
		Win:            4,
		TS:             5,
	}, &PMeta{
		SrcIP:   "sd",
		DstIP:   "dd",
		SrcPort: 100,
		DstPort: 200,
	}, 0, 1)

	t.Log(httplog)
}
