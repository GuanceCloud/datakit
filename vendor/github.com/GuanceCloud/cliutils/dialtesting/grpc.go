// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package dialtesting

import (
	"bufio"
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"strings"
	"text/template"
	"time"

	pdesc "github.com/jhump/protoreflect/desc"
	"github.com/jhump/protoreflect/desc/protoparse"
	"github.com/jhump/protoreflect/dynamic"
	"github.com/jhump/protoreflect/dynamic/grpcdynamic"
	"github.com/jhump/protoreflect/grpcreflect"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/metadata"
)

var (
	_ TaskChild = (*GRPCTask)(nil)
	_ ITask     = (*GRPCTask)(nil)
)

const (
	DefaultGRPCTimeout     = 30 * time.Second
	HealthCheckServiceName = "grpc.health.v1.Health"
	HealthCheckMethodName  = "Check"
)

type GRPCOptCertificate struct {
	IgnoreServerCertificateError bool   `json:"ignore_server_certificate_error,omitempty"`
	PrivateKey                   string `json:"private_key,omitempty"`
	Certificate                  string `json:"certificate,omitempty"`
	CaCert                       string `json:"ca,omitempty"`
}

type GRPCSecret struct {
	NoSaveResponseBody bool `json:"not_save,omitempty"`
}

type GRPCSuccess struct {
	Body         []*SuccessOption `json:"body,omitempty"`
	ResponseTime string           `json:"response_time,omitempty"`
	respTime     time.Duration
}

type GRPCProtoFilesDiscovery struct {
	ProtoFiles  map[string]string `json:"protofiles"`
	FullMethod  string            `json:"full_method"`
	JSONRequest string            `json:"request,omitempty"`
}

type GRPCReflectionDiscovery struct {
	FullMethod  string `json:"full_method"`
	JSONRequest string `json:"request,omitempty"`
}

type GRPCHealthCheckDiscovery struct {
	Service string `json:"service,omitempty"`
}

type GRPCOptRequest struct {
	Metadata       map[string]string         `json:"metadata,omitempty"`
	RequestTimeout string                    `json:"request_timeout,omitempty"`
	ProtoFiles     *GRPCProtoFilesDiscovery  `json:"proto_files,omitempty"`
	Reflection     *GRPCReflectionDiscovery  `json:"reflection,omitempty"`
	HealthCheck    *GRPCHealthCheckDiscovery `json:"health_check,omitempty"`
}

type GRPCAdvanceOption struct {
	RequestOptions *GRPCOptRequest     `json:"request_options,omitempty"`
	Certificate    *GRPCOptCertificate `json:"certificate,omitempty"`
	Secret         *GRPCSecret         `json:"secret,omitempty"`
}

type GRPCTask struct {
	*Task
	Server           string             `json:"server"`
	PostScript       string             `json:"post_script,omitempty"`
	SuccessWhenLogic string             `json:"success_when_logic"`
	SuccessWhen      []*GRPCSuccess     `json:"success_when"`
	AdvanceOptions   *GRPCAdvanceOption `json:"advance_options,omitempty"`

	creds credentials.TransportCredentials

	result           []byte
	reqError         string
	reqCost          time.Duration
	timeout          time.Duration
	postScriptResult *ScriptResult

	rawTask                    *GRPCTask
	healthMethodDescriptor     *pdesc.MethodDescriptor // cached method descriptor for HealthCheck discovery
	protoFilesMethodDescriptor *pdesc.MethodDescriptor // cached method descriptor for ProtoFiles discovery
	reflectionMethodDescriptor *pdesc.MethodDescriptor // cached method descriptor for Reflection discovery
}

func (t *GRPCTask) initTask() {
	if t.Task == nil {
		t.Task = &Task{}
	}
}

