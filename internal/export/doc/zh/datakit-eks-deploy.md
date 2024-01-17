
# Amazon EKS 集成
---

[Amazon Elastic Kubernetes Services (Amazon EKS)](https://aws.amazon.com/eks/){:target="_blank"}  是一项托管容器服务，用于在 AWS 云中运行和扩展 Kubernetes 应用程序。DataKit 为 Amazon EKS 集群提供按命名空间、集群、Pod 不同维度的可观测。客户可以使用现有的 AWS 支持协议来获取支持。


## 架构图 {#architecture-overview}

<figure markdown>
  ![](https://static.guance.com/images/datakit/datakit-eks-architecture-overview.png){ width="800" }
  <figcaption>架构图</figcaption>
</figure>


## 使用 Helm 在 Amazon EKS 集群上部署 Datakit {#helm-install}

### 前置条件 {#prerequisites-helm-install}

- 安装以下工具：[Helm 3.7.1](https://github.com/helm/helm/releases/tag/v3.7.1){:target="_blank"}, [kubectl](https://kubernetes.io/docs/tasks/tools/){:target="_blank"} 和 [AWS CLI](https://aws.amazon.com/cli/){:target="_blank"} 。
- 您有权访问 [Amazon EKS 集群](https://aws.amazon.com/eks/){:target="_blank"} 。
- 你需要提前获取 `DK_DATAWAY`， 您还可以按照以下说明获取：
  1. 进入 [观测云](https://en.guance.com/){:target="_blank"} 网站，参考 [注册](https://docs.guance.com/en/billing/commercial-register/){:target="_blank"} 指南成为 GuanCe 用户。
  2. 点击「集成」菜单，然后选择 「DataKit」页签，复制 `DK_DATAWAY` 参数 如下图：


<figure markdown>
  ![](https://static.guance.com/images/datakit/datakit-eks-zh-get-datawayurl.png){ width="800" }
  <figcaption>复制地址</figcaption>
</figure>
  

### 登录 ECR 仓库 {#login-ecr}

```shell
export HELM_EXPERIMENTAL_OCI=1

aws ecr get-login-password \
    --region us-east-1 | helm registry login \
    --username AWS \
    --password-stdin 709825985650.dkr.ecr.us-east-1.amazonaws.com
```

### Helm 安装（升级） DataKit {#helm-install}

<!-- markdownlint-disable MD046 -->
???+ attention "注意事项"

    Helm 版本必须是 3.7.1
    `datakit.datawayUrl` 必须要修改。

<!-- markdownlint-enable -->

```shell

helm upgrade -i datakit oci://709825985650.dkr.ecr.us-east-1.amazonaws.com/guance/datakit-charts --version 1.20.0 \
     --create-namespace -n datakit \
     --set datakit.datawayUrl="https://us1-openway.guance.com?token=<your-token>"

```

期望输出结果：

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

### 验证部署 {#verify-install}

- 获取部署状态

```shell
helm ls -n datakit
```

期望输出结果：

```txt
datakit  datakit  1  2024-01-12 14:50:07.880846 +0800 CST  deployed  datakit-1.20.0  1.20.0
```

- 观测云平台验证

<figure markdown>
  ![](https://static.guance.com/images/datakit/datakit-eks-zh-verify.png){ width="800" }
  <figcaption>验证</figcaption>
</figure>

## 扩展阅读 {#more-reading}

[K8s 安装](datakit-daemonset-deploy.md)
