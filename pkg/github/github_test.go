package github

import (
	"errors"
	"flag"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"
)

func TestFlags(t *testing.T) {
	var cases = []struct {
		intention string
		want      string
	}{
		{
			"simple",
			"Usage of simple:\n  -token string\n    \t[github] OAuth Token {SIMPLE_TOKEN}\n",
		},
	}

	for _, testCase := range cases {
		t.Run(testCase.intention, func(t *testing.T) {
			fs := flag.NewFlagSet(testCase.intention, flag.ContinueOnError)
			Flags(fs, "")

			var writer strings.Builder
			fs.SetOutput(&writer)
			fs.Usage()

			result := writer.String()

			if result != testCase.want {
				t.Errorf("Flags() = `%s`, want `%s`", result, testCase.want)
			}
		})
	}
}

func TestLastRelease(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.URL.Path, "unknown") {
			w.WriteHeader(http.StatusNotFound)
			return
		}

		if strings.Contains(r.URL.Path, "invalid") {
			w.Write([]byte("{invalid: json}"))
			return
		}

		w.Write([]byte(`{"repository": "vibioh/ketchup", "tag_name": "1.0.0", "body": "this is cool"}`))
	}))

	savedURL := apiURL
	apiURL = server.URL

	defer func() {
		apiURL = savedURL
		server.Close()
	}()

	token := "secret"

	type args struct {
		repository string
	}

	var cases = []struct {
		intention string
		args      args
		want      Release
		wantErr   error
	}{
		{
			"error",
			args{
				repository: "unknown",
			},
			Release{},
			errors.New("unable to get latest release for unknown"),
		},
		{
			"invalid payload",
			args{
				repository: "invalid",
			},
			Release{},
			errors.New("unable to parse release body for invalid"),
		},
		{
			"success",
			args{
				repository: "vibioh/ketchup",
			},
			Release{Repository: "vibioh/ketchup", TagName: "1.0.0", Body: "this is cool"},
			nil,
		},
	}

	for _, tc := range cases {
		t.Run(tc.intention, func(t *testing.T) {
			got, gotErr := New(Config{token: &token}).LastRelease(tc.args.repository)

			failed := false

			if tc.wantErr == nil && gotErr != nil {
				failed = true
			} else if tc.wantErr != nil && gotErr == nil {
				failed = true
			} else if tc.wantErr != nil && !strings.Contains(gotErr.Error(), tc.wantErr.Error()) {
				failed = true
			} else if !reflect.DeepEqual(got, tc.want) {
				failed = true
			}

			if failed {
				t.Errorf("LastRelease() = (%+v, `%s`), want (%+v, `%s`)", got, gotErr, tc.want, tc.wantErr)
			}
		})
	}
}
