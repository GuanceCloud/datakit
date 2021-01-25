## 0.0.30-rc 2021-01-15
## HuaweiCloud SDK Core
- ### Features
    - Support function `ValueOf` to get region information.
- ### Bug Fix
    - None
- ### Change
    - None

## HuaweiCloud SDK CloudBuild
- ### Features
    - Support more interface: `ShowListHistory`.
- ### Bug Fix
    - None
- ### Change
    - None

## HuaweiCloud SDK DGC
- ### Features
    - Support more interfaces: `Job` related interfaces, `Script` related interfaces, `Resource` related interfaces.
- ### Bug Fix
    - None
- ### Change
    - None

## HuaweiCloud SDK IAM
- ### Features
    - None
- ### Bug Fix
    - None
- ### Change
    - Modify the data type of response field `is_domain_owner` from string to boolean of interface `ShowUser` and `CreateUser`.

## HuaweiCloud SDK Live
- ### Features
    - Support more interface: `ListLiveStreamsOnline`.
- ### Bug Fix
    - None
- ### Change
    - None

## HuaweiCloud SDK RDS
- ### Features
    - Support more interfaces: ShowOffSiteBackupPolicy, SetOffSiteBackupPolicy, ListOffSiteBackups, ListOffSiteRestoreTimes, ListOffSiteRestoreTimes
- ### Bug Fix
    - None
- ### Change
    - None

## HuaweiCloud SDK SWR
- ### Features
    - Support `Software Repository for Container` service.
- ### Bug Fix
    - None
- ### Change
    - None


## 0.0.29-beta 2020-12-31
## HuaweiCloud SDK BMS
 - ### Features
    - None
 - ### Bug Fix
    - Fix the problem of interface: ListBareMetalServers.
    - Fix the problem of interface: ListBareMetalServerDetails.
    - Modify interface fields: ShowBaremetalServerInterfaceAttachments.
 - ### Change
    - None

## HuaweiCloud SDK CloudIDE
 - ### Features
    - Support more interface: ShowAccountStatus.
 - ### Bug Fix
    - None
 - ### Change
    - None

## HuaweiCloud SDK DCS
 - ### Features
    - None
 - ### Bug Fix
    - Modify the interface return data type to prevent deserialization failure: 
        - ListSlowlog: change data type of `Tags` from Tag to ResourceTag.
        - ListInstances: change data type of `duration` from int32 to string.
        - ShowBigkeyScanTaskDetails: change data type of `db` from int32 to string.
        - ShowHotkeyTaskDetails: change data type of `db` from int32 to string.
 - ### Change
    - None

## HuaweiCloud SDK DGC
 - ### Features
    - Support `Data Lake Governance Center` service.
 - ### Bug Fix
    - None
 - ### Change
    - None

## HuaweiCloud SDK DIS
 - ### Features
    - Support `Data Ingestion Service`.
 - ### Bug Fix
    - None
 - ### Change
    - None

## HuaweiCloud SDK RDS
 - ### Features
    - None
 - ### Bug Fix
    - None
 - ### Change
    - Interface modification: response type of interface `CreateInstance` adjustment.

## HuaweiCloud SDK SMN
 - ### Features
    - None
 - ### Bug Fix
    - Modify the request parameters of interface `PublishMessage` from Object to Map<String, String>
 - ### Change
    - None


## 0.0.28-beta 2020-12-28
## HuaweiCloud SDK Core
 - ### Features
    - None
 - ### Bug Fix
    - Fix response exception when using temporary AK/SK.
 - ### Change
    - None

## HuaweiCloud SDK DCS
 - ### Features
    - None
 - ### Bug Fix
    - Change property type of `port` from string to integer.
 - ### Change
    - None


## 0.0.27-beta 2020-12-25
## HuaweiCloud SDK DCS
 - ### Features
    - None
 - ### Bug Fix
    - Fix the problem of compilation failure in Mac OS.
 - ### Change
    - Query parameter in interface `ListInstances` modification: id → instance_id.

