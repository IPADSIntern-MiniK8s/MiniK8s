package apiobject

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
    kind: Deployment
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
	Kind       string `json:"kind"`
	Name       string `json:"name"`
	APIVersion string `json:"apiVersion,omitempty"`
}

type MetricSpec struct {
	Type     MetricSourceType      `json:"type"`
	Pods     *PodsMetricSource     `json:"pods,omitempty"`
	Resource *ResourceMetricSource `json:"resource,omitempty"`
}

type PodsMetricSource struct {
	// metric identifies the target metric by name and selector
	Metric MetricIdentifier `json:"metric" protobuf:"bytes,1,name=metric"`
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
	Value *Quantity `json:"value,omitempty" protobuf:"bytes,2,opt,name=value"`
	// averageValue is the target value of the average of the
	// metric across all relevant pods (as a quantity)
	// +optional
	AverageValue *Quantity `json:"averageValue,omitempty"`
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

// Quantity is a fixed-point representation of a number.
// It provides convenient marshaling/unmarshaling in JSON and YAML,
// in addition to String() and AsInt64() accessors.
//
// The serialization format is:
//
// ```
// <quantity>        ::= <signedNumber><suffix>
//
//	(Note that <suffix> may be empty, from the "" case in <decimalSI>.)
//
// <digit>           ::= 0 | 1 | ... | 9
// <digits>          ::= <digit> | <digit><digits>
// <number>          ::= <digits> | <digits>.<digits> | <digits>. | .<digits>
// <sign>            ::= "+" | "-"
// <signedNumber>    ::= <number> | <sign><number>
// <suffix>          ::= <binarySI> | <decimalExponent> | <decimalSI>
// <binarySI>        ::= Ki | Mi | Gi | Ti | Pi | Ei
//
//	(International System of units; See: http://physics.nist.gov/cuu/Units/binary.html)
//
// <decimalSI>       ::= m | "" | k | M | G | T | P | E
//
//	(Note that 1024 = 1Ki but 1000 = 1k; I didn't choose the capitalization.)
//
// <decimalExponent> ::= "e" <signedNumber> | "E" <signedNumber>
// ```
//
// No matter which of the three exponent forms is used, no quantity may represent
// a number greater than 2^63-1 in magnitude, nor may it have more than 3 decimal
// places. Numbers larger or more precise will be capped or rounded up.
// (E.g.: 0.1m will rounded up to 1m.)
// This may be extended in the future if we require larger or smaller quantities.
//
// When a Quantity is parsed from a string, it will remember the type of suffix
// it had, and will use the same type again when it is serialized.
//
// Before serializing, Quantity will be put in "canonical form".
// This means that Exponent/suffix will be adjusted up or down (with a
// corresponding increase or decrease in Mantissa) such that:
//
// - No precision is lost
// - No fractional digits will be emitted
// - The exponent (or suffix) is as large as possible.
//
// The sign will be omitted unless the number is negative.
//
// Examples:
//
// - 1.5 will be serialized as "1500m"
// - 1.5Gi will be serialized as "1536Mi"
//
// Note that the quantity will NEVER be internally represented by a
// floating point number. That is the whole point of this exercise.
//
// Non-canonical values will still parse as long as they are well formed,
// but will be re-emitted in their canonical form. (So always use canonical
// form, or don't diff.)
//
// This format is intended to make it difficult to use these numbers without
// writing some sort of special handling code in the hopes that that will
// cause implementors to also use a fixed point implementation.
type Quantity struct {
	// i is the quantity in int64 scaled form, if d.Dec == nil
	i int64Amount
	// d is the quantity in inf.Dec form if d.Dec != nil
	d infDecAmount
	// s is the generated value of this quantity to avoid recalculation
	s string

	// Change Format at will. See the comment for Canonicalize for
	// more details.
	Format
}

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
	SelectPolicy *ScalingPolicySelect `json:"selectPolicy,omitempty" protobuf:"bytes,1,opt,name=selectPolicy"`
	// policies is a list of potential scaling polices which can be used during scaling.
	// At least one policy must be specified, otherwise the HPAScalingRules will be discarded as invalid
	// +optional
	Policies []HPAScalingPolicy `json:"policies,omitempty" protobuf:"bytes,2,rep,name=policies"`
}

type HPAScalingPolicyType string

const (
	// PodsScalingPolicy is a policy used to specify a change in absolute number of pods.
	PodsScalingPolicy HPAScalingPolicyType = "Pods"
	// PercentScalingPolicy is a policy used to specify a relative amount of change with respect to
	// the current number of pods.
	PercentScalingPolicy HPAScalingPolicyType = "Percent"
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
