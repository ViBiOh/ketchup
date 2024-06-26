package model

import (
	"context"
	"testing"
)

func TestReadUser(t *testing.T) {
	t.Parallel()

	type args struct {
		ctx context.Context
	}

	cases := map[string]struct {
		args args
		want User
	}{
		"empty": {
			args{
				ctx: context.TODO(),
			},
			User{},
		},
		"with User": {
			args{
				ctx: StoreUser(context.TODO(), User{ID: 8000, Email: "nobody@localhost"}),
			},
			User{ID: 8000, Email: "nobody@localhost"},
		},
		"not an User": {
			args{
				ctx: context.WithValue(context.TODO(), ctxUserKey, args{}),
			},
			User{},
		},
	}

	for intention, testCase := range cases {
		t.Run(intention, func(t *testing.T) {
			t.Parallel()

			if got := ReadUser(testCase.args.ctx); got != testCase.want {
				t.Errorf("ReadUser() = %v, want %v", got, testCase.want)
			}
		})
	}
}
