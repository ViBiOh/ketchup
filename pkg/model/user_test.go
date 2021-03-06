package model

import (
	"context"
	"testing"
)

func TestReadUser(t *testing.T) {
	type args struct {
		ctx context.Context
	}

	var cases = []struct {
		intention string
		args      args
		want      User
	}{
		{
			"empty",
			args{
				ctx: context.Background(),
			},
			NoneUser,
		},
		{
			"with User",
			args{
				ctx: StoreUser(context.Background(), User{ID: 8000, Email: "nobody@localhost"}),
			},
			User{ID: 8000, Email: "nobody@localhost"},
		},
		{
			"not an User",
			args{
				ctx: context.WithValue(context.Background(), ctxUserKey, args{}),
			},
			NoneUser,
		},
	}

	for _, tc := range cases {
		t.Run(tc.intention, func(t *testing.T) {
			if got := ReadUser(tc.args.ctx); got != tc.want {
				t.Errorf("ReadUser() = %v, want %v", got, tc.want)
			}
		})
	}
}
