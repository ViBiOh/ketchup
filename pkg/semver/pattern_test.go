package semver

import (
	"fmt"
	"testing"
)

var (
	stableVersion = "1.0.0"
	betaVersion   = "1.0.0-beta1"
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
			"too short",
			safeParsePattern("1"),
			args{
				version: safeParse(stableVersion),
			},
			true,
		},
		{
			"no version",
			safeParsePattern("^latest"),
			args{
				version: safeParse(stableVersion),
			},
			true,
		},
		{
			"latest",
			safeParsePattern("latest"),
			args{
				version: safeParse(stableVersion),
			},
			true,
		},
		{
			"latest beta",
			safeParsePattern("latest"),
			args{
				version: safeParse(betaVersion),
			},
			true,
		},
		{
			"stable",
			safeParsePattern("stable"),
			args{
				version: safeParse(stableVersion),
			},
			true,
		},
		{
			"stable beta",
			safeParsePattern("stable"),
			args{
				version: safeParse(betaVersion),
			},
			false,
		},
		{
			"simple caret",
			safeParsePattern("^2"),
			args{
				version: safeParse(stableVersion),
			},
			false,
		},
		{
			"simple caret match",
			safeParsePattern("^1"),
			args{
				version: safeParse(stableVersion),
			},
			true,
		},
		{
			"simple caret match",
			safeParsePattern("^11"),
			args{
				version: safeParse("12"),
			},
			false,
		},
		{
			"caret",
			safeParsePattern("^1.0"),
			args{
				version: safeParse(stableVersion),
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
				version: safeParse(betaVersion),
			},
			false,
		},
		{
			"caret beta",
			safeParsePattern("^1-0"),
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
				version: safeParse(betaVersion),
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

	for _, tc := range cases {
		t.Run(tc.intention, func(t *testing.T) {
			if got := tc.instance.Check(tc.args.version); got != tc.want {
				t.Errorf("Check() = %t, want %t", got, tc.want)
			}
		})
	}
}

func safeParsePattern(pattern string) Pattern {
	output, err := ParsePattern(pattern)
	if err != nil {
		fmt.Println(err)
	}
	return output
}
