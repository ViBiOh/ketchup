package model

import (
	"errors"
	"strings"
	"testing"
)

func TestString(t *testing.T) {
	var cases = []struct {
		intention string
		instance  RepositoryType
		want      string
	}{
		{
			"github",
			Github,
			"github",
		},
		{
			"helm",
			Helm,
			"helm",
		},
	}

	for _, tc := range cases {
		t.Run(tc.intention, func(t *testing.T) {
			if got := tc.instance.String(); got != tc.want {
				t.Errorf("String() = `%s`, want `%s`", got, tc.want)
			}
		})
	}
}

func TestURL(t *testing.T) {
	var cases = []struct {
		intention string
		instance  Repository
		want      string
	}{
		{
			"helm",
			Repository{
				Type: Helm,
				Name: "app@https://charts.vibioh.fr",
			},
			"https://charts.vibioh.fr",
		},
		{
			"invalid",
			Repository{
				Type: Helm,
				Name: "charts.fr",
			},
			"charts.fr",
		},
		{
			"github",
			Repository{
				Type:    Github,
				Name:    "vibioh/ketchup",
				Version: "1.0.0",
			},
			"https://github.com/vibioh/ketchup/releases/tag/1.0.0",
		},
	}

	for _, tc := range cases {
		t.Run(tc.intention, func(t *testing.T) {
			if got := tc.instance.URL(); got != tc.want {
				t.Errorf("URL() = `%s`, want `%s`", got, tc.want)
			}
		})
	}
}

func TestCompareURL(t *testing.T) {
	type args struct {
		version string
	}

	var cases = []struct {
		intention string
		instance  Repository
		args      args
		want      string
	}{
		{
			"helm",
			Repository{
				Type: Helm,
				Name: "app@https://charts.vibioh.fr",
			},
			args{},
			"https://charts.vibioh.fr",
		},
		{
			"github",
			Repository{
				Type:    Github,
				Name:    "vibioh/ketchup",
				Version: "1.0.0",
			},
			args{
				version: "1.1.0",
			},
			"https://github.com/vibioh/ketchup/compare/1.1.0...1.0.0",
		},
	}

	for _, tc := range cases {
		t.Run(tc.intention, func(t *testing.T) {
			if got := tc.instance.CompareURL(tc.args.version); got != tc.want {
				t.Errorf("URL() = `%s`, want `%s`", got, tc.want)
			}
		})
	}
}

func TestParseRepositoryType(t *testing.T) {
	type args struct {
		value string
	}

	var cases = []struct {
		intention string
		args      args
		want      RepositoryType
		wantErr   error
	}{
		{
			"UpperCase",
			args{
				value: "HELM",
			},
			Helm,
			nil,
		},
		{
			"not found",
			args{
				value: "wrong",
			},
			Github,
			errors.New("invalid value `wrong` for repository type"),
		},
	}

	for _, tc := range cases {
		t.Run(tc.intention, func(t *testing.T) {
			got, gotErr := ParseRepositoryType(tc.args.value)

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
				t.Errorf("ParseRepositoryType() = (`%s`, `%s`), want (`%s`, `%s`)", got, gotErr, tc.want, tc.wantErr)
			}
		})
	}
}