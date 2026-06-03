// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package service

import (
	"fmt"
	"testing"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/cmds"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestServiceCommand(t *testing.T) {
	cmd := NewServiceCmd()
	assert.Equal(t, "service", cmd.Use)
	assert.Equal(t, "Manage datakit service", cmd.Short)
}

func TestServiceShape(t *testing.T) {
	cmd := NewServiceCmd()

	assert.NotNil(t, cmd.PersistentFlags().Lookup("log"))
	assert.NotNil(t, cmd.Commands())

	subcommands := map[string]bool{}
	for _, subcmd := range cmd.Commands() {
		subcommands[subcmd.Name()] = true
	}

	assert.True(t, subcommands["start"])
	assert.True(t, subcommands["stop"])
	assert.True(t, subcommands["restart"])
	assert.True(t, subcommands["uninstall"])
	assert.True(t, subcommands["reinstall"])
}

func TestServiceSubcommandsSetExpectedAction(t *testing.T) {
	cases := map[string]cmds.ServiceAction{
		"start":     cmds.ServiceActionStart,
		"stop":      cmds.ServiceActionStop,
		"restart":   cmds.ServiceActionRestart,
		"uninstall": cmds.ServiceActionUninstall,
		"reinstall": cmds.ServiceActionReinstall,
	}

	for name, want := range cases {
		t.Run(name, func(t *testing.T) {
			origRunServiceFn := runServiceFn
			origLogFlag := logFlag
			origServiceAction := serviceAction
			defer func() {
				runServiceFn = origRunServiceFn
				logFlag = origLogFlag
				serviceAction = origServiceAction
			}()

			called := false
			runServiceFn = func() error {
				called = true
				assert.Equal(t, want, serviceAction)
				return nil
			}

			cmd := NewServiceCmd()
			cmd.SetArgs([]string{name})

			err := cmd.Execute()
			require.NoError(t, err)
			assert.True(t, called)
		})
	}
}

func TestServiceLogFlagIsAvailableToSubcommands(t *testing.T) {
	origRunServiceFn := runServiceFn
	origLogFlag := logFlag
	origServiceAction := serviceAction
	defer func() {
		runServiceFn = origRunServiceFn
		logFlag = origLogFlag
		serviceAction = origServiceAction
	}()

	const wantLog = "/tmp/datakit-service.log"
	runServiceFn = func() error {
		assert.Equal(t, wantLog, logFlag)
		assert.Equal(t, cmds.ServiceActionStart, serviceAction)
		return nil
	}

	for _, args := range [][]string{
		{"--log", wantLog, "start"},
		{"start", "--log", wantLog},
	} {
		t.Run(fmt.Sprint(args), func(t *testing.T) {
			cmd := NewServiceCmd()
			cmd.SetArgs(args)

			err := cmd.Execute()
			require.NoError(t, err)
		})
	}
}
