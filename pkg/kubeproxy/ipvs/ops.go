package ipvs

import (
	"fmt"
	"github.com/mqliang/libipvs"
	log "github.com/sirupsen/logrus"
	"net"
	"os/exec"
	"strconv"
	"syscall"
	"time"
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

func AddService(ip string, port uint16) {
	serviceIP := ip + ":" + strconv.Itoa(int(port))
	if _, ok := Services[serviceIP]; ok {
		return
	}
	svc := addService(ip, port)
	Services[serviceIP] = &ServiceNode{
		Service:   svc,
		Visited:   true,
		Endpoints: map[string]*EndpointNode{},
	}
	log.Info("[kubeproxy] Add service ", serviceIP)
}

func addService(ip string, port uint16) *libipvs.Service {
	// Create a service struct and add it to the ipvs.
	// Equal to the cmd: ipvsadm -A -t 10.10.0.1:8410 -s rr
	svc := &libipvs.Service{
		Address:       net.ParseIP(ip),
		AddressFamily: syscall.AF_INET,
		Protocol:      libipvs.Protocol(syscall.IPPROTO_TCP),
		Port:          port,
		SchedName:     libipvs.RoundRobin,
	}

	if err := handler.NewService(svc); err != nil {
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

	return svc
}

func DeleteService(key string) {
	log.Info("[kubeproxy] Delete service ", key)
	node := Services[key]
	if node != nil {
		deleteService(node.Service)
	}
	delete(Services, key)
}

func deleteService(svc *libipvs.Service) {
	if err := handler.DelService(svc); err != nil {
		fmt.Println(err.Error())
	}
}

func AddEndpoint(key string, ip string, port uint16) {
	svc, exist := Services[key]
	for !exist {
		time.Sleep(1)
		log.Info("[proxy] Add Endpoint: service doesn't exist!")
		svc, exist = Services[key]
	}
	dst := bindEndpoint(svc.Service, ip, port)
	podIP := ip + ":" + strconv.Itoa(int(port))
	svc.Endpoints[podIP] = &EndpointNode{
		Endpoint: dst,
		Visited:  true,
	}
	log.Info("[kubeproxy] Add endpoint ", podIP, " service:", key)
}

func bindEndpoint(svc *libipvs.Service, ip string, port uint16) *libipvs.Destination {
	dst := libipvs.Destination{
		Address:       net.ParseIP(ip),
		AddressFamily: syscall.AF_INET,
		Port:          port,
	}

	//print(svc.Address.String() + ":" + strconv.Itoa(int(svc.Port)))
	args := []string{"-a", "-t", svc.Address.String() + ":" + strconv.Itoa(int(svc.Port)), "-r", ip + ":" + strconv.Itoa(int(port)), "-m"}
	_, err := exec.Command("ipvsadm", args...).CombinedOutput()
	if err != nil {
		fmt.Println(err.Error())
	}

	return &dst
}

func DeleteEndpoint(svcKey string, dstKey string) {
	if svc, ok := Services[svcKey]; ok {
		dst := svc.Endpoints[dstKey].Endpoint
		unbindEndpoint(svc.Service, dst)
		delete(svc.Endpoints, dstKey)
	}
	log.Info("[kubeproxy] Delete endpoint ", dstKey, " service:", svcKey)
}

func unbindEndpoint(svc *libipvs.Service, dst *libipvs.Destination) {
	if err := handler.DelDestination(svc, dst); err != nil {
		fmt.Println(err.Error())
	}
}
