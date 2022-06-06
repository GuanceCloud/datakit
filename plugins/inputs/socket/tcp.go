package socket

import (
	"context"
	"fmt"
	"gitlab.jiagouyun.com/cloudcare-tools/cliutils"
	"net"
	"strings"
	"time"
)

type TcpTask struct {
	Host            string
	Port            string
	Timeout         string
	CurStatus       string
	OwnerExternalID string

	reqCost    time.Duration
	reqDnsCost time.Duration
	reqError   string
	timeout    time.Duration
	ticker     *time.Ticker
	Frequency  string
	ExternalID string
}

func (t *TcpTask) init(debug bool) error {
	if len(t.Timeout) == 0 {
		t.timeout = 10 * time.Second
	} else {
		if timeout, err := time.ParseDuration(t.Timeout); err != nil {
			return err
		} else {
			t.timeout = timeout
		}
	}

	if !debug {
		du, err := time.ParseDuration(t.Frequency)
		if err != nil {
			return err
		}
		if t.ticker != nil {
			t.ticker.Stop()
		}
		t.ticker = time.NewTicker(du)
	}

	if strings.ToLower(t.CurStatus) == StatusStop {
		return nil
	}
	return nil
}

func (t *TcpTask) Init() error {

	return t.init(false)
}

func (t *TcpTask) GetResults() (tags map[string]string, fields map[string]interface{}) {
	tags = map[string]string{
		"dest_host": t.Host,
		"dest_port": t.Port,
		"proto":     "tcp",
	}

	responseTime := int64(t.reqCost) / 1000                     // us
	responseTimeWithDNS := int64(t.reqCost+t.reqDnsCost) / 1000 // us

	fields = map[string]interface{}{
		"response_time":          responseTime,
		"response_time_with_dns": responseTimeWithDNS,
		"success":                int64(-1),
	}

	message := map[string]interface{}{}

	var reasons []string
	if t.reqError != "" {
		reasons = append(reasons, t.reqError)
	}
	if len(reasons) != 0 {
		message[`fail_reason`] = strings.Join(reasons, `;`)
		fields[`fail_reason`] = strings.Join(reasons, `;`)
	} else {
		message["response_time_in_micros"] = responseTime
	}

	if t.reqError == "" && len(reasons) == 0 {
		fields["success"] = int64(1)
	}

	return
}

func (t *TcpTask) Clear() {
	t.reqCost = 0
	t.reqError = ""
}

func (t *TcpTask) Run() error {
	t.Clear()

	var d net.Dialer
	ctx, cancel := context.WithTimeout(context.Background(), t.timeout)
	defer cancel()

	hostIP := net.ParseIP(t.Host)

	if hostIP == nil { // host name
		start := time.Now()
		if ips, err := net.LookupIP(t.Host); err != nil {
			t.reqError = err.Error()
			return err
		} else {
			if len(ips) == 0 {
				err := fmt.Errorf("invalid host: %s, found no ip record", t.Host)
				t.reqError = err.Error()
				return err
			} else {
				t.reqDnsCost = time.Since(start)
				hostIP = ips[0]
			}
		}
	}

	tcpIPAddr := net.JoinHostPort(hostIP.String(), t.Port)

	start := time.Now()
	conn, err := d.DialContext(ctx, "tcp", tcpIPAddr)
	if err != nil {
		t.reqError = err.Error()
		t.reqDnsCost = 0
	} else {
		t.reqCost = time.Since(start)
		err := conn.Close()
		if err != nil {
			return fmt.Errorf("socket input close connection fail : %s", err)
		}
	}

	return nil
}

func (t *TcpTask) ID() string {
	if t.ExternalID == `` {
		return cliutils.XID("dtst_")
	}
	return fmt.Sprintf("_%s", t.ExternalID)
}

func (t *TcpTask) SetStatus(status string) {
	t.CurStatus = status
}

func (t *TcpTask) Status() string {
	return t.CurStatus
}

func (t *TcpTask) Ticker() *time.Ticker {
	return t.ticker
}

func (t *TcpTask) Class() string {
	return ClassTCP
}
