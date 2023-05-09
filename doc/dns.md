# DNS
## Overview
### 主要功能
Kubernetes中的DNS主要有以下几个功能：

1. 服务发现

Kubernetes中的DNS服务可以帮助应用程序发现其他服务。当应用程序需要访问其他服务时，可以使用服务名称作为DNS记录的别名，而不需要知道服务的IP地址和端口号。在Kubernetes集群中，每个Service对象都有一个DNS记录，可以通过服务名称访问到。

2. 域名解析

Kubernetes中的DNS服务可以解析不同对象的域名，包括服务、Pod和ServiceAccount等。这使得在集群内部通信时，不需要使用硬编码的IP地址和端口号，而可以使用相应对象的DNS名称。

3. 集群内部DNS解析

Kubernetes中的DNS服务可以帮助解析Kubernetes内部的DNS名称，例如节点名称、服务IP地址和端口等。

4. 集群外部DNS解析

Kubernetes中的DNS服务还可以帮助解析集群外部的DNS名称，例如解析外部服务的DNS名称或者解析公共DNS记录。

5. 可扩展性

Kubernetes中的DNS服务使用了可扩展性的设计，可以支持多种不同的DNS插件，如CoreDNS、KubeDNS等，而且可以自定义域名后缀，以适应不同的网络拓扑结构和部署场景。

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
  