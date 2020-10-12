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
	"github.com/ViBiOh/ketchup/pkg/service/repository/repositorytest"
)

type testKetchupService struct{}

func (tks testKetchupService) List(ctx context.Context, _, _ uint) ([]model.Ketchup, uint64, error) {
	if model.ReadUser(ctx) == model.NoneUser {
		return nil, 0, errors.New("user not found")
	}

	return []model.Ketchup{
		{Version: "1.0.0", Repository: model.Repository{ID: 1, Name: "vibioh/ketchup", Version: "1.0.0"}, User: model.User{ID: 1, Email: "nobody@localhost"}},
	}, 0, nil
}

func (tks testKetchupService) ListForRepositories(_ context.Context, _ []model.Repository) ([]model.Ketchup, error) {
	return nil, nil
}

func (tks testKetchupService) Create(_ context.Context, _ model.Ketchup) (model.Ketchup, error) {
	return model.NoneKetchup, nil
}

func (tks testKetchupService) Update(_ context.Context, _ model.Ketchup) (model.Ketchup, error) {
	return model.NoneKetchup, nil
}

func (tks testKetchupService) Delete(_ context.Context, _ model.Ketchup) error {
	return nil
}

func TestHandler(t *testing.T) {
	publicPath := "http:/localhost:1080"
	templatesDir = "../../templates"
	testInterface, err := New(Config{uiPath: &publicPath}, testKetchupService{}, nil, repositorytest.NewApp(false))
	if err != nil {
		t.Errorf("unable to create app: %s", err)
	}
	testApp := testInterface.(app)

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
			http.Header{
				"Content-Type": []string{"text/html; charset=UTF-8"},
			},
		},
		{
			"simple",
			httptest.NewRequest(http.MethodGet, "/", nil).WithContext(model.StoreUser(context.Background(), model.User{ID: 1})),
			"vibioh/ketchup</h2>",
			http.StatusOK,
			http.Header{
				"Content-Type": []string{"text/html; charset=UTF-8"},
			},
		},
		{
			"with message",
			httptest.NewRequest(http.MethodGet, "/?messageContent=RedirectedMessage", nil).WithContext(model.StoreUser(context.Background(), model.User{ID: 1})),
			"RedirectedMessage",
			http.StatusOK,
			http.Header{
				"Content-Type": []string{"text/html; charset=UTF-8"},
			},
		},
		{
			"to ketchups",
			httptest.NewRequest(http.MethodGet, "/ketchups/", nil),
			"invalid method GET",
			http.StatusMethodNotAllowed,
			http.Header{
				"Content-Type": []string{"text/html; charset=UTF-8"},
			},
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

func TestPublicHandler(t *testing.T) {
	publicPath := "http:/localhost:1080"
	templatesDir = "../../templates"
	staticDir = "../../static"
	testInterface, err := New(Config{uiPath: &publicPath}, testKetchupService{}, nil, nil)
	if err != nil {
		t.Errorf("unable to create app: %s", err)
	}
	testApp := testInterface.(app)

	var cases = []struct {
		intention  string
		request    *http.Request
		want       string
		wantStatus int
		wantHeader http.Header
	}{
		{
			"static not found",
			httptest.NewRequest(http.MethodGet, "/favicon/unknown", nil),
			`not found`,
			http.StatusNotFound,
			http.Header{
				"Content-Type": []string{"text/plain; charset=utf-8"},
			},
		},
		{
			"static",
			httptest.NewRequest(http.MethodGet, "/favicon/site.webmanifest", nil),
			`"name": "Ketchup"`,
			http.StatusOK,
			http.Header{
				"Content-Type": []string{"text/plain; charset=utf-8"},
			},
		},
		{
			"robots",
			httptest.NewRequest(http.MethodGet, "/robots.txt", nil),
			"User-agent",
			http.StatusOK,
			http.Header{
				"Content-Type": []string{"text/plain; charset=utf-8"},
			},
		},
		{
			"sitemap",
			httptest.NewRequest(http.MethodGet, "/sitemap.xml", nil),
			"<urlset",
			http.StatusOK,
			http.Header{
				"Content-Type": []string{"application/xml"},
			},
		},
		{
			"svg not found",
			httptest.NewRequest(http.MethodGet, "/svg/not-found", nil),
			"¯\\_(ツ)_/¯",
			http.StatusNotFound,
			http.Header{},
		},
		{
			"svg",
			httptest.NewRequest(http.MethodGet, "/svg/times", nil),
			"<svg",
			http.StatusOK,
			http.Header{
				"Content-Type": []string{"image/svg+xml"},
			},
		},
		{
			"signup",
			httptest.NewRequest(http.MethodGet, "/signup", nil),
			"",
			http.StatusMethodNotAllowed,
			http.Header{},
		},
		{
			"public",
			httptest.NewRequest(http.MethodGet, "/", nil),
			"Signup</a>",
			http.StatusOK,
			http.Header{
				"Content-Type": []string{"text/html; charset=UTF-8"},
			},
		},
		{
			"public with message",
			httptest.NewRequest(http.MethodGet, "/?messageContent=welcomeToKetchup", nil),
			"welcomeToKetchup",
			http.StatusOK,
			http.Header{
				"Content-Type": []string{"text/html; charset=UTF-8"},
			},
		},
		{
			"not found",
			httptest.NewRequest(http.MethodGet, "/unkown-path", nil),
			"¯\\_(ツ)_/¯",
			http.StatusNotFound,
			http.Header{
				"Content-Type": []string{"text/plain; charset=utf-8"},
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.intention, func(t *testing.T) {
			writer := httptest.NewRecorder()
			testApp.PublicHandler().ServeHTTP(writer, tc.request)

			if got := writer.Code; got != tc.wantStatus {
				t.Errorf("PublicHandler = %d, want %d", got, tc.wantStatus)
			}

			if got, _ := request.ReadBodyResponse(writer.Result()); !strings.Contains(string(got), tc.want) {
				t.Errorf("PublicHandler = `%s`, want `%s`", string(got), tc.want)
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
