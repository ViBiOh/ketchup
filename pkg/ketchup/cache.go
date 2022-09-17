package ketchup

import (
	"fmt"

	"github.com/ViBiOh/httputils/v4/pkg/sha"
	"github.com/ViBiOh/ketchup/pkg/model"
)

var (
	cacheVersion     = sha.New("vibioh/ketchup/1")[:8]
	cachePrefix      = "ketchup:" + cacheVersion
	cacheSuggestsKey = cachePrefix + ":suggests"
)

func suggestCacheKey(user model.User) string {
	if user.IsZero() {
		return cacheSuggestsKey
	}
	return fmt.Sprintf("%s:user:%d:suggests", cachePrefix, user.ID)
}
