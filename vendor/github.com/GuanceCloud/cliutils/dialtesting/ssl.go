// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package dialtesting

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"strings"
	"text/template"
	"time"
)

var (
	_ TaskChild = (*SSLTask)(nil)
	_ ITask     = (*SSLTask)(nil)
)

type SSLSuccess struct {
	ResponseTime             string           `json:"response_time,omitempty"`
	CertificateExpiresInDays []*ValueSuccess  `json:"ssl_cert_expires_in_days,omitempty"`
	CertificateNotAfter      []*ValueSuccess  `json:"ssl_cert_not_after,omitempty"`
	Subject                  []*SuccessOption `json:"subject,omitempty"`
	Issuer                   []*SuccessOption `json:"issuer,omitempty"`
	TLSVersion               []*SuccessOption `json:"tls_version,omitempty"`

	respTime time.Duration
}

type SSLTask struct {
	*Task
	Host                         string        `json:"host"`
	Port                         string        `json:"port"`
	ServerName                   string        `json:"server_name,omitempty"`
	Timeout                      string        `json:"timeout,omitempty"`
	IgnoreServerCertificateError bool          `json:"ignore_server_certificate_error,omitempty"`
	SuccessWhen                  []*SSLSuccess `json:"success_when"`
	SuccessWhenLogic             string        `json:"success_when_logic"`

	reqCost              time.Duration
	tlsHandshakeCost     time.Duration
	reqError             string
	destIP               string
	timeout              time.Duration
	tlsVersion           string
	sslCertNotBefore     int64
	sslCertNotAfter      int64
	sslCertExpiresInDays int64
	certSubject          string
	certIssuer           string

	rawTask *SSLTask
}

func (t *SSLTask) init() error {
	if t.Timeout == "" {
		t.timeout = 10 * time.Second
	} else {
		timeout, err := time.ParseDuration(t.Timeout)
		if err != nil {
			return err
		}
		t.timeout = timeout
	}

	if len(t.SuccessWhen) == 0 {
		return errors.New(`no any check rule`)
	}

	for _, checker := range t.SuccessWhen {
		if checker.ResponseTime != "" {
			du, err := time.ParseDuration(checker.ResponseTime)
			if err != nil {
				return err
			}
			checker.respTime = du
		}

		for _, v := range checker.Subject {
			if err := genReg(v); err != nil {
				return err
			}
		}
		for _, v := range checker.Issuer {
			if err := genReg(v); err != nil {
				return err
			}
		}
		for _, v := range checker.TLSVersion {
			if err := genReg(v); err != nil {
				return err
			}
		}
	}

	return nil
}

func (t *SSLTask) check() error {
	if t.Host == "" {
		return errors.New("host should not be empty")
	}
	if t.Port == "" {
		return errors.New("port should not be empty")
	}
	return nil
}

func (t *SSLTask) checkResult() (reasons []string, succFlag bool) {
	for _, chk := range t.SuccessWhen {
		if chk.respTime > 0 {
			if t.reqCost >= chk.respTime {
				reasons = append(reasons, fmt.Sprintf("SSL response time(%v) larger equal than %v", t.reqCost, chk.respTime))
			} else {
				succFlag = true
			}
		}

		for _, v := range chk.CertificateExpiresInDays {
			if err := v.check(float64(t.sslCertExpiresInDays)); err != nil {
				reasons = append(reasons, fmt.Sprintf("SSL certificate expires in days check failed: %s", err.Error()))
			} else {
				succFlag = true
			}
		}

		for _, v := range chk.CertificateNotAfter {
			if err := v.check(float64(t.sslCertNotAfter)); err != nil {
				reasons = append(reasons, fmt.Sprintf("SSL certificate not after check failed: %s", err.Error()))
			} else {
				succFlag = true
			}
		}

		for _, v := range chk.Subject {
			if err := v.check(t.certSubject, "SSL certificate subject"); err != nil {
				reasons = append(reasons, err.Error())
			} else {
				succFlag = true
			}
		}

		for _, v := range chk.Issuer {
			if err := v.check(t.certIssuer, "SSL certificate issuer"); err != nil {
				reasons = append(reasons, err.Error())
			} else {
				succFlag = true
			}
		}

		for _, v := range chk.TLSVersion {
			if err := v.check(t.tlsVersion, "TLS version"); err != nil {
				reasons = append(reasons, err.Error())
			} else {
				succFlag = true
			}
		}
	}

	return reasons, succFlag
}

