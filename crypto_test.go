package main

import (
	"encoding/base64"
	"testing"
	"time"
)

func expectStringLooksLikeToken(t *testing.T, token string) {
	data, err := base64.URLEncoding.DecodeString(token)
	if err != nil {
		t.Errorf("string does not look like token: base64 decode: %v", err)
	}

	ft, err := gobDecode[FullToken](data)
	if err != nil {
		t.Errorf("string does not look like token: gob decode %v", err)
	}

	_, err = payloadToBytes(ft.Data)
	if err != nil {
		t.Errorf("string does not look like token: payload to bytes: %v", err)
	}
}

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
