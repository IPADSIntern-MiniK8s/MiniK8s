# CI/CD

### test

[go 覆盖测试工具介绍 - 建站教程 (jiuaidu.com)](https://jiuaidu.com/jianzhan/1046052/)

`go test ./...`可以测试目录下所有的test文件

`go test minik8s/pkg/kubelet/container` 测试指定包下的测试文件

### gitlab runner

#### docker

`docker run -d --name gitlab-runner --restart always   -v /srv/gitlab-runner/config:/etc/gitlab-runner   -v /var/run/docker.sock:/var/run/docker.sock   gitlab/gitlab-runner:v15.10.1`

执行器选择docker 这里镜像需要先在主机上写Dockerfile手动构建好，然后修改`config.toml`配置文件把`pull_policy`修改为`if-not-present`

对于简单测试没问题，但是对于CNI这种复杂的东西，即使加了privilege=true，还是会出现和宿主机上不一样的情况。

#### host

[Install GitLab Runner | GitLab](https://docs.gitlab.com/runner/install/)

交大云主机安装二进制

`nslookup www.ipads.sjtu.edu.cn` 安全组开放所有端口

`gitlab-runner register` 去gitlab网页的settings/cicd复制url和token

执行器选择shell 在主机上给gitlab-runner用户足够的权限

[【汇总】解决GitLab-Runner执行脚本命令无权限_gitlab-runner 提升权限_成为大佬先秃头的博客-CSDN博客](https://blog.csdn.net/qq_39940674/article/details/127616784)

采用这种方法进行CI/CD，gitlab-runner会在主机上的某个目录跑脚本，用的都是主机的环境

- 优点：不需要手动配一个拥有所有环境的镜像；没有容器导致的与主机不一致，跑不起来的情况。
- 缺点：会对主机产生影响；在缺少依赖的情况下无法更换gitlab-runner所在主机。

### .gitlab-ci.yml

1. prepare: 设置go env，防止go test在download时超时

2. test：`go test` 如果测试涉及到的api需要权限，需要加sudo

   创建多个tag为shell的runner，使test阶段并行测试 （目前一共3个）

   需要修改手动`/etc/gitlab-runner/config.toml`的concurrent为3

3. build：`go build` 生成可执行文件在`/home/gitlab-runner/$CI_COMMIT_BRANCH/`目录下

   不同分支build出的文件不会互相覆盖

### 代码同步

同时推送到gitee和gitlab，不然无法用gitlab-runner

[git push origin master一次提交多个远程仓库 - 兜里还剩五块出头 - 博客园 (cnblogs.com)](https://www.cnblogs.com/hmy-666/p/17304317.html)

```shell
root@minik8s-1:/mini-k8s# git remote -v
origin  https://gitee.com/szy_0127/mini-k8s.git (fetch)
origin  https://gitee.com/szy_0127/mini-k8s.git (push)
origin  https://ipads.se.sjtu.edu.cn:2020/520021910933/minik8s.git (push)
```