func (t *GRPCTask) check() error {
	if t.Server == "" {
		return fmt.Errorf("server address is required")
	}
	if t.AdvanceOptions != nil &&
		t.AdvanceOptions.RequestOptions != nil &&
		t.AdvanceOptions.RequestOptions.ProtoFiles != nil &&
		len(t.AdvanceOptions.RequestOptions.ProtoFiles.ProtoFiles) == 0 {
		return fmt.Errorf("proto files not provided")
	}
	if t.getFullMethod() == "" {
		return fmt.Errorf("full method is required")
	}
	if len(t.SuccessWhen) == 0 && t.PostScript == "" {
		return fmt.Errorf(`no any check rule`)
	}

	return nil
}

func (t *GRPCTask) init() error {
	if t.AdvanceOptions == nil || t.AdvanceOptions.RequestOptions == nil {
		return fmt.Errorf("advance options required")
	}
	opt := t.AdvanceOptions
	reqOpt := opt.RequestOptions

	t.timeout = DefaultGRPCTimeout
	if reqOpt.RequestTimeout != "" {
		timeout, err := time.ParseDuration(reqOpt.RequestTimeout)
		if err != nil {
			return fmt.Errorf("invalid timeout %q: %w", reqOpt.RequestTimeout, err)
		}
		t.timeout = timeout
	}

	// init success checker
	for _, checker := range t.SuccessWhen {
		if checker == nil {
			continue
		}
		if checker.ResponseTime != "" {
			du, err := time.ParseDuration(checker.ResponseTime)
			if err != nil {
				return fmt.Errorf("invalid response time %q: %w", checker.ResponseTime, err)
			}
			checker.respTime = du
		}

		// body
		for _, v := range checker.Body {
			if v == nil {
				continue
			}
			if err := genReg(v); err != nil {
				return fmt.Errorf("compile regex failed: %w", err)
			}
		}
	}

	// setup transport credentials
	var err error
	t.creds, err = t.buildTLSCredentials()
	if err != nil {
		return fmt.Errorf("build TLS credentials failed: %w", err)
	}

	// Cache method descriptor if using ProtoFiles discovery
	if reqOpt.ProtoFiles != nil && len(reqOpt.ProtoFiles.ProtoFiles) > 0 {
		_, err := t.findMethodAmongProtofiles()
		if err != nil {
			return fmt.Errorf("find method descriptor failed: %w", err)
		}
	}

	if reqOpt.HealthCheck != nil {
		_, err := t.findHealthCheckMethod()
		if err != nil {
			return fmt.Errorf("find health check method failed: %w", err)
		}
	}

	return nil
}

func (t *GRPCTask) buildTLSCredentials() (credentials.TransportCredentials, error) {
	opt := t.AdvanceOptions
	if opt == nil || opt.Certificate == nil {
		return insecure.NewCredentials(), nil
	}

	cert := opt.Certificate

	// if ignore server certificate error, use insecure TLS config
	if cert.IgnoreServerCertificateError {
		config := &tls.Config{
			InsecureSkipVerify: true, //nolint:gosec
		}
		return credentials.NewTLS(config), nil
	}

	// if CA cert is provided, setup mTLS
	if cert.CaCert != "" {
		caCertPool := x509.NewCertPool()
		if !caCertPool.AppendCertsFromPEM([]byte(cert.CaCert)) {
			return nil, fmt.Errorf("failed to append CA certificate")
		}

		config := &tls.Config{
			RootCAs:    caCertPool,
			MinVersion: tls.VersionTLS12,
		}

		// if client certificate and private key are provided, add them for mTLS
		if cert.Certificate != "" && cert.PrivateKey != "" {
			clientCert, err := tls.X509KeyPair([]byte(cert.Certificate), []byte(cert.PrivateKey))
			if err != nil {
				return nil, fmt.Errorf("failed to load client certificate: %w", err)
			}
			config.Certificates = []tls.Certificate{clientCert}
		}

		return credentials.NewTLS(config), nil
	}

	return insecure.NewCredentials(), nil
}

