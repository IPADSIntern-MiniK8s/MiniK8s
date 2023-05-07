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

- `WithMemoryLimit` 单位是字节，如果容器使用内存超过这个数 会被直接kill

- CPU：

  - `WithCPUs` 将容器进程绑定到指定cpu执行，比如`0-3`绑定到0 1 2 3  ，`1`绑定到1
  - `WithCPUCFS` 调度器，对应到nerdctl 是`--cpus` 会使用这个api，但是网上说这个参数指定cpu核，这个说法不准确，实际上如果这个值为1，会发生cpu0 和cpu1占用率都在50%的情况，即总使用量为1
  - `WithCPUShares` 份额

- port: 仅作标识用，没有意义，所以没有对应api

  [k8s四种port解析：nodePort、port、targetPort、containerPort - 简书 (jianshu.com)](https://www.jianshu.com/p/4b16c995990b) 

### task

containerd的api有一个docker没有的概念task

每个容器创建后，可以开启task，每个task对应一个进程，有对应的api，这时候才会产生新的命名空间

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
nerdctl对于网络的解析太复杂了，对于pause并没有很多乱七八糟的配置，所以直接用ctl启动pause    
在加入pause的namespace后发现，虽然其他容器有通过虚拟网卡向外找到合适的转发接口的能力，但是并没有DNS server。这里解决方法是使用外部的`nerdctl cp`命令将首个容器(pause)的`/etc/resolv.conf` `/etc/hosts`文件复制给每个该pod下的容器  
容器内部部署的服务 可在主机上通过容器ip+容器内端口的方式直接访问到，至此实现pod间通信、主机与pod通信，后续交给kube-proxy

### 原理

有时间再研究

[k8s网络插件之Flannel_林凡修的博客-CSDN博客](https://blog.csdn.net/weixin_43266367/article/details/127836595)

