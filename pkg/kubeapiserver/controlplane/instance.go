package controlplane

import "time"

// Instance contains state for a Kubernetes cluster handlers server instance.
type Instance struct {
	// Name is the name of the instance.
	Name string
	// Namespace is the namespace of the instance.
	Namespace string
	// Cluster is the name of the cluster the instance belongs to.
	Cluster string
	// Version is the version of the instance.
	Version string
	// Image is the image of the instance.
	Image string
	// Replicas is the number of replicas of the instance.
	Replicas int
	// Status is the status of the instance.
	Status string
	// CreatedAt is the creation time of the instance.
	CreatedAt time.Time
	// UpdatedAt is the last update time of the instance.
	UpdatedAt time.Time
}

// InstanceList contains a list of Instance.
