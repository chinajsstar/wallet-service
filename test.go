package main

import (
	"fmt"
)

func main() {
	slice := make([]int, 3, 5)

	for i:=0; i<10; i++ {
		slice = append(slice, i)
		fmt.Printf("pointer address:%d\n", &(slice[0]))
	}

	for i, value := range slice {
		fmt.Printf("i:%d, value:%d\n",i,  value)
	}
}
