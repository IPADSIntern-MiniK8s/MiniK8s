# Scheduler

## Overview

kube-scheduler 是 Kubernetes 控制平面的一个组件，负责为新创建的 Pod 选择一个合适的 Node 节点来运行。当 Kubernetes API Server 接收到创建 Pod 的请求时，会将该请求发送给 kube-scheduler，kube-scheduler 将根据一些规则和条件为该 Pod 分配一个合适的 Node。

kube-scheduler 通过以下几个步骤来为 Pod 分配 Node：

获取 Pod 的调度要求：kube-scheduler 从 Kubernetes API Server 中获取 Pod 的调度要求，包括 Pod 所需的 CPU、内存等资源，以及 Pod 的亲和性和反亲和性规则等。

执行策略：kube-scheduler 将会执行一些策略来为 Pod 选择一个 Node，例如默认策略、负载均衡策略、亲和性策略和节点亲和性策略等。

筛选 Node：kube-scheduler 将基于 Pod 调度要求和策略对集群中的每个 Node 进行筛选，以找到满足 Pod 调度要求的可用 Node。

评分和排序：kube-scheduler 会对可用的 Node 进行评分和排序，以找到最适合运行该 Pod 的 Node。kube-scheduler 根据节点资源使用情况、节点亲和性和反亲和性规则等因素对每个 Node 进行评分，然后选择最高评分的 Node。

绑定 Pod 和 Node：kube-scheduler 选择了最适合运行该 Pod 的 Node 之后，会向 Kubernetes API Server 发送一个绑定请求，将该 Pod 绑定到所选的 Node 上，Kubernetes API Server 将更新该 Pod 的状态并通知 kubelet 在相应的 Node 上创建并运行该 Pod。

## Pod的调度需求

Pod 的调度需求可以通过 Pod 的配置信息来体现。以下是一些 Pod 配置中可以影响 Pod 调度需求的字段和它们的举例：

资源需求（Resource Requirements）：Pod 可以通过容器的资源需求来指定它需要的 CPU 和内存资源。这些需求可以帮助 Kubernetes 调度器决定将 Pod 调度到哪个节点上。
举例：

```yaml
apiVersion: v1
kind: Pod
metadata:
  name: nginx-pod
spec:
  containers:
    - name: nginx
      image: nginx
      resources:
        requests:
          memory: "64Mi"
          cpu: "250m"
        limits:
          memory: "128Mi"
          cpu: "500m"

```

在上面的示例中，容器 nginx 的资源需求分别为 250m CPU 和 64Mi 内存，而其资源限制分别为 500m CPU 和 128Mi 内存。

调度限制（Node Selector）：Pod 可以通过配置调度限制来规定哪些节点可以或不能调度该 Pod。这些限制可以基于节点的标签、容量或亲和性等属性。
举例：

```yaml
apiVersion: v1
kind: Pod
metadata:
  name: nginx-pod
spec:
  nodeSelector:
    disktype: ssd
  containers:
    - name: nginx
      image: nginx
```

在上面的示例中，Pod nginx-pod 的调度限制是其节点必须有一个 disktype 标签值为 ssd。

亲和性和反亲和性（Affinity and Anti-Affinity）：Pod 可以通过配置亲和性和反亲和性规则来指定它应该调度到哪个节点上或不能调度到哪个节点上。这些规则可以基于节点的标签、容量或已运行的 Pod 等属性。
举例：

```yaml
apiVersion: v1
kind: Pod
metadata:
  name: nginx-pod
spec:
  affinity:
    nodeAffinity:
      requiredDuringSchedulingIgnoredDuringExecution:
        nodeSelectorTerms:
        - matchExpressions:
          - key: node-role.kubernetes.io/worker
            operator: Exists
  containers:
    - name: nginx
      image: nginx
```

在上面的示例中，Pod nginx-pod 要求调度到标记有 node-role.kubernetes.io/worker 标签的节点上。

容器亲和性和反亲和性（Affinity and Anti-Affinity）：Pod 中的容器可以通过配置亲和性和反亲和性规则来指定它应该调度到哪个节点上或不能调度到哪个节点上。这些规则可以基于节点的标签、容量或已运行的容器等属性。
举例：

```yaml
apiVersion: v1
kind: Pod
metadata:
  name: nginx-pod
spec:
  containers:
    - name: nginx
      image: nginx
      resources:
        requests:
          memory: "64Mi"
          cpu: "250m"
      env:
        - name: ZONE
          value: "eu-west-1a"
      affinity:
        podAffinity:
          requiredDuringSchedulingIgnoredDuringExecution:
          - labelSelector:
              matchExpressions:
                - key: app
                  operator: In
                  values:
                  - nginx
            topologyKey: "kubernetes.io/hostname"
```

在上面的示例中，Pod 中的容器 nginx 只能调度到在 eu-west-1a 区域中且已经运行了一个 app=nginx 标签的 Pod 的节点上。

调度器扩展程序（Scheduler Extender）：通过实现调度器扩展程序可以在 Kubernetes 默认调度器的基础上增加一些额外的调度算法或逻辑。扩展程序可以接收到调度请求并决定该请求应该被哪个节点处理。
举例：

调度器扩展程序可以基于节点的健康状况或其他第三方条件来决定节点是否适合调度某个 Pod。

总之，Pod 的调度需求可以通过多种方式体现在 Pod 的配置中，调度器会根据这些需求来选择最合适的节点来运行 Pod。

### 目前支持的filter策略

#### configfilter

1. 如果Pod的`NodeSelector`字段不为空，首先用这个字段与Node的`MetaData`的`Label`进行匹配
2. 如果Pod的`Resources`不为空，用这个字段匹配`NodeStatus`中的`Allocatable`,判断是否满足（注意，如果Node相应的字段为空的话，这里不会过滤掉）
3. 目前不支持亲和性
