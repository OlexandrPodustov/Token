package main

import (
	"log"
	"net/http"
	"sync"
	"time"
)

type implService1Client3 struct {
	sync.RWMutex
	chanGetToken chan bool
	ttl          time.Time
}

func newClient3() *implService1Client3 {
	i := &implService1Client3{}
	i.chanGetToken = make(chan bool, 1)

	return i
}

func main() {
	ss := newClient3()
	http.HandleFunc("/main", ss.action)
	log.Fatal(http.ListenAndServe(":8082", nil))
}

func (s *implService1Client3) action(w http.ResponseWriter, req *http.Request) {
	s.hello()
}

func (s *implService1Client3) hello() {
	s.chanGetToken <- true
	if s.isTokenDead() {
		log.Println("after?  ", time.Now().After(s.ttl))
		s.getToken()
	} else {
		s.performRequestWithToken()
	}
}

func (s *implService1Client3) performRequestWithToken() {
	<-s.chanGetToken
}

func (s *implService1Client3) getToken() {
	tokenAlive := 3
	s.ttl = time.Now().Add(time.Duration(tokenAlive) * time.Second)
	<-s.chanGetToken
}

func (s *implService1Client3) isTokenDead() bool {
	return time.Now().After(s.ttl)
}
