# minik8s-group11

## 文档

- [kubectl](https://ipads.se.sjtu.edu.cn:2020/520021910933/minik8s/-/blob/develop/doc/kubectl-api.md)
- [kubelet](https://ipads.se.sjtu.edu.cn:2020/520021910933/minik8s/-/blob/develop/doc/kubelet.md)
- [apiserver](https://ipads.se.sjtu.edu.cn:2020/520021910933/minik8s/-/blob/develop/doc/apiserver.md)
- [kubeproxy](https://ipads.se.sjtu.edu.cn:2020/520021910933/minik8s/-/blob/develop/doc/kubeproxy.md)
- [scheduler](https://ipads.se.sjtu.edu.cn:2020/520021910933/minik8s/-/blob/develop/doc/scheduler.md)
- [dns](https://ipads.se.sjtu.edu.cn:2020/520021910933/minik8s/-/blob/develop/doc/dns.md)
- [ci/cd](https://ipads.se.sjtu.edu.cn:2020/520021910933/minik8s/-/blob/develop/doc/CICD.md)

## 启动

主机3：

1. `./apiserver`
2. `./scheduler` (config中policy指定frequency)

主机1、2：

`./kubelet` (config中指定主机3的地址和自己的IP、子网范围)

## 使用

主机3：

```shell
root@master:~# ./kubectl apply -f test.yaml
url:http://localhost:8080/api/v1/namespaces/testpod/pods
testpod configured
root@master:~# ./kubectl apply -f test1.yaml
url:http://localhost:8080/api/v1/namespaces/testpod/pods
testpod1 configured
```

主机1：

```shell
root@minik8s-1:~# nerdctl -n testpod ps
CONTAINER ID    IMAGE                                                            COMMAND                   CREATED          STATUS    PORTS    NAMES
85f7d62fca86    registry.cn-hangzhou.aliyuncs.com/google_containers/pause:3.6    "/pause"                  8 minutes ago    Up                 testpod-pause
testpod-c1      docker.io/mcastelino/nettools:latest                             "/root/test_mount/te…"    8 minutes ago    Up
testpod-c2      docker.io/mcastelino/nettools:latest                             "/root/test_mount/te…"    8 minutes ago    Up
root@minik8s-1:~# nerdctl -n testpod inspect -f '{{.NetworkSettings.IPAddress}}' testpod-c1
10.2.17.233
```

主机2：

```shell
root@minik8s-2:~# nerdctl -n testpod ps
CONTAINER ID    IMAGE                                                            COMMAND                   CREATED          STATUS    PORTS    NAMES
1a03a3b0dc04    registry.cn-hangzhou.aliyuncs.com/google_containers/pause:3.6    "/pause"                  8 minutes ago    Up                 testpod1-pause
testpod1-c1     docker.io/mcastelino/nettools:latest                             "/root/test_mount/te…"    8 minutes ago    Up
testpod1-c2     docker.io/mcastelino/nettools:latest                             "/root/test_mount/te…"    8 minutes ago    Up
root@minik8s-2:~# nerdctl -n testpod exec -it testpod1-c1 curl 10.2.17.233:12345
http connect success
root@minik8s-2:~# nerdctl -n testpod exec -it testpod1-c1 curl 10.2.17.233:23456
http connect success
```

主机3：

```shell
root@master:~# ./kubectl -n testpod delete pod testpod
url:http://localhost:8080/api/v1/namespaces/testpod/pods/testpod
testpod deleted
root@master:~# ./kubectl -n testpod delete pod testpod1
url:http://localhost:8080/api/v1/namespaces/testpod/pods/testpod1
testpod1 deleted
```

## 测试使用配置

yaml：

```yaml
apiVersion: v1
kind: Pod
metadata:
  name: testpod
  namespace: testpod
  labels:
    app: example
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
      image: docker.io/mcastelino/nettools
      ports:
        - containerPort: 23456
      command:
        - /root/test_mount/test_network
      env:
        - name: port
          value: '23456'
      volumeMounts:
        - name: test-volume
          mountPath: /root/test_mount
  volumes:
    - name: test-volume
      hostPath:
        path: /home/test_mount

```

test_network

```go
package main

import (
    "io"
    "os"
    "net/http"
)

func main() {
    port:=os.Getenv("port")
    http.HandleFunc("/",func(w http.ResponseWriter,request *http.Request){io.WriteString(w,"http connect success\n")})
    _ = http.ListenAndServe(":"+port, nil)
}
```

kubelet-config.yaml (主机1)

```yaml
ApiserverAddr : 192.168.1.13:8080
FlannelSubnet : 10.2.17.1/24
IP            : 192.168.1.12
```

scheduler-config.yaml

```yaml
ApiserverAddr : localhost:8080
Policy: "frequency"
```



