package ketchup

import "testing"

func TestMin(t *testing.T) {
	type args struct {
		a uint64
		b uint64
	}

	cases := map[string]struct {
		args args
		want uint64
	}{
		"a": {
			args{
				a: 1,
				b: 2,
			},
			1,
		},
		"b": {
			args{
				a: 3,
				b: 2,
			},
			2,
		},
	}

	for intention, tc := range cases {
		t.Run(intention, func(t *testing.T) {
			if got := min(tc.args.a, tc.args.b); got != tc.want {
				t.Errorf("min() = %d, want %d", got, tc.want)
			}
		})
	}
}
