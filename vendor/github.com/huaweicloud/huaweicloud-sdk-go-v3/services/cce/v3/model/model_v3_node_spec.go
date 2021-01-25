/*
 * CCE
 *
 * CCE开放API
 *
 */

package model

import (
	"encoding/json"

	"strings"
)

type V3NodeSpec struct {
	//   节点所在的可用区名. 底层实际存在，位于该用户物理可用区组之内的可用区
	Az string `json:"az"`
	// 节点的计费模式：取值为 0（按需付费）、2（自动付费包周期）  自动付费包周期支持普通用户token。 >创建按需节点不影响集群状态；创建包周期节点时，集群状态会转换为“扩容中”。
	BillingMode *int32 `json:"billingMode,omitempty"`
	// 批量创建时节点的个数，必须为大于等于1，小于等于最大限额的正整数。作用于节点池时该项允许为0
	Count int32 `json:"count"`
	// 节点的数据盘参数（目前已支持通过控制台为CCE节点添加第二块数据盘）。  针对专属云节点，参数解释与rootVolume一致
	DataVolumes []V3DataVolume `json:"dataVolumes"`
	// 指定DeH主机的ID，将节点调度到自己的DeH上。\\n>创建节点池添加节点时不支持该参数。
	DedicatedHostId *string `json:"dedicatedHostId,omitempty"`
	// 云服务器组ID，若指定，将节点创建在该云服务器组下
	EcsGroupId *string `json:"ecsGroupId,omitempty"`
	// 创建节点时的扩展参数，可选参数如下： - chargingMode: 节点的计费模式。按需计费，取值为“0”，若不填，则默认为“0”。 - ecs:performancetype：云服务器规格的分类。裸金属节点无该字段。 - orderID: 订单ID，节点付费类型为自动付费包周期类型时必选。 - productID: 产品ID。 - maxPods: 节点最大允许创建的实例数(Pod)，该数量包含系统默认实例，取值范围为16~256。   该设置的目的为防止节点因管理过多实例而负载过重，请根据您的业务需要进行设置。 - periodType:    订购周期类型，取值范围：     - month：月     - year：年   > billingMode为2（自动付费包周期）时生效，且为必选。 - periodNum:   订购周期数，取值范围：     - periodType=month（周期类型为月）时，取值为[1-9]。     - periodType=year（周期类型为年）时，取值为1。   > billingMode为2时生效，且为必选。 - isAutoRenew:   是否自动续订     - “true”：自动续订     - “false”：不自动续订   > billingMode为2时生效，且为必选。 - isAutoPay:   是否自动扣款     - “true”：自动扣款     - “false”：不自动扣款   > billingMode为2时生效，不填写此参数时默认会自动扣款。 - DockerLVMConfigOverride:   Docker数据盘配置项。默认配置示例如下：   ```   \"DockerLVMConfigOverride\":\"dockerThinpool=vgpaas/90%VG;kubernetesLV=vgpaas/10%VG;diskType=evs;lvType=linear\"   ```   包含如下字段：     - userLV：用户空间的大小，示例格式：vgpaas/20%VG     - userPath：用户空间挂载路径，示例格式：/home/wqt-test     - diskType：磁盘类型，目前只有evs、hdd和ssd三种格式     - lvType：逻辑卷的类型，目前支持linear和striped两种，示例格式：striped     - dockerThinpool：Docker盘的空间大小，示例格式：vgpaas/60%VG     - kubernetesLV：Kubelet空间大小，示例格式：vgpaas/20%VG - dockerBaseSize:   Device mapper 模式下，节点上 docker  单容器的可用磁盘空间大小 - init-node-password: 节点初始密码 - offloadNode: 是否为CCE Turbo集群节点 - publicKey: 节点的公钥。 - alpha.cce/preInstall:    安装前执行脚本   > 输入的值需要经过Base64编码，方法为echo -n \"待编码内容\" | base64。 - alpha.cce/postInstall:   安装后执行脚本   > 输入的值需要经过Base64编码，方法为echo -n \"待编码内容\" | base64。 - alpha.cce/NodeImageID: 如果创建裸金属节点，需要使用自定义镜像时用此参数。 - interruption_policy:   竞享实例中断策略，当前仅支持immediate。   - 仅marketType=spot时，该字段才可配置。   - 当interruption_policy=immediate时表示释放策略为立即释放。 - spot_duration_hours:   购买的竞享实例时长。   - 仅interruption_policy=immediate 时该字段才可配置。   - spot_duration_hours须大于0。   - spot_duration_hours的前端最大值从flavor的extra_specs的cond:spot_block:operation:longest_duration_hours字段中查询。 - spot_duration_count：   购买的“竞享实例时长”的个数。   - 仅spot_duration_hours>0时该字段才可配置。   - spot_duration_hours小于6时，spot_duration_count必须等于1。   - spot_duration_hours等于6时，spot_duration_count大于等于1。   - spot_duration_count的前端最大值从flavor的extra_specs的cond:spot_block:operation:longest_duration_count字段中查询。
	ExtendParam map[string]interface{} `json:"extendParam,omitempty"`
	// 节点的规格
	Flavor string `json:"flavor"`
	// 格式为key/value键值对。键值对个数不超过20条。  - Key：必须以字母或数字开头，可以包含字母、数字、连字符、下划线和点，最长63个字符；另外可以使用DNS子域作为前缀，例如example.com/my-key， DNS子域最长253个字符。 - Value：可以为空或者非空字符串，非空字符串必须以字符或数字开头，可以包含字母、数字、连字符、下划线和点，最长63个字符。  示例：  ``` \"k8sTags\": {  \"key\": \"value\" } ```
	K8sTags     map[string]string `json:"k8sTags,omitempty"`
	Login       *Login            `json:"login"`
	NodeNicSpec *NodeNicSpec      `json:"nodeNicSpec,omitempty"`
	// 是否CCE Turbo集群节点 >创建节点池添加节点时不支持该参数。
	OffloadNode *bool `json:"offloadNode,omitempty"`
	// 节点的操作系统类型。  - 对于虚拟机节点，可以配置为“EulerOS”、“CentOS”、“Debian”、“Ubuntu”。默认为\"EulerOS\"。  > 系统会根据集群版本自动选择支持的系统版本。当前集群版本不支持该系统类型，则会报错。  - 对于自动付费包周期的裸金属节点，只支持EulerOS 2.3、EulerOS 2.5、EulerOS 2.8。
	Os         string          `json:"os"`
	PublicIP   *V3NodePublicIp `json:"publicIP,omitempty"`
	RootVolume *V3RootVolume   `json:"rootVolume"`
	// 支持给创建出来的节点加Taints来设置反亲和性，每条Taints包含以下3个参数：  - Key：必须以字母或数字开头，可以包含字母、数字、连字符、下划线和点，最长63个字符；另外可以使用DNS子域作为前缀。 - Value：必须以字符或数字开头，可以包含字母、数字、连字符、下划线和点，最长63个字符。 - Effect：只可选NoSchedule，PreferNoSchedule或NoExecute。  示例：  ``` \"taints\": [{ \"key\": \"status\", \"value\": \"unavailable\", \"effect\": \"NoSchedule\" }, { \"key\": \"looks\", \"value\": \"bad\", \"effect\": \"NoSchedule\" }] ```
	Taints *[]Taint `json:"taints,omitempty"`
	// 云服务器标签，键必须唯一，CCE支持的最大用户自定义标签数量依region而定，自定义标签数上限最少为5个。
	UserTags *[]UserTag `json:"userTags,omitempty"`
}

func (o V3NodeSpec) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "V3NodeSpec struct{}"
	}

	return strings.Join([]string{"V3NodeSpec", string(data)}, " ")
}
