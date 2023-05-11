# APIServer

## 安装etcd

安装的版本：v3.5.8

```bash
wget https://github.com/etcd-io/etcd/releases/download/v3.5.8/etcd-v3.5.8-linux-amd64.tar.gz
tar -zxvf etcd-v3.5.8-linux-amd64.tar.gz
cd etcd-v3.5.8-linux-amd64
sudo cp etcd* /usr/local/bin

# 检验是否安装成功
etcd --version

# 启动etcd
etcd

# 局域网启动
./etcd --listen-client-urls http://0.0.0.0:2379 --advertise-client-urls http://0.0.0.0:2371 --listen-peer-urls http://0.0.0.0:2380
```

## 依赖安装

```bash
go get -u github.com/gin-gonic/gin
go get -u github.com/gorilla/websocket
go get github.com/coreos/etcd/clientv3
```

## watch 实现思路

1. 原始的版本watch初步使用简单的websocket实现
2. 为了和service的接口保持一致，node的watch请求的source放在了request header里面

## api object 信息

### node
在 Kubernetes 中，每个节点（Node）都有一个 Conditions 字段，用于记录有关节点健康状况的信息。Conditions 字段是 NodeStatus 对象的一部分，由 API Server 维护并提供给其他 Kubernetes 组件和工具使用。

Conditions 字段由一组 Condition 对象组成，每个 Condition 对象都表示节点的一个特定方面的健康状况。每个 Condition 对象包含三个属性：

1. Status：表示 Condition 类型的字符串。在 Kubernetes 中，已经定义了一组标准的 Condition 类型，例如 Ready、OutOfDisk、MemoryPressure、DiskPressure 和 PIDPressure 等。
LastHeartbeatTime：表示最后一次收到节点的心跳时间的时间戳。
以下是一些常见的 Node Condition 类型和它们的含义：

Ready：表示节点是否可用于调度 Pod。如果该值为 True，则说明该节点可用；如果该值为 False，则说明该节点不可用；如果该值为 Unknown，则说明节点状态无法确定。
OutOfDisk：表示节点磁盘空间是否耗尽。如果该值为 True，则说明该节点的磁盘空间已经用尽；如果该值为 False，则说明该节点磁盘空间充足；如果该值为 Unknown，则说明该节点的磁盘状态无法确定。
MemoryPressure：表示节点是否出现了内存不足的情况。如果该值为 True，则说明该节点的内存资源已经用尽；如果该值为 False，则说明该节点的内存资源充足；如果该值为 Unknown，则说明该节点的内存状态无法确定。
DiskPressure：表示节点是否出现了磁盘不足的情况。如果该值为 True，则说明该节点的磁盘资源已经用尽；如果该值为 False，则说明该节点的磁盘资源充足；如果该值为 Unknown，则说明该节点的磁盘状态无法确定。
Kubernetes 的调度器会根据节点的 Conditions 字段来判断节点是否适合调度 Pod。例如，如果一个节点的 Ready 值为 False，则调度器不会将 Pod 调度到该节点上。同时，Kubernetes 组件和工具也可以根据节点的 Conditions 字段来监控和报警节点状态的变化。

2. 在 Kubernetes 的 node config 中，status.addresses 数组中的 type 字段用于指定节点的地址类型，可以有以下几种类型
Hostname：节点的主机名。
ExternalIP：节点的外部 IP 地址。
InternalIP：节点的内部 IP 地址。
ExternalDNS：节点的外部 DNS 名称。
InternalDNS：节点的内部 DNS 名称。
### pod
在 Kubernetes 中，Pod 的状态（Status）字段包含了关于 Pod 当前状态的各种信息。Pod 的 Status 字段包括以下几个字段：

Phase：表示 Pod 的当前生命周期阶段。常见的 Phase 值包括 Pending、Running、Succeeded、Failed 和 Unknown。其中，

- Pending 表示 Pod 正在被调度，但是尚未运行任何容器；
- Running 表示 Pod 正在启动容器；
- Succeeded 表示 Pod 中所有容器已经成功被启动；
- Finished 表示 Pod 中所有容器已经成功执行完毕；
- Failed 表示 Pod 中至少有一个容器执行失败；
- Terminating 表示 Pod 已经被删除；
- Unknown 表示 Pod 状态无法确定。

