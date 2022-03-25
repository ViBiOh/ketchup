package model

import (
	"errors"
	"strings"
	"testing"
)

func TestString(t *testing.T) {
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

	for intention, tc := range cases {
		t.Run(intention, func(t *testing.T) {
			if got := tc.instance.String(); got != tc.want {
				t.Errorf("String() = `%s`, want `%s`", got, tc.want)
			}
		})
	}
}

func TestURL(t *testing.T) {
	type args struct {
		pattern string
	}

	cases := map[string]struct {
		instance Repository
		args     args
		want     string
	}{
		"helm": {
			NewHelmRepository(0, "https://charts.vibioh.fr", "app"),
			args{
				pattern: DefaultPattern,
			},
			"https://charts.vibioh.fr",
		},
		"invalid": {
			NewHelmRepository(0, "charts.fr", ""),
			args{
				pattern: DefaultPattern,
			},
			"charts.fr",
		},
		"github": {
			NewGithubRepository(0, "vibioh/ketchup").AddVersion(DefaultPattern, "v1.0.0"),
			args{
				pattern: DefaultPattern,
			},
			"https://github.com/vibioh/ketchup/releases/tag/v1.0.0",
		},
	}

	for intention, tc := range cases {
		t.Run(intention, func(t *testing.T) {
			if got := tc.instance.URL(tc.args.pattern); got != tc.want {
				t.Errorf("URL() = `%s`, want `%s`", got, tc.want)
			}
		})
	}
}

func TestCompareURL(t *testing.T) {
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
			NewHelmRepository(0, "https://charts.vibioh.fr", "app"),
			args{},
			"https://charts.vibioh.fr",
		},
		"github": {
			NewGithubRepository(0, "vibioh/ketchup").AddVersion(DefaultPattern, "v1.1.0"),
			args{
				version: "v1.0.0",
				pattern: DefaultPattern,
			},
			"https://github.com/vibioh/ketchup/compare/v1.0.0...v1.1.0",
		},
	}

	for intention, tc := range cases {
		t.Run(intention, func(t *testing.T) {
			if got := tc.instance.CompareURL(tc.args.version, tc.args.pattern); got != tc.want {
				t.Errorf("URL() = `%s`, want `%s`", got, tc.want)
			}
		})
	}
}

func TestParseRepositoryKind(t *testing.T) {
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

	for intention, tc := range cases {
		t.Run(intention, func(t *testing.T) {
			got, gotErr := ParseRepositoryKind(tc.args.value)

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
