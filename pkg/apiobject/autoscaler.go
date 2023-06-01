package apiobject

import (
	"encoding/json"
	"fmt"
	"minik8s/config"
	"minik8s/pkg/apiobject/utils"
	"strconv"
)

/* an basic example of a autoscaler apiobject:
apiVersion: autoscaling/v2beta2
kind: HorizontalPodAutoscaler
metadata:
  name: hpa-practice
spec:
  minReplicas: 3  # 最小pod数量
  maxReplicas: 6  # 最大pod数量
  metrics:
  - pods:
      metric:
        name: k8s_pod_rate_cpu_core_used_limit
      target:
        averageValue: "80"
        type: AverageValue
    type: Pods
  scaleTargetRef:   # 指定要控制的deploy
    apiVersion:  apps/v1
    kind: Pod
    name: deploy-practice
  behavior:
    scaleDown:
      policies:
      - type: Percent
        value: 10
        periodSeconds: 60 # 每分钟最多10%

*/

// ref:staging\src\k8s.io\api\autoscaling\v2beta2\types.go

type HorizontalPodAutoscaler struct {
	APIVersion string                        `json:"apiVersion,omitempty"`
	Data       MetaData                      `json:"metadata"`
	Spec       HorizontalPodAutoscalerSpec   `json:"spec,omitempty"`
	Status     HorizontalPodAutoscalerStatus `json:"status,omitempty"`
}

type HorizontalPodAutoscalerSpec struct {
	ScaleTargetRef CrossVersionObjectReference      `json:"scaleTargetRef"`
	MinReplicas    int32                            `json:"minReplicas,omitempty"`
	MaxReplicas    int32                            `json:"maxReplicas"`
	Metrics        []MetricSpec                     `json:"metrics,omitempty"`
	Behavior       *HorizontalPodAutoscalerBehavior `json:"behavior,omitempty"`
}

type CrossVersionObjectReference struct {
	Kind       config.ObjType `json:"kind"`
	Name       string         `json:"name"`
	APIVersion string         `json:"apiVersion,omitempty"`
}

type MetricSpec struct {
	Type     MetricSourceType      `json:"type"`
	Pods     *PodsMetricSource     `json:"pods,omitempty"`
	Resource *ResourceMetricSource `json:"resource,omitempty"`
}

type PodsMetricSource struct {
	//// metric identifies the target metric by name and selector
	//Metric MetricIdentifier `json:"metric" protobuf:"bytes,1,name=metric"`
	// target specifies the target value for the given metric
	Target MetricTarget `json:"target" protobuf:"bytes,2,name=target"`
}

type ResourceMetricSource struct {
	// name is the name of the resource in question.
	Name ResourceName `json:"name"`
	// target specifies the target value for the given metric
	Target MetricTarget `json:"target"`
}

type MetricTarget struct {
	// type represents whether the metric type is Utilization, Value, or AverageValue
	Type MetricTargetType `json:"type"`
	// value is the target value of the metric (as a quantity).
	// +optional
	Value *utils.Quantity `json:"value,omitempty"`
	// averageValue is the target value of the average of the
	// metric across all relevant pods (as a quantity)
	// +optional
	AverageValue *utils.Quantity `json:"averageValue,omitempty"`
	// averageUtilization is the target value of the average of the
	// resource metric across all relevant pods, represented as a percentage of
	// the requested value of the resource for the pods.
	// Currently only valid for Resource metric source type
	// +optional
	AverageUtilization *int32 `json:"averageUtilization,omitempty"`
}

type MetricTargetType string

const (
	// UtilizationMetricType declares a MetricTarget is an AverageUtilization value
	UtilizationMetricType MetricTargetType = "Utilization"
	// ValueMetricType declares a MetricTarget is a raw value
	ValueMetricType MetricTargetType = "Value"
	// AverageValueMetricType declares a MetricTarget is an
	AverageValueMetricType MetricTargetType = "AverageValue"
)

type ResourceName string

const (
	// CPU, in cores. (500m = .5 cores)
	ResourceCPU ResourceName = "cpu"
	// Memory, in bytes. (500Gi = 500GiB = 500 * 1024 * 1024 * 1024)
	ResourceMemory ResourceName = "memory"
	// Volume size, in bytes (e,g. 5Gi = 5GiB = 5 * 1024 * 1024 * 1024)
	ResourceStorage ResourceName = "storage"
	// Local ephemeral storage, in bytes. (500Gi = 500GiB = 500 * 1024 * 1024 * 1024)
	// The resource name for ResourceEphemeralStorage is alpha and it can change across releases.
	ResourceEphemeralStorage ResourceName = "ephemeral-storage"
)

