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
					NewKetchup(DefaultPattern, "", NewRepository(10, Github, "")),
					NewKetchup(DefaultPattern, "", NewRepository(1, Github, "")),
					NewKetchup("latest", "", NewRepository(1, Github, "")),
				},
			},
			[]Ketchup{
				NewKetchup("latest", "", NewRepository(1, Github, "")),
				NewKetchup(DefaultPattern, "", NewRepository(1, Github, "")),
				NewKetchup(DefaultPattern, "", NewRepository(10, Github, "")),
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
					NewKetchup("", "", NewRepository(0, Github, "abc")),
					NewKetchup("", "", NewRepository(0, Github, "ghi")),
					NewKetchup("", "", NewRepository(0, Github, "jkl")),
					NewKetchup("", "", NewRepository(0, Github, "def")),
				},
			},
			[]Ketchup{
				NewKetchup("", "", NewRepository(0, Github, "abc")),
				NewKetchup("", "", NewRepository(0, Github, "def")),
				NewKetchup("", "", NewRepository(0, Github, "ghi")),
				NewKetchup("", "", NewRepository(0, Github, "jkl")),
			},
		},
		{
			"semver",
			args{
				array: []Ketchup{
					{Semver: "Minor", Repository: NewRepository(0, Github, "abc")},
					{Semver: "Major", Repository: NewRepository(0, Github, "ghi")},
					{Semver: "Patch", Repository: NewRepository(0, Github, "jkl")},
					{Semver: "", Repository: NewRepository(0, Github, "def")},
				},
			},
			[]Ketchup{
				{Semver: "Major", Repository: NewRepository(0, Github, "ghi")},
				{Semver: "Minor", Repository: NewRepository(0, Github, "abc")},
				{Semver: "Patch", Repository: NewRepository(0, Github, "jkl")},
				{Semver: "", Repository: NewRepository(0, Github, "def")},
			},
		},
		{
			"full",
			args{
				array: []Ketchup{
					{Semver: "Major", Repository: NewRepository(0, Github, "abc")},
					{Semver: "", Repository: NewRepository(0, Github, "abcd")},
					{Semver: "Patch", Repository: NewRepository(0, Github, "jkl")},
					{Semver: "", Repository: NewRepository(0, Github, "defg")},
					{Semver: "Patch", Repository: NewRepository(0, Github, "jjl")},
					{Semver: "Patch", Repository: NewRepository(0, Helm, "jjl")},
					{Semver: "Major", Repository: NewRepository(0, Github, "ghi")},
				},
			},
			[]Ketchup{
				{Semver: "Major", Repository: NewRepository(0, Github, "abc")},
				{Semver: "Major", Repository: NewRepository(0, Github, "ghi")},
				{Semver: "Patch", Repository: NewRepository(0, Github, "jjl")},
				{Semver: "Patch", Repository: NewRepository(0, Github, "jkl")},
				{Semver: "", Repository: NewRepository(0, Github, "abcd")},
				{Semver: "", Repository: NewRepository(0, Github, "defg")},
				{Semver: "Patch", Repository: NewRepository(0, Helm, "jjl")},
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
					NewRelease(NewRepository(10, Github, ""), DefaultPattern, semver.Version{}),
					NewRelease(NewRepository(1, Github, ""), DefaultPattern, semver.Version{}),
				},
			},
			[]Release{
				NewRelease(NewRepository(1, Github, ""), DefaultPattern, semver.Version{}),
				NewRelease(NewRepository(10, Github, ""), DefaultPattern, semver.Version{}),
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
