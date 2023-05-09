# DNS
## Overview
### DNS 记录
以下对象会获得 DNS 记录：
- Services
- Pods

#### 记录格式
- Pod 
  - A/AAAA 记录
    一般而言，Pod 会对应如下 DNS 名字解析：
    ```shell
     pod-ip-address.my-namespace.pod.cluster-domain.example
    ```
    例如，对于一个位于 default 名字空间，IP 地址为 172.17.0.3 的 Pod， 如果集群的域名为 cluster.local，则 Pod 会对应 DNS 名称：
    ```shell
    172-17-0-3.default.pod.cluster.local
    ```
    通过 Service 暴露出来的所有 Pod 都会有如下 DNS 解析名称可用：

    pod-ip-address.service-name.my-namespace.svc.cluster-domain.example

  - Pod 的 hostname 和 subdomain 字段
      当前，创建 Pod 时其主机名（从 Pod 内部观察）取自 Pod 的 metadata.name 值。
      Pod 规约中包含一个可选的 hostname 字段，可以用来指定一个不同的主机名。 当这个字段被设置时，它将优先于 Pod 的名字成为该 Pod 的主机名（同样是从 Pod 内部观察）。 举个例子，给定一个 spec.hostname 设置为 “my-host” 的 Pod， 该 Pod 的主机名将被设置为 “my-host”。
      Pod 规约还有一个可选的 subdomain 字段，可以用来表明该 Pod 是名字空间的子组的一部分。 举个例子，某 Pod 的 spec.hostname 设置为 “foo”，spec.subdomain 设置为 “bar”， 在名字空间 “my-namespace” 中，主机名称被设置成 “foo” 并且对应的完全限定域名（FQDN）为 “foo.bar.my-namespace.svc.cluster-domain.example”（还是从 Pod 内部观察）。
      如果 Pod 所在的名字空间中存在一个无头服务，其名称与子域相同， 则集群的 DNS 服务器还会为 Pod 的完全限定主机名返回 A 和/或 AAAA 记录。
  - 