type HorizontalPodAutoscalerBehavior struct {
	// scaleUp is scaling policy for scaling Up.
	// If not set, the default value is the higher of:
	//   * increase no more than 4 pods per 60 seconds
	//   * double the number of pods per 60 seconds
	// No stabilization is used.
	// +optional
	ScaleUp *HPAScalingRules `json:"scaleUp,omitempty"`
	// scaleDown is scaling policy for scaling Down.
	// If not set, the default value is to allow to scale down to minReplicas pods, with a
	// 300 second stabilization window (i.e., the highest recommendation for
	// the last 300sec is used).
	// +optional
	ScaleDown *HPAScalingRules `json:"scaleDown,omitempty"`
}

type HPAScalingRules struct {
	// StabilizationWindowSeconds is the number of seconds for which past recommendations should be
	// considered while scaling up or scaling down.
	// StabilizationWindowSeconds must be greater than or equal to zero and less than or equal to 3600 (one hour).
	// If not set, use the default values:
	// - For scale up: 0 (i.e. no stabilization is done).
	// - For scale down: 300 (i.e. the stabilization window is 300 seconds long).
	// +optional
	StabilizationWindowSeconds *int32 `json:"stabilizationWindowSeconds,omitempty"`
	// selectPolicy is used to specify which policy should be used.
	// If not set, the default value MaxPolicySelect is used.
	// +optional
	SelectPolicy *ScalingPolicySelect `json:"selectPolicy,omitempty"`
	// policies is a list of potential scaling polices which can be used during scaling.
	// At least one policy must be specified, otherwise the HPAScalingRules will be discarded as invalid
	// +optional
	Policies []HPAScalingPolicy `json:"policies,omitempty"`
}

type HPAScalingPolicyType string

const (
	// PodsScalingPolicy is a policy used to specify a change in absolute number of pods.
	PodsScalingPolicy HPAScalingPolicyType = "Pods"
	// PercentScalingPolicy is a policy used to specify a relative amount of change with respect to
	// the current number of pods.
	PercentScalingPolicy HPAScalingPolicyType = "Percent"
)

type ScalingPolicySelect string

const (
	// MaxPolicySelect selects the policy with the highest possible change.
	MaxPolicySelect ScalingPolicySelect = "Max"
	// MinPolicySelect selects the policy with the lowest possible change.
	MinPolicySelect ScalingPolicySelect = "Min"
	// DisabledPolicySelect disables the scaling in this direction.
	DisabledPolicySelect ScalingPolicySelect = "Disabled"
)

// HPAScalingPolicy is a single policy which must hold true for a specified past interval.
type HPAScalingPolicy struct {
	// Type is used to specify the scaling policy.
	Type HPAScalingPolicyType `json:"type"`
	// Value contains the amount of change which is permitted by the policy.
	// It must be greater than zero
	Value int32 `json:"value"`
	// PeriodSeconds specifies the window of time for which the policy should hold true.
	// PeriodSeconds must be greater than zero and less than or equal to 1800 (30 min).
	PeriodSeconds int32 `json:"periodSeconds"`
}

type MetricSourceType string

const (
	// ObjectMetricSourceType is a metric describing a kubernetes object
	// (for example, hits-per-second on an Ingress object).
	ObjectMetricSourceType MetricSourceType = "Object"
	// PodsMetricSourceType is a metric describing each pod in the current scale
	// target (for example, transactions-processed-per-second).  The values
	// will be averaged together before being compared to the target value.
	PodsMetricSourceType MetricSourceType = "Pods"
	// ResourceMetricSourceType is a resource metric known to Kubernetes, as
	// specified in requests and limits, describing each pod in the current
	// scale target (e.g. CPU or memory).  Such metrics are built in to
	// Kubernetes, and have special scaling options on top of those available
	// to normal per-pod metrics (the "pods" source).
	ResourceMetricSourceType MetricSourceType = "Resource"
	// ContainerResourceMetricSourceType is a resource metric known to Kubernetes, as
	// specified in requests and limits, describing a single container in each pod in the current
	// scale target (e.g. CPU or memory).  Such metrics are built in to
	// Kubernetes, and have special scaling options on top of those available
	// to normal per-pod metrics (the "pods" source).
	ContainerResourceMetricSourceType MetricSourceType = "ContainerResource"
	// ExternalMetricSourceType is a global metric that is not associated
	// with any Kubernetes object. It allows autoscaling based on information
	// coming from components running outside of cluster
	// (for example length of queue in cloud messaging service, or
	// QPS from loadbalancer running outside of cluster).
	ExternalMetricSourceType MetricSourceType = "External"
)

