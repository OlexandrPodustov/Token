package main

import (
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
	t.Parallel()
	for i := 0; i < cycleAmount; i++ {
		_, e := http.Get("http://localhost:8082/main")
		if e != nil {
			t.Fatal(e)
		}
	}
}

func TestFail(t *testing.T) {
	t.Parallel()
	for i := 0; i < cycleAmount; i++ {
		_, e := http.Get("http://localhost:8082/main")
		if e != nil {
			t.Fatal(e)
		}
	}
}

func TestAnother(t *testing.T) {
	t.Parallel()
	for i := 0; i < cycleAmount; i++ {
		_, e := http.Get("http://localhost:8082/main")
		if e != nil {
			t.Fatal(e)
		}
	}
}
