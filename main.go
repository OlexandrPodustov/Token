package main

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"
	"time"
	"token/eapi"
)

const localhost = "http://localhost:8080/"

var j = []byte(`{ "name": "olexa", "password": "pass" }`)

func main() {
	var acquiredToken = new(eapi.JwtToken)

	if !requestTokenized(acquiredToken) {
		acquiredToken = authenticate(j)
	}
	time.Sleep(5 * time.Second)
	if !requestTokenized(acquiredToken) {
		acquiredToken = authenticate(j)
	}
}

func requestTokenized(token *eapi.JwtToken) bool {
	if !tokenAlive(token) {
		token = authenticate(j)
	}
	time.Sleep(5 * time.Second)
	url := localhost + "hello"
	client := &http.Client{}

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		log.Println(err)
	}
	req.Header.Add("Authentication", token.Token)

	resp, err2 := client.Do(req)
	if err2 != nil {
		log.Println(err2)
	}
	defer deferredClose(resp)

	log.Println("requestTokenized", resp.Status)
	if resp.StatusCode != http.StatusOK {
		return false
	}

	return true
}

func authenticate(jsonStr []byte) (token *eapi.JwtToken) {
	url := localhost + "gettoken"
	resp, err := http.Post(url, "", bytes.NewBuffer(jsonStr))
	if err != nil {
		log.Println(err)
	}
	defer deferredClose(resp)

	var t eapi.JwtToken
	err2 := json.NewDecoder(resp.Body).Decode(&t)
	if err2 != nil {
		log.Println(err2)
		return
	}
	return &t
}

func tokenAlive(t *eapi.JwtToken) bool {
	//log.Printf("tokenAlive time.Now().Unix() %+v", time.Now().Unix())
	//log.Printf("tokenAlive t.TimeToLive.Unix() %+v", t.TimeToLive.Unix())
	if time.Now().Unix() >= t.TimeToLive.Unix() {
		log.Println("Token has died")
		return false
	}
	return true
}

func deferredClose(res *http.Response) {
	if err := res.Body.Close(); err != nil {
		log.Println("failed to close response body in deferedClose func")
	}
}
