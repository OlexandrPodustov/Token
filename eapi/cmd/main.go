package main

import (
	"encoding/json"
	"github.com/dgrijalva/jwt-go"
	"log"
	"net/http"
	"sync"
	"time"
	"token/eapi"
)

const tokenTimeToLive = 6

var (
	mySigningKey = []byte("secret")
	db           = make(map[string]struct{})
)

type account struct {
	Name     string
	Password string
}

func main() {
	http.HandleFunc("/hello", requestHandlerTokenized)
	http.HandleFunc("/gettoken", createToken)

	log.Fatal(http.ListenAndServe(":8080", nil))
}

func validateToken(token string) bool {
	var l sync.RWMutex

	l.RLock()
	_, ok := db[token]
	l.RUnlock()

	if !ok {
		log.Println("The token is not valid")
		return false
	}
	log.Println("The token is OK")

	return true
}

func requestHandlerTokenized(w http.ResponseWriter, req *http.Request) {
	//todo: get token from req instead of this kludge
	//t := "tt"
	v := req.Header.Get("Authentication")
	log.Println(" token ", v)

	if !validateToken(v) {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	w.WriteHeader(http.StatusOK)
	return
}

func createToken(w http.ResponseWriter, req *http.Request) {
	var acc account
	var l sync.RWMutex

	err := json.NewDecoder(req.Body).Decode(&acc)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	//log.Println("parsed json - ", acc)

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"username": acc.Name,
		"password": acc.Password,
	})
	tokenString, err := token.SignedString(mySigningKey)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	log.Println("token has been created")

	l.RLock()
	db[tokenString] = struct{}{}
	l.RUnlock()

	tokenCreated := eapi.JwtToken{
		Token:      tokenString,
		TimeToLive: time.Now().Add(tokenTimeToLive * time.Second),
	}
	//log.Println("time.Now() - ", time.Now())

	b, err := json.Marshal(tokenCreated)
	if err != nil {
		log.Println("can't Marshal token", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Write(b)
	w.WriteHeader(http.StatusOK)

	go sanitizer(tokenCreated)

	return
}

func sanitizer(token eapi.JwtToken) {
	var l sync.RWMutex

	//log.Println("db before delete", db)
	//don't know how to trigger smth at some time. temporarily will so this via sleep
	//time.Sleep(token.TimeToLive)
	time.Sleep(tokenTimeToLive * time.Second)
	l.RLock()
	delete(db, token.Token)
	l.RUnlock()
	//log.Println(db)
}
