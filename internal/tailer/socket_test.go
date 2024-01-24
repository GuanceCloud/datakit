// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package tailer

import (
	"fmt"
	"net"
	"testing"
	"time"

	"github.com/GuanceCloud/cliutils/point"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/testutils"

	dkio "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/io"
)

func Test_mkServer(t *testing.T) {
	type args struct {
		socket string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		// we use random port(:0) here, see https://stackoverflow.com/a/43425461/342348
		{
			name:    "case1",
			args:    args{socket: "tcp://127.0.0.1:0"},
			wantErr: false,
		},
		{
			name:    "case2",
			args:    args{socket: "udp://127.0.0.1:0"}, // tcp 和 udp 可以使用同一端口
			wantErr: false,
		},
		{
			name:    "case4",
			args:    args{socket: "udp1://127.0.0.1:0"}, // err socket
			wantErr: true,
		},
		{
			name:    "case5",
			args:    args{socket: "udp127.0.0.1:0"}, // err socket
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotS, err := mkServer(tt.args.socket)
			if (err != nil) != tt.wantErr {
				t.Errorf("case:%s mkServer() error = %v, wantErr %v", tt.name, err, tt.wantErr)
				return
			}

			if gotS != nil {
				if gotS.lis != nil {
					t.Logf("TCP addr: %s", gotS.lis.Addr().String())
				}

				if gotS.conn != nil {
					t.Logf("UDP addr: %+#v", gotS.conn.LocalAddr().String())
				}
			}
		})
	}
}

