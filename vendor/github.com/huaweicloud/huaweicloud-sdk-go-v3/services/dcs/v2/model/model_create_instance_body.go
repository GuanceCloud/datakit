/*
 * DCS
 *
 * DCS V2版本API
 *
 */

package model

import (
	"encoding/json"

	"strings"
)

type CreateInstanceBody struct {
	// 实例名称。  由英文字符开头，只能由英文字母、数字、中划线和下划线组成。  创建单个实例时，名称长度为4到64位的字符串。批量创建实例时，名称长度为4到56位的字符串，且实例名称格式为“自定义名称-n”，其中n从000开始，依次递增。例如，批量创建两个实例，自定义名称为dcs_demo，则两个实例的名称为dcs_demo-000和dcs_demo-001。
	Name string `json:"name"`
	// 缓存引擎：Redis和Memcached。
	Engine string `json:"engine"`
	// 缓存版本。  当缓存引擎为Redis时，取值为3.0、4.0或5.0。  当缓存引擎为Memcached时，该字段为可选，取值为空。
	EngineVersion *string `json:"engine_version,omitempty"`
	// 缓存容量（G Byte） - Redis3.0：单机和主备类型实例取值：2、4、8、16、32、64。Proxy集群实例规格支持64、128、256、512和1024。 - Redis4.0和Redis5.0：单机和主备类型实例取值：0.125、0.25、0.5、1、2、4、8、16、32、64。Cluster集群实例规格支持24、32、48、64、96、128、192、256、384、512、768、1024。 - Memcached：单机和主备类型实例取值：2、4、8、16、32、64。
	Capacity float32 `json:"capacity"`
	// 产品规格编码。具体查询方法，请参考[查询产品规格](https://support.huaweicloud.com/api-dcs/ListFlavors.html)。
	SpecCode string `json:"spec_code"`
	// 创建缓存节点到指定且有资源的可用区Code。创建缓存节点到指定且有资源的可用区Code。具体查询方法，请参考[查询可用区信息](https://support.huaweicloud.com/api-dcs/ListAvailableZones.html)，在查询时，请注意查看该可用区是否有资源。  如果是创建主备、Proxy集群、Cluster集群实例，支持跨可用区部署，可以为备节点指定备可用区。在为节点指定可用区时，用逗号分隔开，具体请查看示例。
	AzCodes []string `json:"az_codes"`
	// 虚拟私有云ID。  获取方法如下：   - 方法1：登录虚拟私有云服务的控制台界面，在虚拟私有云的详情页面查找VPC ID。   - 方法2：通过虚拟私有云服务的API接口查询，具体操作可参考[查询VPC列表](https://support.huaweicloud.com/api-vpc/vpc_api01_0003.html)。
	VpcId string `json:"vpc_id"`
	// 子网的网络ID。  获取方法如下： - 方法1：登录虚拟私有云服务的控制台界面，单击VPC下的子网，进入子网详情页面，查找网络ID。 - 方法2：通过虚拟私有云服务的API接口查询，具体操作可参考[查询子网列表](https://support.huaweicloud.com/api-vpc/vpc_subnet01_0003.html)。
	SubnetId string `json:"subnet_id"`
	// 指定实例所属的安全组。  当engine为Redis且engine_version为3.0时，或engine为Memcached时，该参数为必选。Redis3.0和Memcached实例支持安全组访问控制。  当engine为Redis且engine_version为4.0和5.0时，该参数为可选。Redis4.0和Redis5.0版本实例不支持安全组控制访问，只支持白名单控制。  获取方法如下： - 方法1：登录虚拟私有云服务的控制台界面，在安全组的详情页面查找安全组ID。 - 方法2：通过虚拟私有云服务的API接口查询，具体操作可参考[查询安全组列表](https://support.huaweicloud.com/api-vpc/vpc_sg01_0002.html)。
	SecurityGroupId *string `json:"security_group_id,omitempty"`
	// Redis缓存实例绑定的弹性IP地址的id。  如果开启了公网访问功能（即enable_publicip为true），该字段为必选。
	PublicipId *string `json:"publicip_id,omitempty"`
	// 企业项目ID。
	EnterpriseProjectId *string `json:"enterprise_project_id,omitempty"`
	// 企业项目名称。
	EnterpriseProjectName *string `json:"enterprise_project_name,omitempty"`
	// 实例的描述信息。  长度不超过1024的字符串。 > \\与\"在json报文中属于特殊字符，如果参数值中需要显示\\或者\"字符，请在字符前增加转义字符\\，比如\\\\或者\\\"。
	Description *string `json:"description,omitempty"`
	// Redis缓存实例开启公网访问功能时，是否选择支持ssl。 - true：开启 - false：不开启
	EnableSsl *bool `json:"enable_ssl,omitempty"`
	// 创建缓存实例手动指定的IP地址,支持Redis和Memcached。
	PrivateIp *string `json:"private_ip,omitempty"`
	// 表示批量创建缓存实例时，购买的实例个数。仅Redis和Memcached实例支持批量创建。  默认值：1  取值范围：1-100
	InstanceNum *int32 `json:"instance_num,omitempty"`
	// 维护时间窗开始时间，为UTC时间，格式为HH:mm:ss - 维护时间窗开始和结束时间必须为指定的时间段，可参考[查询维护时间窗时间段](https://support.huaweicloud.com/api-dcs/ListMaintenanceWindows.html)获取。 - 开始时间必须为22:00:00、02:00:00、06:00:00、10:00:00、14:00:00和18:00:00。 - 该参数不能单独为空，若该值为空，则结束时间也为空。系统分配一个默认开始时间02:00:00。
	MaintainBegin *string `json:"maintain_begin,omitempty"`
	// 维护时间窗结束时间，为UTC时间，格式为HH:mm:ss。 - 维护时间窗开始和结束时间必须为指定的时间段，可参考[查询维护时间窗时间段](https://support.huaweicloud.com/api-dcs/ListMaintenanceWindows.html)获取。 - 结束时间在开始时间基础上加四个小时，即当开始时间为22:00:00时，结束时间为02:00:00。 - 该参数不能单独为空，若该值为空，则开始时间也为空，系统分配一个默认结束时间06:00:00。
	MaintainEnd *string `json:"maintain_end,omitempty"`
	// 缓存实例的认证信息 > 当“no_password_access”配置为“false”或未配置时，请求消息中须包含password参数。 Redis类型的缓存实例密码复杂度要求： - 输入长度为8到32位的字符串。 - 新密码不能与旧密码相同。 - 必须包含如下四种字符中的三种组合：   - 小写字母   - 大写字母   - 数字   - 特殊字符包括（`~!@#$%^&*()-_=+\\|[{}]:'\",<.>/?）
	Password *string `json:"password,omitempty"`
	// 是否允许免密码访问缓存实例。 - true：该实例无需密码即可访问。 - false：该实例必须通过密码认证才能访问。 若未配置该参数则默认值为“false”。
	NoPasswordAccess     *bool         `json:"no_password_access,omitempty"`
	BssParam             *BssParam     `json:"bss_param,omitempty"`
	InstanceBackupPolicy *BackupPolicy `json:"instance_backup_policy,omitempty"`
	// 实例标签键值。
	Tags *[]ResourceTag `json:"tags,omitempty"`
	// 当缓存类型为Redis时，则不需要设置，保持为空即可。  当缓存引擎为Memcached，且“no_password_access”为“false”时才需要设置，表示通过密码认证访问缓存实例的认证用户名。  由英文字符开头，只能由英文字母、数字、中划线和下划线组成，长度为1~64的字符。 >   - 当缓存引擎为Memcached时，该参数为可选项。   - 当缓存引擎为Redis时，该参数不需要设置。
	AccessUser *string `json:"access_user,omitempty"`
	// Redis缓存实例是否开启公网访问功能。 - true：开启 - false：不开启
	EnablePublicip *bool `json:"enable_publicip,omitempty"`
	// 实例自定义端口。只有创建Redis4.0和Redis5.0实例才支持自定义端口，Redis3.0和Memcached实例不支持。  创建Redis4.0和Redis5.0实例，如果没发送该参数或该参数为空，表示实例使用默认端口6379。如果自定义端口，端口范围为1~65535的任意数字。
	Port *int32 `json:"port,omitempty"`
	// 支持自定义重命名高危命令。只有创建Redis4.0和Redis5.0实例才支持重命名高危命令，Redis3.0和Memcached实例不支持。  创建Redis4.0和Redis5.0实例，如果没发送该参数或该参数为空，表示没有需要重命名的命令。当前支持重命名的高危命令有command、keys、flushdb、flushall和hgetall，其他命令暂不支持重命名。
	RenameCommands *interface{} `json:"rename_commands,omitempty"`
}

func (o CreateInstanceBody) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "CreateInstanceBody struct{}"
	}

	return strings.Join([]string{"CreateInstanceBody", string(data)}, " ")
}
