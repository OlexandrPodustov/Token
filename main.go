package main

import (
	"bytes"
	"encoding/json"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/OlexandrPodustov/token/eapi"
)

const (
	localhost = "http://localhost:8080"
)

type implService1Client struct {
	sync.RWMutex
	Token     *eapi.JwtToken `json:"token,omitempty"`
	jsonBytes []byte
	ttl       time.Duration
}

func newClient() *implService1Client {
	raw, err := ioutil.ReadFile("conf.json")
	if err != nil {
		log.Fatalln(err)
	}

	var cl implService1Client
	cl.jsonBytes = raw
	cl.Token = &eapi.JwtToken{}

	return &cl
}

func main() {
	ss := newClient()
	http.HandleFunc("/main", ss.action)
	log.Fatal(http.ListenAndServe(":8082", nil))
}

func (s *implService1Client) action(w http.ResponseWriter, req *http.Request) {
	s.hello()
}

func (s *implService1Client) hello() {
	s.Lock()
	if s.isTokenDead() {
		s.getToken()
		s.Unlock()
	} else {
		s.Unlock()
		s.performRequestWithToken()
	}
}

func (s *implService1Client) performRequestWithToken() {
	url := localhost + "/hello"
	client := &http.Client{}

	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		log.Println(err)
		return
	}

	req.Header.Add("Authentication", s.Token.Token)

	resp, err := client.Do(req)
	if resp != nil {
		defer ioCloserErrCheck(resp.Body)
	}

	if err != nil {
		log.Println(err)
		return
	}

	if resp.StatusCode == http.StatusUnauthorized {
		s.Lock()
		if s.isTokenDead() {
			s.getToken()
			s.Unlock()
			s.performRequestWithToken()
		} else {
			s.Unlock()
		}
	}

	if resp.StatusCode != http.StatusOK {
		log.Println("performRequestWithToken status -", resp.Status)
		return
	}
}

func (s *implService1Client) getToken() {
	resp, err := http.Post(localhost+"/login", "application/json", bytes.NewReader(s.jsonBytes))
	if resp != nil {
		defer ioCloserErrCheck(resp.Body)
	}

	if err != nil {
		log.Println(err)
		return
	}

	err = json.NewDecoder(resp.Body).Decode(&s.Token)
	if err != nil {
		log.Println(err)
		return
	}

	_, err = io.Copy(ioutil.Discard, resp.Body)
	if err != nil {
		log.Println(err)
		return
	}

	miniTimeout := 10
	s.ttl = time.Until(s.Token.TimeToLive) - time.Millisecond*time.Duration(miniTimeout)

	time.AfterFunc(s.ttl, func() {
		log.Println("\t get token in advance")

		s.Lock()
		s.getToken()
		s.Unlock()
	})
}

func (s *implService1Client) isTokenDead() bool {
	return time.Now().After(s.Token.TimeToLive)
}

func ioCloserErrCheck(f io.Closer) {
	if err := f.Close(); err != nil {
		log.Println(err)
	}
}
