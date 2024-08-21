// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package httpapi

import (
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"strconv"
	"strings"
	"time"

	dt "github.com/GuanceCloud/cliutils/dialtesting"
	uhttp "github.com/GuanceCloud/cliutils/network/http"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/datakit"
)

type dialtestingDebugRequest struct {
	Task     interface{} `json:"task"`
	TaskType string      `json:"task_type"`
}

var (
	DialtestingDisableInternalNetworkTask      = false
	DialtestingEnableDebugAPI                  = false
	DialtestingDisabledInternalNetworkCidrList = []string{}
)

type dialtestingDebugResponse struct {
	Cost         string                 `json:"cost"`
	ErrorMessage string                 `json:"error_msg"`
	Status       string                 `json:"status"`
	Traceroute   string                 `json:"traceroute"`
	Fields       map[string]interface{} `json:"fields"`
}

func apiDebugDialtestingHandler(w http.ResponseWriter, req *http.Request, whatever ...interface{}) (interface{}, error) {
	var (
		tid        = req.Header.Get(uhttp.XTraceID)
		start      = time.Now()
		t          dt.Task
		traceroute string
		status     = "success"
	)

	reqDebug, err := getAPIDebugDialtestingRequest(req)
	if err != nil {
		l.Errorf("[%s] %s", tid, err.Error())
		return nil, uhttp.Error(ErrInvalidRequest, err.Error())
	}

	taskType := strings.ToUpper(reqDebug.TaskType)
	switch taskType {
	case dt.ClassHTTP:
		t = &dt.HTTPTask{
			Option: map[string]string{"userAgent": fmt.Sprintf("DataKit/%s dialtesting", datakit.Version)},
			AdvanceOptions: &dt.HTTPAdvanceOption{
				RequestOptions: &dt.HTTPOptRequest{
					FollowRedirect: false,
				},
			},
		}
	case dt.ClassTCP:
		t = &dt.TCPTask{}
	case dt.ClassWebsocket:
		t = &dt.WebsocketTask{}
	case dt.ClassICMP:
		t = &dt.ICMPTask{}
	default:
		l.Errorf("unknown task type: %s", taskType)
		return nil, uhttp.Error(ErrInvalidRequest, fmt.Sprintf("unknown task type:%s", taskType))
	}

	bys, err := json.Marshal(reqDebug.Task)
	if err != nil {
		l.Errorf(`json.Marshal: %s`, err.Error())
		return nil, err
	}

	if err := json.Unmarshal(bys, &t); err != nil {
		l.Errorf(`json.Unmarshal: %s`, err.Error())
		return nil, err
	}
	if strings.ToLower(t.Status()) == dt.StatusStop {
		return nil, uhttp.Error(ErrInvalidRequest, "the task status is stop")
	}

	// check internal network
	hostName, err := t.GetHostName()
	if err != nil {
		l.Warnf("get host name error: %s", err.Error())
	} else if isAllowed, err := IsAllowedHost(hostName); err != nil {
		return nil, uhttp.Errorf(ErrInvalidRequest, "dest host is not valid: %s", err.Error())
	} else if !isAllowed {
		return nil, uhttp.Errorf(ErrInvalidRequest, "dest host [%s] is not allowed to be tested", hostName)
	}

	// disable redirect
	if taskType == dt.ClassHTTP {
		httpTask := t.(*dt.HTTPTask)
		if httpTask.AdvanceOptions != nil && httpTask.AdvanceOptions.RequestOptions != nil {
			httpTask.AdvanceOptions.RequestOptions.FollowRedirect = false
		}
	}

	// -- dialtesting debug procedure start --
	if err := defDialtestingMock.debugInit(t); err != nil {
		l.Errorf("[%s] %s", tid, err.Error())
		return nil, uhttp.Error(ErrInvalidRequest, err.Error())
	}
	if err := defDialtestingMock.debugRun(t); err != nil {
		l.Warnf("[%s] %s", tid, err.Error())
	}

	tags, fields := defDialtestingMock.getResults(t)

	failReason, ok := fields["fail_reason"].(string)
	if ok {
		status = "fail"
	}
	if taskType == dt.ClassTCP || taskType == dt.ClassICMP {
		traceroute, _ = fields["traceroute"].(string)
	}

	for k, v := range tags {
		fields[k] = v
	}

	return &dialtestingDebugResponse{
		Cost:         time.Since(start).String(),
		ErrorMessage: failReason,
		Status:       status,
		Traceroute:   traceroute,
		Fields:       fields,
	}, nil
}

