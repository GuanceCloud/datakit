# AWS Private Internet Access

---

## Overview {#overview}

Amazon PrivateLink is a highly available and scalable technology that enables you to connect your VPC privately to services as if these services were within your own VPC. You do not need to use an internet gateway, NAT device, public IP addresses, Amazon Direct Connect connection, or Amazon Site-to-Site VPN connection to allow communication with services in private subnets. Therefore, you can control specific API endpoints, sites, and services accessible from the VPC. Amazon PrivateLink can help you save on some traffic costs.

Benefits of establishing a private network connection:

- **Higher Bandwidth**: Does not consume public bandwidth of business systems, achieving higher bandwidth through endpoint services.
- **More Secure**: Data does not pass through the public internet, ensuring data remains within the private network for enhanced security.
- **Lower Costs**: Compared to high fees for public bandwidth, the cost of virtual internet access is lower.

The architecture is as follows:

```mermaid
flowchart LR
  subgraph Customer_VPC
    dk_a[Availability Zone A - dk]
    dk_b[Availability Zone B - dk]
    dk_c[Availability Zone C - dk]
    plc[Endpoints]

    dk_a --> plc
    dk_b --> plc
    dk_c --> plc
  end

  subgraph <<<custom_key.brand_key>>>_VPC
    pls[Endpoints Service]
    nlb[NLB]
    dw[DW - Availability Zone C]
    pls --> nlb --> dw
  end
  plc --> pls
```


## Prerequisites {#prerequisite}

1. First, select the subscription region, which must match the region where your cloud resources for <<<custom_key.brand_name>>> are deployed.
2. Choose the same VPC network where your cloud resources are deployed. **If multiple VPCs need to connect to the endpoint service, subscribe multiple times, once for each VPC.**

## Subscribe to Service {#sub-service}

### Service Deployment Links {#service-dep}

<<<% if custom_key.brand_key == "truewatch" %>>>
| **Access Region** | **Your Server's Region**       | **Endpoint Service Name**                                      |
| ----------------- | ------------------------------ | -------------------------------------------------------------- |
| Asia-Pacific Region 1 (Singapore) | `ap-southeast-1` (Singapore) | `com.amazonaws.vpce.ap-southeast-1.vpce-svc-08465b643241dce58` |
<<<% else %>>>
| **Access Region** | **Your Server's Region**       | **Endpoint Service Name**                                      |
| ----------------- | ------------------------------ | -------------------------------------------------------------- |
| China Region 2 (Ningxia) | `cn-northwest-1` (Ningxia) | `cn.com.amazonaws.vpce.cn-northwest-1.vpce-svc-070f9283a2c0d1f0c` |
| Overseas Region 1 (Oregon) | `us-west-2` (Oregon)     | `com.amazonaws.vpce.us-west-2.vpce-svc-084745e0ec33f0b44`      |
| Asia-Pacific Region 1 (Singapore) | `ap-southeast-1` (Singapore) | `com.amazonaws.vpce.ap-southeast-1.vpce-svc-070194ed9d834d571` |
<<<% endif %>>>



### Default Endpoint for Private Network Gateway {#region-endpoint}

<<<% if custom_key.brand_key == "truewatch" %>>>
| **Access Region** | **Your Server's Region**       | **Endpoint**                                                |
| ----------------- | ------------------------------ | ------------------------------------------------------------ |
| Asia-Pacific Region 1 (Singapore) | `ap-southeast-1` (Singapore) | `https://ap1-openway.<<<custom_key.brand_main_domain>>>`                            |
<<<% else %>>>
| **Access Region** | **Your Server's Region**       | **Endpoint**                                                |
| ----------------- | ------------------------------ | ------------------------------------------------------------ |
| China Region 2 (Ningxia) | `cn-northwest-1` (Ningxia) | `https://aws-openway.<<<custom_key.brand_main_domain>>>`                             |
| Overseas Region 1 (Oregon) | `us-west-2` (Oregon)     | `https://us1-openway.<<<custom_key.brand_main_domain>>>`                             |
| Asia-Pacific Region 1 (Singapore) | `ap-southeast-1` (Singapore) | `https://ap1-openway.<<<custom_key.brand_main_domain>>>`                            |
<<<% endif %>>>

### Configure Service Subscription {#config-sub}

#### Step One: Authorize Account ID {#accredit-id}
<!-- markdownlint-disable MD032 -->
Open the Amazon console via the following links:

<<<% if custom_key.brand_key == "truewatch" %>>>
- [Console Web](https://console.aws.amazon.com/console/home){:target="_blank"}
<<<% else %>>>
- [China Region](https://console.amazonaws.cn/console/home){:target="_blank"}
- [Overseas Region](https://console.aws.amazon.com/console/home){:target="_blank"}
<<<% endif %>>>
Obtain the account ID in the upper right corner of the console, copy this "Account ID," and **inform** our customer manager at <<<custom_key.brand_name>>> to add it to our whitelist.


#### Step Two: Create Endpoint {#create-endpoint}

1. Open the Amazon VPC console via the following links:
<<<% if custom_key.brand_key == "truewatch" %>>>
   - [VPC](https://console.amazonaws.cn/vpc/){:target="_blank"}
<<<% else %>>>
   - [China Region](https://console.amazonaws.cn/vpc/){:target="_blank"}
   - [Overseas Region](https://console.amazonaws.cn/vpc/){:target="_blank"}
<<<% endif %>>>
<!-- markdownlint-disable MD051 -->
1. Create Security Group:
    - Security group name: private-link
    - Inbound Rules Type: HTTPS
    - Source: 0.0.0.0/0
1. In the navigation pane, select **Endpoint** (Endpoint Service).
1. Create Endpoint
    - Endpoint settings
        - Type: **Endpoint services that use NLBs and GWLBs**
    - Service settings
        - Service name: The current AZ [Service Deployment Links](#service-dep){:target="_blank"}
        - Verify service
    - Network settings
        - VPC: VPC for business services
        - Subnets: Select the business Subnets
        - Security Group: private-link
<!-- markdownlint-enable -->
1. Notify the account manager of <<<custom_key.brand_name>>> for review
1. Wait for the creation to be successful, click on "Operations" of the terminal node - "Modify Private DNS Name", and set 'Enable DNS name'
<!-- markdownlint-enable -->
#### Verification {#verify}

Run the following command on EC2:

```shell
dig us1-openway.<<<custom_key.brand_main_domain>>>
```

Result:

```shell
; <<>> DiG 9.16.38-RH <<>> us1-openway.<<<custom_key.brand_main_domain>>>
;; global options: +cmd
;; Got answer:
;; ->>HEADER<<- opcode: QUERY, status: NOERROR, id: 22545
;; flags: qr rd ra; QUERY: 1, ANSWER: 2, AUTHORITY: 0, ADDITIONAL: 1

;; OPT PSEUDOSECTION:
; EDNS: version: 0, flags:; udp: 4096
;; QUESTION SECTION:
;us1-openway.<<<custom_key.brand_main_domain>>>.      IN    A

;; ANSWER SECTION:
us1-openway.<<<custom_key.brand_main_domain>>>. 296 IN  CNAME     172.31.16.128 

;; Query time: 0 msec
;; SERVER: 172.31.0.2#53(172.31.0.2)
;; WHEN: Thu May 18 11:23:04 UTC 2023
;; MSG SIZE  rcvd: 176
```

### Cost Details {#cost}

Taking Oregon as an example:

| Name                                                        | Cost     | Documentation                                                         | Notes                   |
| ----------------------------------------------------------- | -------- | --------------------------------------------------------------------- | ----------------------- |
| Data transfer out from Amazon EC2 to the internet            | $0.09/GB | [Documentation](https://aws.amazon.com/cn/ec2/pricing/on-demand/#Data_Transfer){:target="_blank"} | Charged by traffic      |
| Interface VPC endpoint                                       | $0.01/H  | [Documentation](https://aws.amazon.com/cn/privatelink/pricing/?nc1=h_ls){:target="_blank"} | Charged by AZ and hour  |
| Data transfer out from interface VPC endpoint                | $0.01/GB | [Documentation](https://aws.amazon.com/cn/privatelink/pricing/?nc1=h_ls){:target="_blank"} | Charged by traffic      |

The main cost components are:

1. Interface VPC endpoint [service charges](https://aws.amazon.com/cn/privatelink/pricing/?nc1=h_ls){:target="_blank"}
2. Traffic charges for the endpoint

Comparison:

Assuming the client transmits **200GB** of outbound traffic and **10GB** of inbound traffic **daily**:

|          |              Internet               | PrivateLink                                                  |
| :------: | :---------------------------------: | ------------------------------------------------------------ |
| Formula  | Internet Outbound Traffic × Internet Outbound Traffic Fee × 30 | Interface VPC Endpoint Service × 3 Availability Zones × 24 Hours × 30 Days + (Interface VPC Endpoint Outbound Traffic Fee × Interface VPC Endpoint Outbound Traffic + Interface VPC Endpoint Inbound Traffic Fee × Interface VPC Endpoint Inbound Traffic) × 30 |
| Calculation |         0.09 × 200 × 30           | 0.01 × 3 × 24 × 30 + (0.01 × 200 + 0.01 × 10) × 30 |
| Monthly Total |             $540.0              | $84.6 |