## HuaweiCloud SDK DDS
 - ### Features
    - Support Document Database Service.
 - ### Bug Fix
    - None
 - ### Change
    - None

## HuaweiCloud SDK Kafka
 - ### Features
    - None
 - ### Bug Fix
    - Fix the problem of compilation failure in Mac OS.
 - ### Change
    - None

## HuaweiCloud SDK RabbitMQ
 - ### Features
    - None
 - ### Bug Fix
    - Fix the problem of compilation failure in Mac OS.
 - ### Change
    - None

## HuaweiCloud SDK RMS
 - ### Features
    - Support more interfaces: `Resources` related interfaces and `Region` related interfaces. 
 - ### Bug Fix
    - None
 - ### Change
    - None


## 0.0.26-beta 2020-12-23
## HuaweiCloud SDK Core
 - ### Features
    - Support Endpoint Resolver: it's supported to use {Service}Region when initializing {ServiceClient} which can automatically backfill endpoint. After choosing a region, the projectId/domainId will be backfilled automatically.
 - ### Bug Fix
    - None
 - ### Change
    - None

## HuaweiCloud SDK BSS
 - ### Features
    - Support more interfaces: ListMeasureUnits.
 - ### Bug Fix
    - None
 - ### Change
    - None

## HuaweiCloud SDK CES
 - ### Features
    - None
 - ### Bug Fix
    - None
 - ### Change
    - Update interface: ShowMetricData

## HuaweiCloud SDK Live
 - ### Features
    - Support more interfaces: ShowStreamPortrait.
 - ### Bug Fix
    - None
 - ### Change
    - None

## HuaweiCloud SDK MPC
 - ### Features
    - Support more interfaces: QualityEnhanceTemplate related interfaces and MergeChannelsTask related interfaces.
 - ### Bug Fix
    - None
 - ### Change
    - None

## HuaweiCloud SDK RDS
 - ### Features
    - Support Relational Database Service.
 - ### Bug Fix
    - None
 - ### Change
    - None

## HuaweiCloud SDK SMN
 - ### Features
    - None
 - ### Bug Fix
    - None
 - ### Change
    - Update field type in message_template_name.


## 0.0.25-beta 2020-12-15
## HuaweiCloud SDK CCE
 - ### Features
    - Support Cloud Container Engine service.
 - ### Bug Fix
    - None
 - ### Change
    - None

## HuaweiCloud SDK ELB
 - ### Features
    - None
 - ### Bug Fix
    - Fix the problem that sending request to interface `CreateListener` returns empty response.
    - Fix the problem that sending request to interface `CreateListener` returns response with wrong type. 
 - ### Change
    - None

## HuaweiCloud SDK FunctionGraph
 - ### Features
    - Support more interfaces: UpdateFunctionReservedInstances.
 - ### Bug Fix
    - None
 - ### Change
    - None

## HuaweiCloud SDK NAT
 - ### Features
    - None
 - ### Bug Fix
    - Fix the problem that using interface `BatchCreateNatGatewayDnatRules` failed. 
 - ### Change
    - None


## 0.0.24-beta 2020-12-04
## HuaweiCloud SDK SMN
 - ### Features
    - Support Simple Message Notification service.
 - ### Bug Fix
    - None
 - ### Change
    - None


## 0.0.23-beta 2020-11-30
## HuaweiCloud SDK BCS
 - ### Features
    - Support BlockChain Service.
 - ### Bug Fix
    - None
 - ### Change
    - None 

## HuaweiCloud SDK BMS
 - ### Features
    - Support Bare Metal Server service.
 - ### Bug Fix
    - None
 - ### Change
    - None

## HuaweiCloud SDK BSS
 - ### Features
    - Support more interfaces: ListUsageTypes, ModPeriodToOnDemand.
 - ### Bug Fix
    - None
 - ### Change
    - None 

