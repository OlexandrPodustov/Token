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

const (
	localhost   = "http://localhost:8080"
	cycleAmount = 10000
)

type service1Client interface {
	hello() error
}

type implService1Client struct {
	sync.RWMutex
	//chanSync     chan struct{}
	//chanGetToken chan bool
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

func (s *implService1Client) isTokenOK() bool {
	if tokenDead(s.Token) {
		log.Println("tokenDead")
		return false
	}
	return true
}

func (s *implService1Client) hello() error {
	//select {
	//case s.chanSync <- struct{}{}:
	if !s.isTokenOK() {
		s.getToken()
	}
	//	select {
	//	case <-s.chanGetToken:
	//		fmt.Println("successfully read after getToken")
	//		<-s.chanSync
	//	default:
	//		fmt.Println("token is not OK")
	//	}
	//default:
	//	log.Println("too many attemts to call getToken")
	//}

	if !s.requestTokenized() {
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

	if resp.StatusCode == http.StatusUnauthorized {
		log.Println("http.StatusUnauthorized")
		//s.getToken()
		//<-s.chanSync
		//<-s.gtCh
		//there is no guarantee that it will be executed only once
		//not with a goto addTokenDoRequest or do.once
	}

	if resp.StatusCode != http.StatusOK {
		log.Println("requestTokenized status -", resp.Status)
		return false
	}

	return true
}

func (s *implService1Client) getToken() {
	// defer func() {
	// 	//s.chanGetToken <- true
	// }()

	log.Print("\t\t/getToken")

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
	//s.chanSync <- struct{}{}
	log.Println("\t\t\t/getToken finished")
}

func (s *implService1Client) getTokenInAdvance() {
	for {
		if time.Now().After(s.Token.TimeToLive /*.Add(-1 * time.Second)*/) {
			//log.Println("getTokenInAdvance")
			s.getToken()
			//<-s.chanSync
			//<-s.gtCh
		}
		//avoid sleep usage
		time.Sleep(time.Second * 1)
	}
}

func tokenDead(t *eapi.JwtToken) bool {
	return time.Now().After(t.TimeToLive)
}

func tokenExist(t *eapi.JwtToken) bool {
	return !(t == nil)
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
		log.Fatalln(err)
	}
	isc.jsonBytes = fileContent

	isc.Token = &eapi.JwtToken{}
	//isc.chanSync = make(chan struct{}, 1)
	//isc.gtCh = make(chan bool, 1)

	go isc.getTokenInAdvance()

	return &isc
}
