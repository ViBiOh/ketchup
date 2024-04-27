package model

import (
	"reflect"
	"sort"
	"strings"
	"testing"

	"github.com/ViBiOh/ketchup/pkg/semver"
)

func TestParseKetchupFrequency(t *testing.T) {
	t.Parallel()

	type args struct {
		value string
	}

	cases := map[string]struct {
		args    args
		want    KetchupFrequency
		wantErr error
	}{
		"UpperCase": {
			args{
				value: "NONE",
			},
			None,
			nil,
		},
		"equal": {
			args{
				value: "Weekly",
			},
			Weekly,
			nil,
		},
		"not found": {
			args{
				value: "wrong",
			},
			Daily,
			ErrUnknownKetchupFrequency,
		},
	}

	for intention, testCase := range cases {
		t.Run(intention, func(t *testing.T) {
			t.Parallel()

			got, gotErr := ParseKetchupFrequency(testCase.args.value)

			failed := false

			if testCase.wantErr == nil && gotErr != nil {
				failed = true
			} else if testCase.wantErr != nil && gotErr == nil {
				failed = true
			} else if testCase.wantErr != nil && !strings.Contains(gotErr.Error(), testCase.wantErr.Error()) {
				failed = true
			} else if got != testCase.want {
				failed = true
			}

			if failed {
				t.Errorf("ParseRepositoryKind() = (`%s`, `%s`), want (`%s`, `%s`)", got, gotErr, testCase.want, testCase.wantErr)
			}
		})
	}
}

func TestKetchupByRepositoryID(t *testing.T) {
	t.Parallel()

	type args struct {
		array []Ketchup
	}

	cases := map[string]struct {
		args args
		want []Ketchup
	}{
		"simple": {
			args{
				array: []Ketchup{
					NewKetchup(DefaultPattern, "", Daily, false, NewGithubRepository(Identifier(10), "")),
					NewKetchup(DefaultPattern, "", Daily, false, NewGithubRepository(Identifier(1), "")),
					NewKetchup("latest", "", Daily, false, NewGithubRepository(Identifier(1), "")),
				},
			},
			[]Ketchup{
				NewKetchup("latest", "", Daily, false, NewGithubRepository(Identifier(1), "")),
				NewKetchup(DefaultPattern, "", Daily, false, NewGithubRepository(Identifier(1), "")),
				NewKetchup(DefaultPattern, "", Daily, false, NewGithubRepository(Identifier(10), "")),
			},
		},
	}

	for intention, testCase := range cases {
		t.Run(intention, func(t *testing.T) {
			t.Parallel()

			sort.Sort(KetchupByRepositoryIDAndPattern(testCase.args.array))
			if got := testCase.args.array; !reflect.DeepEqual(got, testCase.want) {
				t.Errorf("KetchupByRepositoryID() = %+v, want %+v", got, testCase.want)
			}
		})
	}
}

