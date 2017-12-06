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
	ch        chan struct{}
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
	if e := tokenValid(s.Token); e != nil {
		log.Println("e != nil")
		s.Token = &eapi.JwtToken{}
		//<-s.ch
	}

	if tokenDead(s.Token) {
		log.Println("!tokenDead(s.Token)")
		go s.getToken()
		<-s.ch
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

addTokenDoRequest:
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

	if resp.StatusCode == http.StatusUnauthorized {
		log.Println("http.StatusUnauthorized")
		go s.getToken()
		<-s.ch
		//there is no guarantee that it will be executed only once
		goto addTokenDoRequest
	}

	if resp.StatusCode != http.StatusOK {
		log.Println("requestTokenized status -", resp.Status)
		return false
	}

	return true
}

func (s *implService1Client) getToken() {
	log.Print("\t\t/getToken")

	url := localhost + "/login"
	resp, err := http.Post(url, "application/json", bytes.NewReader(s.jsonBytes))
	//http://devs.cloudimmunity.com/gotchas-and-common-mistakes-in-go-golang/index.html#close_http_resp_body
	if resp != nil {
		defer ioCloserErrCheck(resp.Body)
	}
	if err != nil {
		log.Println(err)
		s.ch <- struct{}{}
		return
	}

	err = json.NewDecoder(resp.Body).Decode(&s.Token)
	if err != nil {
		log.Println(err)
		s.ch <- struct{}{}
		return
	}
	_, err = io.Copy(ioutil.Discard, resp.Body)
	if err != nil {
		log.Println(err)
		s.ch <- struct{}{}
		return
	}
	s.ch <- struct{}{}
	log.Println("\t\t\t/getToken finished")
}

func (s *implService1Client) getTokenInAdvance() {
	if s == nil {
		log.Fatalln("struct is nil")
	}
	for {
		if time.Now().After(s.Token.TimeToLive /*.Add(-1 * time.Second)*/) {
			log.Println("getTokenInAdvance")
			//log.Println("before", time.Now(), "\t\t", s.Token.TimeToLive, "\t\t", s.Token.TimeToLive.Add(-1*time.Second))
			go s.getToken()
			<-s.ch
		}
		time.Sleep(time.Second * 1)
	}
}

func tokenDead(t *eapi.JwtToken) bool {
	//return time.Now().Unix() >= t.TimeToLive.Unix()
	return time.Now().After(t.TimeToLive)
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
	isc.ch = make(chan struct{})

	go isc.getTokenInAdvance()

	return &isc
}
