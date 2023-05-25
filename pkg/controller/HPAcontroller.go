package controller

import (
	"fmt"
	"github.com/robfig/cron"
	log "github.com/sirupsen/logrus"
	"github.com/tidwall/gjson"
	"math"
	"minik8s/config"
	"minik8s/pkg/apiobject"
	apiu "minik8s/pkg/apiobject/utils"
	"minik8s/utils"
	"time"
)

/* 主要工作：
1. 监听autoscaler的创建。更改对应replicaset/创建对应replicaset。
2. 监听autoscaler的更改。如果CurrentReplicas和DesiredReplicas数量不一致，则更新对应rs。
3. 监听autoscaler的删除。将对应replicaset的control状态还原。
4. 每隔15s检查hpa的条件是否满足，进行扩缩容。扩缩容逻辑如下：
	1）ScaleTargetRef字段找到对应的pod。Metrics字段提供多个指定的指标，根据每个指标利用metric api向kubelet查询对应pod的资源利用率并计算当前指标值。
	2）将计算出的结果与指标比较。计算出扩缩容的期望副本数。公式： 期望副本数 = ceil[当前副本数 * (当前指标 / 期望指标)]
	3）每个指标都会计算出一个期望副本数。取最大值作为总的期望副本数。
	4）根据Behavior字段定义的扩缩容行为判断总的期望副本数是否满足条件，并确定最终的期望副本数。需要满足的条件有：
		a. 不超过MaxReplicas，不小于MinReplicas。
		b. 上一次扩缩容距今时间大于StabilizationWindowSeconds（扩容默认为0，缩容默认为300s）
		c. 满足HPAScalingPolicy。（如每3秒最多新增10个pod，每20s最多减少10%的pod）。不同policy的限制之间可以设定取最小/最大限制。
	5）根据上述三个条件的限制确定最终副本数，并更新hpa的DesiredReplicas。
*/

type hpaScalerHandler struct {
}

/* ========== Start HPA Handler ========== */

func (s hpaScalerHandler) HandleCreate(message []byte) {
	hpa := &apiobject.HorizontalPodAutoscaler{}
	hpa.UnMarshalJSON(message)

	// find replicaset to control
	switch hpa.Spec.ScaleTargetRef.Kind {
	case config.REPLICA:
		{
			info := utils.GetObject(hpa.Spec.ScaleTargetRef.Kind, hpa.Data.Namespace, hpa.Spec.ScaleTargetRef.Name)
			rs := &apiobject.ReplicationController{}
			rs.UnMarshalJSON([]byte(info))

			rs.Status.OwnerReference = apiobject.OwnerReference{
				Kind:       config.HPA,
				Name:       hpa.Data.Name,
				Controller: true,
			}
			rs.Status.Scale = hpa.Spec.MinReplicas

			utils.UpdateObject(rs, config.REPLICA, rs.Data.Namespace, rs.Data.Name)
		}
	case config.POD:
		{
			//TODO: the pod situation
		}
	}

	log.Info("[hpa controller] Create HPA. Name:", hpa.Data.Name)
}

func (s hpaScalerHandler) HandleDelete(message []byte) {
	hpa := &apiobject.HorizontalPodAutoscaler{}
	hpa.UnMarshalJSON(message)

	switch hpa.Spec.ScaleTargetRef.Kind {
	case config.REPLICA:
		{
			info := utils.GetObject(hpa.Spec.ScaleTargetRef.Kind, hpa.Data.Namespace, hpa.Spec.ScaleTargetRef.Name)
			rs := &apiobject.ReplicationController{}
			rs.UnMarshalJSON([]byte(info))

			rs.Status.OwnerReference.Controller = false
			rs.Status.Scale = 0

			utils.UpdateObject(rs, config.REPLICA, rs.Data.Namespace, rs.Data.Name)
		}
	case config.POD:
		{
			//TODO: the pod situation
		}
	}

	log.Info("[rs controller] Delete HPA. Name:", hpa.Data.Name)
}

