// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package httpapi

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"runtime"
	"strconv"
	"strings"
	"time"

	dt "github.com/GuanceCloud/cliutils/dialtesting"
	uhttp "github.com/GuanceCloud/cliutils/network/http"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/git"
)

type dialtestingDebugRequest struct {
	RequestID string                 `json:"request_id,omitempty"`
	Type      string                 `json:"type,omitempty"`
	Task      interface{}            `json:"task"`
	TaskType  string                 `json:"task_type"`
	Regions   []string               `json:"regions,omitempty"`
	Variables map[string]dt.Variable `json:"variables"` // variable_id => variable
}

const dialtestingDebugDNSLookupTimeout = 15 * time.Second

const dialtestingInternalNetworkDeniedMessage = "The internal network address does not support online testing. However, it can be saved and then used normally."

var errDialtestingInternalNetworkDenied = errors.New(dialtestingInternalNetworkDeniedMessage)

var (
	DialtestingDisableInternalNetworkTask      = false
	DialtestingEnableDebugAPI                  = false
	DialtestingDisabledInternalNetworkCidrList = []string{}
	dialtestingNetPathDebugTaskSetup           func(context.Context, *dt.NetPathTask)
	dialtestingNetPathDebugResultEnricher      func(*dt.NetPathTask, map[string]string, map[string]interface{})
)

func RegisterDialtestingNetPathDebugTaskSetup(setup func(context.Context, *dt.NetPathTask)) {
	dialtestingNetPathDebugTaskSetup = setup
}

func RegisterDialtestingNetPathDebugResultEnricher(
	enrich func(*dt.NetPathTask, map[string]string, map[string]interface{}),
) {
	dialtestingNetPathDebugResultEnricher = enrich
}

type dialtestingDebugResponse struct {
	Cost         string                 `json:"cost"`
	ErrorMessage string                 `json:"error_msg"`
	Status       string                 `json:"status"`
	Traceroute   string                 `json:"traceroute"`
	Fields       map[string]interface{} `json:"fields"`
}

func apiDebugDialtestingHandler(w http.ResponseWriter, req *http.Request, whatever ...interface{}) (interface{}, error) {
	tid := req.Header.Get(uhttp.XTraceID)

	reqDebug, err := getAPIDebugDialtestingRequest(req)
	if err != nil {
		l.Errorf("[%s] %s", tid, err.Error())
		return nil, uhttp.Error(ErrInvalidRequest, err.Error())
	}

	prepared, err := prepareDialtestingDebug(req.Context(), reqDebug, tid)
	if err != nil {
		l.Errorf("[%s] %s", tid, err.Error())
		return nil, err
	}

	result, err := executeDialtestingDebug(prepared, tid)
	if errors.Is(err, errDialtestingInternalNetworkDenied) {
		return nil, uhttp.Error(ErrInvalidRequest, dialtestingInternalNetworkDeniedMessage)
	}
	return result, err
}

type preparedDialtestingDebug struct {
	task     dt.ITask
	taskType string
}

func prepareDialtestingDebug(
	ctx context.Context, reqDebug *dialtestingDebugRequest, tid string,
) (*preparedDialtestingDebug, error) {
	var ct dt.TaskChild
	taskType := strings.ToUpper(reqDebug.TaskType)
	switch taskType {
	case dt.ClassHTTP:
		ct = &dt.HTTPTask{
			AdvanceOptions: &dt.HTTPAdvanceOption{
				RequestOptions: &dt.HTTPOptRequest{
					FollowRedirect: false,
				},
			},
		}
	case dt.ClassTCP:
		ct = &dt.TCPTask{}
	case dt.ClassWebsocket:
		ct = &dt.WebsocketTask{}
	case dt.ClassICMP:
		ct = &dt.ICMPTask{}
	case dt.ClassMulti:
		ct = &dt.MultiTask{}
	case dt.ClassGRPC:
		ct = &dt.GRPCTask{}
	case dt.ClassSSL:
		ct = &dt.SSLTask{}
	case dt.ClassNetPath:
		ct = &dt.NetPathTask{}
	default:
		l.Errorf("unknown task type: %s", taskType)
		return nil, uhttp.Error(ErrInvalidRequest, fmt.Sprintf("unknown task type:%s", taskType))
	}

	bys, err := json.Marshal(reqDebug.Task)
	if err != nil {
		l.Errorf(`json.Marshal: %s`, err.Error())
		return nil, err
	}

	t, err := dt.NewTask(string(bys), ct)
	if err != nil {
		if taskType == dt.ClassNetPath {
			return nil, uhttp.Error(ErrInvalidRequest, "invalid NETPATH task payload")
		}
		return nil, uhttp.Error(ErrInvalidRequest, err.Error())
	}
	if netPathTask, ok := t.(*dt.NetPathTask); ok {
		if dialtestingNetPathDebugTaskSetup == nil {
			return nil, uhttp.Error(ErrInvalidRequest, "NETPATH debug executor is not registered")
		}
		dialtestingNetPathDebugTaskSetup(ctx, netPathTask)
	}

	t.SetOption(map[string]string{"userAgent": fmt.Sprintf("datakit-%s-%s/%s/%s",
		runtime.GOOS, runtime.GOARCH, git.Version, datakit.DKHost)})

	if strings.ToLower(t.Status()) == dt.StatusStop {
		return nil, uhttp.Error(ErrInvalidRequest, "the task status is stop")
	}

	// disable redirect
	if taskType == dt.ClassHTTP {
		httpTask := ct.(*dt.HTTPTask)
		if httpTask.AdvanceOptions != nil && httpTask.AdvanceOptions.RequestOptions != nil {
			httpTask.AdvanceOptions.RequestOptions.FollowRedirect = false
		}
	}

	// -- dialtesting debug procedure start --
	if err := defDialtestingMock.debugInit(t, reqDebug.Variables); err != nil {
		if taskType == dt.ClassNetPath {
			l.Errorf("[%s] NETPATH task %s initialization failed", tid, t.GetExternalID())
			return nil, uhttp.Error(ErrInvalidRequest, "invalid NETPATH task configuration")
		}
		l.Errorf("[%s] %s", tid, err.Error())
		return nil, uhttp.Error(ErrInvalidRequest, err.Error())
	}

	return &preparedDialtestingDebug{task: t, taskType: taskType}, nil
}

