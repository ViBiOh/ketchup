package renderer

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/ViBiOh/httputils/v3/pkg/request"
	"github.com/ViBiOh/ketchup/pkg/model"
)

type testKetchupService struct{}

func (tks testKetchupService) List(ctx context.Context, page, pageSize uint) ([]model.Ketchup, uint64, error) {
	if model.ReadUser(ctx) == model.NoneUser {
		return nil, 0, errors.New("user not found")
	}

	return []model.Ketchup{
		{Version: "1.0.0", Repository: model.Repository{ID: 1, Name: "vibioh/ketchup", Version: "1.0.0"}, User: model.User{ID: 1, Email: "nobody@localhost"}},
	}, 0, nil
}

func (tks testKetchupService) ListForRepositories(ctx context.Context, repositories []model.Repository) ([]model.Ketchup, error) {
	return nil, nil
}

func (tks testKetchupService) Create(ctx context.Context, item model.Ketchup) (model.Ketchup, error) {
	return model.NoneKetchup, nil
}

func (tks testKetchupService) Update(ctx context.Context, item model.Ketchup) (model.Ketchup, error) {
	return model.NoneKetchup, nil
}

func (tks testKetchupService) Delete(ctx context.Context, item model.Ketchup) error {
	return nil
}

func TestHandler(t *testing.T) {
	publicPath := "http:/localhost:1080"
	templatesDir = "../../templates"
	testInterface, err := New(Config{uiPath: &publicPath}, testKetchupService{}, nil)
	if err != nil {
		t.Errorf("unable to create app: %s", err)
	}
	testApp := testInterface.(app)

	wantedHeader := http.Header{}
	wantedHeader.Set("Content-Type", "text/html; charset=UTF-8")

	var cases = []struct {
		intention  string
		request    *http.Request
		want       string
		wantStatus int
		wantHeader http.Header
	}{
		{
			"no user",
			httptest.NewRequest(http.MethodGet, "/", nil),
			"user not found",
			http.StatusInternalServerError,
			wantedHeader,
		},
		{
			"simple",
			httptest.NewRequest(http.MethodGet, "/", nil).WithContext(model.StoreUser(context.Background(), model.User{ID: 1})),
			"vibioh/ketchup</h2>",
			http.StatusOK,
			wantedHeader,
		},
		{
			"with message",
			httptest.NewRequest(http.MethodGet, "/?messageContent=RedirectedMessage", nil).WithContext(model.StoreUser(context.Background(), model.User{ID: 1})),
			"RedirectedMessage",
			http.StatusOK,
			wantedHeader,
		},
	}

	for _, tc := range cases {
		t.Run(tc.intention, func(t *testing.T) {
			writer := httptest.NewRecorder()
			testApp.Handler().ServeHTTP(writer, tc.request)

			if got := writer.Code; got != tc.wantStatus {
				t.Errorf("Handler() = %d, want %d", got, tc.wantStatus)
			}

			if got, _ := request.ReadBodyResponse(writer.Result()); !strings.Contains(string(got), tc.want) {
				t.Errorf("Handler() = `%s`, want `%s`", string(got), tc.want)
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
