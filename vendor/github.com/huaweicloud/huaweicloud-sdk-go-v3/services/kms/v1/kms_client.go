package v1

import (
	http_client "github.com/huaweicloud/huaweicloud-sdk-go-v3/core"
	"github.com/huaweicloud/huaweicloud-sdk-go-v3/services/kms/v1/model"
)

type KmsClient struct {
	HcClient *http_client.HcHttpClient
}

func NewKmsClient(hcClient *http_client.HcHttpClient) *KmsClient {
	return &KmsClient{HcClient: hcClient}
}

func KmsClientBuilder() *http_client.HcHttpClientBuilder {
	builder := http_client.NewHcHttpClientBuilder()
	return builder
}

//- 功能介绍：批量添加删除密钥标签。
func (c *KmsClient) BatchCreateKmsTags(request *model.BatchCreateKmsTagsRequest) (*model.BatchCreateKmsTagsResponse, error) {
	requestDef := GenReqDefForBatchCreateKmsTags()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.BatchCreateKmsTagsResponse), nil
	}
}

//- 功能介绍：撤销授权，授权用户撤销被授权用户操作密钥的权限。 - 说明：    - 创建密钥的用户才能撤销该密钥授权。
func (c *KmsClient) CancelGrant(request *model.CancelGrantRequest) (*model.CancelGrantResponse, error) {
	requestDef := GenReqDefForCancelGrant()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.CancelGrantResponse), nil
	}
}

//- 功能介绍：取消计划删除密钥。  - 说明：密钥处于“计划删除”状态才能取消计划删除密钥。
func (c *KmsClient) CancelKeyDeletion(request *model.CancelKeyDeletionRequest) (*model.CancelKeyDeletionResponse, error) {
	requestDef := GenReqDefForCancelKeyDeletion()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.CancelKeyDeletionResponse), nil
	}
}

//- 功能介绍：退役授权，表示被授权用户不再具有授权密钥的操作权。   例如：用户A授权用户B可以操作密钥A/key，同时授权用户C可以撤销该授权，   那么用户A、B、C均可退役该授权，退役授权后，用户B不再可以使用A/key。 - 须知：      可执行退役授权的主体包括：    - 创建授权的用户；    - 授权中retiring_principal指向的用户；    - 当授权的操作列表中包含retire-grant时，grantee_principal指向的用户。
func (c *KmsClient) CancelSelfGrant(request *model.CancelSelfGrantRequest) (*model.CancelSelfGrantResponse, error) {
	requestDef := GenReqDefForCancelSelfGrant()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.CancelSelfGrantResponse), nil
	}
}

//- 功能介绍：创建数据密钥，返回结果包含明文和密文。
func (c *KmsClient) CreateDatakey(request *model.CreateDatakeyRequest) (*model.CreateDatakeyResponse, error) {
	requestDef := GenReqDefForCreateDatakey()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.CreateDatakeyResponse), nil
	}
}

//- 功能介绍：创建数据密钥，返回结果只包含密文。
func (c *KmsClient) CreateDatakeyWithoutPlaintext(request *model.CreateDatakeyWithoutPlaintextRequest) (*model.CreateDatakeyWithoutPlaintextResponse, error) {
	requestDef := GenReqDefForCreateDatakeyWithoutPlaintext()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.CreateDatakeyWithoutPlaintextResponse), nil
	}
}

//- 功能介绍：创建授权，被授权用户可以对授权密钥进行操作。 - 说明：    - 服务默认主密钥（密钥别名后缀为“/default”）不可以授权。
func (c *KmsClient) CreateGrant(request *model.CreateGrantRequest) (*model.CreateGrantResponse, error) {
	requestDef := GenReqDefForCreateGrant()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.CreateGrantResponse), nil
	}
}

//- 功能介绍：创建用户主密钥，可用来加密数据密钥。  - 说明： 别名“/default”为服务默认主密钥的后缀名，由服务自动创建。因此用户创建的主密钥别名不能与服务默认主密钥的别名相同，即后缀名不能为“/default”。对于开通企业项目的用户，服务默认主密钥属于且只能属于默认企业项目下，且不支持企业资源的迁入迁出。服务默认主密钥为用户提供基础的云上加密功能，满足合规要求。因此，在企业多项目下，其他非默认企业项目下的用户均可使用该密钥。若客户有企业管理资源诉求，请自行创建和使用密钥。
func (c *KmsClient) CreateKey(request *model.CreateKeyRequest) (*model.CreateKeyResponse, error) {
	requestDef := GenReqDefForCreateKey()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.CreateKeyResponse), nil
	}
}

//- 功能介绍：添加密钥标签。
func (c *KmsClient) CreateKmsTag(request *model.CreateKmsTagRequest) (*model.CreateKmsTagResponse, error) {
	requestDef := GenReqDefForCreateKmsTag()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.CreateKmsTagResponse), nil
	}
}

