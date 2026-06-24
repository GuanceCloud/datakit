// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package httpapi

import (
	"encoding/json"
	"net/url"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
	ws "gitlab.jiagouyun.com/cloudcare-tools/datakit/dca/websocket"
	dk "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/datakit"
)

func setupDCAActionTestDirs(t *testing.T) (string, string, string, *ws.DataKit) {
	t.Helper()

	root := t.TempDir()
	confdDir := filepath.Join(root, "conf.d")
	pipelineDir := filepath.Join(root, "pipeline")
	dataDir := filepath.Join(root, "data")
	require.NoError(t, os.MkdirAll(confdDir, dk.ConfPerm))
	require.NoError(t, os.MkdirAll(pipelineDir, dk.ConfPerm))
	require.NoError(t, os.MkdirAll(dataDir, dk.ConfPerm))

	oldConfdDir := dk.ConfdDir
	oldMainConfPath := dk.MainConfPath
	oldPipelineDir := dk.PipelineDir
	t.Cleanup(func() {
		dk.ConfdDir = oldConfdDir
		dk.MainConfPath = oldMainConfPath
		dk.PipelineDir = oldPipelineDir
	})

	dk.ConfdDir = confdDir
	dk.MainConfPath = filepath.Join(confdDir, dk.StrDefaultConfFile)
	dk.PipelineDir = pipelineDir

	datakit := &ws.DataKit{
		ConnID:        "test-conn-id",
		WorkspaceUUID: "test-workspace",
		DataKitRuntimeInfo: ws.DataKitRuntimeInfo{
			ConfdDir:       confdDir,
			PipelineDir:    pipelineDir,
			DataDir:        dataDir,
			GlobalHostTags: map[string]string{"host": "dca-test-host"},
		},
	}

	return confdDir, pipelineDir, dataDir, datakit
}

func setupDCAActionTestDatakit(t *testing.T) *ws.DataKit {
	t.Helper()
	confdDir, pipelineDir, dataDir, datakit := setupDCAActionTestDirs(t)
	require.NotEmpty(t, confdDir)
	require.NotEmpty(t, pipelineDir)
	require.NotEmpty(t, dataDir)
	return datakit
}

func requireActionSuccess(t *testing.T, response *ws.DCAResponse) {
	t.Helper()
	require.Truef(t, response.Success, "response should be successful: %#v", response)
	require.Equal(t, 200, response.Code)
}

func TestGetDatakitStatsAction(t *testing.T) {
	t.Skip("getDatakitStatsAction depends on initialized DataKit runtime metrics; keep the action test placeholder until stats can be injected")

	response := &ws.DCAResponse{}

	getDatakitStatsAction(nil, response, &ws.ActionData{}, &ws.DataKit{})

	requireActionSuccess(t, response)
	require.NotNil(t, response.Content)
}

func TestGetDatakitConfigAction(t *testing.T) {
	confdDir, _, _, datakit := setupDCAActionTestDirs(t)
	configPath := filepath.Join(confdDir, "cpu.conf")
	require.NoError(t, os.WriteFile(configPath, []byte("interval = '10s'\n"), dk.ConfPerm))

	response := &ws.DCAResponse{}
	getDatakitConfigAction(nil, response, &ws.ActionData{
		Query: url.Values{"path": []string{configPath}},
	}, datakit)

	requireActionSuccess(t, response)
	require.Equal(t, "interval = '10s'\n", response.Content)
}

func TestGetDatakitConfigActionDoesNotCreateMissingParentDir(t *testing.T) {
	confdDir, _, _, datakit := setupDCAActionTestDirs(t)
	missingDir := filepath.Join(confdDir, "missing")
	configPath := filepath.Join(missingDir, "cpu.conf")

	response := &ws.DCAResponse{}
	getDatakitConfigAction(nil, response, &ws.ActionData{
		Query: url.Values{"path": []string{configPath}},
	}, datakit)

	require.False(t, response.Success)
	require.NoDirExists(t, missingDir)
}