Conditions：表示 Pod 的当前状态条件。Conditions 是一个包含一组 Condition 对象的数组，每个 Condition 对象表示 Pod 的一个状态条件。常见的 Condition 类型包括 PodScheduled、Ready、ContainersReady 和 Initialized。其中，PodScheduled 表示 Pod 是否已经被调度到某个节点；Ready 表示 Pod 是否已经就绪；ContainersReady 表示 Pod 中的所有容器是否已经就绪；Initialized 表示 Pod 的初始化是否已经完成。
Message：表示 Pod 当前状态的信息。这是一个人类可读的字符串，用于描述 Pod 的当前状态。
Reason：表示 Pod 进入当前状态的原因。这是一个人类可读的字符串，用于描述为什么 Pod 进入当前状态。
HostIP：表示运行 Pod 的节点的 IP 地址。
PodIP：表示 Pod 的 IP 地址。
StartTime：表示 Pod 开始运行的时间。
除此之外，Pod 的 Status 字段还包含以下几个容器相关的字段：

Init Container Statuses：表示 Pod 中 Init Container 的状态信息。Init Container 是一种在 Pod 启动之前运行的容器，用于执行初始化操作。
Container Statuses：表示 Pod 中所有容器的状态信息。这是一个包含一组 ContainerStatus 对象的数组，每个 ContainerStatus 对象表示一个容器的状态。其中，重要的状态字段包括：State、LastState、Ready、RestartCount 和 Image。

## heartbeat
当 Kubernetes API Server 接收到 Worker Node 的心跳信息后，它会根据这些信息更新节点（Node）的状态（Status）。具体来说，Kubernetes API Server 会更新节点对象（Node Object）的 status 字段，该字段包含了节点的各种状态信息，例如节点的 IP 地址、健康状态、容器状态等。

在更新节点状态时，Kubernetes API Server 会根据 kubelet 发送的信息更新以下字段：

node.status.addresses：这个字段包含了节点的 IP 地址信息，包括内部 IP 和外部 IP。Kubernetes API Server 会根据 kubelet 发送的信息更新这个字段的值。

node.status.conditions：这个字段包含了节点的健康状态信息，包括节点是否处于 Ready 状态等。Kubernetes API Server 会根据 kubelet 发送的信息更新这个字段的值。

node.status.capacity：这个字段包含了节点的资源容量信息，例如 CPU、内存等。Kubernetes API Server 不会根据 kubelet 发送的信息更新这个字段的值，而是通过 kubelet 的启动参数或者节点标签来设置这个字段的值。

node.status.allocatable：这个字段包含了节点可用的资源容量信息。Kubernetes API Server 会根据 kubelet 发送的信息更新这个字段的值。

node.status.images：这个字段包含了节点上的镜像信息。Kubernetes API Server 会根据 kubelet 发送的信息更新这个字段的值。

除了上述字段之外，节点状态还包括了其他一些信息，例如节点的标签、节点的名称等。Kubernetes API Server 会根据 kubelet 发送的信息更新这些字段的值。更新节点状态后，其他 Kubernetes 组件（例如调度器、控制器等）可以根据节点状态来进行调度和管理

## 测试命令

1. 启动watch

```shell
wscat -H "X-Source: node-1" -c ws://localhost:8080/api/v1/watch/pods
```

如果使用kubelet发送命令，示例代码如下：

```go
import (
    "net/http"
    "github.com/gorilla/websocket"
)

headers := http.Header{}
headers.Set("X-Source", "my-source")

dialer := websocket.Dialer{}
dialer.Jar = nil // 禁用 cookie
dialer.Header = headers

conn, _, err := dialer.Dial("ws://example.com/api/v1/watch/pods/default", nil)
if err != nil {
    // 处理错误
}
defer conn.Close()

for {
    _, message, err := conn.ReadMessage()
    if err != nil {
        // 处理错误
    }
    // 处理消息
}
```

3. 清除etcd内所有数据

```shell
etcdctl del / --prefix
```

4. etcd查询所有的key
```shell
etcdctl get --prefix ""
```

5. 目前已经加入了scheduler， 如果想要不带scheduler的版本，可以取消`podhandler.go`的`line140-165`的注释，并注释掉`podhandler.go`的`line 167-208`

## 参考资料

Gin安装入门: https://gin-gonic.com/zh-cn/docs/quickstart/
