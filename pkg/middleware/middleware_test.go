package middleware

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ViBiOh/httputils/v3/pkg/request"
	"github.com/ViBiOh/ketchup/pkg/model"
)

type testUserService struct{}

func (tus testUserService) Create(_ context.Context, _ model.User) (model.User, error) {
	return model.NoneUser, nil
}

func (tus testUserService) StoreInContext(ctx context.Context) context.Context {
	return model.StoreUser(ctx, model.User{Email: "nobody@localhost"})
}

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
				w.Write([]byte(model.ReadUser(r.Context()).Email))
			}),
			httptest.NewRequest(http.MethodGet, "/", nil),
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
			writer := httptest.NewRecorder()
			New(testUserService{}).Middleware(tc.next).ServeHTTP(writer, tc.request)

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
