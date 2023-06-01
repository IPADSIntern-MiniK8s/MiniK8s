package main

import (
    "time"
)

func main(){
    var list = make([]uint64,1,1)
    var i uint64
    time.Sleep(time.Second*10)
    //10M*8byte totally cost >100M memory
    for i=0;i<10000000;i++{
	list = append(list,i)
    }
    time.Sleep(time.Duration(120)*time.Second)
}
