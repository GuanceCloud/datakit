package cmds

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetLogFile(t *testing.T) {
	_, err := getLogFile()
	assert.NoError(t, err)
}

func TestUploadLog(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "ok")
	}))

	defer ts.Close()
	err := uploadLog([]string{ts.URL})
	assert.NoError(t, err)
}