func (t *GRPCTask) findMethod(ctx context.Context, conn *grpc.ClientConn) (*pdesc.MethodDescriptor, error) {
	opt := t.AdvanceOptions
	if opt == nil || opt.RequestOptions == nil {
		return nil, fmt.Errorf("request options required")
	}

	reqOpt := opt.RequestOptions

	if reqOpt.ProtoFiles != nil {
		if len(reqOpt.ProtoFiles.ProtoFiles) == 0 {
			return nil, fmt.Errorf("proto files not provided")
		}
		return t.findMethodAmongProtofiles()
	}

	if reqOpt.Reflection != nil {
		return t.findMethodByReflection(ctx, conn)
	}

	if reqOpt.HealthCheck != nil {
		return t.findHealthCheckMethod()
	}

	return nil, fmt.Errorf("no discovery method configured (proto_files, reflection, or health_check)")
}

func (t *GRPCTask) findHealthCheckMethod() (*pdesc.MethodDescriptor, error) {
	if t.healthMethodDescriptor != nil {
		return t.healthMethodDescriptor, nil
	}

	healthFD := grpc_health_v1.File_grpc_health_v1_health_proto
	if healthFD == nil {
		return nil, fmt.Errorf("health check file descriptor not available")
	}

	fd, err := pdesc.WrapFile(healthFD)
	if err != nil {
		return nil, fmt.Errorf("wrap health check file descriptor failed: %w", err)
	}

	sd := fd.FindService(HealthCheckServiceName)
	if sd == nil {
		return nil, fmt.Errorf("health check service %s not found", HealthCheckServiceName)
	}

	md := sd.FindMethodByName(HealthCheckMethodName)
	if md == nil {
		return nil, fmt.Errorf("health check method %s not found", HealthCheckMethodName)
	}
	t.healthMethodDescriptor = md

	return md, nil
}

func (t *GRPCTask) findMethodByReflection(ctx context.Context, conn *grpc.ClientConn) (*pdesc.MethodDescriptor, error) {
	if t.reflectionMethodDescriptor != nil {
		return t.reflectionMethodDescriptor, nil
	}

	opt := t.AdvanceOptions
	if opt == nil || opt.RequestOptions == nil || opt.RequestOptions.Reflection == nil {
		return nil, fmt.Errorf("reflection discovery not configured")
	}

	fullMethod := opt.RequestOptions.Reflection.FullMethod
	if fullMethod == "" {
		return nil, fmt.Errorf("full method is required for reflection discovery")
	}
	fullMethod = strings.TrimPrefix(fullMethod, "/")

	rc := grpcreflect.NewClientAuto(ctx, conn)
	defer rc.Reset()

	slash := strings.LastIndex(fullMethod, "/")
	if slash == -1 {
		return nil, fmt.Errorf("invalid full method name: %s", fullMethod)
	}
	serviceName := fullMethod[:slash]

	fd, err := rc.FileContainingSymbol(serviceName)
	if err != nil {
		return nil, err
	}

	sd := fd.FindService(serviceName)
	if sd == nil {
		return nil, fmt.Errorf("service %s not found", serviceName)
	}

	methodName := fullMethod[slash+1:]
	md := sd.FindMethodByName(methodName)
	if md == nil {
		return nil, fmt.Errorf("method %s not found in service %s", methodName, serviceName)
	}
	t.reflectionMethodDescriptor = md
	return md, nil
}

