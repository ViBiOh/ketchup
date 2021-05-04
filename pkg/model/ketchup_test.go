package model

import (
	"reflect"
	"sort"
	"testing"

	"github.com/ViBiOh/ketchup/pkg/semver"
)

func TestKetchupByRepositoryID(t *testing.T) {
	type args struct {
		array []Ketchup
	}

	var cases = []struct {
		intention string
		args      args
		want      []Ketchup
	}{
		{
			"simple",
			args{
				array: []Ketchup{
					NewKetchup(DefaultPattern, "", NewGithubRepository(10, "")),
					NewKetchup(DefaultPattern, "", NewGithubRepository(1, "")),
					NewKetchup("latest", "", NewGithubRepository(1, "")),
				},
			},
			[]Ketchup{
				NewKetchup("latest", "", NewGithubRepository(1, "")),
				NewKetchup(DefaultPattern, "", NewGithubRepository(1, "")),
				NewKetchup(DefaultPattern, "", NewGithubRepository(10, "")),
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.intention, func(t *testing.T) {
			sort.Sort(KetchupByRepositoryID(tc.args.array))
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

	var cases = []struct {
		intention string
		args      args
		want      []Ketchup
	}{
		{
			"alphabetic",
			args{
				array: []Ketchup{
					NewKetchup("", "", NewGithubRepository(0, "abc")),
					NewKetchup("", "", NewGithubRepository(0, "ghi")),
					NewKetchup("", "", NewGithubRepository(0, "jkl")),
					NewKetchup("", "", NewGithubRepository(0, "def")),
				},
			},
			[]Ketchup{
				NewKetchup("", "", NewGithubRepository(0, "abc")),
				NewKetchup("", "", NewGithubRepository(0, "def")),
				NewKetchup("", "", NewGithubRepository(0, "ghi")),
				NewKetchup("", "", NewGithubRepository(0, "jkl")),
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
				{Semver: "Patch", Repository: NewHelmRepository(0, "jjl", "def")},
				{Semver: "Patch", Repository: NewHelmRepository(0, "jjl", "abc")},
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

	var cases = []struct {
		intention string
		args      args
		want      []Release
	}{
		{
			"simple",
			args{
				array: []Release{
					NewRelease(NewGithubRepository(10, ""), DefaultPattern, semver.Version{}),
					NewRelease(NewGithubRepository(1, ""), DefaultPattern, semver.Version{}),
				},
			},
			[]Release{
				NewRelease(NewGithubRepository(1, ""), DefaultPattern, semver.Version{}),
				NewRelease(NewGithubRepository(10, ""), DefaultPattern, semver.Version{}),
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.intention, func(t *testing.T) {
			sort.Sort(ReleaseByRepositoryID(tc.args.array))
			if got := tc.args.array; !reflect.DeepEqual(got, tc.want) {
				t.Errorf("ReleaseByRepositoryID() = %+v, want %+v", got, tc.want)
			}
		})
	}
}
