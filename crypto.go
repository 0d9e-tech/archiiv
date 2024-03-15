package main

import (
	"bytes"
	"crypto/ed25519"
	"crypto/rand"
	"crypto/sha512"
	"encoding/base64"
	"encoding/gob"
	"errors"
	"fmt"
	"math"
	"math/big"
	"time"
)

type TokenPayload struct {
	Username  string
	Timestamp time.Time
	Nonce     int64
}

type FullToken struct {
	data TokenPayload
	sign []byte
}

func GenerateSecret() string {
	_, priv, err := ed25519.GenerateKey(nil)
	if err != nil {
		panic(err)
	}
	seed := priv.Seed()
	return base64.URLEncoding.EncodeToString(seed)
}

func HashPassword(pwd string) [64]byte {
	return sha512.Sum512([]byte(pwd))
}

func secretToKeys(secretStr string) (priv ed25519.PrivateKey, err error) {
	secret, err := base64.URLEncoding.DecodeString(secretStr)
	if err != nil {
		return nil, err
	}
	priv = ed25519.NewKeyFromSeed(secret)
	return
}

func payloadToBytes(p TokenPayload) []byte {
	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)
	enc.Encode(p)
	return buf.Bytes()
}

func gobEncode(v any) []byte {
	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)
	enc.Encode(v)
	return buf.Bytes()
}

func gobDecode[T any](d []byte) (T, error) {
	dec := gob.NewDecoder(bytes.NewReader(d))
	var v T
	if err := dec.Decode(&v); err != nil {
		return v, fmt.Errorf("gob decode: %v", err)
	}
	return v, nil
}

func Sign(username, secret string) (string, error) {
	// we construct the payload, serialize the payload into []byte, sign
	// the []byte, construct (payload, signature), serialize (payload,
	// signature) into string and return it
	nonce, err := rand.Int(rand.Reader, big.NewInt(math.MaxInt64))
	if err != nil {
		// error here is sus. better take the thing down
		panic(err)
	}

	payload := TokenPayload{
		Username:  username,
		Timestamp: time.Now(),
		Nonce:     nonce.Int64(),
	}

	payloadBytes := payloadToBytes(payload)

	priv, err := secretToKeys(secret)
	if err != nil {
		return "", fmt.Errorf("derive key from secret: %v", err)
	}

	signature, err := priv.Sign(nil, payloadBytes, &ed25519.Options{})
	if err != nil {
		return "", err
	}

	fullTokenBytes := gobEncode(FullToken{
		data: payload,
		sign: signature,
	})

	return base64.URLEncoding.EncodeToString(fullTokenBytes), nil
}

func VerifySignature(dataStr, secret string, maxAge time.Duration) (string, error) {
	data, err := base64.URLEncoding.DecodeString(dataStr)
	if err != nil {
		return "", fmt.Errorf("base64 decode token: %v", err)
	}

	ft, err := gobDecode[FullToken](data)
	if err != nil {
		return "", fmt.Errorf("decode FullToken: %v", err)
	}

	priv, err := secretToKeys(secret)
	if err != nil {
		return "", fmt.Errorf("derive key from secret: %v", err)
	}

	payloadBytes := payloadToBytes(ft.data)

	if !ed25519.Verify(priv.Public().(ed25519.PublicKey), payloadBytes, ft.sign) {
		return "", errors.New("signature is invalid")
	}

	if time.Since(ft.data.Timestamp).Microseconds() > maxAge.Microseconds() {
		return "", errors.New("signature is too old")
	}

	return ft.data.Username, nil
}