var test5kMessge = "agent info: hostname:myclient123 ip:10.200.14.226 serviceType:1800 pid:8842 agentVersion:1.3.0 vmVersion:go1.19.12 serverMetaData:<serverInfo:Go Application serviceInfo:<serviceName:Go (linux, amd64, /usr/local/go) serviceLib:github.com/fsnotify/fsnotify (v1.5.4) serviceLib:github.com/golang/mock (v1.6.0) serviceLib:github.com/golang/protobuf (v1.5.2)serviceLib:github.com/hashicorp/golang-lru (v0.5.4) serviceLib:github.com/hashicorp/hcl (v1.0.0) serviceLib:github.com/magiconair/properties (v1.8.6) serviceLib:github.com/mattn/go-colorable (v0.1.12) serviceLib:github.com/mattn/go-isatty (v0.0.14) serviceLib:github.com/mgutz/ansi (v0.0.0-20200706080929-d51e80ef957d) serviceLib:github.com/mitchellh/mapstructure (v1.5.0) serviceLib:github.com/pelletier/go-toml/v2 (v2.0.1) serviceLib:github.com/pinpoint-apm/pinpoint-go-agent (v1.3.0) serviceLib:github.com/pinpoint-apm/pinpoint-go-agent/plugin/http (v1.3.0) serviceLib:github.com/shirou/gopsutil/v3 (v3.22.7) serviceLib:github.com/sirupsen/logrus (v1.8.1) serviceLib:github.com/spf13/afero (v1.8.2) serviceLib:github.com/spf13/cast (v1.5.0) serviceLib:github.com/spf13/jwalterweatherman (v1.1.0) serviceLib:github.com/spf13/pflag (v1.0.5) serviceLib:github.com/spf13/viper (v1.12.0) serviceLib:github.com/subosito/gotenv (v1.3.0) serviceLib:github.com/tklauser/go-sysconf (v0.3.10) serviceLib:github.com/tklauser/numcpus (v0.4.0) serviceLib:github.com/x-cray/logrus-prefixed-formatter (v0.5.2) serviceLib:golang.org/x/crypto (v0.1.0) serviceLib:golang.org/x/net (v0.1.0) serviceLib:golang.org/x/sys (v0.1.0) serviceLib:golang.org/x/term (v0.1.0)serviceLib:golang.org/x/text (v0.4.0) serviceLib:golang.org/x/time (v0.0.0-20201208040808-7e3f01d25324) serviceLib:google.golang.org/genproto (v0.0.0-20220519153652-3a47de7e79bd) serviceLib:google.golang.org/grpc (v1.49.0) serviceLib:google.golang.org/protobuf (v1.28.1) serviceLib:gopkg.in/ini.v1 (v1.66.4) serviceLib:gopkg.in/natefinch/lumberjack.v2 (v2.0.0) serviceLib:gopkg.in/yaml.v3 (v3.0.1) > > jvmInfo:<vmVersion:go1.19.12 gcType:JVM_GC_TYPE_CMSagent info: hostname:myclient123 ip:10.200.14.226 serviceType:1800 pid:8842 agentVersion:1.3.0 vmVersion:go1.19.12 serverMetaData:<serverInfo:Go Application serviceInfo:<serviceName:Go (linux, amd64, /usr/local/go) serviceLib:github.com/fsnotify/fsnotify (v1.5.4) serviceLib:github.com/golang/mock (v1.6.0) serviceLib:github.com/golang/protobuf (v1.5.2)serviceLib:github.com/hashicorp/golang-lru (v0.5.4) serviceLib:github.com/hashicorp/hcl (v1.0.0) serviceLib:github.com/magiconair/properties (v1.8.6) serviceLib:github.com/mattn/go-colorable (v0.1.12) serviceLib:github.com/mattn/go-isatty (v0.0.14) serviceLib:github.com/mgutz/ansi (v0.0.0-20200706080929-d51e80ef957d) serviceLib:github.com/mitchellh/mapstructure (v1.5.0) serviceLib:github.com/pelletier/go-toml/v2 (v2.0.1) serviceLib:github.com/pinpoint-apm/pinpoint-go-agent (v1.3.0) serviceLib:github.com/pinpoint-apm/pinpoint-go-agent/plugin/http (v1.3.0) serviceLib:github.com/shirou/gopsutil/v3 (v3.22.7) serviceLib:github.com/sirupsen/logrus (v1.8.1) serviceLib:github.com/spf13/afero (v1.8.2) serviceLib:github.com/spf13/cast (v1.5.0) serviceLib:github.com/spf13/jwalterweatherman (v1.1.0) serviceLib:github.com/spf13/pflag (v1.0.5) serviceLib:github.com/spf13/viper (v1.12.0) serviceLib:github.com/subosito/gotenv (v1.3.0) serviceLib:github.com/tklauser/go-sysconf (v0.3.10) serviceLib:github.com/tklauser/numcpus (v0.4.0) serviceLib:github.com/x-cray/logrus-prefixed-formatter (v0.5.2) serviceLib:golang.org/x/crypto (v0.1.0) serviceLib:golang.org/x/net (v0.1.0) serviceLib:golang.org/x/sys (v0.1.0) serviceLib:golang.org/x/term (v0.1.0)serviceLib:golang.org/x/text (v0.4.0) serviceLib:golang.org/x/time (v0.0.0-20201208040808-7e3f01d25324) serviceLib:google.golang.org/genproto (v0.0.0-20220519153652-3a47de7e79bd) serviceLib:google.golang.org/grpc (v1.49.0) serviceLib:google.golang.org/protobuf (v1.28.1) serviceLib:gopkg.in/ini.v1 (v1.66.4) serviceLib:gopkg.in/natefinch/lumberjack.v2 (v2.0.0) serviceLib:gopkg.in/yaml.v3 (v3.0.1) > > jvmInfo:<vmVersion:go1.19.12 gcType:JVM_GC_TYPE_CMSagent info: hostname:myclient123 ip:10.200.14.226 serviceType:1800 pid:8842 agentVersion:1.3.0 vmVersion:go1.19.12 serverMetaData:<serverInfo:Go Application serviceInfo:<serviceName:Go (linux, amd64, /usr/local/go) serviceLib:github.com/fsnotify/fsnotify (v1.5.4) serviceLib:github.com/golang/mock (v1.6.0) serviceLib:github.com/golang/protobuf (v1.5.2)serviceLib:github.com/hashicorp/golang-lru (v0.5.4) serviceLib:github.com/hashicorp/hcl (v1.0.0) serviceLib:github.com/magiconair/properties (v1.8.6) serviceLib:github.com/mattn/go-colorable (v0.1.12) serviceLib:github.com/mattn/go-isatty (v0.0.14) serviceLib:github.com/mgutz/ansi (v0.0.0-20200706080929-d51e80ef957d) serviceLib:github.com/mitchellh/mapstructure (v1.5.0) serviceLib:github.com/pelletier/go-toml/v2 (v2.0.1) serviceLib:github.com/pinpoint-apm/pinpoint-go-agent (v1.3.0) serviceLib:github.com/pinpoint-apm/pinpoint-go-agent/plugin/http (v1.3.0) serviceLib:github.com/shirou/gopsutil/v3 (v3.22.7) serviceLib:github.com/sirupsen/logrus (v1.8.1) serviceLib:github.com/spf13/afero (v1.8.2) serviceLib:github.com/spf13/cast (v1.5.0) serviceLib:github.com/spf13/jwalterweatherman (v1.1.0) serviceLib:github.com/spf13/pflag (v1.0.5) serviceLib:github.com/spf13/viper (v1.12.0) serviceLib:github.com/subosito/gotenv (v1.3.0) serviceLib:github.com/tklauser/go-sysconf (v0.3.10) serviceLib:github.com/tklauser/numcpus (v0.4.0) serviceLib:github.com/x-cray/logrus-prefixed-formatter (v0.5.2) serviceLib:golang.org/x/crypto (v0.1.0) serviceLib:golang.org/x/net (v0.1.0) serviceLib:golang.org/x/sys (v0.1.0) serviceLib:golang.org/x/term (v0.1.0)serviceLib:golang.org/x/text (v0.4.0) serviceLib:golang.org/x/time (v0.0.0-20201208040808-7e3f01d25324) serviceLib:google.golang.org/genproto (v0.0.0-20220519153652-3a47de7e79bd) serviceLib:google.golang.org/grpc (v1.49.0) serviceLib:google.golang.org/protobuf (v1.28.1) serviceLib:gopkg.in/ini.v1 (v1.66.4) serviceLib:gopkg.in/natefinch/lumberjack.v2 (v2.0.0) serviceLib:gopkg.in/yaml.v3 (v3.0.1) > > jvmInfo:<vmVersion:go1.19.12 gcType:JVM_GC_TYPE_CMS"