func TestSaveDatakitConfigAction(t *testing.T) {
	confdDir, _, _, datakit := setupDCAActionTestDirs(t)

	t.Run("valid config saved", func(t *testing.T) {
		configPath := filepath.Join(confdDir, "disk.conf")
		body, err := json.Marshal(saveConfigParam{
			Path:   configPath,
			Config: "interval = '10s'\n",
			IsNew:  true,
		})
		require.NoError(t, err)

		response := &ws.DCAResponse{}
		saveDatakitConfigAction(nil, response, &ws.ActionData{Body: string(body)}, datakit)

		requireActionSuccess(t, response)
		content, err := os.ReadFile(configPath)
		require.NoError(t, err)
		require.Equal(t, "interval = '10s'\n", string(content))
	})

	t.Run("invalid toml rejected", func(t *testing.T) {
		configPath := filepath.Join(confdDir, "invalid.conf")
		body, err := json.Marshal(saveConfigParam{
			Path:   configPath,
			Config: "interval = \n",
			IsNew:  true,
		})
		require.NoError(t, err)

		response := &ws.DCAResponse{}
		saveDatakitConfigAction(nil, response, &ws.ActionData{Body: string(body)}, datakit)

		require.False(t, response.Success)
		require.Equal(t, "toml.format.error", response.ErrorCode)
		require.NoFileExists(t, configPath)
	})

	t.Run("invalid toml does not create parent dir", func(t *testing.T) {
		configDir := filepath.Join(confdDir, "invalid-nested")
		configPath := filepath.Join(configDir, "invalid.conf")
		body, err := json.Marshal(saveConfigParam{
			Path:   configPath,
			Config: "interval = \n",
			IsNew:  true,
		})
		require.NoError(t, err)

		response := &ws.DCAResponse{}
		saveDatakitConfigAction(nil, response, &ws.ActionData{Body: string(body)}, datakit)

		require.False(t, response.Success)
		require.Equal(t, "toml.format.error", response.ErrorCode)
		require.NoDirExists(t, configDir)
	})

	t.Run("existing new config rejected without force", func(t *testing.T) {
		configPath := filepath.Join(confdDir, "existing.conf")
		require.NoError(t, os.WriteFile(configPath, []byte("interval = '10s'\n"), dk.ConfPerm))
		body, err := json.Marshal(saveConfigParam{
			Path:   configPath,
			Config: "interval = '20s'\n",
			IsNew:  true,
		})
		require.NoError(t, err)

		response := &ws.DCAResponse{}
		saveDatakitConfigAction(nil, response, &ws.ActionData{Body: string(body)}, datakit)

		require.False(t, response.Success)
		require.Equal(t, "file.path.exists", response.ErrorCode)
		content, err := os.ReadFile(configPath)
		require.NoError(t, err)
		require.Equal(t, "interval = '10s'\n", string(content))
	})
}

func TestSaveDatakitConfigActionRejectsPathOutsideConfdDir(t *testing.T) {
	confdDir, _, _, datakit := setupDCAActionTestDirs(t)
	escapedPath := confdDir + string(os.PathSeparator) + ".." + string(os.PathSeparator) + "escaped.conf"
	body, err := json.Marshal(saveConfigParam{
		Path:   escapedPath,
		Config: "title = 'escaped'\n",
		IsNew:  true,
	})
	require.NoError(t, err)

	response := &ws.DCAResponse{}
	saveDatakitConfigAction(nil, response, &ws.ActionData{Body: string(body)}, datakit)

	require.False(t, response.Success)
	require.Equal(t, "params.invalid.path_invalid", response.ErrorCode)
	require.NoFileExists(t, filepath.Join(filepath.Dir(confdDir), "escaped.conf"))
}

