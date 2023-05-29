package nginx

import (
	log "github.com/sirupsen/logrus"
	"minik8s/pkg/apiobject"
	"os"
	"os/exec"
	"strconv"
	"text/template"
)

type Location struct {
	Path string
	IP   string
	Port string
}

type Server struct {
	Port       string
	ServerName string
	Locations  []Location
}

type Config struct {
	Servers []Server
}

// GenerateConfig generate the nginx config file
func GenerateConfig(configs []apiobject.DNSRecord) {

	file, err := os.OpenFile("/home/mini-k8s/pkg/kubedns/config/nginx.conf", os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0666)
	if err != nil {
		log.Error("[GenerateConfig] error opening file: ", err)
	}
	defaultHeader := `
worker_processes  5;  ## Default: 1
error_log  ./error.log debug;
pid        ./nginx.pid;
worker_rlimit_nofile 8192;

events {
  worker_connections  4096;  ## Default: 1024
}
`
	bytes, err := file.WriteString(defaultHeader)
	if err != nil {
		log.Error("[GenerateConfig] error writing to file: ", err)
	}
	log.Debug("[GenerateConfig] wrote ", bytes, " bytes to file")

	tmpl := template.Must(template.ParseFiles("/home/mini-k8s/pkg/kubedns/nginx/nginx.tmpl"))

	// generate the servers
	ServerList := make([]Server, 0)
	for _, config := range configs {
		// generate the locations
		locations := make([]Location, 0)
		for _, path := range config.Paths {
			location := Location{
				Path: path.Service,
				IP:   path.Address,
				Port: strconv.Itoa(path.Port),
			}
			locations = append(locations, location)
		}
		server := Server{
			Port:       "80",
			ServerName: config.Host,
			Locations:  locations,
		}
		ServerList = append(ServerList, server)
	}

	config := Config{
		Servers: ServerList,
	}
	err = tmpl.Execute(file, config)
	if err != nil {
		log.Error("[GenerateConfig] error executing template: ", err)
	}

	file.Close()
}


func ReloadNginx() error {
	cmd := exec.Command("pkill", "nginx")
	err := cmd.Run()
	cmd = exec.Command("nginx", "-c", "/home/mini-k8s/pkg/kubedns/config/nginx.conf")
	err = cmd.Run()
	if err != nil {
		log.Error("[ReloadNginx] cmd.Run() failed with ", err.Error())
		return err
	}
	return nil
}