func (t *GRPCTask) findMethodAmongProtofiles() (*pdesc.MethodDescriptor, error) {
	// Return cached method descriptor if available
	if t.protoFilesMethodDescriptor != nil {
		return t.protoFilesMethodDescriptor, nil
	}

	opt := t.AdvanceOptions
	if opt == nil || opt.RequestOptions == nil || opt.RequestOptions.ProtoFiles == nil {
		return nil, fmt.Errorf("proto files discovery not configured")
	}

	protoFiles := opt.RequestOptions.ProtoFiles.ProtoFiles
	fullMethod := opt.RequestOptions.ProtoFiles.FullMethod

	if len(protoFiles) == 0 {
		return nil, fmt.Errorf("proto files not provided")
	}
	if fullMethod == "" {
		return nil, fmt.Errorf("full method is required for proto files discovery")
	}
	fullMethod = strings.TrimPrefix(fullMethod, "/")

	extendedMap, err := buildExtendedProtoMap(protoFiles)
	if err != nil {
		return nil, err
	}

	p := protoparse.Parser{
		Accessor:         protoparse.FileContentsFromMap(extendedMap),
		InferImportPaths: true,
	}

	desc, err := p.ParseFiles(getFileNames(protoFiles)...)
	if err != nil {
		return nil, fmt.Errorf("parse proto files failed: %w", err)
	}

	sepIdx := strings.LastIndex(fullMethod, "/")
	if sepIdx == -1 {
		return nil, fmt.Errorf("invalid fullMethod: %q", fullMethod)
	}

	service := fullMethod[:sepIdx]
	method := fullMethod[sepIdx+1:]
	for _, fd := range desc {
		if sd := fd.FindService(service); sd != nil {
			if md := sd.FindMethodByName(method); md != nil {
				t.protoFilesMethodDescriptor = md
				return md, nil
			}
		}
	}

	return nil, fmt.Errorf("method %s not found in service %s", fullMethod, service)
}

func buildExtendedProtoMap(protoFiles map[string]string) (map[string]string, error) {
	extendedMap := make(map[string]string, len(protoFiles))
	for k, v := range protoFiles {
		extendedMap[k] = v
	}
	var missingImports []string
	for _, content := range protoFiles {
		for _, imp := range extractImports(content) {
			if _, ok := extendedMap[imp]; !ok {
				missingImports = append(missingImports, imp)
			}
		}
	}
	if len(missingImports) > 0 {
		return nil, fmt.Errorf("missing imports: %s", strings.Join(missingImports, ", "))
	}
	return extendedMap, nil
}

// extractImports extracts all import statements from proto file content.
func extractImports(content string) []string {
	var imports []string
	scanner := bufio.NewScanner(strings.NewReader(content))

	for scanner.Scan() {
		line := scanner.Text()

		// Remove inline comments (content after //)
		if commentIdx := strings.Index(line, "//"); commentIdx != -1 {
			line = line[:commentIdx]
		}

		line = strings.TrimSpace(line)

		// Skip empty lines and comment lines
		if line == "" || strings.HasPrefix(line, "//") {
			continue
		}

		// Check if it's an import statement
		if !strings.HasPrefix(line, "import ") {
			continue
		}

		// Extract content in quotes
		// Support both "import" and import formats
		line = strings.TrimPrefix(line, "import")
		line = strings.TrimSpace(line)

		// Remove semicolon if present
		line = strings.TrimSuffix(line, ";")
		line = strings.TrimSpace(line)

		// Extract quoted content
		if strings.HasPrefix(line, `"`) && strings.HasSuffix(line, `"`) {
			importPath := line[1 : len(line)-1]
			if importPath != "" {
				imports = append(imports, importPath)
			}
		} else if start := strings.Index(line, `"`); start != -1 {
			// Handle case where quotes are not at the beginning
			end := strings.LastIndex(line, `"`)
			if end > start {
				importPath := line[start+1 : end]
				if importPath != "" {
					imports = append(imports, importPath)
				}
			}
		}
	}

	return imports
}

func getFileNames(files map[string]string) []string {
	arr := make([]string, 0, len(files))
	for k := range files {
		arr = append(arr, k)
	}
	return arr
}

