package main

import "fmt"

func main() {
	ch := make(chan string)

	go func() {
		ch <- "from channel"
		close(ch)
	}()
	fmt.Println(<-ch)
}
