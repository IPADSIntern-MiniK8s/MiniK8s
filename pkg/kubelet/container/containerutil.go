package container

import (
	"github.com/containerd/containerd/oci"
	"github.com/opencontainers/runtime-spec/specs-go"
	"strconv"
	"strings"
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

func GenerateCMDSpec(CmdLine string) oci.SpecOpts {
	//TODO split by space;
	return oci.WithProcessArgs(CmdLine)
}

type CPUSpecType int

const (
	CPUNone   CPUSpecType = iota
	CPUNumber             //float ,not bind to certain cpu; eg: 1 ,0.5
	CPUCoreID             // bind certain cpus ,start from 0; eg: 0,1, 0-2
	CPUShares             // priority
)

type CPUSpec struct {
	Type  CPUSpecType
	Value string
}

func GenerateCPUSpec(spec CPUSpec) oci.SpecOpts {
	switch spec.Type {
	case CPUNumber:
		cpus, _ := strconv.ParseFloat(spec.Value, 64)
		var (
			period = uint64(100000)
			quota  = int64(cpus * 100000.0)
		)
		return oci.WithCPUCFS(quota, period)
	case CPUCoreID:
		return oci.WithCPUs(spec.Value)
	case CPUShares:
		shares, _ := strconv.ParseUint(spec.Value, 10, 64)
		return oci.WithCPUShares(shares)
	}
	return nil
}

// bytes, if exceed ,the container will be stopped at once
func GenerateMemorySpec(limit uint64) oci.SpecOpts {
	return oci.WithMemoryLimit(limit)
}

func GenerateEnvSpec(envs []string) oci.SpecOpts {
	return oci.WithEnv(envs)
}

func GenerateNamespaceSpec(nsType,path string) oci.SpecOpts{
	return oci.WithLinuxNamespace(specs.LinuxNamespace{Type:specs.LinuxNamespaceType(nsType),Path: path})
}

func PadImageName(image string) string {
	res := image
	if strings.Index(image, ":") == -1 {
		res += ":latest"
	}
	if strings.Index(image, "/") == -1 {
		res = "docker.io/library/" + res
	}
	return res
}
