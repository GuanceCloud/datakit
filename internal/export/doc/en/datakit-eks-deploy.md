
# Amazon EKS Integration
---

[Amazon Elastic Kubernetes Services (Amazon EKS)](https://aws.amazon.com/eks/){:target="_blank"} is a managed container service to run and scale Kubernetes applications in the AWS cloud. For running and extending Kubernetes applications in the AWS cloud. DataKit provides observations for the Amazon EKS Cluster in different dimensions by namespace, cluster, Pod.Customers can use their existing AWS support agreements to obtain support.

<!-- markdownlint-disable MD013 -->
## Architecture overview {#architecture-overview}
<!-- markdownlint-enable -->

![overview](https://static.<<<custom_key.brand_main_domain>>>/images/datakit/datakit-eks-architecture-overview.png){:target="_blank"}


## Using Amazon EKS add-on {#add-on-install}

Deploying DataKit on an Amazon EKS cluster using Amazon EKS add-on.

### Prerequisites {#prerequisites-addon-install}

- Subscribe on the AWS Marketplace [<<<custom_key.brand_name>>> Container Agent](https://aws.amazon.com/marketplace/pp/prodview-tdwkw3qcsimso?sr=0-2&ref_=beagle&applicationId=AWSMPContessa){:target="_blank"} 。
- You have access to an [Amazon EKS Cluster](https://aws.amazon.com/eks/){:target="_blank"} .
- You need to get it in advance `DK_DATAWAY`. You can also obtain it by following the instructions below:
    - You can go to [<<<custom_key.brand_name>>> official website](https://www.guance.one/){:target="_blank"}, [register now](https://auth.<<<custom_key.brand_main_domain>>>/en/businessRegister){:target="_blank"} as a user.
    - Click the 「Integration」 menu, then select the 「DataKit」 TAB and copy the `DK_DATAWAY` parameter
     ![`datakit-eks-en-get-datawayur`](https://static.<<<custom_key.brand_main_domain>>>/images/datakit/datakit-eks-en-get-datawayurl.png){:target="_blank"}

<!-- markdownlint-disable MD046 -->  
=== "Enable DataKit add-on from AWS console"

    - Search Add-ons
      First, in the Amazon EKS Console, go to your EKS cluster and select "Get more Add-ons" on the "add-ons" TAB to find the new third-party EKS add-ons in the cluster Settings of the existing EKS cluster. And search for 'datakit', select '<<<custom_key.brand_name>>> Container Agent', next step.
    
    <figure markdown>
      ![](https://static.<<<custom_key.brand_main_domain>>>/images/datakit/eks-install/get-more-addon.png){ width="800" }
      <figcaption></figcaption>
    </figure>
    
    <figure markdown>
      ![](https://static.<<<custom_key.brand_main_domain>>>/images/datakit/eks-install/search-datakit-addon.png){ width="800" }
      <figcaption></figcaption>
    </figure>

    - Confirm installation
      Select the latest version to install.
    
    <figure markdown>
      ![](https://static.<<<custom_key.brand_main_domain>>>/images/datakit/eks-install/select-install-addon.png){ width="800" }
      <figcaption></figcaption>
    </figure>    
        
    <figure markdown>
      ![](https://static.<<<custom_key.brand_main_domain>>>/images/datakit/eks-install/install-datakit-addon.png){ width="800" }
      <figcaption></figcaption>
    </figure>    

=== "Enable DataKit add-on using AWS CLI"

    ???+ tip
         You need to replace `$YOUR_CLUSTER_NAME` and `$AWS_REGION` accordingly with your actual Amazon EKS cluster name and AWS region.
        
    Install：
    
    ```shell
    aws eks create-addon --addon-name guance_datakit --cluster-name $YOUR_CLUSTER_NAME --region $AWS_REGION
    ```
    
    verify：
    
    ```shell
    aws eks describe-addon --addon-name guance_datakit --cluster-name $YOUR_CLUSTER_NAME --region $AWS_REGION
    ```
<!-- markdownlint-enable -->


### configuration DataKit {#config-addon-datakit}


Set the `token` environment variable:

```shell
token="https://us1-openway.<<<custom_key.brand_main_domain>>>?token=<YOUR-WORKSPACE-TOKEN>"
```

Add token to `env-dataway` secrets:

```shell
envDataway=$(echo -n "$token" | base64)
kubectl patch secret env-dataway -p "{\"data\": {\"datawayUrl\": \"$envDataway\"}}" -n datakit
```

restart DataKit:

```shell
kubectl rollout restart ds datakit -n datakit
```


### Verify Deployment {#verify-addon-install}

- Get deployment status

```shell
helm ls -n datakit
```

Expected output：

```shell
NAME  NAMESPACE REVISION  UPDATED STATUS  CHART APP VERSION
datakit dataki  1 2024-01-12 14:50:07.880846 +0800 CST  deployed  datakit-1.20.0  1.20.0
```

- <<<custom_key.brand_name>>> verify

![`datakit-eks-en-verify`](https://static.<<<custom_key.brand_main_domain>>>/images/datakit/datakit-eks-en-verify.png){:target="_blank"}


<!-- markdownlint-disable MD013 -->
## Deploying DataKit on an Amazon EKS Cluster using Helm {#helm-install}
<!-- markdownlint-enable -->

### Prerequisites {#prerequisites-helm-install}

- Install the following tools: [Helm 3.7.1](https://github.com/helm/helm/releases/tag/v3.7.1){:target="_blank"}, [kubectl](https://kubernetes.io/docs/tasks/tools/){:target="_blank"}, and [AWS CLI](https://aws.amazon.com/cli/){:target="_blank"} .
- You have access to an [Amazon EKS Cluster](https://aws.amazon.com/eks/){:target="_blank"} .
- You need to get it in advance `DK_DATAWAY`. You can also obtain it by following the instructions below:
    - You can go to [<<<custom_key.brand_name>>> official website](https://www.guance.one/){:target="_blank"}, [register now](https://auth.<<<custom_key.brand_main_domain>>>/en/businessRegister){:target="_blank"} as a user.
    - Click the 「Integration」 menu, then select the 「DataKit」 TAB and copy the `DK_DATAWAY` parameter
     ![`datakit-eks-en-get-datawayur`](https://static.<<<custom_key.brand_main_domain>>>/images/datakit/datakit-eks-en-get-datawayurl.png){:target="_blank"}

### Login to the ECR Registry {#login-ecr}

```shell
export HELM_EXPERIMENTAL_OCI=1

aws ecr get-login-password \
    --region us-east-1 | helm registry login \
    --username AWS \
    --password-stdin 709825985650.dkr.ecr.us-east-1.amazonaws.com
```

### Helm installation (Upgrade) DataKit {#helm-install}

<!-- markdownlint-disable MD046 -->
???+ note

    - Helm Version must be 3.7.1.
    - `datakit.datawayUrl` Must be modified.
<!-- markdownlint-enable -->

```shell
helm upgrade -i datakit oci://709825985650.dkr.ecr.us-east-1.amazonaws.com/guance/datakit-charts --version 1.23.5 \
     --create-namespace -n datakit
```

Expected output：

```shell
Release "datakit" does not exist. Installing it now.
Warning: chart media type application/tar+gzip is deprecated
Pulled: 709825985650.dkr.ecr.us-east-1.amazonaws.com/guance/datakit-charts:1.23.5
Digest: sha256:04ce9e0419d8f19898a5a18cda6c35f0ff82cf63e0d95c8693ef0a37ce9d8348
NAME: datakit
LAST DEPLOYED: Fri Jan 12 14:50:07 2024
NAMESPACE: datakit
STATUS: deployed
REVISION: 1
NOTES:
1. Get the application URL by running these commands:
  export POD_NAME=$(kubectl get pods --namespace datakit -l "app.kubernetes.io/name=datakit,app.kubernetes.io/instance=datakit" -o jsonpath="{.items[0].metadata.name}")
  export CONTAINER_PORT=$(kubectl get pod --namespace datakit $POD_NAME -o jsonpath="{.spec.containers[0].ports[0].containerPort}")
  echo "Visit http://127.0.0.1:9527 to use your application"
  kubectl --namespace datakit port-forward $POD_NAME 9527:$CONTAINER_PORT
```

### configuration DataKit {#config-datakit}


Set the `token` environment variable:

```shell
token="https://us1-openway.<<<custom_key.brand_main_domain>>>?token=<YOUR-WORKSPACE-TOKEN>"
```

Add token to `env-dataway` secrets:

```shell
envDataway=$(echo -n "$token" | base64)
kubectl patch secret env-dataway -p "{\"data\": {\"datawayUrl\": \"$envDataway\"}}" -n datakit
```

restart DataKit:

```shell
kubectl rollout restart ds datakit -n datakit
```

### Verify Deployment {#verify-install}

- Get deployment status

```shell
helm ls -n datakit
```

Expected output：

```shell
NAME  NAMESPACE REVISION  UPDATED STATUS  CHART APP VERSION
datakit dataki  1 2024-01-12 14:50:07.880846 +0800 CST  deployed  datakit-1.20.0  1.20.0
```

- <<<custom_key.brand_name>>> verify

![`datakit-eks-en-verify`](https://static.<<<custom_key.brand_main_domain>>>/images/datakit/datakit-eks-en-verify.png){:target="_blank"}


## More Readings {#more-reading}

[K8s deploy](datakit-daemonset-deploy.md)
