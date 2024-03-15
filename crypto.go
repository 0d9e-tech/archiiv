package main

import (
	"bytes"
	"crypto/ed25519"
	"crypto/rand"
	"encoding/base64"
	"encoding/gob"
	"errors"
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
	return base64.StdEncoding.EncodeToString(seed)
}

func secretToKeys(secretStr string) (priv ed25519.PrivateKey, err error) {
	secret, err := base64.StdEncoding.DecodeString(secretStr)
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
		return "", err
	}

	signature, err := priv.Sign(nil, payloadBytes, &ed25519.Options{})
	if err != nil {
		return "", err
	}

	tok := FullToken{
		data: payload,
		sign: signature,
	}

	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)
	enc.Encode(tok)

	return base64.StdEncoding.EncodeToString(buf.Bytes()), nil
}

func VerifySignature(dataStr, secret string, maxAge time.Duration) (string, error) {
	data, err := base64.StdEncoding.DecodeString(dataStr)
	if err != nil {
		return "", err
	}

	dec := gob.NewDecoder(bytes.NewReader(data))
	var ft FullToken
	if err := dec.Decode(&ft); err != nil {
		return "", err
	}

	priv, err := secretToKeys(secret)
	if err != nil {
		return "", err
	}

	payloadBytes := payloadToBytes(ft.data)

	if !ed25519.Verify(priv.Public().(ed25519.PublicKey), payloadBytes, ft.sign) {
		return "", errors.New("Signature is invalid")
	}

	if time.Since(ft.data.Timestamp).Microseconds() > maxAge.Microseconds() {
		return "", errors.New("Signature is too old")
	}

	return ft.data.Username, nil
}