func isAllowedDialtestingDebugHost(ctx context.Context, hosts []string) (bool, error) {
	return isAllowedDialtestingDebugHostWithChecker(ctx, hosts, IsAllowedHostContext)
}

func isAllowedDialtestingDebugHostWithChecker(
	ctx context.Context,
	hosts []string,
	checker func(context.Context, []string) (bool, error),
) (bool, error) {
	dnsCtx, cancel := context.WithTimeout(ctx, dialtestingDebugDNSLookupTimeout)
	defer cancel()
	return checker(dnsCtx, hosts)
}

func executeDialtestingDebug(prepared *preparedDialtestingDebug, tid string) (*dialtestingDebugResponse, error) {
	if err := validateDialtestingDebugDestination(prepared.task); err != nil {
		return nil, err
	}

	start := time.Now()
	t := prepared.task
	taskType := prepared.taskType
	status := "success"
	traceroute := ""

	if err := defDialtestingMock.debugRun(t); err != nil {
		if taskType == dt.ClassNetPath {
			l.Errorf("[%s] NETPATH task %s run failed", tid, t.GetExternalID())
			return nil, uhttp.Error(ErrInvalidRequest, "NETPATH task run failed")
		}
		l.Errorf("[%s] %s", tid, err.Error())
		return nil, uhttp.Error(ErrInvalidRequest, err.Error())
	}

	tags, fields := defDialtestingMock.getResults(t)
	if netPathTask, ok := t.(*dt.NetPathTask); ok && dialtestingNetPathDebugResultEnricher != nil {
		dialtestingNetPathDebugResultEnricher(netPathTask, tags, fields)
	}

	if vars := defDialtestingMock.getVars(t); vars != nil {
		bytes, _ := json.Marshal(vars)
		fields["post_script_variables"] = string(bytes)
	}

	failReason, ok := fields["fail_reason"].(string)
	if ok && (taskType != dt.ClassNetPath || strings.TrimSpace(failReason) != "") {
		status = "fail"
	}
	if taskType == dt.ClassTCP || taskType == dt.ClassICMP || taskType == dt.ClassNetPath {
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

func validateDialtestingDebugDestination(task dt.ITask) error {
	if !DialtestingDisableInternalNetworkTask {
		return nil
	}

	hostNames, err := task.GetHostName()
	if err != nil {
		return uhttp.Errorf(ErrInvalidRequest, "get host name: %s", err.Error())
	}
	isAllowed, err := isAllowedDialtestingDebugHost(context.Background(), hostNames)
	if err != nil {
		return uhttp.Errorf(ErrInvalidRequest, "dest host is not valid: %s", err.Error())
	}
	if !isAllowed {
		return errDialtestingInternalNetworkDenied
	}
	return nil
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
	if reqDebug.TaskType == "" {
		reqDebug.TaskType = reqDebug.Type
	}

	return &reqDebug, nil
}

// IsInternalHost check whether the host is internal host.
// if cidrs is not empty, check whether the host is in the cidrs.
func IsInternalHost(host string, cidrs []string) (bool, error) {
	return IsInternalHostContext(context.Background(), host, cidrs)
}

func IsInternalHostContext(ctx context.Context, host string, cidrs []string) (bool, error) {
	return isInternalHostContext(ctx, host, cidrs, net.DefaultResolver.LookupIP)
}

func isInternalHostContext(
	ctx context.Context,
	host string,
	cidrs []string,
	lookupIP func(context.Context, string, string) ([]net.IP, error),
) (bool, error) {
	ips, err := lookupIP(ctx, "ip", host)
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
func IsAllowedHost(hosts []string) (bool, error) {
	return IsAllowedHostContext(context.Background(), hosts)
}

func IsAllowedHostContext(ctx context.Context, hosts []string) (bool, error) {
	return isAllowedHost(hosts, func(host string) (bool, error) {
		return IsInternalHostContext(ctx, host, DialtestingDisabledInternalNetworkCidrList)
	})
}

func isAllowedHost(hosts []string, checker func(string) (bool, error)) (bool, error) {
	if !DialtestingDisableInternalNetworkTask {
		return true, nil
	}

	for _, host := range hosts {
		isInternal, err := checker(host)
		if err != nil {
			return false, err
		}

		if isInternal {
			return false, nil
		}
	}

	return true, nil
}

func init() { //nolint:gochecknoinits
	parseDialtestingEnvs()

	if DialtestingEnableDebugAPI {
		defaultDialtestingDebugManager = newDialtestingDebugManager(loadDialtestingDebugConfigFromEnv())
		RegHTTPRoute(http.MethodPost, "/v1/dialtesting/debug", apiDebugDialtestingHandler)
		RegHTTPRoute(http.MethodPost, "/v1/dialtesting/debug/runs", apiCreateDialtestingDebugRun)
		RegHTTPRoute(http.MethodGet, "/v1/dialtesting/debug/runs", apiGetDialtestingDebugRun)
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
