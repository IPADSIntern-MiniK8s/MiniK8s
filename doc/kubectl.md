# Mini-K8s kubectl指令手册
Mini-K8s支持的命令如下：
#### kubectl apply

`kubectl apply -f <filename>`

#### kubectl get

`kubectl get <resource> <name> [-n <namespace>]`

`kubectl get <resource>+s [-n <namespace>]`

#### kubectl delete

`kubectl delete <resource> <name> [-n <namespace>]`

由于k8s的Api是基于REST的设计思想，因此，不同种类的HTTP请求也就对应了不同的操作。比较常用的对应关系是：

**GET（SELECT）**：从服务器取出资源（一项或多项）。GET请求对应k8s api的获取信息功能。因此，如果是获取信息的命令都要使用GET方式发起HTTP请求。

**POST（CREATE）**：在服务器新建一个资源。POST请求对应k8s api的创建功能。因此，需要创建Pods、ReplicaSet或者service的时候请使用这种方式发起请求。

**PUT（UPDATE）**：在服务器更新资源（客户端提供改变后的完整资源）。对应更新nodes或Pods的状态、ReplicaSet的自动备份数量等等。

**PATCH（UPDATE）**：在服务器更新资源（客户端提供改变的属性）。

**DELETE（DELETE）**：从服务器删除资源。在稀牛学院的学员使用完毕环境后，可以使用这种方式将Pod删除，释放资源。
