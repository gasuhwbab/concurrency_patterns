package main

import (
	"fmt"
	"time"
)

const maxPending = 10

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

type fetchResult struct {
	fetched []Item
	next    time.Time
	err     error
}

func (s *sub) loop() {
	var pending []Item
	var next time.Time
	var seen = make(map[string]bool)
	var fetchDone chan fetchResult
	var err error
	for {
		var fetchDelay time.Duration
		if now := time.Now(); next.After(now) {
			fetchDelay = next.Sub(now)
		}
		var startFetch <-chan time.Time
		if fetchDone == nil && len(pending) < maxPending {
			startFetch = time.After(fetchDelay) // enable featch case
		}

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
			fetchDone = make(chan fetchResult, 1)
			go func() {
				fetched, next, err := s.fetcher.Fetch()
				fetchDone <- fetchResult{fetched: fetched, next: next, err: err}
			}()
		case result := <-fetchDone:
			fetchDone = nil
			if result.err != nil {
				result.next = time.Now().Add(10 * time.Second)
				break
			}
			for _, it := range result.fetched {
				if !seen[it.Guid] {
					pending = append(pending, it)
					seen[it.Guid] = true
				}
			}
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

type merge struct {
	subs    []Subscription
	updates chan Item
	quit    chan struct{}
	errs    chan error
}

func Merge(subs ...Subscription) Subscription {
	m := &merge{
		subs:    subs,
		updates: make(chan Item),
		quit:    make(chan struct{}),
		errs:    make(chan error),
	}
	for _, sub := range subs {
		go func(s Subscription) {
			for {
				var it Item
				select {
				case <-s.Updates():
				case <-m.quit:
					m.errs <- s.Close()
					return
				}
				select {
				case m.updates <- it:
				case <-m.quit:
					m.errs <- s.Close()
					return
				}
			}
		}(sub)
	}
	return m
}

func (m *merge) Updates() <-chan Item {
	return m.updates
}

func (m *merge) Close() error {
	close(m.quit)
	for range m.subs {
		if err := <-m.errs; err != nil {
			return err
		}
	}
	close(m.updates)
	return nil
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