//- 功能介绍：获取导入密钥的必要参数，包括密钥导入令牌和密钥加密公钥。 - 说明：返回的公钥类型默认为RSA_2048。
func (c *KmsClient) CreateParametersForImport(request *model.CreateParametersForImportRequest) (*model.CreateParametersForImportResponse, error) {
	requestDef := GenReqDefForCreateParametersForImport()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.CreateParametersForImportResponse), nil
	}
}

//- 功能介绍：   生成8~8192bit范围内的随机数。   生成512bit的随机数。
func (c *KmsClient) CreateRandom(request *model.CreateRandomRequest) (*model.CreateRandomResponse, error) {
	requestDef := GenReqDefForCreateRandom()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.CreateRandomResponse), nil
	}
}

//- 功能介绍：解密数据。
func (c *KmsClient) DecryptData(request *model.DecryptDataRequest) (*model.DecryptDataResponse, error) {
	requestDef := GenReqDefForDecryptData()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.DecryptDataResponse), nil
	}
}

//- 功能介绍：解密数据密钥，用指定的主密钥解密数据密钥。
func (c *KmsClient) DecryptDatakey(request *model.DecryptDatakeyRequest) (*model.DecryptDatakeyResponse, error) {
	requestDef := GenReqDefForDecryptDatakey()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.DecryptDatakeyResponse), nil
	}
}

//- 功能介绍：删除密钥材料信息。
func (c *KmsClient) DeleteImportedKeyMaterial(request *model.DeleteImportedKeyMaterialRequest) (*model.DeleteImportedKeyMaterialResponse, error) {
	requestDef := GenReqDefForDeleteImportedKeyMaterial()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.DeleteImportedKeyMaterialResponse), nil
	}
}

//- 功能介绍：计划多少天后删除密钥，可设置7天～1096天内删除密钥。
func (c *KmsClient) DeleteKey(request *model.DeleteKeyRequest) (*model.DeleteKeyResponse, error) {
	requestDef := GenReqDefForDeleteKey()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.DeleteKeyResponse), nil
	}
}

//- 功能介绍：删除密钥标签。
func (c *KmsClient) DeleteTag(request *model.DeleteTagRequest) (*model.DeleteTagResponse, error) {
	requestDef := GenReqDefForDeleteTag()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.DeleteTagResponse), nil
	}
}

//- 功能介绍：禁用密钥，密钥禁用后不可以使用。  - 说明：密钥为启用状态才能禁用密钥。
func (c *KmsClient) DisableKey(request *model.DisableKeyRequest) (*model.DisableKeyResponse, error) {
	requestDef := GenReqDefForDisableKey()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.DisableKeyResponse), nil
	}
}

//- 功能介绍：关闭用户主密钥轮换。
func (c *KmsClient) DisableKeyRotation(request *model.DisableKeyRotationRequest) (*model.DisableKeyRotationResponse, error) {
	requestDef := GenReqDefForDisableKeyRotation()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.DisableKeyRotationResponse), nil
	}
}

//- 功能介绍：启用密钥，密钥启用后才可以使用。  - 说明：密钥为禁用状态才能启用密钥。
func (c *KmsClient) EnableKey(request *model.EnableKeyRequest) (*model.EnableKeyResponse, error) {
	requestDef := GenReqDefForEnableKey()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.EnableKeyResponse), nil
	}
}

//- 功能介绍：开启用户主密钥轮换。 - 说明：   - 开启密钥轮换后，默认轮询间隔时间为365天。   - 默认主密钥及外部导入密钥不支持轮换操作。
func (c *KmsClient) EnableKeyRotation(request *model.EnableKeyRotationRequest) (*model.EnableKeyRotationResponse, error) {
	requestDef := GenReqDefForEnableKeyRotation()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.EnableKeyRotationResponse), nil
	}
}

//- 功能介绍：加密数据，用指定的用户主密钥加密数据。
func (c *KmsClient) EncryptData(request *model.EncryptDataRequest) (*model.EncryptDataResponse, error) {
	requestDef := GenReqDefForEncryptData()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.EncryptDataResponse), nil
	}
}

//- 功能介绍：加密数据密钥，用指定的主密钥加密数据密钥。
func (c *KmsClient) EncryptDatakey(request *model.EncryptDatakeyRequest) (*model.EncryptDatakeyResponse, error) {
	requestDef := GenReqDefForEncryptDatakey()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.EncryptDatakeyResponse), nil
	}
}

//- 功能介绍：导入密钥材料。
func (c *KmsClient) ImportKeyMaterial(request *model.ImportKeyMaterialRequest) (*model.ImportKeyMaterialResponse, error) {
	requestDef := GenReqDefForImportKeyMaterial()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.ImportKeyMaterialResponse), nil
	}
}

//- 功能介绍：查询密钥的授权列表。
func (c *KmsClient) ListGrants(request *model.ListGrantsRequest) (*model.ListGrantsResponse, error) {
	requestDef := GenReqDefForListGrants()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.ListGrantsResponse), nil
	}
}

