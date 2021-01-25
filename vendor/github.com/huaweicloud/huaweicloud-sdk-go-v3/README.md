English | [简体中文](./README_CN.md)

# HuaweiCloud Go Software Development Kit (Go SDK)

The HuaweiCloud Go SDK allows you to easily work with Huawei Cloud services such as Elastic Compute Service (ECS) and Virtual Private Cloud(VPC) without the need to handle API related tasks.

This document introduces how to obtain and use HuaweiCloud Go SDK.

## Getting Started

- To use HuaweiCloud Go SDK, you must have Huawei Cloud account as well as the Access Key and Secret key of the HuaweiCloud account.

    The accessKey is required when initializing `{Service}Client`. You can create an AccessKey in the Huawei Cloud console. For more information, see [My Credentials](https://support.huaweicloud.com/en-us/usermanual-ca/en-us_topic_0046606340.html).

- HuaweiCloud Go SDK requires go 1.14 or later.


## Install Go SDK

HuaweiCloud Go SDK supports go 1.14 or later. Run ``go version`` to check the version of Go.

- Use go get

    Run the following command to install the individual libraries of HuaweiCloud services:

    ``` bash
    # Install the core library
    go get github.com/huaweicloud/huaweicloud-sdk-go-v3
 
    # Install the dependent library
    go get github.com/json-iterator/go
    ```

## Use Go SDK

Take using VPC SDK for example, you need to import `github.com/huaweicloud/huaweicloud-sdk-go-v3/services/vpc/v2`:

1. Import the required modules as follows:

    ``` go
    import (
        "fmt"
        "github.com/huaweicloud/huaweicloud-sdk-go-v3/core/auth/basic"
        "github.com/huaweicloud/huaweicloud-sdk-go-v3/core/config"
        "github.com/huaweicloud/huaweicloud-sdk-go-v3/core/httphandler"
        vpc "github.com/huaweicloud/huaweicloud-sdk-go-v3/services/vpc/v2"
        "github.com/huaweicloud/huaweicloud-sdk-go-v3/services/vpc/v2/model"
        region "github.com/huaweicloud/huaweicloud-sdk-go-v3/services/vpc/v2/region"
        "net/http"
    )
    ```

2. Config `{Service}Client` Configurations

    2.1 Default Configuration

    ``` go
    // Use default configuration
    httpConfig := config.DefaultHttpConfig()
    ```

    2.2  Proxy(Optional)

    ``` go
    // Use proxy if needed
    httpConfig.WithProxy(config.NewProxy().
        WithSchema("http").
        WithHost("proxy.huaweicloud.com").
        WithPort(80).
        WithUsername("testuser").
        WithPassword("password"))))
    ```

    2.3 Connection(Optional)

    ``` go
    // seconds to wait for the server to send data before giving up
    httpConfig.WithTimeout(30);
    ```

    2.4 SSL Certification(Optional)

    ``` go
    // Skip ssl certification checking while using https protocol if needed
    httpConfig.WithIgnoreSSLVerification(true);

3. Initialize the credentials

    **Notice:**
    There are two types of HUAWEI CLOUD services, regional services and global services. 
    Global services currently only support BSS, DevStar, EPS, IAM, RMS.

    For Regional services' authentication, projectId is required. 
    For global services' authentication, domainId is required. 
    
    If you use {Service}Region to initialize {Service}Client, projectId/domainId supports automatic acquisition, you don't need to configure it when initializing Credentials.

    - `ak` is the access key ID for your account.
    - `sk` is the secret access key for your account.
    - `projectId` is the ID of your project depending on your region which you want to operate.
    - `domainId` is the account ID of HUAWEI CLOUD.
    - `securityToken` is the security token when using temporary AK/SK.

    3.1 Use permanent AK/SK
    
    ``` go
    # Regional Services
    auth := basic.NewCredentialsBuilder().
                WithAk(ak).
                WithSk(sk).
                WithProjectId(projectId).
                Build()
   
    # Global Services
    auth := global.NewCredentialsBuilder().
                WithAk(ak).
                WithSk(sk).
                WithDomainId(domainId).
                Build()
    ```
    
    3.2 Use temporary AK/SK
    
    It's preferred to obtain temporary access key, security key and security token first, which could be obtained through permanent access key and security key or through an agency.
    
    Obtaining a temporary access key token through permanent access key and security key, you could refer to document: https://support.huaweicloud.com/en-us/api-iam/iam_04_0002.html . The API mentioned in the document above corresponds to the method of createTemporaryAccessKeyByToken in IAM SDK.
    
    Obtaining a temporary access key and security token through an agency, you could refer to document: https://support.huaweicloud.com/en-us/api-iam/iam_04_0101.html . The API mentioned in the document above corresponds to the method of createTemporaryAccessKeyByAgency in IAM SDK.
    
    ``` go
    # Regional Services
    auth := basic.NewCredentialsBuilder().
                WithAk(ak).
                WithSk(sk).
                WithProjectId(projectId).
                WithSecurityToken(securityToken).
                Build()
   
    # Global Services
    auth := global.NewCredentialsBuilder().
                WithAk(ak).
                WithSk(sk).
                WithDomainId(domainId).
                WithSecurityToken(securityToken).
                Build()
    ```

4. Initialize the {Service}Client instance (Two ways)
        
    4.1 Specify Endpoint when initializing {Service}Client
    ``` go
    // Initialize specified New{Service}Client, take NewVpcClient for example
    client := vpc.NewVpcClient(
        vpc.VpcClientBuilder().
            WithEndpoint(endpoint).
            WithCredential(auth).
            WithHttpConfig(config.DefaultHttpConfig()).  
            Build())
    ```

    **where:**

    - `endpoint` is the service specific endpoints, see [Regions and Endpoints](https://developer.huaweicloud.com/intl/en-us/endpoint)

    4.2 Specify Region when initializing {Service}Client **(Recommended)**
    
    ``` go
    // Initialize specified New{Service}Client, take NewIamClient for example
    client := iam.NewIamClient(
        iam.IamClientBuilder().
            WithRegion(region.CN_NORTH_4).
            WithCredential(auth).
            WithHttpConfig(config.DefaultHttpConfig()).  
            Build())
    ```
    **where:**

    - If you use {Service}Region to initialize {Service}Client, projectId/domainId supports automatic acquisition, you don't need to configure it when initializing Credentials.
    - Multiple ProjectId situation is not supported.
    
5. Send a request and print response.

    ``` go
    // send request and print response, take interface of ListVpcs for example
    request := &model.ListVpcsRequest{
        Body: &model.ListVpcsRequestBody{
            Vpc: &model.ListVpcsOption{
                limit: 1,
            },
        },
    }
    response, err := client.ListVpcs(request)
    if err == nil {
        fmt.Printf("%+v\n",response.Vpc)
    } else {
        fmt.Println(err)
    }
    ```

6. Exceptions

    | Level 1 | Notice | 
    | :---- | :---- | 
    | ServiceResponseError | service response error |
    | url.Error | connect endpoint error |
    
    ``` go
    response, err := client.ListVpcs(request)
    if err == nil {
        fmt.Println(response)
    } else {
        fmt.Println(err)
    }
    ```

7. Original HTTP Listener

    In some situation, you may need to debug your http requests, original http request and response information will be needed. The SDK provides a listener function to obtain the original encrypted http request and response information.

    > :warning:  Warning: The original http log information are used in debugging stage only, please do not print the original http header or body in the production environment. These log information are not encrypted and contain sensitive data such as the password of your ECS virtual machine or the password of your IAM user account, etc.

    ``` go
    func RequestHandler(request http.Request) {
        fmt.Println(request)
    }
   
    func ResponseHandler(response http.Response) {
        fmt.Println(response)
    }

    client := vpc.NewVpcClient(
        vpc.VpcClientBuilder().
            WithEndpoint("{your endpoint}").
            WithCredential(
                basic.NewCredentialsBuilder().
                    WithAk("{your ak string}").
                    WithSk("{your sk string}").
                    WithProjectId("{your project id}").
                       Build()).
            WithHttpConfig(config.DefaultHttpConfig().
                WithIgnoreSSLVerification(true).
                WithHttpHandler(httphandler.
                    NewHttpHandler().
                        AddRequestHandler(RequestHandler).
                        AddResponseHandler(ResponseHandler))).
            Build())
    ```


## Code example

- The following example shows how to query a list of VPC in a specific region, you need to substitute your real `{service} "github.com/huaweicloud/huaweicloud-sdk-go-v3/services/{service}/{version}"` for `vpc "github.com/huaweicloud/huaweicloud-sdk-go-v3/services/vpc/v2"` in actual use.
- Substitute the values for `{your ak string}`, `{your sk string}`, `{your endpoint string}` and `{your project id}`.

    ``` go
    package main

    import (
        "fmt"
        "github.com/huaweicloud/huaweicloud-sdk-go-v3/core/auth/basic"
        "github.com/huaweicloud/huaweicloud-sdk-go-v3/core/config"
        "github.com/huaweicloud/huaweicloud-sdk-go-v3/core/httphandler"
        vpc "github.com/huaweicloud/huaweicloud-sdk-go-v3/services/vpc/v2"
        "github.com/huaweicloud/huaweicloud-sdk-go-v3/services/vpc/v2/model"
        "net/http"
    )

    func RequestHandler(request http.Request) {
        fmt.Println(request)
    }

    func ResponseHandler(response http.Response) {
        fmt.Println(response)
    }

    func main() {
        client := vpc.NewVpcClient(
            vpc.VpcClientBuilder().
                WithEndpoint("{your endpoint}").
                WithCredential(
                    basic.NewCredentialsBuilder().
                        WithAk("{your ak string}").
                        WithSk("{your sk string}").
                        WithProjectId("{your project id}").
                        Build()).
                WithHttpConfig(config.DefaultHttpConfig().
                    WithIgnoreSSLVerification(true).
                    WithHttpHandler(httphandler.
                        NewHttpHandler().
                            AddRequestHandler(RequestHandler).
                            AddResponseHandler(ResponseHandler))).
                Build())

        request := &model.ListVpcsRequest{}
        response, err := client.ListVpcs(request)
        if err == nil {
            fmt.Println("%+v\n",response.Vpc)
        } else {
            fmt.Println(err)
        }
    }
    ```
