package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"sync"
	"time"
	"token/eapi"
)

const localhost = "http://localhost:8080/"

type service1Client interface {
	hello() error
}

type implService1Client struct {
	sync.RWMutex
	UserName string         `json:"user"`
	Password string         `json:"password"`
	Token    *eapi.JwtToken `json:"token, omitempty"`
}

var client service1Client

func init() {
	log.Println("init")
	client = newClient()
}

func main() {
	http.HandleFunc("/main", action)
	log.Fatal(http.ListenAndServe(":8082", nil))
}

func action(w http.ResponseWriter, req *http.Request) {
	err := client.hello()
	if err != nil {
		log.Println(err)
	}
	return
}

func (s *implService1Client) hello() error {
	log.Println("hello")
	err2 := tokenValid(s.Token)
	if err2 != nil {
		s.getToken()
	}
	if tokenAlive(s.Token) {
		s.getToken()
	}

	if !s.requestTokenized() {
		return fmt.Errorf("requestTokenized failed")
	}
	return nil
}
func (s *implService1Client) makeJSON() (string, error) {

	return `{ "name": "olexa", "password": "pass123" }`, nil
}

func (s *implService1Client) requestTokenized() bool {
	url := localhost + "hello"
	client := &http.Client{}

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		log.Println(err)
	}
	req.Header.Add("Authentication", s.Token.Token)

	resp, err2 := client.Do(req)
	if err2 != nil {
		log.Println(err2)
	}
	defer deferredRespClose(resp)

	if resp.StatusCode != http.StatusOK {
		log.Println("requestTokenized nOk", resp.Status)
		return false
	}
	log.Println("requestTokenized  Ok", resp.Status)

	return true
}

func (s *implService1Client) getToken() {
	log.Println("getToken")
	st, err := s.makeJSON()
	if err != nil {
		log.Println(err)
		return
	}

	url := localhost + "login"
	resp, err := http.Post(url, "", bytes.NewBuffer([]byte(st)))
	if err != nil {
		log.Println(err)
		return
	}
	defer deferredRespClose(resp)

	err2 := json.NewDecoder(resp.Body).Decode(&s.Token)
	if err2 != nil {
		log.Println(err2)
		return
	}
}

func getTokenInAdvance() {

}

func tokenAlive(t *eapi.JwtToken) bool {
	log.Println("tokenAlive")
	if time.Now().Unix() >= t.TimeToLive.Unix() {
		//log.Println("Token has died")
		return false
	}
	return true
}

func tokenValid(t *eapi.JwtToken) error {
	log.Println("tokenValid")
	if t == nil {
		return fmt.Errorf("nil Token value")
	}
	return nil
}

func deferredRespClose(res *http.Response) {
	if err := res.Body.Close(); err != nil {
		log.Println("failed to close response body in deferedClose func")
	}
}
func deferredFileClose(f *os.File) {
	if err := f.Close(); err != nil {
		log.Println("failed to close response body in deferedClose func")
	}
}

func newClient() *implService1Client {
	log.Println("newClient")

	var isc = implService1Client{}
	file, err := os.Open("conf.json")
	if err != nil {
		log.Fatalln(err)
	}

	defer deferredFileClose(file)

	err2 := json.NewDecoder(file).Decode(&isc)
	if err2 != nil {
		log.Fatalln(err2)
	}

	return &isc
}
