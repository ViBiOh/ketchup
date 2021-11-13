package ketchup

import (
	"fmt"

	"github.com/ViBiOh/ketchup/pkg/model"
)

func suggestCacheKey(user model.User) string {
	if user.IsZero() {
		return "ketchup:suggests"
	}
	return fmt.Sprintf("ketchup:user:%d:suggests", user.ID)
}
