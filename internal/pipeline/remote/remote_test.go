// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package remote

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/GuanceCloud/cliutils/point"
	"github.com/stretchr/testify/assert"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/pipeline/plval"
)

//------------------------------------------------------------------------------

var errGeneral = fmt.Errorf("test_specific_error")

type fileInfoStruct struct{}

func (fileInfoStruct) Name() string {
	return "useless"
}

func (fileInfoStruct) Size() int64 {
	return 0
}

func (fileInfoStruct) Mode() fs.FileMode {
	return 0
}

func (fileInfoStruct) ModTime() time.Time {
	return time.Time{}
}

func (fileInfoStruct) IsDir() bool {
	return false
}

func (fileInfoStruct) Sys() interface{} {
	return nil
}

// Make sure pipelineRemoteMockerTest implements the IPipelineRemote interface
var _ IPipelineRemote = new(pipelineRemoteMockerTest)

type pipelineRemoteMockerTest struct {
	writeFileData          *FileDataStruct
	readFileData           []byte
	isFileExist            bool
	readDirResult          []fs.FileInfo
	pullPipelineUpdateTime int64
	pullRelationUpdate     bool
	pullRelationUpdateAt   int64
	pullFiles              map[point.Category]map[string]string
	pullRelation           map[point.Category]map[string]string
	pullDefaults           map[point.Category]string
	writeTarStarted        chan struct{}
	writeTarOnce           sync.Once
	writeTarCalls          int
	files                  map[string][]byte

	errMarshal                   error
	errUnMarshal                 error
	errReadFile                  error
	errWriteFile                 error
	errReadDir                   error
	errPullPipeline              error
	errRemove                    error
	errGetNamespacePipelineFiles error
	errReadTarToMap              error
	errWriteTarFromMap           error
}

func newPipelineRemoteMock() *pipelineRemoteMockerTest {
	mock := &pipelineRemoteMockerTest{}

	mock.writeFileData = nil
	mock.readFileData = []byte{}
	mock.isFileExist = false
	mock.readDirResult = []fs.FileInfo{}
	mock.pullPipelineUpdateTime = 0
	mock.pullRelationUpdate = false
	mock.pullPipelineUpdateTime = 0
	mock.errMarshal = nil
	mock.errUnMarshal = nil
	mock.errReadFile = nil
	mock.errWriteFile = nil
	mock.errReadDir = nil
	mock.errPullPipeline = nil
	mock.errRemove = nil
	mock.errGetNamespacePipelineFiles = nil
	mock.errReadTarToMap = nil
	mock.errWriteTarFromMap = nil

	return mock
}

type FileDataStruct struct {
	FileName string
	Bytes    []byte
}

func (mock *pipelineRemoteMockerTest) FileExist(filename string) bool {
	if mock.files != nil {
		_, ok := mock.files[filename]
		return ok
	}
	return mock.isFileExist
}

func (mock *pipelineRemoteMockerTest) Marshal(v interface{}) ([]byte, error) {
	if mock.errMarshal != nil {
		return nil, mock.errMarshal
	}

	return json.Marshal(v)
}

func (mock *pipelineRemoteMockerTest) Unmarshal(data []byte, v interface{}) error {
	if mock.errUnMarshal != nil {
		return mock.errUnMarshal
	}

	return json.Unmarshal(data, v)
}

func (mock *pipelineRemoteMockerTest) ReadFile(filename string) ([]byte, error) {
	if mock.errReadFile != nil {
		return nil, mock.errReadFile
	}

	if mock.files != nil {
		return append([]byte(nil), mock.files[filename]...), nil
	}
	return mock.readFileData, nil
}

func (mock *pipelineRemoteMockerTest) WriteFile(filename string, data []byte, perm fs.FileMode) error {
	if mock.errWriteFile != nil {
		return mock.errWriteFile
	}

	mock.writeFileData = &FileDataStruct{
		FileName: filename,
		Bytes:    data,
	}
	if mock.files != nil {
		mock.files[filename] = append([]byte(nil), data...)
	}
	return nil
}

func (mock *pipelineRemoteMockerTest) ReadDir(dirname string) ([]fs.FileInfo, error) {
	if mock.errReadDir != nil {
		return nil, mock.errReadDir
	}

	return mock.readDirResult, nil
}

