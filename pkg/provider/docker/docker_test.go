package docker

import (
	"net/http"
	"testing"
)

func TestGetNextURL(t *testing.T) {
	t.Parallel()

	type args struct {
		headers  http.Header
		registry string
	}

	upperCaseHeader := http.Header{}
	upperCaseHeader.Add("Link", `rel="next"; /v2/test`)

	lowerCaseHeader := http.Header{}
	lowerCaseHeader.Add("link", `rel="prev"; /v2/prev`)
	lowerCaseHeader.Add("link", `rel="next"; /v2/next`)

	cases := map[string]struct {
		args args
		want string
	}{
		"empty": {
			args{
				headers:  http.Header{},
				registry: "http://127.0.0.1",
			},
			"",
		},
		"uppercase link with next": {
			args{
				headers:  upperCaseHeader,
				registry: "http://127.0.0.1",
			},
			"http://127.0.0.1/v2/test",
		},
		"lowercase link with previous and next inverted": {
			args{
				headers:  lowerCaseHeader,
				registry: "http://127.0.0.1",
			},
			"http://127.0.0.1/v2/next",
		},
	}

	for intention, testCase := range cases {
		t.Run(intention, func(t *testing.T) {
			t.Parallel()

			if got := getNextURL(testCase.args.headers, testCase.args.registry); got != testCase.want {
				t.Errorf("getNextURL() =`%s`, want`%s`", got, testCase.want)
			}
		})
	}
}