//- 功能介绍：查询密钥详细信息。
func (c *KmsClient) ListKeyDetail(request *model.ListKeyDetailRequest) (*model.ListKeyDetailResponse, error) {
	requestDef := GenReqDefForListKeyDetail()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.ListKeyDetailResponse), nil
	}
}

//- 功能介绍：查询用户所有密钥列表。
func (c *KmsClient) ListKeys(request *model.ListKeysRequest) (*model.ListKeysResponse, error) {
	requestDef := GenReqDefForListKeys()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.ListKeysResponse), nil
	}
}

//- 功能介绍：查询密钥实例。通过标签过滤，查询指定用户主密钥的详细信息。
func (c *KmsClient) ListKmsByTags(request *model.ListKmsByTagsRequest) (*model.ListKmsByTagsResponse, error) {
	requestDef := GenReqDefForListKmsByTags()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.ListKmsByTagsResponse), nil
	}
}

//- 功能介绍：查询用户在指定项目下的所有标签集合。
func (c *KmsClient) ListKmsTags(request *model.ListKmsTagsRequest) (*model.ListKmsTagsResponse, error) {
	requestDef := GenReqDefForListKmsTags()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.ListKmsTagsResponse), nil
	}
}

//- 功能介绍：查询用户可以退役的授权列表。
func (c *KmsClient) ListRetirableGrants(request *model.ListRetirableGrantsRequest) (*model.ListRetirableGrantsResponse, error) {
	requestDef := GenReqDefForListRetirableGrants()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.ListRetirableGrantsResponse), nil
	}
}

//- 功能介绍：查询用户主密钥轮换状态。
func (c *KmsClient) ShowKeyRotationStatus(request *model.ShowKeyRotationStatusRequest) (*model.ShowKeyRotationStatusResponse, error) {
	requestDef := GenReqDefForShowKeyRotationStatus()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.ShowKeyRotationStatusResponse), nil
	}
}

//- 功能介绍：查询密钥标签。
func (c *KmsClient) ShowKmsTags(request *model.ShowKmsTagsRequest) (*model.ShowKmsTagsResponse, error) {
	requestDef := GenReqDefForShowKmsTags()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.ShowKmsTagsResponse), nil
	}
}

//- 功能介绍：查询实例数，获取用户已经创建的用户主密钥数量。
func (c *KmsClient) ShowUserInstances(request *model.ShowUserInstancesRequest) (*model.ShowUserInstancesResponse, error) {
	requestDef := GenReqDefForShowUserInstances()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.ShowUserInstancesResponse), nil
	}
}

//- 功能介绍：查询配额，查询用户可以创建的用户主密钥配额总数及当前使用量信息。
func (c *KmsClient) ShowUserQuotas(request *model.ShowUserQuotasRequest) (*model.ShowUserQuotasResponse, error) {
	requestDef := GenReqDefForShowUserQuotas()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.ShowUserQuotasResponse), nil
	}
}

//- 功能介绍：修改用户主密钥别名。 - 说明：    - 服务默认主密钥（密钥别名后缀为“/default”）不可以修改。    - 密钥处于“计划删除”状态，密钥别名不可以修改。
func (c *KmsClient) UpdateKeyAlias(request *model.UpdateKeyAliasRequest) (*model.UpdateKeyAliasResponse, error) {
	requestDef := GenReqDefForUpdateKeyAlias()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.UpdateKeyAliasResponse), nil
	}
}

//- 功能介绍：修改用户主密钥描述信息。 - 说明：    - 服务默认主密钥（密钥别名后缀为“/default”）不可以修改。    - 密钥处于“计划删除”状态，密钥描述不可以修改。
func (c *KmsClient) UpdateKeyDescription(request *model.UpdateKeyDescriptionRequest) (*model.UpdateKeyDescriptionResponse, error) {
	requestDef := GenReqDefForUpdateKeyDescription()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.UpdateKeyDescriptionResponse), nil
	}
}

//- 功能介绍：修改用户主密钥轮换周期。
func (c *KmsClient) UpdateKeyRotationInterval(request *model.UpdateKeyRotationIntervalRequest) (*model.UpdateKeyRotationIntervalResponse, error) {
	requestDef := GenReqDefForUpdateKeyRotationInterval()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.UpdateKeyRotationIntervalResponse), nil
	}
}

//- 功能介绍：查指定API版本信息。
func (c *KmsClient) ShowVersion(request *model.ShowVersionRequest) (*model.ShowVersionResponse, error) {
	requestDef := GenReqDefForShowVersion()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.ShowVersionResponse), nil
	}
}

//- 功能介绍：查询API版本信息列表。
func (c *KmsClient) ShowVersions(request *model.ShowVersionsRequest) (*model.ShowVersionsResponse, error) {
	requestDef := GenReqDefForShowVersions()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.ShowVersionsResponse), nil
	}
}
