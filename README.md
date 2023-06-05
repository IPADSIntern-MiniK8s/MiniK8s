# Minik8s 验收文档

[toc]

## 1. 总体架构与软件栈

| 组件       | 功能                                                         | 使用的软件栈                                      |
| ---------- | ------------------------------------------------------------ | ------------------------------------------------- |
| kubectl    | miniK8s的命令行工具，用于创建，删除，修改，查询集群相关资源  | cobra                                             |
| apiserver  | APIServer作为Kubernetes的中央控制点，负责管理和暴露整个集群的API，并提供对集群资源的访问和操作,并通过ETCD进行资源的持久化 | gin、websocket、etcd                              |
| kubelet    | 管理容器生命周期、容器资源监控                               | containerd、flannel、gin、websocket、cobra、viper |
| kubeproxy  | 管理service的访问入口，通过修改本地的iptables规则或ipvs规则，实现从service到pod的转发 | ipvs                                              |
| controller | 监控各个资源的状态，管理所有资源                             | websocket                                         |
| scheduler  | 调度pod到合适的节点                                          | websocket                                         |
| miniDNS    | 为集群提供DNS服务                                            | coreDNS、nginx                                    |

- 总体架构图
  ![](https://notes.sjtu.edu.cn/uploads/upload_ea4f99b0f092506cb74b9d4f03754823.png)
- 总工作量
  ![10961685695300_.pic.png](https://www.z4a.net/images/2023/06/03/10961685695300_.pic.png)

## 2. 组员分工和贡献度占比

| 组员 | 分工                                        | 贡献度占比    |
| ---- | ------------------------------------------- | ------------- |
| szy  | kubelet容器管理、CNI网络、CI/CD、GPU        | $\frac{1}{3}$ |
| zhr  | APIServer、Scheduler、DNS、Serverless       | $\frac{1}{3}$ |
| zyy  | Kubectl、Replicaset、Service、HPA水平扩缩容 | $\frac{1}{3}$ |

## 3. 项目开发

### 3.1 分支介绍

![](https://notes.sjtu.edu.cn/uploads/upload_e9a98b40cac951977ac68525326a950d.png)

- master: 完成阶段性目标后由develop合并，通过集成测试。
- develop：合并各个功能分支后的主分支，
- 其他：每个分支表示一个新功能/组件，当完成后PR到develop

详见[3.4.2与3.4.6](#3.4 新功能开发流程)

### 3.2 CI/CD介绍

#### 3.2.1 软件栈选择

由于gitee仓库代码的更新需要通知某个部署了执行器的服务器，因此服务器必须具有公网ip且能被gitee仓库访问。为了不增加额外的成本，这里选择采用助教提供的方式，使用ipads的gitlab仓库+交大云主机实现CI/CD。
由于希望测试可以包含更多的部分，例如容器、网络等，而使用docker部署+docker执行器往往会产生与主机不一致的执行结果，不单单是写dockerfile可以解决，故最终采用在云主机上安装gitlab-runner且选择shell作为执行器。
这种方法使得测试环境与开发环境完全一致，避免了权限、软件版本、虚拟网卡等各种问题，且方便拿到build后的结果。缺点是所有需要的环境都必须提前在部署gitlab-runner的主机上进行安装，且进行测试时会对主机环境造成影响。

#### 3.2.2 流程与配置

这里简单将CI/CD分为两部分。

1. 单元测试：使用`go test`测试不同模块的代码
2. 构建可执行文件：利用提前编写好的Makefile编译得到可执行文件，复制到主机上的对应目录。

```makefile
CMDPATH=../cmd
OUTPATH=./bin

kubectl:
	go build -o $(OUTPATH)/kubectl $(CMDPATH)/kubectl.go

kubelet:
	go build -o $(OUTPATH)/kubelet $(CMDPATH)/kubelet.go

apiserver:
	go build -o $(OUTPATH)/apiserver $(CMDPATH)/apiserver.go

scheduler:
	go build -o $(OUTPATH)/scheduler $(CMDPATH)/scheduler.go

controller:
	go build -o $(OUTPATH)/controller $(CMDPATH)/controller.go

kubeproxy:
	go build -o $(OUTPATH)/kubeproxy $(CMDPATH)/kubeproxy.go

serverless:
	go build -o $(OUTPATH)/serverless $(CMDPATH)/serverless.go

all:
	go build -o $(OUTPATH)/kubectl $(CMDPATH)/kubectl.go
	go build -o $(OUTPATH)/kubelet $(CMDPATH)/kubelet.go
	go build -o $(OUTPATH)/apiserver $(CMDPATH)/apiserver.go
	go build -o $(OUTPATH)/scheduler $(CMDPATH)/scheduler.go
	go build -o $(OUTPATH)/controller $(CMDPATH)/controller.go
	go build -o $(OUTPATH)/kubeproxy $(CMDPATH)/kubeproxy.go
	go build -o $(OUTPATH)/serverless $(CMDPATH)/serverless.go

clean:
	rm $(OUTPATH)/*
```

```yaml
stages:
  - prepare
  - test
  - build

prepare:
  stage: prepare
  script:
    - go env -w GOPROXY=https://goproxy.cn
  tags:
    - shell

test-kubelet:
  stage: test
  script:
    - sudo /usr/local/go/bin/go test minik8s/pkg/kubelet/container -cover
    - sudo /usr/local/go/bin/go test minik8s/pkg/kubelet/pod -cover
  tags:
    - shell

test-kubectl:
  stage: test
  script:
    - echo "testing kubectl"
  tags:
    - shell

test-kubeproxy:
  stage: test
  script:
    - echo "testing kubeproxy"
    - sudo /usr/local/go/bin/go test minik8s/pkg/kubeproxy -cover
  tags:
    - shell

test-apiserver:
  stage: test
  script:
    - echo "testing apiserver"
    - sudo /usr/local/go/bin/go test minik8s/pkg/kubeapiserver/storage -cover
    - sudo /usr/local/go/bin/go test minik8s/pkg/kubeapiserver/handlers -cover
  tags:
    - shell

test-scheduler:
  stage: test
  script:
    - echo "testing scheduler"
    - sudo /usr/local/go/bin/go test minik8s/pkg/kubescheduler/policy -cover
    - sudo /usr/local/go/bin/go test minik8s/pkg/kubescheduler/filter -cover
  tags:
    - shell
    
test-serverless:
  stage: test
  script: 
    - echo "testing serverless"
    - sudo /usr/local/go/bin/go test minik8s/pkg/serverless/activator -cover
    - sudo /usr/local/go/bin/go test minik8s/pkg/serverless/workflow -cover
  tags:
    - shell

build:
  stage: build
  script:
    - cd build
    - make all
    - sudo cp -r bin /home/gitlab-runner/$CI_COMMIT_BRANCH/
  tags:
    - shell
```

当任意测试失败时，不会构建最后的可执行文件。
<img src="https://notes.sjtu.edu.cn/uploads/upload_484ba223704a17dfd9e88213d98ef306.png" style="zoom:60%;" />
所有测试通过后，可在主机对应的目录下拿到构建好的可执行文件，且利用gitlab-runner提供的环境变量每个分支构建好的可执行文件不会互相覆盖
<img src="https://notes.sjtu.edu.cn/uploads/upload_11c6e4605438a68b0b9a657d15c4862c.png" style="zoom:60%;" />
各个组件的单元测试环境不会互相冲突时可并行测试，故共构造三个相同的runner进行测试。
<img src="https://notes.sjtu.edu.cn/uploads/upload_e0b1ad4948df487a4870e98e18effbb6.png" style="zoom:50%;" />
<img src="https://notes.sjtu.edu.cn/uploads/upload_e03ac5e97b381f4d04a321914c270dbf.png" style="zoom:67%;" />

### 3.3 软件测试方法

- 单元测试：使用go的testing模块编写测试代码，通过调用函数及判断结果是否符合预期的方式验证被测代码的正确性。
  <img src="https://notes.sjtu.edu.cn/uploads/upload_704d50dcdbcba2535dab630b712d37b6.png" style="zoom:67%;" />
  以kubelet为例，测试包括了拉取镜像、创建删除容器、获取容器资源等。测试精心构造镜像与容器程序，检查了网络、占用资源情况、通信情况等各个功能点，保证基础功能不会随着后续版本更新而被破坏。
- 集成测试：gitlab-runner只跑在一台主机上，因此基于gitlab-runner的集成测试较好的实现方式应该是在build之后跑一个脚本，运行各组件的可执行文件，然后在shell中用其他工具判断功能的准确性，可以测试除了多机之外的所有功能的情况。然而碍于时间与课业压力，这里集成测试并未使用工具或脚本，而是由组员手动完成。在每次单元测试通过后，由组员手动提取出需要的可执行文件，在多机上部署后手动进行集成测试。

### 3.4 新功能开发流程

我们根据项目文档的要求，将MiniK8s分成了不同的功能模块，分别开发，逐步整合。总体上来说，对于新功能的开发，采用了如下工作流：

1. 确定目标和范围：
   根据需求文档和K8s本身的功能确定要开发的新功能的目标和范围。明确功能的需求和期望的实现成果，并在小组内进行讨论和确认。

2. 创建功能分支：
   在git中创建一个新的功能分支，用于开发新功能。这样可以将新功能的开发与主分支的稳定代码分离开来。

3. 设计和规划：
   基于需求文档和参考资料，进行功能的详细设计和规划。确定新功能的架构、模块和接口，并将其与现有代码进行整合。

4. 实施和编码：
   根据设计和规划，开始进行代码的编写。在编程规范上我们使用了Google编程规范，以`Goland`作为主要的开发IDE（`Goland`本身也有一定的代码规范约束），用`logrus`作为log工具。

5. 单元测试：
   每当新功能开发完毕时，小组成员会对于新增功能使用Go的内置测试框架 testing 进行单元测试。编写针对函数、方法或模块的测试用例，验证其功能和边界条件。在新功能合并后，进行集成测试，测试新功能与现有功能的交互和一致性，确保整体系统的稳定性和一致性。比如在APIserver、kubectl和kubelet初步版本实现时，我们进行了集成测试，完成了pod创建的完整流程。

6. 集成和代码审查：
   将新的功能分支整合以pull request的形式合并到develop分支时，小组成员根据pull request的内容进行代码审查，检查代码的质量、性能和安全性。
   <img src="https://notes.sjtu.edu.cn/uploads/upload_3a8da2f1f00314fd96caa68ad29cefcf.png" style="zoom:67%;" />


8. 文档和说明：
   小组成员在开发新功能时，会将新功能的实现、使用方法和配置指南记录在文档（`/doc`下）中，供其他人参考
   <img src="https://notes.sjtu.edu.cn/uploads/upload_bea1fdb0a9896707e80d95c50b3d9d21.png" style="zoom:67%;" />


9. 例会与交流
   每周举行例会，交流本周工作，保证所有成员对于项目代码有全面的理解。使用微信工作群实时交流，遇到与其他成员开发的模块相关的问题时，第一时间解决。

## 4. 系统架构和组件功能

### 4.1 APIObject抽象

#### 4.1.1 Node

Node有如下状态字段：

- Ready：Node可用
- DiskPressure：节点上的磁盘存储空间不足
- MemoryPressure：节点上的内存不足
- NetworkUnavailable：节点不可达

#### 4.1.2 Pod

Pod有如下状态字段：

- Created：这个字段仅给Job使用，表示Job创建了pod但pod在Pending之前的状态。Pod本身不会使用这个状态。
- Pending：接收到Pod请求，但是还没有为Pod调度
- Scheduled：Pod已经调度到对应的Node，但是Pod还没有正常启动
- Running：Pod正在正常运行
- Failed：Pod任务运行失败，即Pod中存在某个容器停止运行且返回值非0
- Finished：Pod任务运行完毕，即Pod中不存在Failed的容器且存在某个容器停止运行且返回值为0
- Terminating：接收到delete Pod的请求，但是Pod还没有完全被终止
- Deleted： Pod被删除
- Unknown：未知状态

#### 4.1.3 Service

重要的状态字段如下：

- ClusterIP：服务对应的虚拟IP，由master随机分配
- NodePort：仅在NodePort模式下使用，访问节点时指定的静态端口号。

#### 4.1.4 ReplicaSet

重要的状态字段如下：

- Replicas：当前副本数
- Scale：当replicaset被其他apiobject控制时使用，用于设置其他资源所要求的期望副本数
- OwnerReference：记录控制replicaset的apiobject的信息
- ReadyReplicas：已经准备完成的副本数（Pod状态为Running）

#### 4.1.5 HorizontalPodAutoscaler

重要的状态字段如下：

- LastScaleTime：上一次扩缩容的时间
- CurrentReplicas：当前副本数
- DesiredReplicas：期望副本数
- CurrentMetrics：从kubelet拿到的最新资源指标值

#### 4.1.6 DNSRecord

包含了配置名称（name）、namespace、配置类型（kind）、主路径（host）、⼦路径（path），以及转发的⽬标Service等信息, 具体参考示例`example/dns/dnsrecord.yaml`

#### 4.1.7 Job

Job本身和Pod非常接近，区别只是Job做短任务，即spec指定的container的cmd会在某段时间内结束。
此外增加两个属性

- backoffLimit：当job执行失败后的再次重试次数
- ttlSecondsAfterFinished：当job执行成功后多久删除job对应的pod

#### 4.1.8 Function

包含了配置名称（name）、配置类型（kind）、代码文件路径（path）等信息，具体参考示例`example/serverless/singlefunc.yaml`

#### 4.1.9 WorkFlow

主要参考了`AWS StepFunction`的实现思路, 以下是具体的格式规定：

- WorkFlow是由若干`State`组成的状态机，具有以下字段：
  - Kind：表示工作流的类型。
  - APIVersion：表示工作流的API版本。
  - Name：表示工作流的名称。
  - Status：表示工作流的状态。
  - StartAt：表示工作流的起始状态。
  - States：是一个映射（map），以状态名称为键，以状态对象为值，表示工作流中各个状态的定义。
  - Comment：表示工作流的一些说明信息。
- WorkFlow支持三种状态类型：任务状态（TaskState）、失败状态（FailState）和选择状态（ChoiceState）
- 任务状态（TaskState）的说明如下：
  * Type：表示状态的类型，这里是 Task。
  * InputPath：用于选择输入参数。
  * ResultPath：用于选择输出参数。
  * Next：表示下一个要执行的状态。
  * End：表示该任务状态是否是工作流的最后一个状态。
- 失败状态（FailState）的说明如下：
  * Type：表示状态的类型，这里是 Fail。
  * Error：表示错误信息。
  * Cause：表示失败原因
- 选择状态（ChoiceState）的说明如下：
  * Type：表示状态的类型，这里是 Choice。
  * Choices：表示选择项的列表，每个选择项包含条件和下一个状态。
  * Default：表示默认的下一个状态，用于处理没有匹配到任何选择条件的情况。
  * 目前支持对于`string`和`numeric`类型的各种比较

具体参考示例`example/serverless/workflow.yaml`

### 4.2 APIserver 

#### 4.2.1 overview

APIserver 是MiniK8s中的核心组件之一，它是整个集群中的控制平面组件。APIserver作为KuberneWorkFlow
tes的中央控制点，负责管理和暴露整个集群的API，并提供对集群资源的访问和操作。

#### 4.2.2 访问管理

APIserver使用**RESTful**风格的API，即通过HTTP协议进行通信，并使用标准的HTTP方法（目前支持GET、POST、PUT、DELETE四种方法）和URL路径来执行操作。

- 每个资源都有一个唯一的URL路径表示，例如/pods、/services等。
- APIServer支持常见的HTTP操作，如获取（GET）、创建（POST）、更新（UPDATE）、删除（DELETE）和获取列表（GET LIST）资源。
- HTTP的请求返回由状态码和response body两部分组成，常见状态码有`200 OK`,  `400 BadRequest`和`500 Internal Server Error ` , 当状态码为`400`或者`500`时，response body中会携带详细的错误信息
- 以下是一些常见的API的基本格式
  - `POST /api/v1/namespaces/{namespace}/{resources}` ——创建资源
  - `POST /api/v1/namespaces/{namespace}/{resources}/{name}/update` ——更新资源
  - `GET /api/v1/namespaces/{namespace}/{resources}/{name}`——获取特定名称的资源
  - `GET /api/v1/namespaces/{namespace}/{resources} `——获取特定namespace下的资源
  - `DELETE /api/v1/namespaces/{namespace}/{resources}/{name} ` ——删除特定名称的资源

这里借助**Gin框架**提供了HTTP的服务支持，通过一个`HandlerTable` (`pkg/kubeapiserver/handlers/handlertable.go`) 进行了统一路由管理，并在启动时根据路由方法选择对应的HTTP方法，并将路由路径和处理程序注册到HTTP服务器中

#### 4.2.3 资源管理

MiniK8s使用**etcd**作为控制面数据持久化容器,APIserver是唯一可以与master etcd交互的组件，etcd中存储着用于存储集群的元数据、配置信息和状态数据，如Pod等资源对象的配置信息都存储在etcd中

资源在etcd中的存储格式是以`JSON`格式表示的键值对形式。每个资源对象都有一个唯一的键（`key`），以及对应的`JSON`值（`value`），其中包含了资源对象的所有属性和配置信息。

对于资源对象的存储采用了分层存储结构，每个资源对象`key`采用了类似目录的方式组织，比如pod的`key`为`/registry/pods/{namespace}/{name}`,  从而可以支持通过前缀快速匹配一系列符合条件资源

#### 4.2.4 Watch机制

APIserver的watch机制是通过**websocket**配合etcd实现的。

当APIServer和特定的组件建立起websocket连接以后，当etcd中的存储内容发生变化之后，变化的信息可以通过这个websocket connection传送到正在监听的组件

APIServer维护着一个`WatchTable`, table中的`key`是监听资源的要求，`value`以该`key`为标准watch的所有websocket连接列表，这里考虑到并发访问安全性，设计实现了一个`ThreadSafeList`，支持websocket连接列表的动态安全修改

当client希望监听某种资源的变化情况时，会给APIserver发送一个websocket连接请求，`url`指明想要监听的资源要求，HTTP服务器中间件`UpgradeToWebSocket()`会拦截websocket请求，获得想要watch的资源`key`值，进行请求升级操作，并将升级后的websocket连接存入`WatchTable`。

当etcd中的内容有所更新时，会根据更新内容的key值，满足该key值的watch key，并将更新的value，发送到所有对应的websocket列表。

得益于etcd资源的分层存储, 该机制支持从所有资源（`/registry`), 特定种类资源（`/registry/{resources}`)，特定namespace的特定资源（`/registry/{resources}/{namespace}`)和特定名字、特定namespace的特定资源`(/registry/{resources}/{namespace}/{name}`)的层次进行watch （但是实际使用中只用到了'特定种类资源'这一粒度）

