package model

import (
	"errors"
	"strings"
	"testing"
)

func TestString(t *testing.T) {
	t.Parallel()

	cases := map[string]struct {
		instance RepositoryKind
		want     string
	}{
		"github": {
			Github,
			"github",
		},
		"helm": {
			Helm,
			"helm",
		},
	}

	for intention, testCase := range cases {
		t.Run(intention, func(t *testing.T) {
			t.Parallel()

			if got := testCase.instance.String(); got != testCase.want {
				t.Errorf("String() = `%s`, want `%s`", got, testCase.want)
			}
		})
	}
}

func TestURL(t *testing.T) {
	t.Parallel()

	type args struct {
		pattern string
	}

	cases := map[string]struct {
		instance Repository
		args     args
		want     string
	}{
		"helm": {
			NewHelmRepository(Identifier(0), "https://charts.vibioh.fr", "app"),
			args{
				pattern: DefaultPattern,
			},
			"https://charts.vibioh.fr",
		},
		"invalid": {
			NewHelmRepository(Identifier(0), "charts.fr", ""),
			args{
				pattern: DefaultPattern,
			},
			"charts.fr",
		},
		"github": {
			NewGithubRepository(Identifier(0), "vibioh/ketchup").AddVersion(DefaultPattern, "v1.0.0"),
			args{
				pattern: DefaultPattern,
			},
			"https://github.com/vibioh/ketchup/releases/tag/v1.0.0",
		},
	}

	for intention, testCase := range cases {
		t.Run(intention, func(t *testing.T) {
			t.Parallel()

			if got := testCase.instance.URL(testCase.args.pattern); got != testCase.want {
				t.Errorf("URL() = `%s`, want `%s`", got, testCase.want)
			}
		})
	}
}

func TestCompareURL(t *testing.T) {
	t.Parallel()

	type args struct {
		version string
		pattern string
	}

	cases := map[string]struct {
		instance Repository
		args     args
		want     string
	}{
		"helm": {
			NewHelmRepository(Identifier(0), "https://charts.vibioh.fr", "app"),
			args{},
			"https://charts.vibioh.fr",
		},
		"github": {
			NewGithubRepository(Identifier(0), "vibioh/ketchup").AddVersion(DefaultPattern, "v1.1.0"),
			args{
				version: "v1.0.0",
				pattern: DefaultPattern,
			},
			"https://github.com/vibioh/ketchup/compare/v1.0.0...v1.1.0",
		},
	}

	for intention, testCase := range cases {
		t.Run(intention, func(t *testing.T) {
			t.Parallel()

			if got := testCase.instance.CompareURL(testCase.args.version, testCase.args.pattern); got != testCase.want {
				t.Errorf("URL() = `%s`, want `%s`", got, testCase.want)
			}
		})
	}
}

func TestParseRepositoryKind(t *testing.T) {
	t.Parallel()

	type args struct {
		value string
	}

	cases := map[string]struct {
		args    args
		want    RepositoryKind
		wantErr error
	}{
		"UpperCase": {
			args{
				value: "HELM",
			},
			Helm,
			nil,
		},
		"not found": {
			args{
				value: "wrong",
			},
			Github,
			errors.New("invalid value `wrong` for repository kind"),
		},
	}

	for intention, testCase := range cases {
		t.Run(intention, func(t *testing.T) {
			t.Parallel()

			got, gotErr := ParseRepositoryKind(testCase.args.value)

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
