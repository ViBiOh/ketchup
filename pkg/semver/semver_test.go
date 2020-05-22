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
			Version{"", 2, 0, 0},
			args{
				other: Version{"", 1, 0, 0},
			},
			true,
		},
		{
			"minor",
			Version{"", 1, 1, 0},
			args{
				other: Version{"", 1, 0, 0},
			},
			true,
		},
		{
			"minor with major greater",
			Version{"", 1, 1, 0},
			args{
				other: Version{"", 2, 0, 0},
			},
			false,
		},
		{
			"patch",
			Version{"", 1, 0, 1},
			args{
				other: Version{"", 1, 0, 0},
			},
			true,
		},
		{
			"patch with major greater",
			Version{"", 1, 0, 1},
			args{
				other: Version{"", 2, 0, 0},
			},
			false,
		},
		{
			"patch with minor greater",
			Version{"", 1, 0, 1},
			args{
				other: Version{"", 1, 1, 0},
			},
			false,
		},
		{
			"equal",
			Version{"", 1, 0, 0},
			args{
				other: Version{"", 1, 0, 0},
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
			Version{"", 1, 0, 0},
			args{
				other: Version{"", 0, 0, 0},
			},
			"Major",
		},
		{
			"minor",
			Version{"", 1, 0, 0},
			args{
				other: Version{"", 1, 2, 0},
			},
			"Minor",
		},
		{
			"patch",
			Version{"", 1, 0, 1},
			args{
				other: Version{"", 1, 0, 0},
			},
			"Patch",
		},
		{
			"equal",
			Version{"", 1, 0, 0},
			args{
				other: Version{"", 1, 0, 0},
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
			"no a semver",
			args{
				version: "release.r60.1",
			},
			NoneVersion,
			errors.New("unable to parse version"),
		},
		{
			"no a semver",
			args{
				version: "v2.2.1.0-0.3.rc3",
			},
			NoneVersion,
			errors.New("unable to parse version"),
		},
		{
			"ignore rc or beta",
			args{
				version: "v2.27.0-rc1",
			},
			NoneVersion,
			errors.New("ignoring rc version"),
		},
		{
			"major and minor only",
			args{
				version: "v1.25+xyz",
			},
			Version{"v1.25+xyz", 1, 25, 0},
			nil,
		},
		{
			"full",
			args{
				version: "v1.2.3",
			},
			Version{"v1.2.3", 1, 2, 3},
			nil,
		},
		{
			"with sha1",
			args{
				version: "v1.2.3+abcdef123456",
			},
			Version{"v1.2.3+abcdef123456", 1, 2, 3},
			nil,
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
