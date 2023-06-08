# Virtual Internet Access

---

## Overview {#overview}

If your Guance Cloud access system is deployed on Alibaba Cloud, this guide will guide you to access the data gateway private network from your host DataKit to the Guance cloud platform by subscribing to the computing nest “**Guance Cloud Data Gateway Virtual Internet**” service.

> Only available for Alibaba Cloud users for now.

Advantages of establishing private network connection:

- **Higher Bandwidth**: It realizes higher bandwidth through the virtual Internet without occupying the public network bandwidth of the business system.
- **Safer**: The data does not pass through the public network, but completely flows in Alibaba Cloud's private network, leading to safer data.
- **Lower Cost**: Compared with the high cost of public network bandwidth, the cost of virtual Internet is lower.

At present, the services that have been released are **cn-hangzhou and cn-beijing**, and those in other regions will be put on shelves soon. The structure is as follows:

![](imgs/aliyun_1.png)

## Subscription Service {#sub-service}

### Service Deployment Link {#service-dep}

| **Access Site** | **Region Where the Server is Located** | **Compute Nest Deployment Link** |
| -------- | ---------------------- | ----------- |
| China 1-hangzhou | cn-hangzhou | [Guance Cloud Data Gateway Virtual Internet-Hangzhou](https://computenest.console.aliyun.com/user/cn-hangzhou/serviceInstanceCreate?ServiceId=service-68c8fee7f0554d6b9baa){:target="_blank"} |
| China 1-hangzhou | cn-beijing | [Guance Cloud Data Gateway Virtual Internet-from Beijing to Hangzhou](https://computenest.console.aliyun.com/user/cn-hangzhou/serviceInstanceCreate?ServiceId=service-af3b4511d9214c9ebaba){:target="_blank"} |  
| China 3-Zhangjiakou | cn-beijing | [Guance Cloud Data Gateway Virtual Internet-from Beijing to Zhangjiakou](https://computenest.console.aliyun.com/user/cn-hangzhou/serviceInstanceCreate?ServiceId=service-a22bc59ed53c4946b8ce){:target="_blank"} | 
| China 3-Zhangjiakou | cn-hangzhou | [Guance Cloud Data Gateway Virtual Internet-from Hangzhou to Zhangjiakou](https://computenest.console.aliyun.com/user/cn-hangzhou/serviceInstanceCreate?ServiceId=service-87a611279d9a42ceaeb2){:target="_blank"} | 

### Private Network Data Gateway Default Endpoint for Different Regions  {#region-endpoint}

| **Access Site** | **Region Where the Server is Located** | **Endpoint** |
| -------- | ---------------------- | ----------- |
| China 1-Hangzhou | cn-hangzhou | https://openway.guance.com  |
| China 1-Hangzhou | cn-beijing | https://beijing-openway.guance.com |  
| China 3-Zhangjiakou | cn-beijing | https://cn3-openway.guance.com | 
| China 3-Zhangjiakou | cn-hangzhou | https://cn3-openway.guance.com | 

**Virtual Internet services in other regions will be released soon.**

### Configure Service Subscriptions {#config-sub}
Sign in with your Alibaba Cloud account and open the above **service deployment link** to subscribe to our virtual Internet service, taking cn-hangzhou as an example:

![](imgs/aliyun_2.png)

1. Select the subscription region first, which must be the same region as the cloud resources deployed by the system you want to access Guance Cloud.
1. Select the same VPC network of cloud resources deployed by the system to be accessed. **If multiple VPCs need to access the virtual Internet, they can subscribe for many times, and each VPC subscribes once.**
1. Select an installation group.
1. Available areas and switches. If multiple available areas and switches are involved, multiple can be added.
1. Select "Use Recommended Custom Domain Name" and use the default recommended domain name, such as the openway.guance.com domain name for cn-hanghou.

Using the default openway.guance.com domain name, it is important to seamlessly switch the data network to a virtual intranet if DataKit has been deployed and implemented within the same VPC.

### Subscription Completion {#sub-com}

After the subscription is completed, the compute nest service will automatically create and configure it under your cloud account:

1. A private network is connected to the terminal node;
2. A cloud-resolved Private Zone that resolves the Endpoint domain name to the default domain.

### Cost {#cost}

The cost situation is mainly divided into two parts:

1. The first part is the private network access fee directly paid by Alibaba Cloud to your Alibaba Cloud account, mainly including the fees for private network connection PrivateLink and cloud analysis PrivateZone services. Refer to Alibaba Cloud official website [Private Network Connection PrivateLink Billing Instructions](https://help.aliyun.com/document_detail/198081.html){:target="_blank"} and [Cloud Parsing PrivateZone Billing Instructions](https://help.aliyun.com/document_detail/71338.html){:target="_blank"}.
2. The second part is the cross-regional network transmission traffic fee. If your Alibaba Cloud resource is connected to the Guance Cloud Hangzhou site in Beijing Region, it will generate cross-regional traffic transmission fee, which will be paid out to your Guance Cloud bill.

## How to Use {#how-to}

After the subscription is completed, it is completely transparent to your DataKit access Guance Cloud. There is no need to modify the DataKit configuration, and the private network connection has been automatically established. You can log in to the cloud host and execute the following ping openway.guance.com command to view the IP pinged. If it is an intranet IP address, it means that you have successfully established a private network connection with the Guance Cloud data gateway:

![](imgs/aliyun_3.png)
