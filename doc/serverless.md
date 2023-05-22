# Serverless

## Overview
### 需要的组件
Autoscaler（自动缩放器）：
Autoscaler 是 Knative 中的一个组件，用于根据工作负载的需求自动调整底层的资源（如 Pod）数量。它通过监测当前工作负载的指标（如请求数、CPU 使用率等）来动态地调整副本数量，以确保应用程序能够根据需求自动扩展或收缩。Autoscaler 可以自动处理流量峰值、负载均衡和资源利用效率等方面的调整，以提供更高的可伸缩性和资源效率。

它主要有两项工作：一个是负责把pod启动起来，另外一个是把启动中的请求转发给pod。

Activator（激活器）：
Activator 是 Knative 中的另一个组件，用于处理请求的激活和暂停。它负责在请求到达时将处于休眠状态的应用程序实例（如休眠的 Pod）唤醒，以处理传入的请求。Activator 监测流量并维护一组活动的应用程序实例，根据需要将流量路由到相应的实例。这种激活和休眠的机制可以帮助节省资源，避免持续运行的实例浪费资源，提高整体的资源利用率。

## Function
> 要求：⽀持Function抽象。⽤⼾可以通过单个⽂件（zip包或代码⽂件）定义函数内容，通过指令上传给 minik8s，并且通过http trigger调⽤函数。
### 实现
1. 在`apiobject`中增加`function`抽象
2. 当`apiserver`接收到`function`注册的请求以后，会使用相应的代码文件制作镜像
3. `apiserver`会将镜像推送到`registry`中, 于此同时，`apiserver`会将`function`的信息存储到`etcd`中

### 坏境搭建
#### docker registry
> 用于镜像的存储和分发
1. 拉取镜像
```shell
docker pull registry
```
2. 启动容器
```shell
docker run -d -p 5000:5000 --restart=always --name registry registry
```
3. 验证容器是否启动成功
```shell
docker ps
```
4. 验证docker registry的功能
```shell
docker pull testcontainers/helloworld
docker tag testcontainers/helloworld localhost:5000/helloworld
docker push localhost:5000/helloworld

# 拉取本地镜像
docker pull localhost:5000/helloworld:latest
# 运行对应的容器
docker run -it --rm localhost:5000/helloworld:latest
```
运行效果检验：
```shell
2023/05/20 09:01:25 DELAY_START_MSEC: 0
2023/05/20 09:01:25 Sleeping for 0 ms
2023/05/20 09:01:25 Starting server on port 8080
2023/05/20 09:01:25 Sleeping for 0 ms
2023/05/20 09:01:25 Starting server on port 8081
2023/05/20 09:01:25 Ready, listening on 8080 and 8081
```
5. 验证是否可以使用containerd运行镜像
```shell
# 从registry中拉取镜像
docker pull localhost:5000/helloworld:latest
# 保存镜像
docker save localhost:5000/helloworld:latest -o helloworld.tar
# 导入镜像
ctr i import helloworld.tar
# 查看镜像
ctr i ls
# 输出如下：
# REF                              TYPE    ...
# localhost:5000/helloworld:latest application/vnd.docker.distribution.manifest.v2+json
# 根据ref信息，运行镜像
ctr run --rm -t localhost:5000/helloworld:latest helloworld
```
运行效果检验：
```shell
2023/05/20 09:09:56 DELAY_START_MSEC: 0
2023/05/20 09:09:56 Sleeping for 0 ms
2023/05/20 09:09:56 Starting server on port 8080
2023/05/20 09:09:56 Sleeping for 0 ms
2023/05/20 09:09:56 Starting server on port 8081
2023/05/20 09:09:56 Ready, listening on 8080 and 8081
```

6. 使用go api操作 docker registry
```shell
go get github.com/docker/docker/client
```

#### docker registry 对应的命令

1. 启动 Docker Registry 容器：确保 Docker Registry 容器正在运行。如果尚未启动，请使用以下命令启动容器：

   ```shell
   docker run -d -p 5000:5000 --restart=always --name registry registry:2
   ```

   这将在后台运行一个 Registry 容器，并将容器的 5000 端口映射到主机的 5000 端口。

