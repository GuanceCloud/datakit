CREATE TABLE `main_workspace` (
  `id` int(11) NOT NULL AUTO_INCREMENT COMMENT '自增 ID',
  `uuid` varchar(48) NOT NULL COMMENT '全局唯一 ID，带 wksp_',
  `name` varchar(128) NOT NULL COMMENT '命名',
  `token` varchar(64) DEFAULT '""' COMMENT '采集数据token',
  `dbUUID` varchar(48) NOT NULL COMMENT 'influx_db uuid对应influx实例的DB名',
  `dataRestoration` json DEFAULT NULL COMMENT '数据权限',
  `dashboardUUID` varchar(48) DEFAULT NULL COMMENT '工作空间概览-视图UUID',
  `exterId` varchar(128) NOT NULL DEFAULT '' COMMENT '外部ID',
  `status` int(11) NOT NULL DEFAULT '0' COMMENT '状态 0: ok/1: 故障/2: 停用/3: 删除',
  `creator` varchar(64) NOT NULL DEFAULT '' COMMENT '创建者 account-id',
  `updator` varchar(64) NOT NULL DEFAULT '' COMMENT '更新者 account-id',
  `createAt` int(11) NOT NULL DEFAULT '-1',
  `deleteAt` int(11) NOT NULL DEFAULT '-1',
  `updateAt` int(11) NOT NULL DEFAULT '-1',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_uuid` (`uuid`) COMMENT 'UUID 做成全局唯一'
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE `biz_scene` (
  `id` int(11) NOT NULL AUTO_INCREMENT COMMENT '自增 ID',
  `uuid` varchar(48) NOT NULL COMMENT '全局唯一 ID，带 scene_',
  `name` varchar(128) NOT NULL DEFAULT '' COMMENT '场景名称',
  `workspaceUUID` varchar(48) NOT NULL DEFAULT '' COMMENT '工作空间UUID',
  `describe` text NOT NULL COMMENT '场景的描述信息',
  `status` int(11) NOT NULL DEFAULT '0' COMMENT '状态 0: ok/1: 故障/2: 停用/3: 删除',
  `creator` varchar(64) NOT NULL DEFAULT '' COMMENT '创建者 account-id',
  `updator` varchar(64) NOT NULL DEFAULT '' COMMENT '更新者 account-id',
  `createAt` int(11) NOT NULL DEFAULT '-1',
  `deleteAt` int(11) NOT NULL DEFAULT '-1',
  `updateAt` int(11) NOT NULL DEFAULT '-1',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_uuid` (`uuid`) COMMENT 'UUID 做成全局唯一',
  KEY `k_ws_uuid` (`workspaceUUID`),
  CONSTRAINT `scene_worspace_fk` FOREIGN KEY (`workspaceUUID`) REFERENCES `main_workspace` (`uuid`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;



CREATE TABLE `biz_node` (
  `id` int(11) NOT NULL AUTO_INCREMENT COMMENT '自增 ID',
  `uuid` varchar(48) NOT NULL COMMENT '全局唯一 ID，带 node_',
  `name` varchar(128) NOT NULL COMMENT '命名',
  `filter` json DEFAULT NULL COMMENT '过滤条件',
  `subTagKeys` json NOT NULL COMMENT '子节点 tag 键值',
  `bindTagValues` json NOT NULL COMMENT '绑定虚拟节点值',
  `workspaceUUID` varchar(48) NOT NULL DEFAULT '' COMMENT '工作空间UUID',
  `sceneUUID` varchar(48) NOT NULL COMMENT '场景 uuid',
  `parentUUID` varchar(48) NOT NULL DEFAULT '' COMMENT '父节点 uuid',
  `templateUUID` varchar(48) NOT NULL DEFAULT '' COMMENT '系统模板 uuid',
  `dashboardUUID` varchar(48) NOT NULL DEFAULT '' COMMENT '视图 uuid',
  `subTemplateUUID` varchar(48) NOT NULL DEFAULT '',
  `subDashboardUUID` varchar(48) NOT NULL DEFAULT '',
  `exclude` json NOT NULL COMMENT '排除项',
  `path` json DEFAULT NULL COMMENT '路径列表',
  `status` int(11) NOT NULL DEFAULT '0' COMMENT '状态 0: ok/1: 故障/2: 停用/3: 删除',
  `creator` varchar(64) NOT NULL DEFAULT '' COMMENT '创建者 account-id',
  `updator` varchar(64) NOT NULL DEFAULT '' COMMENT '更新者 account-id',
  `createAt` int(11) NOT NULL DEFAULT '-1',
  `deleteAt` int(11) NOT NULL DEFAULT '-1',
  `updateAt` int(11) NOT NULL DEFAULT '-1',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_uuid` (`uuid`) COMMENT 'UUID 做成全局唯一',
  KEY `k_ws_uuid` (`workspaceUUID`),
  KEY `scene_node_fk` (`sceneUUID`),
  KEY `node_dashboard_fk` (`dashboardUUID`),
  CONSTRAINT `node_worspace_fk` FOREIGN KEY (`workspaceUUID`) REFERENCES `main_workspace` (`uuid`),
  CONSTRAINT `scene_node_fk` FOREIGN KEY (`sceneUUID`) REFERENCES `biz_scene` (`uuid`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;


CREATE TABLE `biz_dashboard` (
  `id` int(11) NOT NULL AUTO_INCREMENT COMMENT '自增 ID',
  `uuid` varchar(48) NOT NULL COMMENT '全局唯一 ID，带 dshbrd_',
  `workspaceUUID` varchar(48) NOT NULL DEFAULT '' COMMENT '工作空间UUID',
  `name` varchar(128) NOT NULL COMMENT '视图名字',
  `status` int(11) NOT NULL DEFAULT '0' COMMENT '状态 0: ok/1: 故障/2: 停用/3: 删除',
  `chartPos` json DEFAULT NULL COMMENT 'charts 位置信息[{chartUUID:xxx,pos:xxx}]',
  `chartGroupPos` json DEFAULT NULL COMMENT 'chartGroup 位置信息[chartGroupUUIDs]',
  `type` varchar(48) NOT NULL COMMENT '视图类型：仪表板视图',
  `creator` varchar(64) NOT NULL DEFAULT '' COMMENT '创建者 account-id',
  `updator` varchar(64) NOT NULL DEFAULT '' COMMENT '更新者 account-id',
  `createAt` int(11) NOT NULL DEFAULT '-1',
  `deleteAt` int(11) NOT NULL DEFAULT '-1',
  `updateAt` int(11) NOT NULL DEFAULT '-1',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_uuid` (`uuid`) COMMENT 'UUID 做成全局唯一',
  KEY `k_ws_uuid` (`workspaceUUID`),
  CONSTRAINT `dsbrd_worspace_fk` FOREIGN KEY (`workspaceUUID`) REFERENCES `main_workspace` (`uuid`)
) ENGINE=InnoDB  DEFAULT CHARSET=utf8mb4;


CREATE TABLE `biz_chart_group` (
  `id` int(11) NOT NULL AUTO_INCREMENT COMMENT '自增 ID',
  `uuid` varchar(48) NOT NULL DEFAULT '' COMMENT '全局唯一 ID chtg_',
  `name` varchar(64) NOT NULL DEFAULT '' COMMENT '命名,用云备注',
  `dashboardUUID` varchar(65) NOT NULL DEFAULT '视图UUID',
  `workspaceUUID` varchar(48) NOT NULL DEFAULT '' COMMENT '工作空间UUID',
  `status` int(11) NOT NULL DEFAULT '0',
  `creator` varchar(64) NOT NULL DEFAULT '' COMMENT '创建者 account-id',
  `updator` varchar(64) NOT NULL DEFAULT '' COMMENT '更新者 account-id',
  `createAt` int(11) NOT NULL DEFAULT '-1',
  `deleteAt` int(11) NOT NULL DEFAULT '-1',
  `updateAt` int(11) NOT NULL DEFAULT '-1',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_uuid` (`uuid`) COMMENT 'UUID 做成全局唯一',
  KEY `cgroup_worspace_fk` (`workspaceUUID`),
  KEY `k_dashboardUUID` (`dashboardUUID`),
  CONSTRAINT `cgroup_worspace_fk` FOREIGN KEY (`workspaceUUID`) REFERENCES `main_workspace` (`uuid`)
) ENGINE=InnoDB  DEFAULT CHARSET=utf8mb4;


CREATE TABLE `biz_chart` (
  `id` int(11) NOT NULL AUTO_INCREMENT COMMENT '自增 ID',
  `uuid` varchar(48) NOT NULL COMMENT '全局唯一 ID，带 chrt_ 前缀',
  `name` varchar(128) NOT NULL DEFAULT '' COMMENT '命名',
  `workspaceUUID` varchar(48) NOT NULL DEFAULT '' COMMENT '工作空间UUID',
  `chartGroupUUID` varchar(65) DEFAULT NULL COMMENT '图表分组UUID',
  `dashboardUUID` varchar(65) DEFAULT NULL COMMENT '所属视图UUID',
  `type` varchar(48) NOT NULL COMMENT '图表线条类型',
  `extend` json NOT NULL COMMENT '额外拓展字段',
  `status` int(11) NOT NULL DEFAULT '0' COMMENT '状态 0: ok/1: 故障/2: 停用/3: 删除',
  `creator` varchar(64) NOT NULL DEFAULT '' COMMENT '创建者 account-id',
  `updator` varchar(64) NOT NULL DEFAULT '' COMMENT '更新者 account-id',
  `createAt` int(11) NOT NULL DEFAULT '-1',
  `deleteAt` int(11) NOT NULL DEFAULT '-1',
  `updateAt` int(11) NOT NULL DEFAULT '-1',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_uuid` (`uuid`) COMMENT 'UUID 做成全局唯一',
  KEY `k_ws_uuid` (`workspaceUUID`),
  CONSTRAINT `chart_worspace_fk` FOREIGN KEY (`workspaceUUID`) REFERENCES `main_workspace` (`uuid`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;



CREATE TABLE `biz_query` (
  `id` int(11) NOT NULL AUTO_INCREMENT COMMENT '自增 ID',
  `uuid` varchar(48) NOT NULL DEFAULT '' COMMENT '全局唯一 ID qry_',
  `name` varchar(128) NOT NULL DEFAULT '' COMMENT '命名',
  `metric` varchar(256) NOT NULL DEFAULT '' COMMENT 'metric 名称',
  `query` json DEFAULT NULL COMMENT '查询条件, sql 或 json body',
  `qtype` enum('HTTP','TSQL','SQL') NOT NULL COMMENT '查询类型',
  `status` int(11) NOT NULL DEFAULT '0' COMMENT '状态 0: ok/1: 故障/2: 停用/3: 删除',
  `color` varchar(32) NOT NULL DEFAULT '' COMMENT '折线颜色代码',
  `datasource` varchar(48) DEFAULT NULL COMMENT '数据源类型',
  `chartUUID` varchar(48) NOT NULL DEFAULT '' COMMENT '关联的图表UUID',
  `workspaceUUID` varchar(48) NOT NULL DEFAULT '' COMMENT '工作空间UUID',
  `extend` text COMMENT '额外扩展字段',
  `creator` varchar(64) NOT NULL DEFAULT '' COMMENT '创建者 account-id',
  `updator` varchar(64) NOT NULL DEFAULT '' COMMENT '更新者 account-id',
  `createAt` int(11) NOT NULL DEFAULT '-1',
  `deleteAt` int(11) NOT NULL DEFAULT '-1',
  `updateAt` int(11) NOT NULL DEFAULT '-1',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_uuid` (`uuid`) COMMENT 'UUID 做成全局唯一',
  KEY `k_ws_uuid` (`workspaceUUID`),
  KEY `k_chart_uuid` (`chartUUID`) USING BTREE,
  CONSTRAINT `query_chart_fk` FOREIGN KEY (`chartUUID`) REFERENCES `biz_chart` (`uuid`),
  CONSTRAINT `query_worspace_fk` FOREIGN KEY (`workspaceUUID`) REFERENCES `main_workspace` (`uuid`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;


CREATE TABLE `biz_rule_group` (
  `id` int(11) NOT NULL AUTO_INCREMENT COMMENT '自增 ID',
  `uuid` varchar(48) NOT NULL DEFAULT '' COMMENT '全局唯一 ID, rulg_',
  `name` varchar(64) NOT NULL DEFAULT '' COMMENT '命名,用云备注',
  `workspaceUUID` varchar(48) NOT NULL DEFAULT '' COMMENT '工作空间UUID',
  `shareCode` varchar(48) NOT NULL DEFAULT '' COMMENT '分享码',
  `status` int(11) NOT NULL DEFAULT '0',
  `creator` varchar(64) NOT NULL DEFAULT '' COMMENT '创建者 account-id',
  `updator` varchar(64) NOT NULL DEFAULT '' COMMENT '更新者 account-id',
  `createAt` int(11) NOT NULL DEFAULT '-1',
  `deleteAt` int(11) NOT NULL DEFAULT '-1',
  `updateAt` int(11) NOT NULL DEFAULT '-1',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_uuid` (`uuid`) COMMENT 'UUID 做成全局唯一',
  KEY `k_ws_uuid` (`workspaceUUID`)
) ENGINE=InnoDB  DEFAULT CHARSET=utf8mb4;


CREATE TABLE `biz_rule` (
  `id` int(11) NOT NULL AUTO_INCREMENT COMMENT '自增 ID',
  `uuid` varchar(48) NOT NULL DEFAULT '' COMMENT '全局唯一 ID, rul_',
  `workspaceUUID` varchar(48) NOT NULL DEFAULT '' COMMENT '工作空间UUID',
  `ruleGroupUUID` varchar(48) NOT NULL DEFAULT '',
  `type` enum('trigger','baseline') NOT NULL DEFAULT 'trigger',
  `kapaUUID` varchar(48) NOT NULL DEFAULT '' COMMENT '所属Kapa的UUID',
  `jsonScript` json DEFAULT NULL COMMENT 'script的JSON数据',
  `tickInfo` json DEFAULT NULL COMMENT '提交后Kapa 返回的Tasks数据',
  `extend` json DEFAULT NULL COMMENT '额外配置数据',
  `status` int(11) NOT NULL DEFAULT '0' COMMENT '状态 0: ok/1: 故障/2: 停用/3: 删除',
  `creator` varchar(64) NOT NULL DEFAULT '' COMMENT '创建者 account-id',
  `updator` varchar(64) NOT NULL DEFAULT '' COMMENT '更新者 account-id',
  `createAt` int(11) NOT NULL DEFAULT '-1',
  `deleteAt` int(11) NOT NULL DEFAULT '-1',
  `updateAt` int(11) NOT NULL DEFAULT '-1',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_uuid` (`uuid`) COMMENT 'UUID 做成全局唯一',
  KEY `k_ws_uuid` (`workspaceUUID`),
  KEY `rulg_ibfk_2` (`ruleGroupUUID`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;





CREATE TABLE `biz_template` (
  `id` int(11) NOT NULL AUTO_INCREMENT COMMENT '自增 ID',
  `uuid` varchar(48) NOT NULL DEFAULT '' COMMENT '全局唯一 ID temp_',
  `owner` varchar(48) NOT NULL DEFAULT '' COMMENT '工作空间UUID/SYS',
  `name` varchar(128) NOT NULL DEFAULT '' COMMENT '命名',
  `content` json NOT NULL COMMENT '模版内容',
  `extend` json DEFAULT NULL COMMENT '额外扩展字段',
  `status` int(11) NOT NULL DEFAULT '0' COMMENT '状态 0: ok/1: 故障/2: 停用/3: 删除',
  `creator` varchar(64) NOT NULL DEFAULT '' COMMENT '创建者 account-id',
  `updator` varchar(64) NOT NULL DEFAULT '' COMMENT '更新者 account-id',
  `createAt` int(11) NOT NULL DEFAULT '-1',
  `deleteAt` int(11) NOT NULL DEFAULT '-1',
  `updateAt` int(11) NOT NULL DEFAULT '-1',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_uuid` (`uuid`) COMMENT 'UUID 做成全局唯一',
  KEY `template_owner_fk` (`owner`) USING BTREE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;


CREATE TABLE `biz_variable` (
  `id` int(11) NOT NULL AUTO_INCREMENT COMMENT '自增 ID',
  `uuid` varchar(48) NOT NULL DEFAULT '' COMMENT '全局唯一 ID,var_',
  `workspaceUUID` varchar(48) NOT NULL DEFAULT '' COMMENT '工作空间UUID',
  `dashboardUUID` varchar(48) NOT NULL DEFAULT '' COMMENT '视图全局唯一 ID',
  `name` varchar(128) NOT NULL DEFAULT '' COMMENT '变量显示名',
  `code` varchar(128) NOT NULL DEFAULT '' COMMENT '变量名',
  `type` enum('QUERY','CUSTOM_LIST','ALIYUN_INSTANCE') NOT NULL COMMENT '类型',
  `datasource` varchar(48) NOT NULL COMMENT '数据源类型',
  `definition` json DEFAULT NULL COMMENT '解说，原content内容',
  `content` json DEFAULT NULL COMMENT '变量配置数据',
  `status` int(11) NOT NULL DEFAULT '0' COMMENT '状态 0: ok/1: 故障/2: 停用/3: 删除',
  `creator` varchar(64) NOT NULL DEFAULT '' COMMENT '创建者 account-id',
  `updator` varchar(64) NOT NULL DEFAULT '' COMMENT '更新者 account-id',
  `createAt` int(11) NOT NULL DEFAULT '-1',
  `deleteAt` int(11) NOT NULL DEFAULT '-1',
  `updateAt` int(11) NOT NULL DEFAULT '-1',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_uuid` (`uuid`) COMMENT 'UUID 做成全局唯一',
  KEY `k_ws_uuid` (`workspaceUUID`),
  CONSTRAINT `variables_worspace_fk` FOREIGN KEY (`workspaceUUID`) REFERENCES `main_workspace` (`uuid`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;


CREATE TABLE `main_accesskey` (
  `id` int(11) NOT NULL AUTO_INCREMENT COMMENT '自增 ID',
  `uuid` varchar(48) NOT NULL DEFAULT '' COMMENT 'ak 唯一标识',
  `ak` varchar(32) NOT NULL COMMENT 'Access Key',
  `sk` varchar(128) NOT NULL COMMENT 'Secret Key',
  `creator` varchar(64) NOT NULL DEFAULT '' COMMENT '创建者 account-id',
  `updator` varchar(64) NOT NULL DEFAULT '' COMMENT '更新者 account-id',
  `status` int(11) NOT NULL DEFAULT '0' COMMENT '状态 0: 新建/1: 运行/2: 故障/3: 停用/4: 删除',
  `createAt` int(11) NOT NULL DEFAULT '-1' COMMENT '创建时间',
  `updateAt` int(11) NOT NULL DEFAULT '-1' COMMENT '更新时间 ',
  `deleteAt` int(11) NOT NULL DEFAULT '-1' COMMENT '删除时间',
  PRIMARY KEY (`id`) COMMENT 'sk 可以存在相同的情况',
  UNIQUE KEY `uk_ak` (`ak`) COMMENT 'AK 做成全局唯一',
  KEY `idx_ak` (`ak`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;


CREATE TABLE `main_account` (
  `id` int(11) NOT NULL AUTO_INCREMENT COMMENT '自增 ID',
  `uuid` varchar(48) NOT NULL DEFAULT '' COMMENT 'account 唯一标识',
  `name` varchar(64) NOT NULL COMMENT '名称',
  `username` varchar(128) NOT NULL DEFAULT '' COMMENT '登陆账号',
  `password` varchar(128) NOT NULL COMMENT '帐户密码',
  `email` varchar(64) NOT NULL COMMENT '邮箱',
  `mobile` varchar(128) NOT NULL COMMENT '手机号',
  `exterId` varchar(128) NOT NULL DEFAULT '' COMMENT '外部ID',
  `extend` json DEFAULT NULL COMMENT '额外信息',
  `creator` varchar(64) NOT NULL DEFAULT '' COMMENT '创建者 account-id',
  `updator` varchar(64) NOT NULL DEFAULT '' COMMENT '更新者 account-id',
  `status` int(11) NOT NULL DEFAULT '0' COMMENT '状态 0: ok/1: 故障/2: 停用/3: 删除',
  `createAt` int(11) NOT NULL DEFAULT '-1' COMMENT '创建时间',
  `updateAt` int(11) NOT NULL DEFAULT '-1' COMMENT '更新时间 ',
  `deleteAt` int(11) NOT NULL DEFAULT '-1' COMMENT '删除时间',
  PRIMARY KEY (`id`) COMMENT 'sk 可以存在相同的情况',
  UNIQUE KEY `uk_uuid` (`uuid`) COMMENT '全局唯一'
) ENGINE=InnoDB DEFAULT CHARSET=utf8;


CREATE TABLE `main_account_privilege` (
  `id` int(11) NOT NULL AUTO_INCREMENT COMMENT '自增 ID',
  `uuid` varchar(48) NOT NULL COMMENT '全局唯一 ID，带 rolpr_',
  `accountUUID` varchar(48) NOT NULL COMMENT '账号Uuid',
  `entityType` varchar(48) NOT NULL COMMENT '实体类型',
  `entityUUID` varchar(48) NOT NULL COMMENT '实体Uuid',
  `privilegeJson` json NOT NULL COMMENT '权限配置',
  `status` int(11) NOT NULL DEFAULT '0' COMMENT '状态 0: ok/1: 故障/2: 停用/3: 删除',
  `creator` varchar(64) NOT NULL DEFAULT '' COMMENT '创建者 account-id',
  `updator` varchar(64) NOT NULL DEFAULT '' COMMENT '更新者 account-id',
  `createAt` int(11) NOT NULL DEFAULT '-1',
  `deleteAt` int(11) NOT NULL DEFAULT '-1',
  `updateAt` int(11) NOT NULL DEFAULT '-1',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_uuid` (`uuid`) COMMENT 'UUID 做成全局唯一',
  KEY `accountUUID_fk` (`accountUUID`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;


CREATE TABLE `main_account_workspace` (
  `id` int(11) NOT NULL AUTO_INCREMENT COMMENT '自增 ID',
  `uuid` varchar(48) NOT NULL COMMENT '全局唯一 ID，带 wsac_',
  `accountUUID` varchar(48) NOT NULL COMMENT '帐户唯一ID',
  `workspaceUUID` varchar(64) NOT NULL COMMENT '工作空间 uuid',
  `dashboardUUID` varchar(48) DEFAULT NULL COMMENT '视图UUID-与用户绑定',
  `isAdmin` int(1) NOT NULL DEFAULT '0' COMMENT '是否为管理员',
  `status` int(11) NOT NULL DEFAULT '0' COMMENT '状态 0: ok/1: 故障/2: 停用/3: 删除',
  `creator` varchar(64) NOT NULL DEFAULT '' COMMENT '创建者 account-id',
  `updator` varchar(64) NOT NULL DEFAULT '' COMMENT '更新者 account-id',
  `createAt` int(11) NOT NULL DEFAULT '-1',
  `deleteAt` int(11) NOT NULL DEFAULT '-1',
  `updateAt` int(11) NOT NULL DEFAULT '-1',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_uuid` (`uuid`) COMMENT 'UUID 做成全局唯一',
  KEY `k_ws_uuid` (`workspaceUUID`),
  KEY `accountUUID_fk` (`accountUUID`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;


CREATE TABLE `main_agent` (
  `id` int(11) NOT NULL AUTO_INCREMENT COMMENT '自增 ID',
  `uuid` varchar(48) NOT NULL COMMENT 'ftagent的uuid,唯一id',
  `name` varchar(64) NOT NULL DEFAULT '' COMMENT 'agent 名称',
  `creator` varchar(64) NOT NULL DEFAULT '' COMMENT '创建者 account-id',
  `version` varchar(32) NOT NULL DEFAULT '""' COMMENT '当前版本号',
  `host` varchar(64) NOT NULL DEFAULT '""' COMMENT '主机IP, 默认为出口 IP',
  `port` int(11) NOT NULL DEFAULT '0',
  `domainName` varchar(128) NOT NULL DEFAULT ' ' COMMENT 'agent域名',
  `workspaceUUID` varchar(65) NOT NULL DEFAULT '-1' COMMENT '关联的工作空间uuid',
  `status` int(11) NOT NULL DEFAULT '0' COMMENT '状态 0: ok/1: 故障/2: 停用/3: 删除',
  `updator` varchar(64) NOT NULL DEFAULT '',
  `createAt` int(11) NOT NULL DEFAULT '-1' COMMENT '创建时间',
  `uploadAt` int(11) NOT NULL DEFAULT '-1' COMMENT '最后上传时间',
  `deleteAt` int(11) NOT NULL DEFAULT '-1' COMMENT '删除时间',
  `updateAt` int(11) NOT NULL DEFAULT '-1' COMMENT '最后更新时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk` (`uuid`),
  KEY `idx_ws_uuid` (`workspaceUUID`),
  KEY `idx_uuid` (`uuid`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;

CREATE TABLE `main_influx_instance` (
  `id` int(11) NOT NULL AUTO_INCREMENT COMMENT '自增 ID',
  `uuid` varchar(48) NOT NULL DEFAULT '' COMMENT '全局唯一 ID 前缀 iflx_',
  `host` varchar(128) NOT NULL COMMENT '源的配置信息',
  `instanceId` varchar(128) NOT NULL COMMENT '实例ID',
  `dbcount` int(11) NOT NULL DEFAULT '0' COMMENT '当前实例的DB总数量',
  `user` varchar(64) NOT NULL DEFAULT '',
  `pwd` varchar(64) NOT NULL DEFAULT '',
  `status` int(11) NOT NULL DEFAULT '0' COMMENT '状态 0: ok/1: 故障/2: 停用/3: 删除',
  `creator` varchar(64) NOT NULL DEFAULT '' COMMENT '创建者 account-id',
  `updator` varchar(64) NOT NULL DEFAULT '' COMMENT '更新者 account-id',
  `createAt` int(11) NOT NULL DEFAULT '-1',
  `deleteAt` int(11) NOT NULL DEFAULT '-1',
  `updateAt` int(11) NOT NULL DEFAULT '-1',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_uuid` (`uuid`) COMMENT 'UUID 做成全局唯一'
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;


CREATE TABLE `main_influx_db` (
  `id` int(11) NOT NULL AUTO_INCREMENT COMMENT '自增 ID',
  `uuid` varchar(48) NOT NULL DEFAULT '' COMMENT '全局唯一 ID 前缀 ifdb_',
  `db` varchar(48) NOT NULL DEFAULT '' COMMENT 'DB 名称',
  `influxInstanceUUID` varchar(48) NOT NULL DEFAULT '' COMMENT 'instance的UUID',
  `influxRpUUID` varchar(48) NOT NULL DEFAULT '' COMMENT 'influx rp uuid',
  `influxRpName` varchar(48) NOT NULL DEFAULT '' COMMENT 'influx dbrp name',
  `status` int(11) NOT NULL DEFAULT '0' COMMENT '状态 0: ok/1: 故障/2: 停用/3: 删除',
  `creator` varchar(64) NOT NULL DEFAULT '' COMMENT '创建者 account-id',
  `updator` varchar(64) NOT NULL DEFAULT '' COMMENT '更新者 account-id',
  `createAt` int(11) NOT NULL DEFAULT '-1',
  `deleteAt` int(11) NOT NULL DEFAULT '-1',
  `updateAt` int(11) NOT NULL DEFAULT '-1',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_uuid` (`uuid`) COMMENT 'UUID 做成全局唯一',
  UNIQUE KEY `dbrp` (`db`,`influxInstanceUUID`,`influxRpName`),
  KEY `db_isuuid` (`influxInstanceUUID`),
  KEY `db_rpuuid` (`influxRpUUID`),
  CONSTRAINT `db_isuuid` FOREIGN KEY (`influxInstanceUUID`) REFERENCES `main_influx_instance` (`uuid`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;


CREATE TABLE `main_influx_rp` (
  `id` int(11) NOT NULL AUTO_INCREMENT COMMENT '自增 ID',
  `uuid` varchar(48) NOT NULL DEFAULT '' COMMENT '全局唯一 ID 前缀ifrp_',
  `name` varchar(48) NOT NULL DEFAULT '' COMMENT 'rp名称',
  `duration` varchar(48) NOT NULL DEFAULT '0' COMMENT 'InfluxDB保留数据的时间，此处单位为天数(d)',
  `shardGroupDuration` varchar(48) NOT NULL DEFAULT '0' COMMENT 'optional, 此处单位为小时(h)',
  `replication` int(11) NOT NULL DEFAULT '1' COMMENT '每个点的多少独立副本存储在集群中，其中n是数据节点的数量。该子句不能用于单节点实例',
  `status` int(11) NOT NULL DEFAULT '0' COMMENT '状态 0: ok/1: 故障/2: 停用/3: 删除',
  `creator` varchar(64) NOT NULL DEFAULT '' COMMENT '创建者 account-id',
  `updator` varchar(64) NOT NULL DEFAULT '' COMMENT '更新者 account-id',
  `createAt` int(11) NOT NULL DEFAULT '-1',
  `updateAt` int(11) NOT NULL DEFAULT '-1',
  `deleteAt` int(11) NOT NULL DEFAULT '-1',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_uuid` (`uuid`) COMMENT 'UUID 做成全局唯一'
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;


CREATE TABLE `main_inner_app` (
  `id` int(11) NOT NULL AUTO_INCREMENT COMMENT '自增 ID',
  `uuid` varchar(48) NOT NULL COMMENT '全局唯一 ID，关联用，带 inap_ 前缀',
  `akUUID` varchar(64) NOT NULL COMMENT 'ak uuid',
  `version` varchar(64) NOT NULL COMMENT '版本号',
  `domain` varchar(64) NOT NULL COMMENT '域名',
  `name` varchar(64) NOT NULL DEFAULT '' COMMENT '名称',
  `status` int(11) NOT NULL DEFAULT '0' COMMENT '状态 0: ok/1: 故障/2: 停用/3: 删除',
  `creator` varchar(64) NOT NULL DEFAULT '' COMMENT '创建者',
  `updator` varchar(64) NOT NULL DEFAULT '' COMMENT '更新者 account-id',
  `createAt` int(11) NOT NULL DEFAULT '-1',
  `deleteAt` int(11) NOT NULL DEFAULT '-1',
  `updateAt` int(11) NOT NULL DEFAULT '-1',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_uuid` (`uuid`) COMMENT 'UUID 做成全局唯一'
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;


CREATE TABLE `main_kapa` (
  `id` int(11) NOT NULL AUTO_INCREMENT COMMENT '自增 ID',
  `uuid` varchar(48) NOT NULL DEFAULT '' COMMENT '全局唯一 ID 前缀 kapa_',
  `host` varchar(128) NOT NULL COMMENT 'kapa 地址',
  `influxInstanceUUID` varchar(48) NOT NULL DEFAULT '' COMMENT 'source实例uuid',
  `status` int(11) NOT NULL DEFAULT '0' COMMENT '状态 0: ok/1: 故障/2: 停用/3: 删除',
  `creator` varchar(64) NOT NULL DEFAULT '' COMMENT '创建者 account-id',
  `updator` varchar(64) NOT NULL DEFAULT '' COMMENT '更新者 account-id',
  `createAt` int(11) NOT NULL DEFAULT '-1',
  `deleteAt` int(11) NOT NULL DEFAULT '-1',
  `updateAt` int(11) NOT NULL DEFAULT '-1',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_uuid` (`uuid`) COMMENT 'UUID 做成全局唯一',
  KEY `kapa_is_uuid` (`influxInstanceUUID`),
  CONSTRAINT `kapa_is_uuid` FOREIGN KEY (`influxInstanceUUID`) REFERENCES `main_influx_instance` (`uuid`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;


CREATE TABLE `main_manage_account` (
  `id` int(11) NOT NULL AUTO_INCREMENT COMMENT '自增 ID',
  `uuid` varchar(48) NOT NULL DEFAULT '' COMMENT 'account 唯一标识',
  `name` varchar(64) NOT NULL COMMENT '昵称',
  `username` varchar(128) NOT NULL DEFAULT '' COMMENT '账户名',
  `password` varchar(128) NOT NULL COMMENT '帐户密码',
  `email` varchar(64) NOT NULL COMMENT '邮箱',
  `mobile` varchar(128) NOT NULL COMMENT '手机号',
  `creator` varchar(64) NOT NULL DEFAULT '' COMMENT '创建者 account-id',
  `updator` varchar(64) NOT NULL DEFAULT '' COMMENT '更新者 account-id',
  `status` int(11) NOT NULL DEFAULT '0' COMMENT '状态 0: ok/1: 故障/2: 停用/3: 删除',
  `createAt` int(11) NOT NULL DEFAULT '-1' COMMENT '创建时间',
  `updateAt` int(11) NOT NULL DEFAULT '-1' COMMENT '更新时间 ',
  `deleteAt` int(11) NOT NULL DEFAULT '-1' COMMENT '删除时间',
  PRIMARY KEY (`id`) COMMENT 'sk 可以存在相同的情况',
  UNIQUE KEY `uk_uuid` (`uuid`) COMMENT '全局唯一'
) ENGINE=InnoDB AUTO_INCREMENT=26 DEFAULT CHARSET=utf8;


CREATE TABLE `main_manage_role` (
  `id` int(11) NOT NULL AUTO_INCREMENT COMMENT '自增 ID',
  `uuid` varchar(48) NOT NULL COMMENT '全局唯一 ID，带 role_',
  `name` varchar(128) NOT NULL COMMENT '命名',
  `type` enum('super_admin','admin','editor','selector') NOT NULL COMMENT '角色 类型',
  `describe` text COMMENT '角色描述',
  `status` int(11) NOT NULL DEFAULT '0' COMMENT '状态 0: ok/1: 故障/2: 停用/3: 删除',
  `creator` varchar(64) NOT NULL DEFAULT '' COMMENT '创建者 account-id',
  `updator` varchar(64) NOT NULL DEFAULT '' COMMENT '更新者 account-id',
  `createAt` int(11) NOT NULL DEFAULT '-1',
  `deleteAt` int(11) NOT NULL DEFAULT '-1',
  `updateAt` int(11) NOT NULL DEFAULT '-1',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_uuid` (`uuid`) COMMENT 'UUID 做成全局唯一'
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

--- 11月5日
CREATE TABLE `main_subscription` (
  `id` int(11) NOT NULL AUTO_INCREMENT COMMENT '自增 ID',
  `uuid` varchar(48) NOT NULL DEFAULT '' COMMENT '全局唯一 ID 前缀 sbsp_',
  `dbUUID` varchar(48) NOT NULL DEFAULT '' COMMENT 'DB uuid',
  `name` varchar(64) NOT NULL DEFAULT '' COMMENT '订阅名称',
  `kapaUUID` varchar(48) NOT NULL DEFAULT '' COMMENT 'kapa的UUID',
  `status` int(11) NOT NULL DEFAULT '0' COMMENT '状态 0: ok/1: 故障/2: 停用/3: 删除',
  `creator` varchar(64) NOT NULL DEFAULT '' COMMENT '创建者 account-id',
  `updator` varchar(64) NOT NULL DEFAULT '' COMMENT '更新者 account-id',
  `createAt` int(11) NOT NULL DEFAULT '-1',
  `deleteAt` int(11) NOT NULL DEFAULT '-1',
  `updateAt` int(11) NOT NULL DEFAULT '-1',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_uuid` (`uuid`) COMMENT 'UUID 做成全局唯一',
  KEY `db_isuuid` (`dbUUID`),
  KEY `kapa_uuid` (`kapaUUID`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

 -- multiple rp add at 2020-01-15
ALTER TABLE main_subscription ADD COLUMN rpName VARCHAR(48) NOT NULL AFTER dbUUID
 -- data update
UPDATE main_subscription as b SET b.rpName=(SELECT a.influxRpName FROM main_influx_db AS a WHERE a.uuid=b.dbUUID)

-- cq 2020/0212
CREATE TABLE IF NOT EXISTS `main_influx_cq` (
		`id` int(11) NOT NULL AUTO_INCREMENT COMMENT '自增 ID',
		`uuid` varchar(48) NOT NULL DEFAULT '' COMMENT '全局唯一 ID 前缀 ifcq-',
		`cq_name` varchar(48) NOT NULL DEFAULT 'cq_', COMMENT 'CQ measurement name prefix'
		`sample_every` varchar(48) NOT NULL DEFAULT '30m' COMMENT 'RESAMPLE every XXX',
		`sample_for` varchar(48) NOT NULL DEFAULT '1h' COMMENT 'RESAMPLE for XXX',
		`db` varchar(128) NOT NULL DEFAULT '' COMMENT '原始 Measurement 所在 DB 名称',
		`rp` varchar(128) NOT NULL DEFAULT '' COMMENT '不填则只用 @db 的默认 RP',
		`cq_rp` varchar(48) NOT NULL DEFAULT '' COMMENT '不填则用 CQ 默认的 RP, 比如 rpcq',
		`measurement` varchar(48) NOT NULL DEFAULT '' COMMENT '指标名',
		`cq_meas_name` varchar(48) NOT NULL DEFAULT '' COMMENT 'cq 后的指标名，，不指定则  @name',
		`keep_tags` tinyint(1) NOT NULL DEFAULT '1' COMMENT '是否在 CQ 的 measurement 上保 原始 tag， 然原始 tag 会被转换成 field',
		`group_by_time` varchar(48) NOT NULL DEFAULT '30m',
		`group_by_offset` varchar(48) NOT NULL DEFAULT '15m',
		`func_fields`  json NOT NULL COMMENT '["""mean("field-1") AS "field-1-mean"""",...]',
		`creator` varchar(64) NOT NULL DEFAULT '' COMMENT '创建者 account-id',
		`updator` varchar(64) NOT NULL DEFAULT '' COMMENT '更新者 account-id',
		`createAt` int(11) NOT NULL DEFAULT '-1',
		`deleteAt` int(11) NOT NULL DEFAULT '-1',
		`updateAt` int(11) NOT NULL DEFAULT '-1',

		PRIMARY KEY (`id`),
		UNIQUE KEY `uk_uuid` (`uuid`) COMMENT 'UUID 做成全局唯一'
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

