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
     【pod-ip-address】.【命名空间】.pod.cluster.local  
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
- Service

  - kubernetes会为Service创建域名，其域名格式为

    ```shell
    【servic名称】.【命名空间】.svc.cluster.local  
    ```

    cluster.local是k8s默认的集群域

    普通Service的DNS记录是Service本身的IP
    无头Service（Headless Service）的DNS记录则是其选择的 Pod IP 的集合，（无头Service的名称与Pod中配置subdomain一致

### DNS注册

在 Kubernetes 中，每个 Pod 都有一个 DNS 名称，称为 Pod DNS 名称。Pod DNS 名称由以下部分组成：

```shell
pod-ip-address.my-namespace.pod.cluster.local
```

其中，pod-ip-address 是 Pod 的 IP 地址，my-namespace 是 Pod 所在的命名空间，pod 是 Pod 的名称。

当 Pod 启动时，它会在集群 DNS 中注册自己的 DNS 记录。它会向 Kubernetes 内置的 DNS 服务器查询该 Service 的 DNS 域名，这个 DNS 服务器实际上就是 kubelet 启动的 coredns 容器。

除了 Pod DNS 名称，Kubernetes 还使用了一些其他的 DNS 名称。例如，Kubernetes Service 对应的 DNS 名称有：

```shell
my-service.my-namespace.svc.cluster.local
```

其中，my-service 是 Service 的名称，my-namespace 是 Service 所在的命名空间。

当 Service 创建时，它会向 kube-dns 服务注册自己的 DNS 记录。kube-dns 服务会自动将该记录与其他 DNS 记录结合起来，提供一个完整的服务发现机制。

总的来说，在 Kubernetes 中，DNS 注册的过程是自动完成的。当 Pod 或 Service 创建时，它们会向集群 DNS 注册自己的 DNS 记录。kube-dns 服务负责将这些记录与其他服务和 DNS 记录结合起来，提供一个完整的服务发现机制。这使得 Kubernetes 用户可以轻松地在集群中发现和连接其他容器和服务。

#### DNS注册的过程

一共支持两种注册方式：

1. 通过config文件注册
   kubectl接收到client请求以后，发送http请求给apiserver, APIserver将内容转发到coreDNS
2. 通过service、Endpoint、pod的etcd中的信息进行动态更新
   定期向apiserver发送请求，获取service、Endpoint、pod的etcd中的信息，然后更新到coreDNS中

### coreDNS

> core DNS是每个node上运行一个吗？还是控制面上？

CoreDNS是一个Kubernetes集群中的Kubernetes插件，它通常在Kubernetes控制平面上作为一个Deployment或DaemonSet运行。在每个节点上都有一个CoreDNS Pod实例，它监听Kubernetes API服务器的变化并自动更新DNS记录。因此，可以说CoreDNS是在控制平面上运行的。

> coreDNS对于DNS记录的存储

CoreDNS内部包含一个DNS记录存储后端。该后端被称为CoreDNS的存储插件。CoreDNS存储插件的主要作用是管理DNS记录的存储和检索，例如将DNS记录存储在etcd、Consul或文件系统中，并在必要时从这些后端中检索记录。当客户端查询DNS记录时，CoreDNS将首先从存储插件中检索记录，然后将记录返回给客户端。如果记录不存在，则返回一个相应的错误。此外，存储插件还支持动态DNS更新，允许客户端通过API向CoreDNS添加、删除和修改DNS记录。

### pod 访问到service的逻辑

在 Kubernetes 中，Pod 可以通过 Service 的名称进行访问，无需知道具体的后端 Pod 的 IP 地址和端口号。Service 实际上是一个虚拟的逻辑概念，它代表了一组具有相同标签的 Pod，同时为这些 Pod 提供了一个统一的入口。

当 Pod 通过 Service 名称进行访问时，它会向 Kubernetes 的 DNS 服务器发出一个请求。这个请求的格式为`<service-name>.<namespace>.svc.cluster.local`，其中`<service-name>`是 Service 的名称，`<namespace>`是该 Service 所在的命名空间。`svc.cluster.local`是 Kubernetes 集群的默认域名，用于指示请求应该由集群内部的 DNS 服务器进行处理。

Kubernetes 的 DNS 服务器会将这个请求解析成一个或多个后端 Pod 的 IP 地址和端口号，并将其返回给发起请求的 Pod。Pod 将会使用这些信息与后端 Pod 进行通信，这个过程对于 Pod 来说是透明的。

**使用nginx做负载均衡，反向代理？**

## 环境准备

### 安装运行coreDNS

- 安装版本：1.10.1 (linux amd64)
- 安装命令

```shell
 tar -zxvf coredns_1.10.1_linux_amd64.tgz 
 sudo cp coredns /usr/local/bin
```

- 查看版本

```shell
coredns --version
```

- 运行
  在home目录下
```shell
./coredns -dns.port=1053 -conf /home/mini-k8s/pkg/kubedns/config/Corefile
```
- 测试
  - 插入一条信息
  ```shell
  etcdctl put /dns/com/example/sub '{"host":"1.2.3.4"}'
  ```
  - 测试效果
  在本机上测试
  ```shell
   dig @localhost +short -p 1053 www.service.com
  ```
  在其他机器上访问
  ```shell
  dig @192.168.1.13 +short -p 1053 sub.example.com
  dig @192.168.1.13 +short  -p 1053 www.baidu.com
  ```
  
#### 如何将coreDNS作为DNS服务运行
Flannel可以通过在其配置中指定DNS选项来使用CoreDNS作为DNS服务器。具体来说，可以在Flannel的配置文件中添加以下内容：
```shell
{
  "Network": "10.244.0.0/16",
  "Backend": {
    "Type": "udp",
    "Port": 7890
  },
  "DNS": {
    "Type": "coredns",
    "Endpoint": "10.0.0.10:1053",
    "ServiceName": "kube-dns",
    "Domain": "cluster.local"
  }
}

```
### DNS总体架构

![pic1](https://img2022.cnblogs.com/blog/2052820/202207/2052820-20220729201111426-1668551830.png)

当pod1应用想通过dns域名的方式访问pod2则首先根据容器中/etc/resolv.conf内容配置的namserver地址，向dns服务器发出请求，由service将请求抛出转发给kube-dns service，由它进行调度后端的core-dns进行域名解析。解析后请求给kubernetes service进行调度后端etcd数据库返回数据，pod1得到数据后由core-dns转发目的pod2地址解析，最终pod1请求得到pod2。


### 参考资料

[(107条消息) etcd3+coredns设置域名解析\_suiyingday的博客-CSDN博客](https://blog.csdn.net/suiyingday/article/details/90770884#:~:text=vim%20%2Fetc%2Fcoredns%2FCorefile.%3A53%20%7B%20%23%20%E7%9B%91%E5%90%ACtcp%E5%92%8Cudp%E7%9A%8453%E7%AB%AF%E5%8F%A3%20etcd%20%7B,%23%20%E9%85%8D%E7%BD%AE%E5%90%AF%E7%94%A8etcd%E6%8F%92%E4%BB%B6%2C%E5%90%8E%E9%9D%A2%E5%8F%AF%E4%BB%A5%E6%8C%87%E5%AE%9A%E5%9F%9F%E5%90%8D%2C%E4%BE%8B%E5%A6%82%20etcd%20test.com%20%7B%20stubzones%20%23%20%E5%90%AF%E7%94%A8%E5%AD%98%E6%A0%B9%E5%8C%BA%E5%9F%9F%E5%8A%9F%E8%83%BD%E3%80%82)
