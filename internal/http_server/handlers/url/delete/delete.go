package delete

import (
	"errors"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
	"log/slog"
	"net/http"
	resp "url-shortener/internal/lib/api/response"
	"url-shortener/internal/lib/logger/sl"
	"url-shortener/internal/storage"
)

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

		//извлекаем alias из URL параметра
		alias := chi.URLParam(request, "alias")
		if alias == "" {
			log.Error("alias is empty")
			render.JSON(writer, request, resp.Error("invalid request"))
			writer.WriteHeader(http.StatusBadRequest)
			return
		}

		// Вызываем метод удаления URL
		err := urlDeleter.DeleteURL(alias)
		if errors.Is(err, storage.ErrUrlNotFound) {
			log.Info("alias not found", slog.String("alias", alias))
			render.JSON(writer, request, resp.Error("alias not found"))
			writer.WriteHeader(http.StatusNotFound)
			return
		}
		if err != nil {
			log.Error("failed to delete url", sl.Err(err))
			render.JSON(writer, request, resp.Error("internal server error"))
			writer.WriteHeader(http.StatusInternalServerError)
			return
		}

		log.Info("alias deleted", slog.String("alias", alias))

		render.JSON(writer, request, resp.OK())
	}
}
