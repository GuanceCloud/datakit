package cmds

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/config"
)

func TestGetLogFile(t *testing.T) {
	logFile, err := ioutil.TempFile(os.TempDir(), "log")
	assert.NoError(t, err, "create temp log file failed")
	defer os.Remove(logFile.Name()) //nolint: errcheck
	config.Cfg.Logging.Log = logFile.Name()
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

	logFile, err := ioutil.TempFile(os.TempDir(), "log")
	assert.NoError(t, err, "create temp log file failed")
	defer os.Remove(logFile.Name()) //nolint: errcheck
	config.Cfg.Logging.Log = logFile.Name()

	err = uploadLog([]string{ts.URL})
	assert.NoError(t, err)

	err = uploadLog([]string{})
	assert.Error(t, err)

	config.Cfg.Logging.Log = ""
	err = uploadLog([]string{ts.URL})
	assert.Error(t, err, "should be an error, as log file name is empty")
}
