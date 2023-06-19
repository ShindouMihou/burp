package main

import (
	"burp/burper"
	"errors"
	"testing"
)

func TestParse(t *testing.T) {
	calls, err := burper.Parse([]byte("REDIS_PASSWORD=\"[burp: Random(6) AS redis_pass][burp: Use(redis_pass)]\""))
	if err != nil {
		t.Fatal(err)
	}
	if calls[0].Function != "random" {
		t.Fatal(errors.New("function is not being parsed, expected calls[0].Function to be \"random\""))
	}
	if len(calls[0].Args) != 1 {
		t.Fatal(errors.New("function is not being parsed, expected len(calls[0].Args) to be 1"))
	}
	if calls[0].Args[0] != "6" {
		t.Fatal(errors.New("function is not being parsed, expected calls[0].Args[0] to be \"6\""))
	}
	if calls[1].Function != "use" {
		t.Fatal(errors.New("function is not being parsed, expected calls[1].Function to be \"use\""))
	}
	if len(calls[1].Args) != 1 {
		t.Fatal(errors.New("function is not being parsed, expected len(calls[1].Args) to be 1"))
	}
	if calls[1].Args[0] != "redis_pass" {
		t.Fatal(errors.New("function is not being parsed, expected calls[1].Args[0] to be \"redis_pass\""))
	}
}

func BenchmarkParse(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_, err := burper.Parse([]byte("REDIS_PASSWORD=\"[burp: Random(6) AS redis_pass][burp: Use(redis_pass)]\""))
		if err != nil {
			b.Fatal(err)
		}
	}
}
