package main

import (
	"encoding/json"
	"github.com/dgrijalva/jwt-go"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres"
	"io"
	"log"
	"net/http"
	"token/eapi"
)

const tokenTimeToLive = 30

var mySigningKey = []byte("secret")

type account struct {
	Name     string `gorm:"size:255"`
	Password string `gorm:"size:255"`
	eapi.JwtToken
	gorm.Model
}

func lastOrCreate(userName, userPassword string) (res bool) {
	db, err := gorm.Open("postgres", "host=localhost user=postgres sslmode=disable "+
		"password=mysecretpassword")
	if err != nil {
		log.Println("failed to connect database")
	}

	defer func() {
		if err := db.Close(); err != nil {
			log.Println("failed to connect database")
		}
	}()

	var acc account
	db.Last(&acc, "name = ?", userName)
	//log.Printf("before create %+v,%+v,%+v\n\n", acc.Name, acc.Password, acc.Password == userPassword)

	//todo: substitute by switch
	if acc.Name == "" {
		db.AutoMigrate(&account{})
		db.Create(&account{Name: userName, Password: userPassword})
		log.Println("The account was created, didn't existed")
		res = true
	} else if acc.Password != userPassword {
		log.Println("Client has provided wrong password")
	} else if acc.Name != "" && acc.Password == userPassword {
		log.Println("Authentication has been passed successfully")
		res = true
	}

	return res
}

func createToken(w http.ResponseWriter, req *http.Request) {
	var acc account

	err := json.NewDecoder(req.Body).Decode(&acc)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"username": acc.Name,
		"password": acc.Password,
	})

	var tokenString string

	if lastOrCreate(acc.Name, acc.Password) {
		//todo: instead of use mySigningKey - retrieve the secret from a db
		tokenString, err = token.SignedString(mySigningKey)
		log.Println("token has been created")
		if err != nil {
			log.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	}

	tok := eapi.JwtToken{tokenString, tokenTimeToLive}

	//todo: save token into db for further validation
	log.Printf("+%v", tok)
	w.Header().Set("AUTH-TOKEN", tok.Token)

}

func helloServer(w http.ResponseWriter, req *http.Request) {
	w.WriteHeader(401)

	status := http.StatusText(401)
	_, err := io.WriteString(w, status)
	if err != nil {
		log.Println(err)
	}

	//if lastOrCreate("s") {
	//	io.WriteString(w, "hello, world!\n")
	//}
}

func main() {

	http.HandleFunc("/", helloServer)
	http.HandleFunc("/gettoken", createToken)
	log.Fatal(http.ListenAndServe(":8080", nil))
}
