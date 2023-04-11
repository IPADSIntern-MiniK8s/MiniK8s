package container

import (
	"github.com/containerd/containerd/oci"
	"github.com/opencontainers/runtime-spec/specs-go"
)

func GenerateMountSpec(mounts map[string]string) oci.SpecOpts {
	sMounts := make([]specs.Mount, len(mounts), len(mounts))
	i := 0
	for source, destination := range mounts {
		sMounts[i] = specs.Mount{
			Destination: destination,
			Source:      source,
			Type:        "bind",           // if tmpfs, can not persist
			Options:     []string{"bind"}, //otherwise no such device error
		}
		i++
	}
	return oci.WithMounts(sMounts)
}

func GenerateHostnameSpec(hostname string) oci.SpecOpts {
	return oci.WithHostname(hostname)
}

//TODO command

//TODO cpu

//TODO memory

//TODO port
