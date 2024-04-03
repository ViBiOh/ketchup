package ketchup

import "testing"

func TestMin(t *testing.T) {
	t.Parallel()

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

	for intention, testCase := range cases {
		t.Run(intention, func(t *testing.T) {
			t.Parallel()

			if got := min(testCase.args.a, testCase.args.b); got != testCase.want {
				t.Errorf("min() = %d, want %d", got, testCase.want)
			}
		})
	}
}
