package renderer

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/ViBiOh/httputils/v3/pkg/request"
	"github.com/ViBiOh/ketchup/pkg/service"
)

func TestErrorHandler(t *testing.T) {
	publicPath := "http:/localhost:1080"
	templatesDir = "../../templates"
	testInterface, err := New(Config{uiPath: &publicPath}, nil, nil)
	if err != nil {
		t.Errorf("unable to create app: %s", err)
	}
	testApp := testInterface.(app)

	type args struct {
		status int
		err    error
	}

	var cases = []struct {
		intention  string
		args       args
		want       string
		wantStatus int
		wantHeader http.Header
	}{
		{
			"nil",
			args{
				status: http.StatusOK,
				err:    nil,
			},
			"<html",
			http.StatusOK,
			http.Header{},
		},
		{
			"random error",
			args{
				status: http.StatusTooManyRequests,
				err:    errors.New("stop spamming us dude"),
			},
			"stop spamming us dude",
			http.StatusTooManyRequests,
			http.Header{},
		},
		{
			"invalid",
			args{
				status: http.StatusOK,
				err:    service.WrapInvalid(errors.New("wrong name")),
			},
			"Invalid form",
			http.StatusBadRequest,
			http.Header{},
		},
		{
			"internal",
			args{
				status: http.StatusOK,
				err:    service.WrapInternal(errors.New("bad sql")),
			},
			"Oops! Something went wrong.",
			http.StatusInternalServerError,
			http.Header{},
		},
		{
			"not found",
			args{
				status: http.StatusOK,
				err:    service.WrapNotFound(errors.New("vibioh/ketchup")),
			},
			"vibioh/ketchup",
			http.StatusNotFound,
			http.Header{},
		},
		{
			"not found",
			args{
				status: http.StatusOK,
				err:    errors.New("invalid method for form"),
			},
			"invalid method",
			http.StatusMethodNotAllowed,
			http.Header{},
		},
	}

	for _, tc := range cases {
		t.Run(tc.intention, func(t *testing.T) {
			writer := httptest.NewRecorder()
			testApp.errorHandler(writer, tc.args.status, tc.args.err)

			if got := writer.Code; got != tc.wantStatus {
				t.Errorf("errorHandler = %d, want %d", got, tc.wantStatus)
			}

			if got, _ := request.ReadBodyResponse(writer.Result()); !strings.Contains(string(got), tc.want) {
				t.Errorf("errorHandler = `%s`, want `%s`", string(got), tc.want)
			}

			for key := range tc.wantHeader {
				want := tc.wantHeader.Get(key)
				if got := writer.Header().Get(key); got != want {
					t.Errorf("`%s` Header = `%s`, want `%s`", key, got, want)
				}
			}
		})
	}
}
