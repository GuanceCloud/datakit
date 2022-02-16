package cmds

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/config"
)

func TestGetLogFile(t *testing.T) {
	tmpDir, err := ioutil.TempDir("./", "__tmp")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir) // nolint:errcheck
	logFile, err := ioutil.TempFile(tmpDir, "log")
	assert.NoError(t, err, "create temp log file failed")
	config.Cfg.Logging.Log, err = filepath.Abs("./" + logFile.Name())
	if err != nil {
		l.Fatal(err)
	}
	logFileName, err := getLogFile()
	assert.NoError(t, err)
	assert.Contains(t, logFileName, os.TempDir())
	os.Remove(logFileName) //nolint: errcheck
}

func TestUploadLog(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "{\"msg\": \"ok\"}")
	}))
	defer ts.Close()

	tmpDir, err := ioutil.TempDir("./", "__tmp")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir) // nolint:errcheck

	logFile, err := ioutil.TempFile(tmpDir, "log")
	assert.NoError(t, err, "create temp log file failed")
	config.Cfg.Logging.Log, err = filepath.Abs("./" + logFile.Name())
	if err != nil {
		t.Fatal(err)
	}

	err = uploadLog([]string{ts.URL})
	assert.NoError(t, err)

	err = uploadLog([]string{})
	assert.Error(t, err)

	config.Cfg.Logging.Log = ""
	err = uploadLog([]string{ts.URL})
	assert.Error(t, err, "should be an error, as log file name is empty")
}
