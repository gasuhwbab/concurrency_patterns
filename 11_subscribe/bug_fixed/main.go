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
	closing chan chan error
}

func (s *sub) loop() {
	var pending []Item
	var next time.Time
	var err error
	for {
		var firstDelay time.Duration
		if now := time.Now(); next.After(now) {
			firstDelay = next.Sub(now)
		}
		startFetch := time.After(firstDelay)

		var firstItem Item
		//var updates chan Item
		if len(pending) > 0 {
			firstItem = pending[0]
			//updates = s.updates
		}

		select {
		case errc := <-s.closing:
			errc <- err
			close(s.updates)
			return
		case <-startFetch:
			var fetched []Item
			fetched, _, err := s.fetcher.Fetch()
			if err != nil {
				// next = time.Now().Add(10 * time.Second)
				break
			}
			pending = append(pending, fetched...)
		case s.updates <- firstItem:
			pending = pending[1:]
		}
	}
}

func (s *sub) Updates() <-chan Item {
	return s.updates
}

func (s *sub) Close() error {
	errc := make(chan error)
	s.closing <- errc
	return <-errc
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
