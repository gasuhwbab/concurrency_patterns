package main

import (
	"fmt"
	"time"
)

func main() {
	ch1 := make(chan string)
	ch2 := make(chan string)

	go func() {
		ch1 <- "first goroutine"
		close(ch1)
	}()

	go func() {
		ch2 <- "second goroutine"
		close(ch2)
	}()

	WithoutSleep(ch1, ch2) // every time prints "main goroutine"
	WithSleep(ch1, ch2)    // prints "first goroutine" or "second goroutine"
	WithoutSleep(ch1, ch2) // prints anither from WithSleep() function result
}

func WithSleep(ch1, ch2 chan string) {
	time.Sleep(time.Second)
	select {
	case msg := <-ch1:
		fmt.Println(msg)
	case msg := <-ch2:
		fmt.Println(msg)
	default:
		fmt.Println("main goroutine")
	}
}

func WithoutSleep(ch1, ch2 chan string) {
	select {
	case msg := <-ch1:
		fmt.Println(msg)
	case msg := <-ch2:
		fmt.Println(msg)
	default:
		fmt.Println("main goroutine")
	}
}
