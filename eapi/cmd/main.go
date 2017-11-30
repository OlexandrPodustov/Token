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

var inMemDB = database{}

type database struct {
	sync.RWMutex
	mapa map[string]struct{}
}

type account struct {
	Name     string
	Password string
}

func main() {
	inMemDB.mapa = make(map[string]struct{})

	http.HandleFunc("/hello", handlerTokenized)
	http.HandleFunc("/login", createToken)

	log.Fatal(http.ListenAndServe(":8080", nil))
}

func validateToken(token string) bool {
	//inMemDB.Lock()
	inMemDB.RLock()
	_, ok := inMemDB.mapa[token]
	//inMemDB.Unlock()
	inMemDB.RUnlock()

	if !ok {
		log.Println("The token is not valid")
		return false
	}
	//	log.Println("The token is OK")

	return true
}

func handlerTokenized(w http.ResponseWriter, req *http.Request) {
	v := req.Header.Get("Authentication")

	if !validateToken(v) {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	w.WriteHeader(http.StatusOK)
	return
}

func createToken(w http.ResponseWriter, req *http.Request) {
	var receivedAccount account
	mySigningKey := []byte("secret")

	err1 := json.NewDecoder(req.Body).Decode(&receivedAccount)
	if err1 != nil {
		log.Println(err1)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	//log.Println("parsed json - ", receivedAccount)

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"username": receivedAccount.Name,
		"password": receivedAccount.Password,
	})
	tokenString, err2 := token.SignedString(mySigningKey)
	if err2 != nil {
		log.Println(err2)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	inMemDB.Lock()
	log.Println("new token has been written to the map")
	inMemDB.mapa[tokenString] = struct{}{}
	inMemDB.Unlock()

	tokenCreated := eapi.JwtToken{
		Token:      tokenString,
		TimeToLive: time.Now().Add(tokenTimeToLive * time.Second),
	}
	//log.Println("time.Now() - ", time.Now())

	b, err3 := json.Marshal(tokenCreated)
	if err3 != nil {
		log.Println("can't Marshal token", err3)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
	_, err4 := w.Write(b)
	if err4 != nil {
		log.Println(err4)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	go sanitizer(tokenCreated)

	return
}

func sanitizer(token eapi.JwtToken) {
	//don't know how to trigger smth at some time.
	//temporarily will so this via sleep
	time.Sleep(tokenTimeToLive * time.Second)

	inMemDB.Lock()
	delete(inMemDB.mapa, token.Token)
	log.Println("delete completed successfully")
	inMemDB.Unlock()
}