func (mock *pipelineRemoteMockerTest) PullPipeline(ts, relationTS int64) (mFiles, plRelation map[point.Category]map[string]string,
	defaultPl map[point.Category]string, updateTime int64, relationUpdateAt int64, err error,
) {
	if mock.errPullPipeline != nil {
		return nil, nil, nil, 0, 0, mock.errPullPipeline
	}

	files := map[point.Category]map[string]string{
		point.Logging: {
			"123.p": "text123",
			"456.p": "text456",
		},
	}
	relation := map[point.Category]map[string]string{
		point.Logging: {
			"123": "123.p",
			"234": "123.p",
		},
	}
	defaults := map[point.Category]string{
		point.Logging: "123.p",
	}
	if mock.pullFiles != nil {
		files = mock.pullFiles
	}
	if mock.pullRelation != nil {
		relation = mock.pullRelation
	}
	if mock.pullDefaults != nil {
		defaults = mock.pullDefaults
	}
	relationUpdateAt = -1
	if mock.pullRelationUpdate {
		relationUpdateAt = mock.pullRelationUpdateAt
	}
	return files, relation, defaults, mock.pullPipelineUpdateTime, relationUpdateAt, nil
}

func (*pipelineRemoteMockerTest) GetTickerDurationAndBreak() time.Duration {
	return time.Second
}

func (mock *pipelineRemoteMockerTest) Remove(name string) error {
	if mock.errRemove == nil && mock.files != nil {
		delete(mock.files, name)
	}
	return mock.errRemove
}

func (*pipelineRemoteMockerTest) FeedLastError(inputName string, err string) {}

func (mock *pipelineRemoteMockerTest) GetNamespacePipelineFiles(namespace string) ([]string, error) {
	return nil, mock.errGetNamespacePipelineFiles
}

func (mock *pipelineRemoteMockerTest) ReadTarToMap(srcFile string) (map[string]string, error) {
	return nil, mock.errReadTarToMap
}

func (mock *pipelineRemoteMockerTest) WriteTarFromMap(data map[string]string, dest string) error {
	mock.writeTarCalls++
	if mock.files != nil {
		mock.files[dest] = []byte("new archive")
	}
	if mock.writeTarStarted != nil {
		mock.writeTarOnce.Do(func() { close(mock.writeTarStarted) })
	}
	return mock.errWriteTarFromMap
}

