package main

import (
	"bytes"
	"io/ioutil"
	"log"
	"net/http"
	"token/eapi"
)

const localhost = "http://localhost:8080/"

func main() {
	var j = []byte(`{ "name": "olexa", "password": "pass" }`)
	token := authenticate(j)
	log.Println("tst")

	var t eapi.JwtToken
	checkTokenTimeToLive(t)
	requestTokenized(token)

}

func requestTokenized(token string) {
	url := localhost
	resp, err := http.Get(url)
	if err != nil {
		log.Println(err)
	}
	defer deferredClose(resp)

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Println(err)
	}

	log.Println(string(body))

}

func authenticate(jsonStr []byte) (token string) {
	url := localhost + "gettoken"
	resp, err := http.Post(url, "", bytes.NewBuffer(jsonStr))
	if err != nil {
		log.Println(err)
	}
	defer deferredClose(resp)

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Println(err)
	}
	token = string(body)
	return token
}

func checkTokenTimeToLive(t eapi.JwtToken) (ok bool) {

	return
}

func deferredClose(res *http.Response) {
	if err := res.Body.Close(); err != nil {
		log.Println("failed to close response body in deferedClose func")
	}
}
