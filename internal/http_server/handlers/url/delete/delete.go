package delete

import (
	"errors"
	"log/slog"
	"net/http"
	resp "url-shortener/internal/lib/api/response"
	"url-shortener/internal/lib/logger/sl"
	"url-shortener/internal/storage"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
)

//go:generate go run github.com/vektra/mockery/v2@v2 --name=URLDeleter
type URLDeleter interface {
	DeleteURL(alias string) error
}

func Delete(log *slog.Logger, urlDeleter URLDeleter) http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		const op = "handlers.url.delete.Delete"

		log = log.With(
			slog.String("op", op),
			slog.String("request_id", middleware.GetReqID(request.Context())),
		)

		alias := chi.URLParam(request, "alias")
		log.Info("Alias received from chi.URLParam", slog.String("param_alias", alias))

		if alias == "" {
			log.Error("alias is empty")
			render.Status(request, http.StatusBadRequest)
			render.JSON(writer, request, resp.Error("invalid request"))
			return
		}

		err := urlDeleter.DeleteURL(alias)
		if errors.Is(err, storage.ErrUrlNotFound) {
			log.Info("alias not found", slog.String("alias", alias))
			render.Status(request, http.StatusNotFound)
			render.JSON(writer, request, resp.Error("alias not found"))
			return
		}
		if err != nil {
			log.Error("failed to delete url", sl.Err(err))
			render.Status(request, http.StatusInternalServerError)
			render.JSON(writer, request, resp.Error("internal server error"))
			return
		}

		log.Info("alias deleted", slog.String("alias", alias))
		render.JSON(writer, request, resp.OK())
	}
}
