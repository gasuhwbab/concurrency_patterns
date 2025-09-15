package main

import "fmt"

func f(left, right chan int) {
	left <- 1 + <-right
}

func main() {
	const n = 1000
	leftmost := make(chan int)
	right := leftmost
	left := leftmost
	for range n {
		right = make(chan int)
		go f(left, right)
		left = right
	}

	go func(c chan int) {
		c <- 1
	}(right)
	fmt.Println(<-leftmost)
}

//          main
//            |
//			  |
//			  --------------------------------fork
//            |
//            ----------------
//
//
//
