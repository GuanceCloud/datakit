/*
 * BSS
 *
 * Business Support System API
 *
 */

package model

import (
	"encoding/json"

	"strings"
)

type ApplyEnterpriseRealnameAuthsReq struct {
	// |参数名称：客户ID。| |参数约束及描述：客户ID。|
	CustomerId string `json:"customer_id"`
	// |参数名称：认证方案。1：企业证件扫描| |参数的约束及描述：认证方案。1：企业证件扫描|
	IdentifyType int32 `json:"identify_type"`
	// |参数名称：企业证件类型：0：企业营业执照1：事业单位法人证书2：社会团体法人登记证书3：行政执法主体资格证4：组织机构代码证99：其他| |参数的约束及描述：企业证件类型：0：企业营业执照1：事业单位法人证书2：社会团体法人登记证书3：行政执法主体资格证4：组织机构代码证99：其他|
	CertificateType *int32 `json:"certificate_type,omitempty"`
	// |参数名称：企业证件认证时证件附件的文件URL。附件地址必须按照顺序填写，先填写企业证件的附件，如果请求中填写了企业人员信息，再填写企业人员的身份证附件。企业证件顺序为：第1张企业证件照附件，企业人员的证件顺序为：第1张个人身份证的人像面第2张个人身份证的国徽面以营业执照举例，假设存在法人的情况下，第1张上传的是营业执照扫描件abc.023，第2张是法人的身份证人像面照片def004，第3张是法人的国徽面照片gh007，那么上传顺序需要是：abc023def004gh007文件名称区分大小写附件地址必须按照顺序填写，先填写企业证件的附件，如果请求中填写了企业人员信息，再填写企业人员的身份证附件。企业证件顺序为：第1张企业证件照正面，第2张企业证件照反面，个人证件顺序为：第1张个人身份证的人像面第2张个人身份证的国徽面假设不存在法人的情况下，第1张上传的是企业证件正面扫描件abc.023，第2张上传的是企业证件反面扫描件def004，那么上传顺序需要是：abc023def004文件名称区分大小写证件附件目前仅仅支持jpg、jpeg、bmp、png、gif、pdf格式，单个文件最大不超过10M。这个URL是相对URL，不需要包含桶名和download目录，只要包含download目录下的子目录和对应文件名称即可。举例如下：如果上传的证件附件在桶中的位置是：https://bucketname.obs.Endpoint.myhuaweicloud.com/download/abc023.jpg，该字段填写abc023.jpg；如果上传的证件附件在桶中的位置是：https://bucketname.obs.Endpoint.myhuaweicloud.com/download/test/abc023.jpg，该字段填写test/abc023.jpg。| |参数约束以及描述：企业证件认证时证件附件的文件URL。附件地址必须按照顺序填写，先填写企业证件的附件，如果请求中填写了企业人员信息，再填写企业人员的身份证附件。企业证件顺序为：第1张企业证件照附件，企业人员的证件顺序为：第1张个人身份证的人像面第2张个人身份证的国徽面以营业执照举例，假设存在法人的情况下，第1张上传的是营业执照扫描件abc.023，第2张是法人的身份证人像面照片def004，第3张是法人的国徽面照片gh007，那么上传顺序需要是：abc023def004gh007文件名称区分大小写附件地址必须按照顺序填写，先填写企业证件的附件，如果请求中填写了企业人员信息，再填写企业人员的身份证附件。企业证件顺序为：第1张企业证件照正面，第2张企业证件照反面，个人证件顺序为：第1张个人身份证的人像面第2张个人身份证的国徽面假设不存在法人的情况下，第1张上传的是企业证件正面扫描件abc.023，第2张上传的是企业证件反面扫描件def004，那么上传顺序需要是：abc023def004文件名称区分大小写证件附件目前仅仅支持jpg、jpeg、bmp、png、gif、pdf格式，单个文件最大不超过10M。这个URL是相对URL，不需要包含桶名和download目录，只要包含download目录下的子目录和对应文件名称即可。举例如下：如果上传的证件附件在桶中的位置是：https://bucketname.obs.Endpoint.myhuaweicloud.com/download/abc023.jpg，该字段填写abc023.jpg；如果上传的证件附件在桶中的位置是：https://bucketname.obs.Endpoint.myhuaweicloud.com/download/test/abc023.jpg，该字段填写test/abc023.jpg。|
	VerifiedFileUrl []string `json:"verified_file_url"`
	// |参数名称：单位名称。不能全是数字、特殊字符、空格。| |参数约束及描述：单位名称。不能全是数字、特殊字符、空格。|
	CorpName string `json:"corp_name"`
	// |参数名称：单位证件号码。| |参数约束及描述：单位证件号码。|
	VerifiedNumber string `json:"verified_number"`
	// |参数名称：实名认证填写的注册国家。国家的两位字母简码。例如：注册国家为“中国”请填写“CN”。| |参数约束及描述：实名认证填写的注册国家。国家的两位字母简码。例如：注册国家为“中国”请填写“CN”。|
	RegCountry *string `json:"reg_country,omitempty"`
	// |参数名称：实名认证企业注册地址。| |参数约束及描述：实名认证企业注册地址。|
	RegAddress *string `json:"reg_address,omitempty"`
	// |参数名称：华为分给合作伙伴的平台标识。该标识的具体值由华为分配。获取方法请参见如何获取xaccountType的取值| |参数约束及描述：华为分给合作伙伴的平台标识。该标识的具体值由华为分配。获取方法请参见如何获取xaccountType的取值|
	XaccountType     string               `json:"xaccount_type"`
	EnterprisePerson *EnterprisePersonNew `json:"enterprise_person,omitempty"`
}

func (o ApplyEnterpriseRealnameAuthsReq) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "ApplyEnterpriseRealnameAuthsReq struct{}"
	}

	return strings.Join([]string{"ApplyEnterpriseRealnameAuthsReq", string(data)}, " ")
}
