package main

import (
	"net/http"
	"testing"
)

const (
	cycleAmount = 100000
	targetURL   = "http://localhost:8082/main"
)

func TestSuccess(t *testing.T) {
	t.Parallel()

	for i := 0; i < cycleAmount; i++ {
		resp, e := http.Get(targetURL)
		if e != nil {
			t.Error(e)
		}

		e = resp.Body.Close()
		if e != nil {
			t.Error(e)
		}
	}
}

func TestFail(t *testing.T) {
	t.Parallel()

	for i := 0; i < cycleAmount; i++ {
		resp, e := http.Get(targetURL)
		if e != nil {
			t.Error(e)
		}

		e = resp.Body.Close()
		if e != nil {
			t.Error(e)
		}
	}
}

func TestAnother(t *testing.T) {
	t.Parallel()

	for i := 0; i < cycleAmount; i++ {
		resp, e := http.Get(targetURL)
		if e != nil {
			t.Error(e)
		}

		e = resp.Body.Close()
		if e != nil {
			t.Error(e)
		}
	}
}
