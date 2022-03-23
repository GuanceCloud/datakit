package container

import (
	"context"
	"time"

	"github.com/containerd/containerd"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
)

type containerdInput struct {
	client *containerd.Client
	cfg    *containerdInputConfig
}

type containerdInputConfig struct {
	endpoint  string
	extraTags map[string]string
}

func newContainerdInput(cfg *containerdInputConfig) (*containerdInput, error) {
	cli, err := containerd.New(cfg.endpoint)
	if err != nil {
		return nil, err
	}

	return &containerdInput{client: cli, cfg: cfg}, nil
}

func (c *containerdInput) gatherObject() ([]inputs.Measurement, error) {
	var res []inputs.Measurement
	cList, err := c.client.Containers(context.Background())
	if err != nil {
		return nil, err
	}

	for _, container := range cList {
		info, err := container.Info(context.TODO())
		if err != nil {
			continue
		}
		obj := &containerdObject{time: time.Now()}

		imageName, imageShortName, imageTag := ParseImage(info.Image)
		obj.tags = map[string]string{
			"name":             info.ID,
			"container_id":     info.ID,
			"image":            info.Image,
			"image_name":       imageName,
			"image_short_name": imageShortName,
			"image_tag":        imageTag,
			"runtime":          info.Runtime.Name,
			"container_type":   "containerd",
		}
		obj.fields = map[string]interface{}{
			"age": time.Since(info.CreatedAt) / 1e3,
		}

		obj.tags.addValueIfNotEmpty("pod_name", info.Labels[containerLableForPodName])
		obj.tags.addValueIfNotEmpty("pod_namespace", info.Labels[containerLableForPodNamespace])
		for k, v := range c.cfg.extraTags {
			obj.tags[k] = v
		}
		res = append(res, obj)
	}

	return res, nil
}

type containerdObject struct {
	tags   tagsType
	fields fieldsType
	time   time.Time
}

func (c *containerdObject) LineProto() (*io.Point, error) {
	// 此处使用 docker_containers 不合适
	return io.NewPoint(dockerContainerName, c.tags, c.fields, &io.PointOption{Time: c.time, Category: datakit.Object})
}

func (c *containerdObject) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name: dockerContainerName,
		Desc: "containerd 容器对象数据",
		Type: "object",
		Tags: map[string]interface{}{
			"container_id":     inputs.NewTagInfo(`容器 ID（该字段默认被删除）`),
			"name":             inputs.NewTagInfo(`对象数据的指定 ID`),
			"image":            inputs.NewTagInfo("镜像全称，例如 `nginx.org/nginx:1.21.0`"),
			"image_name":       inputs.NewTagInfo("镜像名称，例如 `nginx.org/nginx`"),
			"image_short_name": inputs.NewTagInfo("镜像名称精简版，例如 `nginx`"),
			"image_tag":        inputs.NewTagInfo("镜像tag，例如 `1.21.0`"),
			"container_type":   inputs.NewTagInfo(`容器类型，表明该容器由谁创建，kubernetes/docker/containerd`),
			"pod_name":         inputs.NewTagInfo(`pod 名称（容器由 k8s 创建时存在）`),
			"pod_namespace":    inputs.NewTagInfo(`pod 命名空间（容器由 k8s 创建时存在）`),
		},
		Fields: map[string]interface{}{
			"age": &inputs.FieldInfo{DataType: inputs.Int, Unit: inputs.DurationSecond, Desc: `该容器创建时长，单位秒`},
		},
	}
}
