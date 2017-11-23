package main

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"time"
	"token/eapi"
)

const localhost = "http://localhost:8080/"

func main() {
	var j = []byte(`{ "name": "olexa", "password": "pass" }`)
	token := authenticate(j)
	tokenAlive(token)
	//requestTokenized(token)
}

func requestTokenized(token string) {
	url := localhost + "hello"
	client := &http.Client{}

	req, err := http.NewRequest("GET", url, nil)
	req.Header.Add("Authentication", token)

	resp, err := client.Do(req)
	if err != nil {
		log.Println(err)
	}
	defer deferredClose(resp)

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Println(err)
	}

	log.Println(string(body))
	log.Println(resp.Status)

}

func authenticate(jsonStr []byte) (token *eapi.JwtToken) {
	url := localhost + "gettoken"
	resp, err := http.Post(url, "", bytes.NewBuffer(jsonStr))
	if err != nil {
		log.Println(err)
	}
	defer deferredClose(resp)

	var t eapi.JwtToken
	json.NewDecoder(resp.Body).Decode(&t)
	log.Println(t)
	return &t
}

func tokenAlive(t *eapi.JwtToken) bool {
	if time.Now().Unix() > t.TimeToLive.Unix() {
		log.Println("Token has died", t)
		return false
	}
	return true
}

func deferredClose(res *http.Response) {
	if err := res.Body.Close(); err != nil {
		log.Println("failed to close response body in deferedClose func")
	}
}
