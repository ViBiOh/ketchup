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
					{Repository: Repository{ID: 10}},
					{Repository: Repository{ID: 1}},
				},
			},
			[]Ketchup{
				{Repository: Repository{ID: 1}},
				{Repository: Repository{ID: 10}},
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
					{Repository: Repository{Name: "abc"}},
					{Repository: Repository{Name: "ghi"}},
					{Repository: Repository{Name: "jkl"}},
					{Repository: Repository{Name: "def"}},
				},
			},
			[]Ketchup{
				{Repository: Repository{Name: "abc"}},
				{Repository: Repository{Name: "def"}},
				{Repository: Repository{Name: "ghi"}},
				{Repository: Repository{Name: "jkl"}},
			},
		},
		{
			"semver",
			args{
				array: []Ketchup{
					{Semver: "Minor", Repository: Repository{Name: "abc"}},
					{Semver: "Major", Repository: Repository{Name: "ghi"}},
					{Semver: "Patch", Repository: Repository{Name: "jkl"}},
					{Semver: "", Repository: Repository{Name: "def"}},
				},
			},
			[]Ketchup{
				{Semver: "Major", Repository: Repository{Name: "ghi"}},
				{Semver: "Minor", Repository: Repository{Name: "abc"}},
				{Semver: "Patch", Repository: Repository{Name: "jkl"}},
				{Semver: "", Repository: Repository{Name: "def"}},
			},
		},
		{
			"full",
			args{
				array: []Ketchup{
					{Semver: "Major", Repository: Repository{Name: "abc"}},
					{Semver: "", Repository: Repository{Name: "abcd"}},
					{Semver: "Patch", Repository: Repository{Name: "jkl"}},
					{Semver: "", Repository: Repository{Name: "defg"}},
					{Semver: "Patch", Repository: Repository{Name: "jjl"}},
					{Semver: "Patch", Repository: Repository{Name: "jjl", Kind: 1}},
					{Semver: "Major", Repository: Repository{Name: "ghi"}},
				},
			},
			[]Ketchup{
				{Semver: "Major", Repository: Repository{Name: "abc"}},
				{Semver: "Major", Repository: Repository{Name: "ghi"}},
				{Semver: "Patch", Repository: Repository{Name: "jjl"}},
				{Semver: "Patch", Repository: Repository{Name: "jkl"}},
				{Semver: "", Repository: Repository{Name: "abcd"}},
				{Semver: "", Repository: Repository{Name: "defg"}},
				{Semver: "Patch", Repository: Repository{Name: "jjl", Kind: 1}},
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
					NewRelease(Repository{ID: 10}, semver.Version{}),
					NewRelease(Repository{ID: 1}, semver.Version{}),
				},
			},
			[]Release{
				NewRelease(Repository{ID: 1}, semver.Version{}),
				NewRelease(Repository{ID: 10}, semver.Version{}),
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
