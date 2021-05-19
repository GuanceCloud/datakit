package dockerlog

import (
	"context"
	"testing"

	iod "gitlab.jiagouyun.com/cloudcare-tools/datakit/io"
)

func TestMain(t *testing.T) {
	iod.TestOutput()

	var d = DockerLogs{
		Endpoint:      defaultEndpoint,
		newEnvClient:  NewEnvClient,
		newClient:     NewClient,
		containerList: make(map[string]context.CancelFunc),
		FromBeginning: false,
		Timeout:       "5s",
	}

	d.Run()
}
