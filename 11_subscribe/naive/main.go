package main

import (
	"fmt"
	"time"
)

type Item struct {
	Title, Channel, Guid string
}

func Fetch(domain string) Fetcher

type Fetcher interface {
	Fetch() (items []Item, next time.Time, err error)
}

type Subscription interface {
	Updates() <-chan Item
	Close() error
}

func Subscribe(fetcher Fetcher) Subscription {
	s := &sub{
		fetcher: fetcher,
		updates: make(chan Item),
	}
	go s.loop()
	return s
}

type sub struct {
	fetcher Fetcher   // fetches items
	updates chan Item // delivers items to the user
	closed  bool
	err     error
}

func (s *sub) loop() {
	for {
		if s.closed {
			close(s.updates)
			return
		}
		items, next, err := s.fetcher.Fetch()
		if err != nil {
			s.err = err
			time.Sleep(10 * time.Second)
			continue
		}
		for _, item := range items {
			s.updates <- item
		}
		if now := time.Now(); next.After(now) {
			time.Sleep(next.Sub(now))
		}
	}
}

func (s *sub) Updates() <-chan Item {
	return s.updates
}

func (s *sub) Close() error {
	s.closed = true
	return s.err
}

func Merge(subs ...Subscription) Subscription {
	s := new(sub)
	for _, sub := range subs {
		go func(sub Subscription) {
			for it := range sub.Updates() {
				s.updates <- it
			}
		}(sub)
	}
	return s
}

func main() {
	merged := Merge(
		Subscribe(Fetch("blog.golang.org")),
		Subscribe(Fetch("googleblog.blogspot.com")),
		Subscribe(Fetch("googledevelopers.blogspot.com")),
	)
	time.AfterFunc(3*time.Second, func() { merged.Close() })
	for it := range merged.Updates() {
		fmt.Println(it.Channel, it.Title)
	}

	panic("show me the stacks")
}
