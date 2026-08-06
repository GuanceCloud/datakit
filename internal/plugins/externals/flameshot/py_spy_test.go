// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package flameshot

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBuildPySpyArgs(t *testing.T) {
	stats := &triggerStats{
		PID:          1234,
		Duration:     17,
		PySpyRate:    200,
		PySpySubproc: true,
		PySpyIdle:    true,
	}

	assert.Equal(t,
		[]string{
			"record",
			"--format", "raw",
			"--output", "/tmp/prof",
			"--duration", "17",
			"--pid", "1234",
			"--rate", "200",
			"--subprocesses",
			"--idle",
		},
		buildPySpyArgs(stats, "/tmp/prof"),
	)
}

func TestRunPythonPySpyGeneratesCollapseAttachment(t *testing.T) {
	outputPath := filepath.Join(t.TempDir(), "prof")
	require.NoError(t, os.WriteFile(outputPath, []byte("stale profile"), 0o644))
	restore := stubRunPySpyCommand(t, func(_ context.Context, gotPath string, gotArgs []string) (string, string, error) {
		assert.Equal(t, "/usr/bin/py-spy", gotPath)
		assert.Contains(t, gotArgs, "--format")
		assert.Contains(t, gotArgs, "raw")
		assert.Contains(t, gotArgs, "--output")
		assert.Contains(t, gotArgs, outputPath)
		_, err := os.Stat(outputPath)
		assert.True(t, os.IsNotExist(err))
		require.NoError(t, os.WriteFile(outputPath, []byte("process 1:\"python app.py\";thread (1);main (/app/app.py:1) 1\n"), 0o644))
		return "ok", "", nil
	})
	defer restore()

	stats := &triggerStats{
		Language:    "python",
		PID:         1234,
		Duration:    3,
		PySpyPath:   "/usr/bin/py-spy",
		PySpyOutput: outputPath,
		PySpyRate:   100,
	}

	require.NoError(t, runPythonPySpy(t.Context(), stats))
	assert.Equal(t, outputPath, stats.OutputFile)
	assert.Equal(t, pySpyUploadEventFormat, stats.OutputFormat)
	require.Len(t, stats.Attachments, 1)
	assert.Equal(t, "prof", stats.Attachments[0].FieldName)
	assert.Equal(t, "prof", stats.Attachments[0].FileName)
	assert.Contains(t, string(stats.Attachments[0].Data), "python app.py")
	assert.NotEmpty(t, stats.startTime)
	assert.NotEmpty(t, stats.endTime)
	_, err := os.Stat(outputPath)
	assert.NoError(t, err)
}

func TestRunPythonPySpyCleansOutputOnFailure(t *testing.T) {
	tests := []struct {
		name    string
		prepare func(string) error
		command error
		wantErr string
	}{
		{
			name: "command writes partial output then fails",
			prepare: func(outputPath string) error {
				return os.WriteFile(outputPath, []byte("partial profile"), 0o644)
			},
			command: errors.New("record failed"),
			wantErr: "py-spy exec err",
		},
		{
			name: "output is empty",
			prepare: func(outputPath string) error {
				return os.WriteFile(outputPath, nil, 0o644)
			},
			wantErr: "py-spy output file is empty",
		},
		{
			name: "output cannot be read",
			prepare: func(outputPath string) error {
				return os.Mkdir(outputPath, 0o755)
			},
			wantErr: "read py-spy output file",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			outputPath := filepath.Join(t.TempDir(), "prof")
			restore := stubRunPySpyCommand(t, func(_ context.Context, _ string, _ []string) (string, string, error) {
				require.NoError(t, tt.prepare(outputPath))
				return "", "", tt.command
			})
			defer restore()

			err := runPythonPySpy(t.Context(), &triggerStats{
				PID:         1234,
				Duration:    1,
				PySpyOutput: outputPath,
			})

			require.Error(t, err)
			assert.Contains(t, err.Error(), tt.wantErr)
			_, statErr := os.Stat(outputPath)
			assert.True(t, os.IsNotExist(statErr), "output should be removed, stat error: %v", statErr)
		})
	}
}

func TestResolvePySpyOutputPathUsesUniqueDefault(t *testing.T) {
	oldDefaultOutput := DefaultOutput
	DefaultOutput = t.TempDir()
	t.Cleanup(func() { DefaultOutput = oldDefaultOutput })

	outputPath := resolvePySpyOutputPath(&triggerStats{PID: 4321})

	assert.Equal(t, DefaultOutput, filepath.Dir(outputPath))
	assert.Regexp(t, `^pyspy_4321_\d{8}_\d{6}_\d{9}\.raw$`, filepath.Base(outputPath))
}

func TestRunProfilingDispatchesPythonPySpy(t *testing.T) {
	outputPath := filepath.Join(t.TempDir(), "prof")
	restore := stubRunPySpyCommand(t, func(_ context.Context, _ string, _ []string) (string, string, error) {
		require.NoError(t, os.WriteFile(outputPath, []byte("process 1:\"python app.py\";thread (1);main (/app/app.py:1) 1\n"), 0o644))
		return "", "", nil
	})
	defer restore()

	stats := &triggerStats{
		Language:    "python",
		PID:         1234,
		Duration:    1,
		PySpyOutput: outputPath,
	}

	require.NoError(t, runProfiling(t.Context(), stats))
	assert.Equal(t, []string{"prof"}, stats.attachmentFileNames())
}

func stubRunPySpyCommand(
	t *testing.T,
	fn func(context.Context, string, []string) (string, string, error),
) func() {
	t.Helper()

	old := runPySpyCommand
	runPySpyCommand = fn
	return func() {
		runPySpyCommand = old
	}
}