func (s hpaScalerHandler) HandleUpdate(message []byte) {
	hpa := &apiobject.HorizontalPodAutoscaler{}
	hpa.UnMarshalJSON(message)

	// change the controlled resource according to hpa
	if hpa.Status.CurrentReplicas != hpa.Status.DesiredReplicas {
		switch hpa.Spec.ScaleTargetRef.Kind {
		case config.REPLICA:
			{
				info := utils.GetObject(hpa.Spec.ScaleTargetRef.Kind, hpa.Data.Namespace, hpa.Spec.ScaleTargetRef.Name)
				rs := &apiobject.ReplicationController{}
				rs.UnMarshalJSON([]byte(info))
				// update rs
				rs.Status.Scale = hpa.Status.DesiredReplicas
				utils.UpdateObject(rs, config.REPLICA, rs.Data.Namespace, rs.Data.Name)
				// update hpa
				hpa.Status.CurrentReplicas = hpa.Status.DesiredReplicas
				utils.UpdateObject(hpa, config.HPA, hpa.Data.Namespace, hpa.Data.Name)
			}
		case config.POD:
			{
				//TODO: the pod situation
			}
		}
	}
	/* TODO: scaleTargetRef change */
}

func (s hpaScalerHandler) GetType() config.ObjType {
	return config.HPA
}

/* ========== Start Check Function ========== */

func regularCheck() {
	//创建一个定时任务对象
	c := cron.New()
	//给对象增加定时任务
	c.AddFunc("@every 15s", func() {
		// get all hpa
		info := utils.GetObject(config.HPA, "nil", "")
		hpaList := gjson.Parse(info).Array()
		for _, h := range hpaList {
			hpa := &apiobject.HorizontalPodAutoscaler{}
			hpa.UnMarshalJSON([]byte(h.String()))
			reconcileAutoscaler(hpa)
		}
	})
	//启动定时任务
	c.Start()
}

func reconcileAutoscaler(hpa *apiobject.HorizontalPodAutoscaler) {
	// 1. get all the pods belong to hpa
	var podList []*apiobject.Pod
	switch hpa.Spec.ScaleTargetRef.Kind {
	case config.REPLICA:
		{
			info := utils.GetObject(hpa.Spec.ScaleTargetRef.Kind, hpa.Data.Namespace, hpa.Spec.ScaleTargetRef.Name)
			rs := &apiobject.ReplicationController{}
			rs.UnMarshalJSON([]byte(info))
			podList = GetPodListFromRS(rs)
		}
	}

	// 2. get metric of all the pods
	var metricList []*apiobject.PodMetrics
	for _, pod := range podList {
		metric := getMetricsFromKubelet(pod.Status.HostIp, pod.Data.Namespace, pod.Data.Name)
		metricList = append(metricList, metric)
	}

	// 3. calculate expected replica for each metric requirement of hpa
	var max int32
	max = 0
	for i, m := range hpa.Spec.Metrics {
		expect := computeReplicasForMetric(hpa, &m, i, metricList)
		if expect > 0 {
			max = expect
		}
	}
	desired := max

	// 4. update the autoscaler

	if desired > hpa.Spec.MaxReplicas {
		desired = hpa.Spec.MaxReplicas
	}
	if desired < hpa.Spec.MinReplicas {
		desired = hpa.Spec.MinReplicas
	}

	var duration = time.Since(hpa.Status.LastScaleTime.Time).Seconds()
	var window int32

	if desired < hpa.Status.CurrentReplicas {
		// scale down
		window = int32(300)
		if hpa.Spec.Behavior.ScaleDown != nil {
			res := checkScalingRules(&window, hpa.Spec.Behavior.ScaleDown, hpa, duration)
			if hpa.Status.CurrentReplicas-res > desired {
				desired = hpa.Status.CurrentReplicas - res
			}
		}

	} else if desired < hpa.Status.CurrentReplicas {
		// scale up
		window = int32(0)
		if hpa.Spec.Behavior.ScaleUp != nil {
			res := checkScalingRules(&window, hpa.Spec.Behavior.ScaleUp, hpa, duration)
			if hpa.Status.CurrentReplicas+res < desired {
				desired = hpa.Status.CurrentReplicas + res
			}
		}
	} else {
		return
	}
	if duration > float64(window) {
		hpa.Status.LastScaleTime = apiu.Now()
		hpa.Status.DesiredReplicas = desired
		utils.UpdateObject(hpa, config.HPA, hpa.Data.Namespace, hpa.Data.Name)
	}
}

