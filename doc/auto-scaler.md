# Auto-Scaling

## Metric指标

kubelet通过内置CAdvisor获得资源具体使用情况，auto-scaler通过api向kubelet获取资源使用数据，**不经过api-server**（因为这部分数据不需要持久化）。

### api格式

#### REST API

仅支持GET格式REST请求。

- `/nodes` - all node metrics; type `[]NodeMetrics`
- `/nodes/{node}` - metrics for a specified node; type `NodeMetrics`
- `/namespaces/{namespace}/pods` - all pod metrics within namespace with support for `all-namespaces`; type `[]PodMetrics`
- `/namespaces/{namespace}/pods/{pod}` - metrics for a specified pod; type `PodMetrics`

**向kubelet请求的格式：**

仅支持GET格式REST请求。

- `/node` - node metric; type `NodeMetrics`

- `/pods` - all pod metrics ; type `[]PodMetrics`
- `/namespaces/{namespace}/pods/{pod}` - metrics for a specified pod; type `PodMetrics`

<u>**kubelet在10250端口监听请求**</u>

#### APIObject

详见`pkg/apiobject/metrics.go`

### 单位约定

**资源用量均为int型整数**，单位换算方式如下：

- cpu  ： k8s的1000 = cpu的一个核

   	 如果一台服务器cpu是4核 那么 k8s单位表示就是 4* 1000

- 内存 : k8s的8320MI = 8320 * 1024 * 1024 字节

​    	1MI = 1024*1024 字节

​    	同理 1024MI /1024 = 1G

## controllers

kube-controller-manager 是控制平面的组件， 负责运行控制器进程。

从逻辑上讲， 每个控制器都是一个单独的进程， 但是为了降低复杂性，它们都被编译到同一个可执行文件，并在同一个进程中运行。

这些控制器包括：

节点控制器（Node Controller）：负责在节点出现故障时进行通知和响应

任务控制器（Job Controller）：监测代表一次性任务的 Job 对象，然后创建 Pods 来运行这些任务直至完成

端点分片控制器（EndpointSlice controller）：填充端点分片（EndpointSlice）对象（以提供 Service 和 Pod 之间的链接）。

服务账号控制器（ServiceAccount controller）：为新的命名空间创建默认的服务账号（ServiceAccount）

服务控制器（Service controller）：负责创建和更新 Endpoints 对象（以匹配 Service 对象选择器中定义的任何 Pod）

### replicaSet Controller
- 为了实现replicaSet的功能，需要一个控制器来监控pod的状态，当pod的状态不符合预期时，控制器会自动创建或删除pod，使pod的数量符合预期。
- 这个时候需要为pod增加uid

### deployment Controller

### HPA Controller
#### HorizontalPodAutoscaler 是如何工作的？
Kubernetes 将水平 Pod 自动扩缩实现为一个间歇运行的控制回路（它不是一个连续的过程）。间隔由 kube-controller-manager 的 --horizontal-pod-autoscaler-sync-period 参数设置（默认间隔为 15 秒）。

在每个时间段内，控制器管理器都会根据每个 HorizontalPodAutoscaler 定义中指定的指标查询资源利用率。 控制器管理器找到由 scaleTargetRef 定义的目标资源，然后根据目标资源的 .spec.selector 标签选择 Pod， 并从资源指标 API（针对每个 Pod 的资源指标）或自定义指标获取指标 API（适用于所有其他指标）。

对于按 Pod 统计的资源指标（如 CPU），控制器从资源指标 API 中获取每一个 HorizontalPodAutoscaler 指定的 Pod 的度量值，如果设置了目标使用率，控制器获取每个 Pod 中的容器资源使用情况， 并计算资源使用率。如果设置了 target 值，将直接使用原始数据（不再计算百分比）。 接下来，控制器根据平均的资源使用率或原始值计算出扩缩的比例，进而计算出目标副本数。

需要注意的是，如果 Pod 某些容器不支持资源采集，那么控制器将不会使用该 Pod 的 CPU 使用率。 下面的算法细节章节将会介绍详细的算法。

如果 Pod 使用自定义指示，控制器机制与资源指标类似，区别在于自定义指标只使用原始值，而不是使用率。
如果 Pod 使用对象指标和外部指标（每个指标描述一个对象信息）。 这个指标将直接根据目标设定值相比较，并生成一个上面提到的扩缩比例。 在 autoscaling/v2 版本 API 中，这个指标也可以根据 Pod 数量平分后再计算。
HorizontalPodAutoscaler 的常见用途是将其配置为从聚合 API （metrics.k8s.io、custom.metrics.k8s.io 或 external.metrics.k8s.io）获取指标。 metrics.k8s.io API 通常由名为 Metrics Server 的插件提供，需要单独启动。有关资源指标的更多信息， 请参阅 Metrics Server。

