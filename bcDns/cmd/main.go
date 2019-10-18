package main

import (
	"fmt"
	"time"
)

func main() {
	for t := range time.Tick(time.Second * 2) {
		fmt.Println(t, "hello world")
	}
}