package dockerlog

import (
	"context"
	"testing"
)

func TestMain(t *testing.T) {
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
