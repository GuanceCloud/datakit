package docker_containers

// import (
// 	"github.com/docker/docker/api/types"
// )

type DockerObject struct {
	Name        string `json:"__name"`
	Tags        Tags   `json:"__tags"`
	Carated     int64  `json:"created"`
	Started     int64  `json:"started_at"`
	Finished    int64  `json:"finished_at"`
	Path        string `json:"path"`
	Description string `json:"__description"`
	//Inspect  types.ContainerJSON `json:"inspect"`
}

type Tags struct {
	Class           string `json:"__class"`
	ContainerID     string `json:"container_id"`
	ContainerName   string `json:"container_name"`
	ContainerImage  string `json:"container_image"`
	ContainerStatue string `json:"container_status"`
	Host            string `json:"host"`
	PID             string `json:"pid"`
}
