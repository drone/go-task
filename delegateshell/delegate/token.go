package delegate

import (
	"encoding/hex"
	"time"

	"github.com/google/uuid"
	"gopkg.in/square/go-jose.v2"
	"gopkg.in/square/go-jose.v2/jwt"
)

// Token generates a token with the given expiry to interact with the Harness manager
func Token(audience, issuer, subject, secret string, expiry time.Duration) (string, error) {
	bytes, err := hex.DecodeString(secret)
	if err != nil {
		return "", err
	}

	enc, err := jose.NewEncrypter(
		jose.A128GCM,
		jose.Recipient{Algorithm: jose.DIRECT, Key: bytes},
		(&jose.EncrypterOptions{}).WithType("JWT"),
	)
	if err != nil {
		return "", err
	}

	cl := jwt.Claims{
		Subject:  subject,
		Issuer:   issuer,
		Audience: []string{audience},
		Expiry:   jwt.NewNumericDate(time.Now().Add(expiry)),
		IssuedAt: jwt.NewNumericDate(time.Now()),
		ID:       uuid.New().String(),
	}
	raw, err := jwt.Encrypted(enc).Claims(cl).CompactSerialize()
	if err != nil {
		return "", err
	}

	return raw, nil
}
