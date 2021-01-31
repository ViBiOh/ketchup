package semver

import (
	"fmt"
	"testing"
)

func TestCheck(t *testing.T) {
	type args struct {
		version Version
	}

	var cases = []struct {
		intention string
		instance  Pattern
		args      args
		want      bool
	}{
		{
			"latest",
			safeParsePattern("latest"),
			args{
				version: safeParse("1.0.0"),
			},
			true,
		},
		{
			"latest beta",
			safeParsePattern("latest"),
			args{
				version: safeParse("1.0.0-beta1"),
			},
			true,
		},
		{
			"stable",
			safeParsePattern("stable"),
			args{
				version: safeParse("1.0.0"),
			},
			true,
		},
		{
			"stable beta",
			safeParsePattern("stable"),
			args{
				version: safeParse("1.0.0-beta1"),
			},
			false,
		},
		{
			"caret",
			safeParsePattern("^1.0"),
			args{
				version: safeParse("1.0.0"),
			},
			true,
		},
		{
			"caret minor change",
			safeParsePattern("^1.0"),
			args{
				version: safeParse("1.1.0"),
			},
			true,
		},
		{
			"caret lower major",
			safeParsePattern("^1.0"),
			args{
				version: safeParse("0.1.0"),
			},
			false,
		},
		{
			"caret greater major",
			safeParsePattern("^1.0"),
			args{
				version: safeParse("2.0.0"),
			},
			false,
		},
		{
			"caret no beta",
			safeParsePattern("^1.0"),
			args{
				version: safeParse("1.0.0-beta1"),
			},
			false,
		},
		{
			"caret beta",
			safeParsePattern("^1.1.0-0"),
			args{
				version: safeParse("1.1.0-beta1"),
			},
			true,
		},
		{
			"tilde",
			safeParsePattern("~1.1"),
			args{
				version: safeParse("1.1.1"),
			},
			true,
		},
		{
			"tilde major change",
			safeParsePattern("~1.1"),
			args{
				version: safeParse("2.1.1"),
			},
			false,
		},
		{
			"tilde minor change",
			safeParsePattern("~1.1"),
			args{
				version: safeParse("1.2.1"),
			},
			false,
		},
		{
			"tilde no beta",
			safeParsePattern("~1.0"),
			args{
				version: safeParse("1.0.0-beta1"),
			},
			false,
		},
		{
			"tilde beta",
			safeParsePattern("~1.1.0-0"),
			args{
				version: safeParse("1.1.0-beta1"),
			},
			true,
		},
	}

	fmt.Println(safeParsePattern("^1.0").Check(safeParse("2.0.0")))

	for _, tc := range cases {
		t.Run(tc.intention, func(t *testing.T) {
			if got := tc.instance.Check(tc.args.version); got != tc.want {
				t.Errorf("Check() = %t, want %t", got, tc.want)
			}
		})
	}
}