// go test -v -timeout 30s -run ^TestDoPull$ gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/pipeline/remote
func TestDoPull(t *testing.T) {
	err := plval.InitPlVal(nil, nil, nil, "")
	if err != nil {
		t.Fatal(err)
	}

	const dwURL = "https://openway.guance.com?token=tkn_123"
	const configPath = "/usr/local/datakit/pipeline_remote/.config_fake"
	const relationPath = "/usr/local/datakit/pipeline_remote/.relation_fake_dump.json"

	cases := []struct {
		name                       string
		fileExist                  bool
		pathConfig                 string
		siteURL                    string
		configContent              []byte
		testPullPipelineUpdateTime int64
		testPullRelationUpdate     bool
		testPullRelationUpdateAt   int64
		testReadDirResult          []fs.FileInfo
		failedMarshal              error
		failedReadFile             error
		failedReadDir              error
		failedPullPipeline         error
		failedRemove               error
		expectError                error
	}{
		{
			name:          "update",
			pathConfig:    configPath,
			siteURL:       dwURL,
			configContent: []byte(`{"SiteURL":"https://openway.guance.com?token=tkn_123","UpdateTime":1644318398}`),
		},
		{
			name:           "getPipelineRemoteConfig_fail",
			fileExist:      true,
			failedReadFile: errGeneral,
			expectError:    errGeneral,
		},
		{
			name:               "PullPipeline_fail",
			failedPullPipeline: errGeneral,
			expectError:        errGeneral,
		},
		{
			name:          "alread_up_to_date",
			fileExist:     true,
			pathConfig:    configPath,
			siteURL:       dwURL,
			configContent: []byte(`{"SiteURL":"https://openway.guance.com?token=tkn_123","UpdateTime":1644318398}`),
		},
		{
			name:                       "dumpfile_fail",
			pathConfig:                 configPath,
			siteURL:                    dwURL,
			configContent:              []byte(`{"SiteURL":"https://openway.guance.com?token=tkn_123","UpdateTime":1644318398}`),
			testPullPipelineUpdateTime: 123,
			failedReadDir:              errGeneral,
			expectError:                errGeneral,
		},
		{
			name:                       "updatePipelineRemoteConfig_fail",
			pathConfig:                 configPath,
			siteURL:                    dwURL,
			configContent:              []byte(`{"SiteURL":"https://openway.guance.com?token=tkn_123","UpdateTime":1644318398}`),
			testPullPipelineUpdateTime: 123,
			failedMarshal:              errGeneral,
			expectError:                errGeneral,
		},
		{
			name:                       "updatePipelineRemoteConfig_pass",
			pathConfig:                 configPath,
			siteURL:                    dwURL,
			configContent:              []byte(`{"SiteURL":"https://openway.guance.com?token=tkn_123","UpdateTime":1644318398}`),
			testPullPipelineUpdateTime: 123,
		},
		{
			name:                       "deleteAll_nil",
			pathConfig:                 configPath,
			siteURL:                    dwURL,
			configContent:              []byte(`{"SiteURL":"https://openway.guance.com?token=tkn_123","UpdateTime":1644318398}`),
			testPullPipelineUpdateTime: 1,
		},
		{
			name:                       "deleteAll_error",
			pathConfig:                 configPath,
			siteURL:                    dwURL,
			configContent:              []byte(`{"SiteURL":"https://openway.guance.com?token=tkn_123","UpdateTime":1644318398}`),
			testPullPipelineUpdateTime: 1,
			failedReadDir:              errGeneral,
			expectError:                errGeneral,
		},
		{
			name:                       "removeLocalRemote_continue",
			pathConfig:                 configPath,
			siteURL:                    dwURL,
			configContent:              []byte(`{"SiteURL":"https://openway.guance.com?token=tkn_123","UpdateTime":1644318398}`),
			testPullPipelineUpdateTime: 1,
			testReadDirResult:          []fs.FileInfo{&fileInfoStruct{}},
			failedRemove:               errGeneral,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Logf("TestDoPull: tc.name = %s\n", tc.name)

			mock := newPipelineRemoteMock()
			mock.readFileData = tc.configContent
			mock.isFileExist = tc.fileExist
			mock.errMarshal = tc.failedMarshal
			mock.errReadFile = tc.failedReadFile
			mock.errReadDir = tc.failedReadDir
			mock.errPullPipeline = tc.failedPullPipeline
			mock.pullPipelineUpdateTime = tc.testPullPipelineUpdateTime
			mock.pullRelationUpdate = tc.testPullRelationUpdate
			mock.pullRelationUpdateAt = tc.testPullRelationUpdateAt
			if len(tc.testReadDirResult) > 0 {
				mock.readDirResult = tc.testReadDirResult
			}
			mock.errRemove = tc.failedRemove

			err := doPull(tc.pathConfig, relationPath, tc.siteURL, mock)
			if tc.expectError == nil {
				assert.NoError(t, err)
			} else {
				assert.ErrorIs(t, err, tc.expectError, "doPull found error: %v", err)
			}
		})
	}
}

