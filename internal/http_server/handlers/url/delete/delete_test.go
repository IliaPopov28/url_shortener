package delete_test

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	delHandler "url-shortener/internal/http_server/handlers/url/delete" // Используем псевдоним delHandler
	"url-shortener/internal/http_server/handlers/url/delete/mocks"
	resp "url-shortener/internal/lib/api/response"
	"url-shortener/internal/lib/logger/handlers/slogdiscard"
	"url-shortener/internal/storage"
)

type testCase struct {
	name           string
	alias          string
	mockError      error
	expectedStatus int
	expectedBody   string
}

func TestDeleteHandler(t *testing.T) {
	makeOkBody := func() string {
		jsonBody, _ := json.Marshal(resp.OK())
		return string(jsonBody)
	}
	makeErrorBody := func(msg string) string {
		jsonBody, _ := json.Marshal(resp.Error(msg))
		return string(jsonBody)
	}

	cases := []testCase{
		{
			name:           "Success",
			alias:          "test_alias",
			mockError:      nil,
			expectedStatus: http.StatusOK,
			expectedBody:   makeOkBody(),
		},
		{
			name:           "Success with dot in alias",
			alias:          "kelen.cc",
			mockError:      nil,
			expectedStatus: http.StatusOK,
			expectedBody:   makeOkBody(),
		},
		{
			name:           "Empty alias",
			alias:          "",
			expectedStatus: http.StatusBadRequest,
			expectedBody:   makeErrorBody("invalid request"),
		},
		{
			name:           "Alias not found",
			alias:          "non_existent_alias",
			mockError:      storage.ErrUrlNotFound,
			expectedStatus: http.StatusNotFound,
			expectedBody:   makeErrorBody("alias not found"),
		},
		{
			name:           "Internal error from URLDeleter",
			alias:          "some_alias_for_internal_error",
			mockError:      errors.New("internal storage error"),
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   makeErrorBody("internal server error"),
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			urlDeleterMock := mocks.NewURLDeleter(t)

			// Настраиваем мок только если это необходимо для кейса
			// (т.е. если не ожидается ошибка из-за пустого алиаса до вызова Deleter)
			if tc.alias != "" {
				urlDeleterMock.On("DeleteURL", tc.alias).
					Return(tc.mockError).
					Maybe() // Используем Maybe, так как DeleteURL не всегда будет вызван
			}

			requestPath := "/"
			if tc.alias != "" {
				// Для DELETE запросов параметры обычно передаются в пути
				requestPath = "/" + tc.alias
			}

			// URL-кодируем путь для NewRequest
			parsedPath := &url.URL{Path: requestPath}
			req := httptest.NewRequest(http.MethodDelete, parsedPath.String(), nil)

			// Установка chi.RouteCtxKey для корректной работы chi.URLParam
			rctx := chi.NewRouteContext()
			rctx.URLParams.Add("alias", tc.alias) // Передаем ожидаемый алиас в контекст
			req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

			rr := httptest.NewRecorder()
			router := chi.NewRouter()
			handler := delHandler.Delete(slogdiscard.NewDiscardLogger(), urlDeleterMock)

			// Регистрируем маршрут. Chi должен сам корректно обрабатывать точки в параметрах.
			router.Delete("/{alias}", handler)
			if tc.alias == "" { // Если алиас пустой, используем маршрут без параметра
				router.Delete("/", handler)
			}

			router.ServeHTTP(rr, req)

			// Проверяем статус код
			require.Equal(t, tc.expectedStatus, rr.Code, "Case: %s - Unexpected status code", tc.name)

			// Проверяем тело ответа, если ожидается
			if tc.expectedBody != "" {
				assert.JSONEq(t, tc.expectedBody, rr.Body.String(), "Case: %s - Unexpected response body", tc.name)
			}

			// Проверяем вызовы мока
			if tc.alias != "" { // DeleteURL не должен вызываться для пустого алиаса
				urlDeleterMock.AssertCalled(t, "DeleteURL", tc.alias)
			} else {
				urlDeleterMock.AssertNotCalled(t, "DeleteURL", tc.alias)
			}
		})
	}
}