func getAPIDebugDialtestingRequest(req *http.Request) (*dialtestingDebugRequest, error) {
	body, err := io.ReadAll(req.Body)
	if err != nil {
		return nil, uhttp.Error(ErrInvalidRequest, err.Error())
	}

	var reqDebug dialtestingDebugRequest
	if err := json.Unmarshal(body, &reqDebug); err != nil {
		return nil, uhttp.Error(ErrInvalidRequest, err.Error())
	}

	return &reqDebug, nil
}

// IsInternalHost check whether the host is internal host.
// if cidrs is not empty, check whether the host is in the cidrs.
func IsInternalHost(host string, cidrs []string) (bool, error) {
	ips, err := net.LookupIP(host)
	if err != nil {
		return false, fmt.Errorf("lookup ip failed: %w", err)
	}

	if len(cidrs) > 0 {
		for _, cidr := range cidrs {
			_, ipNet, err := net.ParseCIDR(cidr)
			if err != nil {
				l.Warnf("parse cidr %s failed: %s", cidr, err.Error())
				continue
			}
			for _, ip := range ips {
				if ipNet.Contains(ip) {
					return true, nil
				}
			}
		}
	} else {
		for _, ip := range ips {
			if ip.IsLoopback() ||
				ip.IsPrivate() ||
				ip.IsLinkLocalUnicast() ||
				ip.IsLinkLocalMulticast() ||
				ip.IsUnspecified() {
				return true, nil
			}
		}
	}

	return false, nil
}

// IsAllowedHost check whether the host is allowed to be tested.
func IsAllowedHost(host string) (bool, error) {
	if !DialtestingDisableInternalNetworkTask {
		return true, nil
	}

	isInternal, err := IsInternalHost(host, DialtestingDisabledInternalNetworkCidrList)
	if err != nil {
		return false, err
	} else {
		return !isInternal, nil
	}
}

func init() { //nolint:gochecknoinits
	parseDialtestingEnvs()

	if DialtestingEnableDebugAPI {
		RegHTTPRoute(http.MethodPost, "/v1/dialtesting/debug", apiDebugDialtestingHandler)
	}
}

// parseDialtestingEnvs parse the envs of dialtesting.
func parseDialtestingEnvs() {
	if v := datakit.GetEnv("ENV_INPUT_DIALTESTING_ENABLE_DEBUG_API"); v != "" {
		if isEabled, err := strconv.ParseBool(v); err != nil {
			l.Warnf("parse ENV_INPUT_DIALTESTING_ENABLE_DEBUG_API[%s] error: %s, ignored", v, err.Error())
		} else {
			DialtestingEnableDebugAPI = isEabled
		}
	}

	if v := datakit.GetEnv("ENV_INPUT_DIALTESTING_DISABLE_INTERNAL_NETWORK_TASK"); v != "" {
		if isDisabled, err := strconv.ParseBool(v); err != nil {
			l.Warnf("parse ENV_INPUT_DIALTESTING_DISABLE_INTERNAL_NETWORK_TASK [%s] error: %s, ignored", v, err.Error())
		} else {
			DialtestingDisableInternalNetworkTask = isDisabled
		}
	}

	if v := datakit.GetEnv("ENV_INPUT_DIALTESTING_DISABLED_INTERNAL_NETWORK_CIDR_LIST"); v != "" {
		if err := json.Unmarshal([]byte(v), &DialtestingDisabledInternalNetworkCidrList); err != nil {
			l.Warnf("parse ENV_INPUT_DIALTESTING_DISABLED_INTERNAL_NETWORK_CIDR_LIST[%s] error: %s, ignored", v, err.Error())
		}
	}
}
