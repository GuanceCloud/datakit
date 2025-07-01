// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package logging

import (
	T "testing"

	bstoml "github.com/BurntSushi/toml"
	"github.com/stretchr/testify/assert"
)

func TestEmpty(t *T.T) {
	// placeholder for the package

	t.Run(`sample-decode`, func(t *T.T) {
		ipt := defaultInput()

		_, err := bstoml.Decode(`
  logfiles = [
    # UNIX-like log path example:
    # "/var/log/*.log",  # All log files in the directory
    # "/var/log/*.txt",  # All txt files in the directory
    # "/var/log/sys*",   # All files prefixed with sys in the directory
    # "/var/log/syslog", # Unix-style file path

    # Windows log path example:
    # "C:/path/to/*.txt",
    # or like this(with space in path):
    # '''C:\\path\\to some\\*.txt''',
  ]
  ## Socket currently supports two protocols: tcp/udp. It is recommended to use internal
  ## network ports for security.
  sockets = [
	 "tcp://0.0.0.0:9540",
   "udp://0.0.0.0:9541"
  ]
		`, ipt)
		assert.NoError(t, err)

		assert.NotNil(t, ipt.LogFiles)
		assert.Len(t, ipt.LogFiles, 0)
		assert.NotNilf(t, ipt.LogFiles, "input: %+#v", ipt)
		assert.NotNilf(t, ipt.Sockets, "input: %+#v", ipt)
		assert.Len(t, ipt.Sockets, 2)
	})
}