func TestDeleteDatakitConfigAction(t *testing.T) {
	confdDir, _, _, datakit := setupDCAActionTestDirs(t)

	t.Run("existing config deleted", func(t *testing.T) {
		configPath := filepath.Join(confdDir, "delete-me.conf")
		require.NoError(t, os.WriteFile(configPath, []byte("interval = '10s'\n"), dk.ConfPerm))
		body, err := json.Marshal(map[string]string{"path": configPath, "inputName": "cpu"})
		require.NoError(t, err)

		response := &ws.DCAResponse{}
		deleteDatakitConfigAction(nil, response, &ws.ActionData{Body: string(body)}, datakit)

		requireActionSuccess(t, response)
		require.NoFileExists(t, configPath)
	})

	t.Run("main config delete rejected", func(t *testing.T) {
		require.NoError(t, os.WriteFile(dk.MainConfPath, []byte("default_enabled_inputs = []\n"), dk.ConfPerm))
		body, err := json.Marshal(map[string]string{"path": dk.MainConfPath})
		require.NoError(t, err)

		response := &ws.DCAResponse{}
		deleteDatakitConfigAction(nil, response, &ws.ActionData{Body: string(body)}, datakit)

		require.False(t, response.Success)
		require.Equal(t, "file.path.invalid", response.ErrorCode)
		require.FileExists(t, dk.MainConfPath)
	})

	t.Run("main config delete rejected after path normalization", func(t *testing.T) {
		require.NoError(t, os.WriteFile(dk.MainConfPath, []byte("default_enabled_inputs = []\n"), dk.ConfPerm))
		normalizedMainPath := confdDir + string(os.PathSeparator) + "inputs" + string(os.PathSeparator) +
			".." + string(os.PathSeparator) + dk.StrDefaultConfFile
		body, err := json.Marshal(map[string]string{"path": normalizedMainPath})
		require.NoError(t, err)

		response := &ws.DCAResponse{}
		deleteDatakitConfigAction(nil, response, &ws.ActionData{Body: string(body)}, datakit)

		require.False(t, response.Success)
		require.Equal(t, "file.path.invalid", response.ErrorCode)
		require.FileExists(t, dk.MainConfPath)
	})

	t.Run("missing config delete does not create parent dir", func(t *testing.T) {
		missingDir := filepath.Join(confdDir, "missing-delete")
		configPath := filepath.Join(missingDir, "delete-me.conf")
		body, err := json.Marshal(map[string]string{"path": configPath, "inputName": "cpu"})
		require.NoError(t, err)

		response := &ws.DCAResponse{}
		deleteDatakitConfigAction(nil, response, &ws.ActionData{Body: string(body)}, datakit)

		require.False(t, response.Success)
		require.Equal(t, "file.path.invalid", response.ErrorCode)
		require.NoDirExists(t, missingDir)
	})
}

func TestGetDatakitPipelineAction(t *testing.T) {
	_, pipelineDir, _, datakit := setupDCAActionTestDirs(t)
	require.NoError(t, os.WriteFile(filepath.Join(pipelineDir, "root.p"), []byte("drop()\n"), dk.ConfPerm))
	require.NoError(t, os.WriteFile(filepath.Join(pipelineDir, "ignore.txt"), []byte("ignored\n"), dk.ConfPerm))
	require.NoError(t, os.MkdirAll(filepath.Join(pipelineDir, "logging"), dk.ConfPerm))
	require.NoError(t, os.WriteFile(filepath.Join(pipelineDir, "logging", "log.p"), []byte("drop()\n"), dk.ConfPerm))

	response := &ws.DCAResponse{}
	getDatakitPipelineAction(nil, response, &ws.ActionData{}, datakit)

	requireActionSuccess(t, response)
	pipelines, ok := response.Content.([]pipelineInfo)
	require.True(t, ok)
	require.Len(t, pipelines, 3)
}

func TestGetDatakitPipelineActionSkipsUnreadableCategory(t *testing.T) {
	_, pipelineDir, _, datakit := setupDCAActionTestDirs(t)
	require.NoError(t, os.WriteFile(filepath.Join(pipelineDir, "root.p"), []byte("drop()\n"), dk.ConfPerm))

	unreadableDir := filepath.Join(pipelineDir, "unreadable")
	require.NoError(t, os.MkdirAll(unreadableDir, dk.ConfPerm))

	oldReadPipelineDir := readPipelineDir
	readPipelineDir = func(name string) ([]os.DirEntry, error) {
		if name == unreadableDir {
			return nil, os.ErrPermission
		}
		return oldReadPipelineDir(name)
	}
	t.Cleanup(func() { readPipelineDir = oldReadPipelineDir })

	response := &ws.DCAResponse{}
	getDatakitPipelineAction(nil, response, &ws.ActionData{}, datakit)

	requireActionSuccess(t, response)
	pipelines, ok := response.Content.([]pipelineInfo)
	require.True(t, ok)
	require.Len(t, pipelines, 1)
	require.Equal(t, "root.p", pipelines[0].FileName)
}

