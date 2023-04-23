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

## 实现思路
1. 原始的版本watch初步使用简单的websocket实现

## 测试命令
```shell
wscat -c ws://localhost:8080/api/v1/nodes/node-1/watch
```
## 参考资料

Gin安装入门: https://gin-gonic.com/zh-cn/docs/quickstart/