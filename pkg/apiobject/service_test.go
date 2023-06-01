package apiobject

import (
	"encoding/json"
	"testing"
)

func TestService(t *testing.T) {
	service := &Service{
		APIVersion: "app/v1",
		Data: MetaData{
			Name: "dns-service",
		},
		Spec: ServiceSpec{
			Type: "ClusterIP",
			Ports: []ServicePort{
				{
					Name:       "service-port1",
					Protocol:   "TCP",
					Port:       6800,
					TargetPort: "p1",
				},
				{
					Name:       "service-port2",
					Protocol:   "TCP",
					Port:       6880,
					TargetPort: "p2",
				},
				{
					Name:       "service-port3",
					Protocol:   "TCP",
					Port:       80,
					TargetPort: "p3",
				},
			},
			Selector: map[string]string{
				"app": "dns-test",
			},
		},
	}

	serviceJson, err := json.MarshalIndent(service, "", "  ")
	if err != nil {
		t.Error(err)
	}
	t.Log(string(serviceJson))
}