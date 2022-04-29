# DataKit

[DataKit](https://www.yuque.com/dataflux/datakit) is a hosted infrastructure monitoring platform. This chart adds the DataKit Agent to all nodes in your cluster via a DaemonSet. It also optionally depends on the [kube-state-metrics chart](https://github.com/kubernetes/charts/tree/master/stable/kube-state-metrics). For more information about monitoring Kubernetes with DataKit, please refer to the [DataKit documentation website](https://docs.datadoghq.com/agent/basic_agent_usage/kubernetes/).

## Prerequisites

- Kubernetes 1.14+

- Helm 3.0+

## Quick start

By default, the DataKit Agent runs in a DaemonSet. It can alternatively run inside a Deployment for special use cases.

### Installing the DataKit Chart

To install the chart with the release name `datakit`, retrieve your DataWay url from your [Agent Installation Instructions](https://auth.guance.com/) and run:

```bash
helm install --name datakit \
  --set datakit.dataway_url=<DATAWAY_URL> datakit/datakit
```