## HuaweiCloud SDK CBR
 - ### Features
    - Support more interfaces: MigrateVaultResource
 - ### Bug Fix
    - None
 - ### Change
    - None

## HuaweiCloud SDK CES
 - ### Features
    - Support more interfaces: 
     - ListEvents
     - ListEventDetail
     - CreateResourceGroup
     - UpdateResourceGroup
     - DeleteResourceGroup
     - ListResourceGroup
     - UpdateAlarm
 - ### Bug Fix
    - None
 - ### Change
    - None

## HuaweiCloud SDK DCS
 - ### Features
    - Support Distributed Cache Service.
 - ### Bug Fix
    - None
 - ### Change
    - None

## HuaweiCloud SDK ECS
 - ### Features
    - None 
 - ### Bug Fix
    - None
 - ### Change
    - [Issue 21](https://github.com/huaweicloud/huaweicloud-sdk-java-v3/issues/21) Open related interface.

## HuaweiCloud SDK FunctionGraph
 - ### Features
    - Support FunctionGraph service.
 - ### Bug Fix
    - None
 - ### Change
    - None

## HuaweiCloud SDK IAM
 - ### Features
    - Support more interfaces.
 - ### Bug Fix
    - None
 - ### Change
    - None

## HuaweiCloud SDK Live
 - ### Features
    - None
 - ### Bug Fix
    - None
 - ### Change
    - Name of service client modification: LiveAPIClient → LiveClient.


## 0.0.22-beta 2020-11-17
## HuaweiCloud SDK AS
 - ### Features
    - None
 - ### Bug Fix
    - [Issue 8](https://github.com/huaweicloud/huaweicloud-sdk-go-v3/issues/8) Fix the problem that creating scaling policy failed.
 - ### Change
    - None

## HuaweiCloud SDK DMS
 - ### Features
    - Support Distributed Message Services, provide Kafka premium instances and RabbitMQ premium instances with dedicated resources.
 - ### Bug Fix
    - None
 - ### Change
    - None
 
## HuaweiCloud SDK ECS
 - ### Features
    - None
 - ### Bug Fix
    - None
 - ### Change
    - Property adjustment:  increase property `dry_run` in interfaces `CreatePostPaidServers` and `CreateServers` which means whether parameters will be checked before sending real requests. 


## HuaweiCloud SDK NAT
 - ### Features
    - Support NAT Gateway service.
 - ### Bug Fix
    - None
 - ### Change
    - None 

## HuaweiCloud SDK VPC
 - ### Features
    - Support more interfaces: interfaces related to Network ACLs. 
 - ### Bug Fix
    - None
 - ### Change
    - None 


## 0.0.21-beta 2020-11-11
## HuaweiCloud SDK Core
 - ### Features
    - None
 - ### Bug Fix
    - Update core code from [Pull requests #11](https://github.com/huaweicloud/huaweicloud-sdk-go-v3/pull/11).
 - ### Change
    - None

## HuaweiCloud SDK CBR
 - ### Features
    - None
 - ### Bug Fix
    - None
 - ### Change
    - Interface adjustment: property `object_type` in interface `CreateVault` support key `turbo`.
    - Interface adjustment: property `description` in interface `UpdatePolicy` is removed.

## HuaweiCloud SDK CES
 - ### Features
    - Add examples of interface response and adjust some filed description.
 - ### Bug Fix
    - None
 - ### Change
    - None

## HuaweiCloud SDK CloudPipeline
 - ### Features
    - None
 - ### Bug Fix
    - None
 - ### Change
    - Modify the name of generated Client class: devcloudpipeline_client → cloudPipeline_client
    - Modify the name of generated Meta class: devcloudpipeline_meta → cloudPipeline_meta

## HuaweiCloud SDK DevStar
 - ### Features
    - None
 - ### Bug Fix
    - None
 - ### Change
    - Modify interface parameters and adjust sample code.


## 0.0.20-beta 2020-11-02
## HuaweiCloud SDK CES
 - ### Features
    - None
 - ### Bug Fix
    - None
 - ### Change
    - Interface adjustment: property `alarm_type` in class `CreateAlarmRequestBody` support key `RESOURCE_GROUP`.
    - Interface adjustment: property `meta_data` in class `ListAlarmHistoriesResponse` only returns total number of alarm histories.

## HuaweiCloud SDK ELB
 - ### Features
    - None
 - ### Bug Fix
    - Modify wrong parameter in class `CreateL7ruleRequestBody`.
 - ### Change
    - None


## 0.0.19-beta 2020-10-31
## HuaweiCloud SDK Core
 - ### Features
    - None
 - ### Bug Fix
    - Fix: fix the problem that when query parameter contains enumerated variables the request will fail.
    - [Issue 7](https://github.com/huaweicloud/huaweicloud-sdk-go-v3/issues/7) resolve the problem of using json.Marshal() returns object{}.
 - ### Change
    - None

## HuaweiCloud SDK CBR
 - ### Features
    - Support more interfaces: interfaces related to `TAG`.
 - ### Bug Fix
    - [Issue 17](https://github.com/huaweicloud/huaweicloud-sdk-java-v3/issues/17) fix the problem of interface: ListBackups.
 - ### Change
    - None

## HuaweiCloud SDK CTS
 - ### Features
    - Support more interface: ListQuotas
 - ### Bug Fix
    - None
 - ### Change
    - None

## HuaweiCloud SDK EPS
 - ### Features
    - None
 - ### Bug Fix
    - None
 - ### Change
    - Adjust interfaces' names, replace abbreviations of `EP` with `EnterpriseProject`, the involved interfaces are:
     1. ListEP → ListEnterpriseProject
     2. CreateEP → CreateEnterpriseProject
     3. ShowEP → ShowEnterpriseProject
     4. ModifyEP → ModifyEnterpriseProject
     5. EnableEP → EnableEnterpriseProject
     6. ShowEPQuota → ShowEnterpriseProjectQuota
     7. ShowResourceBindEP → ShowResourceBindEnterpriseProject
     8. DisableEP → DisableEnterpriseProject

## HuaweiCloud SDK Iam
 - ### Features
    - None
 - ### Bug Fix
    - None
 - ### Change
    - Adjust interfaces' names, the involved interfaces are:
     1. KeystoneCreateUserTokenByPasswordAndMFA → KeystoneCreateUserTokenByPasswordAndMfa
     2. CreateUnscopeTokenByIDPInitiated → CreateUnscopeTokenByIdpInitiated

## HuaweiCloud SDK ProjectMan
 - ### Features
    - Support more interfaces: iteration information, user information, project members, project information, project indicators, project statistics, etc.
 - ### Bug Fix
    - None
 - ### Change
    - None


## 0.0.18-beta 2020-10-20
## HuaweiCloud SDK ELB
 - ### Features
    - Support more interfaces of version v2.
 - ### Bug Fix
    - None
 - ### Change
    - None


## 0.0.17-beta 2020-10-14
## HuaweiCloud SDK BSS
 - ### Features
    - Partner center supports exporting product catalog prices.
 - ### Bug Fix
    - None
 - ### Change
    - None

## HuaweiCloud SDK Live
 - ### Features
    - Support more interfaces of version v2 of Live.
 - ### Bug Fix
    - None
 - ### Change
    - None


## 0.0.16-beta 2020-10-12
## HuaweiCloud SDK CTS
 - ### Features
    - None
 - ### Bug Fix
    - None
 - ### Change
    - Delete deprecated interfaces of version v1.

## HuaweiCloud SDK DevStar
 - ### Features
    - None
 - ### Bug Fix
    - Modify the credential type of DevStarClient: the correct credential type is GlobalCredentials.
 - ### Change
    - None


## 0.0.15-beta 2020-09-30
## HuaweiCloud SDK AS
 - ### Features
    - Support Auto Scaling service.
 - ### Bug Fix
    - None
 - ### Change
    - None


## 0.0.14-beta 2020-09-24
## HuaweiCloud SDK BSS
 - ### Features
    - None
 - ### Bug Fix
    - Fix the problem that the class `BssClient` cannot be loaded.
 - ### Change
    - None

## HuaweiCloud SDK EIP
 - ### Features
    - None
 - ### Bug Fix
    - None
 - ### Change
    - Interface `ListPublicips` adjustment: enterprise_project_id does not support multi-value query.


## 0.0.13-beta 2020-09-16
## HuaweiCloud SDK ECS
 - ### Features
    - None
 - ### Bug Fix
    - Fix parameter type of interfaces.
 - ### Change
    - None

## HuaweiCloud SDK BSS
 - ### Features
    - None
 - ### Bug Fix
    - None
 - ### Change
    - Update interfaces.

## HuaweiCloud SDK EIP
 - ### Features
    - None
 - ### Bug Fix
    - Fix the problem that not support transfer multiple values.
 - ### Change
    - None

## HuaweiCloud SDK ELB
 - ### Features
    - None
 - ### Bug Fix
    - Fix the problem that some parameters are wrong.
 - ### Change
    - None

## HuaweiCloud SDK IMS
 - ### Features
    - None
 - ### Bug Fix
    - None
 - ### Change
    - Adjust descriptions of interfaces.

## HuaweiCloud SDK Live
 - ### Features
    - None
 - ### Bug Fix
    - None
 - ### Change
    - Adjust descriptions of banned interface.


## 0.0.12.1-beta 2020-09-16
## HuaweiCloud SDK ECS
 - ### Features
    - None
 - ### Bug Fix
    - Fix parameter type of interfaces.
 - ### Change
    - None

## HuaweiCloud SDK All
 - ### Features
    - None
 - ### Bug Fix
    - None
 - ### Change
    - When the optional parameter type is list, the parameter type will be changed to a pointer, for example, []string to *[]string

## 0.0.12-beta 2020-09-11
## HuaweiCloud SDK KPS
 - ### Features
    - Support KPS service.
 - ### Bug Fix
    - None
 - ### Change
    - None

## HuaweiCloud SDK EVS
 - ### Features
    - Support EVS service.
 - ### Bug Fix
    - None
 - ### Change
    - None

## HuaweiCloud SDK CBR
 - ### Features
    - None
 - ### Bug Fix
    - Fix response value type definition errors.
 - ### Change
    - None

## 0.0.11-beta 2020-09-09
## HuaweiCloud SDK CBR
 - ### Features
    - None
 - ### Bug Fix
    - Fix the problem that resources related interfaces have wrong data type.
 - ### Change
    - None

## HuaweiCloud SDK VPC
 - ### Features
    - None
 - ### Bug Fix
    - Fix the problem that security group related interfaces have wrong data type.
 - ### Change
    - None


## 0.0.10-beta 2020-09-04
## HuaweiCloud SDK Core
 - ### Features
    - None
 - ### Bug Fix
    - Fix the problem that the enumeration code cannot be generated for integer enumeration parameters without format defined in yaml.
 - ### Change
    - Modify User Agent of Http Request header.


# 0.0.9-beta 2020-08-28
## HuaweiCloud SDK CloudPipeline
 - ### Features
    - Support CloudPipeline service.
 - ### Bug Fix
    - None
 - ### Change
    - None

## HuaweiCloud SDK EIP
 - ### Features
    - Support more APIs: tags related APIs and shared bandwidth related APIs.
 - ### Bug Fix
    - Interface BatchCreateBandwidth: modify field billing_info. 
 - ### Change
    - None

## HuaweiCloud SDK IAM
 - ### Features
    - Support more APIs: consistency of console related APIs.
 - ### Bug Fix
    - None
 - ### Change
    - None

## HuaweiCloud SDK ProjectMan
 - ### Features
    - Support Project Management service.
 - ### Bug Fix
    - None
 - ### Change
    - None

## HuaweiCloud SDK VPC
 - ### Features
    - Support more APIs: security group, security group rules, available IP count related APIs.
 - ### Bug Fix
    - None
 - ### Change
    - None


# 0.0.8-beta 2020-08-25
## HuaweiCloud SDK Core
 - ### Features
    - None
 - ### Bug Fix
    - None
 - ### Change
    - Change optional enum variable type to pointer.

## HuaweiCloud SDK ELB
 - ### Features
    - Support Elastic Load Balance service.
 - ### Bug Fix
    - None
 - ### Change
    - None


# 0.0.7-beta 2020-08-20
## HuaweiCloud SDK Core
 - ### Features
    - None
 - ### Bug Fix
    - Fix the problem that some enum variables are unreadable.
 - ### Change
    - Add 'E' as prefix to enum Variables which start with '_'.


# 0.0.6-beta 2020-08-20
## HuaweiCloud SDK Core
 - ### Features
    - None
 - ### Bug Fix
    - Fix the problem of missing the imports when the struct contains enum variables.
 - ### Change
    - None


# 0.0.5-beta 2020-08-19
## HuaweiCloud SDK Core
 - ### Features
    - None
 - ### Bug Fix
    - Fix the deserialization failure of enum variables.
    - Fix the deserialization failure of error response in some scenarios.
 - ### Change
    - None


# 0.0.4-beta 2020-08-18
## HuaweiCloud SDK Core
 - ### Features
    - None
 - ### Bug Fix
    - Fix the problem of missing default values of Go basic type when serializing.
 - ### Change
    - None


# 0.0.3-beta 2020-08-14
## HuaweiCloud SDK APIG
 - ### Features
    - Support API Gateway service.
 - ### Bug Fix
    - None
 - ### Change
    - None

## HuaweiCloud SDK BSS
 - ### Features
    - Support Business Support System.
 - ### Bug Fix
    - None
 - ### Change
    - None


# 0.0.2-beta 2020-8-11
## HuaweiCloud SDK Core
 - ### Features
    - Support temporary AK/SK authentication mode.
 - ### Bug Fix
    - None
 - ### Change
    - None
    
## HuaweiCloud SDK CBR
 - ### Features
    - Support Cloud Backup and Recovery service.
 - ### Bug Fix
    - None
 - ### Change
    - None

## HuaweiCloud SDK CloudIDE
 - ### Features
    - Support Cloud IDE service.
 - ### Bug Fix
    - None
 - ### Change
    - None

## HuaweiCloud SDK ECS
 - ### Features
    - Support Elastic Cloud Server service.
 - ### Bug Fix
    - None
 - ### Change
    - None

## HuaweiCloud SDK EIP
 - ### Features
    - Support Elastic IP service.
 - ### Bug Fix
    - None
 - ### Change
    - None

## HuaweiCloud SDK EVS
 - ### Features
    - Support Elastic Volume Service.
 - ### Bug Fix
    - None
 - ### Change
    - None

## HuaweiCloud SDK IMS
 - ### Features
    - Support Image Management Service.
 - ### Bug Fix
    - None
 - ### Change
    - None
    
## HuaweiCloud SDK Live
 - ### Features
    - Support Live service.
 - ### Bug Fix
    - None
 - ### Change
    - None

## HuaweiCloud SDK MPC
 - ### Features
    - Support Media Processing Center.
 - ### Bug Fix
    - None
 - ### Change
    - None


# __3.0.1-beta__ __2020-07-30__
## First Release
 - ### Supported Services
    - Classroom
    - Cloud Trace Service(CTS)
    - DevStar
    - Enterprise Project Management Service(EPS)
    - Identity and Access Management(IAM)
    - Tag Management Service(TMS)
    - Virtual Private Cloud(VPC)