<img src="https://notes.sjtu.edu.cn/uploads/upload_f01e3e5b07859406005fe6a857180534.png" style="zoom:67%;" />



#### 4.2.5 Heartbeat

APIServer接受节点发送的heartbeat，更新node的`lastHeartbeatTime`, 更新node的健康状态为`Ready`，APIServer使用一个goroutine定期检查node的状态，如果超过一定时间没有接受到node的heartbeat（`30s`)，将node的健康状态标记为`NetworkUnavailable`，此时会进行**reschedule**：如果节点上有正在运行的 Pod，scheduler会检测到节点不可用并重新评估调度决策, scheduler会尝试将受影响的Pod调度到其他可用的节点上, 从而保证集群的高可用性

### 4.3 Scheduler

#### 4.3.1 Overview

Scheduler是 MiniK8s 控制平面的一个组件，负责为新创建的 Pod 选择一个合适的 Node 节点来运行。当 APIServer 接收到创建 Pod 的请求时，会将该请求发送给scheduler，scheduler 将根据一些规则和条件为该 Pod 分配一个合适的 Node。

Scheduler的主要工作可以分为检查pod状态，筛选node，评分和排序这几步。此处`filter`和`policy`采用了工厂方法实现，保证了低耦合性和高拓展性。

<img src="https://notes.sjtu.edu.cn/uploads/upload_b6fd83be1a29be0be311eed25b0df933.png" style="zoom:67%;" />