func (t *GRPCTask) getFullMethod() string {
	if t.AdvanceOptions == nil || t.AdvanceOptions.RequestOptions == nil {
		return ""
	}
	reqOpt := t.AdvanceOptions.RequestOptions
	if reqOpt.ProtoFiles != nil {
		return reqOpt.ProtoFiles.FullMethod
	}
	if reqOpt.Reflection != nil {
		return reqOpt.Reflection.FullMethod
	}
	if reqOpt.HealthCheck != nil {
		return fmt.Sprintf("%s/%s", HealthCheckServiceName, HealthCheckMethodName)
	}
	return ""
}

func (t *GRPCTask) getJSONRequest() string {
	if t.AdvanceOptions == nil || t.AdvanceOptions.RequestOptions == nil {
		return ""
	}
	reqOpt := t.AdvanceOptions.RequestOptions
	if reqOpt.ProtoFiles != nil {
		return reqOpt.ProtoFiles.JSONRequest
	}
	if reqOpt.Reflection != nil {
		return reqOpt.Reflection.JSONRequest
	}
	if reqOpt.HealthCheck != nil && reqOpt.HealthCheck.Service != "" {
		healthReq := map[string]string{"service": reqOpt.HealthCheck.Service}
		jsonReq, _ := json.Marshal(healthReq)
		return string(jsonReq)
	}
	return ""
}

func (t *GRPCTask) run() error {
	opt := t.AdvanceOptions
	if opt == nil || opt.RequestOptions == nil {
		t.reqError = "request options required"
		return nil
	}

	// Create connection (new connection for each run())
	ctx, cancel := context.WithTimeout(context.Background(), t.timeout)
	defer cancel()

	dialOpts := []grpc.DialOption{
		grpc.WithTransportCredentials(t.creds),
		grpc.WithBlock(),
	}

	conn, err := grpc.DialContext(ctx, t.Server, dialOpts...)
	if err != nil {
		t.reqError = fmt.Sprintf("dial grpc server failed: %v", err)
		return nil
	}
	defer func() {
		_ = conn.Close()
	}()

	// Find method
	method, err := t.findMethod(ctx, conn)
	if err != nil {
		t.reqError = err.Error()
		return nil
	}

	reqOpt := opt.RequestOptions

	// Add metadata
	if len(reqOpt.Metadata) > 0 {
		md := metadata.New(reqOpt.Metadata)
		ctx = metadata.NewOutgoingContext(ctx, md)
	}

	// Build request message
	msg := dynamic.NewMessage(method.GetInputType())

	jsonRequest := t.getJSONRequest()
	if jsonRequest != "" {
		if err := msg.UnmarshalJSON([]byte(jsonRequest)); err != nil {
			t.reqError = fmt.Sprintf("invalid message: %v", err)
			return nil
		}
	}

	// Execute RPC call
	rpcStart := time.Now()
	stub := grpcdynamic.NewStub(conn)
	resp, err := stub.InvokeRpc(ctx, method, msg)
	t.reqCost = time.Since(rpcStart)
	if err != nil {
		t.reqError = err.Error()
		return nil
	}

	// dial test message
	dynMsg, ok := resp.(*dynamic.Message)
	if !ok {
		t.reqError = fmt.Sprintf("unexpected response type: expected *dynamic.Message, got %T", resp)
		return nil
	}

	j, err := dynMsg.MarshalJSON()
	if err != nil {
		t.reqError = fmt.Sprintf("marshal response failed: %v", err)
		return nil
	}
	t.result = j

	// run post script if provided
	if t.PostScript != "" {
		result, err := postScriptDoGRPC(t.PostScript, t.result)
		if err != nil {
			t.reqError = err.Error()
			return nil
		}
		t.postScriptResult = result
	}

	return nil
}

func (t *GRPCTask) stop() {
	// close connection in run()
}

func (t *GRPCTask) clear() {
	t.result = nil
	t.reqError = ""
	t.reqCost = 0
	t.postScriptResult = nil
	if t.timeout == 0 {
		t.timeout = DefaultGRPCTimeout
	}
}

func (t *GRPCTask) class() string {
	return ClassGRPC
}

