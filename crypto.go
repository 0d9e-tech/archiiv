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

type tokenPayload struct {
	Username  string
	Timestamp time.Time
	Nonce     int64
}

type fullToken struct {
	Data tokenPayload
	Sign []byte
}

func generateSecret() string {
	_, priv, err := ed25519.GenerateKey(nil)
	if err != nil {
		panic(err)
	}
	seed := priv.Seed()
	return base64.URLEncoding.EncodeToString(seed)
}

func hashPassword(pwd string) [64]byte {
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

func payloadToBytes(p tokenPayload) ([]byte, error) {
	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)
	if err := enc.Encode(p); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func gobEncode(v any) ([]byte, error) {
	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)
	if err := enc.Encode(v); err != nil {
		return nil, fmt.Errorf("gob decode: %w", err)
	}
	return buf.Bytes(), nil
}

func gobDecode[T any](d []byte) (T, error) {
	dec := gob.NewDecoder(bytes.NewReader(d))
	var v T
	if err := dec.Decode(&v); err != nil {
		return v, fmt.Errorf("gob decode: %w", err)
	}
	return v, nil
}

func sign(username, secret string) (string, error) {
	// we construct the payload, serialize the payload into []byte, sign
	// the []byte, construct (payload, signature), serialize (payload,
	// signature) into string and return it
	nonce, err := rand.Int(rand.Reader, big.NewInt(math.MaxInt64))
	if err != nil {
		// error here is sus. better take the thing down
		panic(err)
	}

	payload := tokenPayload{
		Username:  username,
		Timestamp: time.Now(),
		Nonce:     nonce.Int64(),
	}

	payloadBytes, err := payloadToBytes(payload)
	if err != nil {
		return "", fmt.Errorf("payload to bytes: %w", err)
	}

	priv, err := secretToKeys(secret)
	if err != nil {
		return "", fmt.Errorf("derive key from secret: %w", err)
	}

	signature, err := priv.Sign(nil, payloadBytes, &ed25519.Options{})
	if err != nil {
		return "", err
	}

	fullTokenBytes, err := gobEncode(fullToken{Data: payload, Sign: signature})
	if err != nil {
		return "", err
	}

	return base64.URLEncoding.EncodeToString(fullTokenBytes), nil
}

func verifySignature(dataStr, secret string, maxAge time.Duration) (string, error) {
	data, err := base64.URLEncoding.DecodeString(dataStr)
	if err != nil {
		return "", fmt.Errorf("base64 decode token: %w", err)
	}

	ft, err := gobDecode[fullToken](data)
	if err != nil {
		return "", fmt.Errorf("decode FullToken: %w", err)
	}

	priv, err := secretToKeys(secret)
	if err != nil {
		return "", fmt.Errorf("derive key from secret: %w", err)
	}

	payloadBytes, err := payloadToBytes(ft.Data)
	if err != nil {
		return "", fmt.Errorf("payload to bytes: %w", err)
	}

	if !ed25519.Verify(priv.Public().(ed25519.PublicKey), payloadBytes, ft.Sign) {
		return "", errors.New("signature is invalid")
	}

	if time.Since(ft.Data.Timestamp).Microseconds() > maxAge.Microseconds() {
		return "", errors.New("signature is too old")
	}

	return ft.Data.Username, nil
}