#### 4.3.2 Step1: 检查pod状态

在这一步中，通过检查pod的字段信息，确定该pod是否是一个待调度的pod（`Pending`)

#### 4.3.3 Step2: 筛选node (filter)

Scheduler从APIServer查询到当前所有的node，根据以下步骤筛选掉不合适的node：

1. node的健康状态是否是`Ready`
2. 如果pod的`NodeSelector`字段不为空，检查node是否有符合该字段的`Label`
3. 如果pod和node有明确的资源说明，检查node当前`Allocatable`的资源是否能满足pod的`Requests`（这里支持cpu和memory两种资源的检查，并且支持多种单位，如cpu的`m`，memory的`Ki`, `Mi`)

#### 4.3.4 Step3: 评分和排序 (policy)

Scheduler支持了两种schedule的policy，一种根据当前资源的使用率，另一种选择调度pod数量最少的node

- `LeastUtilization`：根据cpu和memory的使用率计算分数，使用率越低，分数越高，如果没有该资源的信息，默认分数为0
- `LeastRequest`: scheduler记录了每个node上调度pod的数量，调度的pod的数量越少，分数越高

Scheduler对于node分数从高到低进行排序，默认选择当前分数最高的node进行调度

Scheduler默认的调度策略为`LeastRequest`，client可以通过config文件指定调度策略。

#### 4.3.5 Corner case

这里考虑了一种情况，在Scheduler选择出分数最高的node，但是在Scheduler将该信息发送给APIServer，APIServer真正将该调度信息发送给对应的node前，该node的状态更新为了不可用（`NetworkUnavailable`），这时pod的调度就会失败，为了尽量避免这种情况发生，Scheduler会将筛选出来的所有node按照分数排序的序列发送给APIServer, 首先会调度到分数最高的node，接下来APIServer会首先调度到分数最高的node，然后利用一个goroutine，每隔一段时间进行检查（`3min`)，如果此时pod还没有调度成功，会尝试将该pod调度到下一个node上。

### 4.4 Controller

#### 4.4.1 Overview

