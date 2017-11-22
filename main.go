package main

import (
	"bytes"
	"io/ioutil"
	"log"
	"net/http"
	"./"
)

func deferredClose(res *http.Response) {
	if err := res.Body.Close(); err != nil {
		panic("failed to close response body in deferedClose func")
	}
}

const localhost = "http://localhost:8080/"

func main() {
	var j = []byte(`{ "name": "olexa", "password": "pass" }`)
	token := authenticate(j)
	println("tst")

	requestTokenized(token)

}

func requestTokenized(token string) {
	url := localhost
	resp, err := http.Get(url)
	if err != nil {
		panic(err)
	}
	defer deferredClose(resp)

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		panic(err)
	}

	log.Println(string(body))

}

func authenticate(jsonStr []byte) (token string) {
	url := localhost + "gettoken"
	resp, err := http.Post(url, "", bytes.NewBuffer(jsonStr))
	if err != nil {
		panic(err)
	}
	defer deferredClose(resp)

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		panic(err)
	}
	token = string(body)
	return token
}

func checkTokenTimeToLive(t JwtToken) (ok bool) {

	return
}
