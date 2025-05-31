package redirect

import (
	"errors"
	"log/slog"
	"net/http"
	"net/url"
	resp "url-shortener/internal/lib/api/response"
	"url-shortener/internal/lib/logger/sl"
	"url-shortener/internal/storage"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
)

//go:generate go run github.com/vektra/mockery/v2@v2 --name=URLGetter
type URLGetter interface {
	GetURL(alias string) (string, error)
}

func Get(log *slog.Logger, urlGetter URLGetter) http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		const op = "handlers.redirect.Get"

		log = log.With(
			slog.String("op", op),
			slog.String("request_id", middleware.GetReqID(request.Context())),
		)

		alias := chi.URLParam(request, "alias")

		if alias == "" {
			log.Info("alias is empty")

			render.Status(request, http.StatusBadRequest)
			render.JSON(writer, request, resp.Error("alias is empty"))
			return
		}

		retrievedURL, err := urlGetter.GetURL(alias)
		if errors.Is(err, storage.ErrUrlNotFound) {
			log.Info("url not found", slog.String("alias", alias))

			render.Status(request, http.StatusBadRequest)
			render.JSON(writer, request, resp.Error("url not found"))

			return
		}
		if err != nil {
			log.Info("failed to get url", sl.Err(err))

			render.Status(request, http.StatusInternalServerError)
			render.JSON(writer, request, resp.Error("internal server error"))

			return
		}

		log.Info("got url", slog.String("url", retrievedURL))

		parsedURL, parseErr := url.Parse(retrievedURL)
		if parseErr != nil {
			log.Error("failed to parse retrieved redirect URL", sl.Err(parseErr), slog.String("original_url", retrievedURL))
			render.Status(request, http.StatusInternalServerError)
			render.JSON(writer, request, resp.Error("internal server error - malformed redirect URL"))
			return
		}

		http.Redirect(writer, request, parsedURL.String(), http.StatusFound)
	}
}