func (t *SSLTask) getResults() (tags map[string]string, fields map[string]interface{}) {
	tags = map[string]string{
		"name":      t.Name,
		"dest_host": t.Host,
		"dest_port": t.Port,
		"dest_ip":   t.destIP,
		"status":    "FAIL",
		"proto":     "ssl",
	}

	if t.ServerName != "" {
		tags["server_name"] = t.ServerName
	}

	for k, v := range t.Tags {
		tags[k] = v
	}

	responseTime := int64(t.reqCost) / 1000
	tlsHandshakeTime := int64(t.tlsHandshakeCost) / 1000
	fields = map[string]interface{}{
		"response_time":      responseTime,
		"tls_handshake_time": tlsHandshakeTime,
		"success":            int64(-1),
	}

	if t.tlsVersion != "" {
		fields["tls_version"] = t.tlsVersion
	}
	if t.sslCertNotBefore > 0 {
		fields["ssl_cert_not_before"] = t.sslCertNotBefore
	}
	if t.sslCertNotAfter > 0 {
		fields["ssl_cert_not_after"] = t.sslCertNotAfter
		fields["ssl_cert_expires_in_days"] = t.sslCertExpiresInDays
	}
	if t.certSubject != "" {
		fields["ssl_cert_subject"] = t.certSubject
	}
	if t.certIssuer != "" {
		fields["ssl_cert_issuer"] = t.certIssuer
	}

	message := map[string]interface{}{}
	var (
		reasons  []string
		succFlag bool
	)
	if t.reqError != "" {
		reasons = append(reasons, t.reqError)
	} else {
		reasons, succFlag = t.checkResult()
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

	message["status"] = tags["status"]
	data, err := json.Marshal(message)
	if err != nil {
		fields[`message`] = err.Error()
	} else if len(data) > MaxMsgSize {
		fields[`message`] = string(data[:MaxMsgSize])
	} else {
		fields[`message`] = string(data)
	}

	return tags, fields
}

func (t *SSLTask) metricName() string {
	return `ssl_dial_testing`
}

func (t *SSLTask) clear() {
	t.reqCost = 0
	t.tlsHandshakeCost = 0
	t.reqError = ""
	t.destIP = ""
	t.tlsVersion = ""
	t.sslCertNotBefore = 0
	t.sslCertNotAfter = 0
	t.sslCertExpiresInDays = 0
	t.certSubject = ""
	t.certIssuer = ""
}

func (t *SSLTask) run() error {
	d := net.Dialer{Timeout: t.timeout}
	ctx, cancel := context.WithTimeout(context.Background(), t.timeout)
	defer cancel()

	addr := net.JoinHostPort(t.Host, t.Port)
	serverName := t.ServerName
	if serverName == "" {
		serverName = t.Host
	}

	cfg := &tls.Config{
		MinVersion:         tls.VersionTLS12,
		ServerName:         serverName,
		InsecureSkipVerify: t.IgnoreServerCertificateError, //nolint:gosec
	}

	start := time.Now()
	rawConn, err := d.DialContext(ctx, "tcp", addr)
	if err != nil {
		t.reqError = err.Error()
		return nil
	}
	conn := tls.Client(rawConn, cfg)
	handshakeStart := time.Now()
	if err := conn.HandshakeContext(ctx); err != nil {
		_ = conn.Close()
		t.reqError = err.Error()
		return nil
	}
	t.tlsHandshakeCost = time.Since(handshakeStart)
	defer conn.Close() //nolint:errcheck

	t.reqCost = time.Since(start)
	if remote := conn.RemoteAddr(); remote != nil {
		if host, _, err := net.SplitHostPort(remote.String()); err == nil {
			t.destIP = host
		}
	}

	cs := conn.ConnectionState()
	t.tlsVersion = tlsVersionString(cs.Version)
	t.extractCertificateInfo(cs)
	return nil
}

func (t *SSLTask) stop() {}

func (t *SSLTask) class() string {
	return ClassSSL
}

func (t *SSLTask) getHostName() ([]string, error) {
	return []string{t.Host}, nil
}

func (t *SSLTask) getVariableValue(variable Variable) (string, error) {
	return "", errors.New("not support")
}

func (t *SSLTask) getRawTask(taskString string) (string, error) {
	task := SSLTask{}

	if err := json.Unmarshal([]byte(taskString), &task); err != nil {
		return "", fmt.Errorf("unmarshal ssl task failed: %w", err)
	}

	task.Task = nil

	bytes, _ := json.Marshal(task)
	return string(bytes), nil
}

func (t *SSLTask) initTask() {
	if t.Task == nil {
		t.Task = &Task{}
	}
}

func (t *SSLTask) setReqError(err string) {
	t.reqError = err
}

func (t *SSLTask) renderTemplate(fm template.FuncMap) error {
	if t.rawTask == nil {
		task := &SSLTask{}
		if err := t.NewRawTask(task); err != nil {
			return fmt.Errorf("new raw task failed: %w", err)
		}
		t.rawTask = task
	}

	task := t.rawTask
	if task == nil {
		return errors.New("raw task is nil")
	}

	if text, err := t.GetParsedString(task.Host, fm); err != nil {
		return fmt.Errorf("render host failed: %w", err)
	} else {
		t.Host = text
	}

	if text, err := t.GetParsedString(task.Port, fm); err != nil {
		return fmt.Errorf("render port failed: %w", err)
	} else {
		t.Port = text
	}

	if text, err := t.GetParsedString(task.ServerName, fm); err != nil {
		return fmt.Errorf("render server name failed: %w", err)
	} else {
		t.ServerName = text
	}

	if err := t.renderSuccessWhen(task, fm); err != nil {
		return err
	}

	return nil
}

func (t *SSLTask) renderSuccessWhen(task *SSLTask, fm template.FuncMap) error {
	if task == nil {
		return nil
	}

	for index, checker := range task.SuccessWhen {
		if text, err := t.GetParsedString(checker.ResponseTime, fm); err != nil {
			return fmt.Errorf("render response time failed: %w", err)
		} else {
			t.SuccessWhen[index].ResponseTime = text
		}

		for optionIndex, option := range checker.Subject {
			if err := t.renderSuccessOption(option, t.SuccessWhen[index].Subject[optionIndex], fm); err != nil {
				return fmt.Errorf("render subject failed: %w", err)
			}
		}

		for optionIndex, option := range checker.Issuer {
			if err := t.renderSuccessOption(option, t.SuccessWhen[index].Issuer[optionIndex], fm); err != nil {
				return fmt.Errorf("render issuer failed: %w", err)
			}
		}

		for optionIndex, option := range checker.TLSVersion {
			if err := t.renderSuccessOption(option, t.SuccessWhen[index].TLSVersion[optionIndex], fm); err != nil {
				return fmt.Errorf("render tls version failed: %w", err)
			}
		}
	}

	return nil
}

func (t *SSLTask) extractCertificateInfo(cs tls.ConnectionState) {
	if len(cs.PeerCertificates) == 0 {
		return
	}

	cert := cs.PeerCertificates[0]
	t.sslCertNotBefore = cert.NotBefore.UnixMicro()
	t.sslCertNotAfter = cert.NotAfter.UnixMicro()
	t.sslCertExpiresInDays = (t.sslCertNotAfter - time.Now().UnixMicro()) / (24 * time.Hour).Microseconds()
	t.certSubject = cert.Subject.String()
	t.certIssuer = cert.Issuer.String()
}

func tlsVersionString(version uint16) string {
	switch version {
	case tls.VersionTLS10:
		return "TLS1.0"
	case tls.VersionTLS11:
		return "TLS1.1"
	case tls.VersionTLS12:
		return "TLS1.2"
	case tls.VersionTLS13:
		return "TLS1.3"
	default:
		return fmt.Sprintf("0x%x", version)
	}
}