func (t *GRPCTask) metricName() string {
	return "grpc_dial_testing"
}

func (t *GRPCTask) checkResult() ([]string, bool) {
	var reasons []string
	var succFlag bool

	if t.reqError != "" {
		return []string{t.reqError}, false
	}
	if t.result == nil {
		return []string{"no response"}, false
	}

	// if no success conditions defined, default to success if no error
	if len(t.SuccessWhen) == 0 && t.PostScript == "" {
		return nil, true
	}

	// check SuccessWhen conditions
	for _, chk := range t.SuccessWhen {
		if chk == nil {
			continue
		}
		// check body
		for _, v := range chk.Body {
			if v == nil {
				continue
			}
			if err := v.check(string(t.result), "response body"); err != nil {
				reasons = append(reasons, err.Error())
			} else {
				succFlag = true
			}
		}

		// check response time
		if chk.respTime > 0 && t.reqCost > chk.respTime {
			reasons = append(reasons,
				fmt.Sprintf("gRPC response time(%v) larger than %v", t.reqCost, chk.respTime))
		} else if chk.respTime > 0 {
			succFlag = true
		}
	}

	// check post script result
	if t.postScriptResult != nil {
		if t.postScriptResult.Result.IsFailed {
			reasons = append(reasons, t.postScriptResult.Result.ErrorMessage)
		} else {
			succFlag = true
		}
	}

	return reasons, succFlag
}

func (t *GRPCTask) getResults() (tags map[string]string, fields map[string]interface{}) {
	tags = map[string]string{
		"name":   t.Name,
		"server": t.Server,
		"method": t.getFullMethod(),
		"status": "FAIL",
		"proto":  "grpc",
	}

	fields = map[string]interface{}{
		"response_time": int64(t.reqCost) / 1000,
		"success":       int64(-1),
	}

	if hostnames, err := t.getHostName(); err == nil && len(hostnames) > 0 {
		tags["dest_host"] = hostnames[0]
	}

	for k, v := range t.Tags {
		tags[k] = v
	}

	message := map[string]interface{}{}

	reasons, succFlag := t.checkResult()

	// check if we should save response body
	notSave := false
	if t.AdvanceOptions != nil && t.AdvanceOptions.Secret != nil && t.AdvanceOptions.Secret.NoSaveResponseBody {
		notSave = true
	}

	// apply SuccessWhenLogic
	switch t.SuccessWhenLogic {
	case "or":
		if succFlag && t.reqError == "" {
			tags["status"] = "OK"
			fields["success"] = int64(1)
		} else {
			message["fail_reason"] = strings.Join(reasons, ";")
			fields["fail_reason"] = strings.Join(reasons, ";")
		}
	default: // "and" or empty (default to "and")
		if succFlag && len(reasons) == 0 && t.reqError == "" {
			tags["status"] = "OK"
			fields["success"] = int64(1)
		} else {
			message["fail_reason"] = strings.Join(reasons, ";")
			fields["fail_reason"] = strings.Join(reasons, ";")
		}
	}

	message["response_time"] = int64(t.reqCost) / 1000
	if t.result != nil && !notSave {
		message["response"] = string(t.result)
	}

	data, err := json.Marshal(message)
	if err != nil {
		fields["message"] = err.Error()
	} else {
		if len(data) > MaxMsgSize {
			fields["message"] = string(data[:MaxMsgSize])
		} else {
			fields["message"] = string(data)
		}
	}

	return tags, fields
}

