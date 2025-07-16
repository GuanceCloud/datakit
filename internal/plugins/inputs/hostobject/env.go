// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package hostobject

import (
	"encoding/json"
	"os"
	"strconv"
	"strings"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/config"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/dkstring"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/export/doc"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs"
)

func (ipt *Input) GetENVDoc() []*inputs.ENVInfo {
	infos := []*inputs.ENVInfo{
		{
			FieldName: "EnableNetVirtualInterfaces",
			ENVName:   "ENABLE_NET_VIRTUAL_INTERFACES",
			ConfField: "enable_net_virtual_interfaces",
			Type:      doc.Boolean,
			Default:   `false`,
			Desc:      "Enable collect network virtual interfaces",
			DescZh:    "允许采集虚拟网卡",
		},

		{
			FieldName: "IgnoreZeroBytesDisk",
			ENVName:   "IGNORE_ZERO_BYTES_DISK",
			ConfField: "ignore_zero_bytes_disk",
			Type:      doc.Boolean,
			Default:   `false`,
			Desc:      "Ignore the disk which space is zero",
			DescZh:    "忽略大小为 0 的磁盘",
		},

		{
			FieldName: "IgnoreFSTypes",
			ENVName:   "IGNORE_FSTYPES",
			ConfField: "ignore_fstypes",
			Type:      doc.String,
			Default:   "`^(tmpfs|autofs|binfmt_misc|devpts|fuse.lxcfs|overlay|proc|squashfs|sysfs)$`",
			Desc:      "Ignore disks with these file systems",
			DescZh:    "磁盘列表采集时忽略特定的文件系统",
		},

		{
			FieldName: "IgnoreMountpoints",
			ENVName:   "IGNORE_MOUNTPOINTS",
			ConfField: "ignore_mountpoints",
			Type:      doc.String,
			Default:   "`^(/usr/local/datakit/.*|/run/containerd/.*)$`",
			Desc:      "Ignore disks with these mount points",
			DescZh:    "磁盘列表采集时忽略特定的挂载点",
		},

		{
			FieldName: "ExcludeDevice",
			ENVName:   "EXCLUDE_DEVICE",
			ConfField: "exclude_device",
			Type:      doc.List,
			Example:   "`/dev/loop0,/dev/loop1`",
			Desc:      "Exclude some with dev prefix",
			DescZh:    "忽略的 device",
		},

		{
			FieldName: "ExtraDevice",
			ENVName:   "EXTRA_DEVICE",
			ConfField: "extra_device",
			Type:      doc.List,
			Example:   "`/nfsdata,other`",
			Desc:      "Additional device",
			DescZh:    "额外增加的 device",
		},

		{
			FieldName: "EnableCloudHostTagsGlobalElection",
			ENVName:   "CLOUD_META_AS_ELECTION_TAGS",
			ConfField: "enable_cloud_host_tags_global_election_tags",
			Type:      doc.Boolean,
			Default:   "true",
			Desc:      "Enable put cloud provider region/zone_id information into global election tags",
			DescZh:    "将云服务商 region/zone_id 信息放入全局选举标签",
		},

		{
			FieldName: "EnableCloudHostTagsGlobalHost",
			ENVName:   "CLOUD_META_AS_HOST_TAGS",
			ConfField: "enable_cloud_host_tags_global_host_tags",
			Type:      doc.Boolean,
			Default:   "true",
			Desc:      "Enable put cloud provider region/zone_id information into global host tags",
			DescZh:    "将云服务商 region/zone_id 信息放入全局主机标签",
		},

		{
			FieldName: "EnableCloudAWSIMDSv2",
			ENVName:   "CLOUD_AWS_IMDS_V2",
			ConfField: "enable_cloud_aws_imds_v2",
			Type:      doc.Boolean,
			Default:   "false",
			Desc:      "Enable AWS IMDSv2",
			DescZh:    "开启 AWS IMDSv2",
		},

		{
			FieldName: "EnableCloudAWSIPv6",
			ENVName:   "CLOUD_AWS_IPV6",
			ConfField: "enable_cloud_aws_ipv6",
			Type:      doc.Boolean,
			Default:   "false",
			Desc:      "Enable AWS IPv6",
			DescZh:    "开启 AWS IPv6",
		},

		{
			FieldName: "Tags",
			ENVName:   "TAGS",
			ConfField: "tags",
		},

		{
			FieldName: "ENVCloud",
			ENVName:   "CLOUD_PROVIDER",
			ConfField: "none",
			Type:      doc.String,
			Example:   "`aliyun/aws/tencent/hwcloud/azure`",
			Desc:      "Designate cloud service provider",
			DescZh:    "指定云服务商",
		},

		{
			FieldName: "CloudMetaURL",
			ENVName:   "CLOUD_META_URL",
			ConfField: "cloud_meta_url",
			Type:      doc.Map,
			Example:   "`{\"tencent\":\"xxx\", \"aliyun\":\"yyy\"}`",
			Desc:      "Cloud metadata URL mapping",
			DescZh:    "云服务商元数据 URL 映射",
		},

		{
			FieldName: "CloudMetaTokenURL",
			ENVName:   "CLOUD_META_TOKEN_URL",
			ConfField: "cloud_meta_token_url",
			Type:      doc.Map,
			Example:   "`{\"aws\":\"xxx\",\"aliyun\":\"yyy\"}`",
			Desc:      "Cloud metadata Token URL mapping",
			DescZh:    "云服务商获取元数据的 Token URL 映射",
		},

		{
			FieldName: "DisableCloudProviderSync",
			ENVName:   "DISABLE_CLOUD_PROVIDER_SYNC",
			ConfField: "disable_cloud_provider_sync",
			Type:      doc.Boolean,
			Example:   "`true`",
			Desc:      "Disable cloud metadata",
			DescZh:    "禁止同步主机云信息",
		},

		{
			ENVName:   "USE_NSENTER",
			ConfField: "use_nsenter",
			Type:      doc.Boolean,
			Example:   "`true`",
			Desc:      "Use nsenter to collect disk usage",
			DescZh:    "用 `nsenter` 方式来采集磁盘用量信息",
		},
	}

	return doc.SetENVDoc("ENV_INPUT_HOSTOBJECT_", infos)
}

