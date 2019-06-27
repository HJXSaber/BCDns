package utils

import (
	"fmt"
	"testing"
	"time"
)

func TestChannel(t *testing.T) {
	c := make(chan int, 5)
	down := make(chan int)
	go func() {
		for i := 0; i < 10; i ++ {
			c <- i
		}
		close(c)
		fmt.Println("Finished")
	}()

	go func() {
		for i := 0; i < 10; i++ {
			fmt.Println(<- c)
			time.Sleep(time.Second)
		}
		close(down)
	}()
	_ = <- down
}