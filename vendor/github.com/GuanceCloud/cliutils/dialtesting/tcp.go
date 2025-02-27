// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package dialtesting

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"strings"
	"time"
)

var (
	_ TaskChild = (*TCPTask)(nil)
	_ ITask = (*TCPTask)(nil)
)

const defaultTCPTimeout = 30 * time.Second

type TCPResponseTime struct {
	IsContainDNS bool   `json:"is_contain_dns"`
	Target       string `json:"target"`

	targetTime time.Duration
}

type TCPSuccess struct {
	ResponseTime    []*TCPResponseTime `json:"response_time,omitempty"`
	Hops            []*ValueSuccess    `json:"hops,omitempty"`
	ResponseMessage []*SuccessOption   `json:"response_message,omitempty"`
}

type TCPTask struct {
	*Task
	Host             string            `json:"host"`
	Port             string            `json:"port"`
	Message          string            `json:"message"`
	Timeout          string            `json:"timeout"`
	EnableTraceroute bool              `json:"enable_traceroute"`
	TracerouteConfig *TracerouteOption `json:"traceroute_config"`
	SuccessWhen      []*TCPSuccess     `json:"success_when"`
	SuccessWhenLogic string            `json:"success_when_logic"`

	reqCost         time.Duration
	reqDNSCost      time.Duration
	reqError        string
	destIP          string
	responseMessage string
	timeout         time.Duration
	traceroute      []*Route
}

