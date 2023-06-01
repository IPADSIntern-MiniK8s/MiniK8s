package main
 
import (
    "io"
    "os"
    "net/http"
)
 
func main() {
    port:=os.Getenv("port")
    http.HandleFunc("/",func(w http.ResponseWriter,request *http.Request){io.WriteString(w,"http connect success\n")})
    _ = http.ListenAndServe(":"+port, nil)
}
