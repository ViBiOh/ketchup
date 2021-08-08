package middleware

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	authModel "github.com/ViBiOh/auth/v2/pkg/model"
	"github.com/ViBiOh/httputils/v4/pkg/request"
	"github.com/ViBiOh/ketchup/pkg/mocks"
	"github.com/ViBiOh/ketchup/pkg/model"
	"github.com/golang/mock/gomock"
)

func TestMiddleware(t *testing.T) {
	var cases = []struct {
		intention  string
		next       http.Handler
		request    *http.Request
		want       string
		wantStatus int
		wantHeader http.Header
	}{
		{
			"simple",
			http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if _, err := w.Write([]byte(model.ReadUser(r.Context()).Email)); err != nil {
					t.Errorf("unable to write: %s", err)
				}
			}),
			httptest.NewRequest(http.MethodGet, "/", nil).WithContext(authModel.StoreUser(context.Background(), authModel.NewUser(1, "test"))),
			"nobody@localhost",
			http.StatusOK,
			http.Header{},
		},
		{
			"nil",
			nil,
			httptest.NewRequest(http.MethodGet, "/", nil),
			"",
			http.StatusOK,
			http.Header{},
		},
	}

	for _, tc := range cases {
		t.Run(tc.intention, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			userService := mocks.NewUserService(ctrl)
			if tc.intention == "simple" {
				userService.EXPECT().StoreInContext(gomock.Any()).Return(model.StoreUser(context.Background(), model.NewUser(1, "nobody@localhost", authModel.NewUser(1, "test"))))
			}

			writer := httptest.NewRecorder()
			New(userService).Middleware(tc.next).ServeHTTP(writer, tc.request)

			if got := writer.Code; got != tc.wantStatus {
				t.Errorf("Middleware = %d, want %d", got, tc.wantStatus)
			}

			if got, _ := request.ReadBodyResponse(writer.Result()); string(got) != tc.want {
				t.Errorf("Middleware = `%s`, want `%s`", string(got), tc.want)
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
