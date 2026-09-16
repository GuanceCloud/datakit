// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package cmds

import (
	"errors"
	"testing"

	"github.com/kardianos/service"
	"github.com/stretchr/testify/require"
)

type serviceControlStub struct {
	service.Service
	platform  string
	status    service.Status
	statusErr error
	actionErr error
	calls     []string
}

func (s *serviceControlStub) Platform() string { return s.platform }

func (s *serviceControlStub) Status() (service.Status, error) {
	s.calls = append(s.calls, "status")
	return s.status, s.statusErr
}

func (s *serviceControlStub) Start() error {
	s.calls = append(s.calls, "start")
	return s.actionErr
}

func (s *serviceControlStub) Stop() error {
	s.calls = append(s.calls, "stop")
	if s.actionErr == nil {
		s.status = service.StatusStopped
	}
	return s.actionErr
}

func (s *serviceControlStub) Restart() error {
	s.calls = append(s.calls, "restart")
	return s.actionErr
}

func (s *serviceControlStub) Install() error {
	s.calls = append(s.calls, "install")
	return nil
}

func TestSystemdServiceControlRecoversFailedState(t *testing.T) {
	for name, control := range map[string]func(service.Service) error{
		"start": startService, "stop": stopService, "restart": restartService,
	} {
		t.Run(name, func(t *testing.T) {
			svc := &serviceControlStub{
				platform: "linux-systemd", status: service.StatusUnknown,
				statusErr: errors.New("service in failed state"),
			}
			require.NoError(t, control(svc))
			require.Equal(t, []string{name}, svc.calls)

			svc.calls = nil
			svc.actionErr = errors.New("systemctl control failed")
			require.ErrorIs(t, control(svc), svc.actionErr)
			require.Equal(t, []string{name}, svc.calls)
		})
	}
}

func TestNonSystemdServiceControlPreservesStatusErrors(t *testing.T) {
	for name, control := range map[string]func(service.Service) error{
		"start": startService, "stop": stopService, "restart": restartService,
	} {
		t.Run(name, func(t *testing.T) {
			svc := &serviceControlStub{platform: "unix-systemv", statusErr: service.ErrNotInstalled}
			require.ErrorIs(t, control(svc), service.ErrNotInstalled)
			require.Equal(t, []string{"status"}, svc.calls)
		})
	}
}

func TestNonSystemdServiceControl(t *testing.T) {
	svc := &serviceControlStub{platform: "unix-systemv", status: service.StatusStopped}
	require.NoError(t, stopService(svc))
	require.Equal(t, []string{"status"}, svc.calls)

	svc.calls = nil
	require.NoError(t, startService(svc))
	require.Equal(t, []string{"status", "install", "start"}, svc.calls)

	svc.status = service.StatusRunning
	svc.calls = nil
	require.NoError(t, startService(svc))
	require.Equal(t, []string{"status"}, svc.calls)

	svc.calls = nil
	require.NoError(t, restartService(svc))
	require.Equal(t, []string{"status", "stop", "status", "install", "start"}, svc.calls)

	svc.status = service.StatusRunning
	svc.calls = nil
	svc.actionErr = errors.New("stop failed")
	require.ErrorIs(t, restartService(svc), svc.actionErr)
	require.Equal(t, []string{"status", "stop"}, svc.calls)
}

func TestServiceCommandHint(t *testing.T) {
	for _, tc := range []struct {
		goos   string
		action ServiceAction
		want   string
	}{
		{"linux", ServiceActionStart, "systemctl start datakit"},
		{"linux", ServiceActionRestart, "systemctl restart datakit"},
		{"windows", ServiceActionStart, "Start-Service -Name datakit"},
		{"windows", ServiceActionRestart, "Restart-Service -Name datakit"},
		{"darwin", ServiceActionStart, "sudo launchctl load -w /Library/LaunchDaemons/com.datakit.plist"},
		{"darwin", ServiceActionRestart, "sudo launchctl unload -w /Library/LaunchDaemons/com.datakit.plist && sudo launchctl load -w /Library/LaunchDaemons/com.datakit.plist"},
	} {
		t.Run(tc.goos+"/"+string(tc.action), func(t *testing.T) {
			require.Equal(t, tc.want, serviceCommandHint(tc.goos, tc.action))
		})
	}
}
