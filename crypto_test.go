package main

import (
	"testing"
	"time"
)

func TestSignVerify(t *testing.T) {
	secret := GenerateSecret()

	msg := "hello world"

	sign, err := Sign(msg, secret)
	if err != nil {
		t.Error(err)
	}

	msg2, err := VerifySignature(sign, secret, 10*time.Second)
	if err != nil {
		t.Error(err)
	}

	if msg != msg2 {
		t.Error("decoded string is not equal")
	}
}
