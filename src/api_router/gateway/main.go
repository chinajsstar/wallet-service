package main

import (
	"../base/service"
	"fmt"
	"time"
	"context"
	"sync"
)

const ServiceCenterName = "center"

func main() {
	wg := &sync.WaitGroup{}

	// 创建服务中心
	centerInstance,_ := service.NewServiceCenter(ServiceCenterName, ":8080", ":8081")

	// 启动服务中心
	ctx, cancel := context.WithCancel(context.Background())
	centerInstance.Start(ctx, wg)

	time.Sleep(time.Second*2)
	for ; ;  {
		fmt.Println("Input 'quit' to quit...")
		var input string
		fmt.Scanln(&input)

		if input == "quit" {
			cancel()
			break;
		}
	}

	fmt.Println("Waiting all routine quit...")
	wg.Wait()
	fmt.Println("All routine is quit...")
}

