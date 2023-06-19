package main

import (
	"burp/burper"
	"burp/burper/functions"
	"testing"
)

func TestTree_Process(t *testing.T) {
	functions.RegisterFunctions()
	if _, err := burper.FromString("REDIS_PASSWORD=\"[burp: Random(6) AS redis_pass][burp: Use(redis_pass)]\""); err != nil {
		t.Fatal(err)
	}
}

func BenchmarkTree_Process(b *testing.B) {
	functions.RegisterFunctions()
	for i := 0; i < b.N; i++ {
		if _, err := burper.FromString("REDIS_PASSWORD=\"[burp: Random(6) AS redis_pass][burp: Use(redis_pass)]\""); err != nil {
			b.Fatal(err)
		}
	}
}
