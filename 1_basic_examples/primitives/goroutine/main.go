package main

import (
	"fmt"
	"time"
)

func main() {
	go func() {
		fmt.Println("from goroutine")
	}()
	time.Sleep(time.Millisecond)
}
