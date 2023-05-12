package apiobject

// dns := DNSConfig{
//    Name: "example-dns",
//    Kind: "CoreDNS",
//    Host: "example.com",
//    Paths: []Path{
//        {Address: "/sub1", Service: "service1", Port: 8080},
//        {Address: "/sub2", Service: "service2", Port: 8081},
//    },
//}

import "testing"

func TestDNSRecord(t *testing.T) {
	return DNSRcord{
		Name: "example-dns",
		Kind: "DNS",
		Host: "example.com",
		Paths: []Path{
			{Address: "/sub1", Service: "service1", Port: 8080},
			{Address: "/sub2", Service: "service2", Port: 8081},
		},
	}
}
