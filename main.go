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

	"./eapi"
)

const (
	localhost   = "http://localhost:8080"
	cycleAmount = 10000
)

//type service1Client interface {
//	hello()
//}

type implService1Client struct {
	sync.RWMutex
	chanGetToken chan bool
	Token        *eapi.JwtToken `json:"token, omitempty"`
	jsonBytes    []byte
	ttl          time.Duration
}

func newClient() *implService1Client {
	var isc = implService1Client{}
	var err error
	isc.jsonBytes, err = ioutil.ReadFile("conf.json")
	if err != nil {
		log.Fatalln(err)
	}
	isc.Token = &eapi.JwtToken{}
	isc.chanGetToken = make(chan bool, 1)

	go isc.getTokenInAdvance()

	return &isc
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
	select {
	case s.chanGetToken <- true:
		if s.isTokenDead() {
			log.Println("hello - get token")
			s.getToken()
		} else {
			<-s.chanGetToken
		}
		//use afnerfunc
	default:
		//log.Println("hello check if token is dead skipped")
	}
	if !s.isTokenDead() {
		if !s.performRequestWithToken() {
			//log.Println("performRequestWithToken failed")
		}
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
		select {
		case s.chanGetToken <- true:
			if s.isTokenDead() {
				log.Println("StatusUnauthorized - get token")
				s.getToken()
				s.performRequestWithToken()
			} else {
				<-s.chanGetToken
				s.performRequestWithToken()
			}
		}
		//s.getToken()
		//<-s.chanSync
		//<-s.gtCh
		//there is no guarantee that it will be executed only once
		//not with a goto addTokenDoRequest or do.once
	}

	if resp.StatusCode != http.StatusOK {
		log.Println("performRequestWithToken status -", resp.Status)
		return false
	}

	return true
}

func (s *implService1Client) getToken() {
	defer func() {
		<-s.chanGetToken
	}()
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
	s.ttl = s.Token.TimeToLive.Sub(time.Now()) // - time.Millisecond*10
	log.Println("\t\t\t/getToken finished")
}

func (s *implService1Client) getTokenInAdvance() {
	for {
		select {
		case <-time.After(s.ttl):
			select {
			case s.chanGetToken <- true:
				if s.isTokenDead() {
					log.Println("getTokenInAdvance - get token")
					s.getToken()
				} else {
					<-s.chanGetToken
				}
			default:

			}
		}

	}
}

func (s *implService1Client) isTokenDead() bool {
	return time.Now().After(s.Token.TimeToLive)
}

func ioCloserErrCheck(f io.Closer) {
	if err := f.Close(); err != nil {
		log.Println(err)
	}
}