func Test_socketLogger_Start(t *testing.T) {
	tcpPort := testutils.RandPort("tcp")
	t.Logf("tcp port = %d ", tcpPort)
	udpPort := testutils.RandPort("tcp")
	t.Logf("udp port = %d", udpPort)
	// 启动socket: tcp,udp 端口
	opt := &Option{
		Source:   "logging",
		Service:  "test_service",
		Pipeline: "",
		Sockets: []string{
			fmt.Sprintf("tcp://127.0.0.1:%d", tcpPort),
			fmt.Sprintf("udp://127.0.0.1:%d", udpPort),
		},
		IgnoreStatus:          []string{"debug"},
		CharacterEncoding:     "utf-8",
		RemoveAnsiEscapeCodes: false,
		IgnoreDeadLog:         time.Minute,
		GlobalTags:            map[string]string{},
		BlockingMode:          true,
		Done:                  nil,
	}

	sl, err := NewWithOpt(opt)
	if err != nil {
		t.Errorf("new sockerLoger err=%v", err)
		return
	}
	feeder := dkio.NewMockedFeeder()
	sl.feeder = feeder

	go sl.Start()
	t.Log(" wait sl.start")
	time.Sleep(time.Second * 3)
	defer sl.Close()

	send(t, "tcp", fmt.Sprintf("127.0.0.1:%d", tcpPort))
	send(t, "udp", fmt.Sprintf("127.0.0.1:%d", udpPort))

	pts, err := feeder.NPoints(10, time.Second*5)
	if err != nil {
		t.Errorf("feeder err=%v", err)
		return
	}

	for _, pt := range pts {
		bts, _ := pt.MarshalJSON()
		t.Logf("pt :%s", string(bts))
		source := pt.GetTag("log_source")
		if source != "socket" {
			t.Errorf("source is %s", source)
		}
		service := pt.GetTag("service")
		if service != "test_service" {
			t.Errorf("source is %s", service)
		}
	}
}

func send(t *testing.T, network string, addr string) {
	t.Helper()
	conn, err := net.DialTimeout(network, addr, time.Second)
	if err != nil {
		t.Errorf("dial network:%s , err=%v", network, err)
		return
	}
	defer conn.Close() //nolint:errcheck
	for i := 0; i < 5; i++ {
		if _, err = conn.Write([]byte("this is logging message\n")); err != nil {
			t.Errorf("conn write err=%v", err)
		}
	}
}

func Test_socketLogger_doSocketV2(t *testing.T) {
	mockFeeder := dkio.NewMockedFeeder()
	sl := socketLogger{
		tcpListeners:    nil,
		udpConns:        nil,
		socketBufferLen: 1024 * 2,
		ignorePatterns:  nil,
		tags:            map[string]string{"host": "testHost"},
		opt:             &Option{Source: "default_source", InputName: "socket"},
		stop:            nil,
		servers:         nil,
		feeder:          mockFeeder,
		ptCache:         make(chan *point.Point, 10),
	}
	stop := make(chan struct{})
	// 开启tcp端口
	// port := ":19530"
	tcpPort := testutils.RandPort("tcp")
	lis, err := net.Listen("tcp", fmt.Sprintf("127.0.0.1:%d", tcpPort))
	if err != nil {
		return
	}
	go func() {
		for {
			conn, err := lis.Accept()
			if err != nil {
				stop <- struct{}{}
				break
			}
			go func(c net.Conn) {
				sl.doSocketV2(c)
			}(conn)
		}
	}()
	time.Sleep(time.Second * 2)
	verify := func(message string) bool {
		ticker := time.NewTicker(time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				return false
			case pt := <-sl.ptCache:
				t.Logf("point message:= %s", pt.LineProto())
				val, ok := pt.Get("message").(string)
				if ok {
					return val == message
				}
			}
		}
	}
	go func() {
		// client send msg
		conn, err := net.Dial("tcp", fmt.Sprintf("127.0.0.1:%d", tcpPort))
		if err != nil {
			stop <- struct{}{}
			return
		}
		defer func() {
			conn.Close()
			time.Sleep(time.Second) // 等待消费完成。
			stop <- struct{}{}
		}()

		_, _ = conn.Write([]byte("this is full message \n"))
		if !verify("this is full message ") {
			t.Errorf("can not verify message")
		}
		_, _ = conn.Write(append([]byte(test5kMessge), []byte("\n")...))
		if !verify(test5kMessge) {
			t.Errorf("can not verify 5k message")
		}
		_, _ = conn.Write([]byte("this is full message \n"))
		if !verify("this is full message ") {
			t.Errorf("can not verify message")
		}
		_, _ = conn.Write([]byte("this is half mess"))
		time.Sleep(time.Millisecond * 10)
		_, _ = conn.Write([]byte("age"))
		_, _ = conn.Write([]byte{'\n'})
		if !verify("this is half message") {
			t.Errorf("can not verify message")
		}
	}()

	<-stop
}