type HorizontalPodAutoscalerStatus struct {
	// observedGeneration is the most recent generation observed by this autoscaler.
	// +optional
	ObservedGeneration *int64 `json:"observedGeneration,omitempty"`

	// lastScaleTime is the last time the HorizontalPodAutoscaler scaled the number of pods,
	// used by the autoscaler to control how often the number of pods is changed.
	// +optional
	LastScaleTime utils.Time `json:"lastScaleTime,omitempty"`

	// currentReplicas is current number of replicas of pods managed by this autoscaler,
	// as last seen by the autoscaler.
	CurrentReplicas int32 `json:"currentReplicas"`

	// desiredReplicas is the desired number of replicas of pods managed by this autoscaler,
	// as last calculated by the autoscaler.
	DesiredReplicas int32 `json:"desiredReplicas"`

	// currentMetrics is the last read state of the metrics used by this autoscaler.
	// +optional
	CurrentMetrics []MetricValueStatus `json:"currentMetrics"`
	//
	//// conditions is the set of conditions required for this autoscaler to scale its target,
	//// and indicates whether or not those conditions are met.
	//// +optional
	//Conditions []HorizontalPodAutoscalerCondition `json:"conditions"`
}

type MetricValueStatus struct {
	// type represents whether the metric type is Utilization, Value, or AverageValue
	Type MetricTargetType `json:"type"`
	// value is the current value of the metric (as a quantity).
	// +optional
	Value *utils.Quantity `json:"value,omitempty"`
	// averageValue is the current value of the average of the
	// metric across all relevant pods (as a quantity)
	// +optional
	AverageValue *utils.Quantity `json:"averageValue,omitempty"`
	// currentAverageUtilization is the current value of the average of the
	// resource metric across all relevant pods, represented as a percentage of
	// the requested value of the resource for the pods.
	// +optional
	AverageUtilization *int32 `json:"averageUtilization,omitempty"`
}

func (h *HorizontalPodAutoscaler) MarshalJSON() ([]byte, error) {
	type Alias HorizontalPodAutoscaler
	return json.Marshal(&struct {
		*Alias
	}{
		Alias: (*Alias)(h),
	})
}

func (h *HorizontalPodAutoscaler) UnMarshalJSON(data []byte) error {
	type Alias HorizontalPodAutoscaler
	aux := &struct {
		*Alias
	}{
		Alias: (*Alias)(h),
	}
	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}
	return nil
}

func (h *HorizontalPodAutoscaler) GetTargetValue(m *MetricSpec) string {
	switch m.Type {
	case ResourceMetricSourceType:
		{
			switch m.Resource.Target.Type {
			case AverageValueMetricType:
				{
					return strconv.Itoa(int(*m.Resource.Target.AverageValue))
				}
			case ValueMetricType:
				{
					return strconv.Itoa(int(*m.Resource.Target.Value))
				}
			case UtilizationMetricType:
				{
					return strconv.Itoa(int(*m.Resource.Target.AverageUtilization)) + "%"
				}

			}
		}
	case PodsMetricSourceType:
		{
			switch m.Pods.Target.Type {
			case AverageValueMetricType:
				{
					return strconv.Itoa(int(*m.Pods.Target.AverageValue))
				}
			case ValueMetricType:
				{
					return strconv.Itoa(int(*m.Pods.Target.Value))
				}
			case UtilizationMetricType:
				{
					return strconv.Itoa(int(*m.Pods.Target.AverageUtilization)) + "%"
				}

			}
		}
	}
	return "0"
}

func (h *HorizontalPodAutoscaler) GetStatusValue(m *MetricValueStatus) string {
	switch m.Type {
	case AverageValueMetricType:
		{
			return strconv.Itoa(int(*m.AverageValue))
		}
	case ValueMetricType:
		{
			return strconv.Itoa(int(*m.Value))
		}
	case UtilizationMetricType:
		{
			return strconv.Itoa(int(*m.AverageUtilization)) + "%"
		}

	}
	return "0"
}

func (h *HorizontalPodAutoscaler) SetStatusValue(m *MetricValueStatus, v float64) {
	switch m.Type {
	case AverageValueMetricType:
		{
			if m.AverageValue == nil {
				var v utils.Quantity
				m.AverageValue = &v
				fmt.Print("here")
			}
			*m.AverageValue = utils.Quantity(v)
		}
	case ValueMetricType:
		{
			if m.Value == nil {
				var v utils.Quantity
				m.Value = &v
			}
			*m.Value = utils.Quantity(v)
		}
	case UtilizationMetricType:
		{
			if m.AverageUtilization == nil {
				var v int32
				m.AverageUtilization = &v
			}
			*m.AverageUtilization = int32(v)
		}
	}
}

func (hpa *HorizontalPodAutoscaler) Union(other *HorizontalPodAutoscaler) {
	if hpa.Status.ObservedGeneration == nil {
		hpa.Status.ObservedGeneration = other.Status.ObservedGeneration
	}
	if hpa.Status.LastScaleTime.IsZero() {
		hpa.Status.LastScaleTime = other.Status.LastScaleTime
	}
	if hpa.Status.CurrentReplicas == 0 {
		hpa.Status.CurrentReplicas = other.Status.CurrentReplicas
	}
	if hpa.Status.DesiredReplicas == 0 {
		hpa.Status.DesiredReplicas = other.Status.DesiredReplicas
	}
	if hpa.Status.CurrentMetrics == nil {
		hpa.Status.CurrentMetrics = other.Status.CurrentMetrics
	}
}
