package remote

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

//------------------------------------------------------------------------------

var (
	writeFileData *FileDataStruct
	readFileData  []byte
	isFileExist   bool
	readDirResult []fs.FileInfo

	errGeneral      = fmt.Errorf("test_specific_error")
	errMarshal      error
	errUnMarshal    error
	errReadFile     error
	errWriteFile    error
	errReadDir      error
	errPullPipeline error
)

func resetVars() {
	writeFileData = nil
	readFileData = []byte{}
	isFileExist = false
	readDirResult = []fs.FileInfo{}

	errMarshal = nil
	errUnMarshal = nil
	errReadFile = nil
	errWriteFile = nil
	errReadDir = nil
	errPullPipeline = nil
}

type pipelineRemoteMockerTest struct{}

type FileDataStruct struct {
	FileName string
	Bytes    []byte
}

func (*pipelineRemoteMockerTest) FileExist(filename string) bool {
	return isFileExist
}

func (*pipelineRemoteMockerTest) Marshal(v interface{}) ([]byte, error) {
	if errMarshal != nil {
		return nil, errMarshal
	}

	return json.Marshal(v)
}

func (*pipelineRemoteMockerTest) Unmarshal(data []byte, v interface{}) error {
	if errUnMarshal != nil {
		return errUnMarshal
	}

	return json.Unmarshal(data, v)
}

func (*pipelineRemoteMockerTest) ReadFile(filename string) ([]byte, error) {
	if errReadFile != nil {
		return nil, errReadFile
	}

	return readFileData, nil
}

func (*pipelineRemoteMockerTest) WriteFile(filename string, data []byte, perm fs.FileMode) error {
	if errWriteFile != nil {
		return errWriteFile
	}

	writeFileData = &FileDataStruct{
		FileName: filename,
		Bytes:    data,
	}
	return nil
}

func (*pipelineRemoteMockerTest) ReadDir(dirname string) ([]fs.FileInfo, error) {
	if errReadDir != nil {
		return nil, errReadDir
	}

	return readDirResult, nil
}

func (*pipelineRemoteMockerTest) PullPipeline(ts int64) (mFiles map[string]string, updateTime int64, err error) {
	if errPullPipeline != nil {
		return nil, 0, errPullPipeline
	}

	return map[string]string{
		"123.p": "text123",
		"456.p": "text456",
	}, 1644318398, nil
}

func (*pipelineRemoteMockerTest) GetTickerDurationAndBreak() (time.Duration, bool) {
	return time.Second, true
}

//------------------------------------------------------------------------------

// go test -v -timeout 30s -run ^TestPullMain$ gitlab.jiagouyun.com/cloudcare-tools/datakit/pipeline/remote
func TestPullMain(t *testing.T) {
	const dwURL = "https://openway.guance.com?token=tkn_3659483096cf4cbabef3e244a917fa90"
	const configPath = "/usr/local/datakit/pipeline_remote/.config_fake"

	cases := []struct {
		name           string
		fileExist      bool
		urls           []string
		pathConfig     string
		siteURL        string
		configContent  []byte
		failedReadFile error
		expectError    error
	}{
		{
			name:          "normal",
			urls:          []string{"https://openway.guance.com?token=tkn_3659483096cf4cbabef3e244a917fa90"},
			pathConfig:    configPath,
			siteURL:       dwURL,
			configContent: []byte(`{"SiteURL":"https://openway.guance.com?token=tkn_3659483096cf4cbabef3e244a917fa90","UpdateTime":1644318398}`),
		},
		{
			name: "urls_zero",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			resetVars()
			isFileExist = tc.fileExist
			errReadFile = tc.failedReadFile

			err := pullMain(tc.urls, &pipelineRemoteMockerTest{})
			assert.Equal(t, tc.expectError, err, "pullMain found error: %v", err)
		})
	}
}

// go test -v -timeout 30s -run ^TestDoPull$ gitlab.jiagouyun.com/cloudcare-tools/datakit/pipeline/remote
func TestDoPull(t *testing.T) {
	const dwURL = "https://openway.guance.com?token=tkn_3659483096cf4cbabef3e244a917fa90"
	const configPath = "/usr/local/datakit/pipeline_remote/.config_fake"

	cases := []struct {
		name               string
		fileExist          bool
		pathConfig         string
		siteURL            string
		configContent      []byte
		failedMarshal      error
		failedReadFile     error
		failedReadDir      error
		failedPullPipeline error
		expectError        error
	}{
		{
			name:          "update",
			pathConfig:    configPath,
			siteURL:       dwURL,
			configContent: []byte(`{"SiteURL":"https://openway.guance.com?token=tkn_3659483096cf4cbabef3e244a917fa90","UpdateTime":1644318398}`),
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
			configContent: []byte(`{"SiteURL":"https://openway.guance.com?token=tkn_3659483096cf4cbabef3e244a917fa90","UpdateTime":1644318398}`),
		},
		{
			name:          "dumpfile_fail",
			pathConfig:    configPath,
			siteURL:       dwURL,
			configContent: []byte(`{"SiteURL":"https://openway.guance.com?token=tkn_3659483096cf4cbabef3e244a917fa90","UpdateTime":1644318398}`),
			failedReadDir: errGeneral,
			expectError:   errGeneral,
		},
		{
			name:          "updatePipelineRemoteConfig_fail",
			pathConfig:    configPath,
			siteURL:       dwURL,
			configContent: []byte(`{"SiteURL":"https://openway.guance.com?token=tkn_3659483096cf4cbabef3e244a917fa90","UpdateTime":1644318398}`),
			failedMarshal: errGeneral,
			expectError:   errGeneral,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			resetVars()
			readFileData = tc.configContent
			isFileExist = tc.fileExist
			errMarshal = tc.failedMarshal
			errReadFile = tc.failedReadFile
			errReadDir = tc.failedReadDir
			errPullPipeline = tc.failedPullPipeline

			err := doPull(tc.pathConfig, tc.siteURL, &pipelineRemoteMockerTest{})
			assert.Equal(t, tc.expectError, err, "doPull found error: %v", err)
		})
	}
}

