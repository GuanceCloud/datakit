package docker_containers

// import (
// 	"github.com/docker/docker/api/types"
// )

type ContainerObject struct {
	Name    string `json:"name"`
	Class   string `json:"class"`
	Content string `json:"content"`

	//Inspect  types.ContainerJSON `json:"inspect"`
}

type ObjectContent struct {
	ContainerID     string `json:"container_id"`
	ContainerName   string `json:"container_name"`
	ContainerImage  string `json:"container_image"`
	ContainerStatue string `json:"container_status"`
	Host            string `json:"host"`
	PID             string `json:"pid"`
	Carated         int64  `json:"created"`
	Started         int64  `json:"started_at"`
	Finished        int64  `json:"finished_at"`
	Path            string `json:"path"`
	Inspect         string `json:"inspect,omitempty"`
}
