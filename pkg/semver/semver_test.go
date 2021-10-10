package semver

import (
	"errors"
	"strings"
	"testing"
)

func TestIsGreater(t *testing.T) {
	type args struct {
		other Version
	}

	var cases = []struct {
		intention string
		instance  Version
		args      args
		want      bool
	}{
		{
			"major",
			safeParse("2.0.0"),
			args{
				other: safeParse("1.0.0"),
			},
			true,
		},
		{
			"minor",
			safeParse("1.1.0"),
			args{
				other: safeParse("1.0.0"),
			},
			true,
		},
		{
			"patch",
			safeParse("1.1.1"),
			args{
				other: safeParse("1.1.0"),
			},
			true,
		},
		{
			"minor with major greater",
			safeParse("2.0.0"),
			args{
				other: safeParse("1.2.0"),
			},
			true,
		},
		{
			"patch with major greater",
			Version{"", 2, 0, 1, 0},
			args{
				other: Version{"", 1, 0, 2, 0},
			},
			true,
		},
		{
			"patch with minor greater",
			Version{"", 1, 2, 1, 0},
			args{
				other: Version{"", 1, 1, 2, 0},
			},
			true,
		},
		{
			"patch with suffix greater",
			Version{"", 1, 1, 1, canary},
			args{
				other: Version{"", 1, 1, 1, beta},
			},
			true,
		},
		{
			"suffix presence",
			safeParse("1.1.1"),
			args{
				other: safeParse("1.1.1-beta1"),
			},
			true,
		},
		{
			"equal",
			safeParse("1.0.0"),
			args{
				other: safeParse("1.0.0"),
			},
			false,
		},
	}

	for _, tc := range cases {
		t.Run(tc.intention, func(t *testing.T) {
			if got := tc.instance.IsGreater(tc.args.other); got != tc.want {
				t.Errorf("IsGreater() = %t, want %t", got, tc.want)
			}
		})
	}
}

func TestCompare(t *testing.T) {
	type args struct {
		other Version
	}

	var cases = []struct {
		intention string
		instance  Version
		args      args
		want      string
	}{
		{
			"major",
			safeParse("1.0.0"),
			args{
				other: Version{"", 0, 0, 0, 0},
			},
			"Major",
		},
		{
			"minor",
			safeParse("1.0.0"),
			args{
				other: safeParse("1.2.0"),
			},
			"Minor",
		},
		{
			"patch",
			Version{"", 1, 0, 1, 0},
			args{
				other: safeParse("1.0.0"),
			},
			"Patch",
		},
		{
			"suffix",
			Version{"", 1, 0, 1, alpha},
			args{
				other: Version{"", 1, 0, 1, beta},
			},
			"Suffix",
		},
		{
			"equal",
			safeParse("1.0.0"),
			args{
				other: safeParse("1.0.0"),
			},
			"",
		},
	}

	for _, tc := range cases {
		t.Run(tc.intention, func(t *testing.T) {
			if got := tc.instance.Compare(tc.args.other); got != tc.want {
				t.Errorf("Compare() = `%s`, want `%s`", got, tc.want)
			}
		})
	}
}

func TestParse(t *testing.T) {
	type args struct {
		version string
	}

	var cases = []struct {
		intention string
		args      args
		want      Version
		wantErr   error
	}{
		{
			"not a semver",
			args{
				version: "release.r60.1",
			},
			Version{},
			errors.New("unable to parse version"),
		},
		{
			"not a semver too many",
			args{
				version: "v2.2.1.0-0.3.rc3",
			},
			Version{},
			errors.New("unable to parse version"),
		},
		{
			"prefixed version",
			args{
				version: "stable-2.10.1",
			},
			Version{},
			errors.New("unable to parse version"),
		},
		{
			"flag rc version",
			args{
				version: "v2.27.0-rc1",
			},
			Version{"v2.27.0-rc1", 2, 27, 0, rc},
			nil,
		},
		{
			"ignore test",
			args{
				version: "1.26.0-test",
			},
			Version{"1.26.0-test", 1, 26, 0, test},
			nil,
		},
		{
			"ignore canary",
			args{
				version: "v10.0.4-canary.1",
			},
			Version{"v10.0.4-canary.1", 10, 0, 4, canary},
			nil,
		},
		{
			"ignore alpha",
			args{
				version: "v0.14.0-alpha20200910",
			},
			Version{"v0.14.0-alpha20200910", 0, 14, 0, alpha},
			nil,
		},
		{
			"major and minor only",
			args{
				version: "v1.25",
			},
			Version{"v1.25", 1, 25, 0, -1},
			nil,
		},
		{
			"major and minor only with release",
			args{
				version: "v1.25-xyz",
			},
			Version{"v1.25-xyz", 1, 25, 0, 0},
			nil,
		},
		{
			"major and minor only with build",
			args{
				version: "v1.25+xyz",
			},
			Version{"v1.25+xyz", 1, 25, 0, -1},
			nil,
		},
		{
			"full",
			args{
				version: "v1.2.3",
			},
			Version{"v1.2.3", 1, 2, 3, -1},
			nil,
		},
		{
			"with sha1",
			args{
				version: "v1.2.3-abcdef123456",
			},
			Version{"v1.2.3-abcdef123456", 1, 2, 3, 0},
			nil,
		},
		{
			"fucking date",
			args{
				version: "v20160726",
			},
			Version{},
			errors.New("version major looks like a date"),
		},
		{
			"fucking node",
			args{
				version: "213123",
			},
			Version{},
			errors.New("version major seems a bit high"),
		},
	}

	for _, tc := range cases {
		t.Run(tc.intention, func(t *testing.T) {
			got, gotErr := Parse(tc.args.version)

			failed := false

			if tc.wantErr == nil && gotErr != nil {
				failed = true
			} else if tc.wantErr != nil && gotErr == nil {
				failed = true
			} else if tc.wantErr != nil && !strings.Contains(gotErr.Error(), tc.wantErr.Error()) {
				failed = true
			} else if got != tc.want {
				failed = true
			}

			if failed {
				t.Errorf("Parse() = (%+v, `%s`), want (%+v, `%s`)", got, gotErr, tc.want, tc.wantErr)
			}
		})
	}
}
