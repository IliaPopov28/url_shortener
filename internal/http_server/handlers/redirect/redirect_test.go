package redirect_test

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"url-shortener/internal/http_server/handlers/redirect"
	"url-shortener/internal/http_server/handlers/redirect/mocks"
	resp "url-shortener/internal/lib/api/response"
	"url-shortener/internal/lib/logger/handlers/slogdiscard"
	"url-shortener/internal/storage"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type testCase struct {
	name             string
	alias            string
	mockURL          string
	mockError        error
	expectedStatus   int
	expectedLocation string
	expectedBody     string
}

func TestRedirectHandler(t *testing.T) {
	makeErrorBody := func(msg string) string {
		jsonBody, _ := json.Marshal(resp.Error(msg))
		return string(jsonBody)
	}

	cases := []testCase{
		{
			name:             "Success",
			alias:            "test_alias",
			mockURL:          "https://google.com",
			expectedStatus:   http.StatusFound,
			expectedLocation: "https://google.com",
		},
		{
			name:           "Empty alias",
			alias:          "",
			expectedStatus: http.StatusBadRequest,
			expectedBody:   makeErrorBody("alias is empty"),
		},
		{
			name:           "Alias not found",
			alias:          "non_existent_alias",
			mockError:      storage.ErrUrlNotFound,
			expectedStatus: http.StatusBadRequest,
			expectedBody:   makeErrorBody("url not found"),
		},
		{
			name:           "Internal error from URLGetter",
			alias:          "some_alias",
			mockError:      errors.New("internal storage error"),
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   makeErrorBody("internal server error"),
		},
		{
			name:             "Success with alias needing URL encoding",
			alias:            "alias with spaces",
			mockURL:          "https://example.com/path_without_spaces",
			expectedStatus:   http.StatusFound,
			expectedLocation: "https://example.com/path_without_spaces",
		},
		{
			name:             "Success with URL to redirect having spaces",
			alias:            "redirect_space_alias",
			mockURL:          "https://example.com/target path with spaces",
			expectedStatus:   http.StatusFound,
			expectedLocation: "https://example.com/target%20path%20with%20spaces",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			urlGetterMock := mocks.NewURLGetter(t)

			if tc.mockURL != "" || tc.mockError != nil {
				urlGetterMock.On("GetURL", tc.alias).
					Return(tc.mockURL, tc.mockError).
					Maybe()
			}

			requestPath := "/"
			if tc.alias != "" {
				requestPath = "/" + tc.alias
			}

			parsedPath := &url.URL{Path: requestPath}
			req := httptest.NewRequest(http.MethodGet, parsedPath.String(), nil)

			rctx := chi.NewRouteContext()
			rctx.URLParams.Add("alias", tc.alias)
			req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

			rr := httptest.NewRecorder()
			router := chi.NewRouter()
			handler := redirect.Get(slogdiscard.NewDiscardLogger(), urlGetterMock)
			router.Get("/{alias}", handler)
			router.Get("/", handler)
			router.ServeHTTP(rr, req)

			require.Equal(t, tc.expectedStatus, rr.Code, "Case: %s - Unexpected status code", tc.name)

			if tc.expectedLocation != "" {
				assert.Equal(t, tc.expectedLocation, rr.Header().Get("Location"), "Case: %s - Unexpected location header", tc.name)
			}

			if tc.expectedBody != "" {
				assert.JSONEq(t, tc.expectedBody, rr.Body.String(), "Case: %s - Unexpected response body", tc.name)
			}

			if tc.mockURL != "" || tc.mockError != nil {
				if tc.alias != "" && tc.name != "Empty alias" {
					urlGetterMock.AssertCalled(t, "GetURL", tc.alias)
				}
			} else if tc.alias != "" && tc.name != "Empty alias" {
			}

			if tc.name == "Empty alias" {
				urlGetterMock.AssertNotCalled(t, "GetURL", tc.alias)
			}
		})
	}
}
