package main

import (
	"net/http"
	"testing"
)

const cycleAmount = 10000

type tc struct {
	input    string
	expected string
}

func TestSuccess(t *testing.T) {
	t.Parallel()
	for i := 0; i < cycleAmount; i++ {
		_, e := http.Get("http://localhost:8082/main")
		if e != nil {
			t.Error(e)
		}
	}
}

func TestFail(t *testing.T) {
	t.Parallel()
	for i := 0; i < cycleAmount; i++ {
		_, e := http.Get("http://localhost:8082/main")
		if e != nil {
			t.Error(e)
		}
	}
}

func TestAnother(t *testing.T) {
	t.Parallel()
	for i := 0; i < cycleAmount; i++ {
		_, e := http.Get("http://localhost:8082/main")
		if e != nil {
			t.Error(e)
		}
	}
}