对 Metrics API 的支持解释了这些不同 API 的稳定性保证和支持状态。

HorizontalPodAutoscaler 控制器访问支持扩缩的相应工作负载资源（例如：Deployment 和 StatefulSet）。 这些资源每个都有一个名为 scale 的子资源，该接口允许你动态设置副本的数量并检查它们的每个当前状态。 有关 Kubernetes API 子资源的一般信息， 请参阅 Kubernetes API 概念。

#### 算法细节
从最基本的角度来看，Pod 水平自动扩缩控制器根据当前指标和期望指标来计算扩缩比例。

期望副本数 = ceil[当前副本数 * (当前指标 / 期望指标)]
例如，如果当前指标值为 200m，而期望值为 100m，则副本数将加倍， 因为 200.0 / 100.0 == 2.0 如果当前值为 50m，则副本数将减半， 因为 50.0 / 100.0 == 0.5。如果比率足够接近 1.0（在全局可配置的容差范围内，默认为 0.1）， 则控制平面会跳过扩缩操作。

如果 HorizontalPodAutoscaler 指定的是 targetAverageValue 或 targetAverageUtilization， 那么将会把指定 Pod 度量值的平均值做为 currentMetricValue。

在检查容差并决定最终值之前，控制平面还会考虑是否缺少任何指标， 以及有多少 Pod Ready。

所有设置了删除时间戳的 Pod（带有删除时间戳的对象正在关闭/移除的过程中）都会被忽略， 所有失败的 Pod 都会被丢弃。

如果某个 Pod 缺失度量值，它将会被搁置，只在最终确定扩缩数量时再考虑。

当使用 CPU 指标来扩缩时，任何还未就绪（还在初始化，或者可能是不健康的）状态的 Pod 或 最近的指标度量值采集于就绪状态前的 Pod，该 Pod 也会被搁置。

由于技术限制，HorizontalPodAutoscaler 控制器在确定是否保留某些 CPU 指标时无法准确确定 Pod 首次就绪的时间。 相反，如果 Pod 未准备好并在其启动后的一个可配置的短时间窗口内转换为准备好，它会认为 Pod “尚未准备好”。 该值使用 --horizontal-pod-autoscaler-initial-readiness-delay 标志配置，默认值为 30 秒。 一旦 Pod 准备就绪，如果它发生在自启动后较长的、可配置的时间内，它就会认为任何向准备就绪的转换都是第一个。 该值由 -horizontal-pod-autoscaler-cpu-initialization-period 标志配置，默认为 5 分钟。

在排除掉被搁置的 Pod 后，扩缩比例就会根据 currentMetricValue/desiredMetricValue 计算出来。

如果缺失某些度量值，控制平面会更保守地重新计算平均值，在需要缩小时假设这些 Pod 消耗了目标值的 100%， 在需要放大时假设这些 Pod 消耗了 0% 目标值。这可以在一定程度上抑制扩缩的幅度。

此外，如果存在任何尚未就绪的 Pod，工作负载会在不考虑遗漏指标或尚未就绪的 Pod 的情况下进行扩缩， 控制器保守地假设尚未就绪的 Pod 消耗了期望指标的 0%，从而进一步降低了扩缩的幅度。

考虑到尚未准备好的 Pod 和缺失的指标后，控制器会重新计算使用率。 如果新的比率与扩缩方向相反，或者在容差范围内，则控制器不会执行任何扩缩操作。 在其他情况下，新比率用于决定对 Pod 数量的任何更改。

注意，平均利用率的 原始 值是通过 HorizontalPodAutoscaler 状态体现的， 而不考虑尚未准备好的 Pod 或缺少的指标，即使使用新的使用率也是如此。

如果创建 HorizontalPodAutoscaler 时指定了多个指标， 那么会按照每个指标分别计算扩缩副本数，取最大值进行扩缩。 如果任何一个指标无法顺利地计算出扩缩副本数（比如，通过 API 获取指标时出错）， 并且可获取的指标建议缩容，那么本次扩缩会被跳过。 这表示，如果一个或多个指标给出的 desiredReplicas 值大于当前值，HPA 仍然能实现扩容。

最后，在 HPA 控制器执行扩缩操作之前，会记录扩缩建议信息。 控制器会在操作时间窗口中考虑所有的建议信息，并从中选择得分最高的建议。 这个值可通过 kube-controller-manager 服务的启动参数 --horizontal-pod-autoscaler-downscale-stabilization 进行配置， 默认值为 5 分钟。 这个配置可以让系统更为平滑地进行缩容操作，从而消除短时间内指标值快速波动产生的影响。

