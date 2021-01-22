/*
 * ServiceStage
 *
 * ServiceStage的API,包括应用管理和仓库授权管理
 *
 */

package model

import (
	"encoding/json"

	"strings"
)

// 构建任务的环境变量。
type BuildInfoParameters struct {
	// 编译命令。默认：  1、根目录存在build.sh：./build.sh  2、根据运行系统，示例如下：  Java和Tomcat：mvn clean package  Nodejs: npm build
	BuildCmd *string `json:"build_cmd,omitempty"`
	// dockerfile地址。默认是根目录./。
	DockerfilePath *string `json:"dockerfile_path,omitempty"`
	// 构建归档组织，默认cas_{project_id}。
	ArtifactNamespace *string `json:"artifact_namespace,omitempty"`
}

func (o BuildInfoParameters) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "BuildInfoParameters struct{}"
	}

	return strings.Join([]string{"BuildInfoParameters", string(data)}, " ")
}