func TestDoPullRejectsInvalidGenerationWithoutPublishingOrPersisting(t *testing.T) {
	manager := plval.NewScriptManager(nil, nil)
	initial, err := manager.PrepareRemoteUpdate(plval.RemoteManagerUpdate{
		Scripts: map[point.Category]map[string]string{
			point.Logging: {"old.p": "drop_key(old)\n"},
		},
		ReplaceScripts:  true,
		Defaults:        map[point.Category]string{point.Logging: "old.p"},
		ReplaceDefaults: true,
		Relation: map[point.Category]map[string]string{
			point.Logging: {"source": "old.p"},
		},
		RelationUpdateAt: 5,
		ReplaceRelation:  true,
	})
	if err != nil {
		t.Fatalf("prepare initial generation: %v", err)
	}
	initial.Commit()
	plval.SetManager(manager)

	mock := newPipelineRemoteMock()
	mock.pullPipelineUpdateTime = 42
	mock.pullRelationUpdate = true
	mock.pullRelationUpdateAt = 9
	mock.pullFiles = map[point.Category]map[string]string{
		point.Logging: {"broken.p": "if"},
	}
	mock.pullDefaults = map[point.Category]string{point.Logging: "broken.p"}
	mock.pullRelation = map[point.Category]map[string]string{
		point.Logging: {"source": "broken.p"},
	}
	oldContentPath := pathContent
	pathContent = filepath.Join(t.TempDir(), pipelineRemoteContentFile)
	defer func() { pathContent = oldContentPath }()

	err = doPull(filepath.Join(t.TempDir(), pipelineRemoteConfigFile),
		filepath.Join(t.TempDir(), pipelineRemoteRelationDumpFile), "site", mock)
	if err == nil {
		t.Fatal("invalid remote generation was accepted")
	}
	if mock.writeTarCalls != 0 || mock.writeFileData != nil {
		t.Fatalf("invalid generation changed disk: tar_calls=%d write=%#v", mock.writeTarCalls, mock.writeFileData)
	}
	if got := manager.RelationUpdateAt(); got != 5 {
		t.Fatalf("relation update time advanced to %d", got)
	}
	lease, ok := manager.Acquire()
	if !ok {
		t.Fatal("acquire retained generation")
	}
	defer lease.Release()
	if name, ok := lease.Relation().Query(point.Logging, "source"); !ok || name != "old.p" {
		t.Fatalf("relation changed to %q", name)
	}
	if script, ok := lease.Manager().QueryScript(point.Logging, "missing"); !ok || script.Name() != "old.p" {
		t.Fatalf("default changed after rejected update: %#v", script)
	}
}