func TestGetDatakitPipelineDetailAction(t *testing.T) {
	_, pipelineDir, _, datakit := setupDCAActionTestDirs(t)
	require.NoError(t, os.WriteFile(filepath.Join(pipelineDir, "demo.p"), []byte("drop()\n"), dk.ConfPerm))

	t.Run("valid pipeline detail returned", func(t *testing.T) {
		response := &ws.DCAResponse{}
		getDatakitPipelineDetailAction(nil, response, &ws.ActionData{
			Query: url.Values{"fileName": []string{"demo.p"}},
		}, datakit)

		requireActionSuccess(t, response)
		detail, ok := response.Content.(pipelineDetailResponse)
		require.True(t, ok)
		require.Equal(t, filepath.Join(pipelineDir, "demo.p"), detail.Path)
		require.Equal(t, "drop()\n", detail.Content)
	})

	t.Run("invalid pipeline name rejected", func(t *testing.T) {
		response := &ws.DCAResponse{}
		getDatakitPipelineDetailAction(nil, response, &ws.ActionData{
			Query: url.Values{"fileName": []string{"demo.txt"}},
		}, datakit)

		require.False(t, response.Success)
		require.Equal(t, 400, response.Code)
		require.Equal(t, "param.invalid", response.ErrorCode)
	})

	t.Run("path traversal pipeline name rejected", func(t *testing.T) {
		outsidePath := filepath.Join(filepath.Dir(pipelineDir), "outside-detail.p")
		require.NoError(t, os.WriteFile(outsidePath, []byte("drop()\n"), dk.ConfPerm))

		response := &ws.DCAResponse{}
		getDatakitPipelineDetailAction(nil, response, &ws.ActionData{
			Query: url.Values{"fileName": []string{"../outside-detail.p"}},
		}, datakit)

		require.False(t, response.Success)
		require.Equal(t, 400, response.Code)
		require.Equal(t, "param.invalid", response.ErrorCode)
	})
}

func TestPatchDatakitPipelineAction(t *testing.T) {
	_, pipelineDir, _, datakit := setupDCAActionTestDirs(t)
	require.NoError(t, os.WriteFile(filepath.Join(pipelineDir, "demo.p"), []byte("drop()\n"), dk.ConfPerm))

	t.Run("existing pipeline updated", func(t *testing.T) {
		body, err := json.Marshal(pipelineInfo{
			FileName: "demo.p",
			Content:  "add_key(status, 'ok')\n",
		})
		require.NoError(t, err)

		response := &ws.DCAResponse{}
		saveDatakitPipelineAction(true)(nil, response, &ws.ActionData{Body: string(body)}, datakit)

		requireActionSuccess(t, response)
		content, err := os.ReadFile(filepath.Join(pipelineDir, "demo.p"))
		require.NoError(t, err)
		require.Equal(t, "add_key(status, 'ok')\n", string(content))
	})

	t.Run("path traversal pipeline name rejected", func(t *testing.T) {
		outsidePath := filepath.Join(filepath.Dir(pipelineDir), "outside-patch.p")
		require.NoError(t, os.WriteFile(outsidePath, []byte("drop()\n"), dk.ConfPerm))
		body, err := json.Marshal(pipelineInfo{
			FileName: "../outside-patch.p",
			Content:  "add_key(status, 'changed')\n",
		})
		require.NoError(t, err)

		response := &ws.DCAResponse{}
		saveDatakitPipelineAction(true)(nil, response, &ws.ActionData{Body: string(body)}, datakit)

		require.False(t, response.Success)
		require.Equal(t, 400, response.Code)
		require.Equal(t, "param.invalid", response.ErrorCode)
		content, err := os.ReadFile(outsidePath)
		require.NoError(t, err)
		require.Equal(t, "drop()\n", string(content))
	})
}

