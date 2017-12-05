//review the synchronization completely
package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"sync"
	"time"

	"./eapi"
)

const localhost = "http://localhost:8080"

type service1Client interface {
	hello() error
}

type implService1Client struct {
	sync.RWMutex
	ch chan struct{}
	//UserName  string         `json:"user"`
	//Password  string         `json:"password"`
	Token     *eapi.JwtToken `json:"token, omitempty"`
	jsonBytes []byte
}

func main() {
	ss := newClient()
	http.HandleFunc("/main", ss.action)
	log.Fatal(http.ListenAndServe(":8082", nil))
}

func (s *implService1Client) action(w http.ResponseWriter, req *http.Request) {
	err := s.hello()
	if err != nil {
		log.Fatalln(err)
	}
}

func (s *implService1Client) hello() error {
	//log.Println("hello")

	//add protection from the two simultaneous requests
	e := tokenValid(s.Token)
	if e != nil {
		log.Println("e != nil")
		s.getToken()
	}
	//add protection from the two simultaneous requests
	if tokenAlive(s.Token) {
		log.Println("!tokenAlive(s.Token)")
		s.getToken()
	}

	ok := s.requestTokenized()
	if !ok {
		return fmt.Errorf("requestTokenized failed")
	}

	return nil
}

func (s *implService1Client) requestTokenized() bool {
	url := localhost + "/hello"
	client := &http.Client{}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		log.Println(err)
		return false
	}
	req.Header.Add("Authentication", s.Token.Token)

	resp, err := client.Do(req)
	//http://devs.cloudimmunity.com/gotchas-and-common-mistakes-in-go-golang/index.html#close_http_resp_body
	if resp != nil {
		defer ioCloserErrCheck(resp.Body)
	}
	if err != nil {
		log.Println(err)
		return false
	}

	//check status tokenexpired - get token and try again
	//+protection from simultaneous requests
	if resp.StatusCode == http.StatusUnauthorized {
		log.Println("http.StatusUnauthorized")
		s.getToken()
	}

	//if status error is different - log
	if resp.StatusCode != http.StatusOK {
		log.Println("requestTokenized status -", resp.Status)
		return false
	}

	return true
}

func (s *implService1Client) getToken() {
	log.Println("/getToken")

	url := localhost + "/login"
	resp, err := http.Post(url, "application/json", bytes.NewReader(s.jsonBytes))
	//http://devs.cloudimmunity.com/gotchas-and-common-mistakes-in-go-golang/index.html#close_http_resp_body
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

	//s.ch <- struct{}{}
	log.Println("/getToken finished")
}

func (s *implService1Client) getTokenInAdvance() {
	if s == nil {
		log.Fatalln("struct is nil")
	}
	for {
		if time.Now().After(s.Token.TimeToLive.Add(-1 * time.Second)) {
			log.Println("\t getTokenInAdvance")
			//log.Println("before", time.Now(), "\t\t", s.Token.TimeToLive, "\t\t", s.Token.TimeToLive.Add(-1*time.Second))
			s.Lock()
			s.getToken()
			//go s.getToken()
			//<-s.ch
			s.Unlock()
			//log.Println("after", time.Now(), "\t", s.Token.TimeToLive)
		}
		time.Sleep(time.Second * 1)
	}
}

func tokenAlive(t *eapi.JwtToken) bool {
	//log.Println("tokenAlive")
	return time.Now().Unix() >= t.TimeToLive.Unix()
}

func tokenValid(t *eapi.JwtToken) error {
	if t == nil {
		return fmt.Errorf("nil Token value")
	}
	return nil
}

func ioCloserErrCheck(f io.Closer) {
	if err := f.Close(); err != nil {
		log.Println(err)
	}
}

func newClient() *implService1Client {
	log.Println("newClient")
	var isc = implService1Client{}
	fileContent, err := ioutil.ReadFile("conf.json")
	if err != nil {
		log.Println(err)
		return nil
	}
	isc.jsonBytes = fileContent

	isc.Token = &eapi.JwtToken{}
	//isc.ch = make(chan struct{}, 1)
	isc.ch = make(chan struct{})
	go isc.getTokenInAdvance()
	return &isc
}

// time package - functions
// how to work with goroutines
