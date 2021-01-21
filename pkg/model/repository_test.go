package model

import (
	"errors"
	"strings"
	"testing"
)

func TestString(t *testing.T) {
	var cases = []struct {
		intention string
		instance  RepositoryKind
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

func TestParseRepositoryKind(t *testing.T) {
	type args struct {
		value string
	}

	var cases = []struct {
		intention string
		args      args
		want      RepositoryKind
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
			errors.New("invalid value `wrong` for repository kind"),
		},
	}

	for _, tc := range cases {
		t.Run(tc.intention, func(t *testing.T) {
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