Controller负责控制管理集群状态的某个特定方面，监控各个资源的状态并将当前状态转变为期望的状态。每个Controller通过[Watch机制](#424-Watch机制)[](#)至少追踪一种类型的资源，并通过给apiserver发送信息对资源进行增删改查操作。

每个Controller的工作流程大致相同：1. 声明自己需要watch的资源。2. 每当资源发生改变时，WatchClient根据`ResourcesVersion`字段判断变化类型（增加/删除/修改）。3. 根据变化类型调用对应的处理函数处理。4. 持续watch，等待下一次改变。可以发现，上述流程中的1，2，4步逻辑相同，所以在具体实现过程中，采用了**工厂方法**的设计模式，定义抽象的`SyncFunc`接口，并由不同的controller各自实现接口中的方法。

<img src="https://notes.sjtu.edu.cn/uploads/upload_39a2263022acab8259fe90bb4ecf19a1.png" style="zoom:67%;" />


在实现设计上，由于多个Controller会分别创建或者更新相同类型的对象，故通过`Metadata.OwnerReference`字段对不同Controller控制的资源加以区分，保证每个Controller只会管理与其相对应的资源。例如，JobController不会删除ReplicasetController控制的Pod。

#### 4.4.2 ControllerManager

ControllerManager是一个守护进程，用于保证所有Controller的正常运行。集群启动时，通过ControllerManager启动所有Controller；在与Apiserver断连后，保证Controller进程不会退出，而是定期重连，这样在Apiserver组件退出后Controller组件不需要重启。

#### 4.4.3 ReplicasetController

ReplicasetController用于维护Replica的副本数为期望值，需要监控和管理的资源为Replica和Pod。主要工作如下：

1. 监听replica资源的创建。
   1. 寻找符合selector条件的pod并将其`Metadata.OwnerReference`字段设置为当前replicaset。
   2. 如果满足条件pod个数小于`Spec.Replicas`值，根据`Spec.Template`设定的模板创建新的pod。
2. 监听replica资源的更新。监听目标replicas数目的更改和selector条件的更改，并创建/删除pod，并删除对应的pod。
3. 监听pod删除。如果满足controller条件，对应当前replica数目减1，根据template创建新的pod。
4. 监听pod更新。查看label是否更改，并更改对应`Status.Replica`和`Metadata.OwnerReference`状态；当pod发生crash或者被kill掉，自动重启pod。

#### 4.4.4 ServiceController

ServiceController用于维护Service和Pod之间的映射关系，需要监控和管理的资源为Service和Pod。通过Service和Pod的变化增删Endpoint资源，从而让kubeproxy实现转发规则的添加。主要工作如下：

1. 监听service资源的创建。
   1. 如果服务类型为`ClusterIP`，则通过etcd从可分配的IP池中为其分配唯一的cluster ip；如果服务类型为`NodePort`，则通过etcd从可分配的Port池中为其分配唯一的静态端口。
   2. 遍历pod列表，找到符合selector条件的pod。创建对应的endpoint资源。
2. 监听service资源的更新。检查selector是否更新，如果更新，删除原先的endpoint并创建新endpoint。
3. 监听service资源的删除。删除对应的endpoint。
4. 监听pod删除。删除pod对应的endpoint。
5. 监听pod更新。如果标签更改，删除/增加endpoint。

通过新增endpoint抽象资源的方式，可以做到一个pod绑定多个service和一个service绑定多个pod，从而实现pod和service抽象的解耦。

#### 4.4.5 HPAController

HPAController用于实现根据任务的负载对Pod的replica数量进行动态扩容和缩容，需要监控和管理的资源为HPA。主要工作如下：

1. 监听autoscaler的创建。更改对应replicaset/创建对应replicaset。

2. 监听autoscaler的更改。如果CurrentReplicas和DesiredReplicas数量不一致，则更新对应replicaset。

3. 监听autoscaler的删除。将对应replicaset的`Metadata.OwnerReference`状态还原。

4. 每隔15s检查hpa的条件是否满足，进行扩缩容。扩缩容逻辑如下：

   1. 根据`Spec.ScaleTargetRef`字段找到对应的replicaset。`Spec.Metrics`字段提供多个指定的指标，根据每个指标利用metric api向kubelet查询对应pod的资源利用率并计算当前指标值。

   2. 将计算出的结果与指标比较。计算出扩缩容的期望副本数。公式： 期望副本数 = ceil[当前副本数 * (当前指标 / 期望指标)]

   3. 每个指标都会计算出一个期望副本数。取最大值作为总的期望副本数。

   4. 根据`Spec.Behavior`字段定义的扩缩容行为判断总的期望副本数是否满足条件，并确定最终的期望副本数。需要满足的条件有：

      a. 不超过MaxReplicas，不小于MinReplicas。

      b. 上一次扩缩容距今时间大于StabilizationWindowSeconds（扩容默认为0，缩容默认为300s）

      c. 满足`Spec.Behavior.*.Policies`字段定义的HPAScalingPolicy。（如每3秒最多新增10个pod，每20s最多减少10%的pod）。不同Policy的限制之间可以由`Spec.Behavior.*.SelectPolicy`字段设定取最小/最大限制。

5. 根据上述三个条件的限制确定最终副本数，并更新hpa的`Status.DesiredReplicas`。

#### 4.4.6 JobController

1. 创建对应的pod
2. 监听对应pod的状态更改，把对应pod状态设置为job状态
3. pod stopped
   - failed：先deletepod再createpod，最多重复backoffLimit次
   - finished：ttlSecondsAfterFinished后删除pod

### 4.5 Kubelet

#### 4.5.1 Overview 

kubelet的职责：

1. 与apiserver保持长连接，监听pod的创建/删除请求，通过containerd api与nerdctl完成容器的创建、网络配置与容器删除。
2. 作为http server，监听获取容器资源占用情况的请求，通过containerd api获取并计算容器在一段时间内的memory、cpu占用情况后返回。
3. 每隔一段时间判断当前主机上所有容器的运行状态，若某容器退出则标记此容器属于的pod的状态为Failed/Finished

#### 4.5.2 运行配置

- ApiserverAddr : 与apiserver通信地址
- FlannelSubnet : 此主机上flannel配置的创建容器时用到的子网网段，告知apiserver而并非配置
- IP            : 此主机ip，告知apiserver而并非配置
- Labels        : 用于nodeSelector的匹配
- ListenAddr    : 监听容器资源情况请求的地址
- CPU   : 用于scheduler调度，配合容器资源申请
- Memory: 用于scheduler调度，配合容器资源申请


使用`viper`库帮助从yaml中获取配置，使用`cobra`指定config路径
`./kubelet -c ./config/kubelet-config.yaml`

#### 4.5.3 容器运行时选择

kubelet作为最底层的管理容器的组件，需要直接与容器运行时交互。由于lab文档推荐使用containerd，这里选择使用containerd作为容器运行时。具体有如下三种方式。

| 方式                                                         | 优点     | 缺点                                      |
| ------------------------------------------------------------ | -------- | ----------------------------------------- |
| containerd api                                               | 效率高   | 文档稀缺，使用困难                        |
| nerdctl                                                      | 使用简单 | 必须通过nerdctl实现的一大部分代码，效率低 |
| grpc                                                         | 封装较好 | 需要额外编译，且需要理解设定好的参数      |
| 最终采用1和2组合的方式，通过阅读containerd和nerdctl两个仓库的源码学习api的使用，在更复杂的情况下使用nerdctl辅助。 |          |                                           |

#### 4.5.4 镜像拉取

containerd api提供的镜像拉取必须指定registry和tag,且如果使用pull方法，不论本地有没有镜像都会从registry拉，而使用自己的registry会有问题，因此最终实现三种方法。

1. dockerhub的镜像：直接使用`client.Pull`
2. 本地存在镜像：先使用`client.ImageService().Get`拿到`image.Image`对象，再用`containerd.NewImage`转为`containerd.Image`对象
3. 自己在某主机上起的docker registry：先用`nerdctl -n namespace pull`拉到本地，随后用方法2

如果需要拓展，可以使用docker resolver进行配置

#### 4.5.5 容器生命周期管理

#### 4.5.5.1 创建

支持配置如下

- WithMounts 挂载，这里只支持host。实际上部署了NFS，之后可以考虑增加NFS模式
- WithDomainname WithHostname：由于pod属于同一个uts namespace，一个pod只有一个hostname
- WithLinuxNamespace：拿到pidproc/pid/ns/uts可以加入其他进程的namespace 
- WithProcessArgs ：启动命令，会覆盖dockerfile中的CMD 
- Withenv： 环境变量 
- WithMemoryLimit 单位是字节，如果容器使用内存超过这个数 会被直接kill。
- CPU：
  - WithCPUs 将容器进程绑定到指定cpu执行，比如0-3绑定到0 1 2 3 ，1绑定到1
  - WithCPUCFS 调度器，对应到nerdctl 是--cpus 会使用这个api，但是网上说这个参数指定cpu核，这个说法不准确，实际上如果这个值为1，会发生cpu0 和cpu1占用率都在50%的情况，即总使用量为1
  - WithCPUShares 份额
- port: 仅作标识用，没有意义，所以没有对应api
- WithContainerLabels：这个功能为container提供label
  配合`client.Containers(ctx,fmt.Sprintf("labels.%q==%s", "pod", pod.Data.Name))`一起使用

#### 4.5.5.2 删除

涉及到的对象有container和task，涉及到的状态有running、stopped
这里统一将container和task全部删除，不考虑中间的stopped状态
注：nerdctl在使用containerd api的基础上自己还会维护状态，因此如果容器使用nerdctl进行创建（pause），则必须使用nerdctl进行删除，否则会产生不一致的问题。

#### 4.5.6 容器资源获取

可以通过containerd的api拿到metrics对象,不过需要Unmarshal，并且对应的接口离其报错，找不到type，只能照着containerd的源码手动用反射

- memory:是一个定值，表示占用内存大小，单位byte

- CPU:进程创建开始之后累计执行的时间，如果跑在2个核上，过了1s，则记为2s 通过与上一次获取的cpu执行时间的delta和时间delta可以计算出CPUPercent，和top展示的cpu%是一模一样的 CPUPercent和容器创建指定的cpu参数可对应，例如指定cpu=1，则cpu%=100%;cpu=2,cpu%=200%（两核跑满）;cpu=500m,cpu%=50%

#### 4.5.7 网络

使用flannel，需要提前安装flannel二进制与cni，cni只有0.9版本及以下才可以直接用flannel。
etcd中写入集群配置信息，随后在cni路径中创建新的网络配置文件，运行flanneld即可完成配置，随后启动容器加入`--net flannel`即可
这里flannel提供以下两个功能

1. DNS：使得容器可以借由主机网卡访问外网服务
2. 跨主机互通：不同主机的容器可互相连接

#### 4.5.8 Pod管理

Pod是我们自己提供的抽象，与容器运行时无关，而kubelet本身是stateless的，因此只需要每次根据pod对应的container信息进行相关的容器操作即可。
每个container命名为`podName-containerName`
启动pause容器后，将此pod内的所有其他容器加入到pause容器的ipc uts pid network namespace。发现只有一开始启动的pause容器具有正常的DNS功能，其他加入此namespace的容器虽然网络联通正常，但是无法使用DNS功能，因此使用nerdctl将pause容器中的`/etc/hosts`与`/etc/resolv.conf`拿出放入此pod的其他容器中，并且加入coredns的地址。
对于apiserver维护的信息，只是自定义的container apiobject，并不是containerd的可以用来获取真实容器信息的对象，使用containerd的添加label并使用filter的方法可以很方便地拿到一个pod对应的所有containers，否则需要通过遍历容器并比较ID来判断。

#### 4.5.9 锁与执行流

观察4.5.1,其中1可视作对容器的写操作，2 3为读操作，是有可能发生冲突的。例如在接收到容器资源占用请求后开始统计某容器资源，此时正好接收到删除某容器的请求，会发生错误。
使用读写锁为每个pod上锁，即`map[string]sync.RWMutex` 其中key为`namespace-podname` 锁必须细粒度，因为2操作非常慢。
go的map本身是线程不安全的，在对于同一个pod同时拿锁时可能创建两个不同的锁，严重时可能导致对于map的修改崩溃，因此将map替换为`sync.Map`。虽然同时写map没问题，但是很可能出现t1创建完并拿完锁之后return t2再次创建并拿锁，原因是并没有另一把锁来让对于map中某个key的访问设为临界区。使用sync.Map提供的`LoadOrStore(key,value)`方法，它会先判断是否存在某个key，如果存在返回`map[key]`，否则设置`map[key]=value`后返回value，最后用value.lock。这个方法只是压缩代码行数到了两行，但仍然不是原子的。最好是有类似`tbb::concurrent_hash_map::accessor`之类的东西。
最终解决方法是使用一把大锁保护`sync.Map`，每次写map都需要拿锁，虽然粒度从pod变大到了整个空间，但是这里锁保护的临界区非常小，只有对于map数据结构的写，很快。并且只有slow path会拿这把大锁，fast path为判断`map[key]`存在的情况下会直接拿锁，并不会拿大锁再创建锁。
在和apiserver连接中断后，任务2会继续做，任务1会每5s再发起一次连接请求，任务3会按之前的频率继续做(由于get请求会失败所以拿不到pod信息)
所有任务用`go`创建不同线程执行，当接收到请求执行耗时任务时，例如创建、删除容器，会再使用`go`额外开辟线程执行，防止阻塞接受下一个指令。

### 4.6 KubeProxy

#### 4.6.1 Overview

Kube-proxy是集群中每个节点（node）上所运行的网络代理， 实现Service概念的一部分，负责维护节点上的一些网络规则， 使得集群内部或外部能够与 Pod 进行网络通信。在本项目中，使用ipvs方式实现了从Service虚拟IP到实际PodIP的转发。

#### 4.6.2 主要工作

简单来说，Kube-proxy的工作就是维护虚拟ServiceIP到实际PodIP之间的映射，从而保证发向虚拟Service的流量能够按照一定的负载均衡策略转发到实际的Pod中。其工作模式与Controller类同，需要监控和管理的资源为Service和Endpoint，主要工作如下：

1. 监听service资源的创建。把对应Cluster IP加入ipvs set中。
2. 监听service资源的删除。在ipvs set中删除对应Cluster IP。
3. 监听endpoint的创建。设置dest规则。
4. 监听endpoint的删除。删除对应dest规则。

#### 4.6.3 工作原理

此处将详细介绍发向虚拟Service的流量是如何转发到实际的Pod中去的。本项目中实现了ClusterIP和NodePort两种服务模式，后者是在前者基础上在每个节点上加了一层从nodeIP：nodePort到ClusterIP的转发规则，故只介绍ClusterIP模式的工作原理。

在Mini-K8s的网络模型中，从某一内部节点发往Pod的流量需要经过以下四步：1. 内部节点通过查询flannel虚拟网卡找到PodIP对应的物理节点。 2. 流量发送至物理节点的网络接口。 3. 流量被转发至flannel网卡对应的网桥。4. 找到Pod的容器对应的veth与host机器上的veth的匹配对，并最终将流量转发至Pod对应的容器中。

例如，从Master节点转发至Pod的流程如下箭头所示，其中eth0是物理节点的网络接口，flannel.1是虚拟网卡，cni0是网桥。

<img src="https://notes.sjtu.edu.cn/uploads/upload_f79c43a5de98e47b72fb1179e91d14c1.png" style="zoom:67%;" />


由此我们可以得知，由于CNI网络模型已经实现了从PodIP到实际Pod的转发，故想要将发往ClusterIP的流量转发至Pod中，只需要实现**从ClusterIP到PodIP**的转发即可。

Kube-proxy通过运行在用户态的ipvsadm提供的CLI接口创建内核级别的ipvs规则，包括将虚拟ip地址添到本地flannel.1网卡以及为虚拟ip添加endpoint（真正的服务节点）。这样，当网络请求下陷到内核态时，会自动根据配置的规则进行包的转发。

ipvs 代理模式基于 netfilter 回调函数，类似于 iptables 模式， 但它使用哈希表作为底层数据结构，在内核空间中生效。 这意味着 IPVS 模式下的 kube-proxy 比 iptables 模式下的 kube-proxy 重定向流量的延迟更低，同步代理规则时性能也更好。 与其他代理模式相比，IPVS 模式还支持更高的网络流量吞吐量。

### 4.7 Kubectl

Kubectl是miniK8s的命令行工具，用于用户和Apiserver进行交互。采用Cobra命令行库实现，支持的命令如下：

#### kubectl apply

```
kubectl apply -f <filename>
```

用于创建/更新资源。

#### kubectl get

```
kubectl get <resource> <name> [-n <namespace>]
kubectl get <resource>+s [-n <namespace>]
```

用于查询资源当前状态。

#### kubectl describe

```
kubectl describe <resource> <name> [-n <namespace>]
kubectl describe <resource>+s [-n <namespace>]
```

用于查询当前资源的详细信息。

#### kubectl delete

```
kubectl delete <resource> <name> [-n <namespace>]
```

用于删除指定资源。

## 5. 所有功能

### 5.1 部署集群

master：

```shell
#!/bin/bash
# check if the ectd is running, if not, start it in the background
# etcd is a progress
if ! pgrep -x "etcd" > /dev/null
then
    echo "etcd is not running, start it"
    nohup etcd &
fi

# check the default systemd-resolved, if it is running, stop it
if pgrep -x "systemd-resolved" > /dev/null
then
    echo "systemd-resolved is running, stop it"
    systemctl stop systemd-resolved
fi

# check if the coredns is running, if not, start it in the background
if ! pgrep -x "coredns" > /dev/null
then
    echo "coredns is not running, start it"
    nohup ./coredns -conf /home/mini-k8s/pkg/kubedns/config/Corefile &
fi

# check the default nginx, if it is running, stop it
if pgrep -x "nginx" > /dev/null
then
    echo "nginx is running, stop it"
    systemctl stop nginx
fi

# start the nginx in the background
echo "start nginx"
nohup nginx -c /home/mini-k8s/pkg/kubedns/config/nginx.conf &

# build the components and run the server
cd /home/mini-k8s/build
make kubectl
make apiserver
make scheduler
make controller
make serverless
make kubeproxy

# create the log directory if not exist
if [ ! -d "/home/mini-k8s/build/log" ]; then
  mkdir /home/mini-k8s/build/log
fi

cd bin

# start the components in different terminals
echo "start the minik8s"
# ./apiserver > ../log/apiserver.log 2> /dev/null &

./apiserver > ../log/apiserver.log 2>&1 &
echo "start apiserver"
sleep 3
./scheduler > ../log/scheduler.log 2>&1 &
echo "start scheduler"
./controller > ../log/controller.log 2>&1 &
echo "start controller"
./kubeproxy > ../log/kubeproxy.log 2>&1 &


chmod +x /home/mini-k8s/pkg/serverless/function/registry.sh
cd /home/mini-k8s/pkg/serverless/function
./registry.sh
cd /home/mini-k8s/build/bin
./serverless  > ../log/serverless.log 2>&1 &
echo "start serverless"
```

worker:

```shell
#!/bin/bash
./kubeproxy > ../log/kubeproxy.log 2>&1 &
./kubelet > ../log/kubelet.log 2>&1 &
```

kubelet默认使用当前目录下的`kubelet-config.yaml`作为配置文件，如果没有默认文件则使用默认配置，详细配置内容见[4.5.2](#4.5.2 运行配置)
可通过`kubectl get nodes`查看集群状态
**之后的所有功能均支持多机**

### 5.2 Pod抽象

支持pod内localhost通信，支持pod间跨node通信
容器支持挂载、环境变量、启动命令、资源限制

```yaml
apiVersion: v1
kind: Pod
metadata:
  name: pod
  namespace: demo
spec:
  containers:
    - name: c1
      image: docker.io/mcastelino/nettools
      ports:
        - containerPort: 12345
      command:
        - /root/test_mount/test_network
      env:
        - name: port
          value: '12345'
      volumeMounts:
        - name: test-volume
          mountPath: /root/test_mount
     - name: c2
      image: ubuntu
      command:
        - /root/test_mount/test_cpu
      volumeMounts:
        - name: test-volume
          mountPath: /root/test_mount
      resources:
        requests:
                cpu: "0.3"
                memory: "50Mi"
        limits:
                cpu: "0.5"
                memory: "100Mi"
    - name: c3
      image: ubuntu
      command:
        - /root/test_mount/test_memory
      volumeMounts:
        - name: test-volume
          mountPath: /root/test_mount
      resources:
        requests:
                cpu: "0.3"
                memory: "50Mi"
        limits:
                cpu: "0.5"
                memory: "100Mi"
  volumes:
    - name: test-volume
      hostPath:
        path: /home/test_mount
```

时序图如下，与k8s完全一致
<img src="https://notes.sjtu.edu.cn/uploads/upload_fb73e291c8c80052029f7a0024690abb.png" style="zoom:67%;" />
具体调度策略见[4.3.4](#4.3.4 Step3: 评分和排序 (policy) )
创建`kubectl apply -f pod.yaml`
查看状态 `kubectl -n namespace get pods`
删除 `kubectl -n namespace delete pod podName`

### 5.3 Service抽象

#### 5.3.1 工作流程

<img src="https://notes.sjtu.edu.cn/uploads/upload_239c715ac23d29fa5e4a73f3fd0e80f2.png" style="zoom:67%;" />


#### 5.3.2 服务类型

**1. ClusterIP**

   通过集群的内部 IP 暴露服务，选择该值，服务只能够在集群内部可以访问，这也是默认的Service类型。ClusterIP类型的service创建时，k8s会通过etcd从可分配的IP池中分配一个IP，该IP全局唯一，且不可修改。所有访问该IP的请求，都会被转发到后端的endpoints中。

   <img src="https://notes.sjtu.edu.cn/uploads/upload_54b40215615e4379c8b7820038b76dd8.png" style="zoom:67%;" />


**2. NodePort**

   通过每个 Node 节点上的 IP 和静态端口（NodePort）暴露服务。NodePort 服务会路由到 ClusterIP 服务，这个 ClusterIP 服务会自动创建。通过请求 NodeIP:NodePort，可以从集群的外部访问一个 NodePort 服务。

   <img src="https://notes.sjtu.edu.cn/uploads/upload_a78d6cdcc888c5b176a06cec2d0b5c1c.png" style="zoom:67%;" />

#### 5.3.3 使用方法

##### 创建服务

创建service时不负责创建service所筛选的pod。因此，如果想要服务可达，需要先创建pod实例（或使用[replicaset抽象](#54-ReplicaSet抽象)创建多个pod）。

支持用声明式的方式创建/修改一个service，只需运行如下命令指定一个配置文件：

```shell
./kubectl apply -f /home/mini-k8s/example/service.yaml 
```

通过`./kubectl get services` 可以看到类似如下输出：

![](https://notes.sjtu.edu.cn/uploads/upload_f978101ccf4d97090b5a55dd12b40420.png)


##### 删除服务

同样的，删除服务时也不会删除对应pod。

```
./kubectl delete service service-practice
```

##### Pod的创建&Endpoint的绑定

Endpoint本质上是clusterIP到PodIP的一对一映射。通过Endpoint抽象，可以实现一个pod对应多个service。Endpoint在service创建时由ServiceController自动创建。对于新添加的满足service的selector筛选条件的pod，ServiceController同样会自动创建对应的Endpoint。除此之外，用户还可以通过自行创建Endpoint，为service添加实例。

配置方式如下：

```yaml
metadata:
  name: endpoint-new
spec:
    svcIP: 10.10.0.1
    svcPort: 23245
    dstIP: 10.2.9.144
    dstPort: 8080
```

并可以通过`./kubectl get endpoints` 查看是否添加成功。

##### 示例

示例详见`/example/service.yaml`文件。

```yaml
apiVersion: v1
kind: Service
metadata:
  name: service-practice
spec:
  selector:
    app: replica-practice
  type: ClusterIP
  ports:
    - name: service-port1
      protocol: TCP
      port: 6692 # ClusterIP对应的端口
      targetPort: p1 # 转发的端口，pod对应的端口
```

重要字段含义如下：

- `metadata.name`：资源的唯一标识符
- `spec.type`：服务类型。仅支持ClusterIP和NodePort。
- `spec.selector`：定义筛选的标签，用于寻找满足条件的pod
- `spec.ports`：定义访问服务的端口和转发的端口。

### 5.4 ReplicaSet抽象

#### 5.4.1 工作流程

<img src="https://notes.sjtu.edu.cn/uploads/upload_112f2142ffcea4303a314ae99fc1deb4.png" style="zoom:67%;" />


*scheduler调度部分被省略

#### 5.4.2 使用方法

##### 创建/修改Repliacaset

mini-K8s支持用声明式的方式创建/修改一个replicaset，只需运行如下命令指定一个配置文件：

```shell
./kubectl apply -f /home/mini-k8s/example/replica.yaml 
```

通过`./kubectl get pods` 可以看到类似如下输出：

![](https://notes.sjtu.edu.cn/uploads/upload_d328f758f69a2edea0ccbb9570e39ce1.png)

##### 扩缩Replicaset

通过更新 `.spec.replicas` 字段，ReplicaSet 可以被轻松地进行扩缩。ReplicasetController能确保受其控制的 Pod 的数量和期望值一致。

在扩容时，ReplicasetController会根据模板创建新的pod。

在缩容时，ReplicasetController会优先选择**创建时间较早**的pod删除。

##### 删除Repliacaset

通过运行以下命令删除Repliacaset以及所有被该Repliacaset控制的pod：

```
./kubectl delete replica replica-practice
```

##### pod的筛选和隔离

replicaset通过`selector`字段筛选满足条件的pod，支持多标签匹配。只有满足所有标签条件的pod才会被选择。**对于单个pod，可以通过改变标签来从 ReplicaSet 中移除 Pod。**

##### pod异常处理

Repliacaset始终致力于保证正在运行的pod数目符合期望副本数。因此，当检测到被控制的pod状态为终止或异常退出时，会删除对应pod并重新启动新的pod。

##### 示例

示例详见`/example/replica.yaml`文件。

重要字段的含义如下：

- `metadata.name`：资源的唯一标识符
- `spec.replicas`：期望副本数
- `spec.selector`：定义筛选的标签，用于寻找满足条件的pod
- `spec.template`：当replicaset对应pod数目不足期望值时，创建的新的pod时使用的模板。注意`.spec.template.metadata.labels` 字段需要与`spec.selector`一致。

```yaml
kind: Replica
apiVersion: apps/v1
metadata:
  name: replica-practice1
spec:
  replicas: 2
  selector:
      app: replica-practice
  template:
    metadata:
      labels:
        app: replica-practice
    spec:
      containers:
        - name: server
          image: docker.io/mcastelino/nettools
          ports:
            - name: p1 # 端口名称
              containerPort: 8080  # 容器端口
          command:
            - /root/test_mount/simple_http_server
          env:
            - name: port
              value: '8080'
          volumeMounts:
            - name: data
              mountPath: /root/test_mount
      volumes:
        - name: data
          hostPath:
            path: /home/test_mount
```

### 5.5 Auto-Scaling

#### 5.5.1 工作流程

<img src="https://notes.sjtu.edu.cn/uploads/upload_692280e6958d05a22ede57e463dc76b9.png" style="zoom:67%;" />
*省略Replicaset更新后和ReplicasetController交互的部分（详见[Replicaset工作流程](#54-ReplicaSet抽象))

#### 5.5.2 扩缩容算法

HPAController每15s进行一次扩缩容检测，根据**监控资源**，**扩缩容标准**，和**扩缩容策略**决定最终扩缩容的行为。扩缩容算法具有一定的自由度，用户可以在Yaml文件中配置相关字段以控制算法行为。MiniK8s中支持的自定义配置如下：

- **监控资源：**

  CPU，memory

- **扩缩容标准：**

  Utilization（占用率）， AverageValue（平均值）

- **扩缩容策略：**

  - scale up， scale down可以分别指定扩容/缩容的策略

  - 策略指定有两种方式：**Pods**和**Percent**，限制一定时间内扩缩容的个数/比例

  - 支持**stabilizationWindowSeconds**字段设定稳定窗口

##### 算法流程

1. 根据`Spec.ScaleTargetRef`字段找到对应的replicaset。`Spec.Metrics`字段提供多个指定的指标，根据每个指标利用metric api向kubelet查询对应pod的度量值并计算当前指标值。如果扩缩容标准设置为`Utilization`，控制器获取每个 Pod 中的容器的资源使用情况， 并计算资源使用率；如果设置为`AverageValue`值，将直接使用原始数据（不再计算百分比）。 

2. 将计算出的结果与指标比较。计算出扩缩容的期望副本数。公式：

   ```
   期望副本数 = ceil[当前副本数 * (当前指标 / 期望指标)]
   ```

   每个指标都会计算出一个期望副本数。取**最大值**作为整体期望副本数。

3. 根据`Spec.Behavior`字段定义的扩缩容行为判断整体期望副本数是否满足条件，并确定最终的期望副本数。需要满足的条件有：

   - 不超过`MaxReplicas`，不小于`MinReplicas`。

   - 上一次扩缩容距今时间大于`StabilizationWindowSeconds`（扩容默认为0，缩容默认为300s）

   - 满足`Spec.Behavior.*.Policies`字段定义的HPAScalingPolicy。（如每3秒最多新增10个pod，每20s最多减少10%的pod）。不同Policy的限制之间可以由`Spec.Behavior.*.SelectPolicy`字段设定取最小/最大限制。

   根据上述三个条件的限制确定最终副本数。

#### 5.5.3 使用方法

##### 创建HPA

mini-K8s支持用声明式的方式创建/修改一个hpa，只需运行如下命令指定一个配置文件：

```shell
./kubectl apply -f /home/mini-k8s/example/hpa/hpa.yaml 
```

当指定控制的资源的副本数小于hpa设置的最小副本数时，首先会扩容至最小副本数。

通过`./kubectl get hpas` 可以看到类似如下输出：

![](https://notes.sjtu.edu.cn/uploads/upload_98d1d05b9b57f44c304753576ab69c5c.png)


##### 示例

详见`/example/hpa`文件夹。

重要字段的含义如下：

- `spec.metrics`：指定期望指标。包括测量资源的类型，测量标准，和期望值。
- `spec.scaleTargetRef`：指定要控制的api object。
- `spec.minReplicas/maxReplicas`：限定扩缩容的最大值/最小值。
- `spec.behavior`：设定扩缩容策略。包括一定时间内扩缩容的个数/比例，以及稳定窗口。

```yaml
apiVersion: apps/v1
kind: HPA
metadata:
  name: hpa-practice
spec:
  minReplicas: 2  # 最小pod数量
  maxReplicas: 5  # 最大pod数量
  metrics:
    - resource:
        name: "memory"
        target:
          averageUtilization: 99
          type: Utilization
      type: Resource
    - resource:
        name: "cpu"
        target:
          averageValue: 1000
          type: AverageValue
      type: Resource
  scaleTargetRef:   # 指定要控制的deploy
    apiVersion:  apps/v1
    kind: replicas
    name: replica-practice
  behavior:
    scaleUp:
      policies:
        - type: Pods
          value: 8
          periodSeconds: 60 # 每分钟最多8个
    scaleDown:
      policies:
        - type: Percent
          value: 10
          periodSeconds: 60 # 每分钟最多10%
      stabilizationWindowSeconds: 30
```

### 5.6 DNS与转发

#### 5.6.1 Overview

MiniK8s集群的DNS服务主要由两部分组成，第一部分由**coreDNS**提供了域名到IP的映射，第二部分由**nginx**提供了不同path到指定service的映射，架构如下所示：

<img src="https://notes.sjtu.edu.cn/uploads/upload_ec406335a8d81e2ee4b9d401a73ab8ce.png" style="zoom:67%;" />


此处以如下dnsrecord为例：

```yaml
kind: dnsrecord
apiVersion: app/v1
name: dns-test1
namespace: default
host: minik8s.com
paths:
  - service: dns-service
    pathName: path1
    port: 22222
  - service: dns-service2
    pathName: path2
    port: 23456
```

#### 5.6.2 coreDNS

> Kubelet在启动容器时，会将`nameserver: [masterIP]`写入容器的`/etc/resolv.conf`文件，从而使得master节点上的coreDNS成为整个集群首选的DNS服务器

coreDNS使用`Corefile` (`pkg/kubedns/config/Corefile`)进行配置，`Corefile`的内容如下所示：

```txt
.:53 {
    etcd {
        endpoint http://localhost:2380
        path /dns
        upstream /etc/resolv.conf
        fallthrough
    }
    forward . 114.114.114.114
    reload 6s
    log
    errors
    loop
    prometheus  # Monitoring plugin
    loadbalance
}
```

coreDNS在master节点的`53`端口监听DNS解析请求，这里使用了etcd插件，用户自定义的域名存储在`/dns`目录下，从而可以支持域名的动态自定义，对于没有匹配的域名信息，转发给上游DNS服务器。

对于如上dnsrecord，会在etcd中更新如下域名信息：

```shell
/dns/com/minik8s
{"host": [masterIp]}
```

 将域名映射到master节点的IP以后，下一步由nginx处理。

#### 5.6.3 Nginx

nginx使用`nginx.conf`文件（`pkg/kubedns/config/nginx.conf`）进行配置, 这个配置文件通过`file template`的方式进行动态更新，每个dnsrecord对应一个HTTP服务器块`server`，它们监听在master节点的`80`端口，其`server_name`对应dnsrecord的`hostname`，通过**反向代理**的方式将不同路径的请求转发到对应的service的`clusterIP` + 指定端口。

以下是如上dnsrecord对应的nginx配置文件，`10.10.0.1`和`10.10.0.2`是对应的service clusterIP。

```conf
worker_processes  5;  ## Default: 1
error_log  ./error.log debug;
pid        ./nginx.pid;
worker_rlimit_nofile 8192;

events {
  worker_connections  4096;  ## Default: 1024
}
http {
    
    server {
        listen 0.0.0.0:80;
        server_name minik8s.com;

        
        location /path1/ {
            access_log /var/log/nginx/access.log;
            proxy_pass http://10.10.0.1:22222/;
        }
        
        location /path2/ {
            access_log /var/log/nginx/access.log;
            proxy_pass http://10.10.0.2:23456/;
        }
        
    }
    
}
```

当client对指定域名和path发送网络请求时，coreDNS首先将域名解析到masterIP，再通过master节点上运行的nginx转发到对应的service的指定端口。

#### 5.6.4 dnsrecord的更新和删除

在dnsrecord更新时，通过更新etcd的方式更新coreDNS中的DNS entries，并更新nginx配置文件`nginx.conf`

### 5.7 容错

MiniK8s主要通过以下操作进行容错：

- 如果只有控制面的APIServer崩溃，其他组件（scheduler，controller，kubelet等组件会定时尝试重新连接）
- 所有的状态都持久化在etcd中，组件都是**stateless**的，如果组件发生崩溃，会重新从etcd中读取数据来保证一致性
- 节点定期给APIServer发送**heartbeat**信息，APIServer通过heartbeat信息检查节点的健康状态
- 节点故障恢复：当一个节点发生故障或不可用时，miniK8s会自动将受影响的容器重新调度到其他健康的节点上。这种故障恢复机制确保应用程序的持续可用性，即使某些节点出现故障也不会影响整个集群的运行

### 5.8 GPU 

#### 5.8.1 Job

提交GPU任务用job完成，只需要一个容器即可。通过查询相关资料发现对于job的运用中，输入输出只能通过挂载这一种方式来完成。
可以在运行kubelet的node上增加label然后用nodeSelector的方式固定job对应的node，用户输入输出都在node上完成。这样的缺点是将node暴露给用户且用户需要切换master与node。最终的解决方案是使用NFS，`/minik8s-sharedata`是minik8s集群都共享的一个文件夹。不论job对应的pod被schedule到哪个node，用户的源文件都能被找到且用户都可以在master上拿到最终结果。

<img src="https://www.z4a.net/images/2023/06/05/gpu.png" style="zoom:67%;" />

为了在不破坏job通用性的情况下满足yaml中配置与slurm对齐，这里采用环境变量的方式传递信息。
对用户的要求：

1. 自己编写编译脚本，包括`make run`与`make build`
2. 需要的输出放在result文件夹中
3. 通过环境变量传入slurm参数
4. 使用提供的gpu-server镜像

```yaml
apiVersion: v1
kind: Job
metadata:
  name: matrix-add
  namespace: gpu
spec:
  containers:
    - name: gpu-server
      image: gpu-server
      command:
        - "./job.py"
      env:
        - name: source-path
          value: /gpu
        - name: job-name
          value: matrix-add
        - name: partition
          value: dgx2
        - name: "N"
          value: "1"
        - name: ntasks-per-node
          value: "1"
        - name: cpus-per-task
          value: "6"
        - name: gres
          value: gpu:1
      volumeMounts:
        - name: share-data
          mountPath: /gpu
  volumes:
    - name: share-data
      hostPath:
        path: /minik8s-sharedata/gpu/matrix-add

  backoffLimit: 3
  ttlSecondsAfterFinished: 10
```

可通过`kubectl get jobs/pods`查看job和对应的pod状态

#### 5.8.2 镜像

容器需要完成的工作

1. 通过环境变量指定参数构造slurm脚本
2. 通过挂载拿到用户的源文件并使用scp上传到hpc服务器
3. 通过ssh与hpc服务器交互，提交GPU任务并拿到job id
4. 通过`sacct`命令等待job完成
5. 通过scp下载结果

这里选择使用python而并非简单的shell脚本+expect是因为更强大的ssh客户端可以通过交互拿到结果，并且用程序语言可以实现更复杂的功能，例如等待job完成而并非提交就结束。

```python
#!/usr/bin/python3
import paramiko
from time import sleep
from scp import SCPClient
from os import getenv
NREAD = 100000
ssh = paramiko.SSHClient()
ssh.set_missing_host_key_policy(paramiko.AutoAddPolicy())
ssh.connect("pilogin.hpc.sjtu.edu.cn",22,"stu1648")
job_submit_tag = "Submitted batch job"
line_finish_tag = "[stu1648@"
PENDING = "PENDING"
COMPLETED = "COMPLETED"
FAILED = "FAILED"
source_path = getenv("source-path")
job_name = getenv("job-name")
partition= getenv("partition")
N = getenv("N")
ntasks_per_node = getenv("ntasks-per-node")
cpus_per_task = getenv("cpus-per-task")
gres = getenv("gres")
if not job_name or not source_path:
    print("env error")
    exit(1)
if source_path[-1] == "/":
    # scp send whole source
    source_path = source_path[:-1]

if not partition:
    partition = "dgx2"
if not N:
    N = 1
if not ntasks_per_node:
    ntasks_per_node = 1
if not cpus_per_task:
    cpus_per_task = 6
if not gres:
    gres = "gpu:1"

def generate_slurm():
    print("generating slurm")
    with open(f"./{job_name}.slurm","w") as f:
        f.write("#!/bin/bash\n")

        f.write(f"#SBATCH --job-name={job_name}\n")
        f.write(f"#SBATCH --partition={partition}\n")
        f.write(f"#SBATCH -N {N}\n")
        f.write(f"#SBATCH --ntasks-per-node={ntasks_per_node}\n")
        f.write(f"#SBATCH --cpus-per-task={cpus_per_task}\n")
        f.write(f"#SBATCH --gres={gres}\n")


        # result must exist . is the same dir as .slurm
        f.write(f"#SBATCH --output=result/output.txt\n")
        f.write(f"#SBATCH --error=result/error.txt\n")

        f.write(f"ulimit -s unlimited\n")
        f.write(f"ulimit -l unlimited\n")

        f.write("module load gcc/8.3.0 cuda/10.1.243-gcc-8.3.0\n")

        f.write("make build\n")
        f.write("make run\n")
def upload_source():
    print("uploading source")
    scp = SCPClient(ssh.get_transport(),socket_timeout=16)
    scp.put(source_path,recursive=True,remote_path=f"~/{job_name}")
    scp.put(f"./{job_name}.slurm",f"~/{job_name}/{job_name}.slurm")
    scp.close()


def download_result(job_id):
    print("downloading result")
    scp = SCPClient(ssh.get_transport(),socket_timeout=16)
    scp = SCPClient(ssh.get_transport(),socket_timeout=16)
    scp.get(f"~/{job_name}/result",recursive=True,local_path=f"{source_path}/")
    scp.close()
def submit_job():
    t = 3
    while t:
        s = ssh.invoke_shell()
        print("starting ssh")
        sleep(2)
        recv = s.recv(NREAD).decode('utf-8')
        if recv.find("stu1648") == -1:
            print("start ssh failed,retrying")
            t -= 1
            sleep(5)
            continue
        print("start ssh success")
        print("sending sbatch")
        s.send(f"cd ~/{job_name} && sbatch ./{job_name}.slurm\n")
        sleep(5)
        
        recv = s.recv(NREAD).decode("utf-8")
        index = recv.find(job_submit_tag)
        if index == -1:
            print("sbatch failed,retrying")
            t -= 1
            sleep(5)
            continue
        print("sbatch success")
        job_id = recv[index+len(job_submit_tag)+1:recv.index(line_finish_tag)-2]
        print(f"{job_id=}")
        print("start checking job status")
        check_status_cmd = f"sacct | grep {job_id} | awk '{{print $6}}'"
        while True:
            s.send(check_status_cmd+"\n")
            sleep(2)
            recv = s.recv(NREAD).decode("utf-8")
            status = recv[recv.index(check_status_cmd)+len(check_status_cmd)+2:recv.index(line_finish_tag)-2]
            print(f"{status=}")
            if status.find(FAILED)!=-1:
                print("job failed")
                return job_id
            if status.find(COMPLETED)==-1:
                sleep(10)
            else:
                return job_id
generate_slurm()
upload_source()
job_id = submit_job()
if job_id:
    download_result(job_id)
print("finish")
```

镜像需要包含的环境为ssh scp python 
且为了避免仓库中泄露hpc服务器的密码，这里使用rsa公私钥的方式登录

```dockerfile
FROM ubuntu
RUN apt-get update
RUN apt-get -y install openssh-server python3-pip vim
RUN pip3 install paramiko scp
WORKDIR /root
RUN mkdir  .ssh
COPY ./id_rsa .ssh/id_rsa
COPY ./known_hosts .ssh/known_hosts
COPY ./job.py job.py
```

安装buildkit，使用`nerdctl -n namespace build`进行构建
或用docker进行构建后push到自己的registry，再使用`nerdctl -n namespace pull`

#### 5.8.3 cuda程序

cuda编程主要就是拆分和并行的元素至不同线程执行。通常来说对于矩阵这种大小不固定的任务，给定thread和block值是不合适的，因此这里选择由硬件参数设置一个合适的thread、block值，每个thread以一定跨度完成不止一个矩阵元素的计算；同时使用dim3方便索引；使用prefetch提前将数据拿到gpu上避免page fault带来的性能损失。
为了让任务更有意义，这里提供输入输出模块，使得用户可以自己提供输入并且输出写到文件中便于更有效地获取结果。同时编写一次cpu的矩阵函数验证gpu核函数的正确性。

```c++
#include <stdio.h>
#include <files.h>
#define CHECK_CORRECTNESS

#define N  10000

__global__ void matrixAddGPU( double * a, double * b, double * c )
{

  int row_begin = blockIdx.x * blockDim.x + threadIdx.x;
  int col_begin = blockIdx.y * blockDim.y + threadIdx.y;
  int stride_row = gridDim.x * blockDim.x;
  int stride_col = gridDim.y * blockDim.y;

  for(int row = row_begin; row < N ;row += stride_row) {
        for(int col= col_begin; col< N ; col+= stride_col) {
                c[row * N + col] = a[row*N+col] + b[row*N+col];
        }
  }
}

void matrixAddCPU( double * a, double * b, double * c )
{

  for( int row = 0; row < N; ++row )
    for( int col = 0; col < N; ++col )
    {
      c[row * N + col] = a[row*N+col]+b[row*N+col];
    }
}

int main()
{
        cudaError_t cudaStatus;

  int deviceId;
  int numberOfSMs;

  cudaGetDevice(&deviceId);
  cudaDeviceGetAttribute(&numberOfSMs, cudaDevAttrMultiProcessorCount, deviceId);
  printf("SM:%d\n",numberOfSMs);//80

  double *a, *b, *c_gpu;

  unsigned long long size = (unsigned long long)N * N * sizeof (double); // Number of bytes of an N x N matrix

  // Allocate memory
  cudaMallocManaged (&a, size);
  cudaMallocManaged (&b, size);
  cudaMallocManaged (&c_gpu, size);
  read_values_from_file("matrix_a_data", a, size);
  read_values_from_file("matrix_b_data", b, size);

  //if too large,invalid configuration argument
  dim3 threads_per_block(32,32,1);
  dim3 number_of_blocks (16*numberOfSMs,16*numberOfSMs, 1);
  cudaMemPrefetchAsync(a, size, deviceId);
  cudaMemPrefetchAsync(b, size, deviceId);
  cudaMemPrefetchAsync(c_gpu, size, deviceId);
  matrixAddGPU <<< number_of_blocks, threads_per_block >>> ( a, b, c_gpu );
        cudaStatus = cudaGetLastError();
        if (cudaStatus != cudaSuccess) {
                fprintf(stderr, "call matrixAddGPU error: %s\n", cudaGetErrorString(cudaStatus));
                return -1;
        }

  cudaDeviceSynchronize(); // Wait for the GPU to finish before proceeding

  // Call the CPU version to check our work
    // Compare the two answers to make sure they are equal
  bool error = false;
  #ifdef CHECK_CORRECTNESS
    double *c_cpu;
    cudaMallocManaged (&c_cpu, size);
    matrixAddCPU( a, b, c_cpu );
    for( int row = 0; row < N && !error; ++row )
      for( int col = 0; col < N && !error; ++col )
        if (c_cpu[row * N + col] != c_gpu[row * N + col])
        {
          printf("FOUND ERROR at c[%d][%d]\n", row, col);
          error = true;
          break;
        }
    cudaFree( c_cpu );
  #endif
  if (!error)
    printf("Success!\n");
  write_values_to_file("result/matrix_c_data", c_gpu, size);
  // Free all our allocated memory
  cudaFree(a);
  cudaFree(b);
  cudaFree( c_gpu );
}
```

矩阵乘法核函数如下

```c++
__global__ void matrixMulGPU( double * a, double * b, double * c )
{

  int row_begin = blockIdx.x * blockDim.x + threadIdx.x;
  int col_begin = blockIdx.y * blockDim.y + threadIdx.y;
  int stride_row = gridDim.x * blockDim.x;
  int stride_col = gridDim.y * blockDim.y;

  for(int row = row_begin; row < N ;row += stride_row) {
        for(int col= col_begin; col< N ; col+= stride_col) {
                double val = 0;
                for(int k = 0; k < N; ++k ){
                        val += a[row * N + k] * b[k * N + col];
                        c[row * N + col] = val;
                }
        }
  }
}
```

### 5.9 Serverless

#### 5.9.1 Overview

serverless模块建立在miniK8s的基本组件之上，依旧使用了APIserver作为功能的统一入口，通过watch机制，触发函数的创建、更新、删除以及function和workflow的调用

#### 5.9.2 函数的创建、更新和删除

目前serverless模块可以实现对于python的支持

* 创建
  1. 当 APIserver接收到 function注册的请求以后，会将function的元信息存储在etcd中
  2. serverless模块watch到function创建的请求后，会使用相应的代码文件制作镜像，镜像的`REPOSITORY `统一为`localhost:5000/[function_name]`,  要求用户上传的函数签名必须为`run`, 对于函数的参数和返回值并没有规定，在容器中使用使用`Flask`实现了一个简单的http server，在`8081`端口监听用户对于函数的调用请求
  3. serverless模块制作完镜像后，将该镜像推送到本地的**docker registry**中，后续node可以从其中拉去制作好的镜像
  4. 为该function生成相应的ReplicaSet，初始`replica`数量为0，在`Spec.Containers[0].Image`指明镜像信息为`master:5000/[function_name]`，从而kubelet在启动pod时会从master节点的docker registry而非dockerhub中拉取镜像；将该ReplicaSet的创建请求发送给APIserver，由APIserver和controller负责后续ReplicaSet的管理
  5. 将创建结果返回给client

- 删除
  1. 首先向APIserver发送请求，删除相应的replicaSet
  2. 从docker registry中删除相应的镜像

- 更新
  1. 当APIserver接收到function的更新信息后，会将function的元信息更新在etcd中
  2. serverless模块watch到function更新的请求后，会将原先的ReplicaSet scale to zero, 同时创建新的ReplicaSet和Image (在代码中，我们虽然删除了原先的ReplicaSet，新创建了一个ReplicaSet，但是由于两个函数的元数据（name等信息）相同，所以创建出来的ReplicaSet的元数据相同，这实际上是一个`scale to zero`→`scale up`的过程)

为了下面自动扩缩容的实现需求，会为所有function维护一个`RecordMap`, 记录了一个该function一个窗口内的访问数量等信息，在创建、删除和更新function的时候也会同步更新`RecordMap`中的内容，这里为了保证并发访问的安全性和效率，对于`RecordMap`的访问采用了一个**读写锁保护**

以下的时序图展示了function创建、更新和删除时各个模块的配合：
<img src="https://www.z4a.net/images/2023/06/03/function1.png" alt="function1.png" style="zoom:80%;" />



#### 5.9.3 函数的调用和自动扩缩容 

函数的调用中蕴涵了自动扩容的机制，函数调用和自动扩缩容的流程如下图所示：

<img src="https://www.z4a.net/images/2023/06/03/trigger.png" alt="trigger.png" style="zoom:80%;" />



- 扩缩容的总体策略是**按需扩容，定期缩容**
- 函数调用时，如果在这次的调用的窗口期内，**函数的实例数大于trigger的次数，小于threshold，会进行扩容，实例数量+1**
- 一个goroutine以30s为一个window，定期对于函数的调用情况进行检查，**将函数的实例数目scale down到上一个window的调用次数，重置当前window函数的调用计数**
- corner case的检查：
  - **情况1**: 新增实例时，如果pod已经启动，但是容器中的http server并没有启动完毕，如果这个时候给pod发送http trigger的请求，就会出错
    - 解决方案：增加`checkConnection`保证目前的实例已经是可用状态
  - **情况2**: 如果多个实例同时请求，新实例对应的pod启动的时间过长（比如函数依赖很多，需要安装很多第三方库），这时候恰好触发了缩容，`available pods`就无法达到最初的`expection`
    - 解决方案：**乐观**思想，超时重试，如果三次retry不成功，就直接按照当前实例情况进行调度

#### 5.9.4 Workflow定义和运⾏流程 

> Workflow的定义详见4.1.9 WorkFlow

- workflow的创建、修改和删除

  workflow的内容持久化在etcd中，APIserver负责workflow的创建修改和删除

- workflow的运行

  当用户发送workflow运行的请求到APIserver时，serverless模块watch到这个请求并获得workflow当前的内容和运行的参数，以**有向图遍历**方式从`startAt`节点运行各个状态上的函数并进行分支选取操作，最后向用户返回运行结果（如果运行出错，返回报错信息）