// ReadEnv used to read ENVs while running under DaemonSet.
func (ipt *Input) ReadEnv(envs map[string]string) {
	if enable, ok := envs["ENV_INPUT_HOSTOBJECT_ENABLE_NET_VIRTUAL_INTERFACES"]; ok {
		b, err := strconv.ParseBool(enable)
		if err != nil {
			l.Warnf("parse ENV_INPUT_HOSTOBJECT_ENABLE_NET_VIRTUAL_INTERFACES to bool: %s, ignore", err)
		} else {
			ipt.EnableNetVirtualInterfaces = b
		}
	}

	if fsList, ok := envs["ENV_INPUT_HOSTOBJECT_EXTRA_DEVICE"]; ok {
		list := strings.Split(fsList, ",")
		l.Debugf("add extra_device from ENV: %v", fsList)
		ipt.ExtraDevice = append(ipt.ExtraDevice, list...)
	}

	if fsList, ok := envs["ENV_INPUT_HOSTOBJECT_EXCLUDE_DEVICE"]; ok {
		list := strings.Split(fsList, ",")
		l.Debugf("add exlude_device from ENV: %v", fsList)
		ipt.ExcludeDevice = append(ipt.ExcludeDevice, list...)
	}

	// https://gitlab.jiagouyun.com/cloudcare-tools/datakit/-/issues/505
	if enable, ok := envs["ENV_INPUT_HOSTOBJECT_ENABLE_ZERO_BYTES_DISK"]; ok {
		b, err := strconv.ParseBool(enable)
		if err != nil {
			l.Warnf("parse ENV_INPUT_HOSTOBJECT_ENABLE_ZERO_BYTES_DISK to bool: %s, ignore", err)
		} else {
			ipt.IgnoreZeroBytesDisk = b
		}
	}

	// https://gitlab.jiagouyun.com/cloudcare-tools/datakit/-/issues/2076
	if enable, ok := envs["ENV_INPUT_HOSTOBJECT_CLOUD_META_AS_ELECTION_TAGS"]; ok {
		b, err := strconv.ParseBool(enable)
		if err != nil {
			l.Warnf("parse ENV_INPUT_HOSTOBJECT_CLOUD_META_AS_ELECTION_TAGS to bool: %s, ignore", err)
		} else {
			ipt.EnableCloudHostTagsGlobalElection = b
		}
	}
	// https://gitlab.jiagouyun.com/cloudcare-tools/datakit/-/issues/2136
	if enable, ok := envs["ENV_INPUT_HOSTOBJECT_CLOUD_META_AS_HOST_TAGS"]; ok {
		b, err := strconv.ParseBool(enable)
		if err != nil {
			l.Warnf("parse ENV_INPUT_HOSTOBJECT_CLOUD_META_AS_HOST_TAGS to bool: %s, ignore", err)
		} else {
			ipt.EnableCloudHostTagsGlobalHost = b
		}
	}
	if tagsStr, ok := envs["ENV_INPUT_HOSTOBJECT_TAGS"]; ok {
		tags := config.ParseGlobalTags(tagsStr)
		for k, v := range tags {
			ipt.Tags[k] = v
		}
	}

	if enable, ok := envs["ENV_INPUT_HOSTOBJECT_CLOUD_AWS_IMDS_V2"]; ok {
		b, err := strconv.ParseBool(enable)
		if err != nil {
			l.Warnf("parse ENV_INPUT_HOSTOBJECT_CLOUD_AWS_IMDS_V2 to bool: %s, ignore", err)
		} else {
			ipt.EnableCloudAWSIMDSv2 = b
		}
	}

	if enable, ok := envs["ENV_INPUT_HOSTOBJECT_CLOUD_AWS_IPV6"]; ok {
		b, err := strconv.ParseBool(enable)
		if err != nil {
			l.Warnf("parse ENV_INPUT_HOSTOBJECT_CLOUD_AWS_IPV6 to bool: %s, ignore", err)
		} else {
			ipt.EnableCloudAWSIPv6 = b
		}
	}

	// ENV_CLOUD_PROVIDER 会覆盖 ENV_INPUT_HOSTOBJECT_TAGS 中填入的 cloud_provider
	if tagsStr, ok := envs["ENV_CLOUD_PROVIDER"]; ok {
		cloudProvider := dkstring.TrimString(tagsStr)
		cloudProvider = strings.ToLower(cloudProvider)
		switch cloudProvider {
		case "aliyun", "tencent", "aws", "hwcloud", "azure":
			ipt.Tags["cloud_provider"] = cloudProvider
		}
	} // ENV_CLOUD_PROVIDER

	if cloudMetaURLStr, ok := envs["ENV_INPUT_HOSTOBJECT_CLOUD_META_URL"]; ok {
		var cloudMetaURL map[string]string
		err := json.Unmarshal([]byte(cloudMetaURLStr), &cloudMetaURL)
		if err != nil {
			l.Warnf("parse ENV_INPUT_HOSTOBJECT_CLOUD_META_URL: %s, ignore", err)
		} else {
			ipt.CloudMetaURL = cloudMetaURL
			l.Infof("loaded cloud_meta_url from ENV: %v", cloudMetaURL)
		}
	}

	if cloudMetaTokenURLStr, ok := envs["ENV_INPUT_HOSTOBJECT_CLOUD_META_TOKEN_URL"]; ok {
		var cloudMetaTokenURL map[string]string
		err := json.Unmarshal([]byte(cloudMetaTokenURLStr), &cloudMetaTokenURL)
		if err != nil {
			l.Warnf("parse ENV_INPUT_HOSTOBJECT_CLOUD_META_TOKEN_URL: %s, ignore", err)
		} else {
			ipt.CloudMetaTokenURL = cloudMetaTokenURL
			l.Infof("loaded cloud_meta_token_url from ENV: %v", cloudMetaTokenURL)
		}
	}

	if s, ok := envs["ENV_INPUT_HOSTOBJECT_DISABLE_CLOUD_PROVIDER_SYNC"]; ok {
		if disabled, err := strconv.ParseBool(s); disabled {
			l.Info("cloud sync disabled")
			ipt.DisableCloudProviderSync = true
		} else {
			l.Debugf("strconv.ParseBool: %s, ignored", err)
		}
	}

	// Default setting: we have add the env HOST_ROOT in datakit.yaml by default
	// but some old deployments may not hava this ENV set.
	ipt.hostRoot = "/rootfs"

	// Deprecated: use ENV_HOST_ROOT
	if v := os.Getenv("HOST_ROOT"); v != "" {
		ipt.hostRoot = v
	}

	if v := os.Getenv("ENV_HOST_ROOT"); v != "" {
		ipt.hostRoot = v
	}

	if str := envs["ENV_INPUT_HOSTOBJECT_IGNORE_FSTYPES"]; str != "" {
		ipt.IgnoreFSTypes = str
	}

	if str := envs["ENV_INPUT_HOSTOBJECT_IGNORE_MOUNTPOINTS"]; str != "" {
		ipt.IgnoreMountpoints = str
	}

	if v := os.Getenv("ENV_INPUT_HOSTOBJECT_USE_NSENTER"); v != "" {
		if b, _ := strconv.ParseBool(v); b {
			ipt.UseNSEnterDiskstatsImpl = true
		}
	}
}
