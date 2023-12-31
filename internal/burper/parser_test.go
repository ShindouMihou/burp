package burper

import (
	"errors"
	"testing"
)

func TestParse(t *testing.T) {
	calls, err := parse([]byte("REDIS_PASSWORD=\"[burp.Random(6) AS redis_pass][burp.redis_pass]\""))
	if err != nil {
		t.Fatal(err)
	}
	if calls[0].Identifier != "random" {
		t.Fatal(errors.New("function is not being parsed, expected calls[0].Identifier to be \"random\""))
	}
	if len(calls[0].Args) != 1 {
		t.Fatal(errors.New("function is not being parsed, expected len(calls[0].Args) to be 1"))
	}
	if calls[0].Args[0] != "6" {
		t.Fatal(errors.New("function is not being parsed, expected calls[0].Args[0] to be \"6\""))
	}
	if calls[1].Identifier != "use" {
		t.Fatal(errors.New("function is not being parsed, expected calls[1].Identifier to be \"use\""))
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
		_, err := parse([]byte("REDIS_PASSWORD=\"[burp.Random(6) AS redis_pass][burp.redis_pass]\""))
		if err != nil {
			b.Fatal(err)
		}
	}
}
