package main

import "testing"

type tc struct {
	input    string
	expected string
}

var testCases = []tc{
	{"in",
		"out",
	},
}

func TestSuccess(t *testing.T) {
	if testCases[0].input != testCases[0].expected {
		t.Fatal("error occured")
	}
}
