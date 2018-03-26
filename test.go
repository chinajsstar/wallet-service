package main

import (
	"fmt"
	"time"
)

func main() {
	channel := make(chan int, 256)

	go func() {
		for {
			fmt.Printf("value is %d\n", <-channel)
		}
	}()

	for i:=0; i<10; i++{
		channel <- i
		time.Sleep(time.Second * 1)
	}


}
