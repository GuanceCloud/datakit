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

type ApplyIndividualRealnameAuthsReq struct {
	// |参数名称：客户ID。| |参数约束及描述：客户ID。|
	CustomerId string `json:"customer_id"`
	// |参数名称：认证方案：0：个人证件认证4：个人银行卡认证。这种方式下，仅仅需要上传一张个人扫脸的图片附件即可。| |参数的约束及描述：认证方案：0：个人证件认证4：个人银行卡认证。这种方式下，仅仅需要上传一张个人扫脸的图片附件即可。|
	IdentifyType int32 `json:"identify_type"`
	// |参数名称：证件类型：0：身份证，上传的附件为3张，第1张是身份证人像面，第2张是身份证国徽面，第3张是个人手持身份证人像面；3：护照，上传的附件为3张，第1张是护照个人资料页，第2张是，护照入境盖章页，第3张是手持护照个人资料页；3：护照，上传的附件为2张，第1张是护照个人资料页，第2张是手持护照个人资料页；5：港澳通行证，上传的附件为3张，第1张是港澳居民来往内地通行证正面（人像面），第2张是港澳居民来往内地通行证反面，第3张是手持港澳居民来往内地通行证人像面；6：台湾通行证，上传的附件为3张，第1张是台湾居民来往大陆通行证正面（人像面），第2张是台湾居民来往大陆通行证反面，第3张是手持台湾居民来往大陆通行证人像面；7：海外驾照，上传的附件为2张，第1张是中国以外驾照正面照片（人像面），第2张是手持中国以外驾照人像面照片；9：港澳居民居住证，上传的附件为3张，第1张是港澳居民居住证人像面，第2张是，港澳居民居住证国徽面，第3张是手持港澳居民居住证人像面照片；10：台湾居民居住证，上传的附件为3张，第1张是台湾居民居住证人像面，第2张是台湾居民居住证国徽面，第3张是手持台湾居民居住证人像面照片。当identifyType=0的时候，该字段需要填写，否则忽略该字段的取值。| |参数的约束及描述：证件类型：0：身份证，上传的附件为3张，第1张是身份证人像面，第2张是身份证国徽面，第3张是个人手持身份证人像面；3：护照，上传的附件为3张，第1张是护照个人资料页，第2张是，护照入境盖章页，第3张是手持护照个人资料页；3：护照，上传的附件为2张，第1张是护照个人资料页，第2张是手持护照个人资料页；5：港澳通行证，上传的附件为3张，第1张是港澳居民来往内地通行证正面（人像面），第2张是港澳居民来往内地通行证反面，第3张是手持港澳居民来往内地通行证人像面；6：台湾通行证，上传的附件为3张，第1张是台湾居民来往大陆通行证正面（人像面），第2张是台湾居民来往大陆通行证反面，第3张是手持台湾居民来往大陆通行证人像面；7：海外驾照，上传的附件为2张，第1张是中国以外驾照正面照片（人像面），第2张是手持中国以外驾照人像面照片；9：港澳居民居住证，上传的附件为3张，第1张是港澳居民居住证人像面，第2张是，港澳居民居住证国徽面，第3张是手持港澳居民居住证人像面照片；10：台湾居民居住证，上传的附件为3张，第1张是台湾居民居住证人像面，第2张是台湾居民居住证国徽面，第3张是手持台湾居民居住证人像面照片。当identifyType=0的时候，该字段需要填写，否则忽略该字段的取值。|
	VerifiedType *int32 `json:"verified_type,omitempty"`
	// |参数名称：个人证件认证时证件附件的文件URL，该URL地址必须按照顺序填写。以身份证举例，譬如身份证人像面文件名称是abc023，国徽面是def004，个人手持身份证人像面是gh007，那么这个地方需要按照abc023def004gh007的顺序填写URL（文件名称区分大小写）。以护照举例，譬如护照个人资料页文件名称是abc023，手持护照个人资料页是def004，那么这个地方需要按照abc023def004的顺序填写URL（文件名称区分大小写）。证件附件目前仅仅支持jpg、jpeg、bmp、png、gif、pdf格式，单个文件最大不超过10M。这个URL是相对URL，不需要包含桶名和download目录，只要包含download目录下的子目录和对应文件名称即可。举例如下：如果上传的证件附件在桶中的位置是：https://bucketname.obs.Endpoint.myhuaweicloud.com/download/abc023.jpg，该字段填写abc023.jpg；如果上传的证件附件在桶中的位置是：https://bucketname.obs.Endpoint.myhuaweicloud.com/download/test/abc023.jpg，该字段填写test/abc023.jpg。| |参数约束以及描述：个人证件认证时证件附件的文件URL，该URL地址必须按照顺序填写。以身份证举例，譬如身份证人像面文件名称是abc023，国徽面是def004，个人手持身份证人像面是gh007，那么这个地方需要按照abc023def004gh007的顺序填写URL（文件名称区分大小写）。以护照举例，譬如护照个人资料页文件名称是abc023，手持护照个人资料页是def004，那么这个地方需要按照abc023def004的顺序填写URL（文件名称区分大小写）。证件附件目前仅仅支持jpg、jpeg、bmp、png、gif、pdf格式，单个文件最大不超过10M。这个URL是相对URL，不需要包含桶名和download目录，只要包含download目录下的子目录和对应文件名称即可。举例如下：如果上传的证件附件在桶中的位置是：https://bucketname.obs.Endpoint.myhuaweicloud.com/download/abc023.jpg，该字段填写abc023.jpg；如果上传的证件附件在桶中的位置是：https://bucketname.obs.Endpoint.myhuaweicloud.com/download/test/abc023.jpg，该字段填写test/abc023.jpg。|
	VerifiedFileUrl []string `json:"verified_file_url"`
	// |参数名称：姓名。| |参数约束及描述：姓名。|
	Name string `json:"name"`
	// |参数名称：证件号码。| |参数约束及描述：证件号码。|
	VerifiedNumber string `json:"verified_number"`
	// |参数名称：变更类型：-1：首次实名认证| |参数的约束及描述：变更类型：-1：首次实名认证|
	ChangeType *int32 `json:"change_type,omitempty"`
	// |参数名称：华为分给合作伙伴的平台标识。该标识的具体值由华为分配。获取方法请参见如何获取xaccountType的取值如何获取xaccountType的取值。| |参数约束及描述：华为分给合作伙伴的平台标识。该标识的具体值由华为分配。获取方法请参见如何获取xaccountType的取值如何获取xaccountType的取值。|
	XaccountType string          `json:"xaccount_type"`
	BankCardInfo *BankCardInfoV2 `json:"bank_card_info,omitempty"`
}

func (o ApplyIndividualRealnameAuthsReq) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "ApplyIndividualRealnameAuthsReq struct{}"
	}

	return strings.Join([]string{"ApplyIndividualRealnameAuthsReq", string(data)}, " ")
}
