# Details

Date : 2023-05-22 17:27:14

Directory /home/mini-k8s

Total : 110 files,  9919 codes, 0 comments, 1381 blanks, all 11300 lines

[Summary](results.md) / Details / [Diff Summary](diff.md) / [Diff Details](diff-details.md)

## Files
| filename | language | code | comment | blank | total |
| :--- | :--- | ---: | ---: | ---: | ---: |
| [cmd/apiserver.go](/cmd/apiserver.go) | Go | 5 | 0 | 3 | 8 |
| [cmd/controller.go](/cmd/controller.go) | Go | 5 | 0 | 3 | 8 |
| [cmd/kubectl.go](/cmd/kubectl.go) | Go | 10 | 0 | 3 | 13 |
| [cmd/kubelet.go](/cmd/kubelet.go) | Go | 47 | 0 | 8 | 55 |
| [cmd/kubeproxy.go](/cmd/kubeproxy.go) | Go | 5 | 0 | 3 | 8 |
| [cmd/scheduler.go](/cmd/scheduler.go) | Go | 52 | 0 | 9 | 61 |
| [cmd/serverless.go](/cmd/serverless.go) | Go | 5 | 0 | 4 | 9 |
| [go.mod](/go.mod) | Go Module File | 95 | 0 | 4 | 99 |
| [go.sum](/go.sum) | Go Checksum File | 1,188 | 0 | 1 | 1,189 |
| [pkg/apiobject/autoscaler.go](/pkg/apiobject/autoscaler.go) | Go | 239 | 0 | 32 | 271 |
| [pkg/apiobject/dnsrecord.go](/pkg/apiobject/dnsrecord.go) | Go | 38 | 0 | 8 | 46 |
| [pkg/apiobject/dnsrecord_test.go](/pkg/apiobject/dnsrecord_test.go) | Go | 10 | 0 | 2 | 12 |
| [pkg/apiobject/doc.go](/pkg/apiobject/doc.go) | Go | 1 | 0 | 1 | 2 |
| [pkg/apiobject/endpoint.go](/pkg/apiobject/endpoint.go) | Go | 37 | 0 | 7 | 44 |
| [pkg/apiobject/function.go](/pkg/apiobject/function.go) | Go | 29 | 0 | 6 | 35 |
| [pkg/apiobject/metrics.go](/pkg/apiobject/metrics.go) | Go | 93 | 0 | 15 | 108 |
| [pkg/apiobject/node.go](/pkg/apiobject/node.go) | Go | 207 | 0 | 23 | 230 |
| [pkg/apiobject/node_test.go](/pkg/apiobject/node_test.go) | Go | 28 | 0 | 6 | 34 |
| [pkg/apiobject/object.go](/pkg/apiobject/object.go) | Go | 24 | 0 | 7 | 31 |
| [pkg/apiobject/pod.go](/pkg/apiobject/pod.go) | Go | 166 | 0 | 25 | 191 |
| [pkg/apiobject/pod_test.go](/pkg/apiobject/pod_test.go) | Go | 19 | 0 | 6 | 25 |
| [pkg/apiobject/replication.go](/pkg/apiobject/replication.go) | Go | 95 | 0 | 18 | 113 |
| [pkg/apiobject/replication_test.go](/pkg/apiobject/replication_test.go) | Go | 23 | 0 | 4 | 27 |
| [pkg/apiobject/service.go](/pkg/apiobject/service.go) | Go | 127 | 0 | 26 | 153 |
| [pkg/apiobject/utils/duration.go](/pkg/apiobject/utils/duration.go) | Go | 41 | 0 | 9 | 50 |
| [pkg/apiobject/utils/quantity.go](/pkg/apiobject/utils/quantity.go) | Go | 10 | 0 | 6 | 16 |
| [pkg/apiobject/utils/time.go](/pkg/apiobject/utils/time.go) | Go | 152 | 0 | 30 | 182 |
| [pkg/controller/HPAcontroller.go](/pkg/controller/HPAcontroller.go) | Go | 281 | 0 | 37 | 318 |
| [pkg/controller/manager.go](/pkg/controller/manager.go) | Go | 19 | 0 | 6 | 25 |
| [pkg/controller/rscontroller.go](/pkg/controller/rscontroller.go) | Go | 188 | 0 | 40 | 228 |
| [pkg/controller/svccontroller.go](/pkg/controller/svccontroller.go) | Go | 188 | 0 | 40 | 228 |
| [pkg/controller/svccontroller_test.go](/pkg/controller/svccontroller_test.go) | Go | 94 | 0 | 15 | 109 |
| [pkg/kubeapiserver/apimachinery/apiserver.go](/pkg/kubeapiserver/apimachinery/apiserver.go) | Go | 98 | 0 | 13 | 111 |
| [pkg/kubeapiserver/apimachinery/routeInstaller.go](/pkg/kubeapiserver/apimachinery/routeInstaller.go) | Go | 40 | 0 | 8 | 48 |
| [pkg/kubeapiserver/controlplane/instance.go](/pkg/kubeapiserver/controlplane/instance.go) | Go | 24 | 0 | 4 | 28 |
| [pkg/kubeapiserver/controlplane/service.go](/pkg/kubeapiserver/controlplane/service.go) | Go | 13 | 0 | 2 | 15 |
| [pkg/kubeapiserver/doc.go](/pkg/kubeapiserver/doc.go) | Go | 1 | 0 | 1 | 2 |
| [pkg/kubeapiserver/handlers/dnshandler.go](/pkg/kubeapiserver/handlers/dnshandler.go) | Go | 222 | 0 | 31 | 253 |
| [pkg/kubeapiserver/handlers/endpointhandler.go](/pkg/kubeapiserver/handlers/endpointhandler.go) | Go | 227 | 0 | 32 | 259 |
| [pkg/kubeapiserver/handlers/functionhandler.go](/pkg/kubeapiserver/handlers/functionhandler.go) | Go | 166 | 0 | 26 | 192 |
| [pkg/kubeapiserver/handlers/handlertable.go](/pkg/kubeapiserver/handlers/handlertable.go) | Go | 32 | 0 | 7 | 39 |
| [pkg/kubeapiserver/handlers/nodehandler.go](/pkg/kubeapiserver/handlers/nodehandler.go) | Go | 126 | 0 | 18 | 144 |
| [pkg/kubeapiserver/handlers/nodehandler_test.go](/pkg/kubeapiserver/handlers/nodehandler_test.go) | Go | 52 | 0 | 11 | 63 |
| [pkg/kubeapiserver/handlers/podhandler.go](/pkg/kubeapiserver/handlers/podhandler.go) | Go | 333 | 0 | 45 | 378 |
| [pkg/kubeapiserver/handlers/podhandler_test.go](/pkg/kubeapiserver/handlers/podhandler_test.go) | Go | 78 | 0 | 22 | 100 |
| [pkg/kubeapiserver/handlers/replicahandler.go](/pkg/kubeapiserver/handlers/replicahandler.go) | Go | 224 | 0 | 32 | 256 |
| [pkg/kubeapiserver/handlers/replicationhandler.go](/pkg/kubeapiserver/handlers/replicationhandler.go) | Go | 216 | 0 | 32 | 248 |
| [pkg/kubeapiserver/handlers/routeInstaller.go](/pkg/kubeapiserver/handlers/routeInstaller.go) | Go | 32 | 0 | 6 | 38 |
| [pkg/kubeapiserver/handlers/servicehandler.go](/pkg/kubeapiserver/handlers/servicehandler.go) | Go | 227 | 0 | 34 | 261 |
| [pkg/kubeapiserver/run.go](/pkg/kubeapiserver/run.go) | Go | 11 | 0 | 3 | 14 |
| [pkg/kubeapiserver/storage/ectd_test.go](/pkg/kubeapiserver/storage/ectd_test.go) | Go | 90 | 0 | 15 | 105 |
| [pkg/kubeapiserver/storage/etcd.go](/pkg/kubeapiserver/storage/etcd.go) | Go | 236 | 0 | 30 | 266 |
| [pkg/kubeapiserver/testing/basicserver_test.go](/pkg/kubeapiserver/testing/basicserver_test.go) | Go | 15 | 0 | 4 | 19 |
| [pkg/kubeapiserver/watch/list.go](/pkg/kubeapiserver/watch/list.go) | Go | 42 | 0 | 16 | 58 |
| [pkg/kubeapiserver/watch/watch.go](/pkg/kubeapiserver/watch/watch.go) | Go | 84 | 0 | 15 | 99 |
| [pkg/kubeapiserver/watch/watchtable.go](/pkg/kubeapiserver/watch/watchtable.go) | Go | 5 | 0 | 4 | 9 |
| [pkg/kubectl/cmd/apply.go](/pkg/kubectl/cmd/apply.go) | Go | 41 | 0 | 8 | 49 |
| [pkg/kubectl/cmd/delete.go](/pkg/kubectl/cmd/delete.go) | Go | 37 | 0 | 7 | 44 |
| [pkg/kubectl/cmd/describe.go](/pkg/kubectl/cmd/describe.go) | Go | 55 | 0 | 12 | 67 |
| [pkg/kubectl/cmd/get.go](/pkg/kubectl/cmd/get.go) | Go | 127 | 0 | 13 | 140 |
| [pkg/kubectl/cmd/root.go](/pkg/kubectl/cmd/root.go) | Go | 27 | 0 | 10 | 37 |
| [pkg/kubectl/doc.go](/pkg/kubectl/doc.go) | Go | 1 | 0 | 0 | 1 |
| [pkg/kubectl/test/http.go](/pkg/kubectl/test/http.go) | Go | 17 | 0 | 3 | 20 |
| [pkg/kubectl/test/http_test.go](/pkg/kubectl/test/http_test.go) | Go | 17 | 0 | 3 | 20 |
| [pkg/kubectl/test/kubectl_test.go](/pkg/kubectl/test/kubectl_test.go) | Go | 18 | 0 | 4 | 22 |
| [pkg/kubectl/utils/utils.go](/pkg/kubectl/utils/utils.go) | Go | 39 | 0 | 8 | 47 |
| [pkg/kubedns/nginx/nginx.tmpl](/pkg/kubedns/nginx/nginx.tmpl) | Go Template File | 14 | 0 | 2 | 16 |
| [pkg/kubedns/nginx/nginxeditor.go](/pkg/kubedns/nginx/nginxeditor.go) | Go | 87 | 0 | 14 | 101 |
| [pkg/kubedns/nginx/nginxeditor_test.go](/pkg/kubedns/nginx/nginxeditor_test.go) | Go | 54 | 0 | 4 | 58 |
| [pkg/kubelet/container/clientutil.go](/pkg/kubelet/container/clientutil.go) | Go | 7 | 0 | 3 | 10 |
| [pkg/kubelet/container/container.go](/pkg/kubelet/container/container.go) | Go | 238 | 0 | 17 | 255 |
| [pkg/kubelet/container/container_test.go](/pkg/kubelet/container/container_test.go) | Go | 234 | 0 | 15 | 249 |
| [pkg/kubelet/container/containerutil.go](/pkg/kubelet/container/containerutil.go) | Go | 75 | 0 | 13 | 88 |
| [pkg/kubelet/kubelet.go](/pkg/kubelet/kubelet.go) | Go | 113 | 0 | 8 | 121 |
| [pkg/kubelet/metricsserver/handler.go](/pkg/kubelet/metricsserver/handler.go) | Go | 21 | 0 | 6 | 27 |
| [pkg/kubelet/metricsserver/metricserver.go](/pkg/kubelet/metricsserver/metricserver.go) | Go | 28 | 0 | 6 | 34 |
| [pkg/kubelet/pod/pod.go](/pkg/kubelet/pod/pod.go) | Go | 223 | 0 | 19 | 242 |
| [pkg/kubelet/pod/pod_test.go](/pkg/kubelet/pod/pod_test.go) | Go | 261 | 0 | 12 | 273 |
| [pkg/kubelet/pod/podutil.go](/pkg/kubelet/pod/podutil.go) | Go | 38 | 0 | 5 | 43 |
| [pkg/kubelet/pod/podutil_test.go](/pkg/kubelet/pod/podutil_test.go) | Go | 29 | 0 | 5 | 34 |
| [pkg/kubelet/run.go](/pkg/kubelet/run.go) | Go | 22 | 0 | 4 | 26 |
| [pkg/kubelet/utils/helper.go](/pkg/kubelet/utils/helper.go) | Go | 24 | 0 | 6 | 30 |
| [pkg/kubeproxy/ipvs/ops.go](/pkg/kubeproxy/ipvs/ops.go) | Go | 124 | 0 | 21 | 145 |
| [pkg/kubeproxy/ipvs/state.go](/pkg/kubeproxy/ipvs/state.go) | Go | 14 | 0 | 5 | 19 |
| [pkg/kubeproxy/proxy.go](/pkg/kubeproxy/proxy.go) | Go | 127 | 0 | 28 | 155 |
| [pkg/kubeproxy/proxy_test.go](/pkg/kubeproxy/proxy_test.go) | Go | 42 | 0 | 6 | 48 |
| [pkg/kubescheduler/doc.go](/pkg/kubescheduler/doc.go) | Go | 1 | 0 | 1 | 2 |
| [pkg/kubescheduler/filter/configfilter.go](/pkg/kubescheduler/filter/configfilter.go) | Go | 210 | 0 | 29 | 239 |
| [pkg/kubescheduler/filter/configfilter_test.go](/pkg/kubescheduler/filter/configfilter_test.go) | Go | 80 | 0 | 26 | 106 |
| [pkg/kubescheduler/filter/templatefilter.go](/pkg/kubescheduler/filter/templatefilter.go) | Go | 9 | 0 | 3 | 12 |
| [pkg/kubescheduler/policy/lrscheduler.go](/pkg/kubescheduler/policy/lrscheduler.go) | Go | 54 | 0 | 9 | 63 |
| [pkg/kubescheduler/policy/lrscheduler_test.go](/pkg/kubescheduler/policy/lrscheduler_test.go) | Go | 41 | 0 | 9 | 50 |
| [pkg/kubescheduler/policy/resourcescheduler.go](/pkg/kubescheduler/policy/resourcescheduler.go) | Go | 60 | 0 | 11 | 71 |
| [pkg/kubescheduler/policy/resourcescheduler_test.go](/pkg/kubescheduler/policy/resourcescheduler_test.go) | Go | 46 | 0 | 13 | 59 |
| [pkg/kubescheduler/policy/templatescheduler.go](/pkg/kubescheduler/policy/templatescheduler.go) | Go | 16 | 0 | 6 | 22 |
| [pkg/kubescheduler/run.go](/pkg/kubescheduler/run.go) | Go | 104 | 0 | 20 | 124 |
| [pkg/kubescheduler/testutils/builder.go](/pkg/kubescheduler/testutils/builder.go) | Go | 107 | 0 | 5 | 112 |
| [pkg/serverless/activator/deploy.go](/pkg/serverless/activator/deploy.go) | Go | 247 | 0 | 35 | 282 |
| [pkg/serverless/activator/deploy_test.go](/pkg/serverless/activator/deploy_test.go) | Go | 26 | 0 | 5 | 31 |
| [pkg/serverless/autoscaler/metric.go](/pkg/serverless/autoscaler/metric.go) | Go | 97 | 0 | 14 | 111 |
| [pkg/serverless/autoscaler/record.go](/pkg/serverless/autoscaler/record.go) | Go | 33 | 0 | 14 | 47 |
| [pkg/serverless/eventfilter/functionwatcher.go](/pkg/serverless/eventfilter/functionwatcher.go) | Go | 128 | 0 | 27 | 155 |
| [pkg/serverless/function/image.go](/pkg/serverless/function/image.go) | Go | 94 | 0 | 17 | 111 |
| [pkg/serverless/imagedata/requirements.txt](/pkg/serverless/imagedata/requirements.txt) | pip requirements | 1 | 0 | 1 | 2 |
| [pkg/serverless/run.go](/pkg/serverless/run.go) | Go | 9 | 0 | 2 | 11 |
| [utils/client.go](/utils/client.go) | Go | 116 | 0 | 12 | 128 |
| [utils/config.go](/utils/config.go) | Go | 12 | 0 | 5 | 17 |
| [utils/http.go](/utils/http.go) | Go | 53 | 0 | 10 | 63 |
| [utils/rand.go](/utils/rand.go) | Go | 88 | 0 | 13 | 101 |
| [utils/utils.go](/utils/utils.go) | Go | 18 | 0 | 9 | 27 |

[Summary](results.md) / Details / [Diff Summary](diff.md) / [Diff Details](diff-details.md)