func TestDoPullPublishesScriptsDefaultsAndRelationAfterOldBatchDrains(t *testing.T) {
	manager := plval.NewScriptManager(nil, nil)
	initial, err := manager.PrepareRemoteUpdate(plval.RemoteManagerUpdate{
		Scripts: map[point.Category]map[string]string{
			point.Logging: {"old.p": "drop_key(old)\n"},
		},
		ReplaceScripts:  true,
		Defaults:        map[point.Category]string{point.Logging: "old.p"},
		ReplaceDefaults: true,
		Relation: map[point.Category]map[string]string{
			point.Logging: {"source": "old.p"},
		},
		RelationUpdateAt: 5,
		ReplaceRelation:  true,
	})
	if err != nil {
		t.Fatalf("prepare initial generation: %v", err)
	}
	initial.Commit()
	plval.SetManager(manager)

	oldLease, ok := manager.Acquire()
	if !ok {
		t.Fatal("acquire old batch")
	}
	mock := newPipelineRemoteMock()
	mock.pullPipelineUpdateTime = 42
	mock.pullRelationUpdate = true
	mock.pullRelationUpdateAt = 9
	mock.pullFiles = map[point.Category]map[string]string{
		point.Logging: {"new.p": "drop_key(new)\n"},
	}
	mock.pullDefaults = map[point.Category]string{point.Logging: "new.p"}
	mock.pullRelation = map[point.Category]map[string]string{
		point.Logging: {"source": "new.p"},
	}
	mock.writeTarStarted = make(chan struct{})
	oldContentPath := pathContent
	pathContent = filepath.Join(t.TempDir(), pipelineRemoteContentFile)
	defer func() { pathContent = oldContentPath }()
	configPath := filepath.Join(t.TempDir(), pipelineRemoteConfigFile)
	relationPath := filepath.Join(t.TempDir(), pipelineRemoteRelationDumpFile)

	result := make(chan error, 1)
	go func() {
		result <- doPull(configPath, relationPath, "site", mock)
	}()
	select {
	case <-mock.writeTarStarted:
	case <-time.After(5 * time.Second):
		t.Fatal("remote candidate was not persisted")
	}
	select {
	case err := <-result:
		t.Fatalf("publication crossed old batch lease: %v", err)
	case <-time.After(100 * time.Millisecond):
	}
	if script, ok := oldLease.Manager().QueryScript(point.Logging, "missing"); !ok || script.Name() != "old.p" {
		t.Fatalf("old batch default changed: %#v", script)
	}
	if name, ok := oldLease.Relation().Query(point.Logging, "source"); !ok || name != "old.p" {
		t.Fatalf("old batch relation changed to %q", name)
	}
	oldLease.Release()
	select {
	case err := <-result:
		if err != nil {
			t.Fatalf("publish generation: %v", err)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("publication did not finish after old batch drained")
	}

	lease, ok := manager.Acquire()
	if !ok {
		t.Fatal("acquire new batch")
	}
	defer lease.Release()
	if script, ok := lease.Manager().QueryScript(point.Logging, "missing"); !ok || script.Name() != "new.p" {
		t.Fatalf("new default was not published: %#v", script)
	}
	if name, ok := lease.Relation().Query(point.Logging, "source"); !ok || name != "new.p" {
		t.Fatalf("new relation = %q", name)
	}
}

func TestDoPullPersistenceFailureRollsBackFilesAndGeneration(t *testing.T) {
	manager := plval.NewScriptManager(nil, nil)
	initial, err := manager.PrepareRemoteUpdate(plval.RemoteManagerUpdate{
		Scripts: map[point.Category]map[string]string{
			point.Logging: {"old.p": "drop_key(old)\n"},
		},
		ReplaceScripts:  true,
		Defaults:        map[point.Category]string{point.Logging: "old.p"},
		ReplaceDefaults: true,
		Relation: map[point.Category]map[string]string{
			point.Logging: {"source": "old.p"},
		},
		RelationUpdateAt: 5,
		ReplaceRelation:  true,
	})
	if err != nil {
		t.Fatalf("prepare initial generation: %v", err)
	}
	initial.Commit()
	plval.SetManager(manager)

	temp := t.TempDir()
	configPath := filepath.Join(temp, pipelineRemoteConfigFile)
	relationPath := filepath.Join(temp, pipelineRemoteRelationDumpFile)
	oldContentPath := pathContent
	pathContent = filepath.Join(temp, pipelineRemoteContentFile)
	defer func() { pathContent = oldContentPath }()
	original := map[string][]byte{
		configPath:   []byte(`{"SiteURL":"site","UpdateTime":5}`),
		pathContent:  []byte("old archive"),
		relationPath: []byte("old relation"),
	}
	mock := newPipelineRemoteMock()
	mock.files = map[string][]byte{}
	for path, data := range original {
		mock.files[path] = append([]byte(nil), data...)
	}
	mock.pullPipelineUpdateTime = 42
	mock.pullRelationUpdate = true
	mock.pullRelationUpdateAt = 9
	mock.pullFiles = map[point.Category]map[string]string{
		point.Logging: {"new.p": "drop_key(new)\n"},
	}
	mock.pullDefaults = map[point.Category]string{point.Logging: "new.p"}
	mock.pullRelation = map[point.Category]map[string]string{
		point.Logging: {"source": "new.p"},
	}
	mock.errMarshal = errGeneral

	err = doPull(configPath, relationPath, "site", mock)
	if !errors.Is(err, errGeneral) {
		t.Fatalf("persistence failure = %v", err)
	}
	for path, want := range original {
		if got := mock.files[path]; !bytes.Equal(got, want) {
			t.Fatalf("%s was not rolled back: got %q want %q", path, got, want)
		}
	}
	if got := manager.RelationUpdateAt(); got != 5 {
		t.Fatalf("failed persistence advanced relation to %d", got)
	}
	if _, ok := manager.QueryScript(point.Logging, "new.p", struct{}{}); ok {
		t.Fatal("failed persistence published the candidate script")
	}
}

// go test -v -timeout 30s -run ^TestDumpFiles$ gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/pipeline/remote
func TestDumpFiles(t *testing.T) {
	cases := []struct {
		name                  string
		files                 map[point.Category]map[string]string
		readDir               []fs.FileInfo
		failedReadDir         error
		failedWriteTarFromMap error
		expectError           error
	}{
		{
			name: "normal",
			files: map[point.Category]map[string]string{
				point.Logging: {
					"123.p": "text123",
					"456.p": "text456",
				},
			},
		},
		{
			name:          "read_dir_fail",
			failedReadDir: errGeneral,
			expectError:   errGeneral,
		},
		{
			name:                  "WriteTarFromMap_fail",
			failedWriteTarFromMap: errGeneral,
			files: map[point.Category]map[string]string{
				point.Logging: {
					"123.p": "text123",
					"456.p": "text456",
				},
			},
			expectError: errGeneral,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			mock := newPipelineRemoteMock()
			mock.errReadDir = tc.failedReadDir
			mock.errWriteTarFromMap = tc.failedWriteTarFromMap

			err := dumpFiles(tc.files, nil, mock)
			assert.Equal(t, tc.expectError, err, "dumpFiles found error: %v", err)
		})
	}
}