func (t *GRPCTask) getVariableValue(variable Variable) (string, error) {
	if variable.PostScript == "" && t.PostScript == "" {
		return "", fmt.Errorf("post_script is empty")
	}

	if variable.TaskVarName == "" {
		return "", fmt.Errorf("task variable name is empty")
	}

	if t.result == nil {
		return "", fmt.Errorf("response body is empty")
	}

	var result *ScriptResult
	var err error
	if variable.PostScript == "" { // use task post script
		result = t.postScriptResult
	} else { // use task variable post script
		if result, err = postScriptDoGRPC(variable.PostScript, t.result); err != nil {
			return "", fmt.Errorf("run pipeline failed: %w", err)
		}
	}

	if result == nil {
		return "", fmt.Errorf("pipeline result is empty")
	}

	value, ok := result.Vars[variable.TaskVarName]
	if !ok {
		return "", fmt.Errorf("task variable name not found")
	}
	return fmt.Sprintf("%v", value), nil
}

func (t *GRPCTask) getHostName() ([]string, error) {
	if t.Server == "" {
		return nil, fmt.Errorf("server address is empty")
	}

	host, _, err := net.SplitHostPort(t.Server)
	if err == nil {
		return []string{host}, nil
	}

	return []string{t.Server}, nil
}

func (t *GRPCTask) getRawTask(taskString string) (string, error) {
	task := GRPCTask{}

	if err := json.Unmarshal([]byte(taskString), &task); err != nil {
		return "", fmt.Errorf("unmarshal grpc task failed: %w", err)
	}

	task.Task = nil

	bytes, err := json.Marshal(task)
	if err != nil {
		return "", fmt.Errorf("marshal grpc task failed: %w", err)
	}
	return string(bytes), nil
}

func (t *GRPCTask) renderSuccessWhen(task *GRPCTask, fm template.FuncMap) error {
	if task == nil {
		return nil
	}

	if task.SuccessWhen != nil {
		for index, checker := range task.SuccessWhen {
			if checker == nil {
				continue
			}

			// body
			if checker.Body != nil {
				for bodyIndex, v := range checker.Body {
					if v != nil {
						if err := t.renderSuccessOption(v, t.SuccessWhen[index].Body[bodyIndex], fm); err != nil {
							return fmt.Errorf("render body failed: %w", err)
						}
					}
				}
			}

			// response time
			if checker.ResponseTime != "" {
				responseTime, err := t.GetParsedString(checker.ResponseTime, fm)
				if err != nil {
					return fmt.Errorf("render response time failed: %w", err)
				}
				t.SuccessWhen[index].ResponseTime = responseTime
			}
		}
	}

	return nil
}

func (t *GRPCTask) setReqError(err string) {
	t.reqError = err
}

func (t *GRPCTask) renderTemplate(fm template.FuncMap) error {
	if t.rawTask == nil {
		task := &GRPCTask{}
		if err := t.NewRawTask(task); err != nil {
			return fmt.Errorf("new raw task failed: %w", err)
		}
		t.rawTask = task
	}

	task := t.rawTask
	if task == nil {
		return errors.New("raw task is nil")
	}

	// server
	server, err := t.GetParsedString(task.Server, fm)
	if err != nil {
		return fmt.Errorf("render server failed: %w", err)
	}
	t.Server = server

	// success when
	if err := t.renderSuccessWhen(task, fm); err != nil {
		return fmt.Errorf("render success when failed: %w", err)
	}

	// advance options
	if err := t.renderAdvanceOptions(task, fm); err != nil {
		return fmt.Errorf("render advance options failed: %w", err)
	}

	return nil
}

func (t *GRPCTask) renderAdvanceOptions(task *GRPCTask, fm template.FuncMap) error {
	if task == nil || task.AdvanceOptions == nil {
		return nil
	}

	// request options
	if err := t.renderRequestOptions(task.AdvanceOptions.RequestOptions, fm); err != nil {
		return fmt.Errorf("render request options failed: %w", err)
	}

	return nil
}

