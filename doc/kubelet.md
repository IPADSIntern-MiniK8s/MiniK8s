# kubelet

## containerd

[containerd/getting-started.md at main · containerd/containerd (github.com)](https://github.com/containerd/containerd/blob/main/docs/getting-started.md)

镜像自带，但需要安装cni，高版本没有自带的flannel，用0.9.1版本可以

### management

管理容器，考虑以下三种方法：

1. containerd启动时会作为grpc server，监听在`unix:///run/containerd/containerd.sock` 可以像k8s一样作为grpc client调定义好的CRI接口。但是我们不需要考虑项目不同模块解耦，也不需要考虑支持其他的容器运行时，对于grpc的调用需要自己构造参数，太复杂，并且试了一下很难跑起来。

2. 用exec+ctl

   这里可以使用containerd写的nerdctl 兼容docker的命令行格式

   [containerd/nerdctl: contaiNERD CTL - Docker-compatible CLI for containerd, with support for Compose, Rootless, eStargz, OCIcrypt, IPFS, ... (github.com)](https://github.com/containerd/nerdctl)

   完全用cli工具技术含量不高，且需要经过nerdctl这个大框架的解析，效率不高。

   可以做一些辅助用途，比如测试、启动pause等。核心的查看容器状态和启动容器还是用containerd的go api

3. containerd api

   实在难用，官方文档一共就readme的几句话，剩下的全靠看源码+猜+看nerdctl源码如何使用
这里研究出的api如下
- 创建容器 包括配置
- 销毁容器
- 获取容器状态
- 获取容器资源信息

### configuration

- `WithMounts` 挂载 需要将type和options同时设为bind，否则会报`no such device`的错

- `WithDomainname` `WithHostname`

- `WithLinuxNamespace`可以加入其他进程的namespace 但是需要先起task 拿到pid`proc/pid/ns/uts`

  启动pause容器后，将此pod内的所有其他容器加入到pause容器的namespace    
  观察containerd的源码可知，就算什么都不配置，默认也是使用了ipc、uts、network、mount、pid这五个命名空间隔离的    
  [k8s之pause容器](https://blog.csdn.net/weixin_40579389/article/details/125941366)按这篇文章的意思 除了mount其他都不需要和pause隔离  
  需要修改的话 一种是自己写配置函数，另一种是使用这个api但只能一个个单独设 

- `WithProcessArgs` 启动命令 只有windows支持`ProcessCmdLine` 不过简单的命令使用起来效果差不多，具体可能涉及到entrypoint 和cmd的区别

- `Withenv`  环境变量 `"a=c"` 

- `WithMemoryLimit` 单位是字节，如果容器使用内存超过这个数 会被直接kill。

  莫名其妙会有bug，报cgroup的错，全网查不到信息，使用“30Mi” 没问题

- CPU：

  - `WithCPUs` 将容器进程绑定到指定cpu执行，比如`0-3`绑定到0 1 2 3  ，`1`绑定到1
  - `WithCPUCFS` 调度器，对应到nerdctl 是`--cpus` 会使用这个api，但是网上说这个参数指定cpu核，这个说法不准确，实际上如果这个值为1，会发生cpu0 和cpu1占用率都在50%的情况，即总使用量为1
  - `WithCPUShares` 份额

- port: 仅作标识用，没有意义，所以没有对应api

  [k8s四种port解析：nodePort、port、targetPort、containerPort - 简书 (jianshu.com)](https://www.jianshu.com/p/4b16c995990b) 
  
- `WithContainerLabels`这个功能为container提供label

  配合`client.Containers(ctx,fmt.Sprintf("labels.%q==%s", "pod", pod.Data.Name))`一起使用

  对于apiserver维护的信息，只是自定义的container apiobject，并不是containerd的可以用来获取真实容器信息的对象，使用containerd的添加label并使用filter的方法可以很方便地拿到一个pod对应的所有containers，否则需要通过遍历容器并比较ID来判断。

### task

containerd的api有一个docker没有的概念task

每个容器创建后，可以开启task，每个task对应一个进程，有对应的api，这时候才会产生新的命名空间

删除容器，先要killtask 然后delete task 最后delete容器

### image

containerd本身管理容器运行时，对于其他功能的提供非常少，包括拉取镜像。使用containerd提供的api只能做到从某个registry拉取，本地image是不行的，不带registry的image也是不行的。

通过观察nerdctl的源码可以得到以下两个扩展：

1. `pkg/imgutil/dockerconfigresolver/dockerconfigresolver.go/New`可知使用docker的resolver可以做到从不同registry(包括自己部署的)拉取镜像
2. `pkg/imgutil/imgutil.go/GetExistingImage`可知containerd提供`NewImage`方法，供`image.Image`对象到`containerd.Image`的转换，这里`image.Image`可以通过`client.ImageService().Get(imageName)`来获取。虽然这里在字符串解析上也必须出现registry的部分，但是实际上不会真的pull，而是从本地获取(`nerdctl image list`可见即可)

如果image在自己部署的registry中但还未被pull，这两种方法都是行不通的，需要自己创建docker resolver然后用WithResolver的配置去pull。这里图方便，解决方案为先用cli工具提前pull，随后紧接着用方法2获取到image对象

### pause
用containerd api设置网络特别麻烦，因此直接用nerdctl跑pause容器，并inspect拿到pid

因此这里直接使用podname+"-pause"为每个pause容器命名

虽然containerd的container对象只能访问.ID，而nerdctl 的 `--name` 设置的是name并不是id 但还是可以通过containerd的filter+label拿到container对象

然而由于一开始pause容器利用nerdctl实现网络配置，nerdctl本身除了调用containerd的api外，自己有维护一个namestore，在创建容器时会 `aquire` 需要在销毁时 `release`

所以销毁容器只调containerd的api是不够的，会导致containerd的容器已经被删掉了，但是nerdctl维护的信息还没删，导致下一次创建同名pause容器会有问题

解决方案可以是照nerdctl的代码找到对应的文件路径 然后照抄 `release`的代码，但这导致minik8s存在与nerdctl耦合的路径配置，较复杂 所以不如销毁pause仍用nerdctl直接实现

## network

containerd相较于docker并没有提供任何网络相关帮助，所以完全依赖CNI插件

`nerdctl network ls` `nerdctl run -net host/none`

CNI插件完成两个目标

1. 让每个容器(实际上就是一个pod)拥有一个虚拟网卡，使其拥有访问外网的能力
2. 支持跨node(物理主机)的pod间通信

[Kubernetes容器网络及Flannel插件详解_边缘计算社区的博客-CSDN博客](https://blog.csdn.net/weixin_41033724/article/details/124976813)

思路：

1. 使用flannel插件创建网络 此时每个node都会出现`flannel.1`的虚拟网卡，可以互相通信
2. 使用`nerdctl run -net flannel pause` 创建pause容器，此时ip在不同node上会在不同子网中进行分配，不会重复
3. 其他容器加入pause容器的network namespace

### flannel

[flannel/running.md at master · flannel-io/flannel · GitHub](https://github.com/flannel-io/flannel/blob/master/Documentation/running.md)

flannel目前已经支持了etcd v3版本，不需要切换v2。

etcd v3 v2的数据是不互通的，flanneld启动时默认会在v3里找数据

[Docker容器使用Flannel通信 - L_Hang - 博客园 (cnblogs.com)](https://www.cnblogs.com/lhang/p/17306765.html)

[Containerd网络管理_containerd 端口映射_班婕妤的博客-CSDN博客](https://blog.csdn.net/weixin_30641567/article/details/123917486)

只有master节点通过apiserver使用etcd，kubelet部署在node上 不需要也不能管理etcd

只需要一个etcd 不需要集群 (flannel如果使用etcd集群会出找不到lease的bug)

master `etcd --listen-peer-urls="http://192.168.1.12:2380,http://localhost:2380" --listen-client-urls="http://192.168.1.12:2379,http://localhost:2379" --initial-advertise-peer-urls="http://192.168.1.12:2380,http://localhost:2380" --advertise-client-urls="http://192.168.1.12:2379,http://localhost:2379"`

master `etcdctl --endpoints "http://192.168.1.12:2379" put /coreos.com/network/config '{"NetWork":"10.2.0.0/16","SubnetMin":"10.2.1.0","SubnetMax": "10.2.20.0","Backend": {"Type": "vxlan"}}'`

node启动`./flanneld-amd64 -etcd-endpoints=http://192.168.1.12:2379 -iface=ens3`

这里ens3是主机上能和外界通信的网卡，如果不设置flannel也会自动找

出现`flannel.1`的网卡。如果修改配置后第一次的flannel1无法消失 出现cni0 重启可以解决

```sh
# vim /etc/cni/net.d/10-flannel.conflist
{
  "name": "flannel",
  "cniVersion": "0.3.1",
  "plugins": [
    {
      "type": "flannel",
      "delegate": {
        "isDefaultGateway": true
      }
    },
    {
      "type": "portmap",
      "capabilities": {
        "portMappings": true
      }
    }
  ]
}
```

`nerdctl run -d -v /home/test_mount:/root/test_mount --net flannel -e port=12345 mcastelino/nettools /root/test_mount/test_network` 测试网络可行
nerdctl对于网络的解析太复杂了，对于pause并没有很多额外的配置，所以直接用ctl启动pause    
在加入pause的namespace后发现，虽然其他容器有通过虚拟网卡向外找到合适的转发接口的能力，但是并没有DNS server。这里解决方法是使用外部的`nerdctl cp`命令将首个容器(pause)的`/etc/resolv.conf` `/etc/hosts`文件复制给每个该pod下的容器  
容器内部部署的服务 可在主机上通过容器ip+容器内端口的方式直接访问到，至此实现pod间通信、主机与pod通信，后续交给kube-proxy

### 原理

[k8s网络插件之Flannel_林凡修的博客-CSDN博客](https://blog.csdn.net/weixin_43266367/article/details/127836595)

### 其他

由于需要给pod增加dns服务，在master上使用coredns作为dns server，解决方法是在resolve.conf中第一条加入master节点的ip（必须53端口）

resolv.conf的逻辑是 如果前一个nameserver连接不上，才会继续向下一个nameserver查找。

如果前一个nameserver连接上了但是没有记录，则会直接报无记录，不会向下一个nameserver查找。

因此必须保证在集群网络通常的情况下，master节点必须启用coredns并且coredns除了minik8s需要的dns服务，必须包括其他通用的dns服务。

## 资源监控

参考`cmd/nerdctl/container_stats_linux.go`

可以通过containerd的api拿到metrics对象,不过需要Unmarshal，并且对应的接口离其报错，找不到type，只能照着containerd的源码手动用反射

### memory
是一个定值，表示占用内存大小，单位byte

### CPU

进程创建开始之后累计执行的时间，如果跑在2个核上，过了1s，则记为2s
通过与上一次获取的cpu执行时间的delta和时间delta可以计算出CPUPercent，和top展示的cpu%是一模一样的
CPUPercent和容器创建指定的cpu参数可对应，例如指定cpu=1，则cpu%=100%;cpu=2,cpu%=200%（两核跑满）;cpu=500m,cpu%=50%

## 总结

kubelet主要做三件事

1. websocket与apiserver保持长连接，监听到pod的创建、销毁状态时进行对应的容器操作。
2. 作为http server接收对于容器资源的请求，获取并计算容器资源后返回。
3. 每隔一段时间检查所有容器的状态，将存在停止容器的pod的状态通过短链接更新给apiserver

针对container，1是写，23是读，可能会发生冲突。例如操作2正在统计某容器资源的时候，该容器被操作1删除。

使用读写锁为每个pod上锁，即`map[string]sync.RWMutex` 其中key为`namespace-podname`

go的map本身是线程不安全的，在对于同一个pod同时拿锁时可能创建两个不同的锁，严重时可能导致对于map的修改崩溃，因此将map替换为`sync.Map`

在和apiserver连接中断后，任务2会继续做，任务1会每5s再发起一次连接请求，任务3会按之前的频率继续做(由于get请求会失败所以拿不到pod信息)

