package delegate

import (
	"time"

	"github.com/patrickmn/go-cache"
	"github.com/sirupsen/logrus"
)

var (
	audience       = "audience"
	issuer         = "issuer"
	expirationTime = 20 * time.Minute
)

type TokenCache struct {
	id     string
	secret string
	expiry time.Duration
	c      *cache.Cache
}

// NewTokenCache creates a token cache which creates a new token
// after the expiry time is over
func NewTokenCache(id, secret string) *TokenCache {
	// purge expired tokens from the cache at expirationTime/3 intervals
	c := cache.New(cache.DefaultExpiration, expirationTime/3)
	return &TokenCache{
		id:     id,
		secret: secret,
		expiry: expirationTime,
		c:      c,
	}
}

// Get returns the value of the account token.
// If the token is cached, it returns from there. Otherwise
// it creates a new token with a new expiration time.
func (t *TokenCache) Get() (string, error) {
	tv, found := t.c.Get(t.id)
	if found {
		return tv.(string), nil
	}
	logrus.WithField("id", t.id).Infoln("refreshing token")
	token, err := Token(audience, issuer, t.id, t.secret, t.expiry)
	if err != nil {
		return "", err
	}
	// refresh token before the expiration time to give some buffer
	t.c.Set(t.id, token, t.expiry/2)
	return token, nil
}
