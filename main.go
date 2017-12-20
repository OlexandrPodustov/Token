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

	"token/eapi"
)

const (
	localhost = "http://localhost:8080"
)

//type service1Client interface {
//	hello()
//}

type implService1Client struct {
	sync.RWMutex
	Token     *eapi.JwtToken `json:"token, omitempty"`
	jsonBytes []byte
	ttl       time.Duration
}

func newClient() *implService1Client {
	var implementationServiceClient = implService1Client{}
	var err error
	implementationServiceClient.jsonBytes, err = ioutil.ReadFile("conf.json")
	if err != nil {
		log.Fatalln(err)
	}
	implementationServiceClient.Token = &eapi.JwtToken{}

	return &implementationServiceClient
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
		log.Println("hello - get token")
		s.getToken()
		s.Unlock()
	} else {
		s.Unlock()
		s.performRequestWithToken()
	}
}

func (s *implService1Client) performRequestWithToken() bool {
	url := localhost + "/hello"
	client := &http.Client{}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		log.Println(err)
		return false
	}

	req.Header.Add("Authentication", s.Token.Token)

	resp, err := client.Do(req)
	if resp != nil {
		defer ioCloserErrCheck(resp.Body)
	}
	if err != nil {
		log.Println(err)
		return false
	}

	if resp.StatusCode == http.StatusUnauthorized {
		log.Println("http.StatusUnauthorized")
		s.Lock()
		if s.isTokenDead() {
			log.Println("StatusUnauthorized - get token")
			s.getToken()
			s.Unlock()
			s.performRequestWithToken()
		}
		s.Unlock()
	}

	if resp.StatusCode != http.StatusOK {
		log.Println("performRequestWithToken status -", resp.Status)
		return false
	}

	return true
}

func (s *implService1Client) getToken() {
	log.Print("\t\t/getToken")

	url := localhost + "/login"
	resp, err := http.Post(url, "application/json", bytes.NewReader(s.jsonBytes))
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
	s.ttl = s.Token.TimeToLive.Sub(time.Now()) - time.Millisecond*10

	time.AfterFunc(s.ttl, func() {
		s.Lock()
		log.Println("\t get token in advance")
		s.getToken()
		s.Unlock()
	})
	log.Println("\t\t\t/getToken finished")
}

func (s *implService1Client) isTokenDead() bool {
	return time.Now().After(s.Token.TimeToLive)
}

func ioCloserErrCheck(f io.Closer) {
	if err := f.Close(); err != nil {
		log.Println(err)
	}
}
