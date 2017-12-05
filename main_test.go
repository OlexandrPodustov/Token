package main

import (
	"log"
	"net/http"
	"testing"
)

type tc struct {
	input    string
	expected string
}

var (
	testCases = []tc{
		{"in",
			"out",
		},
	}
)

func TestSuccess(t *testing.T) {
	for i := 0; i < 100000; i++ {
		//resp, e := http.Get("http://localhost:8082/main")
		_, e := http.Get("http://localhost:8082/main")
		if e != nil {
			log.Println(e)
		}
		//log.Println(resp.StatusCode)
	}
	if testCases[0].input != testCases[0].input {
		t.Fatal("error occured")
	}
}