2. 构建镜像并标记：使用 Docker CLI 构建一个新的镜像，并为该镜像添加 Registry 的地址和标签。例如，假设你有一个名为 `myimage` 的镜像，可以执行以下命令：

   ```shell
   docker build -t myimage .
   docker tag myimage localhost:5000/myimage:latest
   ```

   这将构建 `myimage` 镜像并为其添加 `localhost:5000` Registry 的地址和 `latest` 标签。

3. 推送镜像到 Registry：使用 `docker push` 命令将镜像推送到 Registry。执行以下命令：

   ```shell
   docker push localhost:5000/myimage:latest
   ```

   这将把镜像推送到 `localhost:5000` Registry。

4. 拉取镜像：使用 `docker pull` 命令从 Registry 拉取镜像。执行以下命令：

   ```shell
   docker pull localhost:5000/myimage:latest
   ```

   这将从 Registry 拉取 `myimage` 镜像的最新版本。

5. 运行容器：使用拉取的镜像运行一个容器来验证容器存储的功能。执行以下命令：

   ```shell
   docker run -it --rm localhost:5000/myimage:latest
   ```

   这将在容器中运行 `myimage` 镜像，并在终端中显示容器的输出。如果容器成功运行并显示预期的输出，说明容器存储功能正常。

6. 在另外一台机器上使用这台机器上的镜像
- 获得运行权限
```shell
chmod +x /pkg/serverless/registry.sh
```
- 运行脚本, 获得从registry中获得镜像的权限
```shell
./pkg/serverless/registry.sh
```
- 从registry中拉取镜像
```shell
docker pull 192.168.1.13:5000/helloworld:latest
```

#### flask
1. flask version：2.3.2
2. 每个function都会对应一个flask server，用来接受外界的调用请求，并返回结果
3. 接受POST请求，使用**字典解包**的形式传递参数给function
4. 启动`imagedata`下的`flask`服务，进行测试

function对应的python代码（func.py）：
```python
def run(x, y):
    z = x + y
    print(x)
    print(y)
    print(z)
    print('Hello, World!')
    
```
测试命令：
```shell
curl -X POST -H "Content-Type: application/json" -d '{"x": 3, "y": 5}' http://127.0.0.1:8081/
```

#### 构建image
- 生成镜像
详见`imagedata`下的`Dockerfile`文件，注意这里出现了超时错误，所以我更换了清华源
```shell
docker build -t function /home/mini-k8s/pkg/serverless/imagedata/
```
- 运行镜像
```shell
docker run -p 8081:8081 --name serverless -d function
```
- 进入容器
```shell
docker exec -it serverless /bin/bash
```
- 查看docker对应的ip地址
```shell
docker inspect -f '{{range .NetworkSettings.Networks}}{{.IPAddress}}{{end}}' serverless
```
输出结果：`172.17.0.3`
- 测试
```shell
curl -X POST -H "Content-Type: application/json" -d '{"x": 3, "y": 5}' http://172.17.0.3:8081/
```

### 具体实现
1. upload function
用户发送`function`的注册请求（`api/v1/functions`），`apiserver`会将`function`的信息存储到`etcd`中，同时会将`function`的代码文件制作成镜像，并推送到`registry`中

这个时候，serverless部分会监听`etcd`中`function`的变化，创建相应的replicaSet，初始的replica数量为0，当`function`的replica数量发生变化时，会自动调整`function`的replica数量

2. http trigger
用户发送`http trigger`的请求，`apiserver`会将请求转发给`serverless`，`serverless`会查找是否有合适的pod，向对应的pod发送请求，如果没有合适的pod，首先修改replicaSet的replica数量，当合适的pod被创建以后，向pod发送请求


> 注意事项：
> 1. serverless对应的pod和replicaSet的名称是一致的，并且都在`serverless`的命名空间下
> 2. serverless对应的pod和replicaSet的名称是`function`的名称，所以`function`的名称不能重复


### 参考资料
https://blog.csdn.net/zw0Pi8G5C1x/article/details/123784951

