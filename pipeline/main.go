package main

import "fmt"

func main() {
	nums := []int{1, 2, 3}
	numsCh := convToCh(nums)
	sqNumsCh := sqCh(numsCh)

	for n := range sqNumsCh {
		fmt.Println(n)
	}
}

func convToCh(nums []int) chan int {
	out := make(chan int)
	go func() {
		for _, n := range nums {
			out <- n
		}
		close(out)
	}()
	return out
}

func sqCh(ch <-chan int) chan int {
	out := make(chan int)
	go func() {
		for n := range ch {
			out <- n * n
		}
		close(out)
	}()
	return out
}
