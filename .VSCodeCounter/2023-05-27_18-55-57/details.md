# Details

Date : 2023-05-27 18:55:57

Directory /home/mini-k8s

Total : 122 files,  11953 codes, 0 comments, 1728 blanks, all 13681 lines

[Summary](results.md) / Details / [Diff Summary](diff.md) / [Diff Details](diff-details.md)

## Files
| filename | language | code | comment | blank | total |
| :--- | :--- | ---: | ---: | ---: | ---: |
| [cmd/apiserver.go](/cmd/apiserver.go) | Go | 5 | 0 | 3 | 8 |
| [cmd/controller.go](/cmd/controller.go) | Go | 5 | 0 | 3 | 8 |
| [cmd/kubectl.go](/cmd/kubectl.go) | Go | 10 | 0 | 3 | 13 |
| [cmd/kubelet.go](/cmd/kubelet.go) | Go | 54 | 0 | 10 | 64 |
| [cmd/kubeproxy.go](/cmd/kubeproxy.go) | Go | 5 | 0 | 3 | 8 |
| [cmd/scheduler.go](/cmd/scheduler.go) | Go | 51 | 0 | 9 | 60 |
| [cmd/serverless.go](/cmd/serverless.go) | Go | 5 | 0 | 4 | 9 |
| [config/config.go](/config/config.go) | Go | 14 | 0 | 5 | 19 |
| [go.mod](/go.mod) | Go Module File | 96 | 0 | 4 | 100 |
| [go.sum](/go.sum) | Go Checksum File | 684 | 0 | 1 | 685 |
| [pkg/apiobject/autoscaler.go](/pkg/apiobject/autoscaler.go) | Go | 356 | 0 | 40 | 396 |
| [pkg/apiobject/dnsrecord.go](/pkg/apiobject/dnsrecord.go) | Go | 39 | 0 | 8 | 47 |
| [pkg/apiobject/dnsrecord_test.go](/pkg/apiobject/dnsrecord_test.go) | Go | 10 | 0 | 2 | 12 |
| [pkg/apiobject/doc.go](/pkg/apiobject/doc.go) | Go | 1 | 0 | 1 | 2 |
| [pkg/apiobject/endpoint.go](/pkg/apiobject/endpoint.go) | Go | 37 | 0 | 7 | 44 |
| [pkg/apiobject/function.go](/pkg/apiobject/function.go) | Go | 29 | 0 | 6 | 35 |
| [pkg/apiobject/function_test.go](/pkg/apiobject/function_test.go) | Go | 18 | 0 | 5 | 23 |
| [pkg/apiobject/job.go](/pkg/apiobject/job.go) | Go | 89 | 0 | 13 | 102 |
| [pkg/apiobject/metrics.go](/pkg/apiobject/metrics.go) | Go | 93 | 0 | 15 | 108 |
| [pkg/apiobject/node.go](/pkg/apiobject/node.go) | Go | 207 | 0 | 23 | 230 |
| [pkg/apiobject/node_test.go](/pkg/apiobject/node_test.go) | Go | 28 | 0 | 6 | 34 |
| [pkg/apiobject/object.go](/pkg/apiobject/object.go) | Go | 24 | 0 | 7 | 31 |
| [pkg/apiobject/pod.go](/pkg/apiobject/pod.go) | Go | 171 | 0 | 25 | 196 |
| [pkg/apiobject/pod_test.go](/pkg/apiobject/pod_test.go) | Go | 19 | 0 | 6 | 25 |
| [pkg/apiobject/replication.go](/pkg/apiobject/replication.go) | Go | 101 | 0 | 16 | 117 |
| [pkg/apiobject/replication_test.go](/pkg/apiobject/replication_test.go) | Go | 23 | 0 | 4 | 27 |
| [pkg/apiobject/service.go](/pkg/apiobject/service.go) | Go | 127 | 0 | 26 | 153 |
| [pkg/apiobject/service_test.go](/pkg/apiobject/service_test.go) | Go | 44 | 0 | 3 | 47 |
| [pkg/apiobject/utils/duration.go](/pkg/apiobject/utils/duration.go) | Go | 41 | 0 | 9 | 50 |
| [pkg/apiobject/utils/quantity.go](/pkg/apiobject/utils/quantity.go) | Go | 10 | 0 | 6 | 16 |
| [pkg/apiobject/utils/time.go](/pkg/apiobject/utils/time.go) | Go | 152 | 0 | 30 | 182 |
| [pkg/apiobject/workflow.go](/pkg/apiobject/workflow.go) | Go | 165 | 0 | 26 | 191 |
| [pkg/apiobject/workflow_test.go](/pkg/apiobject/workflow_test.go) | Go | 148 | 0 | 10 | 158 |
| [pkg/controller/HPAcontroller.go](/pkg/controller/HPAcontroller.go) | Go | 306 | 0 | 42 | 348 |
| [pkg/controller/jobcontroller.go](/pkg/controller/jobcontroller.go) | Go | 93 | 0 | 31 | 124 |
| [pkg/controller/manager.go](/pkg/controller/manager.go) | Go | 23 | 0 | 8 | 31 |
| [pkg/controller/rscontroller.go](/pkg/controller/rscontroller.go) | Go | 192 | 0 | 41 | 233 |
| [pkg/controller/svccontroller.go](/pkg/controller/svccontroller.go) | Go | 189 | 0 | 40 | 229 |
| [pkg/controller/svccontroller_test.go](/pkg/controller/svccontroller_test.go) | Go | 95 | 0 | 15 | 110 |
| [pkg/kubeapiserver/apimachinery/apiserver.go](/pkg/kubeapiserver/apimachinery/apiserver.go) | Go | 98 | 0 | 15 | 113 |
| [pkg/kubeapiserver/apimachinery/routeInstaller.go](/pkg/kubeapiserver/apimachinery/routeInstaller.go) | Go | 40 | 0 | 8 | 48 |
| [pkg/kubeapiserver/doc.go](/pkg/kubeapiserver/doc.go) | Go | 1 | 0 | 1 | 2 |
| [pkg/kubeapiserver/handlers/dnshandler.go](/pkg/kubeapiserver/handlers/dnshandler.go) | Go | 254 | 0 | 32 | 286 |
| [pkg/kubeapiserver/handlers/endpointhandler.go](/pkg/kubeapiserver/handlers/endpointhandler.go) | Go | 227 | 0 | 32 | 259 |
| [pkg/kubeapiserver/handlers/functionhandler.go](/pkg/kubeapiserver/handlers/functionhandler.go) | Go | 279 | 0 | 43 | 322 |
| [pkg/kubeapiserver/handlers/handlertable.go](/pkg/kubeapiserver/handlers/handlertable.go) | Go | 57 | 0 | 11 | 68 |
| [pkg/kubeapiserver/handlers/hpahandler.go](/pkg/kubeapiserver/handlers/hpahandler.go) | Go | 269 | 0 | 39 | 308 |
| [pkg/kubeapiserver/handlers/jobhandler.go](/pkg/kubeapiserver/handlers/jobhandler.go) | Go | 226 | 0 | 32 | 258 |
| [pkg/kubeapiserver/handlers/nodehandler.go](/pkg/kubeapiserver/handlers/nodehandler.go) | Go | 126 | 0 | 18 | 144 |
| [pkg/kubeapiserver/handlers/nodehandler_test.go](/pkg/kubeapiserver/handlers/nodehandler_test.go) | Go | 52 | 0 | 11 | 63 |
| [pkg/kubeapiserver/handlers/podhandler.go](/pkg/kubeapiserver/handlers/podhandler.go) | Go | 492 | 0 | 61 | 553 |
| [pkg/kubeapiserver/handlers/podhandler_test.go](/pkg/kubeapiserver/handlers/podhandler_test.go) | Go | 78 | 0 | 22 | 100 |
| [pkg/kubeapiserver/handlers/replicahandler.go](/pkg/kubeapiserver/handlers/replicahandler.go) | Go | 224 | 0 | 32 | 256 |
| [pkg/kubeapiserver/handlers/routeInstaller.go](/pkg/kubeapiserver/handlers/routeInstaller.go) | Go | 32 | 0 | 6 | 38 |
| [pkg/kubeapiserver/handlers/servicehandler.go](/pkg/kubeapiserver/handlers/servicehandler.go) | Go | 227 | 0 | 34 | 261 |
| [pkg/kubeapiserver/handlers/workflowhandler.go](/pkg/kubeapiserver/handlers/workflowhandler.go) | Go | 217 | 0 | 36 | 253 |
| [pkg/kubeapiserver/run.go](/pkg/kubeapiserver/run.go) | Go | 11 | 0 | 3 | 14 |
| [pkg/kubeapiserver/storage/ectd_test.go](/pkg/kubeapiserver/storage/ectd_test.go) | Go | 90 | 0 | 15 | 105 |
| [pkg/kubeapiserver/storage/etcd.go](/pkg/kubeapiserver/storage/etcd.go) | Go | 246 | 0 | 30 | 276 |
| [pkg/kubeapiserver/watch/list.go](/pkg/kubeapiserver/watch/list.go) | Go | 42 | 0 | 16 | 58 |
| [pkg/kubeapiserver/watch/watch.go](/pkg/kubeapiserver/watch/watch.go) | Go | 66 | 0 | 15 | 81 |
| [pkg/kubeapiserver/watch/watchtable.go](/pkg/kubeapiserver/watch/watchtable.go) | Go | 5 | 0 | 4 | 9 |
| [pkg/kubectl/cmd/apply.go](/pkg/kubectl/cmd/apply.go) | Go | 47 | 0 | 8 | 55 |
| [pkg/kubectl/cmd/delete.go](/pkg/kubectl/cmd/delete.go) | Go | 37 | 0 | 7 | 44 |
| [pkg/kubectl/cmd/describe.go](/pkg/kubectl/cmd/describe.go) | Go | 55 | 0 | 12 | 67 |
| [pkg/kubectl/cmd/get.go](/pkg/kubectl/cmd/get.go) | Go | 167 | 0 | 15 | 182 |
| [pkg/kubectl/cmd/root.go](/pkg/kubectl/cmd/root.go) | Go | 27 | 0 | 10 | 37 |
| [pkg/kubectl/doc.go](/pkg/kubectl/doc.go) | Go | 1 | 0 | 0 | 1 |
| [pkg/kubectl/test/http.go](/pkg/kubectl/test/http.go) | Go | 17 | 0 | 3 | 20 |
| [pkg/kubectl/test/http_test.go](/pkg/kubectl/test/http_test.go) | Go | 17 | 0 | 3 | 20 |
| [pkg/kubectl/test/kubectl_test.go](/pkg/kubectl/test/kubectl_test.go) | Go | 18 | 0 | 4 | 22 |
| [pkg/kubectl/utils/utils.go](/pkg/kubectl/utils/utils.go) | Go | 39 | 0 | 9 | 48 |
| [pkg/kubedns/nginx/nginx.tmpl](/pkg/kubedns/nginx/nginx.tmpl) | Go Template File | 14 | 0 | 2 | 16 |
| [pkg/kubedns/nginx/nginxeditor.go](/pkg/kubedns/nginx/nginxeditor.go) | Go | 83 | 0 | 14 | 97 |
| [pkg/kubedns/nginx/nginxeditor_test.go](/pkg/kubedns/nginx/nginxeditor_test.go) | Go | 54 | 0 | 4 | 58 |
| [pkg/kubelet/container/container.go](/pkg/kubelet/container/container.go) | Go | 245 | 0 | 17 | 262 |
| [pkg/kubelet/container/container_test.go](/pkg/kubelet/container/container_test.go) | Go | 262 | 0 | 15 | 277 |
| [pkg/kubelet/container/containerutil.go](/pkg/kubelet/container/containerutil.go) | Go | 64 | 0 | 12 | 76 |
| [pkg/kubelet/image/image.go](/pkg/kubelet/image/image.go) | Go | 52 | 0 | 4 | 56 |
| [pkg/kubelet/image/image_test.go](/pkg/kubelet/image/image_test.go) | Go | 30 | 0 | 3 | 33 |
| [pkg/kubelet/kubelet.go](/pkg/kubelet/kubelet.go) | Go | 133 | 0 | 13 | 146 |
| [pkg/kubelet/metricsserver/handler.go](/pkg/kubelet/metricsserver/handler.go) | Go | 27 | 0 | 6 | 33 |
| [pkg/kubelet/metricsserver/metricserver.go](/pkg/kubelet/metricsserver/metricserver.go) | Go | 28 | 0 | 6 | 34 |
| [pkg/kubelet/pod/pod.go](/pkg/kubelet/pod/pod.go) | Go | 246 | 0 | 28 | 274 |
| [pkg/kubelet/pod/pod_test.go](/pkg/kubelet/pod/pod_test.go) | Go | 312 | 0 | 16 | 328 |
| [pkg/kubelet/pod/podutil.go](/pkg/kubelet/pod/podutil.go) | Go | 38 | 0 | 5 | 43 |
| [pkg/kubelet/pod/podutil_test.go](/pkg/kubelet/pod/podutil_test.go) | Go | 29 | 0 | 5 | 34 |
| [pkg/kubelet/run.go](/pkg/kubelet/run.go) | Go | 25 | 0 | 4 | 29 |
| [pkg/kubelet/utils/helper.go](/pkg/kubelet/utils/helper.go) | Go | 42 | 0 | 8 | 50 |
| [pkg/kubeproxy/ipvs/ops.go](/pkg/kubeproxy/ipvs/ops.go) | Go | 124 | 0 | 21 | 145 |
| [pkg/kubeproxy/ipvs/state.go](/pkg/kubeproxy/ipvs/state.go) | Go | 14 | 0 | 5 | 19 |
| [pkg/kubeproxy/proxy.go](/pkg/kubeproxy/proxy.go) | Go | 128 | 0 | 28 | 156 |
| [pkg/kubeproxy/proxy_test.go](/pkg/kubeproxy/proxy_test.go) | Go | 42 | 0 | 6 | 48 |
| [pkg/kubescheduler/doc.go](/pkg/kubescheduler/doc.go) | Go | 1 | 0 | 1 | 2 |
| [pkg/kubescheduler/filter/configfilter.go](/pkg/kubescheduler/filter/configfilter.go) | Go | 177 | 0 | 24 | 201 |
| [pkg/kubescheduler/filter/configfilter_test.go](/pkg/kubescheduler/filter/configfilter_test.go) | Go | 80 | 0 | 26 | 106 |
| [pkg/kubescheduler/filter/templatefilter.go](/pkg/kubescheduler/filter/templatefilter.go) | Go | 9 | 0 | 3 | 12 |
| [pkg/kubescheduler/policy/lrscheduler.go](/pkg/kubescheduler/policy/lrscheduler.go) | Go | 54 | 0 | 9 | 63 |
| [pkg/kubescheduler/policy/lrscheduler_test.go](/pkg/kubescheduler/policy/lrscheduler_test.go) | Go | 41 | 0 | 9 | 50 |
| [pkg/kubescheduler/policy/resourcescheduler.go](/pkg/kubescheduler/policy/resourcescheduler.go) | Go | 61 | 0 | 12 | 73 |
| [pkg/kubescheduler/policy/resourcescheduler_test.go](/pkg/kubescheduler/policy/resourcescheduler_test.go) | Go | 46 | 0 | 13 | 59 |
| [pkg/kubescheduler/policy/templatescheduler.go](/pkg/kubescheduler/policy/templatescheduler.go) | Go | 16 | 0 | 6 | 22 |
| [pkg/kubescheduler/run.go](/pkg/kubescheduler/run.go) | Go | 111 | 0 | 20 | 131 |
| [pkg/kubescheduler/testutils/builder.go](/pkg/kubescheduler/testutils/builder.go) | Go | 107 | 0 | 5 | 112 |
| [pkg/serverless/activator/deploy.go](/pkg/serverless/activator/deploy.go) | Go | 314 | 0 | 37 | 351 |
| [pkg/serverless/activator/deploy_test.go](/pkg/serverless/activator/deploy_test.go) | Go | 26 | 0 | 9 | 35 |
| [pkg/serverless/autoscaler/metric.go](/pkg/serverless/autoscaler/metric.go) | Go | 101 | 0 | 14 | 115 |
| [pkg/serverless/autoscaler/record.go](/pkg/serverless/autoscaler/record.go) | Go | 33 | 0 | 14 | 47 |
| [pkg/serverless/eventfilter/functionwatcher.go](/pkg/serverless/eventfilter/functionwatcher.go) | Go | 140 | 0 | 19 | 159 |
| [pkg/serverless/eventfilter/workflowwatcher.go](/pkg/serverless/eventfilter/workflowwatcher.go) | Go | 65 | 0 | 9 | 74 |
| [pkg/serverless/function/image.go](/pkg/serverless/function/image.go) | Go | 117 | 0 | 23 | 140 |
| [pkg/serverless/function/image_test.go](/pkg/serverless/function/image_test.go) | Go | 29 | 0 | 8 | 37 |
| [pkg/serverless/imagedata/requirements.txt](/pkg/serverless/imagedata/requirements.txt) | pip requirements | 1 | 0 | 1 | 2 |
| [pkg/serverless/run.go](/pkg/serverless/run.go) | Go | 10 | 0 | 3 | 13 |
| [pkg/serverless/workflow/workflowexecutor.go](/pkg/serverless/workflow/workflowexecutor.go) | Go | 229 | 0 | 38 | 267 |
| [pkg/serverless/workflow/workflowexecutor_test.go](/pkg/serverless/workflow/workflowexecutor_test.go) | Go | 259 | 0 | 27 | 286 |
| [utils/client.go](/utils/client.go) | Go | 140 | 0 | 19 | 159 |
| [utils/http.go](/utils/http.go) | Go | 53 | 0 | 11 | 64 |
| [utils/rand.go](/utils/rand.go) | Go | 88 | 0 | 13 | 101 |
| [utils/resourceutils/unit.go](/utils/resourceutils/unit.go) | Go | 66 | 0 | 15 | 81 |
| [utils/resourceutils/unit_test.go](/utils/resourceutils/unit_test.go) | Go | 26 | 0 | 6 | 32 |
| [utils/utils.go](/utils/utils.go) | Go | 18 | 0 | 9 | 27 |

[Summary](results.md) / Details / [Diff Summary](diff.md) / [Diff Details](diff-details.md)