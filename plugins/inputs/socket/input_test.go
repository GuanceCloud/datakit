package socket

import (
	"fmt"
	"os/exec"
	"regexp"
	"testing"
	"time"

	"gitlab.jiagouyun.com/cloudcare-tools/cliutils"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io"
)

func TestInput_CollectMeasurement(t *testing.T) {
	s := &Input{
		Interval:    datakit.Duration{Duration: time.Second * 5},
		semStop:     cliutils.NewSem(),
		TimeOut:     datakit.Duration{Duration: time.Second * 5},
		SocketProto: []string{"udp"},
	}
	if len(s.SocketProto) == 0 {
		s.SocketProto = []string{"tcp", "udp"}
	}

	// Initialize regexps to validate input data
	validFields := "(bytes_acked|bytes_received|segs_out|segs_in|data_segs_in|data_segs_out|rto)"
	s.validValues = regexp.MustCompile("^" + validFields + ":[0-9]+$")
	s.isNewConnection = regexp.MustCompile(`^\s+.*$`)

	s.lister = socketList
	ssPath, err := exec.LookPath("ss")
	if err != nil {
		io.FeedLastError(inputName, "socket input init error:"+err.Error())
	}
	s.cmdName = ssPath
	for _, proto := range s.SocketProto {
		out, err := s.lister(s.cmdName, proto, s.TimeOut)
		if err != nil {
			fmt.Println(err)
		}
		s.CollectMeasurement(out, proto)
	}

	for _, i := range s.collectCache {
		fmt.Println(i.LineProto())
	}
}