// go test -v -timeout 30s -run ^TestGetPipelineRemoteConfig$ gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/pipeline/remote
func TestGetPipelineRemoteConfig(t *testing.T) {
	const dwURL = "https://openway.guance.com?token=tkn_123"
	const configPath = "/usr/local/datakit/pipeline_remote/.config_fake"

	cases := []struct {
		name               string
		fileExist          bool
		pathConfig         string
		siteURL            string
		configContent      []byte
		failedUnMarshal    error
		failedReadFile     error
		failedRemove       error
		failedReadDir      error
		failedReadTarToMap error
		expectError        error
		expect             int64
	}{
		{
			name:          "normal",
			fileExist:     true,
			pathConfig:    configPath,
			siteURL:       dwURL,
			configContent: []byte(`{"SiteURL":"https://openway.guance.com?token=tkn_123","UpdateTime":1644318398}`),
			expect:        0,
		},
		{
			name:       "config_not_exist",
			pathConfig: "",
		},
		{
			name:           "read_file_fail",
			fileExist:      true,
			pathConfig:     configPath,
			failedReadFile: errGeneral,
			expectError:    errGeneral,
		},
		{
			name:            "json_unmarshal_fail",
			fileExist:       true,
			pathConfig:      configPath,
			failedUnMarshal: errGeneral,
			expectError:     errGeneral,
		},
		{
			name:          "token_changed",
			fileExist:     true,
			pathConfig:    configPath,
			siteURL:       dwURL,
			configContent: []byte(`{"SiteURL":"http://127.0.0.1:9528?token=tkn_123","UpdateTime":1644318398}`),
		},
		{
			name:               "ReadTarToMap_failed",
			fileExist:          true,
			pathConfig:         configPath,
			siteURL:            dwURL,
			configContent:      []byte(`{"SiteURL":"https://openway.guance.com?token=tkn_123","UpdateTime":1644318398}`),
			failedReadTarToMap: errGeneral,
			expect:             0,
		},
		{
			name:          "remove_error",
			fileExist:     true,
			pathConfig:    configPath,
			configContent: []byte(`{"SiteURL":"https://openway.guance.com?token=tkn_123","UpdateTime":1644318398}`),
			failedRemove:  errGeneral,
			failedReadDir: errGeneral,
			expect:        0,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			isFirst = true // variable from package remote

			mock := newPipelineRemoteMock()
			mock.readFileData = tc.configContent
			mock.isFileExist = tc.fileExist
			mock.errUnMarshal = tc.failedUnMarshal
			mock.errReadFile = tc.failedReadFile
			mock.errRemove = tc.failedRemove
			mock.errReadDir = tc.failedReadDir
			mock.errReadTarToMap = tc.failedReadTarToMap

			n, err := getPipelineRemoteConfig(tc.pathConfig, tc.siteURL, mock)
			assert.Equal(t, tc.expectError, err, "getPipelineRemoteConfig found error: %v", err)
			assert.Equal(t, tc.expect, n, "getPipelineRemoteConfig not equal!")
		})
	}
}