func (t *GRPCTask) renderRequestOptions(requestOpt *GRPCOptRequest, fm template.FuncMap) error {
	if requestOpt == nil {
		return nil
	}

	// request timeout
	if requestOpt.RequestTimeout != "" {
		timeout, err := t.GetParsedString(requestOpt.RequestTimeout, fm)
		if err != nil {
			return fmt.Errorf("render timeout failed: %w", err)
		}
		t.AdvanceOptions.RequestOptions.RequestTimeout = timeout
	}

	// metadata
	if len(requestOpt.Metadata) > 0 {
		for k, v := range requestOpt.Metadata {
			key, err := t.GetParsedString(k, fm)
			if err != nil {
				return fmt.Errorf("render metadata key %q failed: %w", k, err)
			}
			value, err := t.GetParsedString(v, fm)
			if err != nil {
				return fmt.Errorf("render metadata value for key %q failed: %w", k, err)
			}
			delete(t.AdvanceOptions.RequestOptions.Metadata, k)
			t.AdvanceOptions.RequestOptions.Metadata[key] = value
		}
	}

	// proto files discovery
	if err := t.renderProtoFiles(requestOpt.ProtoFiles, fm); err != nil {
		return fmt.Errorf("render proto files failed: %w", err)
	}

	// reflection discovery
	if err := t.renderReflection(requestOpt.Reflection, fm); err != nil {
		return fmt.Errorf("render reflection failed: %w", err)
	}

	// health check discovery
	if err := t.renderHealthCheck(requestOpt.HealthCheck, fm); err != nil {
		return fmt.Errorf("render health check failed: %w", err)
	}

	return nil
}

func (t *GRPCTask) renderProtoFiles(protoFiles *GRPCProtoFilesDiscovery, fm template.FuncMap) error {
	if protoFiles == nil {
		return nil
	}

	if protoFiles.FullMethod != "" {
		fullMethod, err := t.GetParsedString(protoFiles.FullMethod, fm)
		if err != nil {
			return fmt.Errorf("render proto files full method failed: %w", err)
		}
		// if full method is changed, clear the cached method descriptor
		if t.AdvanceOptions.RequestOptions.ProtoFiles.FullMethod != fullMethod {
			t.protoFilesMethodDescriptor = nil
		}
		t.AdvanceOptions.RequestOptions.ProtoFiles.FullMethod = fullMethod
	}

	if protoFiles.JSONRequest != "" {
		jsonRequest, err := t.GetParsedString(protoFiles.JSONRequest, fm)
		if err != nil {
			return fmt.Errorf("render proto files JSON request failed: %w", err)
		}
		t.AdvanceOptions.RequestOptions.ProtoFiles.JSONRequest = jsonRequest
	}

	return nil
}

func (t *GRPCTask) renderReflection(reflection *GRPCReflectionDiscovery, fm template.FuncMap) error {
	if reflection == nil {
		return nil
	}

	if reflection.FullMethod != "" {
		fullMethod, err := t.GetParsedString(reflection.FullMethod, fm)
		if err != nil {
			return fmt.Errorf("render reflection full method failed: %w", err)
		}
		// if full method is changed, clear the cached method descriptor
		if t.AdvanceOptions.RequestOptions.Reflection.FullMethod != fullMethod {
			t.reflectionMethodDescriptor = nil
		}
		t.AdvanceOptions.RequestOptions.Reflection.FullMethod = fullMethod
	}

	if reflection.JSONRequest != "" {
		jsonRequest, err := t.GetParsedString(reflection.JSONRequest, fm)
		if err != nil {
			return fmt.Errorf("render reflection JSON request failed: %w", err)
		}
		t.AdvanceOptions.RequestOptions.Reflection.JSONRequest = jsonRequest
	}

	return nil
}

func (t *GRPCTask) renderHealthCheck(healthCheck *GRPCHealthCheckDiscovery, fm template.FuncMap) error {
	if healthCheck == nil {
		return nil
	}

	if healthCheck.Service != "" {
		service, err := t.GetParsedString(healthCheck.Service, fm)
		if err != nil {
			return fmt.Errorf("render health check service failed: %w", err)
		}
		t.AdvanceOptions.RequestOptions.HealthCheck.Service = service
	}

	return nil
}
