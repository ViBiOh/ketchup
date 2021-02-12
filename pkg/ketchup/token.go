package ketchup

import (
	"crypto/rand"
	"fmt"
	"sync"
	"time"

	"github.com/ViBiOh/httputils/v3/pkg/logger"
)

// TokenStore stores single usage token
type TokenStore interface {
	Store(value interface{}, duration time.Duration) string
	Load(key string) (interface{}, bool)
	Delete(key string)
	Clean(currentTime time.Time) error
}

type tokenStore struct {
	store sync.Map
}

// NewTokenStore creates a new token store
func NewTokenStore() TokenStore {
	return &tokenStore{}
}

type mapValue struct {
	content    interface{}
	expiration time.Time
}

func (t *mapValue) isValid() bool {
	return t.expiration.After(time.Now())
}

func (t *tokenStore) Store(value interface{}, duration time.Duration) string {
	token := t.uuid()

	t.store.Store(token, mapValue{
		content:    value,
		expiration: time.Now().Add(duration),
	})

	return token
}

func (t *tokenStore) Load(key string) (interface{}, bool) {
	if value, ok := t.store.Load(key); ok {
		timeValue := value.(mapValue)
		if timeValue.isValid() {
			return timeValue.content, true
		}

		t.Delete(key)
	}

	return nil, false
}

func (t *tokenStore) Delete(key string) {
	t.store.Delete(key)
}

func (t *tokenStore) Clean(_ time.Time) error {
	t.store.Range(func(key, value interface{}) bool {
		timeValue := value.(mapValue)
		if !timeValue.isValid() {
			t.Delete(key.(string))
		}

		return true
	})

	return nil
}

func (t *tokenStore) uuid() string {
	raw := make([]byte, 16)
	if _, err := rand.Read(raw); err != nil {
		logger.Fatal(err)
		return ""
	}

	raw[8] = raw[8]&^0xc0 | 0x80
	raw[6] = raw[6]&^0xf0 | 0x40

	return fmt.Sprintf("%x-%x-%x-%x-%x", raw[0:4], raw[4:6], raw[6:8], raw[8:10], raw[10:])
}
