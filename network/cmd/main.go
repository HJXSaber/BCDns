package main

import (
	"BCDns_0.1/network/service"
	"time"
)

func main(){
	service.P2PNet.BroadcastMsg([]byte("hello"))
	for {
		time.Sleep(time.Second)
	}
}