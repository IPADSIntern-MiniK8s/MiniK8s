
package main

import (
    "fmt"
    "net"
    "net/http"
    "os"
    "strings"
    "io"
)

func main() {
    addrs, err := net.InterfaceAddrs()
    if err != nil {
        fmt.Println(err)
        return
    }

    var ip string
    for _, addr := range addrs {
        //fmt.Println(addr)
	addrStr :=  fmt.Sprintf("%v",addr)
	if strings.Contains(addrStr,"/24"){
		ip = addrStr
	}
    }
    port:=os.Getenv("port")
    http.HandleFunc("/",func(w http.ResponseWriter,request *http.Request){io.WriteString(w,"i'm "+ip+"\n")})
    _ = http.ListenAndServe(":"+port, nil)
}