func TestCreateDatakitPipelineAction(t *testing.T) {
	_, pipelineDir, _, datakit := setupDCAActionTestDirs(t)

	t.Run("new pipeline created", func(t *testing.T) {
		body, err := json.Marshal(pipelineInfo{
			FileName: "created.p",
			Content:  "drop()\n",
		})
		require.NoError(t, err)

		response := &ws.DCAResponse{}
		saveDatakitPipelineAction(false)(nil, response, &ws.ActionData{Body: string(body)}, datakit)

		requireActionSuccess(t, response)
		require.FileExists(t, filepath.Join(pipelineDir, "created.p"))
	})

	t.Run("duplicate pipeline rejected", func(t *testing.T) {
		body, err := json.Marshal(pipelineInfo{
			FileName: "created.p",
			Content:  "drop()\n",
		})
		require.NoError(t, err)

		response := &ws.DCAResponse{}
		saveDatakitPipelineAction(false)(nil, response, &ws.ActionData{Body: string(body)}, datakit)

		require.False(t, response.Success)
		require.Equal(t, 400, response.Code)
		require.Equal(t, "param.invalid.duplicate", response.ErrorCode)
	})

	t.Run("path traversal pipeline name rejected", func(t *testing.T) {
		outsidePath := filepath.Join(filepath.Dir(pipelineDir), "outside-create.p")
		body, err := json.Marshal(pipelineInfo{
			FileName: "../outside-create.p",
			Content:  "drop()\n",
		})
		require.NoError(t, err)

		response := &ws.DCAResponse{}
		saveDatakitPipelineAction(false)(nil, response, &ws.ActionData{Body: string(body)}, datakit)

		require.False(t, response.Success)
		require.Equal(t, 400, response.Code)
		require.Equal(t, "param.invalid", response.ErrorCode)
		require.NoFileExists(t, outsidePath)
	})

	t.Run("codex outside pipeline create rejected", func(t *testing.T) {
		outsidePath := filepath.Join(filepath.Dir(pipelineDir), "codex-outside-create.p")
		body, err := json.Marshal(pipelineInfo{
			FileName: "../codex-outside-create.p",
			Content:  "drop()\n",
		})
		require.NoError(t, err)

		response := &ws.DCAResponse{}
		saveDatakitPipelineAction(false)(nil, response, &ws.ActionData{Body: string(body)}, datakit)

		require.False(t, response.Success)
		require.Equal(t, 400, response.Code)
		require.Equal(t, "param.invalid", response.ErrorCode)
		require.NoFileExists(t, outsidePath)
	})
}

func TestDeleteDatakitPipelineAction(t *testing.T) {
	_, pipelineDir, _, datakit := setupDCAActionTestDirs(t)

	t.Run("existing pipeline deleted", func(t *testing.T) {
		pipelinePath := filepath.Join(pipelineDir, "delete-me.p")
		require.NoError(t, os.WriteFile(pipelinePath, []byte("drop()\n"), dk.ConfPerm))
		body, err := json.Marshal(map[string]string{"fileName": "delete-me.p"})
		require.NoError(t, err)

		response := &ws.DCAResponse{}
		deleteDatakitPipelineAction(nil, response, &ws.ActionData{Body: string(body)}, datakit)

		requireActionSuccess(t, response)
		require.NoFileExists(t, pipelinePath)
	})

	t.Run("path traversal pipeline name rejected", func(t *testing.T) {
		outsidePath := filepath.Join(filepath.Dir(pipelineDir), "outside-delete.p")
		require.NoError(t, os.WriteFile(outsidePath, []byte("drop()\n"), dk.ConfPerm))
		body, err := json.Marshal(map[string]string{"fileName": "../outside-delete.p"})
		require.NoError(t, err)

		response := &ws.DCAResponse{}
		deleteDatakitPipelineAction(nil, response, &ws.ActionData{Body: string(body)}, datakit)

		require.False(t, response.Success)
		require.Equal(t, "param.invalid", response.ErrorCode)
		require.FileExists(t, outsidePath)
	})
}

