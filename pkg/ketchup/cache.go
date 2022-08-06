package ketchup

import (
	"fmt"

	"github.com/ViBiOh/httputils/v4/pkg/sha"
	"github.com/ViBiOh/ketchup/pkg/model"
)

var cacheVersion = sha.New("vibioh/ketchup/1")[:8]

func cachePrefix() string {
	return "ketchup:" + cacheVersion
}

func suggestCacheKey(user model.User) string {
	if user.IsZero() {
		return cachePrefix() + ":suggests"
	}
	return fmt.Sprintf("%s:user:%d:suggests", cachePrefix(), user.ID)
}