// go test -v -timeout 30s -run ^TestUpdatePipelineRemoteConfig$ gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/pipeline/remote
func TestUpdatePipelineRemoteConfig(t *testing.T) {
	const dwURL = "https://openway.guance.com?token=tkn_123"
	const configPath = "/usr/local/datakit/pipeline_remote/.config_fake"
	const ts = 1644820678

	cases := []struct {
		name            string
		pathConfig      string
		siteURL         string
		latestTime      int64
		failedMarshal   error
		failedWriteFile error
		expectError     error
		expect          *FileDataStruct
	}{
		{
			name:       "normal",
			pathConfig: configPath,
			siteURL:    dwURL,
			latestTime: ts,
			expect: &FileDataStruct{
				FileName: configPath,
				Bytes: func() []byte {
					cf := pipelineRemoteConfig{
						SiteURL:    dwURL,
						UpdateTime: ts,
					}
					bys, err := json.Marshal(cf)
					if err != nil {
						panic(err)
					}
					return bys
				}(),
			},
		},
		{
			name:          "json_fail",
			failedMarshal: errGeneral,
			expectError:   errGeneral,
		},
		{
			name:            "write_fail",
			failedWriteFile: errGeneral,
			expectError:     errGeneral,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			mock := newPipelineRemoteMock()
			mock.errMarshal = tc.failedMarshal
			mock.errWriteFile = tc.failedWriteFile

			err := updatePipelineRemoteConfig(tc.pathConfig, tc.siteURL, tc.latestTime, mock)
			assert.Equal(t, tc.expectError, err, "updatePipelineRemoteConfig found error: %v", err)
			assert.Equal(t, tc.expect, mock.writeFileData, "updatePipelineRemoteConfig not equal!")
		})
	}
}

// go test -v -timeout 30s -run ^TestConvertContentMapToThreeMap$ gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/pipeline/remote
func TestConvertContentMapToThreeMap(t *testing.T) {
	cases := []struct {
		name   string
		in     map[string]string
		expect map[string]map[string]string
	}{
		{
			name: "new",
			in: map[string]string{
				"metric/123.p":  "text123",
				"logging/456.p": "text456",
			},
			expect: map[string]map[string]string{
				"metric": {
					"123.p": "text123",
				},
				"logging": {
					"456.p": "text456",
				},
			},
		},
		{
			name: "old",
			in: map[string]string{
				"123.p": "text123",
				"456.p": "text456",
			},
			expect: map[string]map[string]string{
				".": {
					"123.p": "text123",
					"456.p": "text456",
				},
			},
		},
		{
			name: "append",
			in: map[string]string{
				"metric/123.p":   "text123",
				"logging/456.p":  "text456",
				"metric/1234.p":  "text1234",
				"logging/123.p":  "text123",
				"metric/12345.p": "text12345",
			},
			expect: map[string]map[string]string{
				"metric": {
					"123.p":   "text123",
					"1234.p":  "text1234",
					"12345.p": "text12345",
				},
				"logging": {
					"456.p": "text456",
					"123.p": "text123",
				},
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			out := ConvertContentMapToThreeMap(tc.in)
			assert.Equal(t, tc.expect, out)
		})
	}
}

// go test -v -timeout 30s -run ^TestConvertThreeMapToContentMap$ gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/pipeline/remote
func TestConvertThreeMapToContentMap(t *testing.T) {
	cases := []struct {
		name      string
		in        map[point.Category]map[string]string
		inDefault map[point.Category]string
		expect    map[string]string
	}{
		{
			name: "normal",
			in: map[point.Category]map[string]string{
				point.Logging: {
					"123.p":  "text123",
					"1234.p": "text1234",
				},
				point.Metric: {
					"456.p": "text456",
				},
			},
			inDefault: map[point.Category]string{
				point.Logging: "123.p",
			},
			expect: map[string]string{
				"logging/123.p":                 "text123",
				"logging/1234.p":                "text1234",
				"metric/456.p":                  "text456",
				pipelineRemoteDefaultScriptFile: "{\"logging\":\"123.p\"}",
			},
		},
		{
			name: "normal1",
			in: map[point.Category]map[string]string{
				point.Logging: {
					"123.p":  "text123",
					"1234.p": "text1234",
				},
				point.Metric: {
					"456.p": "text456",
				},
			},
			expect: map[string]string{
				"logging/123.p":  "text123",
				"logging/1234.p": "text1234",
				"metric/456.p":   "text456",
			},
		},
		{
			name: "normal2",
			in: map[point.Category]map[string]string{
				point.Logging: {
					"123.p":  "text123",
					"1234.p": "text1234",
				},
				point.Metric: {
					"456.p": "text456",
				},
			},
			inDefault: map[point.Category]string{},
			expect: map[string]string{
				"logging/123.p":                 "text123",
				"logging/1234.p":                "text1234",
				"metric/456.p":                  "text456",
				pipelineRemoteDefaultScriptFile: "{}",
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			out := convertThreeMapToContentMap(tc.in, tc.inDefault)
			assert.Equal(t, tc.expect, out)
		})
	}
}
