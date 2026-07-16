// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package netpath

import (
	"crypto/sha256"
	"crypto/subtle"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"runtime"
	"strings"
	"sync"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/httpapi"
)

var activeCandidateInput = struct { //nolint:gochecknoglobals
	sync.RWMutex
	owner     *Input
	candidate *Input
	started   bool
}{}

func (ipt *Input) RegHTTPHandler() {
	setupLogger()
	if !netpathSupportedOS(runtime.GOOS) {
		l.Warnf("skip %s HTTP handler on unsupported OS %s", inputName, runtime.GOOS)
		return
	}
	if !claimNetpathInput(ipt) {
		l.Warnf("ignore duplicate %s input HTTP registration", inputName)
		return
	}
	httpapi.RemoveHTTPRoute(http.MethodPost, apiPath)
	httpapi.RegHTTPHandler(http.MethodPost, apiPath, dispatchCandidates)
	ipt.initialize()
	activeCandidateInput.Lock()
	defer activeCandidateInput.Unlock()
	if activeCandidateInput.owner != ipt {
		return
	}
	if ipt.Dynamic != nil && ipt.Dynamic.Enabled {
		activeCandidateInput.candidate = ipt
	} else {
		activeCandidateInput.candidate = nil
	}
}

func claimNetpathInput(ipt *Input) bool {
	activeCandidateInput.Lock()
	defer activeCandidateInput.Unlock()
	if activeCandidateInput.owner != nil {
		return false
	}
	activeCandidateInput.owner = ipt
	activeCandidateInput.started = false
	return true
}

func isNetpathInputOwner(ipt *Input) bool {
	activeCandidateInput.RLock()
	defer activeCandidateInput.RUnlock()
	return activeCandidateInput.owner == ipt
}

// dispatchCandidates keeps the Gin route independent from an input instance.
// Gin routes cannot be replaced after the HTTP server starts, while confd can
// stop and recreate an input at runtime.
func dispatchCandidates(w http.ResponseWriter, req *http.Request) {
	activeCandidateInput.RLock()
	ipt := activeCandidateInput.candidate
	activeCandidateInput.RUnlock()
	if ipt == nil {
		http.Error(w, "netpath dynamic candidates disabled", http.StatusNotFound)
		return
	}
	ipt.handleCandidates(w, req)
}

func deactivateCandidateInput(ipt *Input) {
	activeCandidateInput.Lock()
	defer activeCandidateInput.Unlock()
	if activeCandidateInput.owner == ipt {
		activeCandidateInput.owner = nil
		activeCandidateInput.candidate = nil
		activeCandidateInput.started = false
	}
}

func (ipt *Input) handleCandidates(w http.ResponseWriter, req *http.Request) {
	ipt.initialize()
	if ipt.Dynamic == nil || !ipt.Dynamic.Enabled {
		http.Error(w, "netpath dynamic candidates disabled", http.StatusNotFound)
		return
	}
	token := ipt.Dynamic.Token
	if strings.TrimSpace(token) == "" {
		if !requestFromLoopback(req) || !requestToLoopback(req) {
			incCandidate("dropped", "unauthorized")
			http.Error(w, "netpath token is required for non-loopback requests", http.StatusForbidden)
			return
		}
	} else {
		got := req.Header.Get(tokenHeader)
		gotHash := sha256.Sum256([]byte(got))
		tokenHash := sha256.Sum256([]byte(token))
		if subtle.ConstantTimeCompare(gotHash[:], tokenHash[:]) != 1 {
			incCandidate("dropped", "unauthorized")
			http.Error(w, "invalid netpath token", http.StatusUnauthorized)
			return
		}
	}
	defer req.Body.Close() //nolint:errcheck
	req.Body = http.MaxBytesReader(w, req.Body, ipt.Dynamic.MaxBodyBytes)

	dec := json.NewDecoder(req.Body)
	dec.UseNumber()
	payload, err := decodeCandidateRequest(dec, ipt.Dynamic.MaxTestsPerRequest)
	if err != nil {
		reason := "bad_request"
		status := http.StatusBadRequest
		if errors.Is(err, errTooManyCandidateTests) {
			reason = "too_many_tests"
			status = http.StatusRequestEntityTooLarge
		} else if strings.Contains(err.Error(), "http: request body too large") {
			reason = "body_too_large"
			status = http.StatusRequestEntityTooLarge
		}
		incCandidate("dropped", reason)
		http.Error(w, fmt.Sprintf("invalid json: %s", err.Error()), status)
		return
	}
	if err := dec.Decode(&struct{}{}); err != io.EOF {
		incCandidate("dropped", "bad_request")
		http.Error(w, "invalid json: trailing data", http.StatusBadRequest)
		return
	}
	if len(payload.Tests) == 0 {
		incCandidate("dropped", "empty_tests")
		http.Error(w, "missing tests", http.StatusBadRequest)
		return
	}
	if err := validateCandidateIdentity(payload); err != nil {
		incCandidate("dropped", "invalid_identity")
		http.Error(w, err.Error(), http.StatusRequestEntityTooLarge)
		return
	}
	if err := validateCandidateTags(payload); err != nil {
		incCandidate("dropped", "invalid_tags")
		http.Error(w, err.Error(), candidateTagValidationStatus(err))
		return
	}

	if ipt.scheduler == nil {
		incCandidate("dropped", "scheduler_not_ready")
		http.Error(w, "netpath scheduler is not ready", http.StatusServiceUnavailable)
		return
	}

	resp := candidateResponse{DropReasons: map[string]int{}}
	for _, spec := range payload.Tests {
		if reason := candidateDropReason(payload, spec, *ipt.Dynamic); reason != "" {
			resp.Dropped++
			resp.DropReasons[reason]++
			incCandidate("dropped", reason)
			continue
		}
		t, err := taskFromCandidate(payload, spec, *ipt.Dynamic)
		if err != nil {
			reason := candidateErrorReason(err)
			resp.Dropped++
			resp.DropReasons[reason]++
			incCandidate("dropped", reason)
			l.Debugf("drop netpath candidate: %s", err.Error())
			continue
		}
		if ok, reason := ipt.scheduler.enqueue(t); ok {
			resp.Accepted++
			incCandidate("accepted", "ok")
		} else {
			resp.Dropped++
			resp.DropReasons[reason]++
			incCandidate("dropped", reason)
		}
	}
	// Admission is synchronous, so inputCh has already been released here.
	// Report the dynamic contexts that were actually retained by the store.
	resp.QueueSize = ipt.scheduler.store.dynamicLen()
	observeQueues(ipt.scheduler)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		l.Warnf("write netpath candidate response failed: %s", err.Error())
	}
}

