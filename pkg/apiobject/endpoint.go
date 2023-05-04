package apiobject

// Endpoints is a collection of endpoints that implement the actual service.  Example:
//
//	 Name: "mysvc",
//	 Subsets: [
//	   {
//	     Addresses: [{"ip": "10.10.1.1"}, {"ip": "10.10.2.2"}],
//	     Ports: [{"name": "a", "port": 8675}, {"name": "b", "port": 309}]
//	   },
//	   {
//	     Addresses: [{"ip": "10.10.3.3"}],
//	     Ports: [{"name": "a", "port": 93}, {"name": "b", "port": 76}]
//	   },
//	]
type Endpoint struct {
	IP   string `json:"ip"`
	Port int32  `json:"port"`
}

// service中一个port对一个endpoints对象
type Endpoints struct {
	Name    string     `json:"name"` //same as the name of service
	IP      string     `json:"clusterIP"`
	Port    int32      `json:"port"`
	Subsets []Endpoint `json:"subsets"`
}

//type EndpointSubset struct {
//	Addresses []string
//	Ports     []EndpointPort
//}