func TestKetchupByPriority(t *testing.T) {
	t.Parallel()

	type args struct {
		array []Ketchup
	}

	cases := map[string]struct {
		args args
		want []Ketchup
	}{
		"alphabetic": {
			args{
				array: []Ketchup{
					NewKetchup("", "", Daily, false, NewGithubRepository(Identifier(0), "abc")),
					NewKetchup("", "", Daily, false, NewGithubRepository(Identifier(0), "ghi")),
					NewKetchup("", "", Daily, false, NewGithubRepository(Identifier(0), "jkl")),
					NewKetchup("", "", Daily, false, NewGithubRepository(Identifier(0), "def")),
				},
			},
			[]Ketchup{
				NewKetchup("", "", Daily, false, NewGithubRepository(Identifier(0), "abc")),
				NewKetchup("", "", Daily, false, NewGithubRepository(Identifier(0), "def")),
				NewKetchup("", "", Daily, false, NewGithubRepository(Identifier(0), "ghi")),
				NewKetchup("", "", Daily, false, NewGithubRepository(Identifier(0), "jkl")),
			},
		},
		"semver": {
			args{
				array: []Ketchup{
					{Semver: "Minor", Repository: NewGithubRepository(Identifier(0), "abc")},
					{Semver: "Major", Repository: NewGithubRepository(Identifier(0), "ghi")},
					{Semver: "Patch", Repository: NewGithubRepository(Identifier(0), "jkl")},
					{Semver: "", Repository: NewGithubRepository(Identifier(0), "def")},
				},
			},
			[]Ketchup{
				{Semver: "Major", Repository: NewGithubRepository(Identifier(0), "ghi")},
				{Semver: "Minor", Repository: NewGithubRepository(Identifier(0), "abc")},
				{Semver: "Patch", Repository: NewGithubRepository(Identifier(0), "jkl")},
				{Semver: "", Repository: NewGithubRepository(Identifier(0), "def")},
			},
		},
		"full": {
			args{
				array: []Ketchup{
					{Semver: "Major", Repository: NewGithubRepository(Identifier(0), "abc")},
					{Semver: "", Repository: NewGithubRepository(Identifier(0), "abcd")},
					{Semver: "Patch", Repository: NewGithubRepository(Identifier(0), "jkl")},
					{Semver: "", Repository: NewGithubRepository(Identifier(0), "defg")},
					{Semver: "Patch", Repository: NewGithubRepository(Identifier(0), "jjl")},
					{Semver: "Patch", Repository: NewHelmRepository(Identifier(0), "jjl", "def")},
					{Semver: "Patch", Repository: NewHelmRepository(Identifier(0), "jjl", "abc")},
					{Semver: "Major", Repository: NewGithubRepository(Identifier(0), "ghi")},
				},
			},
			[]Ketchup{
				{Semver: "Major", Repository: NewGithubRepository(Identifier(0), "abc")},
				{Semver: "Major", Repository: NewGithubRepository(Identifier(0), "ghi")},
				{Semver: "Patch", Repository: NewGithubRepository(Identifier(0), "jjl")},
				{Semver: "Patch", Repository: NewGithubRepository(Identifier(0), "jkl")},
				{Semver: "", Repository: NewGithubRepository(Identifier(0), "abcd")},
				{Semver: "", Repository: NewGithubRepository(Identifier(0), "defg")},
				{Semver: "Patch", Repository: NewHelmRepository(Identifier(0), "jjl", "abc")},
				{Semver: "Patch", Repository: NewHelmRepository(Identifier(0), "jjl", "def")},
			},
		},
	}

	for intention, testCase := range cases {
		t.Run(intention, func(t *testing.T) {
			t.Parallel()

			sort.Sort(KetchupByPriority(testCase.args.array))
			if got := testCase.args.array; !reflect.DeepEqual(got, testCase.want) {
				t.Errorf("KetchupByPriority() = %+v, want %+v", got, testCase.want)
			}
		})
	}
}

func TestReleaseByRepositoryID(t *testing.T) {
	t.Parallel()

	type args struct {
		array []Release
	}

	cases := map[string]struct {
		args args
		want []Release
	}{
		"simple": {
			args{
				array: []Release{
					NewRelease(NewGithubRepository(Identifier(10), ""), DefaultPattern, semver.Version{}),
					NewRelease(NewGithubRepository(Identifier(1), "stable"), DefaultPattern, semver.Version{}),
					NewRelease(NewGithubRepository(Identifier(1), "~1.10"), DefaultPattern, semver.Version{}),
				},
			},
			[]Release{
				NewRelease(NewGithubRepository(Identifier(1), "stable"), DefaultPattern, semver.Version{}),
				NewRelease(NewGithubRepository(Identifier(1), "~1.10"), DefaultPattern, semver.Version{}),
				NewRelease(NewGithubRepository(Identifier(10), ""), DefaultPattern, semver.Version{}),
			},
		},
	}

	for intention, testCase := range cases {
		t.Run(intention, func(t *testing.T) {
			t.Parallel()

			sort.Sort(ReleaseByRepositoryIDAndPattern(testCase.args.array))
			if got := testCase.args.array; !reflect.DeepEqual(got, testCase.want) {
				t.Errorf("ReleaseByRepositoryID() = %+v, want %+v", got, testCase.want)
			}
		})
	}
}

func TestReleaseByKindAndName(t *testing.T) {
	t.Parallel()

	type args struct {
		array []Release
	}

	cases := map[string]struct {
		args args
		want []Release
	}{
		"simple": {
			args{
				array: []Release{
					NewRelease(NewHelmRepository(Identifier(3), "http://chart", "app"), DefaultPattern, semver.Version{}),
					NewRelease(NewGithubRepository(Identifier(1), "vibioh/github"), DefaultPattern, semver.Version{}),
					NewRelease(NewHelmRepository(Identifier(2), "http://chart", "cron"), DefaultPattern, semver.Version{}),
				},
			},
			[]Release{
				NewRelease(NewGithubRepository(Identifier(1), "vibioh/github"), DefaultPattern, semver.Version{}),
				NewRelease(NewHelmRepository(Identifier(3), "http://chart", "app"), DefaultPattern, semver.Version{}),
				NewRelease(NewHelmRepository(Identifier(2), "http://chart", "cron"), DefaultPattern, semver.Version{}),
			},
		},
	}

	for intention, testCase := range cases {
		t.Run(intention, func(t *testing.T) {
			t.Parallel()

			sort.Sort(ReleaseByKindAndName(testCase.args.array))
			if got := testCase.args.array; !reflect.DeepEqual(got, testCase.want) {
				t.Errorf("ReleaseByRepositoryID() = %+v, want %+v", got, testCase.want)
			}
		})
	}
}
