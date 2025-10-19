package middleware

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	authModel "github.com/ViBiOh/auth/v3/pkg/model"
	"github.com/ViBiOh/httputils/v4/pkg/request"
	"github.com/ViBiOh/ketchup/pkg/mocks"
	"github.com/ViBiOh/ketchup/pkg/model"
	"go.uber.org/mock/gomock"
)

func TestMiddleware(t *testing.T) {
	t.Parallel()

	cases := map[string]struct {
		next       http.Handler
		request    *http.Request
		want       string
		wantStatus int
		wantHeader http.Header
	}{
		"simple": {
			http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if _, err := w.Write([]byte(model.ReadUser(r.Context()).Email)); err != nil {
					t.Errorf("write: %s", err)
				}
			}),
			httptest.NewRequest(http.MethodGet, "/", nil).WithContext(authModel.StoreUser(context.TODO(), authModel.NewUser(1, "test"))),
			"nobody@localhost",
			http.StatusOK,
			http.Header{},
		},
		"nil": {
			nil,
			httptest.NewRequest(http.MethodGet, "/", nil),
			"",
			http.StatusOK,
			http.Header{},
		},
	}

	for intention, testCase := range cases {
		t.Run(intention, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)

			userService := mocks.NewUserService(ctrl)
			if intention == "simple" {
				userService.EXPECT().StoreInContext(gomock.Any()).Return(model.StoreUser(context.TODO(), model.NewUser(1, "nobody@localhost", authModel.NewUser(1, "test"))))
			}

			writer := httptest.NewRecorder()
			New(userService).Middleware(testCase.next).ServeHTTP(writer, testCase.request)

			if got := writer.Code; got != testCase.wantStatus {
				t.Errorf("Middleware = %d, want %d", got, testCase.wantStatus)
			}

			if got, _ := request.ReadBodyResponse(writer.Result()); string(got) != testCase.want {
				t.Errorf("Middleware = `%s`, want `%s`", string(got), testCase.want)
			}

			for key := range testCase.wantHeader {
				want := testCase.wantHeader.Get(key)
				if got := writer.Header().Get(key); got != want {
					t.Errorf("`%s` Header = `%s`, want `%s`", key, got, want)
				}
			}
		})
	}
}
