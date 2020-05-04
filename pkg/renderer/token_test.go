package renderer

import (
	"testing"
	"time"
)

func TestStore(t *testing.T) {
	type args struct {
		value    interface{}
		duration time.Duration
	}

	var cases = []struct {
		intention string
		args      args
		want      int
	}{
		{
			"nil",
			args{
				value:    nil,
				duration: 0,
			},
			36,
		},
		{
			"value",
			args{
				value:    "valid",
				duration: 0,
			},
			36,
		},
	}

	for _, tc := range cases {
		t.Run(tc.intention, func(t *testing.T) {
			if got := NewTokenStore().Store(tc.args.value, tc.args.duration); len(got) != tc.want {
				t.Errorf("Store() = %s, want %d", got, tc.want)
			}
		})
	}
}

func TestLoad(t *testing.T) {
	tokenStore := NewTokenStore()
	presentToken := tokenStore.Store("ok", time.Minute)
	invalidToken := tokenStore.Store("ko", time.Minute*-1)

	type args struct {
		key string
	}

	var cases = []struct {
		intention string
		args      args
		want      interface{}
		wantFound bool
	}{
		{
			"absent",
			args{
				key: "unknown",
			},
			nil,
			false,
		},
		{
			"present",
			args{
				key: presentToken,
			},
			"ok",
			true,
		},
		{
			"invalid",
			args{
				key: invalidToken,
			},
			nil,
			false,
		},
	}

	for _, tc := range cases {
		t.Run(tc.intention, func(t *testing.T) {
			if got, gotFound := tokenStore.Load(tc.args.key); got != tc.want || gotFound != tc.wantFound {
				t.Errorf("Load() = (%s, %t), want (%s, %t)", got, gotFound, tc.want, tc.wantFound)
			}
		})
	}
}

func TestClean(t *testing.T) {
	tokenStore := NewTokenStore()
	presentToken := tokenStore.Store("ok", time.Minute)
	invalidToken := tokenStore.Store("ko", time.Minute*-1)

	var cases = []struct {
		intention string
	}{
		{
			"simple",
		},
	}

	for _, tc := range cases {
		t.Run(tc.intention, func(t *testing.T) {
			_ = tokenStore.Clean(time.Now())

			if _, ok := tokenStore.Load(presentToken); !ok {
				t.Errorf("Clean() deleted `%s`, want kept", presentToken)
			}

			if _, ok := tokenStore.Load(invalidToken); ok {
				t.Errorf("Clean() deleted `%s`, want deleted", invalidToken)
			}
		})
	}
}