func requestFromLoopback(req *http.Request) bool {
	host, _, err := net.SplitHostPort(req.RemoteAddr)
	if err != nil {
		return false
	}
	ip := net.ParseIP(host)
	return ip != nil && ip.IsLoopback()
}

func requestToLoopback(req *http.Request) bool {
	addr, ok := req.Context().Value(http.LocalAddrContextKey).(net.Addr)
	if !ok || addr == nil {
		return false
	}
	host, _, err := net.SplitHostPort(addr.String())
	if err != nil {
		return false
	}
	ip := net.ParseIP(host)
	return ip != nil && ip.IsLoopback()
}

var errTooManyCandidateTests = errors.New("too many netpath candidate tests")

// decodeCandidateRequest bounds the tests array while it is decoded. Decoding
// directly into candidateRequest would allocate the full slice before the
// configured request limit could be checked.
func decodeCandidateRequest(dec *json.Decoder, maxTests int) (candidateRequest, error) {
	var payload candidateRequest
	if maxTests <= 0 {
		return payload, errTooManyCandidateTests
	}

	token, err := dec.Token()
	if err != nil {
		return payload, err
	}
	if delim, ok := token.(json.Delim); !ok || delim != '{' {
		return payload, errors.New("netpath candidate payload must be an object")
	}

	seenTests := false
	for dec.More() {
		fieldToken, err := dec.Token()
		if err != nil {
			return payload, err
		}
		field, ok := fieldToken.(string)
		if !ok {
			return payload, errors.New("invalid netpath candidate field name")
		}

		switch field {
		case "source":
			err = dec.Decode(&payload.Source)
		case "host":
			err = dec.Decode(&payload.Host)
		case "agent_id":
			err = dec.Decode(&payload.AgentID)
		case "sequence":
			err = dec.Decode(&payload.Sequence)
		case "sent_at":
			err = dec.Decode(&payload.SentAt)
		case "tags":
			err = dec.Decode(&payload.Tags)
		case "tests":
			if seenTests {
				return payload, errors.New("duplicate netpath candidate tests field")
			}
			seenTests = true
			payload.Tests, err = decodeCandidateTests(dec, maxTests)
		default:
			var ignored json.RawMessage
			err = dec.Decode(&ignored)
		}
		if err != nil {
			return payload, err
		}
	}

	endToken, err := dec.Token()
	if err != nil {
		return payload, err
	}
	if delim, ok := endToken.(json.Delim); !ok || delim != '}' {
		return payload, errors.New("netpath candidate payload is not terminated")
	}
	return payload, nil
}

func decodeCandidateTests(dec *json.Decoder, maxTests int) ([]candidateSpec, error) {
	token, err := dec.Token()
	if err != nil {
		return nil, err
	}
	if token == nil {
		return nil, nil
	}
	if delim, ok := token.(json.Delim); !ok || delim != '[' {
		return nil, errors.New("netpath candidate tests must be an array")
	}

	tests := make([]candidateSpec, 0, min(maxTests, 16))
	for dec.More() {
		if len(tests) >= maxTests {
			return nil, errTooManyCandidateTests
		}
		var spec candidateSpec
		if err := dec.Decode(&spec); err != nil {
			return nil, err
		}
		tests = append(tests, spec)
	}
	endToken, err := dec.Token()
	if err != nil {
		return nil, err
	}
	if delim, ok := endToken.(json.Delim); !ok || delim != ']' {
		return nil, errors.New("netpath candidate tests array is not terminated")
	}
	return tests, nil
}

