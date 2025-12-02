package watoken

import (
	"crypto/ed25519"
	"encoding/hex"
	"errors"
	"time"

	"aidanwoods.dev/go-paseto"
)

// Payload untuk token JWT-like
type Payload struct {
	Id   string `json:"id"`
	Name string `json:"name"`
	Exp  string `json:"exp"`
}

// EncodeforHours mengenkode token dengan durasi dalam jam
func EncodeforHours(id, name, privateKeyHex string, hours int) (string, error) {
	privateKeyBytes, err := hex.DecodeString(privateKeyHex)
	if err != nil {
		return "", err
	}

	if len(privateKeyBytes) != ed25519.PrivateKeySize {
		return "", errors.New("invalid private key size")
	}

	key := paseto.NewV4AsymmetricSecretKeyFromBytes(privateKeyBytes)
	token := paseto.NewToken()
	token.SetIssuedAt(time.Now())
	token.SetNotBefore(time.Now())
	token.SetExpiration(time.Now().Add(time.Duration(hours) * time.Hour))
	token.SetString("id", id)
	token.SetString("name", name)

	return token.V4Sign(key, nil), nil
}

// Decode mendekode token dengan public key
func Decode(publicKeyHex, tokenString string) (*Payload, error) {
	publicKeyBytes, err := hex.DecodeString(publicKeyHex)
	if err != nil {
		return nil, err
	}

	if len(publicKeyBytes) != ed25519.PublicKeySize {
		return nil, errors.New("invalid public key size")
	}

	key := paseto.NewV4AsymmetricPublicKeyFromBytes(publicKeyBytes)
	parser := paseto.NewParser()

	token, err := parser.ParseV4Public(key, tokenString, nil)
	if err != nil {
		return nil, err
	}

	id, err := token.GetString("id")
	if err != nil {
		return nil, err
	}

	name, err := token.GetString("name")
	if err != nil {
		return nil, err
	}

	exp, err := token.GetExpiration()
	if err != nil {
		return nil, err
	}

	return &Payload{
		Id:   id,
		Name: name,
		Exp:  exp.Format(time.RFC3339),
	}, nil
}

// GenerateKey menghasilkan key pair baru
func GenerateKey() (privateKeyHex, publicKeyHex string) {
	key := paseto.NewV4AsymmetricSecretKey()
	privateKeyBytes := key.ExportBytes()
	publicKeyBytes := key.Public().ExportBytes()

	return hex.EncodeToString(privateKeyBytes), hex.EncodeToString(publicKeyBytes)
}