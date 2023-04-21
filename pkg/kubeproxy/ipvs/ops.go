package ipvs

import (
	"fmt"
	"github.com/mqliang/libipvs"
	"net"
	"os/exec"
	"strconv"
	"syscall"
)

var handler libipvs.IPVSHandle

func Init() {
	h, err := libipvs.New()
	handler = h
	if err != nil {
		fmt.Println(err.Error())
	}

	_, err = exec.Command("sysctl", []string{"net.ipv4.vs.conntrack=1"}...).CombinedOutput()
	if err != nil {
		fmt.Println(err.Error())
	}

}

func TestConfig() {
	svc := addService("10.9.0.1", 12)
	bindEndpoint(svc, "10.2.17.54", 12345)
}

func addService(ip string, port uint16) *libipvs.Service {
	// Create a service struct and add it to the ipvs.
	// Equal to the cmd: ipvsadm -A -t 10.10.0.1:8410 -s rr
	svc := libipvs.Service{
		Address:       net.ParseIP(ip),
		AddressFamily: syscall.AF_INET,
		Protocol:      libipvs.Protocol(syscall.IPPROTO_TCP),
		Port:          port,
		SchedName:     libipvs.RoundRobin,
	}

	if err := handler.NewService(&svc); err != nil {
		fmt.Println(err.Error())
	}

	// Bind the ip address to the NIC (flannel.1 here)
	// Equal to the cmd: ip addr add 10.10.0.1/24 dev flannel.1
	args := []string{"addr", "add", ip + "/24", "dev", "flannel.1"}
	_, err := exec.Command("ip", args...).CombinedOutput()
	if err != nil {
		fmt.Println(err.Error())
	}

	// Configure the iptable: add SNAT rule
	// Equal to the cmd: iptables -t nat -A POSTROUTING -m ipvs  --vaddr 10.9.0.1 --vport 12 -j MASQUERADE
	args = []string{"-t", "nat", "-A", "POSTROUTING", "-m", "ipvs", "--vaddr", ip, "--vport", strconv.Itoa(int(svc.Port)), "-j", "MASQUERADE"}
	_, err = exec.Command("iptables", args...).CombinedOutput()
	if err != nil {
		fmt.Println(err.Error())
	}

	return &svc
}

func bindEndpoint(svc *libipvs.Service, ip string, port uint16) *libipvs.Destination {
	dst := libipvs.Destination{
		Address:       net.ParseIP(ip),
		AddressFamily: syscall.AF_INET,
		Port:          port,
	}

	print(svc.Address.String() + ":" + strconv.Itoa(int(svc.Port)))

	args := []string{"-a", "-t", svc.Address.String() + ":" + strconv.Itoa(int(svc.Port)), "-r", ip + ":" + strconv.Itoa(int(port)), "-m"}
	res, err := exec.Command("ipvsadm", args...).CombinedOutput()
	if err != nil {
		fmt.Println(err.Error())
	}
	println(string(res))
	//if err := handler.NewDestination(svc, &dst); err != nil {
	//	fmt.Println(err.Error())
	//}

	return &dst
}

func unbindEndpoint(svc *libipvs.Service, dst *libipvs.Destination) {
	if err := handler.DelDestination(svc, dst); err != nil {
		fmt.Println(err.Error())
	}
}

func deleteService(svc *libipvs.Service) {
	if err := handler.DelService(svc); err != nil {
		fmt.Println(err.Error())
	}
}
