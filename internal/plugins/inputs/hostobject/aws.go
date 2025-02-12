// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package hostobject

import "time"

const (
	AWSAuthHeader      = "X-aws-ec2-metadata-token"
	AWSTTLHeader       = "X-aws-ec2-metadata-token-ttl-seconds"
	AWSDefaultTokenURL = "http://169.254.169.254/latest/api/token" // nolint:gosec
	AWSMaxTokenTTL     = 21600 * time.Second
)

type aws struct {
	baseURL    string // http://100.100.100.200/latest/meta-data
	authConfig AuthConfig
}

func defaultAWSAuthConfig(ipt *Input) AuthConfig {
	authConfig := AuthConfig{
		Enable: ipt.EnableCloudAWSIMDSv2,
	}
	if ipt.EnableCloudAWSIMDSv2 {
		authConfig.AuthHeader = AWSAuthHeader
		authConfig.TTLHeader = AWSTTLHeader
		authConfig.TokenURL = AWSDefaultTokenURL
		authConfig.MaxTokenTTL = AWSMaxTokenTTL
		authConfig.TokenTTL = ipt.Interval // 这里暂时将token的ttl设置为采集器间隔时间

		if url, ok := ipt.CloudMetaTokenURL[AWS]; ok {
			authConfig.TokenURL = url
		}
	}

	return authConfig
}

func (x *aws) Sync() (map[string]interface{}, error) {
	return map[string]interface{}{
		"cloud_provider":        "aws",
		"description":           x.Description(),
		"instance_id":           x.InstanceID(),
		"instance_name":         x.InstanceName(),
		"instance_type":         x.InstanceType(),
		"instance_charge_type":  x.InstanceChargeType(),
		"instance_network_type": x.InstanceNetworkType(),
		"instance_status":       x.InstanceStatus(),
		"security_group_id":     x.SecurityGroupID(),
		"private_ip":            x.PrivateIP(),
		"zone_id":               x.ZoneID(),
		"region":                x.Region(),
	}, nil
}

func (x *aws) Description() string {
	return Unavailable
}

func (x *aws) InstanceID() string {
	return metaGetV2(x.baseURL+"/instance-id", x.authConfig)
}

func (x *aws) InstanceName() string {
	return Unavailable
}

func (x *aws) InstanceType() string {
	return metaGetV2(x.baseURL+"/instance-type", x.authConfig)
}

func (x *aws) InstanceChargeType() string {
	return Unavailable
}

func (x *aws) InstanceNetworkType() string {
	return Unavailable
}

func (x *aws) InstanceStatus() string {
	return Unavailable
}

func (x *aws) SecurityGroupID() string {
	return metaGetV2(x.baseURL+"/security-groups", x.authConfig)
}

func (x *aws) PrivateIP() string {
	return metaGetV2(x.baseURL+"/local-ipv4", x.authConfig)
}

func (x *aws) ZoneID() string {
	return metaGetV2(x.baseURL+"/placement/availability-zone-id", x.authConfig)
}

func (x *aws) Region() string {
	// 这个在 AWS 文档是没有的：
	//  https://docs.aws.amazon.com/AWSEC2/latest/UserGuide/instancedata-data-retrieval.html
	return metaGetV2(x.baseURL+"/placement/region", x.authConfig)
}
