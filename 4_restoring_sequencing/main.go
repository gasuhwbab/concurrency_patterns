package main

import (
	"fmt"
	"math/rand"
	"time"
)

type Message struct {
	str  string
	wait chan bool
}

func fanIn(inputs ...<-chan Message) <-chan Message {
	c := make(chan Message)
	for i := range inputs {
		go func() {
			for {
				c <- <-inputs[i]
			}
		}()
	}
	return c
}

func boring(msg string) <-chan Message {
	c := make(chan Message)
	wait := make(chan bool)
	go func() {
		for i := 0; ; i++ {
			c <- Message{fmt.Sprintf("%s: %d", msg, i), wait}
			time.Sleep(time.Duration(rand.Intn(1e3)) * time.Millisecond)
			<-wait
		}
	}()
	return c
}

func main() {
	c := fanIn(boring("joe"), boring("ann"))
	for range 5 {
		msg := <-c
		fmt.Println(msg.str)
		msg.wait <- true
	}
	fmt.Println("You're both boring. I'm leaving")
}