func computeReplicasForMetric(hpa *apiobject.HorizontalPodAutoscaler, metric *apiobject.MetricSpec, index int, metricList []*apiobject.PodMetrics) int32 {
	switch metric.Type {
	case apiobject.PodsMetricSourceType:
		{

		}
	case apiobject.ResourceMetricSourceType:
		{
			return computeReplicasForResourceMetric(hpa, metric, index, metricList)
		}
	}
	return 0
}

func computeReplicasForResourceMetric(hpa *apiobject.HorizontalPodAutoscaler, required *apiobject.MetricSpec, index int, metricList []*apiobject.PodMetrics) int32 {
	if required.Resource.Target.Type != apiobject.AverageValueMetricType {
		log.Error("[HPA controller] Not implemented Metric Type for Resource MetricSourceType")
	}
	total := 0
	for _, m := range metricList {
		sum := 0
		for _, s := range m.Containers {
			sum += int(s.Usage[required.Resource.Name])
		}
		total += sum
	}
	avg := float64(total) / float64(len(metricList))
	hpa.SetStatusValue(&hpa.Status.CurrentMetrics[index], avg)
	scale := avg / float64(*required.Resource.Target.AverageValue)
	expect := int32(math.Ceil(scale * float64(hpa.Status.CurrentReplicas)))
	return expect
}

func getMetricsFromKubelet(IP string, ns string, name string) *apiobject.PodMetrics {
	url := fmt.Sprintf("http://%s:10250/%s/%s", IP, ns, name)
	var str []byte
	if info, err := utils.SendRequest("GET", str, url); err != nil {
		log.Error("get object ", info)
		return nil
	} else {
		m := &apiobject.PodMetrics{}
		m.UnMarshalJSON([]byte(info))
		return m
	}
}

func getLimitFromScalingPolicy(p apiobject.HPAScalingPolicy, currentReplica int32, duration float64, selectPolicy apiobject.ScalingPolicySelect, currentRes int32) int32 {
	var res int32
	var limit float64
	switch p.Type {
	case apiobject.PodsScalingPolicy:
		{
			limit = float64(p.Value) / float64(p.PeriodSeconds) * duration
		}
	case apiobject.PercentScalingPolicy:
		{
			limit = float64(p.Value) / 100.0 * float64(currentReplica) / float64(p.PeriodSeconds) * duration
		}
	}
	switch selectPolicy {
	case apiobject.MinPolicySelect:
		{
			if int32(limit) < currentRes {
				res = int32(limit)
			}
		}
	case apiobject.MaxPolicySelect:
		{
			if int32(limit) > currentRes {
				res = int32(limit)
			}
		}
	}
	return res
}

func checkScalingRules(window *int32, rules *apiobject.HPAScalingRules, hpa *apiobject.HorizontalPodAutoscaler, duration float64) int32 {
	// check the window
	if rules.StabilizationWindowSeconds != nil {
		*window = *hpa.Spec.Behavior.ScaleDown.StabilizationWindowSeconds
	}

	// check the policy
	var selectPolicy apiobject.ScalingPolicySelect
	if hpa.Spec.Behavior.ScaleDown.SelectPolicy == nil {
		selectPolicy = apiobject.MaxPolicySelect
	}

	var res int32
	switch selectPolicy {
	case apiobject.MinPolicySelect:
		{
			res = 1<<31 - 1
		}
	case apiobject.MaxPolicySelect:
		{
			res = 0
		}
	}

	for _, p := range hpa.Spec.Behavior.ScaleDown.Policies {
		res = getLimitFromScalingPolicy(p, hpa.Status.CurrentReplicas, duration, selectPolicy, res)
	}
	return res
}