// go test -v -timeout 30s -run ^TestDumpFiles$ gitlab.jiagouyun.com/cloudcare-tools/datakit/pipeline/remote
func TestDumpFiles(t *testing.T) {
	cases := []struct {
		name            string
		files           map[string]string
		readDir         []fs.FileInfo
		failedReadDir   error
		failedWriteFile error
		expectError     error
		expect          []string
	}{
		{
			name: "normal",
			files: map[string]string{
				"123.p": "text123",
				"456.p": "text456",
			},
			expect: []string{
				"/usr/local/datakit/pipeline_remote/123.p",
				"/usr/local/datakit/pipeline_remote/456.p",
			},
		},
		{
			name:          "read_dir_fail",
			failedReadDir: errGeneral,
			expectError:   errGeneral,
		},
		{
			name:            "write_file_fail",
			failedWriteFile: errGeneral,
			files: map[string]string{
				"123.p": "text123",
				"456.p": "text456",
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			resetVars()
			errWriteFile = tc.failedWriteFile
			errReadDir = tc.failedReadDir

			arr, err := dumpFiles(tc.files, &pipelineRemoteMockerTest{})
			assert.Equal(t, tc.expectError, err, "dumpFiles found error: %v", err)

			// cannot compare []string directly because of golang map random sort.
			assert.Equal(t, len(tc.expect), len(arr), "dumpFiles length not equal!")
			mV := make(map[string]struct{})
			for _, v1 := range arr {
				mV[v1] = struct{}{}
			}
			for _, v := range tc.expect {
				if _, ok := mV[v]; !ok {
					assert.Fail(t, "dumpFiles not found: %s", v)
				}
			}
		})
	}
}

// go test -v -timeout 30s -run ^TestGetPipelineRemoteConfig$ gitlab.jiagouyun.com/cloudcare-tools/datakit/pipeline/remote
func TestGetPipelineRemoteConfig(t *testing.T) {
	const dwURL = "https://openway.guance.com?token=tkn_3659483096cf4cbabef3e244a917fa90"
	const configPath = "/usr/local/datakit/pipeline_remote/.config_fake"

	cases := []struct {
		name            string
		fileExist       bool
		pathConfig      string
		siteURL         string
		configContent   []byte
		failedUnMarshal error
		failedReadFile  error
		expectError     error
		expect          int64
	}{
		{
			name:          "normal",
			fileExist:     true,
			pathConfig:    configPath,
			siteURL:       dwURL,
			configContent: []byte(`{"SiteURL":"https://openway.guance.com?token=tkn_3659483096cf4cbabef3e244a917fa90","UpdateTime":1644318398}`),
			expect:        1644318398,
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
			configContent: []byte(`{"SiteURL":"http://127.0.0.1:9528?token=tkn_3659483096cf4cbabef3e244a917fa90","UpdateTime":1644318398}`),
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			resetVars()
			readFileData = tc.configContent
			isFileExist = tc.fileExist
			errUnMarshal = tc.failedUnMarshal
			errReadFile = tc.failedReadFile

			n, err := getPipelineRemoteConfig(tc.pathConfig, tc.siteURL, &pipelineRemoteMockerTest{})
			assert.Equal(t, tc.expectError, err, "getPipelineRemoteConfig found error: %v", err)
			assert.Equal(t, tc.expect, n, "getPipelineRemoteConfig not equal!")
		})
	}
}

// go test -v -timeout 30s -run ^TestUpdatePipelineRemoteConfig$ gitlab.jiagouyun.com/cloudcare-tools/datakit/pipeline/remote
func TestUpdatePipelineRemoteConfig(t *testing.T) {
	const dwURL = "https://openway.guance.com?token=tkn_3659483096cf4cbabef3e244a917fa90"
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
			resetVars()
			errMarshal = tc.failedMarshal
			errWriteFile = tc.failedWriteFile

			err := updatePipelineRemoteConfig(tc.pathConfig, tc.siteURL, tc.latestTime, &pipelineRemoteMockerTest{})
			assert.Equal(t, tc.expectError, err, "updatePipelineRemoteConfig found error: %v", err)
			assert.Equal(t, tc.expect, writeFileData, "updatePipelineRemoteConfig not equal!")
		})
	}
}
