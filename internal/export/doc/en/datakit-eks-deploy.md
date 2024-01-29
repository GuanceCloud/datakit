
# Amazon EKS Integration
---

[Amazon Elastic Kubernetes Services (Amazon EKS)](https://aws.amazon.com/eks/){:target="_blank"} is a managed container service to run and scale Kubernetes applications in the AWS cloud. For running and extending Kubernetes applications in the AWS cloud. DataKit provides observations for the Amazon EKS Cluster in different dimensions by namespace, cluster, Pod.Customers can use their existing AWS support agreements to obtain support.

<!-- markdownlint-disable MD013 -->
## Architecture overview {#architecture-overview}
<!-- markdownlint-enable -->

![overview](https://static.guance.com/images/datakit/datakit-eks-architecture-overview.png){:target="_blank"}
<!-- markdownlint-disable MD013 -->
## Deploying DataKit on an Amazon EKS Cluster using Helm {#helm-install}
<!-- markdownlint-enable -->
### Prerequisites {#prerequisites-helm-install}

- Install the following tools: [Helm 3.7.1](https://github.com/helm/helm/releases/tag/v3.7.1){:target="_blank"}, [kubectl](https://kubernetes.io/docs/tasks/tools/){:target="_blank"}, and [AWS CLI](https://aws.amazon.com/cli/){:target="_blank"} .
- You have access to an [Amazon EKS Cluster](https://aws.amazon.com/eks/){:target="_blank"} .
- You need to get it in advance `DK_DATAWAY`. You can also obtain it by following the instructions below:
  1. You can go to [Guance official website](https://www.guance.one/){:target="_blank"}, [register now](https://auth.guance.com/en/businessRegister){:target="_blank"} as a Guance user.
  2. Click the 「Integration」 menu, then select the 「DataKit」 TAB and copy the `DK_DATAWAY` parameter
     ![`datakit-eks-en-get-datawayur`](https://static.guance.com/images/datakit/datakit-eks-en-get-datawayurl.png){:target="_blank"}

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
???+ attention "Attention"

    - Helm Version must be 3.7.1.
    - `datakit.datawayUrl` Must be modified.
<!-- markdownlint-enable -->

```shell
helm upgrade -i datakit oci://709825985650.dkr.ecr.us-east-1.amazonaws.com/guance/datakit-charts --version 1.20.0 \
     --create-namespace -n datakit \
     --set datakit.datawayUrl="https://us1-openway.guance.com?token=<your-token>"

```

Expected output：

```shell
Release "datakit" does not exist. Installing it now.
Warning: chart media type application/tar+gzip is deprecated
Pulled: 709825985650.dkr.ecr.us-east-1.amazonaws.com/guance/datakit-charts:1.20.0
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

- Guance Cloud verify

![`datakit-eks-en-verify`](https://static.guance.com/images/datakit/datakit-eks-en-verify.png){:target="_blank"}

## More Readings {#more-reading}

[K8s deploy](datakit-daemonset-deploy.md)