func TestTestDatakitPipelineAction(t *testing.T) {
	datakit := setupDCAActionTestDatakit(t)

	t.Run("invalid pipeline script rejected", func(t *testing.T) {
		body, err := json.Marshal(dcaTestParam{
			Category:   "logging",
			ScriptName: "demo",
			Pipeline: map[string]map[string]string{
				"logging": {"demo": "this is not valid pipeline script"},
			},
			Data: []string{"hello"},
		})
		require.NoError(t, err)

		response := &ws.DCAResponse{}
		testDatakitPipelineAction(nil, response, &ws.ActionData{Body: string(body)}, datakit)

		require.False(t, response.Success)
		require.Contains(t, response.Message, "pipeline parse error")
	})

	t.Run("invalid request body rejected", func(t *testing.T) {
		cases := []struct {
			name string
			body string
		}{
			{
				name: "missing pipeline",
				body: `{"category":"logging","script_name":"demo","data":["hello"]}`,
			},
			{
				name: "missing script name",
				body: `{"category":"logging","pipeline":{"logging":{"demo":"drop()"}},"data":["hello"]}`,
			},
			{
				name: "missing category pipeline",
				body: `{"category":"metric","script_name":"demo","pipeline":{"logging":{"demo":"drop()"}},"data":["cpu usage=1"]}`,
			},
			{
				name: "missing script content",
				body: `{"category":"logging","script_name":"demo","pipeline":{"logging":{}},"data":["hello"]}`,
			},
			{
				name: "default category without default pipeline",
				body: `{"category":"default","script_name":"demo","pipeline":{"logging":{"demo":"drop()"}},"data":["hello"]}`,
			},
		}

		for _, tc := range cases {
			t.Run(tc.name, func(t *testing.T) {
				response := &ws.DCAResponse{}
				require.NotPanics(t, func() {
					testDatakitPipelineAction(nil, response, &ws.ActionData{Body: tc.body}, datakit)
				})

				require.False(t, response.Success)
				require.Equal(t, "param.invalid", response.ErrorCode)
			})
		}
	})
}

func TestGetDatakitFilterAction(t *testing.T) {
	_, _, dataDir, datakit := setupDCAActionTestDirs(t)

	t.Run("missing pull file returns empty content", func(t *testing.T) {
		response := &ws.DCAResponse{}
		getDatakitFilterAction(nil, response, &ws.ActionData{}, datakit)

		requireActionSuccess(t, response)
		filter, ok := response.Content.(filterInfo)
		require.True(t, ok)
		require.Empty(t, filter.Content)
		require.Empty(t, filter.FilePath)
	})

	t.Run("pull file content returned", func(t *testing.T) {
		pullPath := filepath.Join(dataDir, ".pull")
		require.NoError(t, os.WriteFile(pullPath, []byte("source = default\n"), dk.ConfPerm))

		response := &ws.DCAResponse{}
		getDatakitFilterAction(nil, response, &ws.ActionData{}, datakit)

		requireActionSuccess(t, response)
		filter, ok := response.Content.(filterInfo)
		require.True(t, ok)
		require.Equal(t, "source = default\n", filter.Content)
		require.Equal(t, pullPath, filter.FilePath)
	})
}

func TestNewWebsocketConnectionAction(t *testing.T) {
	datakit := setupDCAActionTestDatakit(t)
	client, err := ws.NewClient(
		ws.WithWebsocketAddress("ws://127.0.0.1:1/ws"),
		ws.WithDataKit(datakit),
	)
	require.NoError(t, err)
	message := ws.WebsocketMessage{
		Action: ws.NewWebsocketConnectionAction,
		Data: &ws.ActionData{
			Query: url.Values{ws.HeaderNewWebSocketConnectionID: []string{"test-connection"}},
		},
	}

	require.NoError(t, newWebsocketConnectionAction(client, 1, message.Bytes()))
}
