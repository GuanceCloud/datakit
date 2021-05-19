package huaweiyunobject

import (
	"fmt"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/pipeline"
)

// das 数据管理服务 DAS
//  - enterprises 企业版
// ddm 分布式数据库中间件 DDM
//  - instances 实例
// live 视频直播 LIVE
//  - domain 域名
// obs 对象存储服务 OBS
//  - buckets 桶
// ecs 弹性云服务器 ECS
//  - cloudservers 云服务器
// vpc 虚拟私有云 VPC
//  - vpcs 虚拟私有云
//  - bandwidths 共享带宽
//  - securityGroups 安全组
//  - publicips 弹性公网IP
// evs 云硬盘 EVS
//  - volumes 磁盘
// ims 镜像服务 IMS
//  - images 镜像
// rds 云数据库 RDS
//  - instances 实例
// dds 文档数据库服务 DDS
//  - instances 实例
// cdn CDN
//  - domains 域名
// elb 弹性负载均衡 ELB
//  - loadbalancers 负载均衡器
//  - listeners 监听器
// nat NAT网关
//  - natGateways 公网NAT网关
// as 弹性伸缩 AS
//  - scalingGroups 弹性伸缩组
// bms 裸金属服务器 BMS
//  - servers 实例
// cbr 云备份
//  - vault 存储库
// cce 云容器引擎 CCE
//  - clusters 集群
// cci 云容器实例 CCI
//  - pods pods
// cloudsite 云速建站
//  - sites 站点
// css 云搜索服务
//  - clusters 集群
// dayu 智能数据湖运营平台 DAYU
//  - instances 实例
//  - cdmClusters 迁移集群
//  - disStreams 接入通道
//  - workspaces 工作空间
// dcs 分布式缓存服务 DCS
//  - redis Redis实例
//  - memcached Memcached实例
// deh 专属主机 DEH
//  - dedicatedhosts 专属主机
// devcloud 软件开发平台DevCloud
//  - projectman 项目管理
//  - codehub 代码托管
//  - codecheck 代码检查
//  - codeci 编译构建
//  - cloudide CloudIDE
//  - releaseman 发布
//  - testman 云测（测试管理）
//  - apitest 云测（接口测试）
//  - package DevCloud按需服务组合
//  - devcloud_package DevCloud包周期基础版/专业版
//  - classroom Classroom
//  - educourse Classroom教学资源
//  - sandbox Classroom实验沙箱
//  - expcourse Classroom实验内容
// dli 数据湖探索 DLI
//  - queues 队列
// dms 分布式消息服务 DMS
//  - queues 队列
//  - kafkas Kafka实例
//  - rabbitmqs rabbitmq实例
// drs 数据复制服务 DRS
//  - synchronizationJob 实时同步任务
//  - migrationJob 实时迁移任务
//  - dataGuardJob 实时灾备任务
//  - subscriptionJob 数据订阅任务
//  - backupMigrationJob 备份迁移任务
// dss 专属分布式存储服务 DSS
//  - dsspools 存储池
// dws 数据仓库服务
//  - clusters 集群
// fgs 函数工作流 FunctionGraph
//  - functions 函数
// gaussdb 云数据库 GaussDB
//  - instance 实例
//  - nodes 节点
// ges 图引擎服务
//  - graphs 图
// hecs 云耀云服务器 HECS
//  - hcloudservers 实例
// hilens 华为HiLens
//  - commission_skill 定制技能
// ief 智能边缘平台
//  - appinstances 容器应用
// iotda 设备接入 IoTDA
//  - iotda 设备接入
// kms 数据加密服务 DEW
//  - keys 密钥
// kps 密钥对管理服务
//  - keypairs 密钥对
// lts 云日志服务 LTS
//  - topics 日志流
// mrs MapReduce服务
//  - mrs 弹性大数据服务
// nosql 云数据库 GaussDB NoSQL
//  - instances 实例
//  - nodes 节点
// res 推荐系统
//  - inferServices 在线服务
// roma 应用与数据集成平台 ROMA
//  - instances 实例
// servicestage 应用管理与运维平台
//  - cseEngines 微服务引擎

func (ag *agent) run() {
	ag.genAPIClient()

	providers, err := ag.listProviders()
	if err != nil {
		return
	}

	pipeLines := map[string]*pipeline.Pipeline{}

	for {

		select {
		case <-ag.ctx.Done():
			return
		default:
		}

		for _, provider := range providers {

			select {
			case <-ag.ctx.Done():
				return
			default:
			}

			if provider.ResourceTypes == nil || len(*provider.ResourceTypes) == 0 {
				continue
			}
			resType := (*provider.ResourceTypes)[0]
			resources, err := ag.listResources(*provider.Provider, *resType.Name)
			if err != nil || len(resources) == 0 {
				continue
			}

			pipename := fmt.Sprintf("%s_%s.p", inputName, *provider.Provider)
			p := pipeLines[pipename]
			if p == nil {
				p = getPipeline(pipename)
				if p != nil {
					pipeLines[pipename] = p
				}
			}

			for _, res := range resources {

				select {
				case <-ag.ctx.Done():
					return
				default:
				}

				err := ag.parseObject(&res, p)
				if err != nil {
					moduleLogger.Errorf("%s", err)
				}
			}
		}

		datakit.SleepContext(ag.ctx, ag.Interval.Duration)
	}

}