type candidateTagValidationError struct {
	message  string
	tooLarge bool
}

func (e *candidateTagValidationError) Error() string {
	return e.message
}

func candidateTagValidationStatus(err error) int {
	var validationErr *candidateTagValidationError
	if errors.As(err, &validationErr) && validationErr.tooLarge {
		return http.StatusRequestEntityTooLarge
	}
	return http.StatusBadRequest
}

type candidateIdentityField struct {
	name  string
	value string
}

func validateCandidateIdentity(req candidateRequest) error {
	requestBytes, err := candidateIdentitySize([]candidateIdentityField{
		{name: "source", value: req.Source},
		{name: "host", value: req.Host},
		{name: "agent_id", value: req.AgentID},
	})
	if err != nil {
		return fmt.Errorf("invalid request identity: %w", err)
	}

	for i, spec := range req.Tests {
		testBytes, err := candidateIdentitySize([]candidateIdentityField{
			{name: "hostname", value: spec.Hostname},
			{name: "target_ip", value: spec.TargetIP},
			{name: "ip", value: spec.IP},
			{name: "dst_ip", value: spec.DstIP},
			{name: "protocol", value: spec.Protocol},
			{name: "origin", value: spec.Origin},
			{name: "namespace", value: spec.Namespace},
			{name: "source_container_id", value: spec.SourceContainerID},
			{name: "source.hostname", value: spec.Source.Hostname},
			{name: "source.ip", value: spec.Source.IP},
			{name: "source.netns", value: spec.Source.NetNS},
			{name: "source.container_id", value: spec.Source.ContainerID},
			{name: "source.process_name", value: spec.Source.ProcessName},
			{name: "source.service_name", value: spec.Source.ServiceName},
		})
		if err != nil {
			return fmt.Errorf("invalid test %d identity: %w", i, err)
		}
		if requestBytes+testBytes > maxCandidateIdentityBytesPerTest {
			return fmt.Errorf("invalid test %d identity: identity fields exceed %d bytes in total",
				i, maxCandidateIdentityBytesPerTest)
		}
	}
	return nil
}

func candidateIdentitySize(fields []candidateIdentityField) (int, error) {
	total := 0
	for _, field := range fields {
		size := len(field.value)
		if size > maxCandidateIdentityFieldBytes {
			return 0, fmt.Errorf("field %q exceeds %d bytes", field.name, maxCandidateIdentityFieldBytes)
		}
		total += size
	}
	return total, nil
}

func validateCandidateTags(req candidateRequest) error {
	if err := validateTagMap(req.Tags); err != nil {
		return fmt.Errorf("invalid request tags: %w", err)
	}
	for i, spec := range req.Tests {
		if len(req.Tags)+len(spec.Tags) > maxCandidateTags {
			return fmt.Errorf("invalid test %d tags: %w", i, &candidateTagValidationError{
				message:  fmt.Sprintf("at most %d combined request and test tags are allowed", maxCandidateTags),
				tooLarge: true,
			})
		}
		if err := validateTagMap(spec.Tags); err != nil {
			return fmt.Errorf("invalid test %d tags: %w", i, err)
		}
	}
	return nil
}

func validateTagMap(tags map[string]string) error {
	if len(tags) > maxCandidateTags {
		return &candidateTagValidationError{
			message:  fmt.Sprintf("at most %d tags are allowed", maxCandidateTags),
			tooLarge: true,
		}
	}
	for key, value := range tags {
		if strings.TrimSpace(key) == "" {
			return &candidateTagValidationError{message: "tag key must not be empty"}
		}
		if len(key) > maxCandidateTagKeyBytes {
			return &candidateTagValidationError{
				message:  fmt.Sprintf("tag key exceeds %d bytes", maxCandidateTagKeyBytes),
				tooLarge: true,
			}
		}
		if len(value) > maxCandidateTagValueBytes {
			return &candidateTagValidationError{
				message:  fmt.Sprintf("tag value for %q exceeds %d bytes", key, maxCandidateTagValueBytes),
				tooLarge: true,
			}
		}
	}
	return nil
}

func candidateErrorReason(err error) string {
	msg := err.Error()
	switch {
	case errors.Is(err, errIPv6TargetUnsupported):
		return "ipv6_unsupported"
	case strings.Contains(msg, "unsupported netpath protocol"):
		return "unsupported_protocol"
	case strings.Contains(msg, "missing port"):
		return "missing_port"
	case strings.Contains(msg, "missing target"):
		return "missing_target"
	default:
		return "invalid_target"
	}
}

func init() { //nolint:gochecknoinits
	httpapi.RegInputHTTPRouteMatcher(func(method, path string) (string, bool) {
		if method == http.MethodPost && path == apiPath {
			return inputName, true
		}
		return "", false
	})
}