func (t *TCPTask) init() error {
	if len(t.Timeout) == 0 {
		t.timeout = 10 * time.Second
	} else {
		if timeout, err := time.ParseDuration(t.Timeout); err != nil {
			return err
		} else {
			t.timeout = timeout
		}
	}

	if len(t.SuccessWhen) == 0 {
		return fmt.Errorf(`no any check rule`)
	}

	for _, checker := range t.SuccessWhen {
		if checker.ResponseTime != nil {
			for _, v := range checker.ResponseTime {
				du, err := time.ParseDuration(v.Target)
				if err != nil {
					return err
				}
				v.targetTime = du
			}
		}

		// if [checker.Hops] is not nil, set traceroute to be true
		if checker.Hops != nil {
			t.EnableTraceroute = true
		}

		for _, v := range checker.ResponseMessage {
			err := genReg(v)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func (t *TCPTask) check() error {
	if len(t.Host) == 0 {
		return fmt.Errorf("host should not be empty")
	}

	if len(t.Port) == 0 {
		return fmt.Errorf("port should not be empty")
	}

	return nil
}

func (t *TCPTask) checkResult() (reasons []string, succFlag bool) {
	for _, chk := range t.SuccessWhen {
		// check response time
		if chk.ResponseTime != nil {
			for _, v := range chk.ResponseTime {
				reqCost := t.reqCost

				if v.IsContainDNS {
					reqCost += t.reqDNSCost
				}

				if reqCost >= v.targetTime {
					reasons = append(reasons,
						fmt.Sprintf("TCP response time(%v) larger equal than %v", reqCost, v.targetTime))
				} else if v.targetTime > 0 {
					succFlag = true
				}
			}
		}

		// check message
		if chk.ResponseMessage != nil {
			for _, v := range chk.ResponseMessage {
				if err := v.check(t.responseMessage, "response message"); err != nil {
					reasons = append(reasons, err.Error())
				} else {
					succFlag = true
				}
			}
		}

		// check traceroute
		if t.EnableTraceroute {
			hops := float64(len(t.traceroute))

			if hops == 0 {
				reasons = append(reasons, "traceroute failed with no hops")
			} else {
				for _, v := range chk.Hops {
					if err := v.check(hops); err != nil {
						reasons = append(reasons, fmt.Sprintf("traceroute hops check failed: %s", err.Error()))
					} else {
						succFlag = true
					}
				}
			}
		}
	}

	return reasons, succFlag
}

func (t *TCPTask) getResults() (tags map[string]string, fields map[string]interface{}) {
	tags = map[string]string{
		"name":      t.Name,
		"dest_host": t.Host,
		"dest_port": t.Port,
		"dest_ip":   t.destIP,
		"status":    "FAIL",
		"proto":     "tcp",
	}

	responseTime := int64(t.reqCost) / 1000                     // us
	responseTimeWithDNS := int64(t.reqCost+t.reqDNSCost) / 1000 // us

	fields = map[string]interface{}{
		"response_time":          responseTime,
		"response_time_with_dns": responseTimeWithDNS,
		"success":                int64(-1),
	}

	if t.responseMessage != "" {
		fields["response_message"] = t.responseMessage
	}

	if t.EnableTraceroute {
		fields["hops"] = 0
		if t.traceroute == nil {
			fields["traceroute"] = "[]"
		} else {
			tracerouteData, err := json.Marshal(t.traceroute)
			if err == nil && len(tracerouteData) > 0 {
				fields["traceroute"] = string(tracerouteData)
				fields["hops"] = len(t.traceroute)
			} else {
				fields["traceroute"] = "[]"
			}
		}
	}

	for k, v := range t.Tags {
		tags[k] = v
	}

	message := map[string]interface{}{}

	reasons, succFlag := t.checkResult()
	if t.reqError != "" {
		reasons = append(reasons, t.reqError)
	}

	switch t.SuccessWhenLogic {
	case "or":
		if succFlag && t.reqError == "" {
			tags["status"] = "OK"
			fields["success"] = int64(1)
			message["response_time"] = responseTime
		} else {
			message[`fail_reason`] = strings.Join(reasons, `;`)
			fields[`fail_reason`] = strings.Join(reasons, `;`)
		}
	default:
		if len(reasons) != 0 {
			message[`fail_reason`] = strings.Join(reasons, `;`)
			fields[`fail_reason`] = strings.Join(reasons, `;`)
		} else {
			message["response_time"] = responseTime
		}

		if t.reqError == "" && len(reasons) == 0 {
			tags["status"] = "OK"
			fields["success"] = int64(1)
		}
	}

	data, err := json.Marshal(message)
	if err != nil {
		fields[`message`] = err.Error()
	}

	if len(data) > MaxMsgSize {
		fields[`message`] = string(data[:MaxMsgSize])
	} else {
		fields[`message`] = string(data)
	}

	return tags, fields
}

func (t *TCPTask) metricName() string {
	return `tcp_dial_testing`
}

func (t *TCPTask) clear() {
	t.reqCost = 0
	t.reqError = ""
	t.traceroute = nil
}

func (t *TCPTask) run() error {
	var d net.Dialer
	ctx, cancel := context.WithTimeout(context.Background(), t.timeout)
	defer cancel()

	hostIP := net.ParseIP(t.Host)

	if hostIP == nil { // host name
		start := time.Now()
		if ips, err := net.LookupIP(t.Host); err != nil {
			t.reqError = err.Error()
			return nil
		} else {
			if len(ips) == 0 {
				err := fmt.Errorf("invalid host: %s, found no ip record", t.Host)
				t.reqError = err.Error()
				return nil
			} else {
				t.reqDNSCost = time.Since(start)
				hostIP = ips[0] // TODO: support mutiple ip for one host
			}
		}
	}

	t.destIP = hostIP.String()
	tcpIPAddr := net.JoinHostPort(hostIP.String(), t.Port)

	start := time.Now()
	conn, err := d.DialContext(ctx, "tcp", tcpIPAddr)

	if err != nil {
		t.reqError = err.Error()
		t.reqDNSCost = 0
	} else {
		t.reqCost = time.Since(start)
		if t.Message != "" { // send message and get response
			if err := conn.SetDeadline(time.Now().Add(defaultTCPTimeout)); err != nil {
				t.reqError = err.Error()
			} else if _, err := conn.Write([]byte(t.Message)); err != nil {
				t.reqError = err.Error()
			} else {
				buf := make([]byte, 1024)
				if n, err := conn.Read(buf); err != nil {
					t.reqError = err.Error()
				} else {
					t.responseMessage = string(buf[:n])
				}
			}
		}

		_ = conn.Close() //nolint:errcheck
	}

	if t.EnableTraceroute {
		routes, err := TracerouteIP(hostIP.String(), t.TracerouteConfig)
		if err != nil {
			t.reqError = err.Error()
		} else {
			t.traceroute = routes
		}
	}

	return nil
}

func (t *TCPTask) stop() {}

func (t *TCPTask) class() string {
	return ClassTCP
}

func (t *TCPTask) getHostName() ([]string, error) {
	return []string{t.Host}, nil
}

func (t *TCPTask) beforeFirstRender() {
}

func (t *TCPTask) getVariableValue(variable Variable) (string, error) {
	return "", fmt.Errorf("not support")
}

func (t *TCPTask) getRawTask(taskString string) (string, error) {
	task := TCPTask{}

	if err := json.Unmarshal([]byte(taskString), &task); err != nil {
		return "", fmt.Errorf("unmarshal tcp task failed: %w", err)
	}

	task.Task = nil

	bytes, _ := json.Marshal(task)
	return string(bytes), nil
}

func (t *TCPTask) initTask() {
	if t.Task == nil {
		t.Task = &Task{}
	}
}
