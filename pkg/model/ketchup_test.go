package model

import (
	"errors"
	"reflect"
	"sort"
	"strings"
	"testing"

	"github.com/ViBiOh/ketchup/pkg/semver"
)

func TestParseKetchupFrequency(t *testing.T) {
	type args struct {
		value string
	}

	cases := []struct {
		intention string
		args      args
		want      KetchupFrequency
		wantErr   error
	}{
		{
			"UpperCase",
			args{
				value: "NONE",
			},
			None,
			nil,
		},
		{
			"equal",
			args{
				value: "Weekly",
			},
			Weekly,
			nil,
		},
		{
			"not found",
			args{
				value: "wrong",
			},
			Daily,
			errors.New("invalid value `wrong` for ketchup frequency"),
		},
	}

	for _, tc := range cases {
		t.Run(tc.intention, func(t *testing.T) {
			got, gotErr := ParseKetchupFrequency(tc.args.value)

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
				t.Errorf("ParseRepositoryKind() = (`%s`, `%s`), want (`%s`, `%s`)", got, gotErr, tc.want, tc.wantErr)
			}
		})
	}
}

func TestKetchupByRepositoryID(t *testing.T) {
	type args struct {
		array []Ketchup
	}

	cases := []struct {
		intention string
		args      args
		want      []Ketchup
	}{
		{
			"simple",
			args{
				array: []Ketchup{
					NewKetchup(DefaultPattern, "", Daily, false, NewGithubRepository(10, "")),
					NewKetchup(DefaultPattern, "", Daily, false, NewGithubRepository(1, "")),
					NewKetchup("latest", "", Daily, false, NewGithubRepository(1, "")),
				},
			},
			[]Ketchup{
				NewKetchup("latest", "", Daily, false, NewGithubRepository(1, "")),
				NewKetchup(DefaultPattern, "", Daily, false, NewGithubRepository(1, "")),
				NewKetchup(DefaultPattern, "", Daily, false, NewGithubRepository(10, "")),
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.intention, func(t *testing.T) {
			sort.Sort(KetchupByRepositoryIDAndPattern(tc.args.array))
			if got := tc.args.array; !reflect.DeepEqual(got, tc.want) {
				t.Errorf("KetchupByRepositoryID() = %+v, want %+v", got, tc.want)
			}
		})
	}
}

func TestKetchupByPriority(t *testing.T) {
	type args struct {
		array []Ketchup
	}

	cases := []struct {
		intention string
		args      args
		want      []Ketchup
	}{
		{
			"alphabetic",
			args{
				array: []Ketchup{
					NewKetchup("", "", Daily, false, NewGithubRepository(0, "abc")),
					NewKetchup("", "", Daily, false, NewGithubRepository(0, "ghi")),
					NewKetchup("", "", Daily, false, NewGithubRepository(0, "jkl")),
					NewKetchup("", "", Daily, false, NewGithubRepository(0, "def")),
				},
			},
			[]Ketchup{
				NewKetchup("", "", Daily, false, NewGithubRepository(0, "abc")),
				NewKetchup("", "", Daily, false, NewGithubRepository(0, "def")),
				NewKetchup("", "", Daily, false, NewGithubRepository(0, "ghi")),
				NewKetchup("", "", Daily, false, NewGithubRepository(0, "jkl")),
			},
		},
		{
			"semver",
			args{
				array: []Ketchup{
					{Semver: "Minor", Repository: NewGithubRepository(0, "abc")},
					{Semver: "Major", Repository: NewGithubRepository(0, "ghi")},
					{Semver: "Patch", Repository: NewGithubRepository(0, "jkl")},
					{Semver: "", Repository: NewGithubRepository(0, "def")},
				},
			},
			[]Ketchup{
				{Semver: "Major", Repository: NewGithubRepository(0, "ghi")},
				{Semver: "Minor", Repository: NewGithubRepository(0, "abc")},
				{Semver: "Patch", Repository: NewGithubRepository(0, "jkl")},
				{Semver: "", Repository: NewGithubRepository(0, "def")},
			},
		},
		{
			"full",
			args{
				array: []Ketchup{
					{Semver: "Major", Repository: NewGithubRepository(0, "abc")},
					{Semver: "", Repository: NewGithubRepository(0, "abcd")},
					{Semver: "Patch", Repository: NewGithubRepository(0, "jkl")},
					{Semver: "", Repository: NewGithubRepository(0, "defg")},
					{Semver: "Patch", Repository: NewGithubRepository(0, "jjl")},
					{Semver: "Patch", Repository: NewHelmRepository(0, "jjl", "def")},
					{Semver: "Patch", Repository: NewHelmRepository(0, "jjl", "abc")},
					{Semver: "Major", Repository: NewGithubRepository(0, "ghi")},
				},
			},
			[]Ketchup{
				{Semver: "Major", Repository: NewGithubRepository(0, "abc")},
				{Semver: "Major", Repository: NewGithubRepository(0, "ghi")},
				{Semver: "Patch", Repository: NewGithubRepository(0, "jjl")},
				{Semver: "Patch", Repository: NewGithubRepository(0, "jkl")},
				{Semver: "", Repository: NewGithubRepository(0, "abcd")},
				{Semver: "", Repository: NewGithubRepository(0, "defg")},
				{Semver: "Patch", Repository: NewHelmRepository(0, "jjl", "abc")},
				{Semver: "Patch", Repository: NewHelmRepository(0, "jjl", "def")},
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.intention, func(t *testing.T) {
			sort.Sort(KetchupByPriority(tc.args.array))
			if got := tc.args.array; !reflect.DeepEqual(got, tc.want) {
				t.Errorf("KetchupByPriority() = %+v, want %+v", got, tc.want)
			}
		})
	}
}

func TestReleaseByRepositoryID(t *testing.T) {
	type args struct {
		array []Release
	}

	cases := []struct {
		intention string
		args      args
		want      []Release
	}{
		{
			"simple",
			args{
				array: []Release{
					NewRelease(NewGithubRepository(10, ""), DefaultPattern, semver.Version{}),
					NewRelease(NewGithubRepository(1, "stable"), DefaultPattern, semver.Version{}),
					NewRelease(NewGithubRepository(1, "~1.10"), DefaultPattern, semver.Version{}),
				},
			},
			[]Release{
				NewRelease(NewGithubRepository(1, "stable"), DefaultPattern, semver.Version{}),
				NewRelease(NewGithubRepository(1, "~1.10"), DefaultPattern, semver.Version{}),
				NewRelease(NewGithubRepository(10, ""), DefaultPattern, semver.Version{}),
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.intention, func(t *testing.T) {
			sort.Sort(ReleaseByRepositoryIDAndPattern(tc.args.array))
			if got := tc.args.array; !reflect.DeepEqual(got, tc.want) {
				t.Errorf("ReleaseByRepositoryID() = %+v, want %+v", got, tc.want)
			}
		})
	}
}

func TestReleaseByKindAndName(t *testing.T) {
	type args struct {
		array []Release
	}

	cases := []struct {
		intention string
		args      args
		want      []Release
	}{
		{
			"simple",
			args{
				array: []Release{
					NewRelease(NewHelmRepository(3, "http://chart", "app"), DefaultPattern, semver.Version{}),
					NewRelease(NewGithubRepository(1, "vibioh/github"), DefaultPattern, semver.Version{}),
					NewRelease(NewHelmRepository(2, "http://chart", "cron"), DefaultPattern, semver.Version{}),
				},
			},
			[]Release{
				NewRelease(NewGithubRepository(1, "vibioh/github"), DefaultPattern, semver.Version{}),
				NewRelease(NewHelmRepository(3, "http://chart", "app"), DefaultPattern, semver.Version{}),
				NewRelease(NewHelmRepository(2, "http://chart", "cron"), DefaultPattern, semver.Version{}),
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.intention, func(t *testing.T) {
			sort.Sort(ReleaseByKindAndName(tc.args.array))
			if got := tc.args.array; !reflect.DeepEqual(got, tc.want) {
				t.Errorf("ReleaseByRepositoryID() = %+v, want %+v", got, tc.want)
			}
		})
	}
}
