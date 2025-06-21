package tests

import (
	"net/http"
	"net/url"
	"testing"

	"github.com/brianvoe/gofakeit/v6"
	"github.com/gavv/httpexpect/v2"
	"github.com/stretchr/testify/require"

	"url-shortener/internal/http_server/handlers/url/save"
	"url-shortener/internal/lib/api"
	"url-shortener/internal/lib/random"
)

const (
	host = "localhost:8082"
)

func TestURLShortener_HappyPath(t *testing.T) {
	u := url.URL{
		Scheme: "http",
		Host:   host,
	}
	e := httpexpect.Default(t, u.String())

	e.POST("/save").
		WithJSON(save.Request{
			URL:   gofakeit.URL(),
			Alias: random.NewRandomString(10),
		}).
		WithBasicAuth("us", "pass").
		Expect().
		Status(http.StatusOK).
		JSON().Object().
		ContainsKey("alias")
}

func TestURLShortener_SaveRedirect(t *testing.T) {
	testCases := []struct {
		name  string
		url   string
		alias string
		error string
	}{
		{
			name:  "Valid URL",
			url:   gofakeit.URL(),
			alias: gofakeit.Word() + gofakeit.Word(),
			error: "",
		},
		{
			name:  "Invalid URL",
			url:   "invalid",
			alias: gofakeit.Word(),
			error: "field URL must be a valid url",
		},
		{
			name:  "Empty alias",
			url:   gofakeit.URL(),
			error: "",
		},
		{
			name:  "Delete existing URL",
			url:   gofakeit.URL(),
			alias: gofakeit.Word() + gofakeit.Word(),
			error: "",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			u := url.URL{
				Scheme: "http",
				Host:   host,
			}

			e := httpexpect.Default(t, u.String())

			resp := e.POST("/save").
				WithJSON(save.Request{
					URL:   tc.url,
					Alias: tc.alias,
				}).
				WithBasicAuth("us", "pass").
				Expect().
				Status(http.StatusOK).
				JSON().Object()

			if tc.error != "" {
				resp.NotContainsKey("alias")
				resp.Value("error").String().IsEqual(tc.error)
				return
			}

			alias := tc.alias

			if tc.alias == "" {
				resp.Value("alias").String().Length().IsEqual(save.AliasLength)
				alias = resp.Value("alias").String().Raw()
			} else {
				resp.Value("alias").String().NotEmpty()
				alias = resp.Value("alias").String().Raw()
			}

			testRedirect(t, alias, tc.url)

			e.DELETE("/delete/"+alias).
				WithBasicAuth("us", "pass").
				Expect().
				Status(http.StatusOK).
				JSON().Object().
				Value("status").String().IsEqual("OK")

			testRedirectNotFound(t, alias)
		})
	}
}

func testRedirect(t *testing.T, alias, urlToRedirect string) {
	client := http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}

	u := url.URL{
		Scheme: "http",
		Host:   host,
		Path:   alias,
	}

	req, err := http.NewRequest("GET", u.String(), nil)
	require.NoError(t, err)

	req.SetBasicAuth("us", "pass")

	resp, err := client.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	require.Equal(t, http.StatusFound, resp.StatusCode, "redirect status code should be 302")

	redirectedToURL := resp.Header.Get("Location")

	require.Equal(t, urlToRedirect, redirectedToURL)
}

func testRedirectNotFound(t *testing.T, alias string) {
	u := url.URL{
		Scheme: "http",
		Host:   host,
		Path:   alias,
	}

	_, err := api.GetRedirectedURL(u.String())
	require.ErrorIs(t, err, api.ErrInvalidStatusCode)
}
